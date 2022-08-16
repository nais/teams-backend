package graph_test

import (
	"context"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/graph"
	"github.com/nais/console/pkg/graph/model"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/roles"
	"github.com/nais/console/pkg/sqlc"
	"github.com/nais/console/pkg/test"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"

	"github.com/nais/console/pkg/dbmodels"
)

// getAdminUser Create a user object with the admin role, insert it into the provided database, and return the user
func getAdminUser(db *gorm.DB, name, email string) *dbmodels.User {
	role := &dbmodels.Role{}
	db.Where("name = ?", roles.RoleAdmin).Find(role)

	user := &dbmodels.User{
		Email: email,
		Name:  name,
	}
	db.Create(user)
	userRole := &dbmodels.UserRole{
		RoleID: *role.ID,
		UserID: *user.ID,
	}
	db.Create(userRole)
	db.
		Model(user).
		Preload("Role").
		Preload("Role.Authorizations").
		Association("RoleBindings").
		Find(&user.RoleBindings)

	return user
}

func TestQueryResolver_Teams(t *testing.T) {
	db, _ := test.GetTestDBWithRoles()
	db.Create([]dbmodels.Team{
		{
			Slug: "b",
			Name: "B",
		},
		{
			Slug: "a",
			Name: "A",
		},
		{
			Slug: "c",
			Name: "C",
		},
	})

	ch := make(chan reconcilers.Input, 100)
	system := &dbmodels.System{}
	db.Create(system)
	user := getAdminUser(db, "user", "user@example.com")

	ctx := authz.ContextWithUser(context.Background(), user)
	dbc, _ := db.DB()
	resolver := graph.NewResolver(sqlc.New(dbc), db, "example.com", system, ch, nil).Query()

	t.Run("No filter or sort", func(t *testing.T) {
		teams, err := resolver.Teams(ctx, nil, nil, nil)
		assert.NoError(t, err)

		assert.Len(t, teams.Nodes, 3)
		assert.Equal(t, "a", teams.Nodes[0].Slug.String())
		assert.Equal(t, "b", teams.Nodes[1].Slug.String())
		assert.Equal(t, "c", teams.Nodes[2].Slug.String())
	})

	t.Run("Sort name DESC", func(t *testing.T) {
		teams, err := resolver.Teams(ctx, nil, nil, &model.TeamsSort{
			Field:     model.TeamSortFieldName,
			Direction: model.SortDirectionDesc,
		})
		assert.NoError(t, err)

		assert.Len(t, teams.Nodes, 3)
		assert.Equal(t, "c", teams.Nodes[0].Slug.String())
		assert.Equal(t, "b", teams.Nodes[1].Slug.String())
		assert.Equal(t, "a", teams.Nodes[2].Slug.String())
	})
}
