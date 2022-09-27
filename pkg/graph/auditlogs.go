package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/graph/generated"
)

// Actor is the resolver for the actor field.
func (r *auditLogResolver) Actor(ctx context.Context, obj *db.AuditLog) (*string, error) {
	return db.NullStringToStringP(obj.Actor), nil
}

// AuditLog returns generated.AuditLogResolver implementation.
func (r *Resolver) AuditLog() generated.AuditLogResolver { return &auditLogResolver{r} }

type auditLogResolver struct{ *Resolver }
