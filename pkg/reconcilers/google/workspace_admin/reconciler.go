package google_workspace_admin_reconciler

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/nais/teams-backend/pkg/types"

	"github.com/google/uuid"
	"github.com/nais/teams-backend/pkg/auditlogger"
	"github.com/nais/teams-backend/pkg/config"
	"github.com/nais/teams-backend/pkg/db"
	"github.com/nais/teams-backend/pkg/google_token_source"
	"github.com/nais/teams-backend/pkg/logger"
	"github.com/nais/teams-backend/pkg/metrics"
	"github.com/nais/teams-backend/pkg/reconcilers"
	"github.com/nais/teams-backend/pkg/slug"
	"github.com/nais/teams-backend/pkg/sqlc"
	admin_directory_v1 "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

type googleWorkspaceAdminReconciler struct {
	database     db.Database
	auditLogger  auditlogger.AuditLogger
	domain       string
	adminService *admin_directory_v1.Service
	log          logger.Logger
}

const (
	Name              = sqlc.ReconcilerNameGoogleWorkspaceAdmin
	metricsSystemName = "google-admin"
)

func New(database db.Database, auditLogger auditlogger.AuditLogger, domain string, adminService *admin_directory_v1.Service, log logger.Logger) *googleWorkspaceAdminReconciler {
	return &googleWorkspaceAdminReconciler{
		database:     database,
		auditLogger:  auditLogger,
		domain:       domain,
		adminService: adminService,
		log:          log.WithComponent(types.ComponentNameGoogleWorkspaceAdmin),
	}
}

func NewFromConfig(ctx context.Context, database db.Database, cfg *config.Config, log logger.Logger) (reconcilers.Reconciler, error) {
	builder, err := google_token_source.NewFromConfig(cfg)
	if err != nil {
		return nil, err
	}

	ts, err := builder.Admin(ctx)
	if err != nil {
		return nil, fmt.Errorf("get delegated token source: %w", err)
	}

	srv, err := admin_directory_v1.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("retrieve directory client: %w", err)
	}

	return New(database, auditlogger.New(database, types.ComponentNameGoogleWorkspaceAdmin, log), cfg.TenantDomain, srv, log), nil
}

func (r *googleWorkspaceAdminReconciler) Name() sqlc.ReconcilerName {
	return Name
}

func (r *googleWorkspaceAdminReconciler) Reconcile(ctx context.Context, input reconcilers.Input) error {
	state := &reconcilers.GoogleWorkspaceState{}
	err := r.database.LoadReconcilerStateForTeam(ctx, r.Name(), input.Team.Slug, state)
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

	err = r.database.SetReconcilerStateForTeam(ctx, r.Name(), input.Team.Slug, reconcilers.GoogleWorkspaceState{GroupEmail: &grp.Email})
	if err != nil {
		r.log.WithError(err).Error("persiste system state")
	}

	err = r.connectUsers(ctx, grp, input)
	if err != nil {
		return fmt.Errorf("add members to group: %w", err)
	}

	return r.addToGKESecurityGroup(ctx, grp, input)
}

func (r *googleWorkspaceAdminReconciler) Delete(ctx context.Context, teamSlug slug.Slug, correlationID uuid.UUID) error {
	state := &reconcilers.GoogleWorkspaceState{}
	err := r.database.LoadReconcilerStateForTeam(ctx, r.Name(), teamSlug, state)
	if err != nil {
		return fmt.Errorf("load reconciler state for team %q in reconciler %q: %w", teamSlug, r.Name(), err)
	}

	if state.GroupEmail == nil {
		r.log.Warnf("missing group email in reconciler state for team %q in reconciler %q, assume already deleted", teamSlug, r.Name())
		return r.database.RemoveReconcilerStateForTeam(ctx, r.Name(), teamSlug)
	}

	grpEmail := *state.GroupEmail

	err = r.adminService.Groups.Delete(grpEmail).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("delete Google directory group with email %q for team %q: %w", grpEmail, teamSlug, err)
	}

	targets := []auditlogger.Target{
		auditlogger.TeamTarget(teamSlug),
	}
	fields := auditlogger.Fields{
		Action:        types.AuditActionGoogleWorkspaceAdminDelete,
		CorrelationID: correlationID,
	}
	r.auditLogger.Logf(ctx, targets, fields, "Delete Google directory group with email %q", grpEmail)

	return r.database.RemoveReconcilerStateForTeam(ctx, r.Name(), teamSlug)
}

