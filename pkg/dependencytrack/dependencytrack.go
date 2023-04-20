package dependencytrack

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Instances []DependencyTrackInstance

type DependencyTrackInstance struct {
	Endpoint string `json:"endpoint"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (i *Instances) Decode(value string) error {

	*i = make(Instances, 0)
	if value == "" {
		return nil
	}

	err := json.NewDecoder(strings.NewReader(value)).Decode(&i)
	if err != nil {
		return fmt.Errorf("parse dependencytrack instance: %w", err)
	}
	return nil
}
