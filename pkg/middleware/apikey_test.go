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

func getRequest() *http.Request {
	req, _ := http.NewRequest(http.MethodPost, "/query", nil)
	return req
}

func TestApiKeyAuthentication(t *testing.T) {
	t.Run("No authorization header", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		responseWriter := httptest.NewRecorder()
		middleware := middleware.ApiKeyAuthentication(database)
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := authz.UserFromContext(r.Context())
			assert.Nil(t, user)
		})
		req := getRequest()
		middleware(next).ServeHTTP(responseWriter, req)
	})

	t.Run("Unknown API key in header", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		database.
			On("GetUserByApiKey", mock.Anything, "unknown").
			Return(nil, errors.New("user not found")).
			Once()
		responseWriter := httptest.NewRecorder()
		middleware := middleware.ApiKeyAuthentication(database)
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := authz.UserFromContext(r.Context())
			assert.Nil(t, user)
		})
		req := getRequest()
		req.Header.Set("Authorization", "Bearer unknown")
		middleware(next).ServeHTTP(responseWriter, req)
	})

	t.Run("Valid API key", func(t *testing.T) {
		user := &db.User{
			User: &sqlc.User{
				ID:    uuid.New(),
				Email: "user@example.com",
				Name:  "User Name",
			},
		}
		database := db.NewMockDatabase(t)
		database.
			On("GetUserByApiKey", mock.Anything, "user1-key").
			Return(user, nil).
			Once()
		responseWriter := httptest.NewRecorder()
		middleware := middleware.ApiKeyAuthentication(database)
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userInContext := authz.UserFromContext(r.Context())
			assert.NotNil(t, userInContext)
			assert.Equal(t, user, userInContext)
		})
		req := getRequest()
		req.Header.Set("Authorization", "Bearer user1-key")
		middleware(next).ServeHTTP(responseWriter, req)
	})
}
