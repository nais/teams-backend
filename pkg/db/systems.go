package db

import (
	"github.com/nais/console/pkg/sqlc"
)

func (d *database) GetSystemNames() []*sqlc.SystemName {
	values := sqlc.AllSystemNameValues()
	systemNames := make([]*sqlc.SystemName, 0, len(values))
	for _, value := range values {
		systemNames = append(systemNames, &value)
	}
	return systemNames
}
