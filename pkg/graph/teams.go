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

	var team *db.Team
	err = r.database.Transaction(ctx, func(ctx context.Context, dbtx db.Database) error {
		team, err = dbtx.CreateTeam(ctx, input.Name, *input.Slug, input.Purpose)
		if err != nil {
			return err
		}

		if actor.User.IsServiceAccount() {
			return nil
		}

		err = dbtx.SetTeamMemberRole(ctx, actor.User.GetID(), team.ID, sqlc.RoleNameTeamowner)
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

	if !team.Enabled {
		return nil, fmt.Errorf("team is not enabled, unable to synchronize")
	}

	targets := []auditlogger.Target{
		auditlogger.TeamTarget(team.Slug),
	}
	fields := auditlogger.Fields{
		Action:        sqlc.AuditActionGraphqlApiTeamSync,
		CorrelationID: correlationID,
		Actor:         actor,
	}
	r.auditLogger.Logf(ctx, targets, fields, "Synchronize team")

	r.reconcileTeam(ctx, correlationID, *team)

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
func (r *mutationResolver) DisableTeam(ctx context.Context, teamID *uuid.UUID) (*db.Team, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireAuthorization(actor, sqlc.AuthzNameTeamsUpdate, *teamID)
	if err != nil {
		return nil, err
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("unable to create log correlation ID: %w", err)
	}

	team, err := r.database.DisableTeam(ctx, *teamID)
	if err != nil {
		return nil, fmt.Errorf("unable to disable team: %w", err)
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
func (r *mutationResolver) EnableTeam(ctx context.Context, teamID *uuid.UUID) (*db.Team, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireAuthorization(actor, sqlc.AuthzNameTeamsUpdate, *teamID)
	if err != nil {
		return nil, err
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("unable to create log correlation ID: %w", err)
	}

	team, err := r.database.EnableTeam(ctx, *teamID)
	if err != nil {
		return nil, fmt.Errorf("unable to enable team: %w", err)
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
func (r *teamResolver) Metadata(ctx context.Context, obj *db.Team) ([]*db.TeamMetadata, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireAuthorization(actor, sqlc.AuthzNameTeamsRead, obj.ID)
	if err != nil {
		return nil, err
	}

	metadata, err := r.database.GetTeamMetadata(ctx, obj.ID)
	if err != nil {
		return nil, err
	}

	return metadata, nil
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

	rows, err := r.database.GetTeamReconcilerErrors(ctx, obj.ID)
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

// Team returns generated.TeamResolver implementation.
func (r *Resolver) Team() generated.TeamResolver { return &teamResolver{r} }

type teamResolver struct{ *Resolver }
