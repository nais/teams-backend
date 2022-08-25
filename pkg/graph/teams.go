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
		return nil, err
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

	var team *db.Team
	err = r.database.Transaction(ctx, func(ctx context.Context, dbtx db.Database) error {
		team, err = dbtx.GetTeamByID(ctx, *input.TeamID)
		if err != nil {
			return fmt.Errorf("unable to get team: %w", err)
		}

		members, err := dbtx.GetTeamMembers(ctx, team.ID)
		if err != nil {
			return fmt.Errorf("unable to get existing team members: %w", err)
		}

		for _, userID := range input.UserIds {
			var member *db.User = nil
			for _, m := range members {
				if m.ID == *userID {
					member = m
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

func (r *mutationResolver) SynchronizeTeam(ctx context.Context, teamID *uuid.UUID) (*db.Team, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireAuthorization(actor, sqlc.AuthzNameTeamsUpdate, *teamID)
	if err != nil {
		return nil, err
	}

	team, err := r.database.GetTeamByID(ctx, *teamID)
	if err != nil {
		return nil, err
	}

	reconcilerInput, err := reconcilers.CreateReconcilerInput(ctx, r.database, *team)
	if err != nil {
		return nil, err
	}

	r.teamReconciler <- reconcilerInput

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("unable to create log correlation ID: %w", err)
	}

	fields := auditlogger.Fields{
		Action:         sqlc.AuditActionGraphqlApiTeamSetMemberRole,
		CorrelationID:  correlationID,
		ActorEmail:     &actor.User.Email,
		TargetTeamSlug: &team.Slug,
	}
	r.auditLogger.Logf(ctx, fields, "Synchronize team")
	return team, nil
}

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

	var team *db.Team
	err = r.database.Transaction(ctx, func(ctx context.Context, dbtx db.Database) error {
		team, err = dbtx.GetTeamByID(ctx, *input.TeamID)
		if err != nil {
			return fmt.Errorf("team does not exist: %w", err)
		}

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

	var team *db.Team
	err = r.database.Transaction(ctx, func(ctx context.Context, dbtx db.Database) error {
		team, err = dbtx.GetTeamByID(ctx, *input.TeamID)
		if err != nil {
			return fmt.Errorf("team does not exist: %w", err)
		}

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
			r.auditLogger.Logf(ctx, fields, "User team owner")
		}
		return nil
	})

	return team, nil
}

func (r *mutationResolver) SetTeamMemberRole(ctx context.Context, input model.SetTeamMemberRoleInput) (*db.Team, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireAuthorization(actor, sqlc.AuthzNameTeamsUpdate, *input.TeamID)
	if err != nil {
		return nil, err
	}

	team, err := r.database.GetTeamByID(ctx, *input.TeamID)
	if err != nil {
		return nil, err
	}

	members, err := r.database.GetTeamMembers(ctx, *input.TeamID)
	if err != nil {
		return nil, err
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

	desieredRole, err := sqlcRoleFromTeamRole(input.Role)
	if err != nil {
		return nil, err
	}
	err = r.database.SetTeamMemberRole(ctx, *input.UserID, *input.TeamID, desieredRole)
	if err != nil {
		return nil, err
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("unable to create log correlation ID: %w", err)
	}

	fields := auditlogger.Fields{
		Action:          sqlc.AuditActionGraphqlApiTeamSetMemberRole,
		CorrelationID:   correlationID,
		ActorEmail:      &actor.User.Email,
		TargetTeamSlug:  &team.Slug,
		TargetUserEmail: &member.Email,
	}
	r.auditLogger.Logf(ctx, fields, "Set team member role to %q", desieredRole)

	return team, nil
}

func (r *queryResolver) Teams(ctx context.Context) ([]*db.Team, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireGlobalAuthorization(actor, sqlc.AuthzNameTeamsList)
	if err != nil {
		return nil, err
	}

	return r.database.GetTeams(ctx)
}

func (r *queryResolver) Team(ctx context.Context, id *uuid.UUID) (*db.Team, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireAuthorization(actor, sqlc.AuthzNameTeamsRead, *id)
	if err != nil {
		return nil, err
	}

	return r.database.GetTeamByID(ctx, *id)
}

func (r *teamResolver) Purpose(ctx context.Context, obj *db.Team) (*string, error) {
	var purpose *string
	if obj.Purpose.String != "" {
		purpose = &obj.Purpose.String
	}
	return purpose, nil
}

func (r *teamResolver) Metadata(ctx context.Context, obj *db.Team) (map[string]interface{}, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireAuthorization(actor, sqlc.AuthzNameTeamsRead, obj.ID)
	if err != nil {
		return nil, err
	}

	metadata := make(map[string]interface{})
	for k, v := range obj.Metadata {
		metadata[k] = v
	}

	return metadata, nil
}

func (r *teamResolver) AuditLogs(ctx context.Context, obj *db.Team) ([]*db.AuditLog, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireAuthorization(actor, sqlc.AuthzNameAuditLogsRead, obj.ID)
	if err != nil {
		return nil, err
	}

	return r.database.GetAuditLogsForTeam(ctx, obj.Slug)
}

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

// Team returns generated.TeamResolver implementation.
func (r *Resolver) Team() generated.TeamResolver { return &teamResolver{r} }

type teamResolver struct{ *Resolver }
