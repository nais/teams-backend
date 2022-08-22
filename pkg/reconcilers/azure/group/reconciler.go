package azure_group_reconciler

import (
	"context"
	"fmt"
	"github.com/nais/console/pkg/db"
	"strings"

	"github.com/google/uuid"
	helpers "github.com/nais/console/pkg/console"
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
	oauth       clientcredentials.Config
	client      azureclient.Client
	domain      string
}

func New(database db.Database, auditLogger auditlogger.AuditLogger, oauth clientcredentials.Config, client azureclient.Client, domain string) *azureGroupReconciler {
	return &azureGroupReconciler{
		database:    database,
		auditLogger: auditLogger,
		oauth:       oauth,
		client:      client,
		domain:      domain,
	}
}

const Name = sqlc.SystemNameAzureGroup

func NewFromConfig(ctx context.Context, database db.Database, cfg *config.Config, auditLogger auditlogger.AuditLogger) (reconcilers.Reconciler, error) {
	if !cfg.Azure.Enabled {
		return nil, reconcilers.ErrReconcilerNotEnabled
	}

	endpoint := microsoft.AzureADEndpoint(cfg.Azure.TenantID)
	conf := clientcredentials.Config{
		ClientID:     cfg.Azure.ClientID,
		ClientSecret: cfg.Azure.ClientSecret,
		TokenURL:     endpoint.TokenURL,
		AuthStyle:    endpoint.AuthStyle,
		Scopes: []string{
			"https://graph.microsoft.com/.default",
		},
	}

	return New(database, auditLogger, conf, azureclient.New(conf.Client(context.Background())), cfg.TenantDomain), nil
}

func (r *azureGroupReconciler) Name() sqlc.SystemName {
	return Name
}

func (r *azureGroupReconciler) Reconcile(ctx context.Context, input reconcilers.Input) error {
	state := &reconcilers.AzureState{}
	err := r.database.LoadSystemState(ctx, r.Name(), input.Team.ID, state)
	if err != nil {
		return fmt.Errorf("unable to load system state for team '%s' in system '%s': %w", input.Team.Slug, r.Name(), err)
	}

	prefixedName := teamNameWithPrefix(input.Team.Slug)
	grp, created, err := r.client.GetOrCreateGroup(ctx, *state, prefixedName, input.Team.Name, &input.Team.Purpose.String)
	if err != nil {
		return err
	}

	if created {
		r.auditLogger.Logf(ctx, sqlc.AuditActionAzureGroupCreate, input.CorrelationID, r.Name(), nil, &input.Team.Slug, nil, "created Azure AD group: %s", grp)

		id, _ := uuid.Parse(grp.ID)
		err = r.database.SetSystemState(ctx, r.Name(), input.Team.ID, reconcilers.AzureState{GroupID: &id})
		if err != nil {
			log.Errorf("system state not persisted: %s", err)
		}
	}

	err = r.connectUsers(ctx, grp, input)
	if err != nil {
		return fmt.Errorf("%s: add members to group: %s", sqlc.AuditActionAzureGroupAddMembers, err)
	}

	return nil
}

func (r *azureGroupReconciler) connectUsers(ctx context.Context, grp *azureclient.Group, input reconcilers.Input) error {
	consoleTeamMembers := input.Team.Members
	members, err := r.client.ListGroupMembers(ctx, grp)
	if err != nil {
		return fmt.Errorf("%s: list existing members in Azure group '%s': %s", sqlc.AuditActionAzureGroupAddMembers, grp.MailNickname, err)
	}

	consoleUserMap := make(map[string]*db.User)
	localMembers := helpers.DomainUsers(consoleTeamMembers, r.domain)

	membersToRemove := remoteOnlyMembers(members, localMembers)
	for _, member := range membersToRemove {
		remoteEmail := strings.ToLower(member.Mail)
		err = r.client.RemoveMemberFromGroup(ctx, grp, member)
		if err != nil {
			log.Warnf("%s: unable to remove member '%s' from group '%s' in Azure: %s", sqlc.AuditActionAzureGroupDeleteMember, remoteEmail, grp.MailNickname, err)
			continue
		}

		if _, exists := consoleUserMap[remoteEmail]; !exists {
			user, err := r.database.GetUserByEmail(ctx, remoteEmail)
			if err != nil {
				log.Warnf("%s: unable to lookup local user with email '%s': %s", sqlc.AuditActionAzureGroupDeleteMember, remoteEmail, err)
				continue
			}
			consoleUserMap[remoteEmail] = user
		}

		r.auditLogger.Logf(ctx, sqlc.AuditActionAzureGroupDeleteMember, input.CorrelationID, r.Name(), nil, &input.Team.Slug, &remoteEmail, "removed member '%s' from Azure group '%s'", remoteEmail, grp.MailNickname)
	}

	membersToAdd := localOnlyMembers(members, localMembers)
	for _, consoleUser := range membersToAdd {
		member, err := r.client.GetUser(ctx, consoleUser.Email)
		if err != nil {
			log.Warnf("%s: unable to lookup user with email '%s' in Azure: %s", sqlc.AuditActionAzureGroupAddMember, consoleUser.Email, err)
			continue
		}
		err = r.client.AddMemberToGroup(ctx, grp, member)
		if err != nil {
			log.Warnf("%s: unable to add member '%s' to Azure group '%s': %s", sqlc.AuditActionAzureGroupAddMember, consoleUser.Email, grp.MailNickname, err)
			continue
		}

		r.auditLogger.Logf(ctx, sqlc.AuditActionAzureGroupAddMember, input.CorrelationID, r.Name(), nil, &input.Team.Slug, &consoleUser.Email, "added member '%s' to Azure group '%s'", consoleUser.Email, grp.MailNickname)
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

func teamNameWithPrefix(slug string) string {
	return reconcilers.TeamNamePrefix + slug
}
