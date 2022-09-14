package middleware_test

import (
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
			actor := authz.ActorFromContext(r.Context())
			assert.Nil(t, actor)
		})
		req := getRequest()
		middleware(next).ServeHTTP(responseWriter, req)
	})

	t.Run("Unknown API key in header", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		database.
			On("GetServiceAccountByApiKey", mock.Anything, "unknown").
			Return(nil, errors.New("user not found")).
			Once()
		responseWriter := httptest.NewRecorder()
		middleware := middleware.ApiKeyAuthentication(database)
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actor := authz.ActorFromContext(r.Context())
			assert.Nil(t, actor)
		})
		req := getRequest()
		req.Header.Set("Authorization", "Bearer unknown")
		middleware(next).ServeHTTP(responseWriter, req)
	})

	t.Run("Valid API key", func(t *testing.T) {
		serviceAccount := &db.ServiceAccount{
			ID:   uuid.New(),
			Name: "User Name",
		}
		roles := []*db.Role{
			{Name: sqlc.RoleNameAdmin},
		}

		database := db.NewMockDatabase(t)
		database.
			On("GetServiceAccountByApiKey", mock.Anything, "user1-key").
			Return(serviceAccount, nil).
			Once()
		database.
			On("GetUserRoles", mock.Anything, serviceAccount.ID).
			Return(roles, nil).
			Once()

		responseWriter := httptest.NewRecorder()
		middleware := middleware.ApiKeyAuthentication(database)
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actor := authz.ActorFromContext(r.Context())
			assert.NotNil(t, actor)
			assert.Equal(t, serviceAccount, actor.User)
			assert.Equal(t, roles, actor.Roles)
		})
		req := getRequest()
		req.Header.Set("Authorization", "Bearer user1-key")
		middleware(next).ServeHTTP(responseWriter, req)
	})
}
