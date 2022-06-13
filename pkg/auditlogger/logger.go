package auditlogger

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/dbmodels"
)

type auditLogger struct {
	logs chan<- *dbmodels.AuditLog
}

type AuditLog struct {
	Success bool   `gorm:"not null; index"` // True if operation succeeded
	Action  string `gorm:"not null; index"`
	Message string `gorm:"not null"` // Human readable success or error message (log line)
}

type AuditLogger interface {
	Log(action string, success bool, sync dbmodels.Synchronization, targetSystem dbmodels.System, actor *dbmodels.User, targetTeam *dbmodels.Team, targetUser *dbmodels.User, message string, messageArgs ...interface{})
}

func New(logs chan<- *dbmodels.AuditLog) AuditLogger {
	return &auditLogger{
		logs: logs,
	}
}

func (s *auditLogger) Log(action string, success bool, sync dbmodels.Synchronization, targetSystem dbmodels.System, actor *dbmodels.User, targetTeam *dbmodels.Team, targetUser *dbmodels.User, message string, messageArgs ...interface{}) {
	var actorId *uuid.UUID
	var targetTeamId *uuid.UUID
	var targetUserId *uuid.UUID

	if actor != nil && actor.ID != nil {
		actorId = actor.ID
	}

	if targetTeam != nil && targetTeam.ID != nil {
		targetTeamId = targetTeam.ID
	}

	if targetUser != nil && targetUser.ID != nil {
		targetUserId = targetUser.ID
	}

	logEntry := &dbmodels.AuditLog{
		Action:            action,
		Actor:             actor,
		ActorID:           actorId,
		Synchronization:   sync,
		SynchronizationID: *sync.ID,
		TargetSystem:      targetSystem,
		TargetSystemID:    *targetSystem.ID,
		TargetTeam:        targetTeam,
		TargetUser:        targetUser,
		TargetTeamID:      targetTeamId,
		TargetUserID:      targetUserId,
		Success:           success,

		Message: fmt.Sprintf(message, messageArgs...),
	}
	s.logs <- logEntry
}
