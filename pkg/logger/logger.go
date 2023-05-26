package logger

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nais/teams-backend/pkg/sqlc"
	"github.com/sirupsen/logrus"
)

type Logger interface {
	logrus.FieldLogger
	GetInternalLogger() *logrus.Logger
	WithActor(actor string) Logger
	WithCorrelationID(correlationID uuid.UUID) Logger
	WithReconciler(reconciler string) Logger
	WithSystem(system string) Logger
	WithTeamSlug(slug string) Logger
	WithUser(user string) Logger
}

type logger struct {
	*logrus.Entry
}

func (l *logger) GetInternalLogger() *logrus.Logger {
	return l.Entry.Logger
}

func (l *logger) WithTeamSlug(slug string) Logger {
	return &logger{l.WithField("team", slug)}
}

func (l *logger) WithActor(actor string) Logger {
	return &logger{l.WithField("actor", actor)}
}

func (l *logger) WithCorrelationID(correlationID uuid.UUID) Logger {
	return &logger{l.WithField("correlationID", correlationID.String())}
}

func (l *logger) WithUser(user string) Logger {
	return &logger{l.WithField("user", user)}
}

func (l *logger) WithReconciler(reconciler string) Logger {
	return &logger{l.WithField("reconciler", reconciler)}
}

func (l *logger) WithSystem(system string) Logger {
	return &logger{l.WithField("system", system)}
}

func GetLogger(format, level string) (Logger, error) {
	log := logrus.StandardLogger()

	switch format {
	case "json":
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
		})
	case "text":
		log.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: time.RFC3339Nano,
		})
	default:
		return &logger{}, fmt.Errorf("invalid log format: %s", format)
	}

	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return &logger{}, err
	}

	log.SetLevel(lvl)

	return &logger{log.WithField("system", sqlc.SystemNameConsole)}, nil
}
