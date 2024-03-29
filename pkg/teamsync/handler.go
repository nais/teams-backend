package teamsync

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nais/teams-backend/pkg/config"
	"github.com/nais/teams-backend/pkg/db"
	"github.com/nais/teams-backend/pkg/logger"
	"github.com/nais/teams-backend/pkg/metrics"
	"github.com/nais/teams-backend/pkg/reconcilers"
	azure_group_reconciler "github.com/nais/teams-backend/pkg/reconcilers/azure/group"
	dependencytrack_reconciler "github.com/nais/teams-backend/pkg/reconcilers/dependencytrack"
	github_team_reconciler "github.com/nais/teams-backend/pkg/reconcilers/github/team"
	google_gar "github.com/nais/teams-backend/pkg/reconcilers/google/gar"
	google_gcp_reconciler "github.com/nais/teams-backend/pkg/reconcilers/google/gcp"
	google_workspace_admin_reconciler "github.com/nais/teams-backend/pkg/reconcilers/google/workspace_admin"
	nais_deploy_reconciler "github.com/nais/teams-backend/pkg/reconcilers/nais/deploy"
	nais_namespace_reconciler "github.com/nais/teams-backend/pkg/reconcilers/nais/namespace"
	"github.com/nais/teams-backend/pkg/slug"
	"github.com/nais/teams-backend/pkg/sqlc"
	"github.com/nais/teams-backend/pkg/types"
)

type Handler interface {
	SetReconcilerFactories(factories ReconcilerFactories)
	Schedule(input Input) error
	ScheduleAllTeams(ctx context.Context, correlationID uuid.UUID) ([]*db.Team, error)
	InitReconcilers(ctx context.Context) error
	UseReconciler(reconciler db.Reconciler) error
	RemoveReconciler(reconcilerName sqlc.ReconcilerName)
	SyncTeams(ctx context.Context)
	UpdateMetrics(ctx context.Context)
	DeleteTeam(teamSlug slug.Slug, correlationID uuid.UUID) error
	Close()
}

type handler struct {
	activeReconcilers map[sqlc.ReconcilerName]ReconcilerWithRunOrder
	database          db.Database
	syncQueue         Queue
	syncQueueChan     <-chan Input
	log               logger.Logger
	lock              sync.Mutex
	cfg               *config.Config
	factories         ReconcilerFactories
	mainContext       context.Context

	teamsInFlight     map[slug.Slug]struct{}
	teamsInFlightLock sync.Mutex
}

type ReconcilerWithRunOrder struct {
	runOrder   int32
	reconciler reconcilers.Reconciler
}

type ReconcilerFactories map[sqlc.ReconcilerName]reconcilers.ReconcilerFactory

var factories = ReconcilerFactories{
	dependencytrack_reconciler.Name:        dependencytrack_reconciler.NewFromConfig,
	azure_group_reconciler.Name:            azure_group_reconciler.NewFromConfig,
	github_team_reconciler.Name:            github_team_reconciler.NewFromConfig,
	google_workspace_admin_reconciler.Name: google_workspace_admin_reconciler.NewFromConfig,
	google_gcp_reconciler.Name:             google_gcp_reconciler.NewFromConfig,
	nais_namespace_reconciler.Name:         nais_namespace_reconciler.NewFromConfig,
	nais_deploy_reconciler.Name:            nais_deploy_reconciler.NewFromConfig,
	google_gar.Name:                        google_gar.NewFromConfig,
}

const reconcilerTimeout = time.Minute * 15

func NewHandler(ctx context.Context, database db.Database, cfg *config.Config, log logger.Logger) Handler {
	queue, channel := NewQueue()
	return &handler{
		activeReconcilers: make(map[sqlc.ReconcilerName]ReconcilerWithRunOrder),
		database:          database,
		syncQueue:         queue,
		syncQueueChan:     channel,
		teamsInFlight:     make(map[slug.Slug]struct{}),
		cfg:               cfg,
		log:               log,
		factories:         factories,
		mainContext:       ctx,
	}
}

