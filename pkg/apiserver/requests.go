package apiserver

import (
	"github.com/nais/console/pkg/models"
)

type GenericRequest struct {
	ID string `json:"id" path:"id" format:"uuid"`
}

// Could have used models.Team, if not for the fact that "ID" (from URL)
// must be read in through GenericRequest.
type TeamRequest struct {
	GenericRequest
	models.Team
}
