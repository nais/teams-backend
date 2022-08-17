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
	systems := []*dbmodels.System{
		{Name: "B"},
		{Name: "A"},
		{Name: "C"},
	}
	db.Create(systems)

	ch := make(chan reconcilers.Input, 100)
	ctx := context.Background()

	logger := auditlogger.New(db)
	dbc, _ := db.DB()
	resolver := graph.NewResolver(sqlc.New(dbc), db, "example.com", systems[0], ch, logger).Query()

	t.Run("Get systems", func(t *testing.T) {
		systems, err := resolver.Systems(ctx)
		assert.NoError(t, err)

		assert.Len(t, systems, 3)
		assert.Equal(t, "A", systems[0].Name)
		assert.Equal(t, "B", systems[1].Name)
		assert.Equal(t, "C", systems[2].Name)
	})
}
