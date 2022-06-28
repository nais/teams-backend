package google_workspace_admin_reconciler

import (
	"context"
	"fmt"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	helpers "github.com/nais/console/pkg/console"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/google_jwt"
	"github.com/nais/console/pkg/reconcilers"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2/jwt"
	admin_directory_v1 "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

type googleWorkspaceAdminReconciler struct {
	auditLogger auditlogger.AuditLogger
	db          *gorm.DB
	domain      string
	config      *jwt.Config
	system      dbmodels.System
}

const (
	Name                    = "google:workspace-admin"
	OpCreate                = "google:workspace-admin:create"
	OpAddMember             = "google:workspace-admin:add-member"
	OpAddMembers            = "google:workspace-admin:add-members"
	OpDeleteMember          = "google:workspace-admin:delete-member"
	OpAddToGKESecurityGroup = "google:workspace-admin:add-to-gke-security-group"
)

func New(db *gorm.DB, system dbmodels.System, auditLogger auditlogger.AuditLogger, domain string, config *jwt.Config) *googleWorkspaceAdminReconciler {
	return &googleWorkspaceAdminReconciler{
		auditLogger: auditLogger,
		db:          db,
		domain:      domain,
		config:      config,
		system:      system,
	}
}

func NewFromConfig(db *gorm.DB, cfg *config.Config, system dbmodels.System, auditLogger auditlogger.AuditLogger) (reconcilers.Reconciler, error) {
	if !cfg.Google.Enabled {
		return nil, reconcilers.ErrReconcilerNotEnabled
	}

	config, err := google_jwt.GetConfig(cfg.Google.CredentialsFile, cfg.Google.DelegatedUser)

	if err != nil {
		return nil, fmt.Errorf("get google jwt config: %w", err)
	}

	return New(db, system, auditLogger, cfg.TenantDomain, config), nil
}

func (r *googleWorkspaceAdminReconciler) Reconcile(ctx context.Context, input reconcilers.Input) error {
	state := &reconcilers.GoogleWorkspaceState{}
	err := dbmodels.LoadSystemState(r.db, *r.system.ID, *input.Team.ID, state)
	if err != nil {
		return fmt.Errorf("unable to load system state for team '%s' in system '%s': %w", input.Team.Slug, r.system.Name, err)
	}

	client := r.config.Client(ctx)
	srv, err := admin_directory_v1.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("retrieve directory client: %w", err)
	}

	grp, err := r.getOrCreateGroup(srv.Groups, state, input.Corr, input.Team)
	if err != nil {
		return fmt.Errorf("unable to get or create a Google Workspace group for team '%s' in system '%s': %w", input.Team.Slug, r.system.Name, err)
	}

	err = dbmodels.SetSystemState(r.db, *r.system.ID, *input.Team.ID, reconcilers.GoogleWorkspaceState{GroupID: &grp.Id})
	if err != nil {
		log.Errorf("system state not persisted: %s", err)
	}

	err = r.connectUsers(srv.Members, grp, input.Corr, input.Team)
	if err != nil {
		return fmt.Errorf("%s: add members to group: %w", OpAddMembers, err)
	}

	return r.addToGKESecurityGroup(srv.Members, grp, input.Corr, input.Team)
}

func (r *googleWorkspaceAdminReconciler) System() dbmodels.System {
	return r.system
}

func (r *googleWorkspaceAdminReconciler) getOrCreateGroup(groupsService *admin_directory_v1.GroupsService, state *reconcilers.GoogleWorkspaceState, corr dbmodels.Correlation, team dbmodels.Team) (*admin_directory_v1.Group, error) {
	if state.GroupID != nil {
		existingGroup, err := groupsService.Get(*state.GroupID).Do()
		if err == nil {
			return existingGroup, nil
		}
	}

	groupKey := reconcilers.TeamNamePrefix + team.Slug
	email := fmt.Sprintf("%s@%s", groupKey, r.domain)
	newGroup := &admin_directory_v1.Group{
		Email:       email,
		Name:        team.Name,
		Description: helpers.TeamPurpose(team.Purpose),
	}
	group, err := groupsService.Insert(newGroup).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to create Google Directory group: %w", err)
	}

	r.auditLogger.Logf(OpCreate, corr, r.system, nil, &team, nil, "created Google Directory group '%s'", group.Email)

	return group, nil
}

