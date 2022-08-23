package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/console"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/graph/model"
	"github.com/nais/console/pkg/sqlc"
)

func (r *mutationResolver) CreateServiceAccount(ctx context.Context, input model.CreateServiceAccountInput) (*db.User, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) UpdateServiceAccount(ctx context.Context, serviceAccountID *uuid.UUID, input model.UpdateServiceAccountInput) (*db.User, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) DeleteServiceAccount(ctx context.Context, serviceAccountID *uuid.UUID) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

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

func (r *userResolver) Teams(ctx context.Context, obj *db.User) ([]*db.Team, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *userResolver) IsServiceAccount(ctx context.Context, obj *db.User) (bool, error) {
	actor := authz.UserFromContext(ctx)
	err := authz.RequireAuthorization(actor, sqlc.AuthzNameServiceAccountsRead, obj.ID)
	if err != nil {
		return false, err
	}

	return console.IsServiceAccount(*obj, r.tenantDomain), nil
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
