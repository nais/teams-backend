package console_reconciler

import (
	"context"

	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/reconcilers"
)

type consoleSynchronizer struct {
	logs chan<- *dbmodels.AuditLog
}

func New(logs chan<- *dbmodels.AuditLog) *consoleSynchronizer {
	return &consoleSynchronizer{
		logs: logs,
	}
}

func (s *consoleSynchronizer) Name() string {
	return "console"
}

func (s *consoleSynchronizer) Reconcile(ctx context.Context, in reconcilers.Input) error {
	return nil
}
