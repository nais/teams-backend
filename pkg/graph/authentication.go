package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/db"
)

// Me is the resolver for the me field.
func (r *queryResolver) Me(ctx context.Context) (db.AuthenticatedUser, error) {
	return authz.ActorFromContext(ctx).User, nil
}
