package roles_test

import (
	"github.com/google/uuid"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/fixtures"
	"github.com/nais/console/pkg/roles"
	"github.com/nais/console/pkg/test"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"
)

func TestRequireGlobalAuthorization(t *testing.T) {
	t.Run("Nil user", func(t *testing.T) {
		assert.ErrorIs(t, roles.RequireGlobalAuthorization(nil, roles.AuthorizationTeamsCreate), roles.ErrNotAuthorized)
	})

	t.Run("User with no roles", func(t *testing.T) {
		db := getDb()
		user := getUserWithRoles(db, []roles.Role{})
		assert.ErrorIs(t, roles.RequireGlobalAuthorization(user, roles.AuthorizationTeamsCreate), roles.ErrNotAuthorized)
	})

	t.Run("User with insufficient roles", func(t *testing.T) {
		db := getDb()
		user := getUserWithRoles(db, []roles.Role{roles.RoleTeamViewer})
		assert.ErrorIs(t, roles.RequireGlobalAuthorization(user, roles.AuthorizationTeamsCreate), roles.ErrNotAuthorized)
	})

	t.Run("User with sufficient role", func(t *testing.T) {
		db := getDb()
		user := getUserWithRoles(db, []roles.Role{roles.RoleTeamCreator})
		assert.NoError(t, roles.RequireGlobalAuthorization(user, roles.AuthorizationTeamsCreate))
	})
}

func TestRequireAuthorizationForTarget(t *testing.T) {
	targetId, _ := uuid.NewUUID()

	t.Run("Nil user", func(t *testing.T) {
		assert.ErrorIs(t, roles.RequireAuthorization(nil, roles.AuthorizationTeamsCreate, targetId), roles.ErrNotAuthorized)
	})

	t.Run("User with no roles", func(t *testing.T) {
		db := getDb()
		user := getUserWithRoles(db, []roles.Role{})
		assert.ErrorIs(t, roles.RequireAuthorization(user, roles.AuthorizationTeamsCreate, targetId), roles.ErrNotAuthorized)
	})

	t.Run("User with insufficient roles", func(t *testing.T) {
		db := getDb()
		user := getUserWithRoles(db, []roles.Role{roles.RoleTeamViewer})
		assert.ErrorIs(t, roles.RequireAuthorization(user, roles.AuthorizationTeamsUpdate, targetId), roles.ErrNotAuthorized)
	})

	t.Run("User with targetted role", func(t *testing.T) {
		db := getDb()
		user := getUserWithTargettedRole(db, roles.RoleTeamOwner, targetId)
		assert.NoError(t, roles.RequireAuthorization(user, roles.AuthorizationTeamsUpdate, targetId))
	})

	t.Run("User with targetted role for wrong target", func(t *testing.T) {
		wrongId, _ := uuid.NewUUID()
		db := getDb()
		user := getUserWithTargettedRole(db, roles.RoleTeamOwner, wrongId)
		assert.ErrorIs(t, roles.RequireAuthorization(user, roles.AuthorizationTeamsUpdate, targetId), roles.ErrNotAuthorized)
	})

	t.Run("User with global role", func(t *testing.T) {
		db := getDb()
		user := getUserWithRoles(db, []roles.Role{roles.RoleTeamOwner})
		assert.NoError(t, roles.RequireAuthorization(user, roles.AuthorizationTeamsUpdate, targetId))
	})
}

func getDb() *gorm.DB {
	db := test.GetTestDB()
	db.AutoMigrate(&dbmodels.Authorization{}, &dbmodels.Role{}, &dbmodels.RoleAuthorization{}, &dbmodels.User{}, &dbmodels.UserRole{})
	fixtures.CreateRolesAndAuthorizations(db)
	return db
}

func getUserWithTargettedRole(db *gorm.DB, roleName roles.Role, targetId uuid.UUID) *dbmodels.User {
	role := &dbmodels.Role{}
	db.Where("name = ?", roleName).Find(role)

	user := &dbmodels.User{Email: "user@example.com"}
	db.Create(user)

	db.Create(&dbmodels.UserRole{UserID: *user.ID, RoleID: *role.ID, TargetID: &targetId})

	db.
		Model(user).
		Preload("Role").
		Preload("Role.Authorizations").
		Association("RoleBindings").
		Find(&user.RoleBindings)
	return user
}

func getUserWithRoles(db *gorm.DB, roleNames []roles.Role) *dbmodels.User {
	roles := make([]*dbmodels.Role, 0)
	db.Where("name IN (?)", roleNames).Find(&roles)

	user := &dbmodels.User{Email: "user@example.com"}
	db.Create(user)

	for _, role := range roles {
		db.Create(&dbmodels.UserRole{UserID: *user.ID, RoleID: *role.ID})
	}

	db.
		Model(user).
		Preload("Role").
		Preload("Role.Authorizations").
		Association("RoleBindings").
		Find(&user.RoleBindings)
	return user
}
