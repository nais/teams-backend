package db_test

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/sqlc"
	"github.com/stretchr/testify/assert"
	"testing"
)

func setupDatabase(ctx context.Context, dbUrl string) (db.Database, error) {
	dbc, err := pgx.Connect(ctx, dbUrl)
	if err != nil {
		return nil, err
	}
	queries := db.Wrap(sqlc.New(dbc), dbc)
	return db.NewDatabase(queries, dbc), nil
}

const testDbUrl = "postgres://console:console@localhost:3002/console?sslmode=disable"

func TestTransaction(t *testing.T) {
	// MUST START WITH EMPTY DATABASE
	ctx := context.Background()
	database, err := setupDatabase(ctx, testDbUrl)
	if err != nil {
		panic("unable to connect to the database for integration tests. Have you remembered to start it using docker-compose?")
	}

	t.Run("transaction return error", func(t *testing.T) {
		email := "mail@example.com"
		_ = database.Transaction(ctx, func(ctx context.Context, dbtx db.Database) error {
			_, err := dbtx.AddUser(ctx, "name", "mail@example.com")
			assert.NoError(t, err)

			return nil
		})
		//assert.Error(t, err)
		user, _ := database.GetUserByEmail(ctx, email)
		//assert.Error(t, err)
		assert.Nil(t, user)
	})
}
