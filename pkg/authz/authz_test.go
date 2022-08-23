package authz_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/sqlc"
	"github.com/stretchr/testify/assert"
)

func userWithNoRoles() *db.User {
	return &db.User{
		User: &sqlc.User{
			Name:  "User Name",
			Email: "mail@example.com",
		},
	}
}

func userWithGlobalRoleAndAuthorizations(roleName sqlc.RoleName, authorizations []sqlc.AuthzName) *db.User {
	return &db.User{
		User: &sqlc.User{
			Name:  "User Name",
			Email: "mail@example.com",
		},
		Roles: []*db.Role{
			{
				UserRole:       &sqlc.UserRole{TargetID: uuid.NullUUID{}},
				Name:           roleName,
				Authorizations: authorizations,
			},
		},
	}
}

func userWithTargetedRoleAndAuthorizations(roleName sqlc.RoleName, authorizations []sqlc.AuthzName, targetID uuid.UUID) *db.User {
	user := userWithGlobalRoleAndAuthorizations(roleName, authorizations)
	user.Roles[0].TargetID = uuid.NullUUID{
		UUID:  targetID,
		Valid: true,
	}
	return user
}

func TestContextWithUser(t *testing.T) {
	ctx := context.Background()
	assert.Nil(t, authz.UserFromContext(ctx))

	user := &db.User{
		User: &sqlc.User{
			Name:  "User Name",
			Email: "mail@example.com",
		},
	}

	ctx = authz.ContextWithUser(ctx, user)
	assert.Equal(t, user, authz.UserFromContext(ctx))
}

func TestRequireGlobalAuthorization(t *testing.T) {
	t.Run("Nil user", func(t *testing.T) {
		assert.ErrorIs(t, authz.RequireGlobalAuthorization(nil, sqlc.AuthzNameTeamsCreate), authz.ErrNotAuthorized)
	})

	t.Run("User with no roles", func(t *testing.T) {
		user := userWithNoRoles()
		assert.ErrorIs(t, authz.RequireGlobalAuthorization(user, sqlc.AuthzNameTeamsCreate), authz.ErrNotAuthorized)
	})

	t.Run("User with insufficient roles", func(t *testing.T) {
		user := userWithGlobalRoleAndAuthorizations(sqlc.RoleNameTeamviewer, []sqlc.AuthzName{})
		assert.ErrorIs(t, authz.RequireGlobalAuthorization(user, sqlc.AuthzNameTeamsCreate), authz.ErrNotAuthorized)
	})

	t.Run("User with sufficient role", func(t *testing.T) {
		user := userWithGlobalRoleAndAuthorizations(sqlc.RoleNameTeamcreator, []sqlc.AuthzName{
			sqlc.AuthzNameTeamsCreate,
		})
		assert.NoError(t, authz.RequireGlobalAuthorization(user, sqlc.AuthzNameTeamsCreate))
	})
}

func TestRequireAuthorizationForTarget(t *testing.T) {
	targetID, _ := uuid.NewUUID()

	t.Run("Nil user", func(t *testing.T) {
		assert.ErrorIs(t, authz.RequireAuthorization(nil, sqlc.AuthzNameTeamsCreate, targetID), authz.ErrNotAuthorized)
	})

	t.Run("User with no roles", func(t *testing.T) {
		user := userWithNoRoles()
		assert.ErrorIs(t, authz.RequireAuthorization(user, sqlc.AuthzNameTeamsCreate, targetID), authz.ErrNotAuthorized)
	})

	t.Run("User with insufficient roles", func(t *testing.T) {
		user := userWithGlobalRoleAndAuthorizations(sqlc.RoleNameTeamviewer, []sqlc.AuthzName{})
		assert.ErrorIs(t, authz.RequireAuthorization(user, sqlc.AuthzNameTeamsUpdate, targetID), authz.ErrNotAuthorized)
	})

	t.Run("User with targeted role", func(t *testing.T) {
		user := userWithTargetedRoleAndAuthorizations(sqlc.RoleNameTeamowner, []sqlc.AuthzName{sqlc.AuthzNameTeamsUpdate}, targetID)
		assert.NoError(t, authz.RequireAuthorization(user, sqlc.AuthzNameTeamsUpdate, targetID))
	})

	t.Run("User with targeted role for wrong target", func(t *testing.T) {
		wrongID, _ := uuid.NewUUID()
		user := userWithTargetedRoleAndAuthorizations(sqlc.RoleNameTeamowner, []sqlc.AuthzName{sqlc.AuthzNameTeamsUpdate}, wrongID)
		assert.ErrorIs(t, authz.RequireAuthorization(user, sqlc.AuthzNameTeamsUpdate, targetID), authz.ErrNotAuthorized)
	})

	t.Run("User with global role", func(t *testing.T) {
		user := userWithGlobalRoleAndAuthorizations(sqlc.RoleNameTeamowner, []sqlc.AuthzName{sqlc.AuthzNameTeamsUpdate})
		assert.NoError(t, authz.RequireAuthorization(user, sqlc.AuthzNameTeamsUpdate, targetID))
	})
}
