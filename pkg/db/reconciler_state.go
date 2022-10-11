package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/nais/console/pkg/sqlc"
)

// LoadReconcilerStateForTeam Load the team state for a given reconciler into the state parameter
func (d *database) LoadReconcilerStateForTeam(ctx context.Context, reconcilerName sqlc.ReconcilerName, teamID uuid.UUID, state interface{}) error {
	systemState, err := d.querier.GetReconcilerStateForTeam(ctx, sqlc.GetReconcilerStateForTeamParams{
		Reconciler: reconcilerName,
		TeamID:     teamID,
	})
	if err != nil {
		// assume empty state
		systemState = &sqlc.ReconcilerState{State: pgtype.JSONB{}}
	}

	err = systemState.State.AssignTo(state)
	if err != nil {
		return fmt.Errorf("unable to assign state: %w", err)
	}

	return nil
}

// SetReconcilerStateForTeam Update the team state for a given reconciler
func (d *database) SetReconcilerStateForTeam(ctx context.Context, reconcilerName sqlc.ReconcilerName, teamID uuid.UUID, state interface{}) error {
	newState := pgtype.JSONB{}
	err := newState.Set(state)
	if err != nil {
		return fmt.Errorf("unable to set new system state: %w", err)
	}

	return d.querier.SetReconcilerStateForTeam(ctx, sqlc.SetReconcilerStateForTeamParams{
		Reconciler: reconcilerName,
		TeamID:     teamID,
		State:      newState,
	})
}
