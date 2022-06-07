package authz_test

import (
	"context"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/authz"
	helpers "github.com/nais/console/pkg/console"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestContextWithUser(t *testing.T) {
	ctx := context.Background()
	assert.Nil(t, authz.UserFromContext(ctx))

	user := &dbmodels.User{
		Email: helpers.Strp("mail@example.com"),
		Name:  helpers.Strp("User Name"),
	}

	ctx = authz.ContextWithUser(ctx, user)
	assert.Equal(t, user, authz.UserFromContext(ctx))
}

func TestContextWithRoleBindings(t *testing.T) {
	ctx := context.Background()
	assert.Nil(t, authz.RoleBindingsFromContext(ctx))

	roleBindings := []*dbmodels.RoleBinding{
		{
			User: &dbmodels.User{
				Email: helpers.Strp("mail1@example.com"),
			},
		},
		{
			User: &dbmodels.User{
				Email: helpers.Strp("mail2@example.com"),
			},
		},
	}

	ctx = authz.ContextWithRoleBindings(ctx, roleBindings)
	assert.Equal(t, roleBindings, authz.RoleBindingsFromContext(ctx))
}

func TestSimpleAllowDeny(t *testing.T) {
	unusedSystem := uuid.New()
	systemID := uuid.New()
	system := &dbmodels.System{
		Model: dbmodels.Model{
			ID: &systemID,
		},
	}

	// We do not need to set any user or team references on these roles,
	// as roles are prefetched before they are sent to the Authorized() function.
	roles := []*dbmodels.Role{
		{
			SystemID:    &unusedSystem,
			Resource:    "teams",
			AccessLevel: string(authz.AccessLevelRead),
			Permission:  authz.PermissionDeny,
		},
		{
			SystemID:    &systemID,
			Resource:    "teams",
			AccessLevel: string(authz.AccessLevelRead),
			Permission:  authz.PermissionAllow,
		},
		{
			SystemID:    &systemID,
			Resource:    "roles",
			AccessLevel: string(authz.AccessLevelCreate),
			Permission:  authz.PermissionAllow,
		},
	}

	roleBindings := makeRoleBindings(roles)

	assert.NoError(t, authz.RoleBindingsAreAuthorized(roleBindings, system, nil, authz.AccessLevelRead, "teams"))
	assert.EqualError(t, authz.RoleBindingsAreAuthorized(roleBindings, system, nil, authz.AccessLevelCreate, "teams"), "unauthorized")
}

// Roles MUST be defined for any resource to be accessible.
func TestDefaultDeny(t *testing.T) {
	systemID := uuid.New()

	system := &dbmodels.System{
		Model: dbmodels.Model{
			ID: &systemID,
		},
	}

	roleBindings := make([]*dbmodels.RoleBinding, 0)

	assert.EqualError(t, authz.RoleBindingsAreAuthorized(roleBindings, system, nil, authz.AccessLevelCreate, "you"), "unauthorized")
	assert.EqualError(t, authz.RoleBindingsAreAuthorized(roleBindings, system, nil, authz.AccessLevelRead, "shall"), "unauthorized")
	assert.EqualError(t, authz.RoleBindingsAreAuthorized(roleBindings, system, nil, authz.AccessLevelUpdate, "not"), "unauthorized")
	assert.EqualError(t, authz.RoleBindingsAreAuthorized(roleBindings, system, nil, authz.AccessLevelDelete, "pass"), "unauthorized")
}

// Check that an explicit DENY for a resource overrules an explicit ALLOW.
func TestAllowDenyOrdering(t *testing.T) {
	systemID := uuid.New()
	system := &dbmodels.System{
		Model: dbmodels.Model{
			ID: &systemID,
		},
	}
	roles := []*dbmodels.Role{
		{
			SystemID:    &systemID,
			Resource:    "teams",
			AccessLevel: string(authz.AccessLevelRead),
			Permission:  authz.PermissionAllow,
		},
		{
			SystemID:    &systemID,
			Resource:    "teams",
			AccessLevel: string(authz.AccessLevelRead),
			Permission:  authz.PermissionDeny,
		},
	}
	roleBindings := makeRoleBindings(roles)
	assert.EqualError(t, authz.RoleBindingsAreAuthorized(roleBindings, system, nil, authz.AccessLevelRead, "teams"), "unauthorized")
}

// Check that denied reads for a resource does not overrule allowed writes.
func TestExplicitDeny(t *testing.T) {
	systemID := uuid.New()
	system := &dbmodels.System{
		Model: dbmodels.Model{
			ID: &systemID,
		},
	}
	roles := []*dbmodels.Role{
		{
			SystemID:    &systemID,
			Resource:    "teams",
			AccessLevel: string(authz.AccessLevelCreate),
			Permission:  authz.PermissionAllow,
		},
		{
			SystemID:    &systemID,
			Resource:    "teams",
			AccessLevel: string(authz.AccessLevelRead),
			Permission:  authz.PermissionDeny,
		},
	}
	roleBindings := makeRoleBindings(roles)
	assert.EqualError(t, authz.RoleBindingsAreAuthorized(roleBindings, system, nil, authz.AccessLevelRead, "teams"), "unauthorized")
	assert.NoError(t, authz.RoleBindingsAreAuthorized(roleBindings, system, nil, authz.AccessLevelCreate, "teams"))
}

func makeRoleBindings(roles []*dbmodels.Role) []*dbmodels.RoleBinding {
	roleBindings := make([]*dbmodels.RoleBinding, 0)
	for _, role := range roles {
		rb := &dbmodels.RoleBinding{
			Role: role,
		}
		roleBindings = append(roleBindings, rb)
	}
	return roleBindings
}
