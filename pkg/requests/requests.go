package requests

import (
	"github.com/nais/console/pkg/models"
)

type GenericRequest struct {
	ID string `json:"id" path:"id" format:"uuid"`
}

type AuthenticatedRequest struct {
	Authorization string `header:"authorization"`
}

// This class is used to read both "id" from the URL and the team object from JSON.
// Could have used models.Team, if not for the fact that "ID" (from URL)
// must be read in through GenericRequest, since having the "path" annotation
// in models.Team generates bad documentation.
type TeamIDRequest struct {
	GenericRequest
	AuthenticatedRequest
	models.Team
}

type UserRequest struct {
	AuthenticatedRequest
	models.User
}

type UserIDRequest struct {
	GenericRequest
	AuthenticatedRequest
	models.User
}

type DeleteApiKeyRequest struct {
	AuthenticatedRequest
	UserID string `json:"user_id" path:"user_id" format:"uuid"`
}
