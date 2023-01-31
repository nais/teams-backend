package teamsync

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/nais/console/pkg/reconcilers"
	azure_group_reconciler "github.com/nais/console/pkg/reconcilers/azure/group"
	github_team_reconciler "github.com/nais/console/pkg/reconcilers/github/team"
	google_gcp_reconciler "github.com/nais/console/pkg/reconcilers/google/gcp"
	google_workspace_admin_reconciler "github.com/nais/console/pkg/reconcilers/google/workspace_admin"
	nais_deploy_reconciler "github.com/nais/console/pkg/reconcilers/nais/deploy"
	nais_namespace_reconciler "github.com/nais/console/pkg/reconcilers/nais/namespace"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/metrics"
	"github.com/nais/console/pkg/sqlc"
)

type Handler struct {
	activeReconcilers map[sqlc.ReconcilerName]ReconcilerWithRunOrder
	database          db.Database
	auditLogger       auditlogger.AuditLogger
	log               logger.Logger
	lock              sync.Mutex
	cfg               *config.Config
}

type ReconcilerWithRunOrder struct {
	runOrder   int32
	reconciler reconcilers.Reconciler
}

var factories = map[sqlc.ReconcilerName]reconcilers.ReconcilerFactory{
	azure_group_reconciler.Name:            azure_group_reconciler.NewFromConfig,
	github_team_reconciler.Name:            github_team_reconciler.NewFromConfig,
	google_workspace_admin_reconciler.Name: google_workspace_admin_reconciler.NewFromConfig,
	google_gcp_reconciler.Name:             google_gcp_reconciler.NewFromConfig,
	nais_namespace_reconciler.Name:         nais_namespace_reconciler.NewFromConfig,
	nais_deploy_reconciler.Name:            nais_deploy_reconciler.NewFromConfig,
}

func NewHandler(database db.Database, cfg *config.Config, auditLogger auditlogger.AuditLogger, log logger.Logger) *Handler {
	return &Handler{
		activeReconcilers: make(map[sqlc.ReconcilerName]ReconcilerWithRunOrder),
		database:          database,
		cfg:               cfg,
		auditLogger:       auditLogger,
		log:               log,
	}
}

// InitReconcilers initializes the currently enabled reconcilers during startup of Console
func (r *Handler) InitReconcilers(ctx context.Context) error {
	enabledReconcilers, err := r.database.GetEnabledReconcilers(ctx)
	if err != nil {
		return err
	}

	for _, reconciler := range enabledReconcilers {
		if err = r.UseReconciler(ctx, *reconciler); err != nil {
			r.log.WithReconciler(string(reconciler.Name)).WithError(err).Error("use reconciler")
		}
	}

	return nil
}

// UseReconciler will include a reconciler in the list of currently active reconcilers. During the activation this
// function will acquire a lock, preventing other processes from reading from the list of active reconcilers.
func (r *Handler) UseReconciler(ctx context.Context, reconciler db.Reconciler) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	factory, err := getReconcilerFactory(reconciler.Name)
	if err != nil {
		return err
	}

	reconcilerImplementation, err := factory(ctx, r.database, r.cfg, r.auditLogger, r.log)
	if err != nil {
		return err
	}

	r.activeReconcilers[reconciler.Name] = ReconcilerWithRunOrder{
		runOrder:   reconciler.RunOrder,
		reconciler: reconcilerImplementation,
	}
	return nil
}

// RemoveReconciler will remove a reconciler from the list of currently active reconcilers. During the removal of the
// reconciler this function will acquire a lock, preventing other processes from reading from the list.
func (r *Handler) RemoveReconciler(reconcilerName sqlc.ReconcilerName) {
	r.lock.Lock()
	defer r.lock.Unlock()
	delete(r.activeReconcilers, reconcilerName)
}

func (r *Handler) ReconcileTeam(ctx context.Context, input reconcilers.Input) error {
	if !input.Team.Enabled {
		r.log.Infof("team is not enabled, skipping reconciliation")
		return nil
	}

	if len(r.activeReconcilers) == 0 {
		r.log.Warnf("no reconcilers are currently enabled")
		return nil
	}

	r.log.Infof("reconcile team")
	errors := 0
	teamReconcilerTimer := metrics.MeasureReconcileTeamDuration(input.Team.Slug)

	r.lock.Lock()
	var orderedReconcilers []ReconcilerWithRunOrder
	for _, reconcilerWithOrder := range r.activeReconcilers {
		orderedReconcilers = append(orderedReconcilers, reconcilerWithOrder)
	}
	r.lock.Unlock()

	sort.Slice(orderedReconcilers, func(i, j int) bool {
		return orderedReconcilers[i].runOrder < orderedReconcilers[j].runOrder
	})

	for _, reconcilerWithRunOrder := range orderedReconcilers {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		reconcilerImpl := reconcilerWithRunOrder.reconciler
		name := reconcilerImpl.Name()
		log := r.log.WithSystem(string(name))
		metrics.IncReconcilerCounter(name, metrics.ReconcilerStateStarted)

		reconcileTimer := metrics.MeasureReconcilerDuration(name)
		err := reconcilerImpl.Reconcile(ctx, input)
		if err != nil {
			metrics.IncReconcilerCounter(name, metrics.ReconcilerStateFailed)
			log.WithError(err).Error("reconcile")
			errors++
			err = r.database.SetReconcilerErrorForTeam(ctx, input.CorrelationID, input.Team.Slug, name, err)
			if err != nil {
				log.WithError(err).Error("add reconcile error to database")
			}
			continue
		}
		reconcileTimer.ObserveDuration()

		err = r.database.ClearReconcilerErrorsForTeam(ctx, input.Team.Slug, name)
		if err != nil {
			log.WithError(err).Error("purge reconcile errors")
		}

		metrics.IncReconcilerCounter(name, metrics.ReconcilerStateSuccessful)
	}

	if errors > 0 {
		return fmt.Errorf("%d error(s) occurred during reconcile", errors)
	}

	teamReconcilerTimer.ObserveDuration()

	if err := r.database.SetLastSuccessfulSyncForTeam(ctx, input.Team.Slug); err != nil {
		r.log.WithError(err).Error("update last successful sync timestamp")
	}
	return nil
}

func getReconcilerFactory(reconcilerName sqlc.ReconcilerName) (reconcilers.ReconcilerFactory, error) {
	factory, exists := factories[reconcilerName]
	if !exists {
		return nil, fmt.Errorf("reconciler missing factory entry: %q", reconcilerName)
	}
	return factory, nil
}
