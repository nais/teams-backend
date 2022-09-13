package google_workspace_admin_reconciler

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	helpers "github.com/nais/console/pkg/console"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/google_jwt"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/sqlc"
	log "github.com/sirupsen/logrus"
	admin_directory_v1 "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

type googleWorkspaceAdminReconciler struct {
	database     db.Database
	auditLogger  auditlogger.AuditLogger
	domain       string
	adminService *admin_directory_v1.Service
}

const (
	Name = sqlc.SystemNameGoogleWorkspaceAdmin
)

func New(database db.Database, auditLogger auditlogger.AuditLogger, domain string, adminService *admin_directory_v1.Service) *googleWorkspaceAdminReconciler {
	return &googleWorkspaceAdminReconciler{
		database:     database,
		auditLogger:  auditLogger,
		domain:       domain,
		adminService: adminService,
	}
}

func NewFromConfig(ctx context.Context, database db.Database, cfg *config.Config, auditLogger auditlogger.AuditLogger) (reconcilers.Reconciler, error) {
	if !cfg.Google.Enabled {
		return nil, reconcilers.ErrReconcilerNotEnabled
	}

	config, err := google_jwt.GetConfig(cfg.Google.CredentialsFile, cfg.Google.DelegatedUser)
	if err != nil {
		return nil, fmt.Errorf("get google jwt config: %w", err)
	}

	client := config.Client(ctx)
	srv, err := admin_directory_v1.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("retrieve directory client: %w", err)
	}

	return New(database, auditLogger, cfg.TenantDomain, srv), nil
}

func (r *googleWorkspaceAdminReconciler) Name() sqlc.SystemName {
	return Name
}

func (r *googleWorkspaceAdminReconciler) Reconcile(ctx context.Context, input reconcilers.Input) error {
	state := &reconcilers.GoogleWorkspaceState{}
	err := r.database.LoadSystemState(ctx, r.Name(), input.Team.ID, state)
	if err != nil {
		return fmt.Errorf("unable to load system state for team %q in system %q: %w", input.Team.Slug, r.Name(), err)
	}

	grp, err := r.getOrCreateGroup(ctx, state, input)
	if err != nil {
		return fmt.Errorf("unable to get or create a Google Workspace group for team %q in system %q: %w", input.Team.Slug, r.Name(), err)
	}

	err = r.database.SetSystemState(ctx, r.Name(), input.Team.ID, reconcilers.GoogleWorkspaceState{GroupEmail: &grp.Email})
	if err != nil {
		log.Errorf("system state not persisted: %s", err)
	}

	err = r.connectUsers(ctx, grp, input)
	if err != nil {
		return fmt.Errorf("add members to group: %w", err)
	}

	return r.addToGKESecurityGroup(ctx, grp, input)
}

func (r *googleWorkspaceAdminReconciler) getOrCreateGroup(ctx context.Context, state *reconcilers.GoogleWorkspaceState, input reconcilers.Input) (*admin_directory_v1.Group, error) {
	if state.GroupEmail != nil {
		existingGroup, err := r.adminService.Groups.Get(*state.GroupEmail).Do()
		if err == nil {
			return existingGroup, nil
		}
	}

	groupKey := reconcilers.TeamNamePrefix + input.Team.Slug
	email := fmt.Sprintf("%s@%s", groupKey, r.domain)
	newGroup := &admin_directory_v1.Group{
		Email:       email,
		Name:        input.Team.Name,
		Description: helpers.TeamPurpose(&input.Team.Purpose.String),
	}
	group, err := r.adminService.Groups.Insert(newGroup).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to create Google Directory group: %w", err)
	}

	fields := auditlogger.Fields{
		Action:         sqlc.AuditActionGoogleWorkspaceAdminCreate,
		CorrelationID:  input.CorrelationID,
		TargetTeamSlug: &input.Team.Slug,
	}
	r.auditLogger.Logf(ctx, fields, "created Google Directory group %q", group.Email)

	return group, nil
}

