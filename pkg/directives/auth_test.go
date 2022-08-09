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
	db, _ := test.GetTestDB()

	user := &dbmodels.User{
		Name:  "user1",
		Email: "user1@example.com",
	}
	db.Create(user)
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
		assert.EqualError(t, err, "user in context does not exist in database: record not found")
	})
}
