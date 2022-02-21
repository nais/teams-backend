package apiserver

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/gin-gonic/gin"
	"github.com/nais/console/pkg/models"
	"github.com/wI2L/fizz"
	"gorm.io/gorm"
)

type ApiKeysHandler struct {
	db *gorm.DB
}

func (h *ApiKeysHandler) SetupRoutes(parent *fizz.RouterGroup) {
	r := parent.Group(
		"/apikeys",
		"API keys",
		"API keys are used to authenticate users to use the NAIS console without logging on through an identity provider.",
	)

	cruds := &CrudRoute{
		create:   h.Create,
		delete:   h.Delete,
		singular: "API key",
		plural:   "API keys",
		path: map[string]string{
			CrudDelete: "/:user_id",
		},
		description: map[string]string{
			CrudCreate: "Create or rotate a user's API key. Posting to this endpoint will invalidate all other API keys for the given user.",
		},
	}

	cruds.Setup(r)
}

func (h *ApiKeysHandler) Create(_ *gin.Context, key *models.ApiKey) (*models.ApiKey, error) {
	buf := make([]byte, 16)
	_, err := rand.Read(buf)
	if err != nil {
		return nil, err
	}
	key.APIKey = base64.RawURLEncoding.EncodeToString(buf)
	tx := h.db.Transaction(func(tx *gorm.DB) error {
		tx.Delete(&models.ApiKey{}, "user_id = ?", key.UserID)
		tx.Create(key)
		return tx.Error
	})
	return key, tx
}

func (h *ApiKeysHandler) Delete(_ *gin.Context, req *DeleteApiKeyRequest) error {
	key := &models.ApiKey{}
	tx := h.db.First(key, "user_id = ?", req.UserID)
	if tx.Error != nil {
		return tx.Error
	}
	tx = h.db.Delete(key)
	return tx.Error
}