func (h *handler) DeleteTeam(teamSlug slug.Slug, correlationID uuid.UUID) error {
	log := h.log.WithTeamSlug(string(teamSlug))
	errors := 0

	h.lock.Lock()
	orderedReconcilers := getOrderedReconcilers(h.activeReconcilers)
	h.lock.Unlock()

	for _, reconcilerWithRunOrder := range orderedReconcilers {
		if h.mainContext.Err() != nil {
			return h.mainContext.Err()
		}
		reconcilerImpl := reconcilerWithRunOrder.reconciler
		name := reconcilerImpl.Name()
		log := log.WithComponent(types.ComponentName(name))

		err := reconcilerImpl.Delete(h.mainContext, teamSlug, correlationID)
		if err != nil {
			log.WithError(err).Error("delete team")
			errors++
			continue
		}
	}

	if errors > 0 {
		return fmt.Errorf("%d error(s) occurred during delete", errors)
	}

	err := h.database.DeleteTeam(h.mainContext, teamSlug)
	if err != nil {
		log.WithError(err).Error("delete team from database")
	}

	return nil
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
func (h *handler) Schedule(input Input) error {
	return h.syncQueue.Add(input)
}

// InitReconcilers initializes the currently enabled reconcilers during startup of teams-backend
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
// function will acquire a lock, preventing other processes from reading from the list of active reconcilers. If the
// reconciler is already active it will be re-initialized using its factory function.
func (h *handler) UseReconciler(reconciler db.Reconciler) error {
	h.lock.Lock()
	defer h.lock.Unlock()

	factory, err := h.getReconcilerFactory(reconciler.Name)
	if err != nil {
		return err
	}

	reconcilerImplementation, err := factory(h.mainContext, h.database, h.cfg, h.log)
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
	for input := range h.syncQueueChan {
		log := h.log.WithTeamSlug(string(input.TeamSlug))

		if !h.setTeamInFlight(input.TeamSlug) {
			log.Info("already in flight - adding to back of queue")
			time.Sleep(100 * time.Millisecond)
			if err := h.syncQueue.Add(input); err != nil {
				log.WithError(err).Error("failed while re-queueing team that is in flight")
			}
			continue
		}

		ctx, cancel := context.WithTimeout(ctx, reconcilerTimeout)
		err := h.reconcileTeam(ctx, input)
		if err != nil {
			log.WithError(err).Error("reconcile team")
		}

		cancel()
		h.unsetTeamInFlight(input.TeamSlug)
	}
}

// RemoveReconciler will remove a reconciler from the list of currently active reconcilers. During the removal of the
// reconciler this function will acquire a lock, preventing other processes from reading from the list.
func (h *handler) RemoveReconciler(reconcilerName sqlc.ReconcilerName) {
	h.lock.Lock()
	defer h.lock.Unlock()

	delete(h.activeReconcilers, reconcilerName)
}

func (h *handler) ScheduleAllTeams(ctx context.Context, correlationID uuid.UUID) ([]*db.Team, error) {
	teams, err := h.database.GetActiveTeams(ctx)
	if err != nil {
		return nil, err
	}

	for _, team := range teams {
		input := Input{
			TeamSlug:      team.Slug,
			CorrelationID: correlationID,
		}

		err = h.Schedule(input)
		if err != nil {
			return nil, err
		}
	}

	return teams, nil
}

func (h *handler) reconcileTeam(ctx context.Context, input Input) error {
	log := h.log.WithTeamSlug(string(input.TeamSlug))

	team, err := h.database.GetActiveTeamBySlug(ctx, input.TeamSlug)
	if err != nil {
		return err
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
		log := log.WithComponent(types.ComponentName(name))

		reconcilerInput, err := reconcilers.CreateReconcilerInput(ctx, h.database, *team, name)
		if err != nil {
			log.WithError(err).Errorf("get team members for reconciler")
			continue
		}

		metrics.IncReconcilerCounter(name, metrics.ReconcilerStateStarted)

		reconcileTimer := metrics.MeasureReconcilerDuration(name)
		err = reconcilerImpl.Reconcile(ctx, reconcilerInput)
		if err != nil {
			metrics.IncReconcilerCounter(name, metrics.ReconcilerStateFailed)
			log.WithError(err).Error("reconcile")
			errors++
			err = h.database.SetReconcilerErrorForTeam(ctx, input.CorrelationID, team.Slug, name, err)
			if err != nil {
				log.WithError(err).Error("add reconcile error to database")
			}
			continue
		}
		duration := reconcileTimer.ObserveDuration()
		log.Debugf("successful reconcile duration: %s", duration)

		err = h.database.ClearReconcilerErrorsForTeam(ctx, team.Slug, name)
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

	if err := h.database.SetLastSuccessfulSyncForTeam(ctx, team.Slug); err != nil {
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

func (h *handler) setTeamInFlight(slug slug.Slug) bool {
	h.teamsInFlightLock.Lock()
	defer h.teamsInFlightLock.Unlock()

	if _, inFlight := h.teamsInFlight[slug]; !inFlight {
		h.teamsInFlight[slug] = struct{}{}
		return true
	}
	return false
}

func (h *handler) unsetTeamInFlight(slug slug.Slug) {
	h.teamsInFlightLock.Lock()
	defer h.teamsInFlightLock.Unlock()

	delete(h.teamsInFlight, slug)
}
