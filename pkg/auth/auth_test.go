package auth_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/auth"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/stretchr/testify/assert"
)

func TestSimpleAllowDeny(t *testing.T) {
	unusedSystem := uuid.New()
	systemID := uuid.New()

	const (
		allowReadableResource = "should_allow_read"
		allowWritableResource = "should_allow_write"
	)

	system := &dbmodels.System{
		Model: dbmodels.Model{
			ID: &systemID,
		},
	}

	// We do not need to set any user or team references on these roles,
	// as roles are prefetched before they are sent to the Allowed() function.
	roles := []*dbmodels.Role{
		{
			SystemID:    &unusedSystem,
			Resource:    allowReadableResource,
			AccessLevel: auth.AccessRead,
			Permission:  auth.PermissionDeny,
		},
		{
			SystemID:    &systemID,
			Resource:    allowReadableResource,
			AccessLevel: auth.AccessRead,
			Permission:  auth.PermissionAllow,
		},
		{
			SystemID:    &systemID,
			Resource:    allowWritableResource,
			AccessLevel: auth.AccessReadWrite,
			Permission:  auth.PermissionAllow,
		},
	}

	rolebindings := makeRoleBindings(roles)

	// Read access should allow both read and readwrite permissions
	assert.NoError(t, auth.AllowedRoles(rolebindings, system, nil, auth.AccessRead, allowReadableResource))
	assert.NoError(t, auth.AllowedRoles(rolebindings, system, nil, auth.AccessRead, allowWritableResource))

	// Write access should apply only to readwrite permission
	assert.NoError(t, auth.AllowedRoles(rolebindings, system, nil, auth.AccessReadWrite, allowWritableResource))
	assert.Error(t, auth.AllowedRoles(rolebindings, system, nil, auth.AccessReadWrite, allowReadableResource))
}

// Roles MUST be defined for any resource to be accessible.
func TestDefaultDeny(t *testing.T) {
	systemID := uuid.New()

	system := &dbmodels.System{
		Model: dbmodels.Model{
			ID: &systemID,
		},
	}

	rolebindings := make([]*dbmodels.RoleBinding, 0)

	assert.Error(t, auth.AllowedRoles(rolebindings, system, nil, auth.AccessRead, "you"))
	assert.Error(t, auth.AllowedRoles(rolebindings, system, nil, auth.AccessRead, "shall"))
	assert.Error(t, auth.AllowedRoles(rolebindings, system, nil, auth.AccessReadWrite, "not"))
	assert.Error(t, auth.AllowedRoles(rolebindings, system, nil, auth.AccessReadWrite, "pass"))
}

// Check that an explicit DENY for a resource overrules an explicit ALLOW.
func TestAllowDenyOrdering(t *testing.T) {
	systemID := uuid.New()

	const resource = "myresource"

	system := &dbmodels.System{
		Model: dbmodels.Model{
			ID: &systemID,
		},
	}

	roles := []*dbmodels.Role{
		{
			SystemID:    &systemID,
			Resource:    resource,
			AccessLevel: auth.AccessRead,
			Permission:  auth.PermissionAllow,
		},
		{
			SystemID:    &systemID,
			Resource:    resource,
			AccessLevel: auth.AccessRead,
			Permission:  auth.PermissionDeny,
		},
	}

	rolebindings := makeRoleBindings(roles)

	// Access should be denied to both read and write
	assert.Error(t, auth.AllowedRoles(rolebindings, system, nil, auth.AccessRead, resource))
	assert.Error(t, auth.AllowedRoles(rolebindings, system, nil, auth.AccessReadWrite, resource))
}

// Check that denied reads for a resource overrules allowed writes.
func TestExplicitDeny(t *testing.T) {
	systemID := uuid.New()

	const resource = "myresource"

	system := &dbmodels.System{
		Model: dbmodels.Model{
			ID: &systemID,
		},
	}

	roles := []*dbmodels.Role{
		{
			SystemID:    &systemID,
			Resource:    resource,
			AccessLevel: auth.AccessReadWrite,
			Permission:  auth.PermissionAllow,
		},
		{
			SystemID:    &systemID,
			Resource:    resource,
			AccessLevel: auth.AccessRead,
			Permission:  auth.PermissionDeny,
		},
	}

	rolebindings := makeRoleBindings(roles)

	assert.Error(t, auth.AllowedRoles(rolebindings, system, nil, auth.AccessRead, resource))
	assert.Error(t, auth.AllowedRoles(rolebindings, system, nil, auth.AccessReadWrite, resource))
}

// When checking access for multiple resources, fail fast if any of them are denied.
func TestAnyDenyWins(t *testing.T) {
	systemID := uuid.New()

	const firstResource = "first_resource"
	const secondResource = "second_resource"

	system := &dbmodels.System{
		Model: dbmodels.Model{
			ID: &systemID,
		},
	}

	roles := []*dbmodels.Role{
		{
			SystemID:    &systemID,
			Resource:    firstResource,
			AccessLevel: auth.AccessReadWrite,
			Permission:  auth.PermissionAllow,
		},
		{
			SystemID:    &systemID,
			Resource:    secondResource,
			AccessLevel: auth.AccessRead,
			Permission:  auth.PermissionDeny,
		},
	}

	rolebindings := makeRoleBindings(roles)

	assert.Error(t, auth.AllowedRoles(rolebindings, system, nil, auth.AccessRead, firstResource, secondResource))
	assert.Error(t, auth.AllowedRoles(rolebindings, system, nil, auth.AccessReadWrite, firstResource, secondResource))

	// Switch ordering
	assert.Error(t, auth.AllowedRoles(rolebindings, system, nil, auth.AccessRead, secondResource, firstResource))
	assert.Error(t, auth.AllowedRoles(rolebindings, system, nil, auth.AccessReadWrite, secondResource, firstResource))
}

func makeRoleBindings(roles []*dbmodels.Role) []*dbmodels.RoleBinding {
	rolebindings := make([]*dbmodels.RoleBinding, 0)
	for _, role := range roles {
		rb := &dbmodels.RoleBinding{
			Role: role,
		}
		rolebindings = append(rolebindings, rb)
	}
	return rolebindings
}
