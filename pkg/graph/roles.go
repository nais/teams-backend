package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/sqlc"
)

// AssignGlobalRoleToUser is the resolver for the assignGlobalRoleToUser field.
func (r *mutationResolver) AssignGlobalRoleToUser(ctx context.Context, role sqlc.RoleName, userID *uuid.UUID) (*db.User, error) {
	if !role.Valid() {
		return nil, fmt.Errorf("%q is not a valid role", role)
	}

	user, err := r.database.GetUserByID(ctx, *userID)
	if err != nil {
		return nil, err
	}

	err = r.database.AssignGlobalRoleToUser(ctx, *userID, role)
	if err != nil {
		return nil, err
	}

	actor := authz.ActorFromContext(ctx)
	targets := []auditlogger.Target{
		auditlogger.UserTarget(user.Email),
	}
	fields := auditlogger.Fields{
		Action: sqlc.AuditActionGraphqlApiRolesAssignGlobalRole,
		Actor:  actor,
	}
	r.auditLogger.Logf(ctx, targets, fields, "Assign global role %q to user", role)

	return user, nil
}

// RevokeGlobalRoleFromUser is the resolver for the revokeGlobalRoleFromUser field.
func (r *mutationResolver) RevokeGlobalRoleFromUser(ctx context.Context, role sqlc.RoleName, userID *uuid.UUID) (*db.User, error) {
	if !role.Valid() {
		return nil, fmt.Errorf("%q is not a valid role", role)
	}

	user, err := r.database.GetUserByID(ctx, *userID)
	if err != nil {
		return nil, err
	}

	err = r.database.RevokeGlobalRoleFromUser(ctx, *userID, role)
	if err != nil {
		return nil, err
	}

	actor := authz.ActorFromContext(ctx)
	targets := []auditlogger.Target{
		auditlogger.UserTarget(user.Email),
	}
	fields := auditlogger.Fields{
		Action: sqlc.AuditActionGraphqlApiRolesRevokeGlobalRole,
		Actor:  actor,
	}
	r.auditLogger.Logf(ctx, targets, fields, "Revoke global role %q from user", role)

	return user, nil
}

// Roles is the resolver for the roles field.
func (r *queryResolver) Roles(ctx context.Context) ([]sqlc.RoleName, error) {
	return sqlc.AllRoleNameValues(), nil
}

// Name is the resolver for the name field.
func (r *roleResolver) Name(ctx context.Context, obj *db.Role) (sqlc.RoleName, error) {
	return obj.RoleName, nil
}

// TargetID is the resolver for the targetId field.
func (r *roleResolver) TargetID(ctx context.Context, obj *db.Role) (*uuid.UUID, error) {
	if obj.TargetID.Valid {
		return &obj.TargetID.UUID, nil
	}
	return nil, nil
}

// Role returns generated.RoleResolver implementation.
func (r *Resolver) Role() generated.RoleResolver { return &roleResolver{r} }

type roleResolver struct{ *Resolver }
