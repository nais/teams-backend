package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/sqlc"
)

func (d *database) AssignGlobalRoleToUser(ctx context.Context, userID uuid.UUID, roleName sqlc.RoleName) error {
	ID, err := uuid.NewUUID()
	if err != nil {
		return err
	}
	return d.querier.CreateUserRole(ctx, sqlc.CreateUserRoleParams{
		ID:       ID,
		UserID:   userID,
		RoleName: roleName,
	})
}

func (d *database) AssignTargetedRoleToUser(ctx context.Context, userID uuid.UUID, roleName sqlc.RoleName, targetID uuid.UUID) error {
	ID, err := uuid.NewUUID()
	if err != nil {
		return err
	}
	return d.querier.CreateUserRole(ctx, sqlc.CreateUserRoleParams{
		ID:       ID,
		UserID:   userID,
		RoleName: roleName,
		TargetID: nullUUID(&targetID),
	})
}

func (d *database) GetRoleNames(ctx context.Context) ([]sqlc.RoleName, error) {
	return d.querier.GetRoleNames(ctx)
}

// IsGlobal Check if the role is globally assigned or not
func (r Role) IsGlobal() bool {
	return !r.TargetID.Valid
}

// Targets Check if the role targets a specific ID
func (r Role) Targets(targetId uuid.UUID) bool {
	return r.TargetID.Valid && r.TargetID.UUID == targetId
}
