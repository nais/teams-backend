package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v43/github"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	github_team_reconciler "github.com/nais/console/pkg/reconcilers/github/team"
	gcp_team_reconciler "github.com/nais/console/pkg/reconcilers/google/workspace_admin"
	nais_deploy_reconciler "github.com/nais/console/pkg/reconcilers/nais/deploy"
	"github.com/shurcooL/githubv4"

	graphql_handler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/graph"
	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/middleware"
	"github.com/nais/console/pkg/reconcilers"
	console_reconciler "github.com/nais/console/pkg/reconcilers/console"
	"github.com/nais/console/pkg/version"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	admin "google.golang.org/api/admin/directory/v1"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type systemReconcilerPivot struct {
	system     *dbmodels.System
	reconciler reconcilers.Reconciler
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

	cfg, err := config.New()
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

	recs := initReconcilers(cfg, logs)
	for _, rec := range recs {
		log.Infof("Reconciler initialized: %s", rec.Name())
	}

	systems, err := initSystems(ctx, recs, db)
	if err != nil {
		return err
	}

	handler := setupGraphAPI(db, trigger)
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

			err = syncAll(ctx, syncTimeout, db, systems, &pendingTeams)

			if err != nil {
				log.Error(err)
				runTimer.Reset(nextRunGracePeriod)
			}

			if len(pendingTeams) > 0 {
				log.Warnf("%d teams are not fully reconciled.", len(pendingTeams))
			}

			log.Infof("Reconciliation complete.")
		}
	}

	log.Infof("Main program context canceled; exiting.")

	return nil
}

func syncAll(ctx context.Context, timeout time.Duration, db *gorm.DB, systems map[string]systemReconcilerPivot, teams *map[string]*dbmodels.Team) error {
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

		for _, system := range systems {
			input := reconcilers.Input{
				System:          system.system,
				Synchronization: synchronization,
				Team:            team,
			}

			input.Logger().Infof("Starting reconcile")
			err := system.reconciler.Reconcile(ctx, input)

			switch er := err.(type) {
			case nil:
				input.Logger().Infof("Successfully reconciled")
			case *dbmodels.AuditLog:
				er.Log().Error(er.Message)
				teamErrors++
			case error:
				input.Logger().Error(er)
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

func setupLogging() {
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
	})
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: time.RFC3339Nano,
	})

	log.SetLevel(log.DebugLevel)
}

func migrate(db *gorm.DB) error {
	// uuid-ossp is needed for PostgreSQL to generate UUIDs as primary keys
	tx := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
	if tx.Error != nil {
		return tx.Error
	}

	return db.AutoMigrate(
		&dbmodels.ApiKey{},
		&dbmodels.AuditLog{},
		&dbmodels.Role{},
		&dbmodels.Synchronization{},
		&dbmodels.System{},
		&dbmodels.TeamMetadata{},
		&dbmodels.Team{},
		&dbmodels.User{},
	)
}

func initGCP(cfg *config.Config) (*jwt.Config, error) {
	b, err := ioutil.ReadFile(cfg.Google.CredentialsFile)
	if err != nil {
		return nil, fmt.Errorf("read google credentials file: %w", err)
	}

	cf, err := google.JWTConfigFromJSON(
		b,
		admin.AdminDirectoryUserReadonlyScope,
		admin.AdminDirectoryGroupScope,
	)
	if err != nil {
		return nil, fmt.Errorf("initialize google credentials: %w", err)
	}

	cf.Subject = cfg.Google.DelegatedUser

	return cf, nil
}

func initReconcilers(cfg *config.Config, logs chan *dbmodels.AuditLog) []reconcilers.Reconciler {
	logger := auditlogger.New(logs)
	recs := make([]reconcilers.Reconciler, 0)

	// Internal system
	recs = append(recs, console_reconciler.New(logs))

	// GCP
	googleJWT, err := initGCP(cfg)
	if err == nil {
		recs = append(recs, gcp_team_reconciler.New(logger, cfg.Google.Domain, googleJWT))
	} else {
		log.Warnf("GCP team reconciler not configured: %s", err)
	}

	// GitHub
	gh, err := initGitHub(cfg, logger)
	if err == nil {
		recs = append(recs, gh)
	} else {
		log.Warnf("GitHub team reconciler not configured: %s", err)
	}

	// NAIS deploy
	nd, err := initNaisDeploy(cfg, logger)
	if err == nil {
		recs = append(recs, nd)
	} else {
		log.Warnf("NAIS deploy reconciler not configured: %s", err)
	}

	return recs
}

func initGitHub(cfg *config.Config, logger auditlogger.Logger) (reconcilers.Reconciler, error) {
	transport, err := ghinstallation.NewKeyFromFile(
		http.DefaultTransport,
		cfg.GitHub.AppId,
		cfg.GitHub.AppInstallationId,
		cfg.GitHub.PrivateKeyPath,
	)
	if err != nil {
		return nil, err
	}

	// Note that both HTTP clients and transports are safe for concurrent use according to the docs,
	// so we can safely reuse them across objects and concurrent synchronizations.
	httpClient := &http.Client{
		Transport: transport,
	}
	restClient := github.NewClient(httpClient)
	graphClient := githubv4.NewClient(httpClient)

	return github_team_reconciler.New(logger, cfg.GitHub.Organization, restClient.Teams, graphClient), nil
}

func initNaisDeploy(cfg *config.Config, logger auditlogger.Logger) (reconcilers.Reconciler, error) {
	provisionKey, err := hex.DecodeString(cfg.NaisDeploy.ProvisionKey)
	if err != nil {
		return nil, err
	}

	return nais_deploy_reconciler.New(logger, cfg.NaisDeploy.Endpoint, provisionKey), nil
}

func initSystems(ctx context.Context, recs []reconcilers.Reconciler, db *gorm.DB) (map[string]systemReconcilerPivot, error) {
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

	log.Infof("Migrating database schema...")
	err = migrate(db)
	if err != nil {
		return nil, err
	}

	log.Infof("Successfully migrated database schema.")
	return db, nil
}

func setupGraphAPI(db *gorm.DB, trigger chan<- *dbmodels.Team) *graphql_handler.Server {
	resolver := graph.NewResolver(db, trigger)
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

func setupHTTPServer(cfg *config.Config, db *gorm.DB, handler *graphql_handler.Server) (*http.Server, error) {
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
