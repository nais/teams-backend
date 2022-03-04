package gcp_team_reconciler

import (
	"context"

	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/reconcilers"
)

type gcpSynchronizer struct {
	logs chan<- *dbmodels.AuditLog
}

func New(logs chan<- *dbmodels.AuditLog) *gcpSynchronizer {
	return &gcpSynchronizer{
		logs: logs,
	}
}

func (s *gcpSynchronizer) Name() string {
	return "gcp:team"
}

// error -> requeue?
func (s *gcpSynchronizer) Reconcile(ctx context.Context, in reconcilers.Input) error {
	s.logs <- in.AuditLog(nil, 200, "api.create", "successfully synchronized")
	in.Logger().Infof("we did it!")
	return nil
}
