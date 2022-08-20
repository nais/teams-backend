package middleware

import (
	"github.com/nais/console/pkg/db"
	"net/http"
	"time"

	"github.com/nais/console/pkg/authn"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/dbmodels"
)

// Oauth2Authentication If the request has a session cookie, look up the session from the store, and if it exists, try
// to load the user with the email address stored in the session.
func Oauth2Authentication(database db.Database, store authn.SessionStore) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(authn.SessionCookieName)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			session := store.Get(cookie.Value)
			if session == nil {
				next.ServeHTTP(w, r)
				return
			}

			if session.Expires.Before(time.Now()) {
				next.ServeHTTP(w, r)
				return
			}

			user := &dbmodels.User{}
			err = db.Where("email = ?", session.Email).First(user).Error
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := authz.ContextWithUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}
