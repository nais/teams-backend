package middleware_test

import (
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/middleware"
	"github.com/nais/console/pkg/test"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoadUserRoles(t *testing.T) {
	db, _ := test.GetTestDB()
	user1 := &dbmodels.User{Email: "user1@example.com"}
	user2 := &dbmodels.User{Email: "user2@example.com"}
	db.Create([]*dbmodels.User{user1, user2})

	role1 := &dbmodels.Role{Name: "role1"}
	role2 := &dbmodels.Role{Name: "role2"}
	role3 := &dbmodels.Role{Name: "role3"}
	db.Create([]*dbmodels.Role{role1, role2, role3})
	db.Create([]*dbmodels.UserRole{{UserID: *user2.ID, RoleID: *role1.ID}, {UserID: *user2.ID, RoleID: *role3.ID}})

	responseWriter := httptest.NewRecorder()
	middleware := middleware.LoadUserRoles(db)
	req := getRequest()

	t.Run("No user in context", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := authz.UserFromContext(r.Context())
			assert.Nil(t, user)
		})
		middleware(next).ServeHTTP(responseWriter, req)
	})

	t.Run("User with no roles", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := authz.UserFromContext(r.Context())
			assert.NotNil(t, user)
			assert.Len(t, user.RoleBindings, 0)
		})
		middleware(next).ServeHTTP(responseWriter, req.WithContext(authz.ContextWithUser(req.Context(), user1)))
	})

	t.Run("User with roles", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := authz.UserFromContext(r.Context())
			assert.NotNil(t, user)
			assert.Len(t, user.RoleBindings, 2)
			assert.Equal(t, role1.ID, user.RoleBindings[0].Role.ID)
			assert.Equal(t, role3.ID, user.RoleBindings[1].Role.ID)
		})
		middleware(next).ServeHTTP(responseWriter, req.WithContext(authz.ContextWithUser(req.Context(), user2)))
	})
}
