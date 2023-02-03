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
	factories         ReconcilerFactories
	mainContext       context.Context
}

type ReconcilerWithRunOrder struct {
	runOrder   int32
	reconciler reconcilers.Reconciler
}

type ReconcilerFactories map[sqlc.ReconcilerName]reconcilers.ReconcilerFactory

var factories = ReconcilerFactories{
	azure_group_reconciler.Name:            azure_group_reconciler.NewFromConfig,
	github_team_reconciler.Name:            github_team_reconciler.NewFromConfig,
	google_workspace_admin_reconciler.Name: google_workspace_admin_reconciler.NewFromConfig,
	google_gcp_reconciler.Name:             google_gcp_reconciler.NewFromConfig,
	nais_namespace_reconciler.Name:         nais_namespace_reconciler.NewFromConfig,
	nais_deploy_reconciler.Name:            nais_deploy_reconciler.NewFromConfig,
}

func NewHandler(ctx context.Context, database db.Database, cfg *config.Config, auditLogger auditlogger.AuditLogger, log logger.Logger) *Handler {
	return &Handler{
		activeReconcilers: make(map[sqlc.ReconcilerName]ReconcilerWithRunOrder),
		database:          database,
		cfg:               cfg,
		auditLogger:       auditLogger,
		log:               log,
		factories:         factories,
		mainContext:       ctx,
	}
}

func (h *Handler) SetReconcilerFactories(factories ReconcilerFactories) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.factories = factories
}

// InitReconcilers initializes the currently enabled reconcilers during startup of Console
func (h *Handler) InitReconcilers(ctx context.Context) error {
	enabledReconcilers, err := h.database.GetEnabledReconcilers(ctx)
	if err != nil {
		return err
	}

	for _, reconciler := range enabledReconcilers {
		if err = h.UseReconciler(*reconciler); err != nil {
			h.log.WithReconciler(string(reconciler.Name)).WithError(err).Error("use reconciler")
		}
	}

	return nil
}

// UseReconciler will include a reconciler in the list of currently active reconcilers. During the activation this
// function will acquire a lock, preventing other processes from reading from the list of active reconcilers.
func (h *Handler) UseReconciler(reconciler db.Reconciler) error {
	h.lock.Lock()
	defer h.lock.Unlock()

	factory, err := h.getReconcilerFactory(reconciler.Name)
	if err != nil {
		return err
	}

	reconcilerImplementation, err := factory(h.mainContext, h.database, h.cfg, h.auditLogger, h.log)
	if err != nil {
		return err
	}

	h.activeReconcilers[reconciler.Name] = ReconcilerWithRunOrder{
		runOrder:   reconciler.RunOrder,
		reconciler: reconcilerImplementation,
	}
	return nil
}

// RemoveReconciler will remove a reconciler from the list of currently active reconcilers. During the removal of the
// reconciler this function will acquire a lock, preventing other processes from reading from the list.
func (h *Handler) RemoveReconciler(reconcilerName sqlc.ReconcilerName) {
	h.lock.Lock()
	defer h.lock.Unlock()
	delete(h.activeReconcilers, reconcilerName)
}

func (h *Handler) ReconcileTeam(ctx context.Context, input reconcilers.Input) error {
	if !input.Team.Enabled {
		h.log.Infof("team is not enabled, skipping reconciliation")
		return nil
	}

	if len(h.activeReconcilers) == 0 {
		h.log.Warnf("no reconcilers are currently enabled")
		return nil
	}

	log := h.log.WithTeamSlug(string(input.Team.Slug))

	log.Infof("reconcile team")
	errors := 0
	teamReconcilerTimer := metrics.MeasureReconcileTeamDuration()

	h.lock.Lock()
	orderedReconcilers := getOrderedReconcilers(h.activeReconcilers)
	h.lock.Unlock()

	for _, reconcilerWithRunOrder := range orderedReconcilers {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		reconcilerImpl := reconcilerWithRunOrder.reconciler
		name := reconcilerImpl.Name()
		log := log.WithSystem(string(name))
		metrics.IncReconcilerCounter(name, metrics.ReconcilerStateStarted)

		reconcileTimer := metrics.MeasureReconcilerDuration(name)
		err := reconcilerImpl.Reconcile(ctx, input)
		if err != nil {
			metrics.IncReconcilerCounter(name, metrics.ReconcilerStateFailed)
			log.WithError(err).Error("reconcile")
			errors++
			err = h.database.SetReconcilerErrorForTeam(ctx, input.CorrelationID, input.Team.Slug, name, err)
			if err != nil {
				log.WithError(err).Error("add reconcile error to database")
			}
			continue
		}
		duration := reconcileTimer.ObserveDuration()
		h.log.Debugf("successful reconcile duration for team %q with reconciler %q: %s", input.Team.Slug, name, duration)

		err = h.database.ClearReconcilerErrorsForTeam(ctx, input.Team.Slug, name)
		if err != nil {
			log.WithError(err).Error("purge reconcile errors")
		}

		metrics.IncReconcilerCounter(name, metrics.ReconcilerStateSuccessful)
	}

	if errors > 0 {
		return fmt.Errorf("%d error(s) occurred during reconcile", errors)
	}

	duration := teamReconcilerTimer.ObserveDuration()
	h.log.Debugf("successful reconcile duration for team %q: %s", input.Team.Slug, duration)

	if err := h.database.SetLastSuccessfulSyncForTeam(ctx, input.Team.Slug); err != nil {
		h.log.WithError(err).Error("update last successful sync timestamp")
	}
	return nil
}

func (h *Handler) getReconcilerFactory(reconcilerName sqlc.ReconcilerName) (reconcilers.ReconcilerFactory, error) {
	factory, exists := h.factories[reconcilerName]
	if !exists {
		return nil, fmt.Errorf("missing reconciler factory entry: %q", reconcilerName)
	}
	return factory, nil
}

func getOrderedReconcilers(reconcilers map[sqlc.ReconcilerName]ReconcilerWithRunOrder) []ReconcilerWithRunOrder {
	var orderedReconcilers []ReconcilerWithRunOrder
	for _, reconcilerWithOrder := range reconcilers {
		orderedReconcilers = append(orderedReconcilers, reconcilerWithOrder)
	}
	sort.Slice(orderedReconcilers, func(i, j int) bool {
		return orderedReconcilers[i].runOrder < orderedReconcilers[j].runOrder
	})
	return orderedReconcilers
}
