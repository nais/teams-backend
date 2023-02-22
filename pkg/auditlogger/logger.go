package auditlogger

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/console"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
	"github.com/sirupsen/logrus"
)

type AuditLogger interface {
	Logf(ctx context.Context, dbtx db.Database, targets []Target, entry Fields, message string, messageArgs ...interface{}) error
	WithSystemName(systemName sqlc.SystemName) AuditLogger
}

type auditLogger struct {
	systemName sqlc.SystemName
	log        logger.Logger
}

type Target struct {
	Type       sqlc.AuditLogsTargetType
	Identifier string
}

type Fields struct {
	Action        sqlc.AuditAction
	Actor         *authz.Actor
	CorrelationID uuid.UUID
}

func New(log logger.Logger) AuditLogger {
	return &auditLogger{
		log: log,
	}
}

func (l *auditLogger) WithSystemName(systemName sqlc.SystemName) AuditLogger {
	return &auditLogger{
		systemName: systemName,
		log:        l.log.WithSystem(string(systemName)),
	}
}

func (l *auditLogger) Logf(ctx context.Context, dbtx db.Database, targets []Target, fields Fields, message string, messageArgs ...interface{}) error {
	if l.systemName == "" || !l.systemName.Valid() {
		return fmt.Errorf("unable to create auditlog entry: missing or invalid system name")
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

	var actor *string
	if fields.Actor != nil {
		actor = console.Strp(fields.Actor.User.Identity())
	}

	for _, target := range targets {
		err := dbtx.CreateAuditLogEntry(
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

		logFields := logrus.Fields{
			"action":         fields.Action,
			"correlation_id": fields.CorrelationID,
			"target_type":    target.Type,
		}

		log := l.log
		if actor != nil {
			logFields["actor"] = *actor
			log = log.WithActor(*actor)
		}

		switch target.Type {
		case sqlc.AuditLogsTargetTypeTeam:
			log = log.WithTeamSlug(target.Identifier)
		case sqlc.AuditLogsTargetTypeUser:
			log = log.WithUser(target.Identifier)
		case sqlc.AuditLogsTargetTypeReconciler:
			log = log.WithReconciler(target.Identifier)
		default:
			logFields["target_identifier"] = target.Identifier
		}

		log.WithFields(logFields).Infof(message)
	}

	return nil
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

func SystemTarget(name sqlc.SystemName) Target {
	return Target{Type: sqlc.AuditLogsTargetTypeSystem, Identifier: string(name)}
}
