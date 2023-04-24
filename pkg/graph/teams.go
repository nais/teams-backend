package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/graph/apierror"
	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/graph/model"
	"github.com/nais/console/pkg/reconcilers"
	google_gcp_reconciler "github.com/nais/console/pkg/reconcilers/google/gcp"
	"github.com/nais/console/pkg/roles"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
)

// CreateTeam is the resolver for the createTeam field.
func (r *mutationResolver) CreateTeam(ctx context.Context, input model.CreateTeamInput) (*db.Team, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireGlobalAuthorization(actor, roles.AuthorizationTeamsCreate)
	if err != nil {
		return nil, err
	}

	input = input.Sanitize()

	err = input.Validate()
	if err != nil {
		return nil, err
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("create log correlation ID: %w", err)
	}

	var team *db.Team
	err = r.database.Transaction(ctx, func(ctx context.Context, dbtx db.Database) error {
		team, err = dbtx.CreateTeam(ctx, *input.Slug, input.Purpose, input.SlackChannel)
		if err != nil {
			return err
		}

		if actor.User.IsServiceAccount() {
			return dbtx.AssignTeamRoleToServiceAccount(ctx, actor.User.GetID(), sqlc.RoleNameTeamowner, *input.Slug)
		}

		return dbtx.SetTeamMemberRole(ctx, actor.User.GetID(), team.Slug, sqlc.RoleNameTeamowner)
	})
	if err != nil {
		return nil, err
	}

	targets := []auditlogger.Target{
		auditlogger.TeamTarget(team.Slug),
	}
	fields := auditlogger.Fields{
		Action:        sqlc.AuditActionGraphqlApiTeamCreate,
		CorrelationID: correlationID,
		Actor:         actor,
	}
	r.auditLogger.Logf(ctx, r.database, targets, fields, "Team created")

	r.reconcileTeam(ctx, correlationID, team.Slug)

	return team, nil
}

// UpdateTeam is the resolver for the updateTeam field.
func (r *mutationResolver) UpdateTeam(ctx context.Context, slug *slug.Slug, input model.UpdateTeamInput) (*db.Team, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireTeamAuthorization(actor, roles.AuthorizationTeamsUpdate, *slug)
	if err != nil {
		return nil, err
	}

	team, err := r.getTeamBySlug(ctx, *slug)
	if err != nil {
		return nil, apierror.ErrTeamNotExist
	}

	input = input.Sanitize()
	err = input.Validate(r.gcpEnvironments)
	if err != nil {
		return nil, err
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("create log correlation ID: %w", err)
	}

	err = r.database.Transaction(ctx, func(ctx context.Context, dbtx db.Database) error {
		team, err = dbtx.UpdateTeam(ctx, team.Slug, input.Purpose, input.SlackChannel)
		if err != nil {
			return err
		}

		if len(input.SlackAlertsChannels) > 0 {
			for _, slackAlertsChannel := range input.SlackAlertsChannels {
				var err error
				if slackAlertsChannel.ChannelName == nil {
					err = dbtx.RemoveSlackAlertsChannel(ctx, team.Slug, slackAlertsChannel.Environment)
				} else {
					err = dbtx.SetSlackAlertsChannel(ctx, team.Slug, slackAlertsChannel.Environment, *slackAlertsChannel.ChannelName)
				}
				if err != nil {
					return err
				}
			}
		}

		targets := []auditlogger.Target{
			auditlogger.TeamTarget(team.Slug),
		}
		fields := auditlogger.Fields{
			Action:        sqlc.AuditActionGraphqlApiTeamUpdate,
			CorrelationID: correlationID,
			Actor:         actor,
		}

		return r.auditLogger.Logf(ctx, dbtx, targets, fields, "Team configuration saved")
	})
	if err != nil {
		return nil, err
	}

	r.reconcileTeam(ctx, correlationID, team.Slug)

	return team, nil
}

