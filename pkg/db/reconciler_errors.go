package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/sqlc"
)

func (d *database) SetReconcilerErrorForTeam(ctx context.Context, correlationID uuid.UUID, teamID uuid.UUID, reconcilerName sqlc.ReconcilerName, err error) error {
	return d.querier.SetReconcilerErrorForTeam(ctx, sqlc.SetReconcilerErrorForTeamParams{
		CorrelationID: correlationID,
		TeamID:        teamID,
		Reconciler:    reconcilerName,
		ErrorMessage:  err.Error(),
	})
}

func (d *database) GetTeamReconcilerErrors(ctx context.Context, teamID uuid.UUID) ([]*ReconcilerError, error) {
	rows, err := d.querier.GetTeamReconcilerErrors(ctx, teamID)
	if err != nil {
		return nil, err
	}

	errors := make([]*ReconcilerError, 0)
	for _, row := range rows {
		errors = append(errors, &ReconcilerError{ReconcilerError: row})
	}

	return errors, nil
}

func (d *database) ClearReconcilerErrorsForTeam(ctx context.Context, teamID uuid.UUID, reconcilerName sqlc.ReconcilerName) error {
	return d.querier.ClearReconcilerErrorsForTeam(ctx, sqlc.ClearReconcilerErrorsForTeamParams{
		TeamID:     teamID,
		Reconciler: reconcilerName,
	})
}
