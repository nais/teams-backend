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

	"github.com/jackc/pgx/v4"
	"github.com/nais/console/pkg/sqlc"

	"github.com/google/uuid"
	console_reconciler "github.com/nais/console/pkg/reconcilers/console"

	graphql_handler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/authn"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/directives"
	"github.com/nais/console/pkg/fixtures"
	"github.com/nais/console/pkg/graph"
	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/middleware"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/reconcilers/registry"
	"github.com/nais/console/pkg/usersync"
	"github.com/nais/console/pkg/version"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	err := run()
	if err != nil {
		log.Errorf("fatal: %s", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bt, _ := version.BuildTime()
	log.Infof("console.nais.io version %s built on %s", version.Version(), bt)

	cfg, err := config.New()
	if err != nil {
		return err
	}

	err = setupLogging(cfg.LogFormat, cfg.LogLevel)
	if err != nil {
		return err
	}

	gormDB, err := setupDatabase(cfg)
	if err != nil {
		return err
	}

	dbc, err := pgx.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	queries := sqlc.New(dbc)

	systems, err := fixtures.CreateReconcilerSystems(ctx, queries)
	if err != nil {
		return err
	}

	err = fixtures.InsertInitialDataset(gormDB, cfg.TenantDomain, cfg.AdminApiKey)
	if err != nil {
		return err
	}

	if cfg.StaticServiceAccounts != "" {
		err = fixtures.SetupStaticServiceAccounts(gormDB, cfg.StaticServiceAccounts, cfg.TenantDomain)
		if err != nil {
			return err
		}
	}

	// Control channels for goroutine communication
	const maxQueueSize = 4096
	teamReconciler := make(chan reconcilers.Input, maxQueueSize)
	logger := auditlogger.New(gormDB)

	recs, err := initReconcilers(gormDB, cfg, logger, systems)
	if err != nil {
		return err
	}

	log.Infof("Initialized %d reconcilers.", len(recs))

	store := authn.NewStore()
	authHandler, err := setupAuthHandler(cfg, store)
	if err != nil {
		return err
	}

	handler, err := setupGraphAPI(queries, gormDB, dbc, cfg.TenantDomain, systems[console_reconciler.Name], teamReconciler, logger)
	if err != nil {
		return err
	}

	srv, err := setupHTTPServer(cfg, gormDB, handler, authHandler, store)
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

	pendingTeams := make(map[uuid.UUID]reconcilers.Input)

	// Reconcile all teams on startup. All will share the same correlation ID
	id, err := uuid.NewUUID()
	if err != nil {
		return fmt.Errorf("cannot create ID for correlation entry for initial reconcile loop: %w", err)
	}
	corr, err := queries.CreateCorrelation(ctx, id)
	if err != nil {
		return fmt.Errorf("cannot create correlation entry for initial reconcile loop: %w", err)
	}

	allTeams, err := queries.GetTeams(ctx)
	if err != nil {
		return fmt.Errorf("unable to load team for initial reconcile loop: %w", err)
	}
	for _, team := range allTeams {
		members, err := queries.GetTeamMembers(ctx, team.ID)
		if err != nil {
			return fmt.Errorf("unable to load team members for team '%s' for the initial reconcile loop: %w", team.Slug, err)
		}

		metadata, err := queries.GetTeamMetadata(ctx, team.ID)
		if err != nil {
			return fmt.Errorf("unable to load team metadata for team '%s' for the initial reconcile loop: %w", team.Slug, err)
		}

		teamReconciler <- reconcilers.Input{
			Corr:     *corr,
			Team:     team,
			Metadata: metadata,
			Members:  members,
		}
	}

	// User synchronizer
	userSyncTimer := time.NewTimer(1 * time.Second)
	userSyncer, err := usersync.NewFromConfig(cfg, queries, gormDB, systems[console_reconciler.Name], logger)
	if err != nil {
		userSyncTimer.Stop()
		if err != usersync.ErrNotEnabled {
			return err
		}

		log.Warnf("User synchronization disabled: %s", err)
	}

	for ctx.Err() == nil {
		select {
		case <-ctx.Done():
			break

		case input := <-teamReconciler:
			if nextReconcile.Before(time.Now()) {
				nextReconcile = time.Now().Add(immediateReconcile)
				reconcileTimer.Reset(immediateReconcile)
			}
			if _, exists := pendingTeams[input.Team.ID]; !exists {
				log.Infof("Scheduling team '%s' for reconciliation in %s", input.Team.Slug, nextReconcile.Sub(time.Now()))
				pendingTeams[input.Team.ID] = input
			}

		case <-reconcileTimer.C:
			log.Infof("Running reconcile of %d teams...", len(pendingTeams))

			err = reconcileTeams(ctx, gormDB, recs, &pendingTeams)

			if err != nil {
				log.Error(err)
				reconcileTimer.Reset(nextReconcileGracePeriod)
			}

			if len(pendingTeams) > 0 {
				log.Warnf("%d teams are not fully reconciled.", len(pendingTeams))
			}

			log.Infof("Reconciliation complete.")

		case <-userSyncTimer.C:
			log.Infof("Starting user synchronization...")

			ctx, cancel := context.WithTimeout(ctx, time.Second*30)
			err = userSyncer.Sync(ctx)
			cancel()

			if err != nil {
				log.Error(err)
			}

			userSyncTimer.Reset(30 * time.Second)
			log.Infof("User synchronization complete.")
		}
	}

	log.Infof("Main program context canceled; exiting.")

	return nil
}

func reconcileTeams(ctx context.Context, db *gorm.DB, recs []reconcilers.Reconciler, reconcileInputs *map[uuid.UUID]reconcilers.Input) error {
	const reconcileTimeout = 15 * time.Minute
	errors := 0

	ctx, cancel := context.WithTimeout(ctx, reconcileTimeout)
	defer cancel()

	for teamId, input := range *reconcileInputs {
		teamErrors := 0

		for _, reconciler := range recs {
			name := reconciler.System().Name
			log.Infof("Starting reconciler '%s' for team: '%s'", name, input.Team.Name)
			err := reconciler.Reconcile(ctx, input)
			if err != nil {
				log.Error(err)
				// TODO: Solve issue with duplicates
				//err = db.Create(&dbmodels.ReconcileError{
				//	CorrelationID: *input.Corr.ID,
				//	SystemID:      *reconciler.System().ID,
				//	TeamID:        *input.Team.ID,
				//	Message:       err.Error(),
				//}).Error
				//
				//if err != nil {
				//	log.Warnf("unable to store reconcile error to database: %s", err)
				//}

				teamErrors++
				continue
			}

			log.Infof("Successfully finished reconciler '%s' for team: '%s'", name, input.Team.Name)
		}

		if teamErrors == 0 {
			delete(*reconcileInputs, teamId)
		}
		errors += teamErrors
	}

	if errors > 0 {
		return fmt.Errorf("%d error(s) occurred during reconcile", errors)
	}

	return nil
}

func setupAuthHandler(cfg *config.Config, store authn.SessionStore) (*authn.Handler, error) {
	cf := authn.NewGoogle(cfg.OAuth.ClientID, cfg.OAuth.ClientSecret, cfg.OAuth.RedirectURL)
	frontendURL, err := url.Parse(cfg.FrontendURL)
	if err != nil {
		return nil, err
	}
	handler := authn.New(cf, store, *frontendURL)
	return handler, nil
}

func setupLogging(format, level string) error {
	switch format {
	case "json":
		log.SetFormatter(&log.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
		})
	case "text":
		log.SetFormatter(&log.TextFormatter{
			TimestampFormat: time.RFC3339Nano,
		})
	default:
		return fmt.Errorf("invalid log format: %s", format)
	}

	lvl, err := log.ParseLevel(level)
	if err != nil {
		return err
	}

	log.SetLevel(lvl)

	return nil
}

