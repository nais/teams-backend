package directives

import (
	"context"
	"fmt"
	"github.com/nais/console/pkg/graph/model"

	"github.com/99designs/gqlgen/graphql"
	"github.com/nais/console/pkg/authz"
)

type AuthorizeDirective func(ctx context.Context, obj interface{}, next graphql.Resolver, operation model.Operation, targetted *bool) (interface{}, error)

func Authorize() AuthorizeDirective {
	return func(ctx context.Context, obj interface{}, next graphql.Resolver, operation model.Operation, targetted *bool) (interface{}, error) {
		user := authz.UserFromContext(ctx)
		if user == nil {
			return nil, fmt.Errorf("this endpoint requires an authenticated user")
		}

		// FIXME: Check if user is allowed to do the operation based on roles attached to the user

		return next(ctx)
	}
}
