package reconcilers

import (
	"context"
	"errors"
	"github.com/nais/console/pkg/dbmodels"
)

// ErrReconcilerNotEnabled Custom error to use when a reconciler is not enabled via configuration
var ErrReconcilerNotEnabled = errors.New("reconciler not enabled")

// Reconciler Interface for all reconcilers
type Reconciler interface {
	Reconcile(ctx context.Context, corr dbmodels.Correlation, team dbmodels.Team) error
}

// TeamNamePrefix Prefix that can be used for team-like objects in external systems
const TeamNamePrefix = "nais-team-"
