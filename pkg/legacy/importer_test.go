//go:build run_manually_for_data_migration
// +build run_manually_for_data_migration

package legacy_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/nais/console/pkg/legacy"
)

func TestImportTeamsFromLegacyAzure(t *testing.T) {
	const ymlpath = "/Users/kimt/src/navikt/teams/teams.yml"
	const jsonpath = "/Users/kimt/src/navikt/teams/teams.json"
	teams, err := legacy.ReadTeamFiles(ymlpath, jsonpath)
	if err != nil {
		panic(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(teams)
}