func initReconcilers(db *gorm.DB, cfg *config.Config, logger auditlogger.AuditLogger, systems map[string]sqlc.System) ([]reconcilers.Reconciler, error) {
	recs := make([]reconcilers.Reconciler, 0)
	initializers := registry.Reconcilers()

	for _, initializer := range initializers {
		name := initializer.Name
		factory := initializer.Factory
		system, exists := systems[name]
		if !exists {
			return nil, fmt.Errorf("BUG: missing system for reconciler '%s'", name)
		}

		rec, err := factory(db, cfg, system, logger)
		switch err {
		case reconcilers.ErrReconcilerNotEnabled:
			log.Warnf("Reconciler '%s' is disabled through configuration", name)
		default:
			return nil, fmt.Errorf("reconciler '%s': %w", name, err)
		case nil:
			recs = append(recs, rec)
			log.Infof("Reconciler initialized: '%s' -> %T", system.Name, rec)
		}
	}

	return recs, nil
}

func setupDatabase(cfg *config.Config) (*gorm.DB, error) {
	log.Infof("Connecting to database...")
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	log.Infof("Successfully connected to database.")
	log.Infof("Migrating database schema...")
	err = dbmodels.Migrate(db)
	if err != nil {
		return nil, err
	}

	log.Infof("Successfully migrated database schema.")
	return db, nil
}

func setupGraphAPI(queries sqlc.Querier, gormHandle *gorm.DB, db *pgx.Conn, domain string, console sqlc.System, teamReconciler chan<- reconcilers.Input, logger auditlogger.AuditLogger) (*graphql_handler.Server, error) {
	resolver := graph.NewResolver(queries, gormHandle, db, domain, console, teamReconciler, logger)
	gc := generated.Config{}
	gc.Resolvers = resolver
	gc.Directives.Auth = directives.Auth(gormHandle)

	handler := graphql_handler.NewDefaultServer(
		generated.NewExecutableSchema(
			gc,
		),
	)
	handler.SetErrorPresenter(graph.GetErrorPresenter())
	return handler, nil
}

func corsConfig() cors.Options {
	// TODO: Specify a stricter CORS policy
	return cors.Options{
		AllowedOrigins: []string{"http://localhost:*", "https://*"},
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}
}

func setupHTTPServer(cfg *config.Config, db *gorm.DB, graphApi *graphql_handler.Server, authHandler *authn.Handler, store authn.SessionStore) (*http.Server, error) {
	r := chi.NewRouter()

	r.Get("/healthz", func(_ http.ResponseWriter, _ *http.Request) {})

	r.Get("/", playground.Handler("GraphQL playground", "/query"))

	middlewares := []func(http.Handler) http.Handler{
		cors.New(corsConfig()).Handler,
		middleware.ApiKeyAuthentication(db),
		middleware.Oauth2Authentication(db, store),
	}

	// If no other authentication mechanisms produce a authenticated user,
	// fall back to auto-login if it is enabled.
	if len(cfg.AutoLoginUser) > 0 {
		log.Warnf("Auto-login user '%s' is ENABLED for ALL REQUESTS.", cfg.AutoLoginUser)
		log.Warnf("THIS IS A MAJOR SECURITY ISSUE! DO NOT RUN THIS CONFIGURATION IN PRODUCTION!!!")
		middlewares = append(middlewares, middleware.Autologin(db, cfg.AutoLoginUser))
	}

	// Append the role loader middleware after all possible authentication middlewares have been added
	middlewares = append(middlewares, middleware.LoadUserRoles(db))

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
