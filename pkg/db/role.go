package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/sqlc"
)

func (d *database) AssignGlobalRoleToUser(ctx context.Context, userID uuid.UUID, roleName sqlc.RoleName) error {
	return d.querier.CreateUserRole(ctx, sqlc.CreateUserRoleParams{
		UserID:   userID,
		RoleName: roleName,
	})
}

func (d *database) AssignTargetedRoleToUser(ctx context.Context, userID uuid.UUID, roleName sqlc.RoleName, targetID uuid.UUID) error {
	return d.querier.CreateUserRole(ctx, sqlc.CreateUserRoleParams{
		UserID:   userID,
		RoleName: roleName,
		TargetID: nullUUID(&targetID),
	})
}

func (d *database) GetRoleNames() []sqlc.RoleName {
	return sqlc.AllRoleNameValues()
}

// IsGlobal Check if the role is globally assigned or not
func (r Role) IsGlobal() bool {
	return !r.TargetID.Valid
}

// Targets Check if the role targets a specific ID
func (r Role) Targets(targetID uuid.UUID) bool {
	return r.TargetID.Valid && r.TargetID.UUID == targetID
}
