package graph_test

import (
	"context"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/graph"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/sqlc"
	"github.com/nais/console/pkg/test"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/nais/console/pkg/dbmodels"
)

func TestQueryResolver_Systems(t *testing.T) {
	db, _ := test.GetTestDB()
	dbc, _ := db.DB()
	queries := sqlc.New(dbc)

	ctx := context.Background()
	queries.CreateSystem(ctx, "B")
	queries.CreateSystem(ctx, "A")
	queries.CreateSystem(ctx, "C")

	ch := make(chan reconcilers.Input, 100)

	system := &dbmodels.System{}
	db.First(system)

	logger := auditlogger.New(db)
	resolver := graph.NewResolver(queries, db, "example.com", system, ch, logger).Query()

	t.Run("Get systems", func(t *testing.T) {
		systems, err := resolver.Systems(ctx)
		assert.NoError(t, err)

		assert.Len(t, systems, 3)
		assert.Equal(t, "A", systems[0].Name)
		assert.Equal(t, "B", systems[1].Name)
		assert.Equal(t, "C", systems[2].Name)
	})
}
