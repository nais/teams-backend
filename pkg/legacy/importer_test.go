//go:build run_manually_for_data_migration

package legacy_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/legacy"
	log "github.com/sirupsen/logrus"
)

func TestImportTeamsFromLegacyAzure(t *testing.T) {
	const ymlpath = "/Users/kimt/src/navikt/teams/teams.yml"
	const jsonpath = "/Users/kimt/src/navikt/teams/teams.json"

	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	gimp, err := legacy.NewFromConfig(cfg)
	if err != nil {
		panic(err)
	}

	teams, err := legacy.ReadTeamFiles(ymlpath, jsonpath)
	if err != nil {
		panic(err)
	}

	dbteams := make([]*dbmodels.Team, 0, len(teams))

	for _, yamlteam := range teams {
		team := yamlteam.Convert()
		log.Infof("Fetch team info for %s...", *team.Name)
		members, err := gimp.GroupMembers(yamlteam.AzureID)
		if err != nil {
			panic(err)
		}
		team.Users = members
		dbteams = append(dbteams, team)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(dbteams)
}
