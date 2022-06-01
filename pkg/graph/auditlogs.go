package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/graph/model"
)

func (r *auditLogResolver) System(ctx context.Context, obj *dbmodels.AuditLog) (*dbmodels.System, error) {
	var system *dbmodels.System
	err := r.db.Model(&obj).Association("System").Find(&system)
	if err != nil {
		return nil, err
	}
	return system, nil
}

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

// AuditLog returns generated.AuditLogResolver implementation.
func (r *Resolver) AuditLog() generated.AuditLogResolver { return &auditLogResolver{r} }

type auditLogResolver struct{ *Resolver }
