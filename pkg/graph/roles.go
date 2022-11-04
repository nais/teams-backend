package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/sqlc"
)

// Roles is the resolver for the roles field.
func (r *queryResolver) Roles(ctx context.Context) ([]sqlc.RoleName, error) {
	return sqlc.AllRoleNameValues(), nil
}

// Name is the resolver for the name field.
func (r *roleResolver) Name(ctx context.Context, obj *db.Role) (sqlc.RoleName, error) {
	return obj.RoleName, nil
}

// Role returns generated.RoleResolver implementation.
func (r *Resolver) Role() generated.RoleResolver { return &roleResolver{r} }

type roleResolver struct{ *Resolver }
