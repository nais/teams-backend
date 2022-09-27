package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
)

func (d *database) GetAuditLogsForTeam(ctx context.Context, slug slug.Slug) ([]*AuditLog, error) {
	rows, err := d.querier.GetAuditLogsForTeam(ctx, string(slug))
	if err != nil {
		return nil, err
	}

	entries := make([]*AuditLog, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, &AuditLog{AuditLog: row})
	}
	return entries, nil
}

func (d *database) GetAuditLogsForReconciler(ctx context.Context, reconcilerName sqlc.ReconcilerName) ([]*AuditLog, error) {
	rows, err := d.querier.GetAuditLogsForReconciler(ctx, string(reconcilerName))
	if err != nil {
		return nil, err
	}

	entries := make([]*AuditLog, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, &AuditLog{AuditLog: row})
	}
	return entries, nil
}

func (d *database) CreateAuditLogEntry(ctx context.Context, correlationID uuid.UUID, systemName sqlc.SystemName, actor *string, targetType sqlc.AuditLogsTargetType, targetIdentifier string, action sqlc.AuditAction, message string) error {
	return d.querier.CreateAuditLog(ctx, sqlc.CreateAuditLogParams{
		CorrelationID:    correlationID,
		Actor:            nullString(actor),
		SystemName:       systemName,
		TargetType:       targetType,
		TargetIdentifier: targetIdentifier,
		Action:           action,
		Message:          message,
	})
}
