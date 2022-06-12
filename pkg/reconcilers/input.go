package reconcilers

import (
	"fmt"
	"github.com/google/uuid"

	"github.com/nais/console/pkg/dbmodels"
)

// Input data for all reconcilers
type Input struct {
	System          dbmodels.System
	Synchronization dbmodels.Synchronization
	Team            *dbmodels.Team
}

func (in Input) GetAuditLogEntry(user *dbmodels.User, success bool, action, format string, args ...interface{}) *dbmodels.AuditLog {
	var teamId *uuid.UUID
	var userId *uuid.UUID

	if in.Team != nil && in.Team.ID != nil {
		teamId = in.Team.ID
	}

	if user != nil && user.ID != nil {
		userId = user.ID
	}

	return &dbmodels.AuditLog{
		Action:            action,
		Message:           fmt.Sprintf(format, args...),
		Success:           success,
		SynchronizationID: *in.Synchronization.ID,
		SystemID:          *in.System.ID,
		TeamID:            teamId,
		UserID:            userId,
	}
}
