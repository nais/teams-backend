package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/middleware"
	"github.com/nais/console/pkg/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAutologin(t *testing.T) {
	t.Run("Unknown email", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		database.
			On("GetUserByEmail", mock.Anything, "unknown@example.com").
			Return(nil, errors.New("user not found")).
			Once()
		responseWriter := httptest.NewRecorder()
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actor := authz.ActorFromContext(r.Context())
			assert.Nil(t, actor)
		})
		req := getRequest(context.Background())
		middleware := middleware.Autologin(database, "unknown@example.com")
		middleware(next).ServeHTTP(responseWriter, req)
	})

	t.Run("Valid email", func(t *testing.T) {
		user := &db.User{
			User: &sqlc.User{
				ID:    uuid.New(),
				Email: "user@example.com",
				Name:  "User Name",
			},
		}
		roles := []*db.Role{
			{RoleName: sqlc.RoleNameAdmin},
		}

		database := db.NewMockDatabase(t)
		database.
			On("GetUserByEmail", mock.Anything, "user@example.com").
			Return(user, nil).
			Once()
		database.
			On("GetUserRoles", mock.Anything, user.ID).
			Return(roles, nil).
			Once()
		responseWriter := httptest.NewRecorder()
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actor := authz.ActorFromContext(r.Context())
			assert.NotNil(t, actor)
			assert.Equal(t, user, actor.User)
			assert.Equal(t, roles, actor.Roles)
		})
		req := getRequest(context.Background())
		middleware := middleware.Autologin(database, "user@example.com")
		middleware(next).ServeHTTP(responseWriter, req)
	})
}
