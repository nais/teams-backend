package reconcilers

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/nais/teams-backend/pkg/auditlogger"
	"github.com/nais/teams-backend/pkg/config"
	"github.com/nais/teams-backend/pkg/db"
	"github.com/nais/teams-backend/pkg/logger"
	"github.com/nais/teams-backend/pkg/slug"
	"github.com/nais/teams-backend/pkg/sqlc"
)

const (
	ManagedByLabelName  = "managed-by"
	ManagedByLabelValue = "teams-backend"

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
