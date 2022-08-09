package graph_test

import (
	"context"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/graph"
	"github.com/nais/console/pkg/graph/model"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/roles"
	"github.com/nais/console/pkg/test"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/dbmodels"
)

func getSystem() *dbmodels.System {
	systemId, _ := uuid.NewUUID()
	system := &dbmodels.System{
		Model: dbmodels.Model{
			ID: &systemId,
		},
	}
	return system
}

func getAdminUser(db *gorm.DB) *dbmodels.User {
	role := &dbmodels.Role{}
	db.Where("name = ?", roles.RoleAdmin).Find(role)

	user := &dbmodels.User{
		Email: "user@example.com",
		Name:  "User",
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
	system := getSystem()
	user := getAdminUser(db)

	ctx := authz.ContextWithUser(context.Background(), user)
	resolver := graph.NewResolver(db, "example.com", system, ch, nil).Query()

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
