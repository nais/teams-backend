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
	db          *gorm.DB
	system      dbmodels.System
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

func init() {
	registry.Register(Name, NewFromConfig)
}

func New(db *gorm.DB, system dbmodels.System, auditLogger auditlogger.AuditLogger, oauth clientcredentials.Config, client azureclient.Client, domain string) *azureReconciler {
	return &azureReconciler{
		db:          db,
		system:      system,
		auditLogger: auditLogger,
		oauth:       oauth,
		client:      client,
		domain:      domain,
	}
}

func NewFromConfig(db *gorm.DB, cfg *config.Config, system dbmodels.System, auditLogger auditlogger.AuditLogger) (reconcilers.Reconciler, error) {
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

	return New(db, system, auditLogger, conf, azureclient.New(conf.Client(context.Background())), cfg.PartnerDomain), nil
}

func (s *azureReconciler) Reconcile(ctx context.Context, input reconcilers.Input) error {
	prefixedName := teamNameWithPrefix(input.Team.Slug)
	grp, created, err := s.client.GetOrCreateGroup(ctx, prefixedName, input.Team.Name, input.Team.Purpose)
	if err != nil {
		return err
	}

	if created {
		s.auditLogger.Logf(OpCreate, input.Corr, s.system, nil, &input.Team, nil, "created Azure AD group: %s", grp)
	}

	err = s.connectUsers(ctx, grp, input.Corr, input.Team)
	if err != nil {
		return fmt.Errorf("%s: add members to group: %s", OpAddMembers, err)
	}

	return nil
}

func (s *azureReconciler) getUserByEmail(email string) *dbmodels.User {
	localUser := &dbmodels.User{}
	err := s.db.Where("email = ?", email).First(localUser).Error
	if err != nil {
		return nil
	}

	return localUser
}

func (s *azureReconciler) connectUsers(ctx context.Context, grp *azureclient.Group, corr dbmodels.Correlation, team dbmodels.Team) error {
	members, err := s.client.ListGroupMembers(ctx, grp)
	if err != nil {
		return fmt.Errorf("%s: list existing members in Azure group '%s': %s", OpAddMembers, grp.MailNickname, err)
	}

	consoleUserMap := make(map[string]*dbmodels.User)
	localMembers := helpers.DomainUsers(team.Users, s.domain)

	membersToRemove := remoteOnlyMembers(members, localMembers)
	for _, member := range membersToRemove {
		remoteEmail := strings.ToLower(member.Mail)
		err = s.client.RemoveMemberFromGroup(ctx, grp, member)
		if err != nil {
			log.Warnf("%s: unable to remove member '%s' from group '%s' in Azure: %s", OpDeleteMember, remoteEmail, grp.MailNickname, err)
			continue
		}

		if _, exists := consoleUserMap[remoteEmail]; !exists {
			consoleUserMap[remoteEmail] = s.getUserByEmail(remoteEmail)
		}

		s.auditLogger.Logf(OpDeleteMember, corr, s.system, nil, &team, consoleUserMap[remoteEmail], "removed member '%s' from Azure group '%s'", remoteEmail, grp.MailNickname)
	}

	membersToAdd := localOnlyMembers(members, localMembers)
	for _, consoleUser := range membersToAdd {
		member, err := s.client.GetUser(ctx, consoleUser.Email)
		if err != nil {
			log.Warnf("%s: unable to lookup user with email '%s' in Azure: %s", OpAddMember, consoleUser.Email, err)
			continue
		}
		err = s.client.AddMemberToGroup(ctx, grp, member)
		if err != nil {
			log.Warnf("%s: unable to add member '%s' to Azure group '%s': %s", OpAddMember, consoleUser.Email, grp.MailNickname, err)
			continue
		}

		s.auditLogger.Logf(OpAddMember, corr, s.system, nil, &team, consoleUser, "added member '%s' to Azure group '%s'", member.Mail, grp.MailNickname)
	}

	return nil
}

// localOnlyMembers Given a list of Azure group members and a list of Console users, return Console users not present in
// the Azure group member list. The email address is used to compare objects.
func localOnlyMembers(azureGroupMembers []*azureclient.Member, consoleUsers []*dbmodels.User) []*dbmodels.User {
	localUserMap := make(map[string]*dbmodels.User)
	for _, user := range consoleUsers {
		localUserMap[user.Email] = user
	}
	for _, member := range azureGroupMembers {
		delete(localUserMap, strings.ToLower(member.Mail))
	}
	localUsers := make([]*dbmodels.User, 0, len(localUserMap))
	for _, user := range localUserMap {
		localUsers = append(localUsers, user)
	}
	return localUsers
}

// remoteOnlyMembers Given a list of Azure group members and a list of Console users, return Azure group members not
// present in Console user list. The email address is used to compare objects.
func remoteOnlyMembers(azureGroupMembers []*azureclient.Member, consoleUsers []*dbmodels.User) []*azureclient.Member {
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
