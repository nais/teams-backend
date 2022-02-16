package main

import (
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kelseyhightower/envconfig"
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

func run() error {
	cfg := defaultconfig()

	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
	})
	log.SetLevel(log.DebugLevel)

	bt, _ := version.BuildTime()
	log.Infof("console.nais.io version %s built on %s", version.Version(), bt)

	err := envconfig.Process("", cfg)
	if err != nil {
		return err
	}

	log.Infof("Connecting to database...")
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  cfg.DatabaseURL,
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{})

	if err != nil {
		return err
	}

	log.Infof("Successfully connected to database.")

	var tx *gorm.DB

	log.Infof("Migrating database schema...")

	tx = db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
	if tx.Error != nil {
		return tx.Error
	}

	err = db.AutoMigrate(
		&models.AuditLog{},
		&models.Role{},
		&models.Synchronization{},
		&models.System{},
		&models.TeamMetadata{},
		&models.Team{},
		&models.User{},
	)
	if err != nil {
		return err
	}

	log.Infof("Successfully migrated database schema.")

	sock, err := net.Listen("tcp", cfg.ListenAddress)
	if err != nil {
		return err
	}
	defer sock.Close()

	router := gin.New()
	router.GET("/ping", handler)

	return router.RunListener(sock)
}

func handler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, nil)
}
