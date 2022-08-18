package graph_test

import (
	"context"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/graph"
	"github.com/nais/console/pkg/reconcilers"
	console_reconciler "github.com/nais/console/pkg/reconcilers/console"
	"github.com/nais/console/pkg/sqlc"
	"github.com/nais/console/pkg/test"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/nais/console/pkg/dbmodels"
)

func TestQueryResolver_Roles(t *testing.T) {
	ctx := context.Background()
	db, queries, _ := test.GetTestDBAndQueries()
	roles := []*dbmodels.Role{
		{Name: "B"},
		{Name: "A"},
		{Name: "C"},
	}
	system := &sqlc.System{Name: console_reconciler.Name}
	db.Create(roles)

	ch := make(chan reconcilers.Input, 100)

	logger := auditlogger.New(db)
	resolver := graph.NewResolver(queries, db, "example.com", *system, ch, logger).Query()

	t.Run("Get roles", func(t *testing.T) {
		roles, err := resolver.Roles(ctx)
		assert.NoError(t, err)

		assert.Len(t, roles, 3)
		assert.Equal(t, "A", roles[0].Name)
		assert.Equal(t, "B", roles[1].Name)
		assert.Equal(t, "C", roles[2].Name)
	})
}

func TestRoleBindingResolver_Role(t *testing.T) {
	ctx := context.Background()
	db, queries, _ := test.GetTestDBAndQueries()

	system := &sqlc.System{Name: console_reconciler.Name}
	role := &dbmodels.Role{Name: "Some role"}
	user := &dbmodels.User{Email: "user@example.com"}
	db.Create(role)
	db.Create(user)
	userRole := &sqlc.UserRole{
		UserID: *user.ID,
		RoleID: *role.ID,
	}
	db.Create(userRole)

	ch := make(chan reconcilers.Input, 100)

	resolver := graph.NewResolver(queries, db, "example.com", *system, ch, auditlogger.New(db)).RoleBinding()

	retrievedRole, err := resolver.Role(ctx, userRole)
	assert.NoError(t, err)
	assert.Equal(t, "Some role", retrievedRole.Name)
}

func TestRoleBindingResolver_IsGlobal(t *testing.T) {
	ctx := context.Background()
	db, queries, _ := test.GetTestDBAndQueries()

	team := &dbmodels.Team{Slug: "slug", Name: "name"}
	system := &sqlc.System{Name: console_reconciler.Name}
	role1 := &dbmodels.Role{Name: "Some role"}
	role2 := &dbmodels.Role{Name: "Some other role"}
	user := &dbmodels.User{Email: "user@example.com"}
	db.Create(team)
	db.Create(role1)
	db.Create(role2)
	db.Create(user)

	globalUserRole := &dbmodels.UserRole{
		UserID: *user.ID,
		RoleID: *role1.ID,
	}
	targettedUserRole := &dbmodels.UserRole{
		UserID:   *user.ID,
		RoleID:   *role2.ID,
		TargetID: team.ID,
	}
	db.Create(globalUserRole)
	db.Create(targettedUserRole)

	ch := make(chan reconcilers.Input, 100)

	resolver := graph.NewResolver(queries, db, "example.com", *system, ch, auditlogger.New(db)).RoleBinding()

	t.Run("Global role", func(t *testing.T) {
		userRole, _ := queries.GetUserRole(ctx, *globalUserRole.ID)
		isGlobal, err := resolver.IsGlobal(ctx, userRole)
		assert.NoError(t, err)
		assert.True(t, isGlobal)
	})

	t.Run("Targetted role", func(t *testing.T) {
		targettedUserRole, _ := queries.GetUserRole(ctx, *targettedUserRole.ID)
		isGlobal, err := resolver.IsGlobal(ctx, targettedUserRole)
		assert.NoError(t, err)
		assert.False(t, isGlobal)
	})
}
