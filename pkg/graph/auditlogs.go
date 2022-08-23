package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/graph/model"
)

func (r *auditLogResolver) ActorEmail(ctx context.Context, obj *db.AuditLog) (*string, error) {
	return &obj.ActorEmail.String, nil
}

func (r *auditLogResolver) TargetUserEmail(ctx context.Context, obj *db.AuditLog) (*string, error) {
	var targetUserEmail *string
	if obj.TargetUserEmail.String != "" {
		targetUserEmail = &obj.TargetUserEmail.String
	}
	return targetUserEmail, nil
}

func (r *auditLogResolver) TargetTeamSlug(ctx context.Context, obj *db.AuditLog) (*string, error) {
	var targetTeamSlug *string
	if obj.TargetTeamSlug.String != "" {
		targetTeamSlug = &obj.TargetTeamSlug.String
	}
	return targetTeamSlug, nil
}

func (r *queryResolver) AuditLogs(ctx context.Context, pagination *model.Pagination, query *model.AuditLogsQuery, sort *model.AuditLogsSort) (*model.AuditLogs, error) {
	panic(fmt.Errorf("not implemented"))
}

// AuditLog returns generated.AuditLogResolver implementation.
func (r *Resolver) AuditLog() generated.AuditLogResolver { return &auditLogResolver{r} }

type auditLogResolver struct{ *Resolver }
