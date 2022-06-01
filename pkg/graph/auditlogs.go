package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/graph/model"
)

func (r *queryResolver) AuditLogs(ctx context.Context, input *model.QueryAuditLogsInput, sort *model.QueryAuditLogsSortInput) (*model.AuditLogs, error) {
	auditLogs := make([]*dbmodels.AuditLog, 0)

	if sort == nil {
		sort = &model.QueryAuditLogsSortInput{
			Field:     model.AuditLogSortFieldCreatedAt,
			Direction: model.SortDirectionDesc,
		}
	}
	pagination, err := r.paginatedQuery(ctx, input, sort, &dbmodels.AuditLog{}, &auditLogs)
	return &model.AuditLogs{
		Pagination: pagination,
		Nodes:      auditLogs,
	}, err
}
