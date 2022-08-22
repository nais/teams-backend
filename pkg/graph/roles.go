package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/graph/generated"
)

func (r *roleResolver) Name(ctx context.Context, obj *db.Role) (string, error) {
	return string(obj.RoleName), nil
}

func (r *roleResolver) TargetID(ctx context.Context, obj *db.Role) (*uuid.UUID, error) {
	if obj.TargetID.Valid {
		return &obj.TargetID.UUID, nil
	}
	return nil, nil
}

// Role returns generated.RoleResolver implementation.
func (r *Resolver) Role() generated.RoleResolver { return &roleResolver{r} }

type roleResolver struct{ *Resolver }
