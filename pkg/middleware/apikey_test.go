package middleware

import (
	"database/sql"
	"testing"

	"github.com/google/uuid"
	sqliteGo "github.com/mattn/go-sqlite3"
	"github.com/nais/console/pkg/dbmodels"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupDatabase() (*gorm.DB, error) {
	sql.Register("sqlite3_extended",
		&sqliteGo.SQLiteDriver{
			ConnectHook: func(conn *sqliteGo.SQLiteConn) error {
				err := conn.RegisterFunc(
					"uuid_generate_v4",
					func(arguments ...interface{}) (string, error) {
						u, err := uuid.NewUUID()
						if err != nil {
							return "", err
						}
						return u.String(), nil
					},
					true,
				)
				return err
			},
		},
	)

	db, err := gorm.Open(
		sqlite.Open(":memory:"),
		&gorm.Config{},
	)

	if err != nil {
		return nil, err
	}

	err = dbmodels.Migrate(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func setupFixtures(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		tx.Create(&dbmodels.User{
			Model:        dbmodels.Model{},
			SoftDeletes:  dbmodels.SoftDeletes{},
			Email:        nil,
			Name:         nil,
			Teams:        nil,
			RoleBindings: nil,
		})
		return nil
	})
}

func TestApiKeyAuthentication(t *testing.T) {
	db, err := setupDatabase()
	if err != nil {
		panic(err)
	}

	err = setupFixtures(db)
	if err != nil {
		panic(err)
	}

	_ = db
}
