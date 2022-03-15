package auditlogger

import (
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/reconcilers"
)

type logger struct {
	logs chan<- *dbmodels.AuditLog
}

type Logger interface {
	Errorf(in reconcilers.Input, operation string, format string, args ...interface{}) error
	UserErrorf(in reconcilers.Input, operation string, user *dbmodels.User, format string, args ...interface{}) error
	Logf(in reconcilers.Input, operation string, format string, args ...interface{})
	UserLogf(in reconcilers.Input, operation string, user *dbmodels.User, format string, args ...interface{})
}

func New(logs chan<- *dbmodels.AuditLog) Logger {
	return &logger{
		logs: logs,
	}
}

func (s *logger) Errorf(in reconcilers.Input, operation string, format string, args ...interface{}) error {
	return in.AuditLog(nil, false, operation, format, args...)
}

func (s *logger) UserErrorf(in reconcilers.Input, operation string, user *dbmodels.User, format string, args ...interface{}) error {
	return in.AuditLog(user, false, operation, format, args...)
}

func (s *logger) Logf(in reconcilers.Input, operation string, format string, args ...interface{}) {
	s.UserLogf(in, operation, nil, format, args...)
}

func (s *logger) UserLogf(in reconcilers.Input, operation string, user *dbmodels.User, format string, args ...interface{}) {
	logLine := in.AuditLog(user, true, operation, format, args...)
	s.logs <- logLine
}
