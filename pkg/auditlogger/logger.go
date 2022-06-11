package auditlogger

import (
	"github.com/nais/console/pkg/dbmodels"
)

type logger struct {
	logs chan<- *dbmodels.AuditLog
}

// Input Log input interface
type Input interface {
	GetAuditLogEntry(user *dbmodels.User, success bool, action, format string, args ...interface{}) *dbmodels.AuditLog
}

type Logger interface {
	Errorf(in Input, operation string, format string, args ...interface{}) error
	UserErrorf(in Input, operation string, user *dbmodels.User, format string, args ...interface{}) error
	Logf(in Input, operation string, format string, args ...interface{})
	UserLogf(in Input, operation string, user *dbmodels.User, format string, args ...interface{})
}

func New(logs chan<- *dbmodels.AuditLog) Logger {
	return &logger{
		logs: logs,
	}
}

func (s *logger) Errorf(in Input, operation string, format string, args ...interface{}) error {
	return in.GetAuditLogEntry(nil, false, operation, format, args...)
}

func (s *logger) UserErrorf(in Input, operation string, user *dbmodels.User, format string, args ...interface{}) error {
	return in.GetAuditLogEntry(user, false, operation, format, args...)
}

func (s *logger) Logf(in Input, operation string, format string, args ...interface{}) {
	s.UserLogf(in, operation, nil, format, args...)
}

func (s *logger) UserLogf(in Input, operation string, user *dbmodels.User, format string, args ...interface{}) {
	s.logs <- in.GetAuditLogEntry(user, true, operation, format, args...)
}
