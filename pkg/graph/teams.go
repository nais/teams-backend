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
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
	log "github.com/sirupsen/logrus"
)

// CreateTeam is the resolver for the createTeam field.
func (r *mutationResolver) CreateTeam(ctx context.Context, input model.CreateTeamInput) (*db.Team, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireGlobalAuthorization(actor, sqlc.AuthzNameTeamsCreate)
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
		team, err = dbtx.CreateTeam(ctx, *input.Slug, input.Purpose)
		if err != nil {
			return err
		}

		if actor.User.IsServiceAccount() {
			return nil
		}

		err = dbtx.SetTeamMemberRole(ctx, actor.User.GetID(), team.Slug, sqlc.RoleNameTeamowner)
		if err != nil {
			return err
		}

		return nil
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
	r.auditLogger.Logf(ctx, targets, fields, "Team created")

	r.reconcileTeam(ctx, correlationID, *team)

	return team, nil
}

// UpdateTeam is the resolver for the updateTeam field.
func (r *mutationResolver) UpdateTeam(ctx context.Context, slug *slug.Slug, input model.UpdateTeamInput) (*db.Team, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireTeamAuthorization(actor, sqlc.AuthzNameTeamsUpdate, *slug)
	if err != nil {
		return nil, err
	}

	team, err := r.database.GetTeamBySlug(ctx, *slug)
	if err != nil {
		log.Errorf("get team %q: %s", *slug, err)
		return nil, apierror.ErrTeamNotExist
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

	team, err = r.database.UpdateTeam(ctx, team.Slug, input.Purpose)
	if err != nil {
		return nil, err
	}

	targets := []auditlogger.Target{
		auditlogger.TeamTarget(team.Slug),
	}
	fields := auditlogger.Fields{
		Action:        sqlc.AuditActionGraphqlApiTeamUpdate,
		CorrelationID: correlationID,
		Actor:         actor,
	}

	r.auditLogger.Logf(ctx, targets, fields, "Team configuration saved")

	r.reconcileTeam(ctx, correlationID, *team)

	return team, nil
}

// RemoveUsersFromTeam is the resolver for the removeUsersFromTeam field.
func (r *mutationResolver) RemoveUsersFromTeam(ctx context.Context, slug *slug.Slug, userIds []*uuid.UUID) (*db.Team, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireTeamAuthorization(actor, sqlc.AuthzNameTeamsUpdate, *slug)
	if err != nil {
		return nil, err
	}

	team, err := r.database.GetTeamBySlug(ctx, *slug)
	if err != nil {
		log.Errorf("get team %q: %s", *slug, err)
		return nil, apierror.ErrTeamNotExist
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
			r.auditLogger.Logf(ctx, targets, fields, "Removed user")
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	r.reconcileTeam(ctx, correlationID, *team)

	return team, nil
}

// SynchronizeTeam is the resolver for the synchronizeTeam field.
func (r *mutationResolver) SynchronizeTeam(ctx context.Context, slug *slug.Slug) (*model.TeamSync, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireTeamAuthorization(actor, sqlc.AuthzNameTeamsUpdate, *slug)
	if err != nil {
		return nil, err
	}

	team, err := r.database.GetTeamBySlug(ctx, *slug)
	if err != nil {
		log.Errorf("get team %q: %s", *slug, err)
		return nil, apierror.ErrTeamNotExist
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("create log correlation ID: %w", err)
	}

	if !team.Enabled {
		return nil, apierror.Errorf("Synchronization of this team has been disabled. Unfortunately, there's nothing you can do to resolve the situation. Please contact the NAIS team for support.")
	}

	targets := []auditlogger.Target{
		auditlogger.TeamTarget(team.Slug),
	}
	fields := auditlogger.Fields{
		Action:        sqlc.AuditActionGraphqlApiTeamSync,
		CorrelationID: correlationID,
		Actor:         actor,
	}
	r.auditLogger.Logf(ctx, targets, fields, "Manually scheduled for synchronization")

	r.reconcileTeam(ctx, correlationID, *team)

	return &model.TeamSync{
		Team:          team,
		CorrelationID: &correlationID,
	}, nil
}

// AddTeamMembers is the resolver for the addTeamMembers field.
func (r *mutationResolver) AddTeamMembers(ctx context.Context, slug *slug.Slug, userIds []*uuid.UUID) (*db.Team, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireTeamAuthorization(actor, sqlc.AuthzNameTeamsUpdate, *slug)
	if err != nil {
		return nil, err
	}

	team, err := r.database.GetTeamBySlug(ctx, *slug)
	if err != nil {
		log.Errorf("get team %q: %s", *slug, err)
		return nil, apierror.ErrTeamNotExist
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
			r.auditLogger.Logf(ctx, targets, fields, "Add team member")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	r.reconcileTeam(ctx, correlationID, *team)

	return team, nil
}

// AddTeamOwners is the resolver for the addTeamOwners field.
func (r *mutationResolver) AddTeamOwners(ctx context.Context, slug *slug.Slug, userIds []*uuid.UUID) (*db.Team, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireTeamAuthorization(actor, sqlc.AuthzNameTeamsUpdate, *slug)
	if err != nil {
		return nil, err
	}

	team, err := r.database.GetTeamBySlug(ctx, *slug)
	if err != nil {
		log.Errorf("get team %q: %s", *slug, err)
		return nil, apierror.ErrTeamNotExist
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
			r.auditLogger.Logf(ctx, targets, fields, "Add team owner")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	r.reconcileTeam(ctx, correlationID, *team)

	return team, nil
}

// SetTeamMemberRole is the resolver for the setTeamMemberRole field.
func (r *mutationResolver) SetTeamMemberRole(ctx context.Context, slug *slug.Slug, userID *uuid.UUID, role model.TeamRole) (*db.Team, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireTeamAuthorization(actor, sqlc.AuthzNameTeamsUpdate, *slug)
	if err != nil {
		return nil, err
	}

	team, err := r.database.GetTeamBySlug(ctx, *slug)
	if err != nil {
		log.Errorf("get team %q: %s", *slug, err)
		return nil, apierror.ErrTeamNotExist
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

	r.auditLogger.Logf(ctx, targets, fields, "Assign %q to %s", desiredRole, member.Email)

	r.reconcileTeam(ctx, correlationID, *team)

	return team, nil
}

// DisableTeam is the resolver for the disableTeam field.
func (r *mutationResolver) DisableTeam(ctx context.Context, slug *slug.Slug) (*db.Team, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireTeamAuthorization(actor, sqlc.AuthzNameTeamsUpdate, *slug)
	if err != nil {
		return nil, err
	}

	team, err := r.database.GetTeamBySlug(ctx, *slug)
	if err != nil {
		log.Errorf("get team %q: %s", *slug, err)
		return nil, apierror.ErrTeamNotExist
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("create log correlation ID: %w", err)
	}

	team, err = r.database.DisableTeam(ctx, team.Slug)
	if err != nil {
		return nil, fmt.Errorf("disable team: %w", err)
	}

	targets := []auditlogger.Target{
		auditlogger.TeamTarget(team.Slug),
	}
	fields := auditlogger.Fields{
		Action:        sqlc.AuditActionGraphqlApiTeamDisable,
		Actor:         actor,
		CorrelationID: correlationID,
	}
	r.auditLogger.Logf(ctx, targets, fields, "Disable team")

	r.reconcileTeam(ctx, correlationID, *team)

	return team, nil
}

// EnableTeam is the resolver for the enableTeam field.
func (r *mutationResolver) EnableTeam(ctx context.Context, slug *slug.Slug) (*db.Team, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireTeamAuthorization(actor, sqlc.AuthzNameTeamsUpdate, *slug)
	if err != nil {
		return nil, err
	}

	team, err := r.database.GetTeamBySlug(ctx, *slug)
	if err != nil {
		log.Errorf("get team %q: %s", *slug, err)
		return nil, apierror.ErrTeamNotExist
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("create log correlation ID: %w", err)
	}

	team, err = r.database.EnableTeam(ctx, team.Slug)
	if err != nil {
		return nil, fmt.Errorf("enable team: %w", err)
	}

	targets := []auditlogger.Target{
		auditlogger.TeamTarget(team.Slug),
	}
	fields := auditlogger.Fields{
		Action:        sqlc.AuditActionGraphqlApiTeamEnable,
		Actor:         actor,
		CorrelationID: correlationID,
	}
	r.auditLogger.Logf(ctx, targets, fields, "Enable team")

	r.reconcileTeam(ctx, correlationID, *team)

	return team, nil
}

// Teams is the resolver for the teams field.
func (r *queryResolver) Teams(ctx context.Context) ([]*db.Team, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireGlobalAuthorization(actor, sqlc.AuthzNameTeamsList)
	if err != nil {
		return nil, err
	}

	return r.database.GetTeams(ctx)
}

// Team is the resolver for the team field.
func (r *queryResolver) Team(ctx context.Context, slug *slug.Slug) (*db.Team, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireTeamAuthorization(actor, sqlc.AuthzNameTeamsRead, *slug)
	if err != nil {
		return nil, err
	}

	team, err := r.database.GetTeamBySlug(ctx, *slug)
	if err != nil {
		log.Errorf("get team %q: %s", *slug, err)
		return nil, apierror.ErrTeamNotExist
	}

	return team, nil
}

// Metadata is the resolver for the metadata field.
func (r *teamResolver) Metadata(ctx context.Context, obj *db.Team) ([]*db.TeamMetadata, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireTeamAuthorization(actor, sqlc.AuthzNameTeamsRead, obj.Slug)
	if err != nil {
		return nil, err
	}

	metadata, err := r.database.GetTeamMetadata(ctx, obj.Slug)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

// AuditLogs is the resolver for the auditLogs field.
func (r *teamResolver) AuditLogs(ctx context.Context, obj *db.Team) ([]*db.AuditLog, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireTeamAuthorization(actor, sqlc.AuthzNameAuditLogsRead, obj.Slug)
	if err != nil {
		return nil, err
	}

	return r.database.GetAuditLogsForTeam(ctx, obj.Slug)
}

// Members is the resolver for the members field.
func (r *teamResolver) Members(ctx context.Context, obj *db.Team) ([]*model.TeamMember, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireGlobalAuthorization(actor, sqlc.AuthzNameUsersList)
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
	err := authz.RequireTeamAuthorization(actor, sqlc.AuthzNameTeamsRead, obj.Slug)
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
	naisNamespaceState := &reconcilers.GoogleGcpNaisNamespaceState{
		Namespaces: make(map[string]slug.Slug),
	}
	azureADState := &reconcilers.AzureState{}
	naisDeployKeyState := &reconcilers.NaisDeployKeyState{}

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

		for env, projectID := range gcpProjectState.Projects {
			gcpProjects = append(gcpProjects, &model.GcpProject{
				Environment: env,
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

	return &model.ReconcilerState{
		GitHubTeamSlug:            gitHubState.Slug,
		GoogleWorkspaceGroupEmail: googleWorkspaceState.GroupEmail,
		GcpProjects:               gcpProjects,
		NaisNamespaces:            naisNamespaces,
		AzureADGroupID:            azureADState.GroupID,
		NaisDeployKeyProvisioned:  naisDeployKeyState.Provisioned,
	}, nil
}

// Team returns generated.TeamResolver implementation.
func (r *Resolver) Team() generated.TeamResolver { return &teamResolver{r} }

type teamResolver struct{ *Resolver }
