package azure_group_reconciler

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	helpers "github.com/nais/console/pkg/console"
	"github.com/nais/console/pkg/sqlc"
	log "github.com/sirupsen/logrus"
	"strings"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/azureclient"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/reconcilers"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/oauth2/microsoft"
	"gorm.io/gorm"
)

type azureGroupReconciler struct {
	queries     sqlc.Querier
	db          *gorm.DB
	system      sqlc.System
	auditLogger auditlogger.AuditLogger
	oauth       clientcredentials.Config
	client      azureclient.Client
	domain      string
}

const (
	Name           = "azure:group"
	OpCreate       = "azure:group:create"
	OpAddMember    = "azure:group:add-member"
	OpAddMembers   = "azure:group:add-members"
	OpDeleteMember = "azure:group:delete-member"
)

func New(queries sqlc.Querier, db *gorm.DB, system sqlc.System, auditLogger auditlogger.AuditLogger, oauth clientcredentials.Config, client azureclient.Client, domain string) *azureGroupReconciler {
	return &azureGroupReconciler{
		queries:     queries,
		db:          db,
		system:      system,
		auditLogger: auditLogger,
		oauth:       oauth,
		client:      client,
		domain:      domain,
	}
}

func NewFromConfig(queries sqlc.Querier, db *gorm.DB, cfg *config.Config, system sqlc.System, auditLogger auditlogger.AuditLogger) (reconcilers.Reconciler, error) {
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

	return New(queries, db, system, auditLogger, conf, azureclient.New(conf.Client(context.Background())), cfg.TenantDomain), nil
}

func (r *azureGroupReconciler) Reconcile(ctx context.Context, input reconcilers.Input) error {
	state := &reconcilers.AzureState{}
	err := dbmodels.LoadSystemState(r.db, r.system.ID, input.Team.ID, state)
	if err != nil {
		return fmt.Errorf("unable to load system state for team '%s' in system '%s': %w", input.Team.Slug, r.system.Name, err)
	}

	prefixedName := teamNameWithPrefix(input.Team.Slug)
	grp, created, err := r.client.GetOrCreateGroup(ctx, *state, prefixedName, input.Team.Name, &input.Team.Purpose.String)
	if err != nil {
		return err
	}

	if created {
		r.auditLogger.Logf(OpCreate, input.Corr, r.system, nil, input.Team, nil, "created Azure AD group: %s", grp)

		id, _ := uuid.Parse(grp.ID)
		err = dbmodels.SetSystemState(r.db, r.system.ID, input.Team.ID, reconcilers.AzureState{GroupID: &id})
		if err != nil {
			log.Errorf("system state not persisted: %s", err)
		}
	}

	err = r.connectUsers(ctx, grp, input.Corr, *input.Team, input.Members)
	if err != nil {
		return fmt.Errorf("%s: add members to group: %s", OpAddMembers, err)
	}

	return nil
}

func (r *azureGroupReconciler) System() sqlc.System {
	return r.system
}

func (r *azureGroupReconciler) connectUsers(ctx context.Context, grp *azureclient.Group, corr sqlc.Correlation, team sqlc.Team, consoleTeamMembers []*sqlc.User) error {
	members, err := r.client.ListGroupMembers(ctx, grp)
	if err != nil {
		return fmt.Errorf("%s: list existing members in Azure group '%s': %s", OpAddMembers, grp.MailNickname, err)
	}

	consoleUserMap := make(map[string]*sqlc.User)
	localMembers := helpers.DomainUsers(consoleTeamMembers, r.domain)

	membersToRemove := remoteOnlyMembers(members, localMembers)
	for _, member := range membersToRemove {
		remoteEmail := strings.ToLower(member.Mail)
		err = r.client.RemoveMemberFromGroup(ctx, grp, member)
		if err != nil {
			log.Warnf("%s: unable to remove member '%s' from group '%s' in Azure: %s", OpDeleteMember, remoteEmail, grp.MailNickname, err)
			continue
		}

		if _, exists := consoleUserMap[remoteEmail]; !exists {
			// user := r.queries.GetUserByEmail(ctx, remoteEmail)
			user, err := r.queries.GetUserByEmail(ctx, remoteEmail)
			if err != nil {
				log.Warnf("%s: unable to lookup local user with email '%s': %s", OpDeleteMember, remoteEmail, err)
				continue
			}
			consoleUserMap[remoteEmail] = user
		}

		r.auditLogger.Logf(OpDeleteMember, corr, r.system, nil, &team, consoleUserMap[remoteEmail], "removed member '%s' from Azure group '%s'", remoteEmail, grp.MailNickname)
	}

	membersToAdd := localOnlyMembers(members, localMembers)
	for _, consoleUser := range membersToAdd {
		member, err := r.client.GetUser(ctx, consoleUser.Email)
		if err != nil {
			log.Warnf("%s: unable to lookup user with email '%s' in Azure: %s", OpAddMember, consoleUser.Email, err)
			continue
		}
		err = r.client.AddMemberToGroup(ctx, grp, member)
		if err != nil {
			log.Warnf("%s: unable to add member '%s' to Azure group '%s': %s", OpAddMember, consoleUser.Email, grp.MailNickname, err)
			continue
		}

		r.auditLogger.Logf(OpAddMember, corr, r.system, nil, &team, consoleUser, "added member '%s' to Azure group '%s'", member.Mail, grp.MailNickname)
	}

	return nil
}

// localOnlyMembers Given a list of Azure group members and a list of Console users, return Console users not present in
// the Azure group member list. The email address is used to compare objects.
func localOnlyMembers(azureGroupMembers []*azureclient.Member, consoleUsers []*sqlc.User) []*sqlc.User {
	localUserMap := make(map[string]*sqlc.User)
	for _, user := range consoleUsers {
		localUserMap[user.Email] = user
	}
	for _, member := range azureGroupMembers {
		delete(localUserMap, strings.ToLower(member.Mail))
	}
	localUsers := make([]*sqlc.User, 0, len(localUserMap))
	for _, user := range localUserMap {
		localUsers = append(localUsers, user)
	}
	return localUsers
}

// remoteOnlyMembers Given a list of Azure group members and a list of Console users, return Azure group members not
// present in Console user list. The email address is used to compare objects.
func remoteOnlyMembers(azureGroupMembers []*azureclient.Member, consoleUsers []*sqlc.User) []*azureclient.Member {
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

func teamNameWithPrefix(slug dbmodels.Slug) string {
	return reconcilers.TeamNamePrefix + string(slug)
}
