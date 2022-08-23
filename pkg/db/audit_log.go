package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
)

type AuditLog struct {
	*sqlc.AuditLog
}

func (d *database) GetAuditLogsForTeam(ctx context.Context, slug slug.Slug) ([]*AuditLog, error) {
	rows, err := d.querier.GetAuditLogsForTeam(ctx, &slug)
	if err != nil {
		return nil, err
	}

	entries := make([]*AuditLog, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, &AuditLog{AuditLog: row})
	}
	return entries, nil
}

func (d *database) AddAuditLog(ctx context.Context, correlationID uuid.UUID, systemName sqlc.SystemName, actorEmail *string, targetTeamSlug *slug.Slug, targetUserEmail *string, action sqlc.AuditAction, message string) error {
	id, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	err = d.querier.CreateAuditLog(ctx, sqlc.CreateAuditLogParams{
		ID:              id,
		CorrelationID:   correlationID,
		ActorEmail:      nullString(actorEmail),
		SystemName:      systemName,
		TargetTeamSlug:  targetTeamSlug,
		TargetUserEmail: nullString(targetUserEmail),
		Action:          action,
		Message:         message,
	})
	if err != nil {
		return err
	}

	return nil
}
