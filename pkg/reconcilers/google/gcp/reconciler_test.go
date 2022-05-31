//go:build adhoc_integration_test

package google_gcp_reconciler_test

import (
	"context"
	helpers "github.com/nais/console/pkg/console"
	"testing"

	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/reconcilers"
	google_gcp_reconciler "github.com/nais/console/pkg/reconcilers/google/gcp"
	"github.com/stretchr/testify/assert"
)

func TestCreateProjectName(t *testing.T) {
	// different organization names don't show up in name, but are reflected in the hash
	assert.Equal(t, "happyteam-prod-488a", google_gcp_reconciler.CreateProjectID("nais.io", "production", "happyteam"))
	assert.Equal(t, "happyteam-prod-5534", google_gcp_reconciler.CreateProjectID("bais.io", "production", "happyteam"))

	// environments that get truncated produce different hashes
	assert.Equal(t, "sadteam-prod-04d4", google_gcp_reconciler.CreateProjectID("nais.io", "production", "sadteam"))
	assert.Equal(t, "sadteam-prod-6ce6", google_gcp_reconciler.CreateProjectID("nais.io", "producers", "sadteam"))

	// team names that get truncated produce different hashes
	assert.Equal(t, "happyteam-is-very-ha-prod-4b2d", google_gcp_reconciler.CreateProjectID("bais.io", "production", "happyteam-is-very-happy"))
	assert.Equal(t, "happyteam-is-very-ha-prod-4801", google_gcp_reconciler.CreateProjectID("bais.io", "production", "happyteam-is-very-happy-and-altogether-too-long"))
}

func TestGCPReconciler(t *testing.T) {
	ctx := context.Background()

	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	rec, err := google_gcp_reconciler.NewFromConfig(nil, cfg, nil)
	if err != nil {
		panic(err)
	}

	teamSlug := dbmodels.Slug("foo")
	input := reconcilers.Input{
		Team: &dbmodels.Team{
			Slug:    &teamSlug,
			Name:    helpers.Strp("Hello, World!"),
			Purpose: helpers.Strp("you know it"),
		},
	}

	err = rec.Reconcile(ctx, input)

	assert.NoError(t, err)
}
