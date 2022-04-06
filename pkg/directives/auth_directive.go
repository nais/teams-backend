package directives

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/nais/console/pkg/auth"
	"github.com/nais/console/pkg/dbmodels"
	"gorm.io/gorm"
)

type Directive func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error)

// Make sure there is an authenticated user making this request.
// Also fetches all roles connected to that user, either solo or through a team, and puts them into the context.
func Auth(db *gorm.DB) Directive {
	return func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
		user := auth.UserFromContext(ctx)
		if user == nil {
			return nil, fmt.Errorf("this endpoint requires authentication")
		}

		roleBindings, err := loadUserRoleBindings(db, user)
		if err != nil {
			return nil, err
		}

		ctx = auth.ContextWithRoleBindings(ctx, roleBindings)
		return next(ctx)
	}
}

func loadUserRoleBindings(db *gorm.DB, user *dbmodels.User) ([]*dbmodels.RoleBinding, error) {
	u := &dbmodels.User{}

	tx := db.Preload("RoleBindings").Preload("RoleBindings.Role").Find(u, "id = ?", user.ID)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return u.RoleBindings, nil
}
