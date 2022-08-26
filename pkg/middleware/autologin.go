package middleware

import (
	"net/http"

	"github.com/nais/console/pkg/db"

	"github.com/nais/console/pkg/authz"
)

// Autologin Authenticates ALL HTTP requests as a specific user. It goes without saying, but please do not use
// this in production.
func Autologin(database db.Database, email string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			user, err := database.GetUserByEmail(ctx, email)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			roles, err := database.GetUserRoles(ctx, user.ID)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx = authz.ContextWithActor(ctx, user, roles)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}
