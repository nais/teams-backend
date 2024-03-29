package db

import (
	"context"

	"github.com/nais/teams-backend/pkg/types"

	"github.com/google/uuid"
	"github.com/nais/teams-backend/pkg/slug"
	"github.com/nais/teams-backend/pkg/sqlc"
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

func (d *database) CreateAuditLogEntry(ctx context.Context, correlationID uuid.UUID, componentName types.ComponentName, actor *string, targetType types.AuditLogsTargetType, targetIdentifier string, action types.AuditAction, message string) error {
	return d.querier.CreateAuditLog(ctx, sqlc.CreateAuditLogParams{
		CorrelationID:    correlationID,
		Actor:            actor,
		ComponentName:    string(componentName),
		TargetType:       string(targetType),
		TargetIdentifier: targetIdentifier,
		Action:           string(action),
		Message:          message,
	})
}

func (d *database) GetAuditLogsForCorrelationID(ctx context.Context, correlationID uuid.UUID) ([]*AuditLog, error) {
	rows, err := d.querier.GetAuditLogsForCorrelationID(ctx, correlationID)
	if err != nil {
		return nil, err
	}

	entries := make([]*AuditLog, len(rows))
	for i, row := range rows {
		entries[i] = &AuditLog{AuditLog: row}
	}

	return entries, nil
}
