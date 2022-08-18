package auditlogger

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/sqlc"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type auditLogger struct {
	db *gorm.DB
}

type AuditLogger interface {
	Logf(action string, corr sqlc.Correlation, targetSystem sqlc.System, actor *dbmodels.User, targetTeam *dbmodels.Team, targetUser *dbmodels.User, message string, messageArgs ...interface{}) error
}

func New(db *gorm.DB) AuditLogger {
	return &auditLogger{
		db: db,
	}
}

func (l *auditLogger) Logf(action string, corr sqlc.Correlation, targetSystem sqlc.System, actor *dbmodels.User, targetTeam *dbmodels.Team, targetUser *dbmodels.User, message string, messageArgs ...interface{}) error {
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

	// tmp fix to use dbmodels.System instead of sqlc.System for the audit log
	system := &dbmodels.System{}
	err := l.db.Where("id = ?", targetSystem.ID).First(system).Error
	if err != nil {
		return fmt.Errorf("unable to fetch system entry from DB: %w", err)
	}

	// tmp fix to use dbmodels.Correlation instead of sqlc.Correlation for the audit log
	correlation := &dbmodels.Correlation{}
	err = l.db.Where("id = ?", corr.ID).First(correlation).Error
	if err != nil {
		return fmt.Errorf("unable to fetch correlation entry from DB: %w", err)
	}

	logEntry := &dbmodels.AuditLog{
		Action:         action,
		Actor:          actor,
		ActorID:        actorId,
		Correlation:    *correlation,
		CorrelationID:  *correlation.ID,
		TargetSystem:   *system,
		TargetSystemID: *system.ID,
		TargetTeam:     targetTeam,
		TargetUser:     targetUser,
		TargetTeamID:   targetTeamId,
		TargetUserID:   targetUserId,

		Message: fmt.Sprintf(message, messageArgs...),
	}
	err = l.db.Omit(clause.Associations).Create(logEntry).Error
	if err != nil {
		return fmt.Errorf("store audit log line in database: %s", err)
	}

	logEntry.Log().Infof(logEntry.Message)
	return err
}
