package db

import (
	"context"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/sqlc"
)

func (d *database) AddAuditLog(ctx context.Context, correlationId uuid.UUID, actorEmail *string, systemName *sqlc.SystemName, targetTeamSlug, targetUserEmail *string, action sqlc.AuditAction, message string) error {
	id, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	nullSystemName := sqlc.NullSystemName{}
	nullSystemName.Scan(systemName)

	err = d.querier.CreateAuditLog(ctx, sqlc.CreateAuditLogParams{
		ID:              id,
		CorrelationID:   correlationId,
		ActorEmail:      nullString(actorEmail),
		SystemName:      nullSystemName,
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
