package teamsync

import (
	"github.com/google/uuid"
	"github.com/nais/console/pkg/slug"
)

type Input struct {
	CorrelationID uuid.UUID
	TeamSlug      slug.Slug
}
