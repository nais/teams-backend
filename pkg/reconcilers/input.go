package reconcilers

import (
	"context"
	"fmt"

	"github.com/nais/console/pkg/dbmodels"
)

// All synchronizers must implement the Reconciler interface.
type Reconciler interface {
	Reconcile(ctx context.Context, s Input) error
}

// Input data for all synchronizers.
//
// Team is a fully populated database object and contains child references to
// users, user roles, team roles, and any other metadata.
type Input struct {
	System          *dbmodels.System
	Synchronization *dbmodels.Synchronization
	Team            *dbmodels.Team
}

var ErrReconcilerNotEnabled = fmt.Errorf("reconciler not enabled")

const TeamNamePrefix = "nais-team-"

// Helper method to quickly create an audit log line based on the current synchronization.
// FIXME: improve API
func (s *Input) AuditLog(user *dbmodels.User, success bool, action, format string, args ...interface{}) *dbmodels.AuditLog {
	return &dbmodels.AuditLog{
		Action:          action,
		Message:         fmt.Sprintf(format, args...),
		Success:         success,
		Synchronization: s.Synchronization,
		System:          s.System,
		Team:            s.Team,
		User:            user,
	}
}
