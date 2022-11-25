package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	graphql_handler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/authn"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/directives"
	"github.com/nais/console/pkg/fixtures"
	"github.com/nais/console/pkg/graph"
	"github.com/nais/console/pkg/graph/apierror"
	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/metrics"
	"github.com/nais/console/pkg/middleware"
	"github.com/nais/console/pkg/reconcilers"
	azure_group_reconciler "github.com/nais/console/pkg/reconcilers/azure/group"
	github_team_reconciler "github.com/nais/console/pkg/reconcilers/github/team"
	google_gcp_reconciler "github.com/nais/console/pkg/reconcilers/google/gcp"
	google_workspace_admin_reconciler "github.com/nais/console/pkg/reconcilers/google/workspace_admin"
	nais_deploy_reconciler "github.com/nais/console/pkg/reconcilers/nais/deploy"
	nais_namespace_reconciler "github.com/nais/console/pkg/reconcilers/nais/namespace"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
	"github.com/nais/console/pkg/usersync"
	"github.com/nais/console/pkg/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		fmt.Printf("fatal: %s", err)
		os.Exit(1)
	}

	log, err := logger.GetLogger(cfg.LogFormat, cfg.LogLevel)
	if err != nil {
		fmt.Printf("fatal: %s", err)
		os.Exit(2)
	}

	err = run(cfg, log)
	if err != nil {
		log.WithError(err).Error("fatal error in run()")
		os.Exit(3)
	}
}

func run(cfg *config.Config, log logger.Logger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bt, _ := version.BuildTime()
	log.Infof("console.nais.io version %s built on %s", version.Version(), bt)

	database, err := setupDatabase(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}

	err = fixtures.CreateAdminServiceAccount(ctx, database, cfg.AdminApiKey)
	if err != nil {
		return err
	}

	err = fixtures.CreateNaisVerification(ctx, database)
	if err != nil {
		return err
	}

	if cfg.StaticServiceAccounts != "" {
		err = fixtures.SetupStaticServiceAccounts(ctx, database, cfg.StaticServiceAccounts)
		if err != nil {
			return err
		}
	}

	// Control channels for goroutine communication
	const maxQueueSize = 4096
	teamReconciler := make(chan reconcilers.Input, maxQueueSize)
	auditLogger := auditlogger.New(database, log)

	authHandler, err := setupAuthHandler(cfg, database, log)
	if err != nil {
		return err
	}

	gcpEnvironments := make([]string, 0)
	if cfg.GCP.Clusters != "" {
		clusterInfo, err := google_gcp_reconciler.GetClusterInfoFromJson(cfg.GCP.Clusters)
		if err != nil {
			return err
		}

		for env := range clusterInfo {
			gcpEnvironments = append(gcpEnvironments, env)
		}
	}

	handler := setupGraphAPI(database, cfg.TenantDomain, teamReconciler, auditLogger.WithSystemName(sqlc.SystemNameGraphqlApi), gcpEnvironments, log)
	srv, err := setupHTTPServer(cfg, database, handler, authHandler)
	if err != nil {
		return err
	}

	log.Infof("Ready to accept requests.")

	go func() {
		err := srv.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Error(err)
		}
		log.Infof("HTTP server finished, terminating...")
		cancel()
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-signals
		log.Infof("Received signal %s, terminating...", sig)
		cancel()
	}()

	const nextReconcileGracePeriod = 15 * time.Second
	const immediateReconcile = 1 * time.Second

	nextReconcile := time.Time{}
	reconcileTimer := time.NewTimer(1 * time.Second)
	reconcileTimer.Stop()

	pendingTeams := make(map[slug.Slug]reconcilers.Input)

	// Reconcile all teams on startup. All will share the same correlation ID
	correlationID, err := uuid.NewUUID()
	if err != nil {
		return fmt.Errorf("cannot create ID for correlation entry for initial reconcile loop: %w", err)
	}

	allTeams, err := database.GetTeams(ctx)
	if err != nil {
		return fmt.Errorf("unable to load team for initial reconcile loop: %w", err)
	}
	for _, team := range allTeams {
		input, err := reconcilers.CreateReconcilerInput(ctx, database, *team)
		if err != nil {
			return fmt.Errorf("unable to create input for initial reconcile loop: %w", err)
		}

		// override correlation id as we want to group all actions in the initial reconcile loop
		teamReconciler <- input.WithCorrelationID(correlationID)
	}

	// User synchronizer
	userSyncTimer := time.NewTimer(1 * time.Second)
	userSyncer, err := usersync.NewFromConfig(cfg, database, auditLogger.WithSystemName(sqlc.SystemNameUsersync), log)
	if err != nil {
		userSyncTimer.Stop()
		if err != usersync.ErrNotEnabled {
			return err
		}

		log.Warnf("User synchronization disabled: %s", err)
	}

	defer log.Infof("Main program context canceled; exiting.")

	for ctx.Err() == nil {
		select {
		case <-ctx.Done():
			return nil

		case input := <-teamReconciler:
			if nextReconcile.Before(time.Now()) {
				nextReconcile = time.Now().Add(immediateReconcile)
				reconcileTimer.Reset(immediateReconcile)
			}

			log.Infof("Scheduling team %q for reconciliation in %s", input.Team.Slug, time.Until(nextReconcile))
			pendingTeams[input.Team.Slug] = input

		case <-reconcileTimer.C:
			log.Infof("Running reconcile of %d teams...", len(pendingTeams))

			err = reconcileTeams(ctx, database, &pendingTeams, cfg, auditLogger, log)

			if err != nil {
				log.Error(err)
				reconcileTimer.Reset(nextReconcileGracePeriod)
			}

			if len(pendingTeams) > 0 {
				log.Warnf("%d teams are not fully reconciled.", len(pendingTeams))
			}

			log.Infof("Reconciliation complete.")

		case <-userSyncTimer.C:
			const interval = time.Hour * 1
			const timeout = time.Second * 30

			log.Infof("Starting user synchronization...")

			ctx, cancel := context.WithTimeout(ctx, timeout)
			err = userSyncer.Sync(ctx)
			cancel()

			if err != nil {
				log.Error(err)
			}

			userSyncTimer.Reset(interval)
			log.Infof("User synchronization complete; next run at %s", time.Now().Add(interval))
		}
	}

	return nil
}

