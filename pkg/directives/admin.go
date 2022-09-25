package directives

import (
	"context"
	"fmt"

	"github.com/nais/console/pkg/sqlc"

	"github.com/99designs/gqlgen/graphql"
	"github.com/nais/console/pkg/authz"
)

// Admin Require a user with the admin role to allow the request
func Admin() DirectiveFunc {
	return func(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
		actor := authz.ActorFromContext(ctx)
		if !actor.Authenticated() {
			return nil, fmt.Errorf("this endpoint requires an authenticated user")
		}

		err := authz.RequireRole(actor, sqlc.RoleNameAdmin)
		if err != nil {
			return nil, fmt.Errorf("this endpoint requires a user with the admin role")
		}

		return next(ctx)
	}
}
