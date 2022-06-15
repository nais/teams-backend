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

func TestQueryResolver_Systems(t *testing.T) {
	db := test.GetTestDB()
	db.AutoMigrate(&dbmodels.System{})
	db.Create([]dbmodels.System{
		{
			Name: "B",
		},
		{
			Name: "A",
		},
		{
			Name: "C",
		},
	})

	ch := make(chan reconcilers.Input, 100)
	system := getSystem()
	ctx := context.Background()

	logger := auditlogger.New(db)
	resolver := graph.NewResolver(db, "example.com", system, ch, logger).Query()

	t.Run("No filter or sort", func(t *testing.T) {
		systems, err := resolver.Systems(ctx, nil, nil, nil)
		assert.NoError(t, err)

		assert.Len(t, systems.Nodes, 3)
		assert.Equal(t, "A", systems.Nodes[0].Name)
		assert.Equal(t, "B", systems.Nodes[1].Name)
		assert.Equal(t, "C", systems.Nodes[2].Name)
	})

	t.Run("Sort name DESC", func(t *testing.T) {
		systems, err := resolver.Systems(ctx, nil, nil, &model.SystemsSort{
			Field:     model.SystemSortFieldName,
			Direction: model.SortDirectionDesc,
		})
		assert.NoError(t, err)

		assert.Len(t, systems.Nodes, 3)
		assert.Equal(t, "C", systems.Nodes[0].Name)
		assert.Equal(t, "B", systems.Nodes[1].Name)
		assert.Equal(t, "A", systems.Nodes[2].Name)
	})
}
