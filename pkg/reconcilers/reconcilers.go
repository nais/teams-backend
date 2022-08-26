package reconcilers

import (
	"context"
	"errors"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/sqlc"
)

// ErrReconcilerNotEnabled Custom error to use when a reconciler is not enabled via configuration
var ErrReconcilerNotEnabled = errors.New("reconciler not enabled")

// Reconciler Interface for all reconcilers
type Reconciler interface {
	Reconcile(ctx context.Context, input Input) error
	Name() sqlc.SystemName
}

// TeamNamePrefix Prefix that can be used for team-like objects in external systems
const TeamNamePrefix = "nais-team-"

// ReconcilerFactory The constructor function for all reconcilers
type ReconcilerFactory func(context.Context, db.Database, *config.Config, auditlogger.AuditLogger) (Reconciler, error)
