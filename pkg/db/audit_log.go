package db

import (
	"context"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/sqlc"
)

type AuditLog struct {
	*sqlc.AuditLog
}

func (d *database) GetAuditLogsForTeam(ctx context.Context, slug string) ([]*AuditLog, error) {
	targetTeamSlug := nullString(&slug)
	rows, err := d.querier.GetAuditLogsForTeam(ctx, targetTeamSlug)
	if err != nil {
		return nil, err
	}

	entries := make([]*AuditLog, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, &AuditLog{AuditLog: row})
	}
	return entries, nil
}

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
