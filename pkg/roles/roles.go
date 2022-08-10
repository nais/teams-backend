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

var ErrNotAuthorized = errors.New("not authorized")

// RequireGlobalAuthorization Require an actor to have a specific authorization through a globally assigned role. The
// role bindings must already be attached to the actor.
func RequireGlobalAuthorization(actor *dbmodels.User, requiredAuthorization Authorization) error {
	if actor == nil {
		return ErrNotAuthorized
	}

	authorizations := make(map[dbmodels.Authorization]struct{})

	for _, roleBinding := range actor.RoleBindings {
		if roleBinding.TargetID == nil {
			for _, authorization := range roleBinding.Role.Authorizations {
				authorizations[authorization] = struct{}{}
			}
		}
	}

	return authorized(authorizations, requiredAuthorization)
}

// RequireAuthorization Require an actor to have a specific authorization through a globally assigned or a correctly
// targetted role. The role bindings must already be attached to the actor.
func RequireAuthorization(actor *dbmodels.User, requiredAuthorization Authorization, target uuid.UUID) error {
	if actor == nil {
		return ErrNotAuthorized
	}

	authorizations := make(map[dbmodels.Authorization]struct{})

	for _, roleBinding := range actor.RoleBindings {
		if roleBinding.TargetID == nil || *roleBinding.TargetID == target {
			for _, authorization := range roleBinding.Role.Authorizations {
				authorizations[authorization] = struct{}{}
			}
		}
	}

	return authorized(authorizations, requiredAuthorization)
}

// RequireAuthorizationOrTargetMatch Require an actor to have a specific authorization through a globally assigned or a
// correctly targetted role. The role bindings must already be attached to the actor. If the actor matches the target,
// the action will be allowed.
func RequireAuthorizationOrTargetMatch(actor *dbmodels.User, requiredAuthorization Authorization, target uuid.UUID) error {
	if actor != nil && *actor.ID == target {
		return nil
	}

	return RequireAuthorization(actor, requiredAuthorization, target)
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
