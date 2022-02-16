package main

import (
	"fmt"

	"github.com/nais/console/pkg/models"
	"github.com/nais/console/pkg/version"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	err := run()
	if err != nil {
		panic(err)
	}
}

func run() error {
	fmt.Println("hello, world!")
	fmt.Printf("version %s\n", version.Version())

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  "user=postgres password=postgres dbname=postgres host=127.0.0.1 port=5432 sslmode=disable",
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{})

	if err != nil {
		return err
	}

	var tx *gorm.DB

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

	return err
}
