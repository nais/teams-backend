package legacy

import (
	"encoding/json"
	"os"
)

// read the prod-output.json or dev-output.json files from nais/teams/gcp-projects/
func ReadGcpTeamCacheFile(filename string) (map[string]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data := struct {
		TeamProjectIDMapping struct {
			Value map[string]string `json:"value"`
		} `json:"team_projectid_mapping"`
	}{}
	json.NewDecoder(f).Decode(&data)
	return data.TeamProjectIDMapping.Value, nil
}

// merge the outputs of team cache files, dev and prod
func MergeGcpTeamCacheFiles(dev, prod map[string]string) map[string]map[string]string {
	output := map[string]map[string]string{}
	for slug, projectID := range dev {
		output[slug] = make(map[string]string)
		output[slug]["dev"] = projectID
	}
	for slug, projectID := range prod {
		if output[slug] == nil {
			output[slug] = make(map[string]string)
		}
		output[slug]["prod"] = projectID
	}
	return output
}

/*
teamname {
	dev: foo
	prod: bar
}

*/
