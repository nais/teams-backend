package directives

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/nais/console/pkg/authz"
)

// Auth Make sure there is an authenticated user making this request.
func Auth() DirectiveFunc {
	return func(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
		actor := authz.ActorFromContext(ctx)
		if !actor.Authenticated() {
			return nil, fmt.Errorf("this endpoint requires an authenticated user")
		}

		return next(ctx)
	}
}
