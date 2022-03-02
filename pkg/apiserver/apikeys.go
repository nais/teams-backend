package apiserver

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/gin-gonic/gin"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/requests"
	"gorm.io/gorm"
)

type ApiKeysHandler struct {
	db *gorm.DB
}

func (h *ApiKeysHandler) Create(_ *gin.Context, key *dbmodels.ApiKey) (*dbmodels.ApiKey, error) {
	buf := make([]byte, 16)
	_, err := rand.Read(buf)
	if err != nil {
		return nil, err
	}
	key.APIKey = base64.RawURLEncoding.EncodeToString(buf)
	tx := h.db.Transaction(func(tx *gorm.DB) error {
		tx.Delete(&dbmodels.ApiKey{}, "user_id = ?", key.UserID)
		tx.Create(key)
		return tx.Error
	})
	return key, tx
}

func (h *ApiKeysHandler) Delete(_ *gin.Context, req *requests.DeleteApiKeyRequest) error {
	keys := make([]*dbmodels.ApiKey, 0)
	tx := h.db.Find(&keys, "user_id = ?", req.UserID)
	if tx.Error != nil {
		return tx.Error
	}
	tx = h.db.Delete(keys)
	return tx.Error
}
