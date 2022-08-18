package test

import (
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/fixtures"
	"github.com/nais/console/pkg/sqlc"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const dsn = ":memory:"

// GetTestDB Get an in-memory SQLite database instance, used for testing. This function will also run DB migration so
// all tables should be present.
func GetTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = dbmodels.Migrate(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// GetTestDBAndQueries Get an in-memory SQLite database instance, used for testing. This function will also run DB
// migration so all tables should be present. This function will also return the sqlc queries instance.
func GetTestDBAndQueries() (*gorm.DB, *sqlc.Queries, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	dbc, err := db.DB()
	if err != nil {
		return nil, nil, err
	}

	err = dbmodels.Migrate(db)
	if err != nil {
		return nil, nil, err
	}

	return db, sqlc.New(dbc), nil
}

// GetTestDBWithRoles Get a complete test database with all roles added
func GetTestDBWithRoles() (*gorm.DB, error) {
	db, err := GetTestDB()
	if err != nil {
		return nil, err
	}

	err = fixtures.CreateRolesAndAuthorizations(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func GetTestDBAndQueriesWithRoles() (*gorm.DB, *sqlc.Queries, error) {
	db, queries, err := GetTestDBAndQueries()
	if err != nil {
		return nil, nil, err
	}

	err = fixtures.CreateRolesAndAuthorizations(db)
	if err != nil {
		return nil, nil, err
	}

	return db, queries, nil
}
