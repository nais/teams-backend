package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/sqlc"
)

// AddUserRole implements Database
func (d *database) AddUserRole(ctx context.Context, userID uuid.UUID, roleName sqlc.RoleName) error {
	id, err := uuid.NewUUID()
	if err != nil {
		return err
	}
	return d.querier.AddRoleToUser(ctx, sqlc.AddRoleToUserParams{
		ID:       id,
		UserID:   userID,
		RoleName: roleName,
	})
}

// IsGlobal Check if the role is globally assigned or not
func (r Role) IsGlobal() bool {
	return !r.TargetID.Valid
}

// Targets Check if the role targets a specific ID
func (r Role) Targets(targetId uuid.UUID) bool {
	return r.TargetID.Valid && r.TargetID.UUID == targetId
}
