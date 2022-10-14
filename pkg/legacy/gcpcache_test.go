package legacy_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/nais/console/pkg/legacy"
)

func TestReadGcpTeamCacheFile(t *testing.T) {
	x, err := legacy.ReadGcpTeamCacheFile("/Users/kimt/src/nais/teams/gcp-projects/prod-output.json")
	if err != nil {
		panic(err)
	}
	json.NewEncoder(os.Stdout).Encode(x)
}

func TestMergeGcpTeamCacheFiles(t *testing.T) {
	dev, _ := legacy.ReadGcpTeamCacheFile("/Users/kimt/src/nais/teams/gcp-projects/dev-output.json")
	prod, _ := legacy.ReadGcpTeamCacheFile("/Users/kimt/src/nais/teams/gcp-projects/prod-output.json")
	merged := legacy.MergeGcpTeamCacheFiles(dev, prod)
	json.NewEncoder(os.Stdout).Encode(merged)
}
