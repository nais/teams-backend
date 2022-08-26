package console_reconciler

import (
	"context"

	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/sqlc"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/reconcilers"
)

type consoleReconciler struct{}

const (
	Name = sqlc.SystemNameConsole
)

func New() *consoleReconciler {
	return &consoleReconciler{}
}

func NewFromConfig(_ context.Context, _ db.Database, _ *config.Config, _ auditlogger.AuditLogger) (reconcilers.Reconciler, error) {
	return New(), nil
}

func (r *consoleReconciler) Reconcile(_ context.Context, _ reconcilers.Input) error {
	return nil
}

func (r *consoleReconciler) Name() sqlc.SystemName {
	return Name
}
