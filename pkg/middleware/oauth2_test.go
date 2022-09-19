package middleware_test

import (
	"context"
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
)

func TestOauth2Authentication(t *testing.T) {
	t.Run("No cookie in request", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		responseWriter := httptest.NewRecorder()
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actor := authz.ActorFromContext(r.Context())
			assert.Nil(t, actor)
		})
		req := getRequest(context.Background())
		middleware := middleware.Oauth2Authentication(database)
		middleware(next).ServeHTTP(responseWriter, req)
	})

	t.Run("Invalid cookie value", func(t *testing.T) {
		ctx := context.Background()
		database := db.NewMockDatabase(t)
		responseWriter := httptest.NewRecorder()
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actor := authz.ActorFromContext(r.Context())
			assert.Nil(t, actor)
		})
		req := getRequest(ctx)
		req.AddCookie(&http.Cookie{
			Name:  authn.SessionCookieName,
			Value: "unknown-session-key",
		})
		middleware := middleware.Oauth2Authentication(database)
		middleware(next).ServeHTTP(responseWriter, req)
	})

	t.Run("Valid cookie, no session in store", func(t *testing.T) {
		ctx := context.Background()
		database := db.NewMockDatabase(t)
		responseWriter := httptest.NewRecorder()
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actor := authz.ActorFromContext(r.Context())
			assert.Nil(t, actor)
		})
		req := getRequest(ctx)
		sessionID := uuid.New()
		req.AddCookie(&http.Cookie{
			Name:  authn.SessionCookieName,
			Value: sessionID.String(),
		})
		notFoundErr := errors.New("not found")
		database.
			On("GetSessionByID", ctx, sessionID).
			Return(nil, notFoundErr).
			Once()
		middleware := middleware.Oauth2Authentication(database)
		middleware(next).ServeHTTP(responseWriter, req)
	})

	t.Run("Valid cookie with matching session", func(t *testing.T) {
		ctx := context.Background()
		sessionID := uuid.New()
		userID := uuid.New()
		user := &db.User{
			ID:    userID,
			Email: "user@example.com",
			Name:  "User Name",
		}
		roles := []*db.Role{
			{Name: sqlc.RoleNameAdmin},
		}
		session := &db.Session{Session: &sqlc.Session{
			ID:      sessionID,
			UserID:  userID,
			Expires: time.Now().Add(10 * time.Second),
		}}
		extendedSession := &db.Session{Session: &sqlc.Session{
			ID:      sessionID,
			UserID:  userID,
			Expires: time.Now().Add(30 * time.Minute),
		}}

		database := db.NewMockDatabase(t)
		database.
			On("GetSessionByID", ctx, sessionID).
			Return(session, nil).
			Once()
		database.
			On("GetUserByID", ctx, userID).
			Return(user, nil).
			Once()
		database.
			On("GetUserRoles", ctx, user.ID).
			Return(roles, nil).
			Once()
		database.
			On("ExtendSession", ctx, sessionID).
			Return(extendedSession, nil).
			Once()

		responseWriter := httptest.NewRecorder()
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actor := authz.ActorFromContext(r.Context())
			assert.NotNil(t, actor)
			assert.Equal(t, user, actor.User)
			assert.Equal(t, roles, actor.Roles)
		})
		req := getRequest(ctx)
		req.AddCookie(&http.Cookie{
			Name:  authn.SessionCookieName,
			Value: sessionID.String(),
		})

		middleware := middleware.Oauth2Authentication(database)
		middleware(next).ServeHTTP(responseWriter, req)
	})

	t.Run("Valid cookie with matching expired session", func(t *testing.T) {
		ctx := context.Background()
		sessionID := uuid.New()
		userID := uuid.New()
		database := db.NewMockDatabase(t)
		responseWriter := httptest.NewRecorder()
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actor := authz.ActorFromContext(r.Context())
			assert.Nil(t, actor)
		})
		req := getRequest(ctx)
		req.AddCookie(&http.Cookie{
			Name:  authn.SessionCookieName,
			Value: sessionID.String(),
		})
		session := &db.Session{Session: &sqlc.Session{
			ID:      sessionID,
			UserID:  userID,
			Expires: time.Now().Add(-10 * time.Second),
		}}
		database.
			On("GetSessionByID", ctx, sessionID).
			Return(session, nil).
			Once()
		database.
			On("DeleteSession", ctx, sessionID).
			Return(nil).
			Once()

		middleware := middleware.Oauth2Authentication(database)
		middleware(next).ServeHTTP(responseWriter, req)
	})

	t.Run("Valid cookie with matching session with invalid userID in session", func(t *testing.T) {
		ctx := context.Background()
		sessionID := uuid.New()
		userID := uuid.New()
		database := db.NewMockDatabase(t)
		responseWriter := httptest.NewRecorder()
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actor := authz.ActorFromContext(r.Context())
			assert.Nil(t, actor)
		})
		req := getRequest(ctx)
		req.AddCookie(&http.Cookie{
			Name:  authn.SessionCookieName,
			Value: sessionID.String(),
		})
		session := &db.Session{Session: &sqlc.Session{
			ID:      sessionID,
			UserID:  userID,
			Expires: time.Now().Add(10 * time.Second),
		}}
		database.
			On("GetSessionByID", ctx, sessionID).
			Return(session, nil).
			Once()
		database.
			On("GetUserByID", ctx, userID).
			Return(nil, errors.New("not found")).
			Once()
		database.
			On("DeleteSession", ctx, sessionID).
			Return(nil).
			Once()

		middleware := middleware.Oauth2Authentication(database)
		middleware(next).ServeHTTP(responseWriter, req)
	})
}
