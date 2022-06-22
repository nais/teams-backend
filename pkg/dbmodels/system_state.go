package dbmodels

import (
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UpdateStateFunc func(newState interface{}) error

// LoadSystemState Fetch the team state for a given system, and load the current state into the passed state parameter.
// If the team has no state in the system yet, a blank one will be createdThe UpdateStateFunc return function can be
// used to update the state if needed.
func LoadSystemState(db *gorm.DB, systemId, teamId uuid.UUID, state interface{}) (UpdateStateFunc, error) {
	systemState := &SystemState{
		SystemID: systemId,
		TeamID:   teamId,
	}

	err := db.Where("system_id = ? AND team_id = ?", systemId, teamId).FirstOrCreate(systemState).Error
	if err != nil {
		return nil, fmt.Errorf("get system state: %w", err)
	}

	err = systemState.State.AssignTo(&state)
	if err != nil {
		return nil, fmt.Errorf("unable to assign state: %w", err)
	}

	return func(newState interface{}) error {
		state = newState
		err = systemState.State.Set(newState)
		if err != nil {
			return fmt.Errorf("system state not set: %w", err)
		}

		err = db.Save(systemState).Error
		if err != nil {
			return fmt.Errorf("system state not persisted: %w", err)
		}

		return nil
	}, nil
}