func (r *googleWorkspaceAdminReconciler) getOrCreateGroup(ctx context.Context, state *reconcilers.GoogleWorkspaceState, input reconcilers.Input) (*admin_directory_v1.Group, error) {
	if state.GroupEmail != nil {
		grp, err := r.adminService.Groups.Get(*state.GroupEmail).Do()
		if err != nil {
			metrics.IncExternalCallsByError(metricsSystemName, err)
			return nil, err
		}
		metrics.IncExternalCalls(metricsSystemName, grp.HTTPStatusCode)
		return grp, err
	}

	groupKey := reconcilers.TeamNamePrefix + input.Team.Slug
	email := fmt.Sprintf("%s@%s", groupKey, r.domain)
	newGroup := &admin_directory_v1.Group{
		Email:       email,
		Name:        string(groupKey),
		Description: input.Team.Purpose,
	}
	group, err := r.adminService.Groups.Insert(newGroup).Do()
	if err != nil {
		metrics.IncExternalCallsByError(metricsSystemName, err)
		return nil, fmt.Errorf("unable to create Google Directory group: %w", err)
	}
	metrics.IncExternalCalls(metricsSystemName, group.HTTPStatusCode)

	targets := []auditlogger.Target{
		auditlogger.TeamTarget(input.Team.Slug),
	}
	fields := auditlogger.Fields{
		Action:        types.AuditActionGoogleWorkspaceAdminCreate,
		CorrelationID: input.CorrelationID,
	}
	r.auditLogger.Logf(ctx, targets, fields, "Created Google Directory group %q", group.Email)

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
	metrics.IncExternalCallsByError(metricsSystemName, err)
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

	teamsBackendUserMap := make(map[string]*db.User)
	membersToRemove := remoteOnlyMembers(membersAccordingToGoogle, input.TeamMembers)
	for _, member := range membersToRemove {
		remoteMemberEmail := strings.ToLower(member.Email)
		err = r.adminService.Members.Delete(grp.Id, member.Id).Do()
		metrics.IncExternalCallsByError(metricsSystemName, err)
		if err != nil {
			r.log.WithError(err).Warnf("delete member %q from Google Directory group %q", remoteMemberEmail, grp.Email)
			continue
		}

		if _, exists := teamsBackendUserMap[remoteMemberEmail]; !exists {
			user, err := r.database.GetUserByEmail(ctx, remoteMemberEmail)
			if err == sql.ErrNoRows {
				r.log.Warnf("Remote Google user %q not found in local database", remoteMemberEmail)
			} else if err != nil {
				return err
			} else {
				teamsBackendUserMap[remoteMemberEmail] = user
			}
		}

		targets := []auditlogger.Target{
			auditlogger.TeamTarget(input.Team.Slug),
			auditlogger.UserTarget(remoteMemberEmail),
		}
		fields := auditlogger.Fields{
			Action:        types.AuditActionGoogleWorkspaceAdminDeleteMember,
			CorrelationID: input.CorrelationID,
		}
		r.auditLogger.Logf(ctx, targets, fields, "Deleted member %q from Google Directory group %q", member.Email, grp.Email)
	}

	membersToAdd := localOnlyMembers(membersAccordingToGoogle, input.TeamMembers)
	for _, user := range membersToAdd {
		member := &admin_directory_v1.Member{
			Email: user.Email,
		}
		operation, err := r.adminService.Members.Insert(grp.Id, member).Do()
		if err != nil {
			// Getting an error from this call is not really an "error" per se.
			// Thus, it is overridden in the metrics to avoid pumping up the error rate.
			metrics.IncExternalCalls(metricsSystemName, 0)
			r.log.WithError(err).Warnf("add member %q to Google Directory group %q", member.Email, grp.Email)
			continue
		}
		metrics.IncExternalCalls(metricsSystemName, operation.HTTPStatusCode)

		targets := []auditlogger.Target{
			auditlogger.TeamTarget(input.Team.Slug),
			auditlogger.UserTarget(user.Email),
		}
		fields := auditlogger.Fields{
			Action:        types.AuditActionGoogleWorkspaceAdminAddMember,
			CorrelationID: input.CorrelationID,
		}
		r.auditLogger.Logf(ctx, targets, fields, "Added member %q to Google Directory group %q", member.Email, grp.Email)
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
		metrics.IncExternalCallsByError(metricsSystemName, err)
		if ok && googleError.Code == http.StatusConflict {
			return nil
		}
		return fmt.Errorf("add group %q to GKE security group %q: %s", member.Email, groupKey, err)
	}

	targets := []auditlogger.Target{
		auditlogger.TeamTarget(input.Team.Slug),
	}
	fields := auditlogger.Fields{
		Action:        types.AuditActionGoogleWorkspaceAdminAddToGkeSecurityGroup,
		CorrelationID: input.CorrelationID,
	}
	r.auditLogger.Logf(ctx, targets, fields, "Added group %q to GKE security group %q", member.Email, groupKey)

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

// remoteOnlyMembers Given a list of Google group members and a list of teams-backend users, return Google group members not
// present in teams-backend user list.
func remoteOnlyMembers(googleGroupMembers []*admin_directory_v1.Member, teamsBackendUsers []*db.User) []*admin_directory_v1.Member {
	googleGroupMemberMap := make(map[string]*admin_directory_v1.Member)
	for _, member := range googleGroupMembers {
		googleGroupMemberMap[strings.ToLower(member.Email)] = member
	}
	for _, user := range teamsBackendUsers {
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
func localOnlyMembers(googleGroupMembers []*admin_directory_v1.Member, teamsBackendUsers []*db.User) []*db.User {
	localUserMap := make(map[string]*db.User)
	for _, user := range teamsBackendUsers {
		localUserMap[user.Email] = user
	}
	for _, member := range googleGroupMembers {
		delete(localUserMap, strings.ToLower(member.Email))
	}
	teamsBackendUsers = make([]*db.User, 0, len(localUserMap))
	for _, user := range localUserMap {
		teamsBackendUsers = append(teamsBackendUsers, user)
	}
	return teamsBackendUsers
}
