package db

import (
	"context"

	"github.com/nais/teams-backend/pkg/slug"

	"github.com/google/uuid"
	"github.com/nais/teams-backend/pkg/sqlc"
)

func (d *database) SetReconcilerErrorForTeam(ctx context.Context, correlationID uuid.UUID, slug slug.Slug, reconcilerName sqlc.ReconcilerName, err error) error {
	return d.querier.SetReconcilerErrorForTeam(ctx, sqlc.SetReconcilerErrorForTeamParams{
		CorrelationID: correlationID,
		TeamSlug:      slug,
		Reconciler:    reconcilerName,
		ErrorMessage:  err.Error(),
	})
}

func (d *database) GetTeamReconcilerErrors(ctx context.Context, slug slug.Slug) ([]*ReconcilerError, error) {
	rows, err := d.querier.GetTeamReconcilerErrors(ctx, slug)
	if err != nil {
		return nil, err
	}

	errors := make([]*ReconcilerError, 0)
	for _, row := range rows {
		errors = append(errors, &ReconcilerError{ReconcilerError: row})
	}

	return errors, nil
}

func (d *database) ClearReconcilerErrorsForTeam(ctx context.Context, slug slug.Slug, reconcilerName sqlc.ReconcilerName) error {
	return d.querier.ClearReconcilerErrorsForTeam(ctx, sqlc.ClearReconcilerErrorsForTeamParams{
		TeamSlug:   slug,
		Reconciler: reconcilerName,
	})
}
