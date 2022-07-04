package middleware

import (
	"net/http"

	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/dbmodels"
	"gorm.io/gorm"
)

// Autologin Authenticates ALL HTTP requests as a specific user. It goes without saying, but please do not use
// this in production.
func Autologin(db *gorm.DB, email string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			user := &dbmodels.User{}
			err := db.Where("email = ?", email).First(user).Error
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
