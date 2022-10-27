package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/sqlc"
)

func (d *database) CreateAPIKey(ctx context.Context, apiKey string, serviceAccountID uuid.UUID) error {
	return d.querier.CreateAPIKey(ctx, sqlc.CreateAPIKeyParams{
		ApiKey:           apiKey,
		ServiceAccountID: serviceAccountID,
	})
}

func (d *database) RemoveApiKeysFromServiceAccount(ctx context.Context, serviceAccountID uuid.UUID) error {
	return d.querier.RemoveApiKeysFromServiceAccount(ctx, serviceAccountID)
}
