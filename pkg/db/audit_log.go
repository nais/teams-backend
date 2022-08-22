package db

import (
	"context"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/sqlc"
)

func (d *database) AddAuditLog(ctx context.Context, correlationID uuid.UUID, systemName sqlc.SystemName, actorEmail, targetTeamSlug, targetUserEmail *string, action sqlc.AuditAction, message string) error {
	id, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	err = d.querier.CreateAuditLog(ctx, sqlc.CreateAuditLogParams{
		ID:              id,
		CorrelationID:   correlationID,
		ActorEmail:      nullString(actorEmail),
		SystemName:      systemName,
		TargetTeamSlug:  nullString(targetTeamSlug),
		TargetUserEmail: nullString(targetUserEmail),
		Action:          action,
		Message:         message,
	})
	if err != nil {
		return err
	}

	return nil
}
