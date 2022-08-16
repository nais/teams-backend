package graph_test

import (
	"context"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/graph"
	"github.com/nais/console/pkg/graph/model"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/sqlc"
	"github.com/nais/console/pkg/test"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/nais/console/pkg/dbmodels"
)

func TestQueryResolver_Roles(t *testing.T) {
	db, _ := test.GetTestDB()
	roles := []*dbmodels.Role{
		{Name: "B"},
		{Name: "A"},
		{Name: "C"},
	}
	system := &dbmodels.System{}
	db.Create(roles)
	db.Create(system)

	ch := make(chan reconcilers.Input, 100)
	ctx := context.Background()

	logger := auditlogger.New(db)
	dbc, _ := db.DB()
	resolver := graph.NewResolver(sqlc.New(dbc), db, "example.com", system, ch, logger).Query()

	t.Run("No filter or sort", func(t *testing.T) {
		roles, err := resolver.Roles(ctx, nil, nil, nil)
		assert.NoError(t, err)

		assert.Len(t, roles.Nodes, 3)
		assert.Equal(t, "A", roles.Nodes[0].Name)
		assert.Equal(t, "B", roles.Nodes[1].Name)
		assert.Equal(t, "C", roles.Nodes[2].Name)
	})

	t.Run("Sort name DESC", func(t *testing.T) {
		roles, err := resolver.Roles(ctx, nil, nil, &model.RolesSort{
			Field:     model.RoleSortFieldName,
			Direction: model.SortDirectionDesc,
		})
		assert.NoError(t, err)

		assert.Len(t, roles.Nodes, 3)
		assert.Equal(t, "C", roles.Nodes[0].Name)
		assert.Equal(t, "B", roles.Nodes[1].Name)
		assert.Equal(t, "A", roles.Nodes[2].Name)
	})
}

func TestRoleBindingResolver_Role(t *testing.T) {
	db, _ := test.GetTestDB()

	system := &dbmodels.System{}
	role := &dbmodels.Role{Name: "Some role"}
	user := &dbmodels.User{Email: "user@example.com"}
	db.Create(role)
	db.Create(user)
	db.Create(system)

	userRole := &dbmodels.UserRole{
		UserID: *user.ID,
		RoleID: *role.ID,
	}
	db.Create(userRole)

	ch := make(chan reconcilers.Input, 100)
	dbc, _ := db.DB()

	resolver := graph.NewResolver(sqlc.New(dbc), db, "example.com", system, ch, auditlogger.New(db)).RoleBinding()

	role, err := resolver.Role(context.Background(), userRole)
	assert.NoError(t, err)
	assert.Equal(t, "Some role", role.Name)
}

func TestRoleBindingResolver_IsGlobal(t *testing.T) {
	db, _ := test.GetTestDB()

	team := &dbmodels.Team{Slug: "slug", Name: "name"}
	system := &dbmodels.System{}
	role1 := &dbmodels.Role{Name: "Some role"}
	role2 := &dbmodels.Role{Name: "Some other role"}
	user := &dbmodels.User{Email: "user@example.com"}
	db.Create(team)
	db.Create(role1)
	db.Create(role2)
	db.Create(user)
	db.Create(system)

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

	dbc, _ := db.DB()
	resolver := graph.NewResolver(sqlc.New(dbc), db, "example.com", system, ch, auditlogger.New(db)).RoleBinding()
	ctx := context.Background()

	t.Run("Global role", func(t *testing.T) {
		isGlobal, err := resolver.IsGlobal(ctx, globalUserRole)
		assert.NoError(t, err)
		assert.True(t, isGlobal)
	})

	t.Run("Targetted role", func(t *testing.T) {
		isGlobal, err := resolver.IsGlobal(ctx, targettedUserRole)
		assert.NoError(t, err)
		assert.False(t, isGlobal)
	})
}