func reconcileTeams(
	ctx context.Context,
	database db.Database,
	reconcileInputs *map[slug.Slug]reconcilers.Input,
	cfg *config.Config,
	auditLogger auditlogger.AuditLogger,
	log logger.Logger,
) error {
	const reconcileTimeout = 15 * time.Minute

	enabledReconcilers, err := initReconcilers(ctx, database, cfg, auditLogger, log)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, reconcileTimeout)
	defer cancel()

	errors := 0
	for teamSlug, input := range *reconcileInputs {
		log := log.WithTeamSlug(string(teamSlug))
		if !input.Team.Enabled {
			log.Infof("team is not enabled, skipping and removing from queue")
			delete(*reconcileInputs, teamSlug)
			continue
		}

		teamErrors := 0

		for _, reconciler := range enabledReconcilers {
			name := reconciler.Name()
			log := log.WithSystem(string(name))
			metrics.IncReconcilerCounter(name, metrics.ReconcilerStateStarted, log)

			err = reconciler.Reconcile(ctx, input)
			if err != nil {
				metrics.IncReconcilerCounter(name, metrics.ReconcilerStateFailed, log)
				log.WithError(err).Error("reconcile")
				teamErrors++
				err = database.SetReconcilerErrorForTeam(ctx, input.CorrelationID, input.Team.Slug, name, err)
				if err != nil {
					log.WithError(err).Error("add reconcile error to database")
				}
				continue
			}

			err = database.ClearReconcilerErrorsForTeam(ctx, input.Team.Slug, name)
			if err != nil {
				log.WithError(err).Error("purge reconcile errors")
			}

			metrics.IncReconcilerCounter(name, metrics.ReconcilerStateSuccessful, log)
		}

		if teamErrors == 0 {
			if err = database.SetLastSuccessfulSyncForTeam(ctx, teamSlug); err != nil {
				log.WithError(err).Error("update last successful sync timestamp")
			}
			delete(*reconcileInputs, teamSlug)
		}
		errors += teamErrors
	}

	if errors > 0 {
		return fmt.Errorf("%d error(s) occurred during reconcile", errors)
	}

	return nil
}

func setupAuthHandler(cfg *config.Config, database db.Database, log logger.Logger) (authn.Handler, error) {
	cf := authn.NewGoogle(cfg.OAuth.ClientID, cfg.OAuth.ClientSecret, cfg.OAuth.RedirectURL)
	frontendURL, err := url.Parse(cfg.FrontendURL)
	if err != nil {
		return nil, err
	}
	handler := authn.New(cf, database, *frontendURL, log)
	return handler, nil
}

