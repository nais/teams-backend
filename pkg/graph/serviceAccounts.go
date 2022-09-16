package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/sqlc"
)

// Roles is the resolver for the roles field.
func (r *serviceAccountResolver) Roles(ctx context.Context, obj *db.ServiceAccount) ([]*db.Role, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireAuthorizationOrTargetMatch(actor, sqlc.AuthzNameUsersUpdate, obj.ID)
	if err != nil {
		return nil, err
	}

	return r.database.GetUserRoles(ctx, obj.ID)
}

// ServiceAccount returns generated.ServiceAccountResolver implementation.
func (r *Resolver) ServiceAccount() generated.ServiceAccountResolver {
	return &serviceAccountResolver{r}
}

type serviceAccountResolver struct{ *Resolver }
