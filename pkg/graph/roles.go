package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/sqlc"
)

func (r *queryResolver) Roles(ctx context.Context) ([]*sqlc.Role, error) {
	roles, err := r.queries.GetRoles(ctx)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *roleBindingResolver) Role(ctx context.Context, obj *sqlc.UserRole) (*sqlc.Role, error) {
	role, err := r.queries.GetRole(ctx, obj.RoleID)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *roleBindingResolver) IsGlobal(ctx context.Context, obj *sqlc.UserRole) (bool, error) {
	return !obj.TargetID.Valid, nil
}

func (r *roleBindingResolver) TargetID(ctx context.Context, obj *sqlc.UserRole) (*uuid.UUID, error) {
	if obj.TargetID.Valid {
		return &obj.TargetID.UUID, nil
	}

	return nil, nil
}

// RoleBinding returns generated.RoleBindingResolver implementation.
func (r *Resolver) RoleBinding() generated.RoleBindingResolver { return &roleBindingResolver{r} }

type roleBindingResolver struct{ *Resolver }
