package directives

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/dbmodels"
	"gorm.io/gorm"
)

type Directive func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error)

// Auth Make sure there is an authenticated user making this request.
// Also fetches all roles connected to that user and puts them into the context.
func Auth(db *gorm.DB) Directive {
	return func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
		user := authz.UserFromContext(ctx)
		if user == nil {
			return nil, fmt.Errorf("this endpoint requires an authenticated user")
		}

		roleBindings, err := loadUserRoleBindings(db, user)
		if err != nil {
			return nil, err
		}

		ctx = authz.ContextWithRoleBindings(ctx, roleBindings)
		return next(ctx)
	}
}

func loadUserRoleBindings(db *gorm.DB, user *dbmodels.User) ([]*dbmodels.RoleBinding, error) {
	newUser := &dbmodels.User{}
	tx := db.Where("ID = ?", user.ID).Preload("RoleBindings").Preload("RoleBindings.Role").Preload("RoleBindings.Role.System").First(newUser)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return newUser.RoleBindings, nil
}
