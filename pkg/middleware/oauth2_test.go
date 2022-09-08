package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/authn"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/middleware"
	"github.com/nais/console/pkg/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOauth2Authentication(t *testing.T) {
	t.Run("No cookie in request", func(t *testing.T) {
		store := authn.NewStore()
		database := db.NewMockDatabase(t)
		responseWriter := httptest.NewRecorder()
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actor := authz.ActorFromContext(r.Context())
			assert.Nil(t, actor)
		})
		req := getRequest()
		middleware := middleware.Oauth2Authentication(database, store)
		middleware(next).ServeHTTP(responseWriter, req)
	})

	t.Run("Valid cookie, no session in store", func(t *testing.T) {
		store := authn.NewStore()
		database := db.NewMockDatabase(t)
		responseWriter := httptest.NewRecorder()
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actor := authz.ActorFromContext(r.Context())
			assert.Nil(t, actor)
		})
		req := getRequest()
		req.AddCookie(&http.Cookie{
			Name:  authn.SessionCookieName,
			Value: "unknown-session-key",
		})
		store.Create(&authn.Session{
			Key:     "session-key",
			Expires: time.Now().Add(10 * time.Second),
			Email:   "user1@example.com",
		})
		middleware := middleware.Oauth2Authentication(database, store)
		middleware(next).ServeHTTP(responseWriter, req)
	})

	t.Run("Valid cookie with matching session", func(t *testing.T) {
		store := authn.NewStore()
		user := &db.User{
			ID:    uuid.New(),
			Email: "user1@example.com",
			Name:  "User Name",
		}
		roles := []*db.Role{
			{Name: sqlc.RoleNameAdmin},
		}

		database := db.NewMockDatabase(t)
		database.
			On("GetUserByEmail", mock.Anything, "user1@example.com").
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
		req := getRequest()
		req.AddCookie(&http.Cookie{
			Name:  authn.SessionCookieName,
			Value: "session-key-1",
		})
		store.Create(&authn.Session{
			Key:     "session-key-1",
			Expires: time.Now().Add(10 * time.Second),
			Email:   "user1@example.com",
		})
		store.Create(&authn.Session{
			Key:     "session-key-2",
			Expires: time.Now().Add(10 * time.Second),
			Email:   "user2@example.com",
		})

		middleware := middleware.Oauth2Authentication(database, store)
		middleware(next).ServeHTTP(responseWriter, req)
	})

	t.Run("Valid cookie with matching expired session", func(t *testing.T) {
		store := authn.NewStore()
		database := db.NewMockDatabase(t)
		responseWriter := httptest.NewRecorder()
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actor := authz.ActorFromContext(r.Context())
			assert.Nil(t, actor)
		})
		req := getRequest()
		req.AddCookie(&http.Cookie{
			Name:  authn.SessionCookieName,
			Value: "session-key-1",
		})
		store.Create(&authn.Session{
			Key:     "session-key-1",
			Expires: time.Now().Add(-10 * time.Second),
			Email:   "user1@example.com",
		})
		middleware := middleware.Oauth2Authentication(database, store)
		middleware(next).ServeHTTP(responseWriter, req)
	})

	t.Run("Valid cookie with matching session with invalid email in session", func(t *testing.T) {
		store := authn.NewStore()
		database := db.NewMockDatabase(t)
		database.
			On("GetUserByEmail", mock.Anything, "user1@example.com").
			Return(nil, errors.New("user not found")).
			Once()
		responseWriter := httptest.NewRecorder()
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actor := authz.ActorFromContext(r.Context())
			assert.Nil(t, actor)
		})
		req := getRequest()
		req.AddCookie(&http.Cookie{
			Name:  authn.SessionCookieName,
			Value: "session-key-1",
		})
		store.Create(&authn.Session{
			Key:     "session-key-1",
			Expires: time.Now().Add(10 * time.Second),
			Email:   "user1@example.com",
		})
		middleware := middleware.Oauth2Authentication(database, store)
		middleware(next).ServeHTTP(responseWriter, req)
	})
}
