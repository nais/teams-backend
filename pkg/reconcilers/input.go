package reconcilers

import (
	"context"
	"fmt"

	"github.com/nais/console/pkg/dbmodels"
	log "github.com/sirupsen/logrus"
)

// All synchronizers must implement the Reconciler interface.
type Reconciler interface {
	Name() string
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

// Helper method to quickly create an audit log line based on the current synchronization.
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

// Helper method to create a log entry with synchronization and system metadata.
func (s *Input) Logger() *log.Entry {
	return log.StandardLogger().WithFields(log.Fields{
		"correlation_id": s.Synchronization.ID.String(),
		"system":         s.System.Name,
		"team":           *s.Team.Slug,
	})
}
