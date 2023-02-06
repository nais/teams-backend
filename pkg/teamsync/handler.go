package teamsync

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/metrics"
	"github.com/nais/console/pkg/reconcilers"
	azure_group_reconciler "github.com/nais/console/pkg/reconcilers/azure/group"
	github_team_reconciler "github.com/nais/console/pkg/reconcilers/github/team"
	google_gcp_reconciler "github.com/nais/console/pkg/reconcilers/google/gcp"
	google_workspace_admin_reconciler "github.com/nais/console/pkg/reconcilers/google/workspace_admin"
	nais_deploy_reconciler "github.com/nais/console/pkg/reconcilers/nais/deploy"
	nais_namespace_reconciler "github.com/nais/console/pkg/reconcilers/nais/namespace"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
)

type Handler interface {
	SetReconcilerFactories(factories ReconcilerFactories)
	Schedule(input reconcilers.Input) error
	InitReconcilers(ctx context.Context) error
	UseReconciler(reconciler db.Reconciler) error
	RemoveReconciler(reconcilerName sqlc.ReconcilerName)
	SyncTeams(ctx context.Context)
	UpdateMetrics(ctx context.Context)
	Close()
}

type handler struct {
	activeReconcilers map[sqlc.ReconcilerName]ReconcilerWithRunOrder
	database          db.Database
	syncQueue         Queue
	queueInputs       map[slug.Slug]reconcilers.Input
	teamsInFlight     map[slug.Slug]bool
	syncQueueChan     <-chan slug.Slug
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

const (
	reconcilerTimeout  = time.Minute * 15
	teamSyncMaxRetries = 10
)

func NewHandler(ctx context.Context, database db.Database, cfg *config.Config, auditLogger auditlogger.AuditLogger, log logger.Logger) Handler {
	queue, channel := NewQueue()
	return &handler{
		activeReconcilers: make(map[sqlc.ReconcilerName]ReconcilerWithRunOrder),
		database:          database,
		queueInputs:       make(map[slug.Slug]reconcilers.Input, 0),
		syncQueue:         queue,
		syncQueueChan:     channel,
		teamsInFlight:     make(map[slug.Slug]bool),
		cfg:               cfg,
		auditLogger:       auditLogger,
		log:               log,
		factories:         factories,
		mainContext:       ctx,
	}
}

func (h *handler) Close() {
	h.syncQueue.Close()
}

func (h *handler) SetReconcilerFactories(factories ReconcilerFactories) {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.factories = factories
}

// Schedule a team for sync
func (h *handler) Schedule(input reconcilers.Input) error {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.queueInputs[input.Team.Slug] = input
	h.syncQueue.Add(input.Team.Slug)
	return nil
}

// InitReconcilers initializes the currently enabled reconcilers during startup of Console
func (h *handler) InitReconcilers(ctx context.Context) error {
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
func (h *handler) UseReconciler(reconciler db.Reconciler) error {
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

func (h *handler) SyncTeams(ctx context.Context) {
	for slug := range h.syncQueueChan {
		if h.getTeamInFlight(slug) {
			h.syncQueue.Add(slug)
			continue
		}

		h.setTeamInFlight(slug, true)
		input := h.popInput(slug)

		log := h.log.WithTeamSlug(string(slug))
		if input.NumSyncAttempts > teamSyncMaxRetries {
			metrics.IncReconcilerMaxAttemptsExhaustion()
			log.Errorf("reconcile has failed %d times for team %q, giving up", teamSyncMaxRetries, slug)
			h.setTeamInFlight(slug, false)
			continue
		}

		ctx, cancel := context.WithTimeout(ctx, reconcilerTimeout)
		err := h.reconcileTeam(ctx, *input)
		if err != nil {
			h.requeueInput(slug, *input)
			log.WithError(err).Error("reconcile team")
		}

		cancel()
		h.setTeamInFlight(slug, false)
	}
}

// RemoveReconciler will remove a reconciler from the list of currently active reconcilers. During the removal of the
// reconciler this function will acquire a lock, preventing other processes from reading from the list.
func (h *handler) RemoveReconciler(reconcilerName sqlc.ReconcilerName) {
	h.lock.Lock()
	defer h.lock.Unlock()

	delete(h.activeReconcilers, reconcilerName)
}

func (h *handler) reconcileTeam(ctx context.Context, input reconcilers.Input) error {
	log := h.log.WithTeamSlug(string(input.Team.Slug))

	if !input.Team.Enabled {
		log.Infof("team is not enabled, skipping reconciliation")
		return nil
	}

	if len(h.activeReconcilers) == 0 {
		log.Warnf("no reconcilers are currently enabled")
		return nil
	}

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
		log.Debugf("successful reconcile duration: %s", duration)

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
	log.Debugf("successful reconcile duration: %s", duration)

	if err := h.database.SetLastSuccessfulSyncForTeam(ctx, input.Team.Slug); err != nil {
		log.WithError(err).Error("update last successful sync timestamp")
	}
	return nil
}

func (h *handler) getReconcilerFactory(reconcilerName sqlc.ReconcilerName) (reconcilers.ReconcilerFactory, error) {
	factory, exists := h.factories[reconcilerName]
	if !exists {
		return nil, fmt.Errorf("missing reconciler factory entry: %q", reconcilerName)
	}
	return factory, nil
}

func (h *handler) UpdateMetrics(ctx context.Context) {
	for ctx.Err() == nil {
		select {
		case <-ctx.Done():
			return
		case <-time.After(10 * time.Second):
			metrics.SetPendingTeamCount(len(h.syncQueueChan))
		}
	}
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

func (h *handler) setTeamInFlight(slug slug.Slug, inFlight bool) {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.teamsInFlight[slug] = inFlight
}

func (h *handler) getTeamInFlight(slug slug.Slug) bool {
	h.lock.Lock()
	defer h.lock.Unlock()

	return h.teamsInFlight[slug]
}

func (h *handler) popInput(slug slug.Slug) *reconcilers.Input {
	h.lock.Lock()
	defer h.lock.Unlock()

	input, ok := h.queueInputs[slug]
	if !ok {
		return nil
	}
	delete(h.queueInputs, slug)
	return &input
}

func (h *handler) requeueInput(slug slug.Slug, input reconcilers.Input) {
	h.lock.Lock()
	defer h.lock.Unlock()

	_, exists := h.queueInputs[slug]
	if !exists {
		input.NumSyncAttempts++
		h.queueInputs[slug] = input
	}
	h.syncQueue.Add(slug)
}
