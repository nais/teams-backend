package reconcilers

import (
	"context"
	"errors"
)

// ErrReconcilerNotEnabled Custom error to use when a reconciler is not enabled via configuration
var ErrReconcilerNotEnabled = errors.New("reconciler not enabled")

// Reconciler Interface for all reconcilers
type Reconciler interface {
	Reconcile(ctx context.Context, input Input) error
}

// TeamNamePrefix Prefix that can be used for team-like objects in external systems
const TeamNamePrefix = "nais-team-"
