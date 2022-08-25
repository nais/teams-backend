package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/sqlc"
)

func (d *database) CreateAPIKey(ctx context.Context, apiKey string, userID uuid.UUID) error {
	return d.querier.CreateAPIKey(ctx, sqlc.CreateAPIKeyParams{
		ApiKey: apiKey,
		UserID: userID,
	})
}