package auditlogger

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/console"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
	log "github.com/sirupsen/logrus"
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
	Logf(ctx context.Context, targets []Target, entry Fields, message string, messageArgs ...interface{}) error
	WithSystemName(systemName sqlc.SystemName) AuditLogger
}

func New(database db.Database) AuditLogger {
	return &auditLogger{
		database: database,
	}
}

type Target struct {
	Type       sqlc.AuditLogsTargetType
	Identifier string
}

func UserTarget(email string) Target {
	return Target{Type: sqlc.AuditLogsTargetTypeUser, Identifier: email}
}

func TeamTarget(slug slug.Slug) Target {
	return Target{Type: sqlc.AuditLogsTargetTypeTeam, Identifier: string(slug)}
}

func ReconcilerTarget(name sqlc.ReconcilerName) Target {
	return Target{Type: sqlc.AuditLogsTargetTypeReconciler, Identifier: string(name)}
}

type Fields struct {
	Action        sqlc.AuditAction
	Actor         *authz.Actor
	CorrelationID uuid.UUID
}

func (l *auditLogger) Logf(ctx context.Context, targets []Target, fields Fields, message string, messageArgs ...interface{}) error {
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

	for _, target := range targets {
		var actor *string
		if fields.Actor != nil {
			actor = console.Strp(fields.Actor.User.Identity())
		}
		err := l.database.CreateAuditLogEntry(
			ctx,
			fields.CorrelationID,
			l.systemName,
			actor,
			target.Type,
			target.Identifier,
			fields.Action,
			message,
		)
		if err != nil {
			return fmt.Errorf("create audit log entry: %w", err)
		}

		logFields := log.Fields{
			"action":            fields.Action,
			"correlation_id":    fields.CorrelationID,
			"system_name":       l.systemName,
			"target_type":       target.Type,
			"target_identifier": target.Identifier,
		}
		if actor != nil {
			logFields["actor"] = *actor
		}

		log.StandardLogger().WithFields(logFields).Infof(message)
	}

	return nil
}
