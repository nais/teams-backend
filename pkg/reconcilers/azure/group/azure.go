package azure_group_reconciler

import (
	"context"
	"fmt"
	helpers "github.com/nais/console/pkg/console"
	log "github.com/sirupsen/logrus"
	"strings"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/azureclient"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/reconcilers/registry"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/oauth2/microsoft"
	"gorm.io/gorm"
)

// gitHubReconciler creates teams on GitHub and connects users to them.
type azureReconciler struct {
	system      dbmodels.System
	auditLogger auditlogger.AuditLogger
	oauth       clientcredentials.Config
	client      azureclient.Client
	domain      string
}

const (
	Name            = "azure:group"
	OpCreate        = "azure:group:create"
	OpAddMember     = "azure:group:add-member"
	OpAddMembers    = "azure:group:add-members"
	OpDeleteMember  = "azure:group:delete-member"
	OpDeleteMembers = "azure:group:delete-members"
)

func init() {
	registry.Register(Name, NewFromConfig)
}

func New(system dbmodels.System, auditLogger auditlogger.AuditLogger, oauth clientcredentials.Config, client azureclient.Client, domain string) *azureReconciler {
	return &azureReconciler{
		system:      system,
		auditLogger: auditLogger,
		oauth:       oauth,
		client:      client,
		domain:      domain,
	}
}

func NewFromConfig(_ *gorm.DB, cfg *config.Config, system dbmodels.System, auditLogger auditlogger.AuditLogger) (reconcilers.Reconciler, error) {
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

	return New(system, auditLogger, conf, azureclient.New(conf.Client(context.Background())), cfg.PartnerDomain), nil
}

func (s *azureReconciler) Reconcile(ctx context.Context, input reconcilers.Input) error {
	prefixedName := teamNameWithPrefix(input.Team.Slug)
	grp, created, err := s.client.GetOrCreateGroup(ctx, prefixedName, input.Team.Name, input.Team.Purpose)
	if err != nil {
		return err
	}

	if created {
		s.auditLogger.Log(OpCreate, input.Corr, s.system, nil, &input.Team, nil, "created Azure AD group: %s", grp)
	}

	err = s.connectUsers(ctx, grp, input.Corr, input.Team)
	if err != nil {
		return fmt.Errorf("%s: add members to group: %s", OpAddMembers, err)
	}

	return nil
}

func (s *azureReconciler) connectUsers(ctx context.Context, grp *azureclient.Group, corr dbmodels.Correlation, team dbmodels.Team) error {
	members, err := s.client.ListGroupMembers(ctx, grp)
	if err != nil {
		return fmt.Errorf("%s: list existing members in Azure group: %s", OpAddMembers, err)
	}

	localMembers := helpers.DomainUsers(team.Users, s.domain)

	deleteMembers := extraMembers(members, localMembers)
	createUsers := missingUsers(members, localMembers)

	for _, member := range deleteMembers {
		err = s.client.RemoveMemberFromGroup(ctx, grp, member)
		if err != nil {
			return fmt.Errorf("%s: delete member '%s' from Azure group '%s': %s", OpDeleteMember, member.Mail, grp.MailNickname, err)
		}

		// FIXME: connect audit log with database user
		s.auditLogger.Log(OpDeleteMember, corr, s.system, nil, &team, nil, "deleted member '%s' from Azure group '%s'", member.Mail, grp.MailNickname)
	}

	for _, user := range createUsers {
		member, err := s.client.GetUser(ctx, user.Email)
		if err != nil {
			log.Warnf("%s: Unable to lookup user with email '%s' in Azure", OpAddMember, user.Email)
			continue
		}
		err = s.client.AddMemberToGroup(ctx, grp, member)
		if err != nil {
			return fmt.Errorf("%s: add member '%s' to Azure group '%s': %s", OpAddMember, member.Mail, grp.MailNickname, err)
		}

		// FIXME: connect audit log with database user
		s.auditLogger.Log(OpAddMember, corr, s.system, nil, &team, nil, "added member '%s' to Azure group '%s'", member.Mail, grp.MailNickname)
	}

	return nil
}

// Given a list of Azure group members and a list of users,
// return users not present in members list.
func missingUsers(members []*azureclient.Member, users []*dbmodels.User) []*dbmodels.User {
	userMap := make(map[string]*dbmodels.User)
	for _, user := range users {
		userMap[user.Email] = user
	}
	for _, member := range members {
		delete(userMap, strings.ToLower(member.Mail))
	}
	users = make([]*dbmodels.User, 0, len(userMap))
	for _, user := range userMap {
		users = append(users, user)
	}
	return users
}

// Given a list of Azure group members and a list of users,
// return members not present in user list.
func extraMembers(members []*azureclient.Member, users []*dbmodels.User) []*azureclient.Member {
	memberMap := make(map[string]*azureclient.Member)
	for _, member := range members {
		memberMap[strings.ToLower(member.Mail)] = member
	}
	for _, user := range users {
		delete(memberMap, user.Email)
	}
	members = make([]*azureclient.Member, 0, len(memberMap))
	for _, member := range memberMap {
		members = append(members, member)
	}
	return members
}

func teamNameWithPrefix(slug dbmodels.Slug) string {
	return reconcilers.TeamNamePrefix + string(slug)
}
