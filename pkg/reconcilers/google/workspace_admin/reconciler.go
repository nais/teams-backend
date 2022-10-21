package google_workspace_admin_reconciler

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
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
	Name = sqlc.ReconcilerNameGoogleWorkspaceAdmin
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

func (r *googleWorkspaceAdminReconciler) Name() sqlc.ReconcilerName {
	return Name
}

func (r *googleWorkspaceAdminReconciler) Reconcile(ctx context.Context, input reconcilers.Input) error {
	state := &reconcilers.GoogleWorkspaceState{}
	err := r.database.LoadReconcilerStateForTeam(ctx, r.Name(), input.Team.ID, state)
	if err != nil {
		return fmt.Errorf("unable to load system state for team %q in system %q: %w", input.Team.Slug, r.Name(), err)
	}

	grp, err := r.getOrCreateGroup(ctx, state, input)
	if err != nil {
		return fmt.Errorf("unable to get or create a Google Workspace group for team %q in system %q: %w", input.Team.Slug, r.Name(), err)
	}

	err = r.syncGroupInfo(ctx, input.Team, grp)
	if err != nil {
		return err
	}

	err = r.database.SetReconcilerStateForTeam(ctx, r.Name(), input.Team.ID, reconcilers.GoogleWorkspaceState{GroupEmail: &grp.Email})
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
		return r.adminService.Groups.Get(*state.GroupEmail).Do()
	}

	groupKey := reconcilers.TeamNamePrefix + input.Team.Slug
	email := fmt.Sprintf("%s@%s", groupKey, r.domain)
	newGroup := &admin_directory_v1.Group{
		Email:       email,
		Name:        string(input.Team.Slug),
		Description: input.Team.Purpose,
	}
	group, err := r.adminService.Groups.Insert(newGroup).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to create Google Directory group: %w", err)
	}

	targets := []auditlogger.Target{
		auditlogger.TeamTarget(input.Team.Slug),
	}
	fields := auditlogger.Fields{
		Action:        sqlc.AuditActionGoogleWorkspaceAdminCreate,
		CorrelationID: input.CorrelationID,
	}
	r.auditLogger.Logf(ctx, targets, fields, "created Google Directory group %q", group.Email)

	return group, nil
}

// getGoogleGroupMembers Get all members of a Google Workspace Group
func getGoogleGroupMembers(ctx context.Context, service *admin_directory_v1.MembersService, groupID string) ([]*admin_directory_v1.Member, error) {
	members := make([]*admin_directory_v1.Member, 0)
	callback := func(fragments *admin_directory_v1.Members) error {
		members = append(members, fragments.Members...)
		return nil
	}
	err := service.
		List(groupID).
		Context(ctx).
		Pages(ctx, callback)
	if err != nil {
		return nil, fmt.Errorf("list existing members in Google Directory group: %w", err)
	}

	return members, nil
}

func (r *googleWorkspaceAdminReconciler) connectUsers(ctx context.Context, grp *admin_directory_v1.Group, input reconcilers.Input) error {
	membersAccordingToGoogle, err := getGoogleGroupMembers(ctx, r.adminService.Members, grp.Id)
	if err != nil {
		return fmt.Errorf("list existing members in Google Directory group: %w", err)
	}

	consoleUserMap := make(map[string]*db.User)
	membersToRemove := remoteOnlyMembers(membersAccordingToGoogle, input.TeamMembers)
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

		targets := []auditlogger.Target{
			auditlogger.TeamTarget(input.Team.Slug),
			auditlogger.UserTarget(remoteMemberEmail),
		}
		fields := auditlogger.Fields{
			Action:        sqlc.AuditActionGoogleWorkspaceAdminDeleteMember,
			CorrelationID: input.CorrelationID,
		}
		r.auditLogger.Logf(ctx, targets, fields, "deleted member %q from Google Directory group %q", member.Email, grp.Email)
	}

	membersToAdd := localOnlyMembers(membersAccordingToGoogle, input.TeamMembers)
	for _, user := range membersToAdd {
		member := &admin_directory_v1.Member{
			Email: user.Email,
		}
		_, err = r.adminService.Members.Insert(grp.Id, member).Do()
		if err != nil {
			log.Warnf("add member %q to Google Directory group %q: %s", member.Email, grp.Email, err)
			continue
		}
		targets := []auditlogger.Target{
			auditlogger.TeamTarget(input.Team.Slug),
			auditlogger.UserTarget(user.Email),
		}
		fields := auditlogger.Fields{
			Action:        sqlc.AuditActionGoogleWorkspaceAdminAddMember,
			CorrelationID: input.CorrelationID,
		}
		r.auditLogger.Logf(ctx, targets, fields, "added member %q to Google Directory group %q", member.Email, grp.Email)
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

	targets := []auditlogger.Target{
		auditlogger.TeamTarget(input.Team.Slug),
	}
	fields := auditlogger.Fields{
		Action:        sqlc.AuditActionGoogleWorkspaceAdminAddToGkeSecurityGroup,
		CorrelationID: input.CorrelationID,
	}
	r.auditLogger.Logf(ctx, targets, fields, "added group %q to GKE security group %q", member.Email, groupKey)

	return nil
}

func (r *googleWorkspaceAdminReconciler) syncGroupInfo(ctx context.Context, team db.Team, group *admin_directory_v1.Group) error {
	if team.Purpose == group.Description {
		return nil
	}

	group.Description = team.Purpose
	group.ForceSendFields = []string{"Description"}
	_, err := r.adminService.Groups.Patch(group.Id, group).Context(ctx).Do()
	if err != nil {
		return err
	}

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

// localOnlyMembers Given a list of Google group members and a list of users, return users not present in members
// directory.
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
