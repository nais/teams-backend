package middleware_test

import (
	"errors"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/middleware"
	"github.com/nais/console/pkg/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
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
			user := authz.UserFromContext(r.Context())
			assert.Nil(t, user)
		})
		req := getRequest()
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
		database := db.NewMockDatabase(t)
		database.
			On("GetUserByEmail", mock.Anything, "user@example.com").
			Return(user, nil).
			Once()
		responseWriter := httptest.NewRecorder()
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userInContext := authz.UserFromContext(r.Context())
			assert.NotNil(t, userInContext)
			assert.Equal(t, user, userInContext)
		})
		req := getRequest()
		middleware := middleware.Autologin(database, "user@example.com")
		middleware(next).ServeHTTP(responseWriter, req)
	})

}
