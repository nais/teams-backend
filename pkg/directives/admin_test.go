package directives_test

import (
	"context"
	"testing"

	"github.com/nais/console/pkg/sqlc"

	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/db"

	"github.com/nais/console/pkg/directives"
	"github.com/stretchr/testify/assert"
)

func TestAdmin(t *testing.T) {
	var obj interface{}
	var nextHandler func(ctx context.Context) (res interface{}, err error)

	t.Run("No user in context", func(t *testing.T) {
		nextHandler = func(ctx context.Context) (res interface{}, err error) {
			panic("Should not be executed")
		}
		_, err := directives.Admin()(context.Background(), obj, nextHandler)
		assert.EqualError(t, err, "this endpoint requires an authenticated user")
	})

	t.Run("User with no admin role", func(t *testing.T) {
		nextHandler = func(ctx context.Context) (res interface{}, err error) {
			panic("Should not be executed")
		}
		user := &db.User{}
		ctx := authz.ContextWithActor(context.Background(), user, []*db.Role{{UserRole: &sqlc.UserRole{RoleName: sqlc.RoleNameTeamcreator}}})
		_, err := directives.Admin()(ctx, obj, nextHandler)
		assert.EqualError(t, err, "this endpoint requires a user with the admin role")
	})

	t.Run("User with no admin role", func(t *testing.T) {
		nextHandler = func(ctx context.Context) (res interface{}, err error) {
			return "executed", nil
		}
		user := &db.User{}
		ctx := authz.ContextWithActor(context.Background(), user, []*db.Role{{UserRole: &sqlc.UserRole{RoleName: sqlc.RoleNameAdmin}}})
		result, err := directives.Admin()(ctx, obj, nextHandler)
		assert.NoError(t, err)
		assert.Equal(t, "executed", result)
	})
}
