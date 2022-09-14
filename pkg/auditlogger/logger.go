package auditlogger

import (
	"context"
	"fmt"

	"github.com/nais/console/pkg/slug"

	"github.com/nais/console/pkg/db"
	log "github.com/sirupsen/logrus"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/sqlc"
)

type auditLogger struct {
	database   db.Database
	systemName sqlc.SystemName
}

func (l *auditLogger) WithSystemName(systemName sqlc.SystemName) AuditLogger {
	return &auditLogger{
		database:   l.database,
		systemName: systemName,
	}
}

type AuditLogger interface {
	Logf(ctx context.Context, entry Fields, message string, messageArgs ...interface{}) error
	WithSystemName(systemName sqlc.SystemName) AuditLogger
}

func New(database db.Database) AuditLogger {
	return &auditLogger{
		database: database,
	}
}

type Fields struct {
	Action         sqlc.AuditAction
	CorrelationID  uuid.UUID
	Actor          *string
	TargetTeamSlug *slug.Slug
	TargetUser     *string
}

func (l *auditLogger) Logf(ctx context.Context, fields Fields, message string, messageArgs ...interface{}) error {
	if l.systemName == "" || !l.systemName.Valid() {
		return fmt.Errorf("unable to create auditlog entry: missing or invalid systemName")
	}

	if fields.Action == "" || !fields.Action.Valid() {
		return fmt.Errorf("unable to create auditlog entry: missing or invalid audit action")
	}

	if fields.CorrelationID == uuid.Nil {
		id, err := uuid.NewUUID()
		if err != nil {
			return fmt.Errorf("missing correlation ID in fields and unable to generate one: %w", err)
		}
		fields.CorrelationID = id
	}

	message = fmt.Sprintf(message, messageArgs...)
	err := l.database.CreateAuditLogEntry(
		ctx,
		fields.CorrelationID,
		l.systemName,
		fields.Actor,
		fields.TargetTeamSlug,
		fields.TargetUser,
		fields.Action,
		message,
	)
	if err != nil {
		return fmt.Errorf("create audit log entry: %w", err)
	}

	logFields := log.Fields{
		"action":         fields.Action,
		"correlation_id": fields.CorrelationID,
		"system_name":    l.systemName,
	}
	if fields.Actor != nil {
		logFields["actor"] = str(fields.Actor)
	}
	if fields.TargetTeamSlug != nil {
		logFields["target_team_slug"] = str(fields.TargetTeamSlug.StringP())
	}
	if fields.TargetUser != nil {
		logFields["target_user"] = str(fields.TargetUser)
	}

	log.StandardLogger().WithFields(logFields).Infof(message)

	return nil
}

func str(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
