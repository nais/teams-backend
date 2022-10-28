package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/sqlc"
)

func (d *database) AssignGlobalRoleToUser(ctx context.Context, userID uuid.UUID, roleName sqlc.RoleName) error {
	return d.querier.AssignGlobalRoleToUser(ctx, sqlc.AssignGlobalRoleToUserParams{
		UserID:   userID,
		RoleName: roleName,
	})
}

func (d *database) AssignGlobalRoleToServiceAccount(ctx context.Context, serviceAccountID uuid.UUID, roleName sqlc.RoleName) error {
	return d.querier.AssignGlobalRoleToServiceAccount(ctx, sqlc.AssignGlobalRoleToServiceAccountParams{
		ServiceAccountID: serviceAccountID,
		RoleName:         roleName,
	})
}

func (d *database) RevokeGlobalRoleFromUser(ctx context.Context, userID uuid.UUID, roleName sqlc.RoleName) error {
	return d.querier.RevokeGlobalRoleFromUser(ctx, sqlc.RevokeGlobalRoleFromUserParams{
		UserID:   userID,
		RoleName: roleName,
	})
}

func (d *database) AssignTargetedRoleToUser(ctx context.Context, userID uuid.UUID, roleName sqlc.RoleName, targetID uuid.UUID) error {
	return d.querier.AssignTargetedRoleToUser(ctx, sqlc.AssignTargetedRoleToUserParams{
		UserID:   userID,
		RoleName: roleName,
		TargetID: nullUUID(&targetID),
	})
}

// IsGlobal Check if the role is globally assigned or not
func (r Role) IsGlobal() bool {
	return r.TargetID == nil
}

// Targets Check if the role targets a specific ID
func (r Role) Targets(targetID uuid.UUID) bool {
	return r.TargetID != nil && *r.TargetID == targetID
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
	return d.querier.Transaction(ctx, func(ctx context.Context, querier Querier) error {
		err := querier.RevokeTargetedRoleFromUser(ctx, sqlc.RevokeTargetedRoleFromUserParams{
			UserID:   userID,
			TargetID: nullUUID(&teamID),
			RoleName: sqlc.RoleNameTeamowner,
		})
		if err != nil {
			return err
		}

		err = querier.RevokeTargetedRoleFromUser(ctx, sqlc.RevokeTargetedRoleFromUserParams{
			UserID:   userID,
			TargetID: nullUUID(&teamID),
			RoleName: sqlc.RoleNameTeammember,
		})
		if err != nil {
			return err
		}

		err = querier.AssignTargetedRoleToUser(ctx, sqlc.AssignTargetedRoleToUserParams{
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

func (d *database) roleFromRoleBinding(ctx context.Context, roleName sqlc.RoleName, targetID uuid.NullUUID) (*Role, error) {
	authorizations, err := d.querier.GetRoleAuthorizations(ctx, roleName)
	if err != nil {
		return nil, err
	}

	var id *uuid.UUID
	if targetID.Valid {
		id = &targetID.UUID
	}

	return &Role{
		Authorizations: authorizations,
		RoleName:       roleName,
		TargetID:       id,
	}, nil
}
