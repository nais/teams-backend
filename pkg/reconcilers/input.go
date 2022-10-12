package reconcilers

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/db"
)

// Input Input for reconcilers
type Input struct {
	CorrelationID uuid.UUID
	Team          db.Team
	TeamMembers   []*db.User
}

// CreateReconcilerInput Helper function to create input for reconcilers, with members already set on the team object
func CreateReconcilerInput(ctx context.Context, database db.Database, team db.Team) (Input, error) {
	if team.Disabled {
		return Input{}, fmt.Errorf("team %q is disabled", team.Slug)
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return Input{}, err
	}

	members, err := database.GetTeamMembers(ctx, team.ID)
	if err != nil {
		return Input{}, err
	}

	return Input{
		CorrelationID: correlationID,
		Team:          team,
		TeamMembers:   members,
	}, nil
}

func (i Input) WithCorrelationID(correlationID uuid.UUID) Input {
	i.CorrelationID = correlationID

	return i
}
