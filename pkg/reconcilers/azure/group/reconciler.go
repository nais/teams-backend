package azure_group_reconciler

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/teams-backend/pkg/types"

	"github.com/google/uuid"
	"github.com/nais/teams-backend/pkg/auditlogger"
	"github.com/nais/teams-backend/pkg/azureclient"
	"github.com/nais/teams-backend/pkg/config"
	"github.com/nais/teams-backend/pkg/db"
	"github.com/nais/teams-backend/pkg/logger"
	"github.com/nais/teams-backend/pkg/reconcilers"
	"github.com/nais/teams-backend/pkg/slug"
	"github.com/nais/teams-backend/pkg/sqlc"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/oauth2/microsoft"
)

type azureGroupReconciler struct {
	database    db.Database
	auditLogger auditlogger.AuditLogger
	client      azureclient.Client
	domain      string
	log         logger.Logger
}

func New(database db.Database, auditLogger auditlogger.AuditLogger, client azureclient.Client, domain string, log logger.Logger) *azureGroupReconciler {
	return &azureGroupReconciler{
		database:    database,
		auditLogger: auditLogger,
		client:      client,
		domain:      domain,
		log:         log.WithComponent(types.ComponentNameAzureGroup),
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

func NewFromConfig(ctx context.Context, database db.Database, cfg *config.Config, log logger.Logger) (reconcilers.Reconciler, error) {
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

	return New(database, auditlogger.New(database, types.ComponentNameAzureGroup, log), azureclient.New(conf.Client(context.Background())), cfg.TenantDomain, log), nil
}

func (r *azureGroupReconciler) Name() sqlc.ReconcilerName {
	return Name
}

func (r *azureGroupReconciler) Reconcile(ctx context.Context, input reconcilers.Input) error {
	state := &reconcilers.AzureState{}
	err := r.database.LoadReconcilerStateForTeam(ctx, r.Name(), input.Team.Slug, state)
	if err != nil {
		return fmt.Errorf("unable to load system state for team %q in system %q: %w", input.Team.Slug, r.Name(), err)
	}

	prefixedName := teamNameWithPrefix(input.Team.Slug)
	grp, created, err := r.client.GetOrCreateGroup(ctx, state.GroupID, prefixedName, input.Team.Purpose)
	if err != nil {
		return err
	}

	if created {
		targets := []auditlogger.Target{
			auditlogger.TeamTarget(input.Team.Slug),
		}
		fields := auditlogger.Fields{
			Action:        types.AuditActionAzureGroupCreate,
			CorrelationID: input.CorrelationID,
		}
		r.auditLogger.Logf(ctx, targets, fields, "Created Azure AD group %q with ID %q", grp.MailNickname, grp.ID)

		id, _ := uuid.Parse(grp.ID)
		err = r.database.SetReconcilerStateForTeam(ctx, r.Name(), input.Team.Slug, reconcilers.AzureState{GroupID: &id})
		if err != nil {
			r.log.WithError(err).Error("persiste system state")
		}
	}

	err = r.connectUsers(ctx, grp, input)
	if err != nil {
		return fmt.Errorf("add members to group: %s", err)
	}

	return nil
}

func (r *azureGroupReconciler) Delete(ctx context.Context, teamSlug slug.Slug, correlationID uuid.UUID) error {
	state := &reconcilers.AzureState{}
	err := r.database.LoadReconcilerStateForTeam(ctx, r.Name(), teamSlug, state)
	if err != nil {
		return fmt.Errorf("load reconciler state for team %q in reconciler %q: %w", teamSlug, r.Name(), err)
	}

	if state.GroupID == nil {
		r.log.Warnf("missing group ID in reconciler state for team %q in reconciler %q, assume already deleted", teamSlug, r.Name())
		return r.database.RemoveReconcilerStateForTeam(ctx, r.Name(), teamSlug)
	}

	grpID := *state.GroupID

	err = r.client.DeleteGroup(ctx, grpID)
	if err != nil {
		return fmt.Errorf("delete Azure AD group with ID %q for team %q: %w", grpID, teamSlug, err)
	}

	targets := []auditlogger.Target{
		auditlogger.TeamTarget(teamSlug),
	}
	fields := auditlogger.Fields{
		Action:        types.AuditActionAzureGroupDelete,
		CorrelationID: correlationID,
	}
	r.auditLogger.Logf(ctx, targets, fields, "Delete Azure AD group with ID %q", grpID)

	return r.database.RemoveReconcilerStateForTeam(ctx, Name, teamSlug)
}

func (r *azureGroupReconciler) connectUsers(ctx context.Context, grp *azureclient.Group, input reconcilers.Input) error {
	members, err := r.client.ListGroupMembers(ctx, grp)
	if err != nil {
		return fmt.Errorf("list existing members in Azure group %q: %s", grp.MailNickname, err)
	}

	teamsBackendUserMap := make(map[string]*db.User)
	membersToRemove := remoteOnlyMembers(members, input.TeamMembers)
	for _, member := range membersToRemove {
		remoteEmail := strings.ToLower(member.Mail)
		err = r.client.RemoveMemberFromGroup(ctx, grp, member)
		if err != nil {
			r.log.WithError(err).Errorf("remove member %q from group %q in Azure", remoteEmail, grp.MailNickname)
			continue
		}

		if _, exists := teamsBackendUserMap[remoteEmail]; !exists {
			user, err := r.database.GetUserByEmail(ctx, remoteEmail)
			if err != nil {
				r.log.WithError(err).Warnf("lookup local user with email %q", remoteEmail)
				continue
			}
			teamsBackendUserMap[remoteEmail] = user
		}

		targets := []auditlogger.Target{
			auditlogger.TeamTarget(input.Team.Slug),
			auditlogger.UserTarget(remoteEmail),
		}
		fields := auditlogger.Fields{
			Action:        types.AuditActionAzureGroupDeleteMember,
			CorrelationID: input.CorrelationID,
		}
		r.auditLogger.Logf(ctx, targets, fields, "Removed member %q from Azure group %q", remoteEmail, grp.MailNickname)
	}

	membersToAdd := localOnlyMembers(members, input.TeamMembers)
	for _, teamsBackendUser := range membersToAdd {
		member, err := r.client.GetUser(ctx, teamsBackendUser.Email)
		if err != nil {
			r.log.WithError(err).Warnf("lookup user with email %q in Azure", teamsBackendUser.Email)
			continue
		}
		err = r.client.AddMemberToGroup(ctx, grp, member)
		if err != nil {
			r.log.WithError(err).Warnf("add member %q to Azure group %q", teamsBackendUser.Email, grp.MailNickname)
			continue
		}

		targets := []auditlogger.Target{
			auditlogger.TeamTarget(input.Team.Slug),
			auditlogger.UserTarget(teamsBackendUser.Email),
		}
		fields := auditlogger.Fields{
			Action:        types.AuditActionAzureGroupAddMember,
			CorrelationID: input.CorrelationID,
		}
		r.auditLogger.Logf(ctx, targets, fields, "Added member %q to Azure group %q", teamsBackendUser.Email, grp.MailNickname)
	}

	return nil
}

// localOnlyMembers Given a list of Azure group members and a list of teams-backend users, return teams-backend users
// not present in the Azure group member list. The email address is used to compare objects.
func localOnlyMembers(azureGroupMembers []*azureclient.Member, teamsBackendUsers []*db.User) []*db.User {
	localUserMap := make(map[string]*db.User)
	for _, user := range teamsBackendUsers {
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

// remoteOnlyMembers Given a list of Azure group members and a list of teams-backend users, return Azure group members
// not present in teams-backend user list. The email address is used to compare objects.
func remoteOnlyMembers(azureGroupMembers []*azureclient.Member, teamsBackendUsers []*db.User) []*azureclient.Member {
	azureGroupMemberMap := make(map[string]*azureclient.Member)
	for _, member := range azureGroupMembers {
		azureGroupMemberMap[strings.ToLower(member.Mail)] = member
	}
	for _, user := range teamsBackendUsers {
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
