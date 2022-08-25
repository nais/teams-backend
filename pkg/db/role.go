package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/sqlc"
)

func (d *database) AssignGlobalRoleToUser(ctx context.Context, userID uuid.UUID, roleName sqlc.RoleName) error {
	return d.querier.AddGlobalUserRole(ctx, sqlc.AddGlobalUserRoleParams{
		UserID:   userID,
		RoleName: roleName,
	})
}

func (d *database) AssignTargetedRoleToUser(ctx context.Context, userID uuid.UUID, roleName sqlc.RoleName, targetID uuid.UUID) error {
	return d.querier.AddTargetedUserRole(ctx, sqlc.AddTargetedUserRoleParams{
		UserID:   userID,
		RoleName: roleName,
		TargetID: nullUUID(&targetID),
	})
}

// IsGlobal Check if the role is globally assigned or not
func (r Role) IsGlobal() bool {
	return !r.TargetID.Valid
}

// Targets Check if the role targets a specific ID
func (r Role) Targets(targetID uuid.UUID) bool {
	return r.TargetID.Valid && r.TargetID.UUID == targetID
}

func (d *database) UserIsTeamOwner(ctx context.Context, userID, teamID uuid.UUID) (bool, error) {
	roles, err := d.querier.GetUserRoles(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, role := range roles {
		if role.RoleName == sqlc.RoleNameTeamowner && role.TargetID.UUID == teamID {
			return true, nil
		}
	}

	return false, nil
}

func (d *database) SetTeamMemberRole(ctx context.Context, userID uuid.UUID, teamID uuid.UUID, role sqlc.RoleName) error {
	return d.querier.Transaction(ctx, func(querier Querier) error {
		err := querier.RemoveTargetedUserRole(ctx, sqlc.RemoveTargetedUserRoleParams{
			UserID:   userID,
			TargetID: nullUUID(&teamID),
			RoleName: sqlc.RoleNameTeamowner,
		})
		if err != nil {
			return err
		}

		err = querier.RemoveTargetedUserRole(ctx, sqlc.RemoveTargetedUserRoleParams{
			UserID:   userID,
			TargetID: nullUUID(&teamID),
			RoleName: sqlc.RoleNameTeammember,
		})
		if err != nil {
			return err
		}

		err = querier.AddTargetedUserRole(ctx, sqlc.AddTargetedUserRoleParams{
			UserID:   userID,
			TargetID: nullUUID(&teamID),
			RoleName: role,
		})
		if err != nil {
			return err
		}

		return nil
	})
}
