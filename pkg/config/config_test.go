package config_test

import (
	"testing"

	"github.com/nais/teams-backend/pkg/config"
	"github.com/nais/teams-backend/pkg/gcp"
	"github.com/stretchr/testify/assert"
)

func TestParseEnvironments(t *testing.T) {
	newCfg := func() *config.Config {
		return &config.Config{
			GCP: config.GCP{
				Clusters: map[string]gcp.Cluster{
					"dev":      gcp.Cluster{},
					"dev-gcp":  gcp.Cluster{},
					"prod":     gcp.Cluster{},
					"prod-gcp": gcp.Cluster{},
				},
			},
			OnpremClusters: []string{"dev-fss", "prod-fss", "dev-sbs", "prod-sbs"},
		}
	}

	t.Run("ignore gcp env en works", func(t *testing.T) {
		cfg := newCfg()
		cfg.IgnoredEnvironments = []string{"dev", "prod"}
		cfg.ParseEnvironments()

		assert.NotContains(t, cfg.Environments, "dev")
		assert.NotContains(t, cfg.Environments, "prod")
		assert.Contains(t, cfg.Environments, "dev-gcp")
		assert.Contains(t, cfg.Environments, "prod-gcp")

		assert.NotContains(t, cfg.GCP.Clusters, "dev")
		assert.NotContains(t, cfg.GCP.Clusters, "prod")
		assert.Contains(t, cfg.GCP.Clusters, "dev-gcp")
		assert.Contains(t, cfg.GCP.Clusters, "prod-gcp")
	})

	t.Run("ignore onprem env en works", func(t *testing.T) {
		cfg := newCfg()
		cfg.IgnoredEnvironments = []string{"dev-sbs", "prod-sbs"}
		cfg.ParseEnvironments()

		assert.NotContains(t, cfg.Environments, "dev-sbs")
		assert.NotContains(t, cfg.Environments, "prod-sbs")
		assert.Contains(t, cfg.Environments, "dev-fss")
		assert.Contains(t, cfg.Environments, "prod-fss")

		assert.NotContains(t, cfg.OnpremClusters, "dev-sbs")
		assert.NotContains(t, cfg.OnpremClusters, "prod-sbs")
		assert.Contains(t, cfg.OnpremClusters, "dev-fss")
		assert.Contains(t, cfg.OnpremClusters, "prod-fss")
	})

	t.Run("ignore all envs en works", func(t *testing.T) {
		cfg := newCfg()
		cfg.IgnoredEnvironments = []string{"dev", "prod", "dev-sbs", "prod-sbs"}
		cfg.ParseEnvironments()

		assert.NotContains(t, cfg.Environments, "dev-sbs")
		assert.NotContains(t, cfg.Environments, "prod-sbs")
		assert.Contains(t, cfg.Environments, "dev-fss")
		assert.Contains(t, cfg.Environments, "prod-fss")

		assert.NotContains(t, cfg.Environments, "dev")
		assert.NotContains(t, cfg.Environments, "prod")
		assert.Contains(t, cfg.Environments, "dev-gcp")
		assert.Contains(t, cfg.Environments, "prod-gcp")

		assert.NotContains(t, cfg.OnpremClusters, "dev-sbs")
		assert.NotContains(t, cfg.OnpremClusters, "prod-sbs")
		assert.Contains(t, cfg.OnpremClusters, "dev-fss")
		assert.Contains(t, cfg.OnpremClusters, "prod-fss")

		assert.NotContains(t, cfg.GCP.Clusters, "dev")
		assert.NotContains(t, cfg.GCP.Clusters, "prod")
		assert.Contains(t, cfg.GCP.Clusters, "dev-gcp")
		assert.Contains(t, cfg.GCP.Clusters, "prod-gcp")
	})

}
