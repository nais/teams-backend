package reconcilers

import (
	"github.com/google/uuid"
	"github.com/nais/console/pkg/db"
)

// Input Input for reconcilers
type Input struct {
	CorrelationID uuid.UUID
	Team          *db.Team
}
