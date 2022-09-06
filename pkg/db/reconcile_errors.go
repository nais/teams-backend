package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/sqlc"
)

type ReconcileError struct {
	*sqlc.ReconcileError
}

func (d *database) AddTeamReconcileError(ctx context.Context, correlationID uuid.UUID, teamID uuid.UUID, systemName sqlc.SystemName, err error) error {
	return d.querier.AddTeamReconcileError(ctx, sqlc.AddTeamReconcileErrorParams{
		CorrelationID: correlationID,
		TeamID:        teamID,
		SystemName:    systemName,
		ErrorMessage:  err.Error(),
	})
}

func (d *database) GetTeamReconcileErrors(ctx context.Context, teamID uuid.UUID) ([]*ReconcileError, error) {
	rows, err := d.querier.GetTeamReconcileErrors(ctx, teamID)
	if err != nil {
		return nil, err
	}

	errors := make([]*ReconcileError, 0)
	for _, run := range rows {
		errors = append(errors, &ReconcileError{ReconcileError: run})
	}

	return errors, nil
}

func (d *database) PurgeTeamReconcileErrors(ctx context.Context, teamID uuid.UUID, systemName sqlc.SystemName) error {
	return d.querier.PurgeTeamReconcileErrors(ctx, sqlc.PurgeTeamReconcileErrorsParams{
		TeamID:     teamID,
		SystemName: systemName,
	})
}
