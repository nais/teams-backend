package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/graph/model"
	"github.com/nais/console/pkg/sqlc"
)

func (r *mutationResolver) CreateAPIKey(ctx context.Context, userID *uuid.UUID) (*model.APIKey, error) {
	actor := authz.UserFromContext(ctx)
	err := authz.RequireAuthorization(actor, sqlc.AuthzNameServiceAccountsUpdate, *userID)
	if err != nil {
		return nil, err
	}

	serviceAccount, err := r.database.GetUserByID(ctx, *userID)
	if err != nil {
		return nil, err
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("unable to create correlation ID for audit log: %w", err)
	}

	buf := make([]byte, 16)
	_, err = rand.Read(buf)
	if err != nil {
		return nil, err
	}
	apiKey := base64.RawURLEncoding.EncodeToString(buf)

	err = r.database.CreateAPIKey(ctx, apiKey, *userID)
	if err != nil {
		return nil, fmt.Errorf("unable to create API key for user: %w", err)
	}

	fields := auditlogger.Fields{
		Action:          sqlc.AuditActionGraphqlApiApiKeyCreate,
		CorrelationID:   correlationID,
		ActorEmail:      &actor.Email,
		TargetUserEmail: &serviceAccount.Email,
	}
	r.auditLogger.Logf(ctx, fields, "API key created")

	return &model.APIKey{
		Key: apiKey,
	}, nil
}

func (r *mutationResolver) DeleteAPIKey(ctx context.Context, userID *uuid.UUID) (bool, error) {
	actor := authz.UserFromContext(ctx)
	err := authz.RequireAuthorization(actor, sqlc.AuthzNameServiceAccountsUpdate, *userID)
	if err != nil {
		return false, err
	}

	err = r.database.RemoveApiKeysFromUser(ctx, *userID)
	return err == nil, err
}
