package db

import (
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate/v4"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/nais/console/sqlc/schemas"
)

func NewDatabase(q Querier) Database {
	return &database{querier: q}
}

func NullStringToStringP(ns sql.NullString) *string {
	var strP *string
	if ns.String != "" {
		strP = &ns.String
	}
	return strP
}

func Migrate(connString string) error {
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

	return nil
}

func nullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{
		String: *s,
		Valid:  true,
	}
}

func nullUUID(ID *uuid.UUID) uuid.NullUUID {
	if ID == nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{
		UUID:  *ID,
		Valid: true,
	}
}
