package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	graphql_handler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
	"github.com/kelseyhightower/envconfig"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/graph"
	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/middleware"
	"github.com/nais/console/pkg/reconcilers"
	console_reconciler "github.com/nais/console/pkg/reconcilers/console"
	gcp_team_reconciler "github.com/nais/console/pkg/reconcilers/gcp/team"
	"github.com/nais/console/pkg/version"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type systemReconcilerPivot struct {
	system     *dbmodels.System
	reconciler reconcilers.Reconciler
}

type config struct {
	DatabaseURL   string `envconfig:"CONSOLE_DATABASE_URL"`
	ListenAddress string `envconfig:"CONSOLE_LISTEN_ADDRESS"`
}

func defaultconfig() *config {
	return &config{
		DatabaseURL:   "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
		ListenAddress: "127.0.0.1:3000",
	}
}

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

	cfg, err := configure()
	if err != nil {
		return err
	}

	db, err := setupDatabase(cfg)
	if err != nil {
		return err
	}

	// Control channels for goroutine communication
	const maxQueueSize = 4096
	logs := make(chan *dbmodels.AuditLog, maxQueueSize)
	trigger := make(chan *dbmodels.Team, maxQueueSize)

	systems, err := initSystems(ctx, cfg, db, logs)
	if err != nil {
		return err
	}

	handler := setupGraphAPI(db)
	srv, err := setupHTTPServer(cfg, db, handler)
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

	nextRun := time.Time{}
	runTimer := time.NewTimer(1 * time.Second)

	for ctx.Err() == nil {
		select {
		case <-ctx.Done():
			break
		case <-trigger:
			if nextRun.Before(time.Now()) {
				runTimer.Reset(nextRunGracePeriod)
			}
		case <-runTimer.C:
			// run sync
			// fixme: only for specific team
			err = syncAll(ctx, 5*time.Minute, db, systems)
			if err != nil {
				log.Error(err)
				runTimer.Reset(nextRunGracePeriod)
			}
		}
	}

	log.Infof("Main program context canceled; exiting.")

	return nil
}

func syncAll(ctx context.Context, timeout time.Duration, db *gorm.DB, systems map[string]systemReconcilerPivot) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	synchronization := &dbmodels.Synchronization{}
	tx := db.WithContext(ctx).Save(synchronization)
	if tx.Error != nil {
		return fmt.Errorf("cannot create synchronization reference: %w", tx.Error)
	}

	for _, system := range systems {
		input := reconcilers.Input{
			System:          system.system,
			Synchronization: synchronization,
			Team:            nil,
		}

		input.Logger().Infof("Starting reconcile")
		err := system.reconciler.Reconcile(ctx, input)
		if err != nil {
			return err
		}
		input.Logger().Infof("Finished reconcile")
	}

	return nil
}

func setupLogging() {
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
	})

	log.SetLevel(log.DebugLevel)
}

func configure() (*config, error) {
	cfg := defaultconfig()

	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func migrate(db *gorm.DB) error {
	// uuid-ossp is needed for PostgreSQL to generate UUIDs as primary keys
	tx := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
	if tx.Error != nil {
		return tx.Error
	}

	return db.AutoMigrate(
		&dbmodels.AuditLog{},
		&dbmodels.Role{},
		&dbmodels.Synchronization{},
		&dbmodels.System{},
		&dbmodels.TeamMetadata{},
		&dbmodels.Team{},
		&dbmodels.User{},
		&dbmodels.ApiKey{},
	)
}

func initSystems(ctx context.Context, cfg *config, db *gorm.DB, logs chan *dbmodels.AuditLog) (map[string]systemReconcilerPivot, error) {
	recs := []reconcilers.Reconciler{
		console_reconciler.New(logs),
		gcp_team_reconciler.New(logs),
	}

	systems := make(map[string]systemReconcilerPivot)
	for _, reconciler := range recs {
		systemName := reconciler.Name()
		sys := &dbmodels.System{}
		tx := db.WithContext(ctx).First(sys, "name = ?", systemName)

		// System not found in database, try to create
		if sys.ID == nil {
			sys.Name = systemName
			tx = db.WithContext(ctx).Save(sys)
		}

		if tx.Error != nil {
			return nil, tx.Error
		}
		systems[systemName] = systemReconcilerPivot{
			system:     sys,
			reconciler: reconciler,
		}
	}

	// TODO: filter out non-configured systems

	return systems, nil
}

func setupDatabase(cfg *config) (*gorm.DB, error) {
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

	log.Infof("Migrating database schema...")
	err = migrate(db)
	if err != nil {
		return nil, err
	}

	log.Infof("Successfully migrated database schema.")
	return db, nil
}

func setupGraphAPI(db *gorm.DB) *graphql_handler.Server {
	resolver := graph.NewResolver(db)
	gc := generated.Config{}
	gc.Resolvers = resolver
	gc.Directives.Auth = middleware.ApiKeyDirective()
	gc.Directives.Acl = middleware.ACLDirective(db)

	handler := graphql_handler.NewDefaultServer(
		generated.NewExecutableSchema(
			gc,
		),
	)
	return handler
}

func setupHTTPServer(cfg *config, db *gorm.DB, handler *graphql_handler.Server) (*http.Server, error) {
	r := chi.NewRouter()
	r.Get("/", playground.Handler("GraphQL playground", "/query"))
	r.Route("/query", func(r chi.Router) {
		r.Use(middleware.ApiKeyAuthentication(db))
		r.Post("/", handler.ServeHTTP)
	})
	srv := &http.Server{
		Addr:    cfg.ListenAddress,
		Handler: r,
	}
	return srv, nil
}
