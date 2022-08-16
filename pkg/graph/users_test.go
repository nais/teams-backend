package graph_test

import (
	"context"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/graph"
	"github.com/nais/console/pkg/graph/model"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/sqlc"
	"github.com/nais/console/pkg/test"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/nais/console/pkg/dbmodels"
)

func TestQueryResolver_Users(t *testing.T) {
	db, _ := test.GetTestDBWithRoles()
	users := []*dbmodels.User{
		{Name: "A", Email: "b@example.com"},
		{Name: "B", Email: "a@example.com"},
		{Name: "C", Email: "c@example.com"},
	}
	system := &dbmodels.System{}
	db.Create(users)
	db.Create(system)

	ch := make(chan reconcilers.Input, 100)

	adminUser := getAdminUser(db, "D", "d@example.com")
	ctx := authz.ContextWithUser(context.Background(), adminUser)

	logger := auditlogger.New(db)
	dbc, _ := db.DB()
	resolver := graph.NewResolver(sqlc.New(dbc), db, "example.com", system, ch, logger).Query()

	t.Run("No filter or sort", func(t *testing.T) {
		users, err := resolver.Users(ctx, nil, nil, nil)
		assert.NoError(t, err)

		assert.Len(t, users.Nodes, 4)
		assert.Equal(t, "A", users.Nodes[0].Name)
		assert.Equal(t, "B", users.Nodes[1].Name)
		assert.Equal(t, "C", users.Nodes[2].Name)
		assert.Equal(t, "D", users.Nodes[3].Name)
	})

	t.Run("Sort name DESC", func(t *testing.T) {
		users, err := resolver.Users(ctx, nil, nil, &model.UsersSort{
			Field:     model.UserSortFieldName,
			Direction: model.SortDirectionDesc,
		})
		assert.NoError(t, err)

		assert.Len(t, users.Nodes, 4)
		assert.Equal(t, "D", users.Nodes[0].Name)
		assert.Equal(t, "C", users.Nodes[1].Name)
		assert.Equal(t, "B", users.Nodes[2].Name)
		assert.Equal(t, "A", users.Nodes[3].Name)
	})
}

func TestUserResolver_Roles(t *testing.T) {
	db, _ := test.GetTestDB()
	role := &dbmodels.Role{Name: "Some role"}
	userWithRoles := &dbmodels.User{Email: "a@example.com"}
	userWithNoRoles := &dbmodels.User{Email: "b@example.com"}
	system := &dbmodels.System{}
	db.Create(role)
	db.Create(userWithRoles)
	db.Create(userWithNoRoles)
	db.Create(system)
	roleBinding := &dbmodels.UserRole{UserID: *userWithRoles.ID, RoleID: *role.ID}
	db.Create(roleBinding)

	ch := make(chan reconcilers.Input, 100)
	dbc, _ := db.DB()
	resolver := graph.NewResolver(sqlc.New(dbc), db, "example.com", system, ch, auditlogger.New(db)).User()
	ctx := context.Background()

	t.Run("No user in context", func(t *testing.T) {
		userRoles, err := resolver.Roles(ctx, userWithNoRoles)
		assert.ErrorIs(t, err, authz.ErrNotAuthorized)
		assert.Nil(t, userRoles)
	})

	t.Run("Not authorized", func(t *testing.T) {
		userRoles, err := resolver.Roles(authz.ContextWithUser(ctx, userWithRoles), userWithNoRoles)
		assert.ErrorIs(t, err, authz.ErrNotAuthorized)
		assert.Nil(t, userRoles)
	})

	t.Run("No roles", func(t *testing.T) {
		userRoles, err := resolver.Roles(authz.ContextWithUser(ctx, userWithNoRoles), userWithNoRoles)
		assert.NoError(t, err)
		assert.Len(t, userRoles, 0)
	})

	t.Run("User with roles", func(t *testing.T) {
		userRoles, err := resolver.Roles(authz.ContextWithUser(ctx, userWithRoles), userWithRoles)
		assert.NoError(t, err)
		assert.Len(t, userRoles, 1)
	})
}
