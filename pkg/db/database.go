package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nais/teams-backend/pkg/logger"
	"github.com/nais/teams-backend/pkg/metrics"
	"github.com/nais/teams-backend/pkg/sqlc"
	"github.com/nais/teams-backend/sqlc/schemas"
)

const databaseConnectRetries = 5

func New(ctx context.Context, dbUrl string, log logger.Logger) (Database, error) {
	config, err := pgxpool.ParseConfig(dbUrl)
	if err != nil {
		return nil, err
	}

	var dbc *pgxpool.Pool
	for i := 0; i < databaseConnectRetries; i++ {
		dbc, err = pgxpool.ConnectConfig(ctx, config)
		if err == nil {
			break
		}

		log.Warnf("unable to connect to the database: %s", err)
		time.Sleep(time.Second * time.Duration(i+1))
	}

	if dbc == nil {
		return nil, fmt.Errorf("giving up connecting to the database after %d attempts: %w", databaseConnectRetries, err)
	}

	err = runMigrations(dbc.Config().ConnString())
	if err != nil {
		return nil, err
	}

	return &database{
		querier: &Queries{
			Queries:  sqlc.New(dbc),
			connPool: dbc,
		},
	}, nil
}

func runMigrations(connString string) error {
	d, err := iofs.New(schemas.FS, ".")
	if err != nil {
		return err
	}
	defer d.Close()

	m, err := migrate.NewWithSourceInstance("iofs", d, connString)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	version, dirty, err := m.Version()
	if err != nil {
		return err
	}

	metrics.SetSchemaVersion(version, dirty)

	return nil
}
