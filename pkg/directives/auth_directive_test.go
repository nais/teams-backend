package directives_test

import (
	"context"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/directives"
	"github.com/nais/console/pkg/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAuth(t *testing.T) {
	var obj interface{}
	var nextHandler func(ctx context.Context) (res interface{}, err error)
	db := test.GetTestDB()
	db.AutoMigrate(&dbmodels.User{}, &dbmodels.RoleBinding{})

	userWithNoRoleBindings := &dbmodels.User{
		Name:  "user1",
		Email: "user1@example.com",
	}
	userWithRoleBindings := &dbmodels.User{
		Name:  "user2",
		Email: "user2@example.com",
		RoleBindings: []*dbmodels.RoleBinding{
			{
				Role: dbmodels.Role{
					Name: "role_for_user2",
					System: dbmodels.System{
						Name: "system_for_user2",
					},
				},
			},
		},
	}
	userWithOtherRoleBindings := &dbmodels.User{
		Name:  "user3",
		Email: "user3@example.com",
		RoleBindings: []*dbmodels.RoleBinding{
			{
				Role: dbmodels.Role{
					Name: "role_for_user3",
					System: dbmodels.System{
						Name: "system_for_user3",
					},
				},
			},
		},
	}

	db.Create(userWithNoRoleBindings).Create(userWithRoleBindings).Create(userWithOtherRoleBindings)
	auth := directives.Auth(db)

	t.Run("No user in context", func(t *testing.T) {
		nextHandler = func(ctx context.Context) (res interface{}, err error) {
			panic("Should not be executed")
		}
		_, err := auth(context.Background(), obj, nextHandler)
		assert.EqualError(t, err, "this endpoint requires an authenticated user")
	})

	t.Run("Unknown user in context", func(t *testing.T) {
		nextHandler = func(ctx context.Context) (res interface{}, err error) {
			panic("Should not be executed")
		}
		user := &dbmodels.User{Name: "user that does not exist in the DB"}
		_, err := auth(authz.ContextWithUser(context.Background(), user), obj, nextHandler)
		assert.EqualError(t, err, "record not found")
	})

	t.Run("User in context with no role bindings in database", func(t *testing.T) {
		nextHandler = func(ctx context.Context) (res interface{}, err error) {
			roleBindings := authz.RoleBindingsFromContext(ctx)
			assert.Len(t, roleBindings, 0)
			return res, err
		}
		res, err := auth(authz.ContextWithUser(context.Background(), userWithNoRoleBindings), obj, nextHandler)
		assert.NoError(t, err)
		assert.Equal(t, res, obj)
	})

	t.Run("User in context with role bindings in database", func(t *testing.T) {
		nextHandler = func(ctx context.Context) (res interface{}, err error) {
			roleBindings := authz.RoleBindingsFromContext(ctx)
			assert.Len(t, roleBindings, 1)
			assert.Equal(t, "role_for_user2", roleBindings[0].Role.Name)
			assert.Equal(t, "system_for_user2", roleBindings[0].Role.System.Name)
			return res, err
		}
		res, err := auth(authz.ContextWithUser(context.Background(), userWithRoleBindings), obj, nextHandler)
		assert.NoError(t, err)
		assert.Equal(t, res, obj)
	})
}
