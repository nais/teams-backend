package auditlogger

import (
	"context"
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
	Logf(ctx context.Context, action sqlc.AuditAction, correlationID uuid.UUID, systemName sqlc.SystemName, actorEmail *string, targetTeamSlug *string, targetUserEmail *string, message string, messageArgs ...interface{}) error
}

func New(database db.Database) AuditLogger {
	return &auditLogger{
		database: database,
	}
}

func (l *auditLogger) Logf(ctx context.Context, action sqlc.AuditAction, correlationID uuid.UUID, systemName sqlc.SystemName, actorEmail *string, targetTeamSlug *string, targetUserEmail *string, message string, messageArgs ...interface{}) error {
	message = fmt.Sprintf(message, messageArgs...)
	err := l.database.AddAuditLog(ctx, correlationID, systemName, actorEmail, targetTeamSlug, targetUserEmail, action, message)
	if err != nil {
		return fmt.Errorf("create audit log entry: %w", err)
	}

	log.StandardLogger().WithFields(log.Fields{
		"action":            action,
		"correlation_id":    correlationID,
		"system_name":       systemName,
		"actor_email":       str(actorEmail),
		"target_team_slug":  str(targetTeamSlug),
		"target_user_email": str(targetUserEmail),
	}).Infof(message)
	return nil
}

func str(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
