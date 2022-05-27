package console_reconciler

import (
	"context"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/reconcilers/registry"
	"gorm.io/gorm"
)

type consoleReconciler struct{}

const (
	Name = "console"
)

func init() {
	registry.Register(Name, NewFromConfig)
}

func New() *consoleReconciler {
	return &consoleReconciler{}
}

func NewFromConfig(_ *gorm.DB, _ *config.Config, _ auditlogger.Logger) (reconcilers.Reconciler, error) {
	return New(), nil
}

func (s *consoleReconciler) Reconcile(ctx context.Context, in reconcilers.Input) error {
	return nil
}
