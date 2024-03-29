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
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/nais/teams-backend/pkg/auditlogger"
	"github.com/nais/teams-backend/pkg/authn"
	"github.com/nais/teams-backend/pkg/config"
	"github.com/nais/teams-backend/pkg/db"
	"github.com/nais/teams-backend/pkg/deployproxy"
	"github.com/nais/teams-backend/pkg/directives"
	"github.com/nais/teams-backend/pkg/fixtures"
	"github.com/nais/teams-backend/pkg/graph"
	"github.com/nais/teams-backend/pkg/graph/apierror"
	"github.com/nais/teams-backend/pkg/graph/dataloader"
	"github.com/nais/teams-backend/pkg/graph/generated"
	"github.com/nais/teams-backend/pkg/logger"
	"github.com/nais/teams-backend/pkg/middleware"
	"github.com/nais/teams-backend/pkg/teamsync"
	"github.com/nais/teams-backend/pkg/types"
	"github.com/nais/teams-backend/pkg/usersync"
	"github.com/nais/teams-backend/pkg/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	reconcilerWorkers    = 10
	fullTeamSyncInterval = time.Minute * 30
	userSyncInterval     = time.Minute * 15
	userSyncTimeout      = time.Second * 30
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
	log.Infof("teams-backend version %s built on %s", version.Version(), bt)

	database, err := db.New(ctx, cfg.DatabaseURL)
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

	wg := sync.WaitGroup{}
	teamSync := teamsync.NewHandler(ctx, database, cfg, log)
	defer func() {
		teamSync.Close()
		wg.Wait()
	}()
	err = teamSync.InitReconcilers(ctx)
	if err != nil {
		return err
	}

	for i := 0; i < reconcilerWorkers; i++ {
		wg.Add(1)
		go func(ctx context.Context) {
			teamSync.SyncTeams(ctx)
			defer wg.Done()
		}(ctx)
	}

	fullTeamSyncTimer := time.NewTimer(time.Second * 1)
	go teamSync.UpdateMetrics(ctx)

	var userSyncer *usersync.UserSynchronizer
	userSync := make(chan uuid.UUID, 1)
	userSyncTimer := time.NewTimer(10 * time.Second)
	userSyncTimer.Stop()
	userSyncRuns := usersync.NewRunsHandler(cfg.UserSync.RunsToStore)
	if cfg.UserSync.Enabled {
		userSyncer, err = usersync.NewFromConfig(cfg, database, log, userSyncRuns)
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

	handler := setupGraphAPI(teamSync, database, deployProxy, cfg.TenantDomain, userSync, cfg.Environments, log, userSyncRuns)
	srv := setupHTTPServer(cfg, database, handler, authHandler)

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

		case <-fullTeamSyncTimer.C:
			log.Infof("start full team sync")

			correlationID, err := uuid.NewUUID()
			if err != nil {
				log.WithError(err).Errorf("create correlation ID for full team sync")
				fullTeamSyncTimer.Reset(time.Second * 1)
				break
			}

			teams, err := teamSync.ScheduleAllTeams(ctx, correlationID)
			if err != nil {
				log.WithError(err).Errorf("full team sync")
				fullTeamSyncTimer.Reset(time.Second * 1)
				break
			}

			log.Infof("%d teams scheduled for sync", len(teams))
			fullTeamSyncTimer.Reset(fullTeamSyncInterval)
		}
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

func setupGraphAPI(teamSync teamsync.Handler, database db.Database, deployProxy deployproxy.Proxy, domain string, userSync chan<- uuid.UUID, gcpEnvironments []string, log logger.Logger, userSyncRuns *usersync.RunsHandler) *graphql_handler.Server {
	resolver := graph.NewResolver(teamSync, database, deployProxy, domain, userSync, auditlogger.New(database, types.ComponentNameGraphqlApi, log), gcpEnvironments, log, userSyncRuns)
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

func setupHTTPServer(cfg *config.Config, database db.Database, graphApi *graphql_handler.Server, authHandler authn.Handler) *http.Server {
	r := chi.NewRouter()

	r.Handle("/metrics", promhttp.Handler())
	r.Get("/healthz", func(_ http.ResponseWriter, _ *http.Request) {})

	r.Get("/", playground.Handler("GraphQL playground", "/query"))

	iapIssuer := middleware.IAPAuthentication(database, cfg.IAP.Audience)
	if cfg.IAP.Insecure {
		iapIssuer = middleware.IAPInsecureAuthentication(database)
	}

	dataLoaders := dataloader.NewLoaders(database)
	middlewares := []func(http.Handler) http.Handler{
		cors.New(corsConfig(cfg.FrontendURL)).Handler,
		middleware.ApiKeyAuthentication(database),
		iapIssuer,
		middleware.Oauth2Authentication(database, authHandler),
		dataloader.Middleware(dataLoaders),
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

	return &http.Server{
		Addr:    cfg.ListenAddress,
		Handler: r,
	}
}
