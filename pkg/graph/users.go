package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/graph/model"
	"github.com/nais/console/pkg/sqlc"
)

func (r *queryResolver) Users(ctx context.Context) ([]*db.User, error) {
	actor := authz.UserFromContext(ctx)
	err := authz.RequireGlobalAuthorization(actor, sqlc.AuthzNameUsersList)
	if err != nil {
		return nil, err
	}

	return r.database.GetUsers(ctx)
}

func (r *queryResolver) User(ctx context.Context, id *uuid.UUID) (*db.User, error) {
	actor := authz.UserFromContext(ctx)
	err := authz.RequireGlobalAuthorization(actor, sqlc.AuthzNameUsersList)
	if err != nil {
		return nil, err
	}

	return r.database.GetUserByID(ctx, *id)
}

func (r *queryResolver) Me(ctx context.Context) (*db.User, error) {
	return authz.UserFromContext(ctx), nil
}

func (r *userResolver) Teams(ctx context.Context, obj *db.User) ([]*model.UserTeam, error) {
	actor := authz.UserFromContext(ctx)
	err := authz.RequireGlobalAuthorization(actor, sqlc.AuthzNameTeamsList)
	if err != nil {
		return nil, err
	}

	teams, err := r.database.GetUserTeams(ctx, obj.ID)
	if err != nil {
		return nil, err
	}

	userTeams := make([]*model.UserTeam, 0, len(teams))
	for _, team := range teams {
		isOwner, err := r.database.UserIsTeamOwner(ctx, obj.ID, team.ID)
		if err != nil {
			return nil, err
		}

		role := model.TeamRoleMember
		if isOwner {
			role = model.TeamRoleOwner
		}

		userTeams = append(userTeams, &model.UserTeam{
			Team: team,
			Role: role,
		})
	}

	return userTeams, nil
}

func (r *userResolver) Roles(ctx context.Context, obj *db.User) ([]*db.Role, error) {
	actor := authz.UserFromContext(ctx)
	err := authz.RequireAuthorizationOrTargetMatch(actor, sqlc.AuthzNameUsersUpdate, obj.ID)
	if err != nil {
		return nil, err
	}

	return r.database.GetUserRoles(ctx, obj.ID)
}

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type userResolver struct{ *Resolver }
