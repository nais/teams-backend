package reconcilers

import (
	"context"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/db"
)

// Input Input for reconcilers
type Input struct {
	CorrelationID uuid.UUID
	Team          db.Team
}

// CreateReconcilerInput Helper function to create input for reconcilers, with members already set on the team object
func CreateReconcilerInput(ctx context.Context, database db.Database, team db.Team) (Input, error) {
	correlationID, err := uuid.NewUUID()
	if err != nil {
		return Input{}, err
	}

	members, err := database.GetTeamMembers(ctx, team.ID)
	if err != nil {
		return Input{}, err
	}

	team.Members = members

	return Input{
		CorrelationID: correlationID,
		Team:          team,
	}, nil
}

func (i Input) WithCorrelationID(correlationID uuid.UUID) Input {
	i.CorrelationID = correlationID

	return i
}
