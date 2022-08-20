package db

import (
	"context"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/sqlc"
)

type AuditLog struct {
	*sqlc.AuditLog
}

func (d *database) AddAuditLog(ctx context.Context, auditLog AuditLog) (*AuditLog, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	entry, err := d.querier.CreateAuditLog(ctx, sqlc.CreateAuditLogParams{
		ID:              id,
		CorrelationID:   auditLog.CorrelationID,
		ActorEmail:      auditLog.ActorEmail,
		SystemName:      auditLog.SystemName,
		TargetUserEmail: auditLog.TargetUserEmail,
		TargetTeamSlug:  auditLog.TargetTeamSlug,
		Action:          auditLog.Action,
		Message:         auditLog.Message,
	})
	if err != nil {
		return nil, err
	}

	return &AuditLog{AuditLog: entry}, nil
}
