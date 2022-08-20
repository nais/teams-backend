package reconcilers

import (
	"context"
	"errors"
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
