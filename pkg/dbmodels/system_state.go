package dbmodels

import (
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LoadSystemState Load the team state for a given system into the state parameter
func LoadSystemState(db *gorm.DB, systemId, teamId uuid.UUID, state interface{}) error {
	systemState := &SystemState{
		SystemID: systemId,
		TeamID:   teamId,
	}

	err := db.Where("system_id = ? AND team_id = ?", systemId, teamId).FirstOrCreate(systemState).Error
	if err != nil {
		return fmt.Errorf("get system state: %w", err)
	}

	err = systemState.State.AssignTo(state)
	if err != nil {
		return fmt.Errorf("unable to assign state: %w", err)
	}

	return nil
}

// UpdateSystemState Update the team state for a given system
func UpdateSystemState(db *gorm.DB, systemId, teamId uuid.UUID, state interface{}) error {
	systemState := &SystemState{
		SystemID: systemId,
		TeamID:   teamId,
	}

	err := db.Where("system_id = ? AND team_id = ?", systemId, teamId).FirstOrCreate(systemState).Error
	if err != nil {
		return fmt.Errorf("get system state: %w", err)
	}

	err = systemState.State.Set(state)
	if err != nil {
		return fmt.Errorf("system state not set: %w", err)
	}

	err = db.Save(systemState).Error
	if err != nil {
		return fmt.Errorf("system state not persisted: %w", err)
	}

	return nil
}
