package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/nais/console/pkg/sqlc"
)

// LoadSystemState Load the team state for a given system into the state parameter
func (d *database) LoadSystemState(ctx context.Context, systemName sqlc.SystemName, teamID uuid.UUID, state interface{}) error {
	systemState, err := d.querier.GetTeamSystemState(ctx, sqlc.GetTeamSystemStateParams{
		SystemName: systemName,
		TeamID:     teamID,
	})
	if err != nil {
		// assume empty state
		systemState = &sqlc.SystemState{State: pgtype.JSONB{}}
	}

	err = systemState.State.AssignTo(state)
	if err != nil {
		return fmt.Errorf("unable to assign state: %w", err)
	}

	return nil
}

// SetSystemState Update the team state for a given system
func (d *database) SetSystemState(ctx context.Context, systemName sqlc.SystemName, teamID uuid.UUID, state interface{}) error {
	newState := pgtype.JSONB{}
	err := newState.Set(state)
	if err != nil {
		return fmt.Errorf("unable to set new system state: %w", err)
	}

	return d.querier.SetTeamSystemState(ctx, sqlc.SetTeamSystemStateParams{
		SystemName: systemName,
		TeamID:     teamID,
		State:      newState,
	})
}
