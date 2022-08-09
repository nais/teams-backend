package test

import (
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/fixtures"
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
