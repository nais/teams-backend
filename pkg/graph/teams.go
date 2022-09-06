package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/graph/model"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/sqlc"
)

// CreateTeam is the resolver for the createTeam field.
func (r *mutationResolver) CreateTeam(ctx context.Context, input model.CreateTeamInput) (*db.Team, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireGlobalAuthorization(actor, sqlc.AuthzNameTeamsCreate)
	if err != nil {
		return nil, err
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("unable to create log correlation ID: %w", err)
	}

	team, err := r.database.AddTeam(ctx, input.Name, *input.Slug, input.Purpose, actor.User.ID)
	if err != nil {
		return nil, err
	}

	fields := auditlogger.Fields{
		Action:         sqlc.AuditActionGraphqlApiTeamCreate,
		CorrelationID:  correlationID,
		ActorEmail:     &actor.User.Email,
		TargetTeamSlug: &team.Slug,
	}
	r.auditLogger.Logf(ctx, fields, "Team created")

	reconcilerInput, err := reconcilers.CreateReconcilerInput(ctx, r.database, *team)
	if err != nil {
		return nil, fmt.Errorf("unable to reconcile team: %w", err)
	}

	r.teamReconciler <- reconcilerInput

	return team, nil
}

// UpdateTeam is the resolver for the updateTeam field.
func (r *mutationResolver) UpdateTeam(ctx context.Context, teamID *uuid.UUID, input model.UpdateTeamInput) (*db.Team, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireAuthorization(actor, sqlc.AuthzNameTeamsUpdate, *teamID)
	if err != nil {
		return nil, err
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("unable to create log correlation ID: %w", err)
	}

	team, err := r.database.UpdateTeam(ctx, *teamID, input.Name, input.Purpose)
	if err != nil {
		return nil, fmt.Errorf("unable to update team: %w", err)
	}

	fields := auditlogger.Fields{
		Action:         sqlc.AuditActionGraphqlApiTeamUpdate,
		CorrelationID:  correlationID,
		ActorEmail:     &actor.User.Email,
		TargetTeamSlug: &team.Slug,
	}
	r.auditLogger.Logf(ctx, fields, "Team updated")

	reconcilerInput, err := reconcilers.CreateReconcilerInput(ctx, r.database, *team)
	if err != nil {
		return nil, err
	}

	r.teamReconciler <- reconcilerInput.WithCorrelationID(correlationID)

	return team, nil
}

