package graph_test

import (
	"context"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/graph"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/sqlc"
	"github.com/nais/console/pkg/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQueryResolver_Systems(t *testing.T) {
	ctx := context.Background()
	db, queries, _ := test.GetTestDBAndQueries()

	id1, _ := uuid.NewUUID()
	id2, _ := uuid.NewUUID()
	id3, _ := uuid.NewUUID()

	system, _ := queries.CreateSystem(ctx, sqlc.CreateSystemParams{ID: id1, Name: "B"})
	queries.CreateSystem(ctx, sqlc.CreateSystemParams{ID: id2, Name: "C"})
	queries.CreateSystem(ctx, sqlc.CreateSystemParams{ID: id3, Name: "A"})

	ch := make(chan reconcilers.Input, 100)

	logger := auditlogger.New(db)
	resolver := graph.NewResolver(queries, db, "example.com", *system, ch, logger).Query()

	t.Run("Get systems", func(t *testing.T) {
		systems, err := resolver.Systems(ctx)
		assert.NoError(t, err)

		assert.Len(t, systems, 3)
		assert.Equal(t, "A", systems[0].Name)
		assert.Equal(t, "B", systems[1].Name)
		assert.Equal(t, "C", systems[2].Name)
	})
}
