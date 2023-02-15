package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/graph/apierror"
	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/graph/model"
	"github.com/nais/console/pkg/sqlc"
)

// EnableReconciler is the resolver for the enableReconciler field.
func (r *mutationResolver) EnableReconciler(ctx context.Context, name sqlc.ReconcilerName) (*db.Reconciler, error) {
	if !name.Valid() {
		return nil, apierror.Errorf("%q is not a valid name", name)
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("create log correlation ID: %w", err)
	}

	_, err = r.database.GetReconciler(ctx, name)
	if err != nil {
		r.log.WithError(err).Errorf("unable to get reconciler: %q", name)
		return nil, apierror.Errorf("Unable to get reconciler: %q", name)
	}

	configs, err := r.database.GetReconcilerConfig(ctx, name)
	if err != nil {
		r.log.WithError(err).Errorf("unable to get reconciler config")
		return nil, apierror.Errorf("Unable to get reconciler config")
	}

	missingOptions := make([]string, 0)
	for _, config := range configs {
		if !config.Configured {
			missingOptions = append(missingOptions, string(config.Key))
		}
	}

	if len(missingOptions) != 0 {
		r.log.WithError(err).Errorf("reconciler is not fully configured")
		return nil, apierror.Errorf("Reconciler is not fully configured, missing one or more options: %s", strings.Join(missingOptions, ", "))
	}

	reconciler, err := r.database.EnableReconciler(ctx, name)
	if err != nil {
		r.log.WithError(err).Errorf("unable to enable reconciler")
		return nil, apierror.Errorf("Unable to enable reconciler")
	}

	err = r.teamSyncHandler.UseReconciler(*reconciler)
	if err != nil {
		if _, err := r.database.DisableReconciler(ctx, name); err != nil {
			r.log.WithError(err).Errorf("reconciler was enabled, but initialization failed, and we were unable to disable the reconciler.")
			return nil, apierror.Errorf("Reconciler was enabled, but initialization failed, and we were unable to disable the reconciler.")
		}

		r.log.WithError(err).Errorf("reconciler will not be enabled because of an initialization failure. Please verify that you have entered correct configuratiom values.")
		return nil, apierror.Errorf("Reconciler will not be enabled because of an initialization failure. Please verify that you have entered correct configuratiom values.")
	}

	actor := authz.ActorFromContext(ctx)
	targets := []auditlogger.Target{
		auditlogger.ReconcilerTarget(name),
	}
	fields := auditlogger.Fields{
		Action:        sqlc.AuditActionGraphqlApiReconcilersEnable,
		Actor:         actor,
		CorrelationID: correlationID,
	}
	r.auditLogger.Logf(ctx, r.database, targets, fields, "Enable reconciler: %q", name)

	_, err = r.teamSyncHandler.ScheduleAllTeams(ctx, correlationID)
	if err != nil {
		r.log.WithError(err).Errorf("reconcile all teams")
	}

	return reconciler, nil
}

// DisableReconciler is the resolver for the disableReconciler field.
func (r *mutationResolver) DisableReconciler(ctx context.Context, name sqlc.ReconcilerName) (*db.Reconciler, error) {
	if !name.Valid() {
		return nil, apierror.Errorf("%q is not a valid name", name)
	}

	var reconciler *db.Reconciler
	var err error
	err = r.database.Transaction(ctx, func(ctx context.Context, dbtx db.Database) error {
		reconciler, err = dbtx.GetReconciler(ctx, name)
		if err != nil {
			return err
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
	targets := []auditlogger.Target{
		auditlogger.ReconcilerTarget(name),
	}
	fields := auditlogger.Fields{
		Action: sqlc.AuditActionGraphqlApiReconcilersDisable,
		Actor:  actor,
	}
	r.auditLogger.Logf(ctx, r.database, targets, fields, "Disable reconciler: %q", name)
	r.teamSyncHandler.RemoveReconciler(name)

	return reconciler, nil
}

// ConfigureReconciler is the resolver for the configureReconciler field.
func (r *mutationResolver) ConfigureReconciler(ctx context.Context, name sqlc.ReconcilerName, config []*model.ReconcilerConfigInput) (*db.Reconciler, error) {
	if !name.Valid() {
		return nil, apierror.Errorf("%q is not a valid name", name)
	}

	reconcilerConfig := make(map[sqlc.ReconcilerConfigKey]string)
	for _, entry := range config {
		reconcilerConfig[entry.Key] = entry.Value
	}

	reconciler, err := r.database.ConfigureReconciler(ctx, name, reconcilerConfig)
	if err != nil {
		return nil, err
	}

	if reconciler.Enabled {
		err = r.teamSyncHandler.UseReconciler(*reconciler)
		if err != nil {
			r.log.WithError(err).Errorf("use reconciler: %q", reconciler.Name)
		}
	}

	actor := authz.ActorFromContext(ctx)
	targets := []auditlogger.Target{
		auditlogger.ReconcilerTarget(name),
	}
	fields := auditlogger.Fields{
		Action: sqlc.AuditActionGraphqlApiReconcilersConfigure,
		Actor:  actor,
	}
	r.auditLogger.Logf(ctx, r.database, targets, fields, "Configure reconciler: %q", name)

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
	targets := []auditlogger.Target{
		auditlogger.ReconcilerTarget(name),
	}
	fields := auditlogger.Fields{
		Action: sqlc.AuditActionGraphqlApiReconcilersReset,
		Actor:  actor,
	}
	r.auditLogger.Logf(ctx, r.database, targets, fields, "Reset reconciler: %q", name)

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

// AuditLogs is the resolver for the auditLogs field.
func (r *reconcilerResolver) AuditLogs(ctx context.Context, obj *db.Reconciler) ([]*db.AuditLog, error) {
	return r.database.GetAuditLogsForReconciler(ctx, obj.Name)
}

// Value is the resolver for the value field.
func (r *reconcilerConfigResolver) Value(ctx context.Context, obj *db.ReconcilerConfig) (*string, error) {
	switch value := obj.Value.(type) {
	case string:
		return &value, nil
	}

	return nil, nil
}

// Reconciler returns generated.ReconcilerResolver implementation.
func (r *Resolver) Reconciler() generated.ReconcilerResolver { return &reconcilerResolver{r} }

// ReconcilerConfig returns generated.ReconcilerConfigResolver implementation.
func (r *Resolver) ReconcilerConfig() generated.ReconcilerConfigResolver {
	return &reconcilerConfigResolver{r}
}

type (
	reconcilerResolver       struct{ *Resolver }
	reconcilerConfigResolver struct{ *Resolver }
)
