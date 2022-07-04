package middleware

import (
	"github.com/nais/console/pkg/authz"
	"gorm.io/gorm"
	"net/http"
)

// LoadUserRoles Attach roles to the authenticated user in the request context, if one exists.
func LoadUserRoles(db *gorm.DB) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			user := authz.UserFromContext(r.Context())

			if user == nil {
				next.ServeHTTP(w, r)
				return
			}

			err := db.
				Model(user).
				Preload("Role").
				Preload("Role.Authorizations").
				Association("RoleBindings").
				Find(&user.RoleBindings)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			next.ServeHTTP(w, r)
			return
		}
		return http.HandlerFunc(fn)
	}
}
