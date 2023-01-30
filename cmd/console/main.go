package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
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
	"github.com/nais/console/pkg/deployproxy"
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
	"github.com/nais/console/pkg/sqlc"
	"github.com/nais/console/pkg/usersync"
	"github.com/nais/console/pkg/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	reconcilerTimeout = time.Minute * 15
	reconcilerWorkers = 10
	userSyncInterval  = time.Minute * 15
	userSyncTimeout   = time.Second * 30
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

	firstRun, err := database.IsFirstRun(ctx)
	if err != nil {
		return err
	}
	if firstRun {
		log.Infof("first run detected ")
		firstRunLogger := log.WithField("system", "first-run")
		if err := fixtures.SetupDefaultReconcilers(ctx, firstRunLogger, cfg.FirstRunEnableReconcilers, database); err != nil {
			return err
		}

		if err := database.FirstRunComplete(ctx); err != nil {
			return err
		}
	}

	err = fixtures.SetupStaticServiceAccounts(ctx, database, cfg.StaticServiceAccounts)
	if err != nil {
		return err
	}

	auditLogger := auditlogger.New(log)

	teamReconcilerQueue, teamReconcilerQueueChannel := reconcilers.NewReconcilerQueue()
	wg := sync.WaitGroup{}
	defer func() {
		teamReconcilerQueue.Close()
		wg.Wait()
	}()

	for i := 0; i < reconcilerWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for input := range teamReconcilerQueueChannel {
				ctx, cancel := context.WithTimeout(ctx, reconcilerTimeout)
				err = reconcileTeam(ctx, database, input, cfg, auditLogger, log)
				if err != nil {
					log.WithTeamSlug(string(input.Team.Slug)).WithError(err).Error("reconcile team")
					teamReconcilerQueue.Add(input)
				}

				cancel()
			}
		}()
	}

	var userSyncer *usersync.UserSynchronizer
	userSync := make(chan uuid.UUID, 1)
	userSyncTimer := time.NewTimer(10 * time.Second)
	userSyncTimer.Stop()
	if cfg.UserSync.Enabled {
		userSyncer, err = usersync.NewFromConfig(cfg, database, auditLogger.WithSystemName(sqlc.SystemNameUsersync), log)
		if err != nil {
			return err
		}

		userSyncTimer.Reset(time.Second * 1)
	}

	authHandler, err := setupAuthHandler(cfg, database, log)
	if err != nil {
		return err
	}

	deployProxy, err := deployproxy.NewProxy(cfg.NaisDeploy.DeployKeyEndpoint, cfg.NaisDeploy.ProvisionKey, log)
	if err != nil {
		log.Warnf("Deploy proxy is not configured: %v", err)
	}

	handler := setupGraphAPI(database, deployProxy, cfg.TenantDomain, teamReconcilerQueue, userSync, auditLogger.WithSystemName(sqlc.SystemNameGraphqlApi), cfg.Environments, log)
	srv, err := setupHTTPServer(cfg, database, handler, authHandler)
	if err != nil {
		return err
	}

	log.Infof("ready to accept requests at %s.", cfg.ListenAddress)

	go func() {
		err := srv.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Error(err)
		}
		log.Info("HTTP server finished, terminating...")
		cancel()
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-signals
		log.Infof("received signal %s, terminating...", sig)
		cancel()
	}()

	defer log.Info("main program context canceled; exiting.")

	for ctx.Err() == nil {
		metrics.SetPendingTeamCount(len(teamReconcilerQueueChannel))

		select {
		case <-ctx.Done():
			return nil

		case correlationID := <-userSync:
			if userSyncer == nil {
				log.Infof("user sync is disabled")
				break
			}

			log.Debug("starting user synchronization...")
			ctx, cancel := context.WithTimeout(ctx, userSyncTimeout)
			err = userSyncer.Sync(ctx, correlationID)
			cancel()

			if err != nil {
				log.WithError(err).Error("sync users")
			}

			log.Debugf("user sync complete")

		case <-userSyncTimer.C:
			nextUserSync := time.Now().Add(userSyncInterval)
			userSyncTimer.Reset(userSyncInterval)
			log.Debugf("scheduled user sync triggered; next run at %s", nextUserSync)

			correlationID, err := uuid.NewUUID()
			if err != nil {
				log.WithError(err).Errorf("unable to create correlation ID for user sync")
				break
			}

			userSync <- correlationID
		}
	}

	return nil
}

func reconcileTeam(ctx context.Context, database db.Database, input reconcilers.Input, cfg *config.Config, auditLogger auditlogger.AuditLogger, log logger.Logger) error {
	enabledReconcilers, err := initReconcilers(ctx, database, cfg, auditLogger, log)
	if err != nil {
		return err
	}

	log = log.WithTeamSlug(string(input.Team.Slug))

	if !input.Team.Enabled {
		log.Info("team is not enabled, skipping and removing from queue")
		return nil
	}

	errors := 0

	for _, reconciler := range enabledReconcilers {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		name := reconciler.Name()
		log = log.WithSystem(string(name))
		metrics.IncReconcilerCounter(name, metrics.ReconcilerStateStarted)

		err = reconciler.Reconcile(ctx, input)
		if err != nil {
			metrics.IncReconcilerCounter(name, metrics.ReconcilerStateFailed)
			log.WithError(err).Error("reconcile")
			errors++
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

		metrics.IncReconcilerCounter(name, metrics.ReconcilerStateSuccessful)
	}

	if errors > 0 {
		return fmt.Errorf("%d error(s) occurred during reconcile", errors)
	}

	if err = database.SetLastSuccessfulSyncForTeam(ctx, input.Team.Slug); err != nil {
		log.WithError(err).Error("update last successful sync timestamp")
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
			log.WithReconciler(string(name)).WithError(err).Error("initialize")
			continue
		}

		recs = append(recs, rec)
		log.WithReconciler(string(name)).Info("initialized")
	}

	return recs, nil
}

func setupGraphAPI(database db.Database, deployProxy deployproxy.Proxy, domain string, teamReconcilerQueue reconcilers.ReconcilerQueue, userSync chan<- uuid.UUID, auditLogger auditlogger.AuditLogger, gcpEnvironments []string, log logger.Logger) *graphql_handler.Server {
	resolver := graph.NewResolver(database, deployProxy, domain, teamReconcilerQueue, userSync, auditLogger, gcpEnvironments, log)
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
