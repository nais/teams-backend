package legacy

import (
	"encoding/json"
	"os"

	"github.com/nais/console/pkg/dbmodels"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Teams struct {
	Teams []*Team
}

type Team struct {
	AzureID               string
	Name                  string
	Description           string
	SlackChannel          string `yaml:"slack-channel"`
	PlatformAlertsChannel string `yaml:"platform-alerts-channel"`
}

func (t *Team) Convert() *dbmodels.Team {
	meta := make([]*dbmodels.TeamMetadata, 0)

	if len(t.SlackChannel) > 0 {
		meta = append(meta, &dbmodels.TeamMetadata{
			Key:   "slack-channel-generic",
			Value: &t.SlackChannel,
		})
	}

	if len(t.PlatformAlertsChannel) > 0 {
		meta = append(meta, &dbmodels.TeamMetadata{
			Key:   "slack-channel-platform-alerts",
			Value: &t.PlatformAlertsChannel,
		})
	}

	slug := dbmodels.Slug(t.Name)
	return &dbmodels.Team{
		Slug:     slug,
		Name:     t.Name,
		Purpose:  &t.Description,
		Metadata: meta,
	}
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

	for azureid, slug := range idmap {
		_, exists := teammap[slug]
		if !exists {
			log.Errorf("no team for azure mapping (%s, %s)", slug, azureid)
			continue
		}
		teammap[slug].AzureID = azureid
	}

	return teammap, nil
}
