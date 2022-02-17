package main

import (
	"github.com/google/uuid"
	"github.com/loopfz/gadgeto/tonic"
	"github.com/wI2L/fizz"
	"github.com/wI2L/fizz/openapi"
	"net"
	"os"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kelseyhightower/envconfig"
	"github.com/nais/console/pkg/apiserver"
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

	sock, err := net.Listen("tcp", cfg.ListenAddress)
	if err != nil {
		return err
	}
	defer sock.Close()

	srv := apiserver.New(db)

	router := gin.New()
	router.Use(gin.Recovery())

	f := fizz.NewFromEngine(router)
	err = f.Generator().OverrideDataType(reflect.TypeOf(&uuid.UUID{}), "string", "uuid")
	if err != nil {
		return err
	}

	infos := &openapi.Info{
		Title:       "NAIS Console",
		Description: `NAIS Console`,
		Version:     "1.0.0",
	}
	f.GET("/openapi.yaml", nil, f.OpenAPI(infos, "yaml"))

	v1 := f.Group("/api/v1", "Version 1", "Version 1 of the API")
	v1.POST("/teams", nil, tonic.Handler(srv.PostTeam, 201))
	v1.GET("/teams", nil, tonic.Handler(srv.GetTeams, 200))
	v1.GET("/teams/:id", nil, tonic.Handler(srv.GetTeam, 200))
	v1.PUT("/teams/:id", nil, srv.PutTeam)
	v1.DELETE("/teams/:id", nil, srv.DeleteTeam)

	return router.RunListener(sock)
}
