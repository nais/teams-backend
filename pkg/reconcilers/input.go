package reconcilers

import (
	"github.com/google/uuid"
	"github.com/nais/console/pkg/db"
)

// Input Input for reconcilers
type Input struct {
	CorrelationId uuid.UUID
	Team          *db.Team
}
