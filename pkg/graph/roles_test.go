package graph_test

import (
	"context"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/graph"
	"github.com/nais/console/pkg/graph/model"
	"github.com/nais/console/pkg/reconcilers"
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
	resolver := graph.NewResolver(db, "example.com", system, ch, logger).Query()

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

func TestQueryResolver_Authorizations(t *testing.T) {
	db, _ := test.GetTestDB()
	emptyRole := &dbmodels.Role{Name: "Empty role"}
	db.Create(emptyRole)

	roleWithAuthorizations := &dbmodels.Role{Name: "Non-empty role", Authorizations: []dbmodels.Authorization{
		{Name: "authz-1"},
		{Name: "authz-2"},
	}}
	db.Create(roleWithAuthorizations)
	system := &dbmodels.System{}
	db.Create(system)

	ch := make(chan reconcilers.Input, 100)
	ctx := context.Background()

	logger := auditlogger.New(db)
	resolver := graph.NewResolver(db, "example.com", system, ch, logger).Role()

	t.Run("Role with no authorizations", func(t *testing.T) {
		authorizations, err := resolver.Authorizations(ctx, emptyRole)
		assert.NoError(t, err)
		assert.Len(t, authorizations, 0)
	})

	t.Run("Role with authorizations", func(t *testing.T) {
		authorizations, err := resolver.Authorizations(ctx, roleWithAuthorizations)
		assert.NoError(t, err)
		assert.Len(t, authorizations, 2)
	})
}
