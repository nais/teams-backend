package db

import (
	"github.com/nais/console/pkg/sqlc"
)

func (d *database) GetSystemNames() []sqlc.SystemName {
	return sqlc.AllSystemNameValues()
}
