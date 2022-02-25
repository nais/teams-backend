package main

import (
	"context"
	"net"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql"
	graphql_handler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/kelseyhightower/envconfig"
	"github.com/nais/console/pkg/graph"
	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/models"
	"github.com/nais/console/pkg/version"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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

func setupLogging() {
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
	})

	log.SetLevel(log.DebugLevel)

	gin.DefaultWriter = log.StandardLogger().WriterLevel(log.DebugLevel)
	gin.DefaultErrorWriter = log.StandardLogger().WriterLevel(log.ErrorLevel)

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
		&models.AuditLog{},
		&models.Role{},
		&models.Synchronization{},
		&models.System{},
		&models.TeamMetadata{},
		&models.Team{},
		&models.User{},
		&models.ApiKey{},
	)
}

func run() error {
	setupLogging()

	bt, _ := version.BuildTime()
	log.Infof("console.nais.io version %s built on %s", version.Version(), bt)

	cfg, err := configure()
	if err != nil {
		return err
	}

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
		return err
	}
	log.Infof("Successfully connected to database.")

	log.Infof("Migrating database schema...")
	err = migrate(db)
	if err != nil {
		return err
	}
	log.Infof("Successfully migrated database schema.")

	resolver := graph.NewResolver(db)
	gc := generated.Config{}
	gc.Resolvers = resolver
	gc.Directives.Authentication = func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
		// FIXME: implement authentication
		ctx = context.WithValue(ctx, "authenticated", true)
		return next(ctx)
	}

	handler := graphql_handler.NewDefaultServer(
		generated.NewExecutableSchema(
			gc,
		),
	)

	sock, err := net.Listen("tcp", cfg.ListenAddress)
	if err != nil {
		return err
	}
	defer sock.Close()

	router := gin.New()
	router.GET("/", gin.WrapH(playground.Handler("GraphQL playground", "/query")))
	router.POST("/query", gin.WrapH(handler))

	log.Infof("Ready to accept requests.")

	return router.RunListener(sock)
}
