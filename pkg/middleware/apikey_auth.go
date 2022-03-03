package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gin-gonic/gin"
	"github.com/nais/console/pkg/dbmodels"
	"gorm.io/gorm"
)

type Directive func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error)

type AuthenticatedRequest struct {
	Authorization string `header:"authorization"`
}

var userCtxKey = &contextKey{"user"}

type contextKey struct {
	name string
}

func ApiKeyAuthentication(db *gorm.DB) Directive {
	return func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
		user := ctx.Value("user")
		if user == nil {
			return nil, fmt.Errorf("not authenticated")
		}
		return next(ctx)
	}
}

func GinApiKeyAuthentication(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := &AuthenticatedRequest{}
		err := c.BindHeader(req)
		if err != nil {
			c.Next()
			return
		}

		if !strings.HasPrefix(req.Authorization, "Bearer ") || len(req.Authorization) < 8 {
			c.Next()
			return
		}

		key := &dbmodels.ApiKey{
			APIKey: req.Authorization[7:],
		}

		tx := db.Preload("User").First(key, "api_key = ?", key.APIKey)
		if tx.Error != nil {
			c.Next()
			return
		}

		c.Set("user", key.User)

		c.Next()
	}
}
