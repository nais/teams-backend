package middleware

import (
	"net/http"

	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/dbmodels"
	"gorm.io/gorm"
)

// Authenticates ALL HTTP requests as a specific user.
//
// It goes without saying, but please do not use this in production.
func AutologinMiddleware(db *gorm.DB, email string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			user := &dbmodels.User{}
			tx := db.First(user, "email = ?", email)
			if tx.Error != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := authz.ContextWithUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}
