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

// SynchronizeUsers is the resolver for the synchronizeUsers field.
func (r *mutationResolver) SynchronizeUsers(ctx context.Context) (*model.UserSync, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireGlobalAuthorization(actor, sqlc.AuthzNameUsersyncSynchronize)
	if err != nil {
		return nil, err
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("create log correlation ID: %w", err)
	}

	targets := []auditlogger.Target{
		auditlogger.SystemTarget(sqlc.SystemNameUsersync),
	}
	fields := auditlogger.Fields{
		Action:        sqlc.AuditActionGraphqlApiUsersSync,
		Actor:         actor,
		CorrelationID: correlationID,
	}
	r.auditLogger.Logf(ctx, r.database, targets, fields, "Trigger user sync")
	r.userSync <- correlationID

	return &model.UserSync{
		CorrelationID: &correlationID,
	}, nil
}

// Users is the resolver for the users field.
func (r *queryResolver) Users(ctx context.Context) ([]*db.User, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireGlobalAuthorization(actor, sqlc.AuthzNameUsersList)
	if err != nil {
		return nil, err
	}

	return r.database.GetUsers(ctx)
}

// User is the resolver for the user field.
func (r *queryResolver) User(ctx context.Context, id *uuid.UUID) (*db.User, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireGlobalAuthorization(actor, sqlc.AuthzNameUsersList)
	if err != nil {
		return nil, err
	}

	return r.database.GetUserByID(ctx, *id)
}

// UserByEmail is the resolver for the userByEmail field.
func (r *queryResolver) UserByEmail(ctx context.Context, email string) (*db.User, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireGlobalAuthorization(actor, sqlc.AuthzNameUsersList)
	if err != nil {
		return nil, err
	}

	return r.database.GetUserByEmail(ctx, email)
}

// Teams is the resolver for the teams field.
func (r *userResolver) Teams(ctx context.Context, obj *db.User) ([]*model.TeamMembership, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireGlobalAuthorization(actor, sqlc.AuthzNameTeamsList)
	if err != nil {
		return nil, err
	}

	teams, err := r.database.GetUserTeams(ctx, obj.ID)
	if err != nil {
		return nil, err
	}

	userTeams := make([]*model.TeamMembership, 0, len(teams))
	for _, team := range teams {
		isOwner, err := r.database.UserIsTeamOwner(ctx, obj.ID, team.Slug)
		if err != nil {
			return nil, err
		}

		role := model.TeamRoleMember
		if isOwner {
			role = model.TeamRoleOwner
		}

		userTeams = append(userTeams, &model.TeamMembership{
			Team: team,
			Role: role,
		})
	}

	return userTeams, nil
}

// Roles is the resolver for the roles field.
func (r *userResolver) Roles(ctx context.Context, obj *db.User) ([]*db.Role, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireRole(actor, sqlc.RoleNameAdmin)
	if err != nil && actor.User.GetID() != obj.ID {
		return nil, err
	}

	return r.database.GetUserRoles(ctx, obj.ID)
}

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type userResolver struct{ *Resolver }
