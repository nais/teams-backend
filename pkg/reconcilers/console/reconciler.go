package console_reconciler

import (
	"context"
	"github.com/nais/console/pkg/dbmodels"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/reconcilers"
	"gorm.io/gorm"
)

type consoleReconciler struct {
	system dbmodels.System
}

const (
	Name         = "console"
	OpCreateTeam = "console:team:create"
	OpSyncTeam   = "console:team:sync"
)

func New(system dbmodels.System) *consoleReconciler {
	return &consoleReconciler{
		system: system,
	}
}

func NewFromConfig(_ *gorm.DB, _ *config.Config, system dbmodels.System, _ auditlogger.AuditLogger) (reconcilers.Reconciler, error) {
	return New(system), nil
}

func (r *consoleReconciler) Reconcile(_ context.Context, _ reconcilers.Input) error {
	return nil
}

func (r *consoleReconciler) System() dbmodels.System {
	return r.system
}
