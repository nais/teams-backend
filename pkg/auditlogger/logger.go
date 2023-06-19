package auditlogger

import (
	"context"
	"fmt"

	"github.com/nais/teams-backend/pkg/types"

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
	Logf(ctx context.Context, targets []Target, fields Fields, message string, messageArgs ...interface{})
}

type auditLogger struct {
	componentName types.ComponentName
	db            db.Database
	log           logger.Logger
}

type auditLoggerForTesting struct {
	entries []Entry
}

type Target struct {
	Type       types.AuditLogsTargetType
	Identifier string
}

type Fields struct {
	Action        types.AuditAction
	Actor         *authz.Actor
	CorrelationID uuid.UUID
}

type Entry struct {
	Context context.Context
	Targets []Target
	Fields  Fields
	Message string
}

func New(db db.Database, componentName types.ComponentName, log logger.Logger) AuditLogger {
	return &auditLogger{
		componentName: componentName,
		db:            db,
		log:           log.WithComponent(componentName),
	}
}

func NewAuditLoggerForTesting() *auditLoggerForTesting {
	return &auditLoggerForTesting{
		entries: make([]Entry, 0),
	}
}

func (l *auditLoggerForTesting) Logf(ctx context.Context, targets []Target, fields Fields, message string, messageArgs ...interface{}) {
	l.entries = append(l.entries, Entry{
		Context: ctx,
		Targets: targets,
		Fields:  fields,
		Message: fmt.Sprintf(message, messageArgs...),
	})
}

func (l *auditLoggerForTesting) Entries() []Entry {
	return l.entries
}

// Logf Write the audit log entry to the database, and generate a system log entry. Do not call this function inside of
// a database transaction as it will generate a system log entry.
func (l *auditLogger) Logf(ctx context.Context, targets []Target, fields Fields, message string, messageArgs ...interface{}) {
	if fields.Action == "" {
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
		err := l.db.CreateAuditLogEntry(
			ctx,
			fields.CorrelationID,
			l.componentName,
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
		case types.AuditLogsTargetTypeTeam:
			log = log.WithTeamSlug(target.Identifier)
		case types.AuditLogsTargetTypeUser:
			log = log.WithUser(target.Identifier)
		case types.AuditLogsTargetTypeReconciler:
			log = log.WithReconciler(target.Identifier)
		default:
			logFields["target_identifier"] = target.Identifier
		}

		log.WithFields(logFields).Infof(message)
	}
}

func UserTarget(email string) Target {
	return Target{Type: types.AuditLogsTargetTypeUser, Identifier: email}
}

func TeamTarget(slug slug.Slug) Target {
	return Target{Type: types.AuditLogsTargetTypeTeam, Identifier: string(slug)}
}

func ReconcilerTarget(name sqlc.ReconcilerName) Target {
	return Target{Type: types.AuditLogsTargetTypeReconciler, Identifier: string(name)}
}

func ComponentTarget(name types.ComponentName) Target {
	return Target{Type: types.AuditLogsTargetTypeSystem, Identifier: string(name)}
}
