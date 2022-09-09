package directives_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/directives"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuth(t *testing.T) {
	var obj interface{}
	var nextHandler func(ctx context.Context) (res interface{}, err error)

	t.Run("No user in context", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		auth := directives.Auth(database)

		nextHandler = func(ctx context.Context) (res interface{}, err error) {
			panic("Should not be executed")
		}
		_, err := auth(context.Background(), obj, nextHandler)
		assert.EqualError(t, err, "this endpoint requires an authenticated user")
	})

	t.Run("Unknown user in context", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		auth := directives.Auth(database)

		userID := uuid.New()
		user := &db.User{ID: userID}

		database.
			On("GetUserByID", mock.Anything, userID).
			Return(nil, errors.New("record not found")).
			Once()

		nextHandler = func(ctx context.Context) (res interface{}, err error) {
			panic("Should not be executed")
		}
		_, err := auth(authz.ContextWithActor(context.Background(), user, []*db.Role{}), obj, nextHandler)
		assert.EqualError(t, err, "user in context does not exist in database: record not found")
	})
}
