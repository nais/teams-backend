package graph_test

import (
	"context"
	"github.com/nais/console/pkg/authz"
	helpers "github.com/nais/console/pkg/console"
	"github.com/nais/console/pkg/graph"
	"github.com/nais/console/pkg/graph/model"
	"github.com/nais/console/pkg/test"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/dbmodels"
)

func getContextWithAddedRole(system *dbmodels.System, resource authz.Resource, accessLevel string) context.Context {
	roles := []*dbmodels.RoleBinding{
		{
			Role: &dbmodels.Role{
				Resource:    string(resource),
				AccessLevel: accessLevel,
				SystemID:    system.ID,
			},
		},
	}

	return authz.ContextWithRoleBindings(context.Background(), roles)
}

func getSystem() *dbmodels.System {
	systemId, _ := uuid.NewUUID()
	system := &dbmodels.System{
		Model: dbmodels.Model{
			ID: &systemId,
		},
	}
	return system
}

func TestQueryResolver_Teams(t *testing.T) {
	db := test.GetTestDB()
	db.AutoMigrate(&dbmodels.Team{})
	db.Create([]dbmodels.Team{
		{
			Slug: dbmodels.SlugP("b"),
			Name: helpers.Strp("B"),
		},
		{
			Slug: dbmodels.SlugP("a"),
			Name: helpers.Strp("A"),
		},
		{
			Slug: dbmodels.SlugP("c"),
			Name: helpers.Strp("C"),
		},
	})

	ch := make(chan *dbmodels.Team, 100)
	system := getSystem()

	ctx := getContextWithAddedRole(system, graph.ResourceTeams, authz.AccessRead)
	resolver := graph.NewResolver(db, system, ch).Query()

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
