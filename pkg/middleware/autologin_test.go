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

func TestAutologin(t *testing.T) {
	db := test.GetTestDB()
	db.AutoMigrate(&dbmodels.User{})
	user1 := &dbmodels.User{Email: "user1@example.com"}
	user2 := &dbmodels.User{Email: "user2@example.com"}
	db.Create([]*dbmodels.User{user1, user2})

	responseWriter := httptest.NewRecorder()

	t.Run("Unknown email", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := authz.UserFromContext(r.Context())
			assert.Nil(t, user)
		})

		req := getRequest()
		middleware := middleware.Autologin(db, "unknown@example.com")
		middleware(next).ServeHTTP(responseWriter, req)
	})

	t.Run("Valid email", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := authz.UserFromContext(r.Context())
			assert.NotNil(t, user)
			assert.Equal(t, user2.ID, user.ID)
		})

		req := getRequest()
		middleware := middleware.Autologin(db, "user2@example.com")
		middleware(next).ServeHTTP(responseWriter, req)
	})
}
