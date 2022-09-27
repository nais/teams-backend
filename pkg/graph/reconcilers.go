package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/console"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/graph/model"
	"github.com/nais/console/pkg/sqlc"
)

// EnableReconciler is the resolver for the enableReconciler field.
func (r *mutationResolver) EnableReconciler(ctx context.Context, name sqlc.ReconcilerName) (*db.Reconciler, error) {
	if !name.Valid() {
		return nil, fmt.Errorf("%q is not a valid name", name)
	}

	var reconciler *db.Reconciler
	var err error
	err = r.database.Transaction(ctx, func(ctx context.Context, dbtx db.Database) error {
		reconciler, err = dbtx.GetReconciler(ctx, name)
		if err != nil {
			return err
		}

		if reconciler.Enabled {
			return fmt.Errorf("reconciler is already enabled")
		}

		configs, err := dbtx.GetReconcilerConfig(ctx, name)
		if err != nil {
			return err
		}

		missingOptions := make([]string, 0)
		for _, config := range configs {
			if !config.Configured {
				missingOptions = append(missingOptions, string(config.Key))
			}
		}

		if len(missingOptions) != 0 {
			return fmt.Errorf("reconciler is not fully configured, missing one or more options: %s", strings.Join(missingOptions, ", "))
		}

		reconciler, err = dbtx.EnableReconciler(ctx, name)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	actor := authz.ActorFromContext(ctx)
	fields := auditlogger.Fields{
		Action: sqlc.AuditActionGraphqlApiReconcilersEnable,
		Actor:  console.Strp(actor.User.Identity()),
	}
	r.auditLogger.Logf(ctx, fields, "Enable reconciler: %q", name)

	return reconciler, nil
}

// DisableReconciler is the resolver for the disableReconciler field.
func (r *mutationResolver) DisableReconciler(ctx context.Context, name sqlc.ReconcilerName) (*db.Reconciler, error) {
	if !name.Valid() {
		return nil, fmt.Errorf("%q is not a valid name", name)
	}

	var reconciler *db.Reconciler
	var err error
	err = r.database.Transaction(ctx, func(ctx context.Context, dbtx db.Database) error {
		reconciler, err = dbtx.GetReconciler(ctx, name)
		if err != nil {
			return err
		}

		if !reconciler.Enabled {
			return fmt.Errorf("reconciler is already disabled")
		}

		reconciler, err = dbtx.DisableReconciler(ctx, name)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	actor := authz.ActorFromContext(ctx)
	fields := auditlogger.Fields{
		Action: sqlc.AuditActionGraphqlApiReconcilersDisable,
		Actor:  console.Strp(actor.User.Identity()),
	}
	r.auditLogger.Logf(ctx, fields, "Disable reconciler: %q", name)

	return reconciler, nil
}

// ConfigureReconciler is the resolver for the configureReconciler field.
func (r *mutationResolver) ConfigureReconciler(ctx context.Context, name sqlc.ReconcilerName, config []*model.ReconcilerConfigInput) (*db.Reconciler, error) {
	if !name.Valid() {
		return nil, fmt.Errorf("%q is not a valid name", name)
	}

	reconcilerConfig := make(map[sqlc.ReconcilerConfigKey]string)
	for _, entry := range config {
		reconcilerConfig[entry.Key] = entry.Value
	}

	reconciler, err := r.database.ConfigureReconciler(ctx, name, reconcilerConfig)
	if err != nil {
		return nil, err
	}

	actor := authz.ActorFromContext(ctx)
	fields := auditlogger.Fields{
		Action: sqlc.AuditActionGraphqlApiReconcilersConfigure,
		Actor:  console.Strp(actor.User.Identity()),
	}
	r.auditLogger.Logf(ctx, fields, "Configure reconciler: %q", name)

	return reconciler, nil
}

// ResetReconciler is the resolver for the resetReconciler field.
func (r *mutationResolver) ResetReconciler(ctx context.Context, name sqlc.ReconcilerName) (*db.Reconciler, error) {
	if !name.Valid() {
		return nil, fmt.Errorf("%q is not a valid name", name)
	}

	var reconciler *db.Reconciler
	var err error
	err = r.database.Transaction(ctx, func(ctx context.Context, dbtx db.Database) error {
		reconciler, err = dbtx.ResetReconcilerConfig(ctx, name)
		if err != nil {
			return err
		}

		if !reconciler.Enabled {
			return nil
		}

		reconciler, err = dbtx.DisableReconciler(ctx, name)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	actor := authz.ActorFromContext(ctx)
	fields := auditlogger.Fields{
		Action: sqlc.AuditActionGraphqlApiReconcilersReset,
		Actor:  console.Strp(actor.User.Identity()),
	}
	r.auditLogger.Logf(ctx, fields, "Reset reconciler: %q", name)

	return reconciler, nil
}

// Reconcilers is the resolver for the reconcilers field.
func (r *queryResolver) Reconcilers(ctx context.Context) ([]*db.Reconciler, error) {
	return r.database.GetReconcilers(ctx)
}

// Config is the resolver for the config field.
func (r *reconcilerResolver) Config(ctx context.Context, obj *db.Reconciler) ([]*db.ReconcilerConfig, error) {
	return r.database.GetReconcilerConfig(ctx, obj.Name)
}

// Configured is the resolver for the configured field.
func (r *reconcilerResolver) Configured(ctx context.Context, obj *db.Reconciler) (bool, error) {
	configs, err := r.database.GetReconcilerConfig(ctx, obj.Name)
	if err != nil {
		return false, err
	}

	for _, config := range configs {
		if !config.Configured {
			return false, nil
		}
	}

	return true, nil
}

// Reconciler returns generated.ReconcilerResolver implementation.
func (r *Resolver) Reconciler() generated.ReconcilerResolver { return &reconcilerResolver{r} }

type reconcilerResolver struct{ *Resolver }
