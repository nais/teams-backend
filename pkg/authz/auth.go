package authz

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/nais/console/pkg/dbmodels"
)

const (
	userContextKey     = "user"
	roleBindingsCtxKey = "roleBindings"

	AccessLevelCreate AccessLevel = "C"
	AccessLevelRead   AccessLevel = "R"
	AccessLevelUpdate AccessLevel = "U"
	AccessLevelDelete AccessLevel = "D"

	PermissionAllow = "allow"
	PermissionDeny  = "deny"

	ResourceTeams Resource = "teams"
	ResourceRoles Resource = "roles"
)

type AccessLevel string
type Resource string

func (r Resource) Format(args ...interface{}) Resource {
	return Resource(fmt.Sprintf(string(r), args...))
}

// ContextWithUser Return a context with a user module stored.
func ContextWithUser(ctx context.Context, user *dbmodels.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// ContextWithRoleBindings Insert user/team roles into a context
func ContextWithRoleBindings(ctx context.Context, roleBindings []*dbmodels.RoleBinding) context.Context {
	return context.WithValue(ctx, roleBindingsCtxKey, roleBindings)
}

// UserFromContext Finds any authenticated user from the context. Requires that a middleware has stored a user in the first place.
func UserFromContext(ctx context.Context) *dbmodels.User {
	user, _ := ctx.Value(userContextKey).(*dbmodels.User)
	return user
}

// RoleBindingsFromContext Finds any authenticated user from the context. Requires that a middleware authenticator has set the user.
func RoleBindingsFromContext(ctx context.Context) []*dbmodels.RoleBinding {
	roles, _ := ctx.Value(roleBindingsCtxKey).([]*dbmodels.RoleBinding)
	return roles
}

// Authorized Check if the user currently stored in the context is au
func Authorized(ctx context.Context, system *dbmodels.System, team *dbmodels.Team, requiredAccessLevel AccessLevel, resource Resource) error {
	return RoleBindingsAreAuthorized(RoleBindingsFromContext(ctx), system, team, requiredAccessLevel, resource)
}

// RoleBindingsAreAuthorized Check if the roles defined in the access control list have the necessary access level for the specified resource
// Returns nil if action is allowed, or an error if denied.
//
// The first matching rule wins out, but note that DENY rules are prioritized before ALLOW rules.
func RoleBindingsAreAuthorized(roleBindings []*dbmodels.RoleBinding, system *dbmodels.System, team *dbmodels.Team, requiredAccessLevel AccessLevel, resource Resource) error {
	unauthorized := fmt.Errorf("unauthorized")

	// Sort 'deny' roles before 'allow' so that we can fail fast.
	sort.Slice(roleBindings, func(i, j int) bool {
		return roleBindings[i].Role.Permission > roleBindings[j].Role.Permission
	})

	for _, roleBinding := range roleBindings {
		// ignore role binding if systems does not match
		if roleBinding.Role.SystemID != *system.ID {
			continue
		}

		// ignore role binding if resource does not match
		if Resource(roleBinding.Role.Resource) != resource {
			continue
		}

		// if the role binding is for a specific team, and the check does not require a team, ignore it
		if team == nil && roleBinding.Team != nil {
			continue
		}

		// if a team is required, check if the role binding matches that team, or if it is set globally
		if team != nil && roleBinding.Team != nil && *roleBinding.Team.ID != *team.ID {
			continue
		}

		// At this point, the role we are working with is pointing directly to the resource in question.
		requiredLevel := hasRequiredAccessLevel(AccessLevel(roleBinding.Role.AccessLevel), requiredAccessLevel)

		if roleBinding.Role.Permission == PermissionDeny && requiredLevel {
			return unauthorized
		}

		if roleBinding.Role.Permission == PermissionAllow && requiredLevel {
			return nil
		}
	}

	return unauthorized
}

func hasRequiredAccessLevel(hasAccessLevel, requiredAccessLevel AccessLevel) bool {
	return strings.Contains(string(hasAccessLevel), string(requiredAccessLevel))
}
