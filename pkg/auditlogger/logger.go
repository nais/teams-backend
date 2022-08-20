package auditlogger

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/nais/console/pkg/db"
	log "github.com/sirupsen/logrus"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/sqlc"
)

type auditLogger struct {
	database db.Database
}

type AuditLogger interface {
	Logf(ctx context.Context, action sqlc.AuditAction, correlationId uuid.UUID, systemName *sqlc.SystemName, actorEmail *string, targetTeamSlug *string, targetUserEmail *string, message string, messageArgs ...interface{}) error
}

func New(database db.Database) AuditLogger {
	return &auditLogger{
		database: database,
	}
}

func (l *auditLogger) Logf(ctx context.Context, action sqlc.AuditAction, correlationId uuid.UUID, systemName *sqlc.SystemName, actorEmail *string, targetTeamSlug *string, targetUserEmail *string, message string, messageArgs ...interface{}) error {
	nullSystemName := sqlc.NullSystemName{}
	if systemName != nil {
		nullSystemName.SystemName = *systemName
	}

	logEntry := &db.AuditLog{}
	logEntry.Action = action
	logEntry.CorrelationID = correlationId
	logEntry.SystemName = nullSystemName
	logEntry.ActorEmail = nullString(actorEmail)
	logEntry.TargetTeamSlug = nullString(targetTeamSlug)
	logEntry.TargetUserEmail = nullString(targetUserEmail)
	logEntry.Message = fmt.Sprintf(message, messageArgs...)
	logEntry, err := l.database.AddAuditLog(ctx, *logEntry)
	if err != nil {
		return fmt.Errorf("create audit log entry: %w", err)
	}

	log.StandardLogger().WithFields(log.Fields{
		"action":            action,
		"correlation_id":    correlationId,
		"system_name":       systemName,
		"actor_email":       actorEmail,
		"target_team_slug":  targetTeamSlug,
		"target_user_email": targetUserEmail,
	})
	return nil
}

func nullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}

	return sql.NullString{
		String: *s,
		Valid:  true,
	}
}
