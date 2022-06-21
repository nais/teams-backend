package dbmodels

import (
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GetSystemState Get the team state for a given system. If the team has no state in the system yet, a blank one will be
// created, and then returned.
func GetSystemState(db *gorm.DB, systemId, teamId uuid.UUID) (*SystemState, error) {
	systemState := &SystemState{
		SystemID: systemId,
		TeamID:   teamId,
	}

	err := db.Where("system_id = ? AND team_id = ?", systemId, teamId).FirstOrCreate(systemState).Error
	if err != nil {
		return nil, fmt.Errorf("get system state: %w", err)
	}

	return systemState, nil
}
