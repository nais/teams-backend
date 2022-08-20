package directives

import (
	"context"
	"fmt"
	"github.com/nais/console/pkg/db"

	"github.com/99designs/gqlgen/graphql"
	"github.com/nais/console/pkg/authz"
)

type AuthDirective func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error)

// Auth Make sure there is an authenticated user making this request.
func Auth(database db.Database) AuthDirective {
	return func(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
		user := authz.UserFromContext(ctx)
		if user == nil {
			return nil, fmt.Errorf("this endpoint requires an authenticated user")
		}

		user, err := database.GetUserByID(ctx, user.ID)
		if err != nil {
			return nil, fmt.Errorf("user in context does not exist in database: %w", err)
		}

		return next(ctx)
	}
}
