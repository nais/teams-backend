package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql"
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
	"github.com/vektah/gqlparser/v2/gqlerror"
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
	setupLogging()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bt, _ := version.BuildTime()
	log.Infof("console.nais.io version %s built on %s", version.Version(), bt)

	cfg, err := config.New()
	if err != nil {
		return err
	}

	db, err := setupDatabase(cfg)
	if err != nil {
		return err
	}

	err = fixtures.EnsureSystemsExistInDatabase(ctx, db)
	if err != nil {
		return err
	}

	err = fixtures.InsertRootUser(ctx, db)
	if err != nil {
		return err
	}

	sysEntries, err := systems(ctx, db)
	if err != nil {
		return err
	}

	// Control channels for goroutine communication
	const maxQueueSize = 4096
	logs := make(chan *dbmodels.AuditLog, maxQueueSize)
	trigger := make(chan *dbmodels.Team, maxQueueSize)
	logger := auditlogger.New(logs)

	recs, err := initReconcilers(cfg, logger, sysEntries)
	if err != nil {
		return err
	}

	for system, rec := range recs {
		log.Infof("Reconciler initialized: '%s' -> %T", system.Name, rec)
	}
	log.Infof("Initialized %d reconcilers.", len(recs))

	store := authn.NewStore()
	authhandler := setupAuthHandler(cfg, store)
	handler := setupGraphAPI(db, sysEntries["console"], trigger)
	srv, err := setupHTTPServer(cfg, db, handler, authhandler, store)
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

	const nextRunGracePeriod = 15 * time.Second
	const immediateRun = 1 * time.Second
	const syncTimeout = 15 * time.Minute

	nextRun := time.Time{}
	runTimer := time.NewTimer(1 * time.Second)
	runTimer.Stop()
	pendingTeams := make(map[string]*dbmodels.Team)

	// Synchronize every team on startup
	allTeams := make([]*dbmodels.Team, 0)
	db.Preload("Users").Find(&allTeams)
	for _, team := range allTeams {
		trigger <- team
	}

	// Asynchronously record all audit log in database
	go func() {
		for logLine := range logs {
			tx := db.Save(logLine)
			if tx.Error != nil {
				log.Errorf("store audit log line in database: %s", tx.Error)
			}
			logLine.Log().Infof(logLine.Message)
		}
	}()

	// User synchronizer
	userSyncTimer := time.NewTimer(1 * time.Second)
	usersyncer, err := usersync.NewFromConfig(cfg, db, logger)
	if err != nil {
		userSyncTimer.Stop()
		if err == usersync.ErrNotEnabled {
			log.Warnf("User synchronization disabled: %s", err)
		} else {
			return err
		}
	}

	for ctx.Err() == nil {
		select {
		case <-ctx.Done():
			break

		case team := <-trigger:
			if nextRun.Before(time.Now()) {
				nextRun = time.Now().Add(immediateRun)
				runTimer.Reset(immediateRun)
			}
			if pendingTeams[*team.Slug] == nil {
				log.Infof("Scheduling team '%s' for reconciliation in %s", *team.Slug, nextRun.Sub(time.Now()))
				pendingTeams[*team.Slug] = team
			}

		case <-runTimer.C:
			log.Infof("Running reconcile of %d teams...", len(pendingTeams))

			err = syncAll(ctx, syncTimeout, db, recs, &pendingTeams)

			if err != nil {
				log.Error(err)
				runTimer.Reset(nextRunGracePeriod)
			}

			if len(pendingTeams) > 0 {
				log.Warnf("%d teams are not fully reconciled.", len(pendingTeams))
			}

			log.Infof("Reconciliation complete.")

		case <-userSyncTimer.C:
			log.Infof("Starting user synchronization...")

			ctx, cancel := context.WithTimeout(ctx, time.Second*30)
			err = usersyncer.FetchAll(ctx)
			cancel()

			switch er := err.(type) {
			case nil:
				break
			case *dbmodels.AuditLog:
				er.Log().Error(er.Message)
			default:
				log.Error(err)
			}

			userSyncTimer.Reset(30 * time.Second)
			log.Infof("User synchronization complete.")
		}
	}

	log.Infof("Main program context canceled; exiting.")

	return nil
}

