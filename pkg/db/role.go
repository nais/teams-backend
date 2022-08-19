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
