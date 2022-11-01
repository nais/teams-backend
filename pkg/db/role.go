package db

import (
	"context"

	"github.com/nais/console/pkg/slug"

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

// IsGlobal Check if the role is globally assigned or not
func (r Role) IsGlobal() bool {
	return r.TargetServiceAccountID == nil && r.TargetTeamSlug == nil
}

// TargetsTeam Check if the role targets a specific team
func (r Role) TargetsTeam(targetsTeamSlug slug.Slug) bool {
	return r.TargetTeamSlug != nil && *r.TargetTeamSlug == targetsTeamSlug
}

// TargetsServiceAccount Check if the role targets a specific service account
func (r Role) TargetsServiceAccount(targetServiceAccountID uuid.UUID) bool {
	return r.TargetServiceAccountID != nil && *r.TargetServiceAccountID == targetServiceAccountID
}

func (d *database) UserIsTeamOwner(ctx context.Context, userID uuid.UUID, teamSlug slug.Slug) (bool, error) {
	roles, err := d.querier.GetUserRoles(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, role := range roles {
		if role.RoleName == sqlc.RoleNameTeamowner && role.TargetTeamSlug != nil && *role.TargetTeamSlug == teamSlug {
			return true, nil
		}
	}

	return false, nil
}

func (d *database) SetTeamMemberRole(ctx context.Context, userID uuid.UUID, teamSlug slug.Slug, role sqlc.RoleName) error {
	return d.querier.Transaction(ctx, func(ctx context.Context, querier Querier) error {
		err := querier.RemoveUserFromTeam(ctx, sqlc.RemoveUserFromTeamParams{
			UserID:         userID,
			TargetTeamSlug: &teamSlug,
		})
		if err != nil {
			return err
		}

		err = querier.AssignTeamRoleToUser(ctx, sqlc.AssignTeamRoleToUserParams{
			UserID:         userID,
			TargetTeamSlug: &teamSlug,
			RoleName:       role,
		})
		if err != nil {
			return err
		}

		return nil
	})
}

func (d *database) roleFromRoleBinding(ctx context.Context, roleName sqlc.RoleName, targetServiceAccountID uuid.NullUUID, targetTeamSlug *slug.Slug) (*Role, error) {
	authorizations, err := d.querier.GetRoleAuthorizations(ctx, roleName)
	if err != nil {
		return nil, err
	}

	var saID *uuid.UUID
	if targetServiceAccountID.Valid {
		saID = &targetServiceAccountID.UUID
	}

	return &Role{
		Authorizations:         authorizations,
		RoleName:               roleName,
		TargetServiceAccountID: saID,
		TargetTeamSlug:         targetTeamSlug,
	}, nil
}
