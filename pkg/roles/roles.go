package roles

type Authorization string
type Role string

const (
	AuthorizationAuditLogsRead         Authorization = "audit_logs.read"
	AuthorizationServiceAccountsCreate Authorization = "service_accounts.create"
	AuthorizationServiceAccountsDelete Authorization = "service_accounts.delete"
	AuthorizationServiceAccountsList   Authorization = "service_accounts.list"
	AuthorizationServiceAccountsRead   Authorization = "service_accounts.read"
	AuthorizationServiceAccountsUpdate Authorization = "service_accounts.update"
	AuthorizationSystemStatesDelete    Authorization = "system_states.delete"
	AuthorizationSystemStatesRead      Authorization = "system_states.read"
	AuthorizationSystemStatesUpdate    Authorization = "system_states.update"
	AuthorizationTeamsCreate           Authorization = "teams.create"
	AuthorizationTeamsDelete           Authorization = "teams.delete"
	AuthorizationTeamsList             Authorization = "teams.list"
	AuthorizationTeamsRead             Authorization = "teams.read"
	AuthorizationTeamsUpdate           Authorization = "teams.update"
	AuthorizationUsersList             Authorization = "users.list"
	AuthorizationUsersUpdate           Authorization = "users.update"

	RoleAdmin                 Role = "Admin"
	RoleServiceAccountCreator Role = "Service account creator"
	RoleServiceAccountOwner   Role = "Service account owner"
	RoleTeamCreator           Role = "Team creator"
	RoleTeamMember            Role = "Team member"
	RoleTeamOwner             Role = "Team owner"
	RoleTeamViewer            Role = "Team viewer"
	RoleUserAdmin             Role = "User admin"
	RoleUserViewer            Role = "User viewer"
)
