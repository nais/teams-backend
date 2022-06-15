package console_reconciler

import (
	"context"
	"github.com/nais/console/pkg/dbmodels"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/reconcilers/registry"
	"gorm.io/gorm"
)

type consoleReconciler struct{}

const (
	Name             = "console"
	OpReconcileStart = "console:reconcile:start"
	OpReconcileEnd   = "console:reconcile:end"
	OpCreateTeam     = "console:team:create"
	OpSyncTeam       = "console:team:sync"
)

func init() {
	registry.Register(Name, NewFromConfig)
}

func New() *consoleReconciler {
	return &consoleReconciler{}
}

func NewFromConfig(_ *gorm.DB, _ *config.Config, _ dbmodels.System, _ auditlogger.AuditLogger) (reconcilers.Reconciler, error) {
	return New(), nil
}

func (s *consoleReconciler) Reconcile(_ context.Context, _ reconcilers.Input) error {
	return nil
}
