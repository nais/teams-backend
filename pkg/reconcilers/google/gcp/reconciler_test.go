package google_gcp_reconciler_test

import (
	"testing"

	"github.com/nais/console/pkg/reconcilers/google/gcp"
	"github.com/stretchr/testify/assert"
)

func TestGenerateProjectID(t *testing.T) {
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

func TestGetClusterInfoFromJson(t *testing.T) {
	t.Run("empty string", func(t *testing.T) {
		info, err := google_gcp_reconciler.GetClusterInfoFromJson("")
		assert.Nil(t, info)
		assert.EqualError(t, err, "parse GCP cluster info: EOF")
	})

	t.Run("empty JSON object", func(t *testing.T) {
		info, err := google_gcp_reconciler.GetClusterInfoFromJson("{}")
		assert.NoError(t, err)
		assert.Empty(t, info)
	})

	t.Run("JSON with clusters", func(t *testing.T) {
		jsonData := `{
			"env1": {"team_folder_id": 123, "cluster_project_id": "some-id-123"},
			"env2": {"team_folder_id": 456, "cluster_project_id": "some-id-456"}
		}`
		info, err := google_gcp_reconciler.GetClusterInfoFromJson(jsonData)
		assert.NoError(t, err)

		assert.Contains(t, info, "env1")
		assert.Equal(t, int64(123), info["env1"].TeamFolderID)
		assert.Equal(t, "some-id-123", info["env1"].ProjectID)

		assert.Contains(t, info, "env2")
		assert.Equal(t, int64(456), info["env2"].TeamFolderID)
		assert.Equal(t, "some-id-456", info["env2"].ProjectID)
	})
}
