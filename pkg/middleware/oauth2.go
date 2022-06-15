package middleware

import (
	"net/http"
	"time"

	"github.com/nais/console/pkg/authn"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/dbmodels"
	"gorm.io/gorm"
)

func Oauth2Authentication(db *gorm.DB, store authn.SessionStore) func(next http.Handler) http.Handler {
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
			err = db.First(user, "email = ?", session.Email).Error
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
