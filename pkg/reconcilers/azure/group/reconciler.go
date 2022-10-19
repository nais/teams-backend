package azure_group_reconciler

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/slug"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/sqlc"
	log "github.com/sirupsen/logrus"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/azureclient"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/reconcilers"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/oauth2/microsoft"
)

type azureGroupReconciler struct {
	database    db.Database
	auditLogger auditlogger.AuditLogger
	client      azureclient.Client
	domain      string
}

func New(database db.Database, auditLogger auditlogger.AuditLogger, client azureclient.Client, domain string) *azureGroupReconciler {
	return &azureGroupReconciler{
		database:    database,
		auditLogger: auditLogger,
		client:      client,
		domain:      domain,
	}
}

const Name = sqlc.ReconcilerNameAzureGroup

type reconcilerConfig struct {
	clientID     string
	clientSecret string
	tenantID     string
}

func convertDatabaseConfig(ctx context.Context, database db.Database) (*reconcilerConfig, error) {
	config, err := database.DangerousGetReconcilerConfigValues(ctx, Name)
	if err != nil {
		return nil, err
	}

	return &reconcilerConfig{
		clientSecret: config.GetValue(sqlc.ReconcilerConfigKeyAzureClientSecret),
		clientID:     config.GetValue(sqlc.ReconcilerConfigKeyAzureClientID),
		tenantID:     config.GetValue(sqlc.ReconcilerConfigKeyAzureTenantID),
	}, nil
}

func NewFromConfig(ctx context.Context, database db.Database, cfg *config.Config, auditLogger auditlogger.AuditLogger) (reconcilers.Reconciler, error) {
	config, err := convertDatabaseConfig(ctx, database)
	if err != nil {
		return nil, err
	}

	endpoint := microsoft.AzureADEndpoint(config.tenantID)
	conf := clientcredentials.Config{
		ClientID:     config.clientID,
		ClientSecret: config.clientSecret,
		TokenURL:     endpoint.TokenURL,
		AuthStyle:    endpoint.AuthStyle,
		Scopes: []string{
			"https://graph.microsoft.com/.default",
		},
	}

	return New(database, auditLogger, azureclient.New(conf.Client(context.Background())), cfg.TenantDomain), nil
}

func (r *azureGroupReconciler) Name() sqlc.ReconcilerName {
	return Name
}

func (r *azureGroupReconciler) Reconcile(ctx context.Context, input reconcilers.Input) error {
	state := &reconcilers.AzureState{}
	err := r.database.LoadReconcilerStateForTeam(ctx, r.Name(), input.Team.ID, state)
	if err != nil {
		return fmt.Errorf("unable to load system state for team %q in system %q: %w", input.Team.Slug, r.Name(), err)
	}

	prefixedName := teamNameWithPrefix(input.Team.Slug)
	grp, created, err := r.client.GetOrCreateGroup(ctx, *state, prefixedName, input.Team.Name, &input.Team.Purpose)
	if err != nil {
		return err
	}

	if created {
		targets := []auditlogger.Target{
			auditlogger.TeamTarget(input.Team.Slug),
		}
		fields := auditlogger.Fields{
			Action:        sqlc.AuditActionAzureGroupCreate,
			CorrelationID: input.CorrelationID,
		}
		r.auditLogger.Logf(ctx, targets, fields, "created Azure AD group: %s", grp)

		id, _ := uuid.Parse(grp.ID)
		err = r.database.SetReconcilerStateForTeam(ctx, r.Name(), input.Team.ID, reconcilers.AzureState{GroupID: &id})
		if err != nil {
			log.Errorf("system state not persisted: %s", err)
		}
	}

	err = r.connectUsers(ctx, grp, input)
	if err != nil {
		return fmt.Errorf("add members to group: %s", err)
	}

	return nil
}