func (r *googleWorkspaceAdminReconciler) connectUsers(ctx context.Context, grp *admin_directory_v1.Group, input reconcilers.Input) error {
	membersAccordingToGoogle, err := r.adminService.Members.List(grp.Id).Do()
	if err != nil {
		return fmt.Errorf("list existing members in Google Directory group: %w", err)
	}

	consoleUserMap := make(map[string]*db.User)
	localMembers := helpers.DomainUsers(input.TeamMembers, r.domain)

	membersToRemove := remoteOnlyMembers(membersAccordingToGoogle.Members, localMembers)
	for _, member := range membersToRemove {
		remoteMemberEmail := strings.ToLower(member.Email)
		err = r.adminService.Members.Delete(grp.Id, member.Id).Do()
		if err != nil {
			log.Warnf("delete member %q from Google Directory group %q: %s", remoteMemberEmail, grp.Email, err)
			continue
		}

		if _, exists := consoleUserMap[remoteMemberEmail]; !exists {
			user, err := r.database.GetUserByEmail(ctx, remoteMemberEmail)
			if err != nil {
				return err
			}
			consoleUserMap[remoteMemberEmail] = user
		}

		fields := auditlogger.Fields{
			Action:          sqlc.AuditActionGoogleWorkspaceAdminDeleteMember,
			CorrelationID:   input.CorrelationID,
			TargetTeamSlug:  &input.Team.Slug,
			TargetUserEmail: &remoteMemberEmail,
		}
		r.auditLogger.Logf(ctx, fields, "deleted member %q from Google Directory group %q", member.Email, grp.Email)
	}

	membersToAdd := localOnlyMembers(membersAccordingToGoogle.Members, localMembers)
	for _, user := range membersToAdd {
		member := &admin_directory_v1.Member{
			Email: user.Email,
		}
		_, err = r.adminService.Members.Insert(grp.Id, member).Do()
		if err != nil {
			log.Warnf("add member %q to Google Directory group %q: %s", member.Email, grp.Email, err)
			continue
		}
		fields := auditlogger.Fields{
			Action:          sqlc.AuditActionGoogleWorkspaceAdminAddMember,
			CorrelationID:   input.CorrelationID,
			TargetTeamSlug:  &input.Team.Slug,
			TargetUserEmail: &user.Email,
		}
		r.auditLogger.Logf(ctx, fields, "added member %q to Google Directory group %q", member.Email, grp.Email)
	}

	return nil
}

func (r *googleWorkspaceAdminReconciler) addToGKESecurityGroup(ctx context.Context, grp *admin_directory_v1.Group, input reconcilers.Input) error {
	const groupPrefix = "gke-security-groups@"
	groupKey := groupPrefix + r.domain

	member := &admin_directory_v1.Member{
		Email: grp.Email,
	}

	_, err := r.adminService.Members.Insert(groupKey, member).Do()
	if err != nil {
		googleError, ok := err.(*googleapi.Error)
		if ok && googleError.Code == http.StatusConflict {
			return nil
		}
		return fmt.Errorf("add group %q to GKE security group %q: %s", member.Email, groupKey, err)
	}

	fields := auditlogger.Fields{
		Action:         sqlc.AuditActionGoogleWorkspaceAdminAddToGkeSecurityGroup,
		CorrelationID:  input.CorrelationID,
		TargetTeamSlug: &input.Team.Slug,
	}
	r.auditLogger.Logf(ctx, fields, "added group %q to GKE security group %q", member.Email, groupKey)

	return nil
}

// remoteOnlyMembers Given a list of Google group members and a list of Console users, return Google group members not
// present in Console user list.
func remoteOnlyMembers(googleGroupMembers []*admin_directory_v1.Member, consoleUsers []*db.User) []*admin_directory_v1.Member {
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
func localOnlyMembers(googleGroupMembers []*admin_directory_v1.Member, consoleUsers []*db.User) []*db.User {
	localUserMap := make(map[string]*db.User)
	for _, user := range consoleUsers {
		localUserMap[user.Email] = user
	}
	for _, member := range googleGroupMembers {
		delete(localUserMap, strings.ToLower(member.Email))
	}
	consoleUsers = make([]*db.User, 0, len(localUserMap))
	for _, user := range localUserMap {
		consoleUsers = append(consoleUsers, user)
	}
	return consoleUsers
}
