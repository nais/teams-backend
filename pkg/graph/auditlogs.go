package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/graph/generated"
)

func (r *auditLogResolver) ActorEmail(ctx context.Context, obj *db.AuditLog) (*string, error) {
	return db.NullStringToStringP(obj.ActorEmail), nil
}

func (r *auditLogResolver) TargetUserEmail(ctx context.Context, obj *db.AuditLog) (*string, error) {
	return db.NullStringToStringP(obj.TargetUserEmail), nil
}

func (r *auditLogResolver) TargetTeamSlug(ctx context.Context, obj *db.AuditLog) (*string, error) {
	return db.NullStringToStringP(obj.TargetTeamSlug), nil
}

// AuditLog returns generated.AuditLogResolver implementation.
func (r *Resolver) AuditLog() generated.AuditLogResolver { return &auditLogResolver{r} }

type auditLogResolver struct{ *Resolver }
