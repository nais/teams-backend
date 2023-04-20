package dependencytrack

import (
	"os"
	"testing"

	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/assert"
)

type Config struct {
	DependencyTrack DependencyTrack
}

type DependencyTrack struct {
	// One dependency track instance for each cluster.
	Instances Instances `envconfig:"CONSOLE_DEPENDENCYTRACK"`
}

func TestEnvconfigToInstances(t *testing.T) {
	err := os.Setenv("CONSOLE_DEPENDENCYTRACK", `[{"endpoint": "https://dependencytrack.nais.io", "username": "admin", "password": "password"}]`)
	assert.NoError(t, err)
	var cfg Config
	err = envconfig.Process("myapp", &cfg)
	assert.NoError(t, err)
	assert.Equal(t, Instances{{
		Endpoint: "https://dependencytrack.nais.io",
		Username: "admin",
		Password: "password",
	}}, cfg.DependencyTrack.Instances)

}
