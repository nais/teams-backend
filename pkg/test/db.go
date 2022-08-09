package test

import (
	"database/sql"
	"github.com/google/uuid"
	sqliteGo "github.com/mattn/go-sqlite3"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/fixtures"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	driverName = "sqlite3_extended"
	dsn        = ":memory:"
)

func newUUID(arguments ...interface{}) (string, error) {
	u, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// GetTestDB Get an in-memory SQLite database instance, used for testing. This function will also run DB migration so
// all tables should be present.
func GetTestDB() (*gorm.DB, error) {
	if !stringInSlice(driverName, sql.Drivers()) {
		sql.Register(driverName,
			&sqliteGo.SQLiteDriver{
				ConnectHook: func(conn *sqliteGo.SQLiteConn) error {
					return conn.RegisterFunc("uuid_generate_v4", newUUID, true)
				},
			},
		)
	}

	conn, _ := sql.Open(driverName, dsn)
	db, _ := gorm.Open(sqlite.Dialector{
		Conn: conn,
	}, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	err := dbmodels.Migrate(db)
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