// RemoveUsersFromTeam is the resolver for the removeUsersFromTeam field.
func (r *mutationResolver) RemoveUsersFromTeam(ctx context.Context, input model.RemoveUsersFromTeamInput) (*db.Team, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireAuthorization(actor, sqlc.AuthzNameTeamsUpdate, *input.TeamID)
	if err != nil {
		return nil, err
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("unable to create log correlation ID: %w", err)
	}

	team, err := r.database.GetTeamByID(ctx, *input.TeamID)
	if err != nil {
		return nil, fmt.Errorf("unable to get team: %w", err)
	}

	err = r.database.Transaction(ctx, func(ctx context.Context, dbtx db.Database) error {
		members, err := dbtx.GetTeamMembers(ctx, team.ID)
		if err != nil {
			return fmt.Errorf("unable to get existing team members: %w", err)
		}

		for _, userID := range input.UserIds {
			var member *db.User = nil
			for _, m := range members {
				if m.ID == *userID {
					member = m
					break
				}
			}
			if member == nil {
				return fmt.Errorf("user %q not in team %q", *userID, *input.TeamID)
			}

			err = dbtx.RemoveUserFromTeam(ctx, *userID, *input.TeamID)
			if err != nil {
				return err
			}

			fields := auditlogger.Fields{
				Action:          sqlc.AuditActionGraphqlApiTeamRemoveMember,
				CorrelationID:   correlationID,
				ActorEmail:      &actor.User.Email,
				TargetTeamSlug:  &team.Slug,
				TargetUserEmail: &member.Email,
			}
			r.auditLogger.Logf(ctx, fields, "Removed user")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return team, nil
}

// SynchronizeTeam is the resolver for the synchronizeTeam field.
func (r *mutationResolver) SynchronizeTeam(ctx context.Context, teamID *uuid.UUID) (*model.TeamSync, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireAuthorization(actor, sqlc.AuthzNameTeamsUpdate, *teamID)
	if err != nil {
		return nil, err
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("unable to create log correlation ID: %w", err)
	}

	team, err := r.database.GetTeamByID(ctx, *teamID)
	if err != nil {
		return nil, fmt.Errorf("unable to get team: %w", err)
	}

	fields := auditlogger.Fields{
		Action:         sqlc.AuditActionGraphqlApiTeamSync,
		CorrelationID:  correlationID,
		ActorEmail:     &actor.User.Email,
		TargetTeamSlug: &team.Slug,
	}
	r.auditLogger.Logf(ctx, fields, "Synchronize team")

	reconcilerInput, err := reconcilers.CreateReconcilerInput(ctx, r.database, *team)
	if err != nil {
		return nil, err
	}

	r.teamReconciler <- reconcilerInput.WithCorrelationID(correlationID)

	return &model.TeamSync{
		Team:          team,
		CorrelationID: &correlationID,
	}, nil
}

// AddTeamMembers is the resolver for the addTeamMembers field.
func (r *mutationResolver) AddTeamMembers(ctx context.Context, input model.AddTeamMembersInput) (*db.Team, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireAuthorization(actor, sqlc.AuthzNameTeamsUpdate, *input.TeamID)
	if err != nil {
		return nil, err
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("unable to create log correlation ID: %w", err)
	}

	team, err := r.database.GetTeamByID(ctx, *input.TeamID)
	if err != nil {
		return nil, fmt.Errorf("unable to get team: %w", err)
	}

	err = r.database.Transaction(ctx, func(ctx context.Context, dbtx db.Database) error {
		for _, userID := range input.UserIds {
			user, err := dbtx.GetUserByID(ctx, *userID)
			if err != nil {
				return err
			}

			err = dbtx.SetTeamMemberRole(ctx, *userID, *input.TeamID, sqlc.RoleNameTeammember)
			if err != nil {
				return err
			}

			fields := auditlogger.Fields{
				Action:          sqlc.AuditActionGraphqlApiTeamAddMember,
				CorrelationID:   correlationID,
				ActorEmail:      &actor.User.Email,
				TargetTeamSlug:  &team.Slug,
				TargetUserEmail: &user.Email,
			}
			r.auditLogger.Logf(ctx, fields, "Add team member")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return team, nil
}

// AddTeamOwners is the resolver for the addTeamOwners field.
func (r *mutationResolver) AddTeamOwners(ctx context.Context, input model.AddTeamOwnersInput) (*db.Team, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireAuthorization(actor, sqlc.AuthzNameTeamsUpdate, *input.TeamID)
	if err != nil {
		return nil, err
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("unable to create log correlation ID: %w", err)
	}

	team, err := r.database.GetTeamByID(ctx, *input.TeamID)
	if err != nil {
		return nil, fmt.Errorf("unable to get team: %w", err)
	}

	err = r.database.Transaction(ctx, func(ctx context.Context, dbtx db.Database) error {
		for _, userID := range input.UserIds {
			user, err := dbtx.GetUserByID(ctx, *userID)
			if err != nil {
				return err
			}

			err = dbtx.SetTeamMemberRole(ctx, *userID, *input.TeamID, sqlc.RoleNameTeamowner)
			if err != nil {
				return err
			}

			fields := auditlogger.Fields{
				Action:          sqlc.AuditActionGraphqlApiTeamAddOwner,
				CorrelationID:   correlationID,
				ActorEmail:      &actor.User.Email,
				TargetTeamSlug:  &team.Slug,
				TargetUserEmail: &user.Email,
			}
			r.auditLogger.Logf(ctx, fields, "Add team owner")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return team, nil
}

// SetTeamMemberRole is the resolver for the setTeamMemberRole field.
func (r *mutationResolver) SetTeamMemberRole(ctx context.Context, input model.SetTeamMemberRoleInput) (*db.Team, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireAuthorization(actor, sqlc.AuthzNameTeamsUpdate, *input.TeamID)
	if err != nil {
		return nil, err
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("unable to create log correlation ID: %w", err)
	}

	team, err := r.database.GetTeamByID(ctx, *input.TeamID)
	if err != nil {
		return nil, fmt.Errorf("unable to get team: %w", err)
	}

	members, err := r.database.GetTeamMembers(ctx, *input.TeamID)
	if err != nil {
		return nil, fmt.Errorf("unable to get team members: %w", err)
	}

	var member *db.User = nil
	for _, m := range members {
		if m.ID == *input.UserID {
			member = m
			break
		}
	}
	if member == nil {
		return nil, fmt.Errorf("user %q not in team %q", *input.UserID, *input.TeamID)
	}

	desiredRole, err := sqlcRoleFromTeamRole(input.Role)
	if err != nil {
		return nil, err
	}

	err = r.database.SetTeamMemberRole(ctx, *input.UserID, *input.TeamID, desiredRole)
	if err != nil {
		return nil, err
	}

	fields := auditlogger.Fields{
		Action:          sqlc.AuditActionGraphqlApiTeamSetMemberRole,
		CorrelationID:   correlationID,
		ActorEmail:      &actor.User.Email,
		TargetTeamSlug:  &team.Slug,
		TargetUserEmail: &member.Email,
	}
	r.auditLogger.Logf(ctx, fields, "Set team member role to %q", desiredRole)

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
func (r *queryResolver) Team(ctx context.Context, id *uuid.UUID) (*db.Team, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireAuthorization(actor, sqlc.AuthzNameTeamsRead, *id)
	if err != nil {
		return nil, err
	}

	return r.database.GetTeamByID(ctx, *id)
}

// Purpose is the resolver for the purpose field.
func (r *teamResolver) Purpose(ctx context.Context, obj *db.Team) (*string, error) {
	var purpose *string
	if obj.Purpose.String != "" {
		purpose = &obj.Purpose.String
	}
	return purpose, nil
}

// Metadata is the resolver for the metadata field.
func (r *teamResolver) Metadata(ctx context.Context, obj *db.Team) (map[string]interface{}, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireAuthorization(actor, sqlc.AuthzNameTeamsRead, obj.ID)
	if err != nil {
		return nil, err
	}

	metadata, err := r.database.GetTeamMetadata(ctx, obj.ID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for k, v := range metadata {
		result[k] = v
	}

	return result, nil
}

// AuditLogs is the resolver for the auditLogs field.
func (r *teamResolver) AuditLogs(ctx context.Context, obj *db.Team) ([]*db.AuditLog, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireAuthorization(actor, sqlc.AuthzNameAuditLogsRead, obj.ID)
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

	users, err := r.database.GetTeamMembers(ctx, obj.ID)
	if err != nil {
		return nil, err
	}

	members := make([]*model.TeamMember, len(users))
	for idx, user := range users {
		isOwner, err := r.database.UserIsTeamOwner(ctx, user.ID, obj.ID)
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
	err := authz.RequireAuthorization(actor, sqlc.AuthzNameTeamsRead, obj.ID)
	if err != nil {
		return nil, err
	}

	rows, err := r.database.GetTeamReconcileErrors(ctx, obj.ID)
	if err != nil {
		return nil, err
	}

	syncErrors := make([]*model.SyncError, 0)
	for _, row := range rows {
		syncErrors = append(syncErrors, &model.SyncError{
			CreatedAt: row.CreatedAt,
			System:    row.SystemName,
			Error:     row.ErrorMessage,
		})
	}

	return syncErrors, nil
}

// Team returns generated.TeamResolver implementation.
func (r *Resolver) Team() generated.TeamResolver { return &teamResolver{r} }

type teamResolver struct{ *Resolver }
