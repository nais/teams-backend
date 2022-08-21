// Code generated by mockery v2.14.0. DO NOT EDIT.

package db

import (
	context "context"

	sqlc "github.com/nais/console/pkg/sqlc"
	mock "github.com/stretchr/testify/mock"

	uuid "github.com/google/uuid"
)

// MockDatabase is an autogenerated mock type for the Database type
type MockDatabase struct {
	mock.Mock
}

// AddAuditLog provides a mock function with given fields: ctx, correlationID, actorEmail, systemName, targetTeamSlug, targetUserEmail, action, message
func (_m *MockDatabase) AddAuditLog(ctx context.Context, correlationID uuid.UUID, actorEmail *string, systemName *sqlc.SystemName, targetTeamSlug *string, targetUserEmail *string, action sqlc.AuditAction, message string) error {
	ret := _m.Called(ctx, correlationID, actorEmail, systemName, targetTeamSlug, targetUserEmail, action, message)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, *string, *sqlc.SystemName, *string, *string, sqlc.AuditAction, string) error); ok {
		r0 = rf(ctx, correlationID, actorEmail, systemName, targetTeamSlug, targetUserEmail, action, message)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// AddTeam provides a mock function with given fields: ctx, name, slug, purpose
func (_m *MockDatabase) AddTeam(ctx context.Context, name string, slug string, purpose *string) (*Team, error) {
	ret := _m.Called(ctx, name, slug, purpose)

	var r0 *Team
	if rf, ok := ret.Get(0).(func(context.Context, string, string, *string) *Team); ok {
		r0 = rf(ctx, name, slug, purpose)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Team)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, *string) error); ok {
		r1 = rf(ctx, name, slug, purpose)
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

// GetRoleNames provides a mock function with given fields: ctx
func (_m *MockDatabase) GetRoleNames(ctx context.Context) ([]sqlc.RoleName, error) {
	ret := _m.Called(ctx)

	var r0 []sqlc.RoleName
	if rf, ok := ret.Get(0).(func(context.Context) []sqlc.RoleName); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]sqlc.RoleName)
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

// GetTeamBySlug provides a mock function with given fields: ctx, slug
func (_m *MockDatabase) GetTeamBySlug(ctx context.Context, slug string) (*Team, error) {
	ret := _m.Called(ctx, slug)

	var r0 *Team
	if rf, ok := ret.Get(0).(func(context.Context, string) *Team); ok {
		r0 = rf(ctx, slug)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Team)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, slug)
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
