package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/graph/model"
	"github.com/nais/console/pkg/sqlc"
)

func (r *mutationResolver) CreateAPIKey(ctx context.Context, userID *uuid.UUID, revokeExistingAPIKeys bool) (*model.APIKey, error) {
	actor := authz.UserFromContext(ctx)
	err := authz.RequireAuthorization(actor, sqlc.AuthzNameServiceAccountsUpdate, *userID)
	if err != nil {
		return nil, err
	}

	serviceAccount, err := r.database.GetUserByID(ctx, *userID)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 16)
	_, err = rand.Read(buf)
	if err != nil {
		return nil, err
	}

	correlationId, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("unable to create correlation ID for audit log: %w", err)
	}

	key := base64.RawURLEncoding.EncodeToString(buf)

	err = r.database.CreateAPIKey(ctx, key, *userID)
	if err != nil {
		return nil, fmt.Errorf("unable to create API key for user: %w", err)
	}

	systemName := sqlc.SystemNameConsole
	r.auditLogger.Logf(ctx, sqlc.AuditActionConsoleApiKeyCreate, correlationId, &systemName, &actor.Email, nil, &serviceAccount.Email, "API key created")

	return &model.APIKey{
		APIKey: key,
	}, nil
}