func syncAll(ctx context.Context, timeout time.Duration, db *gorm.DB, systems map[*dbmodels.System]reconcilers.Reconciler, teams *map[string]*dbmodels.Team) error {
	errors := 0

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	synchronization := &dbmodels.Synchronization{}
	tx := db.WithContext(ctx).Save(synchronization)
	if tx.Error != nil {
		return fmt.Errorf("cannot create synchronization reference: %w", tx.Error)
	}

	for key, team := range *teams {
		teamErrors := 0

		for system, reconciler := range systems {
			input := reconcilers.Input{
				System:          system,
				Synchronization: synchronization,
				Team:            team,
			}

			if input.Team == nil || input.Team.Slug == nil {
				panic("BUG: refusing to create team with empty slug")
			}

			log.Info(input.AuditLog(nil, true, "console:reconcile:start", "Starting reconcile"))
			err := reconciler.Reconcile(ctx, input)

			switch er := err.(type) {
			case nil:
				log.Info(input.AuditLog(nil, true, "console:reconcile:end", "Successfully reconciled"))
			case *dbmodels.AuditLog:
				er.Log().Error(er.Message)
				teamErrors++
			case error:
				log.Error(input.AuditLog(nil, true, "console:reconcile:end", er.Error()))
				teamErrors++
			}
		}

		if teamErrors == 0 {
			delete(*teams, key)
		}
		errors += teamErrors
	}

	if errors > 0 {
		return fmt.Errorf("%d systems returned errors during reconcile", errors)
	}

	return nil
}

func setupAuthHandler(cfg *config.Config, store authn.SessionStore) *authn.Handler {
	cf := authn.NewGoogle(cfg.OAuth.ClientID, cfg.OAuth.ClientSecret, cfg.OAuth.RedirectURL)
	handler := authn.New(cf, store)
	return handler
}

func setupLogging() {
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
	})
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: time.RFC3339Nano,
	})

	log.SetLevel(log.DebugLevel)
}

func systems(ctx context.Context, db *gorm.DB) (map[string]*dbmodels.System, error) {
	dbsystems := make([]*dbmodels.System, 0)
	tx := db.WithContext(ctx).Find(&dbsystems)
	if tx.Error != nil {
		return nil, tx.Error
	}
	results := make(map[string]*dbmodels.System)
	for _, sys := range dbsystems {
		results[sys.Name] = sys
	}
	return results, nil
}

func initReconcilers(cfg *config.Config, logger auditlogger.Logger, systems map[string]*dbmodels.System) (map[*dbmodels.System]reconcilers.Reconciler, error) {
	recs := make(map[*dbmodels.System]reconcilers.Reconciler)

	factories := registry.Reconcilers()
	for key, factory := range factories {
		rec, err := factory(cfg, logger)
		switch err {
		case reconcilers.ErrReconcilerNotEnabled:
			log.Warnf("Reconciler '%s' is disabled through configuration", key)
		default:
			return nil, fmt.Errorf("reconciler '%s': %w", key, err)
		case nil:
			recs[systems[key]] = rec
		}
	}

	return recs, nil
}

func setupDatabase(cfg *config.Config) (*gorm.DB, error) {
	log.Infof("Connecting to database...")
	db, err := gorm.Open(
		postgres.New(
			postgres.Config{
				DSN:                  cfg.DatabaseURL,
				PreferSimpleProtocol: true, // disables implicit prepared statement usage
			},
		),
		&gorm.Config{},
	)
	if err != nil {
		return nil, err
	}
	log.Infof("Successfully connected to database.")

	// uuid-ossp is needed for PostgreSQL to generate UUIDs as primary keys
	tx := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
	if tx.Error != nil {
		return nil, fmt.Errorf("install postgres uuid extension: %w", tx.Error)
	}

	log.Infof("Migrating database schema...")
	err = dbmodels.Migrate(db)
	if err != nil {
		return nil, err
	}

	log.Infof("Successfully migrated database schema.")
	return db, nil
}

func setupGraphAPI(db *gorm.DB, console *dbmodels.System, trigger chan<- *dbmodels.Team) *graphql_handler.Server {
	resolver := graph.NewResolver(db, console, trigger)
	gc := generated.Config{}
	gc.Resolvers = resolver
	gc.Directives.Auth = directives.Auth(db)

	handler := graphql_handler.NewDefaultServer(
		generated.NewExecutableSchema(
			gc,
		),
	)
	handler.SetErrorPresenter(func(ctx context.Context, e error) *gqlerror.Error {
		err := graphql.DefaultErrorPresenter(ctx, e)

		if errors.Is(err, gorm.ErrRecordNotFound) {
			err.Message = "Not found"
			err.Extensions = map[string]interface{}{
				"code": "404",
			}
		}

		return err
	})
	return handler
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

func setupHTTPServer(cfg *config.Config, db *gorm.DB, graphapi *graphql_handler.Server, authhandler *authn.Handler, store authn.SessionStore) (*http.Server, error) {
	r := chi.NewRouter()

	r.Get("/", playground.Handler("GraphQL playground", "/query"))
	r.Route("/query", func(r chi.Router) {
		r.Use(
			cors.New(corsConfig()).Handler,
			middleware.ApiKeyAuthentication(db),
			middleware.Oauth2Authentication(db, store),
		)
		r.Post("/", graphapi.ServeHTTP)
	})

	r.Route("/oauth2", func(r chi.Router) {
		r.Get("/login", authhandler.Login)
		r.Get("/logout", authhandler.Logout)
		r.Get("/callback", authhandler.Callback)
	})

	srv := &http.Server{
		Addr:    cfg.ListenAddress,
		Handler: r,
	}
	return srv, nil
}
