package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/graph/model"
	console_reconciler "github.com/nais/console/pkg/reconcilers/console"
	"github.com/nais/console/pkg/roles"
	"gorm.io/gorm"
)

func (r *mutationResolver) CreateAPIKey(ctx context.Context, userID *uuid.UUID) (*model.APIKey, error) {
	actor := authz.UserFromContext(ctx)
	err := authz.RequireAuthorization(actor, roles.AuthorizationServiceAccountsUpdate, *userID)
	if err != nil {
		return nil, err
	}

	serviceAccount := &dbmodels.User{}
	err = r.db.Where("id = ?", userID).First(serviceAccount).Error
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 16)
	_, err = rand.Read(buf)
	if err != nil {
		return nil, err
	}

	corr, err := r.createCorrelation(ctx)
	if err != nil {
		return nil, err
	}
	key := &dbmodels.ApiKey{
		APIKey: base64.RawURLEncoding.EncodeToString(buf),
		UserID: *userID,
	}
	err = r.db.Transaction(func(tx *gorm.DB) error {
		err = tx.Where("user_id = ?", key.UserID).Delete(&dbmodels.ApiKey{}).Error
		if err != nil {
			return err
		}

		err = db.CreateTrackedObject(ctx, tx, key)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	r.auditLogger.Logf(console_reconciler.OpCreateApiKey, *corr, r.system, actor, nil, serviceAccount, "API key created")

	return &model.APIKey{
		APIKey: key.APIKey,
	}, nil
}

func (r *mutationResolver) DeleteAPIKey(ctx context.Context, userID *uuid.UUID) (bool, error) {
	actor := authz.UserFromContext(ctx)
	err := authz.RequireAuthorization(actor, roles.AuthorizationServiceAccountsUpdate, *userID)
	if err != nil {
		return false, err
	}

	serviceAccount := &dbmodels.User{}
	err = r.db.Where("id = ?", userID).First(serviceAccount).Error
	if err != nil {
		return false, err
	}

	corr, err := r.createCorrelation(ctx)
	if err != nil {
		return false, err
	}

	err = r.db.Where("user_id = ?", userID).Delete(&dbmodels.ApiKey{}).Error
	if err != nil {
		return false, err
	}

	r.auditLogger.Logf(console_reconciler.OpDeleteApiKey, *corr, r.system, actor, nil, serviceAccount, "API key deleted")

	return true, nil
}