// RemoveUsersFromTeam is the resolver for the removeUsersFromTeam field.
func (r *mutationResolver) RemoveUsersFromTeam(ctx context.Context, slug *slug.Slug, userIds []*uuid.UUID) (*db.Team, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireTeamAuthorization(actor, roles.AuthorizationTeamsUpdate, *slug)
	if err != nil {
		return nil, err
	}

	team, err := r.getTeamBySlug(ctx, *slug)
	if err != nil {
		return nil, err
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("create log correlation ID: %w", err)
	}

	err = r.database.Transaction(ctx, func(ctx context.Context, dbtx db.Database) error {
		members, err := dbtx.GetTeamMembers(ctx, team.Slug)
		if err != nil {
			return fmt.Errorf("get team members of %q: %w", *slug, err)
		}

		memberFromUserID := func(userId uuid.UUID) *db.User {
			for _, m := range members {
				if m.ID == userId {
					return m
				}
			}
			return nil
		}

		for _, userID := range userIds {
			member := memberFromUserID(*userID)
			if member == nil {
				return apierror.Errorf("The user %q is not a member of team %q.", *userID, *slug)
			}

			err = dbtx.RemoveUserFromTeam(ctx, *userID, team.Slug)
			if err != nil {
				return err
			}

			targets := []auditlogger.Target{
				auditlogger.TeamTarget(team.Slug),
				auditlogger.UserTarget(member.Email),
			}
			fields := auditlogger.Fields{
				Action:        sqlc.AuditActionGraphqlApiTeamRemoveMember,
				CorrelationID: correlationID,
				Actor:         actor,
			}
			r.auditLogger.Logf(ctx, dbtx, targets, fields, "Removed user: %q", member.Email)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	r.reconcileTeam(ctx, correlationID, team.Slug)

	return team, nil
}

// SynchronizeTeam is the resolver for the synchronizeTeam field.
func (r *mutationResolver) SynchronizeTeam(ctx context.Context, slug *slug.Slug) (*model.TeamSync, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireTeamAuthorization(actor, roles.AuthorizationTeamsSynchronize, *slug)
	if err != nil {
		return nil, err
	}

	team, err := r.getTeamBySlug(ctx, *slug)
	if err != nil {
		return nil, err
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("create log correlation ID: %w", err)
	}

	targets := []auditlogger.Target{
		auditlogger.TeamTarget(team.Slug),
	}
	fields := auditlogger.Fields{
		Action:        sqlc.AuditActionGraphqlApiTeamSync,
		CorrelationID: correlationID,
		Actor:         actor,
	}
	r.auditLogger.Logf(ctx, r.database, targets, fields, "Manually scheduled for synchronization")

	r.reconcileTeam(ctx, correlationID, team.Slug)

	return &model.TeamSync{
		CorrelationID: &correlationID,
	}, nil
}

// SynchronizeAllTeams is the resolver for the synchronizeAllTeams field.
func (r *mutationResolver) SynchronizeAllTeams(ctx context.Context) (*model.TeamSync, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireGlobalAuthorization(actor, roles.AuthorizationTeamsSynchronize)
	if err != nil {
		return nil, err
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("create log correlation ID: %w", err)
	}

	teams, err := r.teamSyncHandler.ScheduleAllTeams(ctx, correlationID)
	if err != nil {
		return nil, err
	}

	targets := make([]auditlogger.Target, 0, len(teams))
	for _, entry := range teams {
		targets = append(targets, auditlogger.TeamTarget(entry.Team.Slug))
	}
	fields := auditlogger.Fields{
		Action:        sqlc.AuditActionGraphqlApiTeamSync,
		Actor:         actor,
		CorrelationID: correlationID,
	}
	r.auditLogger.Logf(ctx, r.database, targets, fields, "Manually scheduled for synchronization")

	return &model.TeamSync{
		CorrelationID: &correlationID,
	}, nil
}

// AddTeamMembers is the resolver for the addTeamMembers field.
func (r *mutationResolver) AddTeamMembers(ctx context.Context, slug *slug.Slug, userIds []*uuid.UUID) (*db.Team, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireTeamAuthorization(actor, roles.AuthorizationTeamsUpdate, *slug)
	if err != nil {
		return nil, err
	}

	team, err := r.getTeamBySlug(ctx, *slug)
	if err != nil {
		return nil, err
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("create log correlation ID: %w", err)
	}

	err = r.database.Transaction(ctx, func(ctx context.Context, dbtx db.Database) error {
		for _, userID := range userIds {
			user, err := dbtx.GetUserByID(ctx, *userID)
			if err != nil {
				return err
			}

			err = dbtx.SetTeamMemberRole(ctx, *userID, team.Slug, sqlc.RoleNameTeammember)
			if err != nil {
				return err
			}

			targets := []auditlogger.Target{
				auditlogger.TeamTarget(team.Slug),
				auditlogger.UserTarget(user.Email),
			}
			fields := auditlogger.Fields{
				Action:        sqlc.AuditActionGraphqlApiTeamAddMember,
				CorrelationID: correlationID,
				Actor:         actor,
			}
			r.auditLogger.Logf(ctx, dbtx, targets, fields, "Add team member: %q", user.Email)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	r.reconcileTeam(ctx, correlationID, team.Slug)

	return team, nil
}

// AddTeamOwners is the resolver for the addTeamOwners field.
func (r *mutationResolver) AddTeamOwners(ctx context.Context, slug *slug.Slug, userIds []*uuid.UUID) (*db.Team, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireTeamAuthorization(actor, roles.AuthorizationTeamsUpdate, *slug)
	if err != nil {
		return nil, err
	}

	team, err := r.getTeamBySlug(ctx, *slug)
	if err != nil {
		return nil, err
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("create log correlation ID: %w", err)
	}

	err = r.database.Transaction(ctx, func(ctx context.Context, dbtx db.Database) error {
		for _, userID := range userIds {
			user, err := dbtx.GetUserByID(ctx, *userID)
			if err != nil {
				return err
			}

			err = dbtx.SetTeamMemberRole(ctx, *userID, team.Slug, sqlc.RoleNameTeamowner)
			if err != nil {
				return err
			}

			targets := []auditlogger.Target{
				auditlogger.TeamTarget(team.Slug),
				auditlogger.UserTarget(user.Email),
			}
			fields := auditlogger.Fields{
				Action:        sqlc.AuditActionGraphqlApiTeamAddOwner,
				CorrelationID: correlationID,
				Actor:         actor,
			}
			r.auditLogger.Logf(ctx, dbtx, targets, fields, "Add team owner: %q", user.Email)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	r.reconcileTeam(ctx, correlationID, team.Slug)

	return team, nil
}

// SetTeamMemberRole is the resolver for the setTeamMemberRole field.
func (r *mutationResolver) SetTeamMemberRole(ctx context.Context, slug *slug.Slug, userID *uuid.UUID, role model.TeamRole) (*db.Team, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireTeamAuthorization(actor, roles.AuthorizationTeamsUpdate, *slug)
	if err != nil {
		return nil, err
	}

	team, err := r.getTeamBySlug(ctx, *slug)
	if err != nil {
		return nil, err
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("create log correlation ID: %w", err)
	}

	members, err := r.database.GetTeamMembers(ctx, team.Slug)
	if err != nil {
		return nil, fmt.Errorf("get team members: %w", err)
	}

	var member *db.User = nil
	for _, m := range members {
		if m.ID == *userID {
			member = m
			break
		}
	}
	if member == nil {
		return nil, fmt.Errorf("user %q not in team %q", *userID, *slug)
	}

	desiredRole, err := sqlcRoleFromTeamRole(role)
	if err != nil {
		return nil, err
	}

	err = r.database.SetTeamMemberRole(ctx, *userID, team.Slug, desiredRole)
	if err != nil {
		return nil, err
	}

	targets := []auditlogger.Target{
		auditlogger.TeamTarget(team.Slug),
		auditlogger.UserTarget(member.Email),
	}
	fields := auditlogger.Fields{
		Action:        sqlc.AuditActionGraphqlApiTeamSetMemberRole,
		CorrelationID: correlationID,
		Actor:         actor,
	}

	r.auditLogger.Logf(ctx, r.database, targets, fields, "Assign %q to %s", desiredRole, member.Email)

	r.reconcileTeam(ctx, correlationID, team.Slug)

	return team, nil
}

// RequestTeamDeletion is the resolver for the requestTeamDeletion field.
func (r *mutationResolver) RequestTeamDeletion(ctx context.Context, slug *slug.Slug) (*db.TeamDeleteKey, error) {
	actor := authz.ActorFromContext(ctx)
	if actor.User.IsServiceAccount() {
		return nil, apierror.Errorf("Service accounts are not allowed to request a team deletion.")
	}

	err := authz.RequireTeamAuthorization(actor, roles.AuthorizationTeamsUpdate, *slug)
	if err != nil {
		return nil, err
	}

	team, err := r.getTeamBySlug(ctx, *slug)
	if err != nil {
		return nil, err
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("create log correlation ID: %w", err)
	}

	deleteKey, err := r.database.CreateTeamDeleteKey(ctx, *slug, actor.User.GetID())
	if err != nil {
		return nil, fmt.Errorf("create team delete key: %w", err)
	}

	targets := []auditlogger.Target{
		auditlogger.TeamTarget(team.Slug),
	}
	fields := auditlogger.Fields{
		Action:        sqlc.AuditActionGraphqlApiTeamsRequestDelete,
		Actor:         actor,
		CorrelationID: correlationID,
	}
	r.auditLogger.Logf(ctx, r.database, targets, fields, "Request team deletion")

	return deleteKey, nil
}

// ConfirmTeamDeletion is the resolver for the confirmTeamDeletion field.
func (r *mutationResolver) ConfirmTeamDeletion(ctx context.Context, key *uuid.UUID) (*uuid.UUID, error) {
	deleteKey, err := r.database.GetTeamDeleteKey(ctx, *key)
	if err != nil {
		return nil, apierror.Errorf("Unknown deletion key: %q", key)
	}

	actor := authz.ActorFromContext(ctx)
	if actor.User.IsServiceAccount() {
		return nil, apierror.Errorf("Service accounts are not allowed to confirm a team deletion.")
	}
	err = authz.RequireTeamAuthorization(actor, roles.AuthorizationTeamsUpdate, deleteKey.TeamSlug)
	if err != nil {
		return nil, err
	}

	if actor.User.GetID() == deleteKey.CreatedBy {
		return nil, apierror.Errorf("You cannot confirm your own delete key.")
	}

	if deleteKey.ConfirmedAt.Valid {
		return nil, apierror.Errorf("Key has already been confirmed, team is currently being deleted.")
	}

	if deleteKey.HasExpired() {
		return nil, apierror.Errorf("Team delete key has expired, you need to request a new key.")
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("create log correlation ID: %w", err)
	}

	err = r.database.ConfirmTeamDeleteKey(ctx, *key)
	if err != nil {
		return nil, fmt.Errorf("confirm team delete key: %w", err)
	}

	go r.teamSyncHandler.DeleteTeam(deleteKey.TeamSlug, correlationID)

	targets := []auditlogger.Target{
		auditlogger.TeamTarget(deleteKey.TeamSlug),
	}
	fields := auditlogger.Fields{
		Action:        sqlc.AuditActionGraphqlApiTeamsDelete,
		Actor:         actor,
		CorrelationID: correlationID,
	}
	r.auditLogger.Logf(ctx, r.database, targets, fields, "Delete team")

	return &correlationID, nil
}

// Teams is the resolver for the teams field.
func (r *queryResolver) Teams(ctx context.Context) ([]*db.Team, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireGlobalAuthorization(actor, roles.AuthorizationTeamsList)
	if err != nil {
		return nil, err
	}

	return r.database.GetTeams(ctx)
}

// Team is the resolver for the team field.
func (r *queryResolver) Team(ctx context.Context, slug *slug.Slug) (*db.Team, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireTeamAuthorization(actor, roles.AuthorizationTeamsRead, *slug)
	if err != nil {
		return nil, err
	}

	team, err := r.getTeamBySlug(ctx, *slug)
	if err != nil {
		return nil, err
	}

	return team, nil
}

// DeployKey is the resolver for the deployKey field.
func (r *queryResolver) DeployKey(ctx context.Context, slug *slug.Slug) (string, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireTeamAuthorization(actor, roles.AuthorizationDeployKeyView, *slug)
	if err != nil {
		return "", err
	}

	if r.deployProxy == nil {
		return "", fmt.Errorf("deploy proxy is not configured")
	}

	deployKey, err := r.deployProxy.GetApiKey(ctx, *slug)
	if err != nil {
		return "", err
	}

	return deployKey, nil
}

// TeamDeleteKey is the resolver for the teamDeleteKey field.
func (r *queryResolver) TeamDeleteKey(ctx context.Context, key *uuid.UUID) (*db.TeamDeleteKey, error) {
	deleteKey, err := r.database.GetTeamDeleteKey(ctx, *key)
	if err != nil {
		return nil, apierror.Errorf("Unknown deletion key: %q", key)
	}

	actor := authz.ActorFromContext(ctx)
	if actor.User.IsServiceAccount() {
		return nil, apierror.Errorf("Service accounts are not allowed to get team delete keys.")
	}
	err = authz.RequireTeamAuthorization(actor, roles.AuthorizationTeamsUpdate, deleteKey.TeamSlug)
	if err != nil {
		return nil, err
	}

	return deleteKey, nil
}

// AuditLogs is the resolver for the auditLogs field.
func (r *teamResolver) AuditLogs(ctx context.Context, obj *db.Team) ([]*db.AuditLog, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireTeamAuthorization(actor, roles.AuthorizationAuditLogsRead, obj.Slug)
	if err != nil {
		return nil, err
	}

	return r.database.GetAuditLogsForTeam(ctx, obj.Slug)
}

// Members is the resolver for the members field.
func (r *teamResolver) Members(ctx context.Context, obj *db.Team) ([]*model.TeamMember, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireGlobalAuthorization(actor, roles.AuthorizationUsersList)
	if err != nil {
		return nil, err
	}

	users, err := r.database.GetTeamMembers(ctx, obj.Slug)
	if err != nil {
		return nil, err
	}

	members := make([]*model.TeamMember, len(users))
	for idx, user := range users {
		isOwner, err := r.database.UserIsTeamOwner(ctx, user.ID, obj.Slug)
		if err != nil {
			return nil, err
		}

		role := model.TeamRoleMember
		if isOwner {
			role = model.TeamRoleOwner
		}

		members[idx] = &model.TeamMember{
			User: user,
			Role: role,
		}
	}
	return members, nil
}

// SyncErrors is the resolver for the syncErrors field.
func (r *teamResolver) SyncErrors(ctx context.Context, obj *db.Team) ([]*model.SyncError, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireTeamAuthorization(actor, roles.AuthorizationTeamsRead, obj.Slug)
	if err != nil {
		return nil, err
	}

	rows, err := r.database.GetTeamReconcilerErrors(ctx, obj.Slug)
	if err != nil {
		return nil, err
	}

	syncErrors := make([]*model.SyncError, 0)
	for _, row := range rows {
		syncErrors = append(syncErrors, &model.SyncError{
			CreatedAt:  row.CreatedAt,
			Reconciler: row.Reconciler,
			Error:      row.ErrorMessage,
		})
	}

	return syncErrors, nil
}

// LastSuccessfulSync is the resolver for the lastSuccessfulSync field.
func (r *teamResolver) LastSuccessfulSync(ctx context.Context, obj *db.Team) (*time.Time, error) {
	if !obj.LastSuccessfulSync.Valid {
		return nil, nil
	}
	return &obj.LastSuccessfulSync.Time, nil
}

// ReconcilerState is the resolver for the reconcilerState field.
func (r *teamResolver) ReconcilerState(ctx context.Context, obj *db.Team) (*model.ReconcilerState, error) {
	gitHubState := &reconcilers.GitHubState{}
	googleWorkspaceState := &reconcilers.GoogleWorkspaceState{}
	gcpProjectState := &reconcilers.GoogleGcpProjectState{
		Projects: make(map[string]reconcilers.GoogleGcpEnvironmentProject),
	}
	naisNamespaceState := &reconcilers.NaisNamespaceState{
		Namespaces: make(map[string]slug.Slug),
	}
	azureADState := &reconcilers.AzureState{}
	naisDeployKeyState := &reconcilers.NaisDeployKeyState{}
	googleGarState := &reconcilers.GoogleGarState{}

	queriedFields := GetQueriedFields(ctx)

	if _, inQuery := queriedFields["gitHubTeamSlug"]; inQuery {
		err := r.database.LoadReconcilerStateForTeam(ctx, sqlc.ReconcilerNameGithubTeam, obj.Slug, gitHubState)
		if err != nil {
			return nil, apierror.Errorf("Unable to load the existing GCP project state.")
		}
	}

	if _, inQuery := queriedFields["googleWorkspaceGroupEmail"]; inQuery {
		err := r.database.LoadReconcilerStateForTeam(ctx, sqlc.ReconcilerNameGoogleWorkspaceAdmin, obj.Slug, googleWorkspaceState)
		if err != nil {
			return nil, apierror.Errorf("Unable to load the existing Google Workspace state.")
		}
	}

	gcpProjects := make([]*model.GcpProject, 0)
	if _, inQuery := queriedFields["gcpProjects"]; inQuery {
		err := r.database.LoadReconcilerStateForTeam(ctx, sqlc.ReconcilerNameGoogleGcpProject, obj.Slug, gcpProjectState)
		if err != nil {
			return nil, apierror.Errorf("Unable to load the existing GCP project state.")
		}
		if gcpProjectState.Projects == nil {
			gcpProjectState.Projects = make(map[string]reconcilers.GoogleGcpEnvironmentProject)
		}

		for env, projectID := range gcpProjectState.Projects {
			gcpProjects = append(gcpProjects, &model.GcpProject{
				Environment: env,
				ProjectName: google_gcp_reconciler.GetProjectDisplayName(obj.Slug, env),
				ProjectID:   projectID.ProjectID,
			})
		}
	}

	naisNamespaces := make([]*model.NaisNamespace, 0)
	if _, inQuery := queriedFields["naisNamespaces"]; inQuery {
		err := r.database.LoadReconcilerStateForTeam(ctx, sqlc.ReconcilerNameNaisNamespace, obj.Slug, naisNamespaceState)
		if err != nil {
			return nil, apierror.Errorf("Unable to load the existing GCP project state.")
		}
		if naisNamespaceState.Namespaces == nil {
			naisNamespaceState.Namespaces = make(map[string]slug.Slug)
		}

		for environment, namespace := range naisNamespaceState.Namespaces {
			naisNamespaces = append(naisNamespaces, &model.NaisNamespace{
				Environment: environment,
				Namespace:   &namespace,
			})
		}
	}

	if _, inQuery := queriedFields["azureADGroupId"]; inQuery {
		err := r.database.LoadReconcilerStateForTeam(ctx, sqlc.ReconcilerNameAzureGroup, obj.Slug, azureADState)
		if err != nil {
			return nil, apierror.Errorf("Unable to load the existing Azure AD state.")
		}
	}

	if _, inQuery := queriedFields["naisDeployKeyProvisioned"]; inQuery {
		err := r.database.LoadReconcilerStateForTeam(ctx, sqlc.ReconcilerNameNaisDeploy, obj.Slug, naisDeployKeyState)
		if err != nil {
			return nil, apierror.Errorf("Unable to load the existing NAIS deploy key state.")
		}
	}

	if _, inQuery := queriedFields["garRepositoryName"]; inQuery {
		err := r.database.LoadReconcilerStateForTeam(ctx, sqlc.ReconcilerNameGoogleGcpGar, obj.Slug, googleGarState)
		if err != nil {
			return nil, apierror.Errorf("Unable to load the existing GAR state.")
		}
	}

	return &model.ReconcilerState{
		GitHubTeamSlug:            gitHubState.Slug,
		GoogleWorkspaceGroupEmail: googleWorkspaceState.GroupEmail,
		GcpProjects:               gcpProjects,
		NaisNamespaces:            naisNamespaces,
		AzureADGroupID:            azureADState.GroupID,
		NaisDeployKeyProvisioned:  naisDeployKeyState.Provisioned,
		GarRepositoryName:         googleGarState.RepositoryName,
	}, nil
}

// SlackAlertsChannels is the resolver for the slackAlertsChannels field.
func (r *teamResolver) SlackAlertsChannels(ctx context.Context, obj *db.Team) ([]*model.SlackAlertsChannel, error) {
	channels := make([]*model.SlackAlertsChannel, 0, len(r.gcpEnvironments))
	existingChannels, err := r.database.GetSlackAlertsChannels(ctx, obj.Slug)
	if err != nil {
		return nil, err
	}

	for _, environment := range r.gcpEnvironments {
		var channel *string
		if value, exists := existingChannels[environment]; exists {
			channel = &value
		}
		channels = append(channels, &model.SlackAlertsChannel{
			Environment: environment,
			ChannelName: channel,
		})
	}
	return channels, nil
}

// GitHubRepositories is the resolver for the gitHubRepositories field.
func (r *teamResolver) GitHubRepositories(ctx context.Context, obj *db.Team) ([]*reconcilers.GitHubRepository, error) {
	state := &reconcilers.GitHubState{}
	err := r.database.LoadReconcilerStateForTeam(ctx, sqlc.ReconcilerNameGithubTeam, obj.Slug, state)
	if err != nil {
		return nil, apierror.Errorf("Unable to load the GitHub state for the team.")
	}
	return state.Repositories, nil
}

// CreatedBy is the resolver for the createdBy field.
func (r *teamDeleteKeyResolver) CreatedBy(ctx context.Context, obj *db.TeamDeleteKey) (*db.User, error) {
	return r.database.GetUserByID(ctx, obj.CreatedBy)
}

// Team is the resolver for the team field.
func (r *teamDeleteKeyResolver) Team(ctx context.Context, obj *db.TeamDeleteKey) (*db.Team, error) {
	return r.database.GetTeamBySlug(ctx, obj.TeamSlug)
}

// Team returns generated.TeamResolver implementation.
func (r *Resolver) Team() generated.TeamResolver { return &teamResolver{r} }

// TeamDeleteKey returns generated.TeamDeleteKeyResolver implementation.
func (r *Resolver) TeamDeleteKey() generated.TeamDeleteKeyResolver { return &teamDeleteKeyResolver{r} }

type (
	teamResolver          struct{ *Resolver }
	teamDeleteKeyResolver struct{ *Resolver }
)
