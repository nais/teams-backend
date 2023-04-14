package dependencytrack

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_All(t *testing.T) {
	c := NewClient("http://localhost:9001/api/v1", "admin", "yolo")

	team := "yolo"
	teams, err := c.GetTeams(context.TODO())
	assert.NoError(t, err)

	uuid := GetTeamUuid(teams, team)
	if uuid == "" {
		team, err := c.CreateTeam(context.TODO(), team, []Permission{
			ViewPortfolioPermission,
		})
		assert.NoError(t, err)
		uuid = team.Uuid
	}

	err = c.CreateUser(context.TODO(), "user@dev-nais.io")
	assert.NoError(t, err)

	err = c.AddToTeam(context.TODO(), "user@dev-nais.io", uuid)
	assert.NoError(t, err)
	fmt.Printf("team uuid: %s\n", uuid)
}

func TestClient_GetTeams(t *testing.T) {
	c := NewClient("http://localhost:9001/api/v1", "admin", "yolo")
	teams, err := c.GetTeams(context.TODO())
	assert.NoError(t, err)
	fmt.Printf("%+v\n", teams)
}
