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

		tx := db.Preload("Roles").Preload("Teams.Roles").Find(user, "id = ?", user.ID)
		if tx.Error != nil {
			return nil, tx.Error
		}

		roles := flattenRoles(user)

		ctx = auth.ContextWithRoles(ctx, roles)
		return next(ctx)
	}
}

func flattenRoles(user *dbmodels.User) []*dbmodels.Role {
	roles := make([]*dbmodels.Role, 0)
	for _, role := range user.Roles {
		roles = append(roles, role)
	}
	for _, team := range user.Teams {
		for _, role := range team.Roles {
			roles = append(roles, role)
		}
	}
	return roles
}
