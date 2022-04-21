package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/graph/model"
)

func (r *queryResolver) Auditlogs(ctx context.Context, input model.AuditLogInput) (*model.AuditLogs, error) {
	var count int64
	auditLogs := make([]*dbmodels.AuditLog, 0)
	query := &dbmodels.AuditLog{
		SystemID:          input.SystemID,
		SynchronizationID: input.SynchronizationID,
		UserID:            input.UserID,
		TeamID:            input.TeamID,
	}
	tx := r.db.WithContext(ctx).
		Where(query).
		Order("created_at DESC").
		Find(&auditLogs)
	tx.Count(&count)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &model.AuditLogs{
		Pagination: &model.Pagination{
			Results: int(count),
		},
		Nodes: auditLogs,
	}, nil
}
