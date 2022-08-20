package auditlogger

import (
	"fmt"
	"github.com/nais/console/pkg/db"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/sqlc"
	"gorm.io/gorm/clause"
)

type auditLogger struct {
	database db.Database
}

type AuditLogger interface {
	Logf(action string, correlationId uuid.UUID, targetSystemName sqlc.SystemName, actor *db.User, targetTeam *db.Team, targetUser *db.User, message string, messageArgs ...interface{}) error
}

func New(database db.Database) AuditLogger {
	return &auditLogger{
		database: database,
	}
}

func (l *auditLogger) Logf(action string, correlationId uuid.UUID, targetSystemName sqlc.SystemName, actor *db.User, targetTeam *db.Team, targetUser *db.User, message string, messageArgs ...interface{}) error {
	var actorId *uuid.UUID
	var targetTeamId *uuid.UUID
	var targetUserId *uuid.UUID

	if actor != nil {
		actorId = &actor.ID
	}

	if targetTeam != nil {
		targetTeamId = &targetTeam.ID
	}

	if targetUser != nil {
		targetUserId = &targetUser.ID
	}

	// tmp fix to use dbmodels.User instead of sqlc.User for the actor
	actorGorm := &dbmodels.User{}
	err := l.db.Where("id = ?", actor.ID).First(actorGorm).Error
	if err != nil {
		return fmt.Errorf("unable to fetch user from DB: %w", err)
	}

	// tmp fix to use dbmodels.User instead of sqlc.User for the target user
	targetUserGorm := &dbmodels.User{}
	err = l.db.Where("id = ?", targetUser.ID).First(targetUserGorm).Error
	if err != nil {
		return fmt.Errorf("unable to fetch user from DB: %w", err)
	}

	// tmp fix to use dbmodels.System instead of sqlc.System for the audit log
	system := &dbmodels.System{}
	err = l.db.Where("id = ?", targetSystem.ID).First(system).Error
	if err != nil {
		return fmt.Errorf("unable to fetch system entry from DB: %w", err)
	}

	// tmp fix to use dbmodels.Correlation instead of sqlc.Correlation for the audit log
	correlation := &dbmodels.Correlation{}
	err = l.db.Where("id = ?", corr.ID).First(correlation).Error
	if err != nil {
		return fmt.Errorf("unable to fetch correlation entry from DB: %w", err)
	}

	// tmp fix to use dbmodels.Team instead of sqlc.Team for the audit log
	team := &dbmodels.Team{}
	err = l.db.Where("id = ?", targetTeam.ID).First(team).Error
	if err != nil {
		return fmt.Errorf("unable to fetch team entry from DB: %w", err)
	}

	logEntry := &dbmodels.AuditLog{
		Action:         action,
		Actor:          actorGorm,
		ActorID:        actorId,
		Correlation:    *correlation,
		CorrelationID:  *correlation.ID,
		TargetSystem:   *system,
		TargetSystemID: *system.ID,
		TargetTeam:     team,
		TargetUser:     targetUserGorm,
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
