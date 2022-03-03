package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/nais/console/pkg/auth"
	"github.com/nais/console/pkg/dbmodels"
	"gorm.io/gorm"
)

type Directive func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error)

func ApiKeyDirective() Directive {
	return func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
		user := auth.UserFromContext(ctx)
		if user == nil {
			return nil, fmt.Errorf("this endpoint requires authentication")
		}
		return next(ctx)
	}
}

func ApiKeyAuthentication(db *gorm.DB) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") || len(authHeader) < 8 {
				next.ServeHTTP(w, r)
				return
			}

			key := &dbmodels.ApiKey{
				APIKey: authHeader[7:],
			}

			tx := db.Preload("User").First(key, "api_key = ?", key.APIKey)
			if tx.Error != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := auth.ContextWithUser(r.Context(), key.User)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}
