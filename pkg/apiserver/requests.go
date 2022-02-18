package apiserver

import (
	"github.com/nais/console/pkg/models"
)

type GenericRequest struct {
	ID string `json:"id" path:"id" format:"uuid"`
}

// This class is used to read both "id" from the URL and the team object from JSON.
// Could have used models.Team, if not for the fact that "ID" (from URL)
// must be read in through GenericRequest, since having the "path" annotation
// in models.Team generates bad documentation.
type TeamRequest struct {
	GenericRequest
	models.Team
}
