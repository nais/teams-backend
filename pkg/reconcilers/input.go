package reconcilers

import (
	"context"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/sqlc"
)

// Input Input for reconcilers
type Input struct {
	CorrelationID uuid.UUID
	Team          db.Team
	TeamMembers   []*db.User
}

// CreateReconcilerInput Helper function to create input for reconcilers
func CreateReconcilerInput(ctx context.Context, database db.Database, team db.Team, reconcilerName sqlc.ReconcilerName) (Input, error) {
	correlationID, err := uuid.NewUUID()
	if err != nil {
		return Input{}, err
	}

	members, err := database.GetTeamMembersForReconciler(ctx, team.Slug, reconcilerName)
	if err != nil {
		return Input{}, err
	}

	return Input{
		CorrelationID: correlationID,
		Team:          team,
		TeamMembers:   members,
	}, nil
}
