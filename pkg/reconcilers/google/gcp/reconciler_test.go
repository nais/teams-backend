package google_gcp_reconciler_test

import (
	"testing"

	"github.com/nais/console/pkg/reconcilers/google/gcp"
	"github.com/stretchr/testify/assert"
)

func TestCreateProjectName(t *testing.T) {
	// different organization names don't show up in name, but are reflected in the hash
	assert.Equal(t, "happyteam-prod-488a", google_gcp_reconciler.GenerateProjectID("nais.io", "production", "happyteam"))
	assert.Equal(t, "happyteam-prod-5534", google_gcp_reconciler.GenerateProjectID("bais.io", "production", "happyteam"))

	// environments that get truncated produce different hashes
	assert.Equal(t, "sadteam-prod-04d4", google_gcp_reconciler.GenerateProjectID("nais.io", "production", "sadteam"))
	assert.Equal(t, "sadteam-prod-6ce6", google_gcp_reconciler.GenerateProjectID("nais.io", "producers", "sadteam"))

	// team names that get truncated produce different hashes
	assert.Equal(t, "happyteam-is-very-ha-prod-4b2d", google_gcp_reconciler.GenerateProjectID("bais.io", "production", "happyteam-is-very-happy"))
	assert.Equal(t, "happyteam-is-very-ha-prod-4801", google_gcp_reconciler.GenerateProjectID("bais.io", "production", "happyteam-is-very-happy-and-altogether-too-long"))
}
