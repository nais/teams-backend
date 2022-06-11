package reconcilers

import (
	"fmt"

	"github.com/nais/console/pkg/dbmodels"
)

// Input data for all reconcilers
type Input struct {
	System          *dbmodels.System
	Synchronization *dbmodels.Synchronization
	Team            *dbmodels.Team
}

func (in Input) GetAuditLogEntry(user *dbmodels.User, success bool, action, format string, args ...interface{}) *dbmodels.AuditLog {
	return &dbmodels.AuditLog{
		Action:          action,
		Message:         fmt.Sprintf(format, args...),
		Success:         success,
		Synchronization: in.Synchronization,
		System:          in.System,
		Team:            in.Team,
		User:            user,
	}
}
