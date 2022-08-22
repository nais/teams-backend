package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/graph/model"
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

func (r *queryResolver) Users(ctx context.Context, pagination *model.Pagination, query *model.UsersQuery, sort *model.UsersSort) (*model.Users, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) User(ctx context.Context, id *uuid.UUID) (*db.User, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Me(ctx context.Context) (*db.User, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *userResolver) HasAPIKey(ctx context.Context, obj *db.User) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *userResolver) IsServiceAccount(ctx context.Context, obj *db.User) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *userResolver) CreatedAt(ctx context.Context, obj *db.User) (*time.Time, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *userResolver) Roles(ctx context.Context, obj *db.User) ([]*model.UserRole, error) {
	panic(fmt.Errorf("not implemented"))
}

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type userResolver struct{ *Resolver }