func (r *azureGroupReconciler) connectUsers(ctx context.Context, grp *azureclient.Group, input reconcilers.Input) error {
	members, err := r.client.ListGroupMembers(ctx, grp)
	if err != nil {
		return fmt.Errorf("list existing members in Azure group %q: %s", grp.MailNickname, err)
	}

	consoleUserMap := make(map[string]*db.User)
	membersToRemove := remoteOnlyMembers(members, input.TeamMembers)
	for _, member := range membersToRemove {
		remoteEmail := strings.ToLower(member.Mail)
		err = r.client.RemoveMemberFromGroup(ctx, grp, member)
		if err != nil {
			log.Warnf("unable to remove member %q from group %q in Azure: %s", remoteEmail, grp.MailNickname, err)
			continue
		}

		if _, exists := consoleUserMap[remoteEmail]; !exists {
			user, err := r.database.GetUserByEmail(ctx, remoteEmail)
			if err != nil {
				log.Warnf("unable to lookup local user with email %q: %s", remoteEmail, err)
				continue
			}
			consoleUserMap[remoteEmail] = user
		}

		targets := []auditlogger.Target{
			auditlogger.TeamTarget(input.Team.Slug),
			auditlogger.UserTarget(remoteEmail),
		}
		fields := auditlogger.Fields{
			Action:        sqlc.AuditActionAzureGroupDeleteMember,
			CorrelationID: input.CorrelationID,
		}
		r.auditLogger.Logf(ctx, targets, fields, "removed member %q from Azure group %q", remoteEmail, grp.MailNickname)
	}

	membersToAdd := localOnlyMembers(members, input.TeamMembers)
	for _, consoleUser := range membersToAdd {
		member, err := r.client.GetUser(ctx, consoleUser.Email)
		if err != nil {
			log.Warnf("unable to lookup user with email %q in Azure: %s", consoleUser.Email, err)
			continue
		}
		err = r.client.AddMemberToGroup(ctx, grp, member)
		if err != nil {
			log.Warnf("unable to add member %q to Azure group %q: %s", consoleUser.Email, grp.MailNickname, err)
			continue
		}

		targets := []auditlogger.Target{
			auditlogger.TeamTarget(input.Team.Slug),
			auditlogger.UserTarget(consoleUser.Email),
		}
		fields := auditlogger.Fields{
			Action:        sqlc.AuditActionAzureGroupAddMember,
			CorrelationID: input.CorrelationID,
		}
		r.auditLogger.Logf(ctx, targets, fields, "added member %q to Azure group %q", consoleUser.Email, grp.MailNickname)
	}

	return nil
}

// localOnlyMembers Given a list of Azure group members and a list of Console users, return Console users not present in
// the Azure group member list. The email address is used to compare objects.
func localOnlyMembers(azureGroupMembers []*azureclient.Member, consoleUsers []*db.User) []*db.User {
	localUserMap := make(map[string]*db.User)
	for _, user := range consoleUsers {
		localUserMap[user.Email] = user
	}
	for _, member := range azureGroupMembers {
		delete(localUserMap, strings.ToLower(member.Mail))
	}
	localUsers := make([]*db.User, 0, len(localUserMap))
	for _, user := range localUserMap {
		localUsers = append(localUsers, user)
	}
	return localUsers
}

// remoteOnlyMembers Given a list of Azure group members and a list of Console users, return Azure group members not
// present in Console user list. The email address is used to compare objects.
func remoteOnlyMembers(azureGroupMembers []*azureclient.Member, consoleUsers []*db.User) []*azureclient.Member {
	azureGroupMemberMap := make(map[string]*azureclient.Member)
	for _, member := range azureGroupMembers {
		azureGroupMemberMap[strings.ToLower(member.Mail)] = member
	}
	for _, user := range consoleUsers {
		delete(azureGroupMemberMap, user.Email)
	}
	azureGroupMembers = make([]*azureclient.Member, 0, len(azureGroupMemberMap))
	for _, member := range azureGroupMemberMap {
		azureGroupMembers = append(azureGroupMembers, member)
	}
	return azureGroupMembers
}

func teamNameWithPrefix(slug slug.Slug) string {
	return reconcilers.TeamNamePrefix + string(slug)
}
