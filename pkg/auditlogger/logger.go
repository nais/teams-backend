package auditlogger

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nais/teams-backend/pkg/authz"
	"github.com/nais/teams-backend/pkg/db"
	"github.com/nais/teams-backend/pkg/helpers"
	"github.com/nais/teams-backend/pkg/logger"
	"github.com/nais/teams-backend/pkg/slug"
	"github.com/nais/teams-backend/pkg/sqlc"
	"github.com/sirupsen/logrus"
)

type AuditLogger interface {
	Logf(ctx context.Context, dbtx db.Database, targets []Target, entry Fields, message string, messageArgs ...interface{})
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

func (l *auditLogger) Logf(ctx context.Context, dbtx db.Database, targets []Target, fields Fields, message string, messageArgs ...interface{}) {
	if l.systemName == "" || !l.systemName.Valid() {
		l.log.Errorf("unable to create auditlog entry: missing or invalid system name")
		return
	}

	if fields.Action == "" || !fields.Action.Valid() {
		l.log.Errorf("unable to create auditlog entry: missing or invalid audit action")
		return
	}

	if fields.CorrelationID == uuid.Nil {
		id, err := uuid.NewUUID()
		if err != nil {
			l.log.WithError(err).Errorf("missing correlation ID in fields and unable to generate one")
			return
		}
		fields.CorrelationID = id
	}

	message = fmt.Sprintf(message, messageArgs...)

	var actor *string
	if fields.Actor != nil {
		actor = helpers.Strp(fields.Actor.User.Identity())
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
			l.log.WithError(err).Errorf("create audit log entry")
			return
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
