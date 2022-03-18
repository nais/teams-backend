package auth

import (
	"context"
	"fmt"
	"sort"

	"github.com/nais/console/pkg/dbmodels"
)

var userCtxKey = &contextKey{"user"}
var rolesCtxKey = &contextKey{"roles"}

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
func RolesFromContext(ctx context.Context) []*dbmodels.Role {
	roles, _ := ctx.Value(rolesCtxKey).([]*dbmodels.Role)
	return roles
}

// Insert user/team roles into a context
func ContextWithRoles(ctx context.Context, roles []*dbmodels.Role) context.Context {
	return context.WithValue(ctx, rolesCtxKey, roles)
}

func Allowed(ctx context.Context, system *dbmodels.System, accessLevel string, resources ...Resource) error {
	return AllowedRoles(RolesFromContext(ctx), system, accessLevel, resources...)
}

// Check if the roles defined in the access control list allow read or write access to one or more specific system resources.
// Returns nil if action is allowed, or an error if denied.
//
// The first matching rule wins out, but note that DENY rules are prioritized before ALLOW rules.
func AllowedRoles(roles []*dbmodels.Role, system *dbmodels.System, accessLevel string, resources ...Resource) error {

	//goland:noinspection GoErrorStringFormat
	unauthorized := fmt.Errorf("YOU DIDN'T SAY THE MAGIC WORD!")

	// Sort 'deny' roles before 'allow' so that we can fail fast.
	sort.Slice(roles, func(i, j int) bool {
		return roles[i].Permission > roles[j].Permission
	})

	for _, role := range roles {

		// Skip unmatching systems
		if role.SystemID != system.ID {
			continue
		}

		for _, resource := range resources {
			// Skip unmatching resources
			if Resource(role.Resource) != resource {
				continue
			}

			// At this point, the role we are working with is pointing directly to the resource in question.
			if role.Permission == PermissionDeny || !hasAccessLevel(accessLevel, role.AccessLevel) {
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
