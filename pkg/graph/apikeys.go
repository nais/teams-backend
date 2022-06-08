package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/graph/model"
	"gorm.io/gorm"
)

func (r *mutationResolver) CreateAPIKey(ctx context.Context, input model.APIKeyInput) (*model.APIKey, error) {
	buf := make([]byte, 16)
	_, err := rand.Read(buf)
	if err != nil {
		return nil, err
	}
	key := &dbmodels.ApiKey{
		APIKey: base64.RawURLEncoding.EncodeToString(buf),
		UserID: *input.UserID,
	}
	err = r.db.Transaction(func(tx *gorm.DB) error {
		tx.Delete(&dbmodels.ApiKey{}, "user_id = ?", key.UserID)
		tx.Create(key)
		return tx.Error
	})
	if err != nil {
		return nil, err
	}
	return &model.APIKey{
		APIKey: key.APIKey,
	}, nil
}

func (r *mutationResolver) DeleteAPIKey(ctx context.Context, input model.APIKeyInput) (bool, error) {
	keys := make([]*dbmodels.ApiKey, 0)
	tx := r.db.Find(&keys, "user_id = ?", input.UserID)
	if tx.Error != nil {
		return false, tx.Error
	}
	tx = r.db.Delete(keys)
	return tx.Error == nil, tx.Error
}
