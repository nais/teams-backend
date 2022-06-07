package dbmodels

import (
	"github.com/google/uuid"
)

var (
	AdminUserID = &uuid.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xa}

	TeamEditorRoleID  = &uuid.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xa, 0}
	TeamCreatorRoleID = &uuid.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xb, 0}
	TeamViewerRoleID  = &uuid.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xc, 0}
	RoleEditorRoleID  = &uuid.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xd, 0}
)
