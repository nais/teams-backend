package legacy

import (
	"encoding/json"
	"os"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
	"gopkg.in/yaml.v3"
)

type Teams struct {
	Teams []*Team
}

type Team struct {
	Name                  string
	Description           string
	SlackChannel          string `yaml:"slack-channel"`
	PlatformAlertsChannel string `yaml:"platform-alerts-channel"`
	GCPProjectIDs         map[string]string
	AzureGroupID          uuid.UUID
}

func (t *Team) Convert() *db.Team {
	var slackAlertsChannel string

	if len(t.PlatformAlertsChannel) > 0 {
		slackAlertsChannel = t.PlatformAlertsChannel
	} else if len(t.SlackChannel) > 0 {
		slackAlertsChannel = t.SlackChannel
	}

	return &db.Team{
		Team: &sqlc.Team{
			Slug:               slug.Slug(t.Name),
			Purpose:            t.Description,
			SlackAlertsChannel: slackAlertsChannel,
		},
	}
}

func ReadTeamFiles(ymlPath, jsonPath string, log logger.Logger) (map[string]*Team, error) {
	yf, err := os.Open(ymlPath)
	if err != nil {
		return nil, err
	}
	defer yf.Close()

	jf, err := os.Open(jsonPath)
	if err != nil {
		return nil, err
	}
	defer jf.Close()

	jdec := json.NewDecoder(jf)
	ydec := yaml.NewDecoder(yf)

	teams := &Teams{}
	err = ydec.Decode(&teams)
	if err != nil {
		return nil, err
	}

	teammap := make(map[string]*Team)
	for _, team := range teams.Teams {
		teammap[team.Name] = team
	}

	idmap := make(map[string]string)
	err = jdec.Decode(&idmap)
	if err != nil {
		return nil, err
	}

	for azureID, name := range idmap {
		_, exists := teammap[name]
		if !exists {
			log.Errorf("no team for azure mapping (%q, %q)", name, azureID)
			continue
		}
		teammap[name].AzureGroupID = uuid.MustParse(azureID)
	}

	return teammap, nil
}
