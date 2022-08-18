package reconcilers

import (
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/sqlc"
)

// Input Input for reconcilers
type Input struct {
	Corr sqlc.Correlation
	Team dbmodels.Team
}
