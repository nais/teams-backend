package graph

import (
	"github.com/nais/console/pkg/auth"
)

const (
	ResourceTeams        auth.Resource = "teams"
	ResourceSpecificTeam auth.Resource = "teams:%s"
	ResourceCreateTeam   auth.Resource = "createTeam"
)
