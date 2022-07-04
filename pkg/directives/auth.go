package directives

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/nais/console/pkg/authz"
	"gorm.io/gorm"
)

type AuthDirective func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error)

// Auth Make sure there is an authenticated user making this request.
func Auth(db *gorm.DB) AuthDirective {
	return func(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
		user := authz.UserFromContext(ctx)
		if user == nil {
			return nil, fmt.Errorf("this endpoint requires an authenticated user")
		}

		err := db.Where("id = ?", user.ID).First(user).Error
		if err != nil {
			return nil, fmt.Errorf("user in context does not exist in database: %w", err)
		}

		return next(ctx)
	}
}
