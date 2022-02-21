package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nais/console/pkg/models"
	"github.com/nais/console/pkg/requests"
	"gorm.io/gorm"
)

func ApiKeyAuthentication(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := &requests.AuthenticatedRequest{}
		err := c.BindHeader(req)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if !strings.HasPrefix(req.Authorization, "Bearer ") || len(req.Authorization) < 8 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "authorization header must have 'Bearer' prefix"})
			return
		}

		key := &models.ApiKey{
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
