package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/console/pkg/sqlc"
)

func (d *database) GetReconciler(ctx context.Context, reconcilerName sqlc.ReconcilerName) (*Reconciler, error) {
	reconciler, err := d.querier.GetReconciler(ctx, reconcilerName)
	if err != nil {
		return nil, err
	}

	return &Reconciler{Reconciler: reconciler}, nil
}

func (d *database) GetReconcilers(ctx context.Context) ([]*Reconciler, error) {
	rows, err := d.querier.GetReconcilers(ctx)
	if err != nil {
		return nil, err
	}

	return wrapReconcilers(rows), nil
}

func (d *database) GetEnabledReconcilers(ctx context.Context) ([]*Reconciler, error) {
	rows, err := d.querier.GetEnabledReconcilers(ctx)
	if err != nil {
		return nil, err
	}

	return wrapReconcilers(rows), nil
}

func (d *database) ConfigureReconciler(ctx context.Context, reconcilerName sqlc.ReconcilerName, config map[string]string) (*Reconciler, error) {
	reconciler, err := d.querier.GetReconciler(ctx, reconcilerName)
	if err != nil {
		return nil, err
	}

	err = d.querier.Transaction(ctx, func(ctx context.Context, querier Querier) error {
		rows, err := querier.GetReconcilerConfig(ctx, reconcilerName)
		if err != nil {
			return err
		}

		if len(rows) == 0 {
			return fmt.Errorf("reconciler %q does not have any configuration options", reconcilerName)
		}

		validOptions := make(map[string]struct{})
		for _, row := range rows {
			validOptions[row.Key] = struct{}{}
		}

		for key, value := range config {
			if _, exists := validOptions[key]; !exists {
				return fmt.Errorf("unknown configuration option %q for reconciler %q. Valid options: %s", key, reconcilerName, strings.Join(getKeys(validOptions), ", "))
			}

			err := querier.ConfigureReconciler(ctx, sqlc.ConfigureReconcilerParams{
				Reconciler: reconcilerName,
				Key:        key,
				Value:      value,
			})
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &Reconciler{Reconciler: reconciler}, nil
}

func (d *database) GetReconcilerConfig(ctx context.Context, reconcilerName sqlc.ReconcilerName) ([]*ReconcilerConfig, error) {
	rows, err := d.querier.GetReconcilerConfig(ctx, reconcilerName)
	if err != nil {
		return nil, err
	}

	config := make([]*ReconcilerConfig, 0, len(rows))
	for _, row := range rows {
		config = append(config, &ReconcilerConfig{GetReconcilerConfigRow: row})
	}

	return config, nil
}

func (d *database) ResetReconcilerConfig(ctx context.Context, reconcilerName sqlc.ReconcilerName) (*Reconciler, error) {
	reconciler, err := d.querier.GetReconciler(ctx, reconcilerName)
	if err != nil {
		return nil, err
	}

	err = d.querier.ResetReconcilerConfig(ctx, reconcilerName)
	if err != nil {
		return nil, err
	}

	return &Reconciler{Reconciler: reconciler}, nil
}

func (d *database) EnableReconciler(ctx context.Context, reconcilerName sqlc.ReconcilerName) (*Reconciler, error) {
	reconciler, err := d.querier.EnableReconciler(ctx, reconcilerName)
	if err != nil {
		return nil, err
	}

	return &Reconciler{Reconciler: reconciler}, nil
}

func (d *database) DisableReconciler(ctx context.Context, reconcilerName sqlc.ReconcilerName) (*Reconciler, error) {
	reconciler, err := d.querier.DisableReconciler(ctx, reconcilerName)
	if err != nil {
		return nil, err
	}

	return &Reconciler{Reconciler: reconciler}, nil
}

func wrapReconcilers(rows []*sqlc.Reconciler) []*Reconciler {
	reconcilers := make([]*Reconciler, 0, len(rows))
	for _, row := range rows {
		reconcilers = append(reconcilers, &Reconciler{Reconciler: row})
	}
	return reconcilers
}

func getKeys(values map[string]struct{}) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}
