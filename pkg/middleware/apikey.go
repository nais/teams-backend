package middleware

import (
	"github.com/nais/console/pkg/db"
	"net/http"
	"strings"

	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/dbmodels"
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

			key := &dbmodels.ApiKey{}
			err := db.Preload("User").Where("api_key = ?", authHeader[7:]).First(key).Error
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := authz.ContextWithUser(r.Context(), &key.User)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}
