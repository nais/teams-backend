package roles

type Role string

const (
	RoleAdmin                 Role = "Admin"
	RoleServiceAccountCreator Role = "Service account creator"
	RoleServiceAccountOwner   Role = "Service account owner"
	RoleTeamCreator           Role = "Team creator"
	RoleTeamMember            Role = "Team member"
	RoleTeamOwner             Role = "Team owner"
	RoleTeamViewer            Role = "Team viewer"
)
