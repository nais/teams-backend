package reconcilers

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
)

const (
	ManagedByLabelName  = "managed-by"
	ManagedByLabelValue = "console"

	// TeamNamePrefix Prefix that can be used for team-like objects in external systems
	TeamNamePrefix              = "nais-team-"
	CnrmServiceAccountAccountID = "nais-sa-cnrm"
)

// Reconciler Interface for all reconcilers
type Reconciler interface {
	Reconcile(ctx context.Context, input Input) error
	Delete(ctx context.Context, teamSlug slug.Slug, correlationID uuid.UUID) error
	Name() sqlc.ReconcilerName
}

// ErrReconcilerNotEnabled Custom error to use when a reconciler is not enabled via configuration
var ErrReconcilerNotEnabled = errors.New("reconciler not enabled")

// ReconcilerFactory The constructor function for all reconcilers
type ReconcilerFactory func(context.Context, db.Database, *config.Config, auditlogger.AuditLogger, logger.Logger) (Reconciler, error)
