package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nais/console/pkg/dbmodels"
	"gorm.io/gorm"
)

type AuthenticatedRequest struct {
	Authorization string `header:"authorization"`
}

func ApiKeyAuthentication(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := &AuthenticatedRequest{}
		err := c.BindHeader(req)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if !strings.HasPrefix(req.Authorization, "Bearer ") || len(req.Authorization) < 8 {
			c.AbortWithStatus(http.StatusBadRequest)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "authorization header must have 'Bearer' prefix"})
			return
		}

		key := &dbmodels.ApiKey{
			APIKey: req.Authorization[7:],
		}

		tx := db.Preload("User").First(key, "api_key = ?", key.APIKey)
		if tx.Error != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
			return
		}

		c.Set("authorized_user", key.User)

		c.Next()
	}
}
