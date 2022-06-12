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

func (r *auditLogResolver) Synchronization(ctx context.Context, obj *dbmodels.AuditLog) (*dbmodels.Synchronization, error) {
	var synchronization *dbmodels.Synchronization
	err := r.db.Model(&obj).Association("Synchronization").Find(&synchronization)
	if err != nil {
		return nil, err
	}
	return synchronization, nil
}

func (r *auditLogResolver) User(ctx context.Context, obj *dbmodels.AuditLog) (*dbmodels.User, error) {
	if obj.UserID == nil {
		return nil, nil
	}
	var user *dbmodels.User
	err := r.db.Model(&obj).Association("User").Find(&user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *auditLogResolver) Team(ctx context.Context, obj *dbmodels.AuditLog) (*dbmodels.Team, error) {
	if obj.TeamID == nil {
		return nil, nil
	}
	var team *dbmodels.Team
	err := r.db.Model(&obj).Association("Team").Find(&team)
	if err != nil {
		return nil, err
	}
	return team, nil
}

func (r *queryResolver) AuditLogs(ctx context.Context, pagination *model.Pagination, query *model.AuditLogsQuery, sort *model.AuditLogsSort) (*model.AuditLogs, error) {
	auditLogs := make([]*dbmodels.AuditLog, 0)

	if sort == nil {
		sort = &model.AuditLogsSort{
			Field:     model.AuditLogSortFieldCreatedAt,
			Direction: model.SortDirectionDesc,
		}
	}
	pageInfo, err := r.paginatedQuery(ctx, pagination, query, sort, &dbmodels.AuditLog{}, &auditLogs)
	return &model.AuditLogs{
		PageInfo: pageInfo,
		Nodes:    auditLogs,
	}, err
}

// AuditLog returns generated.AuditLogResolver implementation.
func (r *Resolver) AuditLog() generated.AuditLogResolver { return &auditLogResolver{r} }

type auditLogResolver struct{ *Resolver }
