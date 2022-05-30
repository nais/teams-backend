package azure_group_reconciler

import (
	"context"
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
	logger auditlogger.Logger
	oauth  clientcredentials.Config
	client azureclient.Client
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

func New(logger auditlogger.Logger, oauth clientcredentials.Config, client azureclient.Client) *azureReconciler {
	return &azureReconciler{
		logger: logger,
		oauth:  oauth,
		client: client,
	}
}

func NewFromConfig(_ *gorm.DB, cfg *config.Config, logger auditlogger.Logger) (reconcilers.Reconciler, error) {
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

	return New(logger, conf, azureclient.New(conf.Client(context.Background()))), nil
}

func (s *azureReconciler) Reconcile(ctx context.Context, in reconcilers.Input) error {
	grp, err := s.client.GetOrCreateGroup(ctx, teamNameWithPrefix(in.Team.Slug), *in.Team.Name, *in.Team.Purpose)
	if err != nil {
		return err
	}

	// FIXME: support changing metadata (group name, description)

	err = s.connectUsers(ctx, grp, in)
	if err != nil {
		return s.logger.Errorf(in, OpAddMembers, "add members to group: %s", err)
	}

	return nil
}

func (s *azureReconciler) connectUsers(ctx context.Context, grp *azureclient.Group, in reconcilers.Input) error {
	members, err := s.client.ListGroupMembers(ctx, grp)
	if err != nil {
		return s.logger.Errorf(in, OpAddMembers, "list existing members in Azure group: %s", err)
	}

	deleteMembers := extraMembers(members, in.Team.Users)
	createUsers := missingUsers(members, in.Team.Users)

	for _, member := range deleteMembers {
		// FIXME: connect audit log with database user, if exists
		err = s.client.RemoveMemberFromGroup(ctx, grp, member)
		if err != nil {
			return s.logger.UserErrorf(in, OpDeleteMember, nil, "delete member '%s' from Azure group '%s': %s", member.Mail, grp.MailNickname, err)
		}
		s.logger.UserLogf(in, OpDeleteMember, nil, "deleted member '%s' from Azure group '%s'", member.Mail, grp.MailNickname)
	}

	s.logger.Logf(in, OpDeleteMembers, "all unmanaged members successfully deleted from Azure group '%s'", grp.MailNickname)

	for _, user := range createUsers {
		if user.Email == nil {
			continue
		}
		member, err := s.client.GetUser(ctx, *user.Email)
		if err != nil {
			s.logger.UserLogf(in, OpAddMember, user, "Unable to lookup user with email '%s' in Azure", *user.Email)
			continue
		}
		err = s.client.AddMemberToGroup(ctx, grp, member)
		if err != nil {
			return s.logger.UserErrorf(in, OpAddMember, user, "add member '%s' to Azure group '%s': %s", member.Mail, grp.MailNickname, err)
		}
		s.logger.UserLogf(in, OpAddMember, user, "added member '%s' to Azure group '%s'", member.Mail, grp.MailNickname)
	}

	s.logger.Logf(in, OpAddMembers, "all members successfully added to Azure group '%s'", grp.MailNickname)

	return nil
}

// Given a list of Azure group members and a list of users,
// return users not present in members list.
func missingUsers(members []*azureclient.Member, users []*dbmodels.User) []*dbmodels.User {
	userMap := make(map[string]*dbmodels.User)
	for _, user := range users {
		if user.Email == nil {
			continue
		}
		userMap[*user.Email] = user
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
		if user.Email == nil {
			continue
		}
		delete(memberMap, *user.Email)
	}
	members = make([]*azureclient.Member, 0, len(memberMap))
	for _, member := range memberMap {
		members = append(members, member)
	}
	return members
}

func teamNameWithPrefix(slug *dbmodels.Slug) string {
	if slug == nil {
		panic("nil slug passed to teamNameWithPrefix")
	}
	return reconcilers.TeamNamePrefix + slug.String()
}
