package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/graph/generated"
)

func (r *queryResolver) Roles(ctx context.Context) ([]*db.Role, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *roleResolver) Name(ctx context.Context, obj *db.Role) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

// Role returns generated.RoleResolver implementation.
func (r *Resolver) Role() generated.RoleResolver { return &roleResolver{r} }

type roleResolver struct{ *Resolver }
