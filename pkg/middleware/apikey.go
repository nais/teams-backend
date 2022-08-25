package middleware

import (
	"net/http"
	"strings"

	"github.com/nais/console/pkg/db"

	"github.com/nais/console/pkg/authz"
)

// ApiKeyAuthentication If the request has an authorization header, we will try to pull the user who owns it from the
// database and put the user into the context.
func ApiKeyAuthentication(database db.Database) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") || len(authHeader) < 8 {
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()
			user, err := database.GetUserByApiKey(ctx, authHeader[7:])
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
