package middleware_test

import (
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/middleware"
	"github.com/nais/console/pkg/test"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nais/console/pkg/dbmodels"
)

func getRequest() *http.Request {
	req, _ := http.NewRequest(http.MethodPost, "/query", nil)
	return req
}

func TestApiKeyAuthentication(t *testing.T) {
	db := test.GetTestDB()
	db.AutoMigrate(&dbmodels.ApiKey{}, &dbmodels.User{})
	user1 := &dbmodels.User{Email: "user1@example.com"}
	user2 := &dbmodels.User{Email: "user2@example.com"}
	user3 := &dbmodels.User{Email: "user3@example.com"}
	db.Create([]*dbmodels.User{user1, user2, user3})

	apiKey1 := &dbmodels.ApiKey{APIKey: "user1-key", UserID: *user1.ID}
	apiKey2 := &dbmodels.ApiKey{APIKey: "user2-key", UserID: *user2.ID}
	db.Create([]*dbmodels.ApiKey{apiKey1, apiKey2})

	middleware := middleware.ApiKeyAuthentication(db)
	responseWriter := httptest.NewRecorder()

	t.Run("No authorization header", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := authz.UserFromContext(r.Context())
			assert.Nil(t, user)
		})

		req := getRequest()
		middleware(next).ServeHTTP(responseWriter, req)
	})

	t.Run("Unknown API key in header", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := authz.UserFromContext(r.Context())
			assert.Nil(t, user)
		})

		req := getRequest()
		req.Header.Set("Authorization", "Bearer unknown")
		middleware(next).ServeHTTP(responseWriter, req)
	})

	t.Run("Valid API key", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := authz.UserFromContext(r.Context())
			assert.NotNil(t, user)
			assert.Equal(t, "user1@example.com", user.Email)
		})

		req := getRequest()
		req.Header.Set("Authorization", "Bearer user1-key")
		middleware(next).ServeHTTP(responseWriter, req)
	})
}
