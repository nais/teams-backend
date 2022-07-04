package middleware_test

import (
	"github.com/nais/console/pkg/authn"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/middleware"
	"github.com/nais/console/pkg/test"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOauth2Authentication(t *testing.T) {
	db := test.GetTestDB()
	db.AutoMigrate(&dbmodels.User{})
	user1 := &dbmodels.User{Email: "user1@example.com"}
	user2 := &dbmodels.User{Email: "user2@example.com"}
	db.Create([]*dbmodels.User{user1, user2})

	responseWriter := httptest.NewRecorder()
	store := authn.NewStore()

	t.Run("No cookie in request", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := authz.UserFromContext(r.Context())
			assert.Nil(t, user)
		})

		req := getRequest()
		middleware := middleware.Oauth2Authentication(db, store)
		middleware(next).ServeHTTP(responseWriter, req)
	})

	t.Run("Valid cookie, no session in store", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := authz.UserFromContext(r.Context())
			assert.Nil(t, user)
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
		middleware := middleware.Oauth2Authentication(db, store)
		middleware(next).ServeHTTP(responseWriter, req)
	})

	t.Run("Valid cookie with matching session", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := authz.UserFromContext(r.Context())
			assert.NotNil(t, user)
			assert.Equal(t, "user1@example.com", user.Email)
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

		middleware := middleware.Oauth2Authentication(db, store)
		middleware(next).ServeHTTP(responseWriter, req)
	})

	t.Run("Valid cookie with matching expired session", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := authz.UserFromContext(r.Context())
			assert.Nil(t, user)
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

		middleware := middleware.Oauth2Authentication(db, store)
		middleware(next).ServeHTTP(responseWriter, req)
	})

	t.Run("Valid cookie with matching session with invalid email in session", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := authz.UserFromContext(r.Context())
			assert.Nil(t, user)
		})

		req := getRequest()
		req.AddCookie(&http.Cookie{
			Name:  authn.SessionCookieName,
			Value: "session-key-1",
		})
		store.Create(&authn.Session{
			Key:     "session-key-1",
			Expires: time.Now().Add(10 * time.Second),
			Email:   "user3@example.com",
		})

		middleware := middleware.Oauth2Authentication(db, store)
		middleware(next).ServeHTTP(responseWriter, req)
	})
}
