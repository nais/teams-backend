//go:build adhoc_integration_test

package nais_namespace_reconciler_test

import (
	"context"
	"database/sql"
	helpers "github.com/nais/console/pkg/console"
	"testing"

	"github.com/google/uuid"
	sqliteGo "github.com/mattn/go-sqlite3"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/reconcilers"
	nais_namespace_reconciler "github.com/nais/console/pkg/reconcilers/nais/namespace"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestNaisNamespaceReconciler(t *testing.T) {
	ctx := context.Background()

	ch := make(chan *dbmodels.AuditLog, 100)
	logger := auditlogger.New(ch)
	defer close(ch)

	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	db, err := setupDatabase()
	if err != nil {
		panic(err)
	}

	rec, err := nais_namespace_reconciler.NewFromConfig(db, cfg, logger)
	if err != nil {
		panic(err)
	}

	sysid := uuid.New()

	teamName := dbmodels.Slug("foo")
	err = rec.Reconcile(ctx, reconcilers.Input{
		System: &dbmodels.System{
			Model: dbmodels.Model{
				ID: &sysid,
			},
		},
		Team: &dbmodels.Team{
			Slug: &teamName,
			SystemState: []*dbmodels.SystemState{
				{
					SystemID:    &sysid,
					Environment: helpers.Strp("dev"),
					Key:         "google-project-id",
					Value:       "this-is-the-google-project-id",
				},
			},
		},
	})

	assert.NoError(t, err)
}

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
