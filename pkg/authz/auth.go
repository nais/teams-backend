package authz

import (
	"context"
	"fmt"
	"sort"

	"github.com/nais/console/pkg/dbmodels"
)

var userCtxKey = &contextKey{"user"}
var roleBindingsCtxKey = &contextKey{"rolebindings"}

type contextKey struct {
	name string
}

const (
	AccessRead      = "read"
	AccessReadWrite = "readwrite"
	AccessCreate    = "create"

	PermissionAllow = "allow"
	PermissionDeny  = "deny"
)

type Resource string

func (r Resource) Format(args ...interface{}) Resource {
	return Resource(fmt.Sprintf(string(r), args...))
}

// Finds any authenticated user from the context. Requires that a middleware authenticator has set the user.
func UserFromContext(ctx context.Context) *dbmodels.User {
	user, _ := ctx.Value(userCtxKey).(*dbmodels.User)
	return user
}

// Insert a user object into a context.
func ContextWithUser(ctx context.Context, user *dbmodels.User) context.Context {
	return context.WithValue(ctx, userCtxKey, user)
}

// Finds any authenticated user from the context. Requires that a middleware authenticator has set the user.
func RoleBindingsFromContext(ctx context.Context) []*dbmodels.RoleBinding {
	roles, _ := ctx.Value(roleBindingsCtxKey).([]*dbmodels.RoleBinding)
	return roles
}

// Insert user/team roles into a context
func ContextWithRoleBindings(ctx context.Context, roles []*dbmodels.RoleBinding) context.Context {
	return context.WithValue(ctx, roleBindingsCtxKey, roles)
}

func Allowed(ctx context.Context, system *dbmodels.System, team *dbmodels.Team, accessLevel string, resources ...Resource) error {
	return AllowedRoles(RoleBindingsFromContext(ctx), system, team, accessLevel, resources...)
}

// Check if the roles defined in the access control list allow read or write access to one or more specific system resources.
// Returns nil if action is allowed, or an error if denied.
//
// The first matching rule wins out, but note that DENY rules are prioritized before ALLOW rules.
func AllowedRoles(roleBindings []*dbmodels.RoleBinding, system *dbmodels.System, team *dbmodels.Team, accessLevel string, resources ...Resource) error {

	//goland:noinspection GoErrorStringFormat
	unauthorized := fmt.Errorf("YOU DIDN'T SAY THE MAGIC WORD!")

	// Sort 'deny' roles before 'allow' so that we can fail fast.
	sort.Slice(roleBindings, func(i, j int) bool {
		return roleBindings[i].Role.Permission > roleBindings[j].Role.Permission
	})

	for _, roleBinding := range roleBindings {

		// Skip unmatching systems
		if *roleBinding.Role.SystemID != *system.ID {
			continue
		}

		// If a team permission is needed, check that the rolebinding team matches, or is set to nil (global permission).
		if team != nil && roleBinding.Team != nil && *roleBinding.Team.ID != *team.ID {
			continue
		}

		for _, resource := range resources {
			// Skip unmatching resources
			if Resource(roleBinding.Role.Resource) != resource {
				continue
			}

			// At this point, the role we are working with is pointing directly to the resource in question.
			if roleBinding.Role.Permission == PermissionDeny || !hasAccessLevel(accessLevel, roleBinding.Role.AccessLevel) {
				return unauthorized
			}

			return nil
		}
	}

	return unauthorized
}

func hasAccessLevel(needed, have string) bool {
	switch needed {
	case AccessRead:
		return have == AccessRead || have == AccessReadWrite
	case AccessReadWrite, AccessCreate:
		return have == needed
	default:
		return false
	}
}