func setupDatabase(ctx context.Context, dbUrl string) (db.Database, error) {
	dbc, err := pgxpool.Connect(ctx, dbUrl)
	if err != nil {
		return nil, err
	}

	err = db.Migrate(dbc.Config().ConnString())
	if err != nil {
		return nil, err
	}

	queries := db.Wrap(sqlc.New(dbc), dbc)
	return db.NewDatabase(queries), nil
}

func initReconcilers(
	ctx context.Context,
	database db.Database,
	cfg *config.Config,
	auditLogger auditlogger.AuditLogger,
	log logger.Logger,
) ([]reconcilers.Reconciler, error) {
	factories := map[sqlc.ReconcilerName]reconcilers.ReconcilerFactory{
		azure_group_reconciler.Name:            azure_group_reconciler.NewFromConfig,
		github_team_reconciler.Name:            github_team_reconciler.NewFromConfig,
		google_workspace_admin_reconciler.Name: google_workspace_admin_reconciler.NewFromConfig,
		google_gcp_reconciler.Name:             google_gcp_reconciler.NewFromConfig,
		nais_namespace_reconciler.Name:         nais_namespace_reconciler.NewFromConfig,
		nais_deploy_reconciler.Name:            nais_deploy_reconciler.NewFromConfig,
	}

	enabledReconcilers, err := database.GetEnabledReconcilers(ctx)
	if err != nil {
		return nil, err
	}

	recs := make([]reconcilers.Reconciler, 0)
	for _, reconciler := range enabledReconcilers {
		name := reconciler.Name
		log := log.WithSystem(string(name))

		factory, exists := factories[name]
		if !exists {
			log.WithError(fmt.Errorf("reconciler missing factory entry")).Error("check for factory")
			continue
		}

		rec, err := factory(ctx, database, cfg, auditLogger.WithSystemName(reconcilers.ReconcilerNameToSystemName(name)), log)
		if err != nil {
			log.WithError(err).Error("unable to create reconciler")
			continue
		}

		recs = append(recs, rec)
		log.Info("initialized")
	}

	return recs, nil
}

func setupGraphAPI(database db.Database, domain string, teamReconciler chan<- reconcilers.Input, auditLogger auditlogger.AuditLogger, gcpEnvironments []string, log logger.Logger) *graphql_handler.Server {
	resolver := graph.NewResolver(database, domain, teamReconciler, auditLogger, gcpEnvironments, log)
	gc := generated.Config{}
	gc.Resolvers = resolver
	gc.Directives.Admin = directives.Admin()
	gc.Directives.Auth = directives.Auth()
	gc.Complexity.User.Teams = func(childComplexity int) int {
		return 10 * childComplexity
	}
	gc.Complexity.Team.Members = func(childComplexity int) int {
		return 10 * childComplexity
	}

	handler := graphql_handler.NewDefaultServer(
		generated.NewExecutableSchema(
			gc,
		),
	)
	handler.SetErrorPresenter(apierror.GetErrorPresenter(log))
	handler.Use(extension.FixedComplexityLimit(1000))

	return handler
}

func corsConfig(frontendUrl string) cors.Options {
	return cors.Options{
		AllowedOrigins:   []string{frontendUrl},
		AllowedMethods:   []string{"*"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}
}

func setupHTTPServer(cfg *config.Config, database db.Database, graphApi *graphql_handler.Server, authHandler authn.Handler) (*http.Server, error) {
	r := chi.NewRouter()

	r.Handle("/metrics", promhttp.Handler())
	r.Get("/healthz", func(_ http.ResponseWriter, _ *http.Request) {})

	r.Get("/", playground.Handler("GraphQL playground", "/query"))

	middlewares := []func(http.Handler) http.Handler{
		cors.New(corsConfig(cfg.FrontendURL)).Handler,
		middleware.ApiKeyAuthentication(database),
		middleware.Oauth2Authentication(database, authHandler),
	}

	r.Route("/query", func(r chi.Router) {
		r.Use(middlewares...)
		r.Post("/", graphApi.ServeHTTP)
	})

	r.Route("/oauth2", func(r chi.Router) {
		r.Get("/login", authHandler.Login)
		r.Get("/logout", authHandler.Logout)
		r.Get("/callback", authHandler.Callback)
	})

	srv := &http.Server{
		Addr:    cfg.ListenAddress,
		Handler: r,
	}
	return srv, nil
}
