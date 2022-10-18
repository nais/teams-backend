package legacy

import (
	"database/sql"
	"encoding/json"
	"os"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
	log "github.com/sirupsen/logrus"
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

func (t *Team) Convert() (*db.Team, []db.TeamMetadata) {
	metadata := make([]db.TeamMetadata, 0)

	if len(t.SlackChannel) > 0 {
		metadata = append(metadata, db.TeamMetadata{
			Key:   "slack-channel-generic",
			Value: &t.SlackChannel,
		})
	}

	if len(t.PlatformAlertsChannel) > 0 {
		metadata = append(metadata, db.TeamMetadata{
			Key:   "slack-channel-platform-alerts",
			Value: &t.PlatformAlertsChannel,
		})
	}

	desc := sql.NullString{
		String: "Team created by NAIS console",
		Valid:  true,
	}

	if t.Description != "" {
		desc.String = t.Description
	}

	return &db.Team{
		Team: &sqlc.Team{
			Slug:    slug.Slug(t.Name),
			Name:    t.Name,
			Purpose: desc,
		},
	}, metadata
}

func ReadTeamFiles(ymlPath, jsonPath string) (map[string]*Team, error) {
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
