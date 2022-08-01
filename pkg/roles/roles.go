package roles

import (
	"errors"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/dbmodels"
)

type Authorization string
type Role string

const (
	AuthorizationAuditLogsRead         Authorization = "audit_logs.read"
	AuthorizationServiceAccountsCreate Authorization = "service_accounts.create"
	AuthorizationServiceAccountsDelete Authorization = "service_accounts.delete"
	AuthorizationServiceAccountList    Authorization = "service_accounts.list"
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

	RoleAdmin                 Role = "Admin"
	RoleServiceAccountCreator Role = "Service account creator"
	RoleServiceAccountOwner   Role = "Service account owner"
	RoleTeamCreator           Role = "Team creator"
	RoleTeamMember            Role = "Team member"
	RoleTeamOwner             Role = "Team owner"
	RoleTeamViewer            Role = "Team viewer"
	RoleUserViewer            Role = "User viewer"
)

var ErrNotAuthorized = errors.New("not authorized")

// RequireGlobalAuthorization Require a user to have a specific authorization through a globally assigned role. The role
// bindings must already be attached to the user.
func RequireGlobalAuthorization(user *dbmodels.User, requiredAuthorization Authorization) error {
	if user == nil {
		return ErrNotAuthorized
	}

	authorizations := make(map[dbmodels.Authorization]struct{})

	for _, roleBinding := range user.RoleBindings {
		if roleBinding.TargetID == nil {
			for _, authorization := range roleBinding.Role.Authorizations {
				authorizations[authorization] = struct{}{}
			}
		}
	}

	return authorized(authorizations, requiredAuthorization)
}

// RequireAuthorization Require a user to have a specific authorization through a globally assigned or a correctly
// targetted role. The role bindings must already be attached to the user.
func RequireAuthorization(user *dbmodels.User, requiredAuthorization Authorization, target uuid.UUID) error {
	if user == nil {
		return ErrNotAuthorized
	}

	authorizations := make(map[dbmodels.Authorization]struct{})

	for _, roleBinding := range user.RoleBindings {
		if roleBinding.TargetID == nil || *roleBinding.TargetID == target {
			for _, authorization := range roleBinding.Role.Authorizations {
				authorizations[authorization] = struct{}{}
			}
		}
	}

	return authorized(authorizations, requiredAuthorization)
}

// authorized Check if one of the authorizations in the map matches the required authorization.
func authorized(authorizations map[dbmodels.Authorization]struct{}, requiredAuthorization Authorization) error {
	for authorization, _ := range authorizations {
		if Authorization(authorization.Name) == requiredAuthorization {
			return nil
		}
	}

	return ErrNotAuthorized
}
