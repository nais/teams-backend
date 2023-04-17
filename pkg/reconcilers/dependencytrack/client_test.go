package dependencytrack

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_All(t *testing.T) {
	c := NewClient("http://localhost:9001/api/v1", "admin", "yolo")

	teamName := "yolo"
	teams, err := c.GetTeams(context.TODO())
	assert.NoError(t, err)

	team := GetTeam(teams, teamName)
	uuid := team.Uuid
	if uuid == "" {
		team, err := c.CreateTeam(context.TODO(), teamName, []Permission{
			ViewPortfolioPermission,
		})
		assert.NoError(t, err)
		uuid = team.Uuid
	}

	err = c.CreateUser(context.TODO(), "user@dev-nais.io")
	assert.NoError(t, err)

	err = c.AddToTeam(context.TODO(), "user@dev-nais.io", uuid)
	assert.NoError(t, err)
	fmt.Printf("teamName uuid: %s\n", uuid)
}

func TestClient_GetTeams(t *testing.T) {
	c := NewClient("http://localhost:9001/api/v1", "admin", "yolo")
	teams, err := c.GetTeams(context.TODO())
	assert.NoError(t, err)
	fmt.Printf("%+v\n", teams)
}

func TestClient_DeleteTeam(t *testing.T) {
	c := NewClient("http://localhost:9001/api/v1", "admin", "yolo")
	err := c.DeleteTeam(context.TODO(), "teamname")
	assert.NoError(t, err)
}
