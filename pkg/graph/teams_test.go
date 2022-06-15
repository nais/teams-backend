package graph_test

import (
	"context"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/graph"
	"github.com/nais/console/pkg/graph/model"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/test"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/dbmodels"
)

func getContextWithAddedRoleBinding(system *dbmodels.System, resource authz.Resource, accessLevel authz.AccessLevel, permission string) context.Context {
	roles := []*dbmodels.RoleBinding{
		{
			Role: dbmodels.Role{
				Resource:    string(resource),
				AccessLevel: string(accessLevel),
				Permission:  permission,
				SystemID:    *system.ID,
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

	ch := make(chan reconcilers.ReconcileTeamInput, 100)
	system := getSystem()

	ctx := getContextWithAddedRoleBinding(system, authz.ResourceTeams, authz.AccessLevelRead, authz.PermissionAllow)
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

func TestQueryResolver_TeamsNoPermission(t *testing.T) {
	resolver := graph.NewResolver(test.GetTestDB(), "example.com", getSystem(), make(chan reconcilers.ReconcileTeamInput, 100), nil).Query()
	_, err := resolver.Teams(context.Background(), nil, nil, nil)
	assert.EqualError(t, err, "unauthorized")
}