func (r *googleWorkspaceAdminReconciler) connectUsers(membersService *admin_directory_v1.MembersService, grp *admin_directory_v1.Group, corr dbmodels.Correlation, team dbmodels.Team) error {
	membersAccordingToGoogle, err := membersService.List(grp.Id).Do()
	if err != nil {
		return fmt.Errorf("%s: list existing members in Google Directory group: %w", OpAddMembers, err)
	}

	consoleUserMap := make(map[string]*dbmodels.User)
	localMembers := helpers.DomainUsers(team.Users, r.domain)

	membersToRemove := remoteOnlyMembers(membersAccordingToGoogle.Members, localMembers)
	for _, member := range membersToRemove {
		remoteMemberEmail := strings.ToLower(member.Email)
		err = membersService.Delete(grp.Id, member.Id).Do()
		if err != nil {
			log.Warnf("%s: delete member '%s' from Google Directory group '%s': %s", OpDeleteMember, remoteMemberEmail, grp.Email, err)
			continue
		}

		if _, exists := consoleUserMap[remoteMemberEmail]; !exists {
			consoleUserMap[remoteMemberEmail] = dbmodels.GetUserByEmail(r.db, remoteMemberEmail)
		}

		r.auditLogger.Logf(OpDeleteMember, corr, r.system, nil, &team, consoleUserMap[remoteMemberEmail], "deleted member '%s' from Google Directory group '%s'", member.Email, grp.Email)
	}

	membersToAdd := localOnlyMembers(membersAccordingToGoogle.Members, localMembers)
	for _, user := range membersToAdd {
		member := &admin_directory_v1.Member{
			Email: user.Email,
		}
		_, err = membersService.Insert(grp.Id, member).Do()
		if err != nil {
			log.Warnf("%s: add member '%s' to Google Directory group '%s': %s", OpAddMember, member.Email, grp.Email, err)
			continue
		}
		r.auditLogger.Logf(OpAddMember, corr, r.system, nil, &team, user, "added member '%s' to Google Directory group '%s'", member.Email, grp.Email)
	}

	return nil
}

func (r *googleWorkspaceAdminReconciler) addToGKESecurityGroup(membersService *admin_directory_v1.MembersService, grp *admin_directory_v1.Group, corr dbmodels.Correlation, team dbmodels.Team) error {
	const groupPrefix = "gke-security-groups@"
	groupKey := groupPrefix + r.domain

	member := &admin_directory_v1.Member{
		Email: grp.Email,
	}

	_, err := membersService.Insert(groupKey, member).Do()
	if err != nil {
		googleError, ok := err.(*googleapi.Error)
		if ok && googleError.Code == http.StatusConflict {
			return nil
		}
		return fmt.Errorf("%s: add group '%s' to GKE security group '%s': %s", OpAddToGKESecurityGroup, member.Email, groupKey, err)
	}

	r.auditLogger.Logf(OpAddToGKESecurityGroup, corr, r.system, nil, &team, nil, "added group '%s' to GKE security group '%s'", member.Email, groupKey)

	return nil
}

// remoteOnlyMembers Given a list of Google group members and a list of Console users, return Google group members not
// present in Console user list.
func remoteOnlyMembers(googleGroupMembers []*admin_directory_v1.Member, consoleUsers []*dbmodels.User) []*admin_directory_v1.Member {
	googleGroupMemberMap := make(map[string]*admin_directory_v1.Member)
	for _, member := range googleGroupMembers {
		googleGroupMemberMap[strings.ToLower(member.Email)] = member
	}
	for _, user := range consoleUsers {
		delete(googleGroupMemberMap, user.Email)
	}
	googleGroupMembers = make([]*admin_directory_v1.Member, 0, len(googleGroupMemberMap))
	for _, member := range googleGroupMemberMap {
		googleGroupMembers = append(googleGroupMembers, member)
	}
	return googleGroupMembers
}

// Given a list of Google group members and a list of users,
// return users not present in members directory.
func localOnlyMembers(googleGroupMembers []*admin_directory_v1.Member, consoleUsers []*dbmodels.User) []*dbmodels.User {
	localUserMap := make(map[string]*dbmodels.User)
	for _, user := range consoleUsers {
		localUserMap[user.Email] = user
	}
	for _, member := range googleGroupMembers {
		delete(localUserMap, strings.ToLower(member.Email))
	}
	consoleUsers = make([]*dbmodels.User, 0, len(localUserMap))
	for _, user := range localUserMap {
		consoleUsers = append(consoleUsers, user)
	}
	return consoleUsers
}
