package roles

import (
	"errors"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/dbmodels"
)

type Authorization string
type Role string

const (
	RoleAdmin                  Role = "Admin"
	RoleServiceAccountCreaetor Role = "Service account creator"
	RoleServiceAccountOwner    Role = "Service account owner"
	RoleTeamCreator            Role = "Team creator"
	RoleTeamMember             Role = "Team member"
	RoleTeamOwner              Role = "Team owner"
	RoleTeamViewer             Role = "Team viewer"
)
