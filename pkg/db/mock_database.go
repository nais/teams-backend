// Code generated by mockery v2.14.0. DO NOT EDIT.

package db

import (
	context "context"

	slug "github.com/nais/console/pkg/slug"
	mock "github.com/stretchr/testify/mock"

	sqlc "github.com/nais/console/pkg/sqlc"

	uuid "github.com/google/uuid"
)

// MockDatabase is an autogenerated mock type for the Database type
type MockDatabase struct {
	mock.Mock
}

// AddAuditLog provides a mock function with given fields: ctx, correlationID, systemName, actorEmail, targetTeamSlug, targetUserEmail, action, message
func (_m *MockDatabase) AddAuditLog(ctx context.Context, correlationID uuid.UUID, systemName sqlc.SystemName, actorEmail *string, targetTeamSlug *slug.Slug, targetUserEmail *string, action sqlc.AuditAction, message string) error {
	ret := _m.Called(ctx, correlationID, systemName, actorEmail, targetTeamSlug, targetUserEmail, action, message)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, sqlc.SystemName, *string, *slug.Slug, *string, sqlc.AuditAction, string) error); ok {
		r0 = rf(ctx, correlationID, systemName, actorEmail, targetTeamSlug, targetUserEmail, action, message)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// AddServiceAccount provides a mock function with given fields: ctx, name, email, userID
func (_m *MockDatabase) AddServiceAccount(ctx context.Context, name slug.Slug, email string, userID uuid.UUID) (*User, error) {
	ret := _m.Called(ctx, name, email, userID)

	var r0 *User
	if rf, ok := ret.Get(0).(func(context.Context, slug.Slug, string, uuid.UUID) *User); ok {
		r0 = rf(ctx, name, email, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, slug.Slug, string, uuid.UUID) error); ok {
		r1 = rf(ctx, name, email, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AddTeam provides a mock function with given fields: ctx, name, _a2, purpose, userID
func (_m *MockDatabase) AddTeam(ctx context.Context, name string, _a2 slug.Slug, purpose *string, userID uuid.UUID) (*Team, error) {
	ret := _m.Called(ctx, name, _a2, purpose, userID)

	var r0 *Team
	if rf, ok := ret.Get(0).(func(context.Context, string, slug.Slug, *string, uuid.UUID) *Team); ok {
		r0 = rf(ctx, name, _a2, purpose, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Team)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, slug.Slug, *string, uuid.UUID) error); ok {
		r1 = rf(ctx, name, _a2, purpose, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AddUser provides a mock function with given fields: ctx, name, email
func (_m *MockDatabase) AddUser(ctx context.Context, name string, email string) (*User, error) {
	ret := _m.Called(ctx, name, email)

	var r0 *User
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *User); ok {
		r0 = rf(ctx, name, email)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, name, email)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AddUserToTeam provides a mock function with given fields: ctx, userID, teamID
func (_m *MockDatabase) AddUserToTeam(ctx context.Context, userID uuid.UUID, teamID uuid.UUID) error {
	ret := _m.Called(ctx, userID, teamID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, uuid.UUID) error); ok {
		r0 = rf(ctx, userID, teamID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// AssignGlobalRoleToUser provides a mock function with given fields: ctx, userID, roleName
func (_m *MockDatabase) AssignGlobalRoleToUser(ctx context.Context, userID uuid.UUID, roleName sqlc.RoleName) error {
	ret := _m.Called(ctx, userID, roleName)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, sqlc.RoleName) error); ok {
		r0 = rf(ctx, userID, roleName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// AssignTargetedRoleToUser provides a mock function with given fields: ctx, userID, roleName, targetID
func (_m *MockDatabase) AssignTargetedRoleToUser(ctx context.Context, userID uuid.UUID, roleName sqlc.RoleName, targetID uuid.UUID) error {
	ret := _m.Called(ctx, userID, roleName, targetID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, sqlc.RoleName, uuid.UUID) error); ok {
		r0 = rf(ctx, userID, roleName, targetID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateAPIKey provides a mock function with given fields: ctx, apiKey, userID
func (_m *MockDatabase) CreateAPIKey(ctx context.Context, apiKey string, userID uuid.UUID) error {
	ret := _m.Called(ctx, apiKey, userID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, uuid.UUID) error); ok {
		r0 = rf(ctx, apiKey, userID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteUser provides a mock function with given fields: ctx, userID
func (_m *MockDatabase) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	ret := _m.Called(ctx, userID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, userID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAuditLogsForTeam provides a mock function with given fields: ctx, _a1
func (_m *MockDatabase) GetAuditLogsForTeam(ctx context.Context, _a1 slug.Slug) ([]*AuditLog, error) {
	ret := _m.Called(ctx, _a1)

	var r0 []*AuditLog
	if rf, ok := ret.Get(0).(func(context.Context, slug.Slug) []*AuditLog); ok {
		r0 = rf(ctx, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*AuditLog)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, slug.Slug) error); ok {
		r1 = rf(ctx, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetRoleNames provides a mock function with given fields:
func (_m *MockDatabase) GetRoleNames() []sqlc.RoleName {
	ret := _m.Called()

	var r0 []sqlc.RoleName
	if rf, ok := ret.Get(0).(func() []sqlc.RoleName); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]sqlc.RoleName)
		}
	}

	return r0
}

// GetSystemNames provides a mock function with given fields:
func (_m *MockDatabase) GetSystemNames() []sqlc.SystemName {
	ret := _m.Called()

	var r0 []sqlc.SystemName
	if rf, ok := ret.Get(0).(func() []sqlc.SystemName); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]sqlc.SystemName)
		}
	}

	return r0
}

// GetTeamByID provides a mock function with given fields: ctx, ID
func (_m *MockDatabase) GetTeamByID(ctx context.Context, ID uuid.UUID) (*Team, error) {
	ret := _m.Called(ctx, ID)

	var r0 *Team
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *Team); ok {
		r0 = rf(ctx, ID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Team)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, ID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTeamBySlug provides a mock function with given fields: ctx, _a1
func (_m *MockDatabase) GetTeamBySlug(ctx context.Context, _a1 slug.Slug) (*Team, error) {
	ret := _m.Called(ctx, _a1)

	var r0 *Team
	if rf, ok := ret.Get(0).(func(context.Context, slug.Slug) *Team); ok {
		r0 = rf(ctx, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Team)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, slug.Slug) error); ok {
		r1 = rf(ctx, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTeamMembers provides a mock function with given fields: ctx, teamID
func (_m *MockDatabase) GetTeamMembers(ctx context.Context, teamID uuid.UUID) ([]*User, error) {
	ret := _m.Called(ctx, teamID)

	var r0 []*User
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) []*User); ok {
		r0 = rf(ctx, teamID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, teamID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTeams provides a mock function with given fields: ctx
func (_m *MockDatabase) GetTeams(ctx context.Context) ([]*Team, error) {
	ret := _m.Called(ctx)

	var r0 []*Team
	if rf, ok := ret.Get(0).(func(context.Context) []*Team); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*Team)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetUserByApiKey provides a mock function with given fields: ctx, APIKey
func (_m *MockDatabase) GetUserByApiKey(ctx context.Context, APIKey string) (*User, error) {
	ret := _m.Called(ctx, APIKey)

	var r0 *User
	if rf, ok := ret.Get(0).(func(context.Context, string) *User); ok {
		r0 = rf(ctx, APIKey)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, APIKey)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetUserByEmail provides a mock function with given fields: ctx, email
func (_m *MockDatabase) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	ret := _m.Called(ctx, email)

	var r0 *User
	if rf, ok := ret.Get(0).(func(context.Context, string) *User); ok {
		r0 = rf(ctx, email)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, email)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetUserByID provides a mock function with given fields: ctx, ID
func (_m *MockDatabase) GetUserByID(ctx context.Context, ID uuid.UUID) (*User, error) {
	ret := _m.Called(ctx, ID)

	var r0 *User
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *User); ok {
		r0 = rf(ctx, ID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, ID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetUserRoles provides a mock function with given fields: ctx, userID
func (_m *MockDatabase) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*Role, error) {
	ret := _m.Called(ctx, userID)

	var r0 []*Role
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) []*Role); ok {
		r0 = rf(ctx, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*Role)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetUserTeams provides a mock function with given fields: ctx, userID
func (_m *MockDatabase) GetUserTeams(ctx context.Context, userID uuid.UUID) ([]*Team, error) {
	ret := _m.Called(ctx, userID)

	var r0 []*Team
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) []*Team); ok {
		r0 = rf(ctx, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*Team)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetUsers provides a mock function with given fields: ctx
func (_m *MockDatabase) GetUsers(ctx context.Context) ([]*User, error) {
	ret := _m.Called(ctx)

	var r0 []*User
	if rf, ok := ret.Get(0).(func(context.Context) []*User); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetUsersByEmail provides a mock function with given fields: ctx, email
func (_m *MockDatabase) GetUsersByEmail(ctx context.Context, email string) ([]*User, error) {
	ret := _m.Called(ctx, email)

	var r0 []*User
	if rf, ok := ret.Get(0).(func(context.Context, string) []*User); ok {
		r0 = rf(ctx, email)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, email)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LoadSystemState provides a mock function with given fields: ctx, systemName, teamID, state
func (_m *MockDatabase) LoadSystemState(ctx context.Context, systemName sqlc.SystemName, teamID uuid.UUID, state interface{}) error {
	ret := _m.Called(ctx, systemName, teamID, state)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, sqlc.SystemName, uuid.UUID, interface{}) error); ok {
		r0 = rf(ctx, systemName, teamID, state)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveApiKeysFromUser provides a mock function with given fields: ctx, userID
func (_m *MockDatabase) RemoveApiKeysFromUser(ctx context.Context, userID uuid.UUID) error {
	ret := _m.Called(ctx, userID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, userID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveUserRoles provides a mock function with given fields: ctx, userID
func (_m *MockDatabase) RemoveUserRoles(ctx context.Context, userID uuid.UUID) error {
	ret := _m.Called(ctx, userID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, userID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetSystemState provides a mock function with given fields: ctx, systemName, teamID, state
func (_m *MockDatabase) SetSystemState(ctx context.Context, systemName sqlc.SystemName, teamID uuid.UUID, state interface{}) error {
	ret := _m.Called(ctx, systemName, teamID, state)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, sqlc.SystemName, uuid.UUID, interface{}) error); ok {
		r0 = rf(ctx, systemName, teamID, state)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetUserName provides a mock function with given fields: ctx, userID, name
func (_m *MockDatabase) SetUserName(ctx context.Context, userID uuid.UUID, name string) (*User, error) {
	ret := _m.Called(ctx, userID, name)

	var r0 *User
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string) *User); ok {
		r0 = rf(ctx, userID, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, string) error); ok {
		r1 = rf(ctx, userID, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Transaction provides a mock function with given fields: ctx, fn
func (_m *MockDatabase) Transaction(ctx context.Context, fn TransactionFunc) error {
	ret := _m.Called(ctx, fn)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, TransactionFunc) error); ok {
		r0 = rf(ctx, fn)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateTeam provides a mock function with given fields: ctx, teamID, name, purpose
func (_m *MockDatabase) UpdateTeam(ctx context.Context, teamID uuid.UUID, name *string, purpose *string) (*Team, error) {
	ret := _m.Called(ctx, teamID, name, purpose)

	var r0 *Team
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, *string, *string) *Team); ok {
		r0 = rf(ctx, teamID, name, purpose)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Team)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, *string, *string) error); ok {
		r1 = rf(ctx, teamID, name, purpose)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UserIsTeamOwner provides a mock function with given fields: ctx, userID, teamID
func (_m *MockDatabase) UserIsTeamOwner(ctx context.Context, userID uuid.UUID, teamID uuid.UUID) (bool, error) {
	ret := _m.Called(ctx, userID, teamID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, uuid.UUID) bool); ok {
		r0 = rf(ctx, userID, teamID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, uuid.UUID) error); ok {
		r1 = rf(ctx, userID, teamID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewMockDatabase interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockDatabase creates a new instance of MockDatabase. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockDatabase(t mockConstructorTestingTNewMockDatabase) *MockDatabase {
	mock := &MockDatabase{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
