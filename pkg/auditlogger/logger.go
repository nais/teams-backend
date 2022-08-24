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

func (a auditLogger) WithSystemName(systemName sqlc.SystemName) AuditLogger {
	a.systemName = systemName
	return &a
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
	Action          sqlc.AuditAction
	CorrelationID   uuid.UUID
	ActorEmail      *string
	TargetTeamSlug  *slug.Slug
	TargetUserEmail *string
}

func (l *auditLogger) Logf(ctx context.Context, fields Fields, message string, messageArgs ...interface{}) error {
	message = fmt.Sprintf(message, messageArgs...)
	err := l.database.AddAuditLog(
		ctx,
		fields.CorrelationID,
		l.systemName,
		fields.ActorEmail,
		fields.TargetTeamSlug,
		fields.TargetUserEmail,
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
	if fields.ActorEmail != nil {
		logFields["actor_email"] = str(fields.ActorEmail)
	}
	if fields.TargetTeamSlug != nil {
		logFields["target_team_slug"] = str(fields.TargetTeamSlug.StringP())
	}
	if fields.TargetUserEmail != nil {
		logFields["target_user_email"] = str(fields.TargetUserEmail)
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
