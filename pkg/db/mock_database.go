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

// AssignGlobalRoleToServiceAccount provides a mock function with given fields: ctx, serviceAccountID, roleName
func (_m *MockDatabase) AssignGlobalRoleToServiceAccount(ctx context.Context, serviceAccountID uuid.UUID, roleName sqlc.RoleName) error {
	ret := _m.Called(ctx, serviceAccountID, roleName)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, sqlc.RoleName) error); ok {
		r0 = rf(ctx, serviceAccountID, roleName)
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

// ClearReconcilerErrorsForTeam provides a mock function with given fields: ctx, _a1, reconcilerName
func (_m *MockDatabase) ClearReconcilerErrorsForTeam(ctx context.Context, _a1 slug.Slug, reconcilerName sqlc.ReconcilerName) error {
	ret := _m.Called(ctx, _a1, reconcilerName)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, slug.Slug, sqlc.ReconcilerName) error); ok {
		r0 = rf(ctx, _a1, reconcilerName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ConfigureReconciler provides a mock function with given fields: ctx, reconcilerName, config
func (_m *MockDatabase) ConfigureReconciler(ctx context.Context, reconcilerName sqlc.ReconcilerName, config map[sqlc.ReconcilerConfigKey]string) (*Reconciler, error) {
	ret := _m.Called(ctx, reconcilerName, config)

	var r0 *Reconciler
	if rf, ok := ret.Get(0).(func(context.Context, sqlc.ReconcilerName, map[sqlc.ReconcilerConfigKey]string) *Reconciler); ok {
		r0 = rf(ctx, reconcilerName, config)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Reconciler)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, sqlc.ReconcilerName, map[sqlc.ReconcilerConfigKey]string) error); ok {
		r1 = rf(ctx, reconcilerName, config)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateAPIKey provides a mock function with given fields: ctx, apiKey, serviceAccountID
func (_m *MockDatabase) CreateAPIKey(ctx context.Context, apiKey string, serviceAccountID uuid.UUID) error {
	ret := _m.Called(ctx, apiKey, serviceAccountID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, uuid.UUID) error); ok {
		r0 = rf(ctx, apiKey, serviceAccountID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateAuditLogEntry provides a mock function with given fields: ctx, correlationID, systemName, actor, targetType, targetIdentifier, action, message
func (_m *MockDatabase) CreateAuditLogEntry(ctx context.Context, correlationID uuid.UUID, systemName sqlc.SystemName, actor *string, targetType sqlc.AuditLogsTargetType, targetIdentifier string, action sqlc.AuditAction, message string) error {
	ret := _m.Called(ctx, correlationID, systemName, actor, targetType, targetIdentifier, action, message)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, sqlc.SystemName, *string, sqlc.AuditLogsTargetType, string, sqlc.AuditAction, string) error); ok {
		r0 = rf(ctx, correlationID, systemName, actor, targetType, targetIdentifier, action, message)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateServiceAccount provides a mock function with given fields: ctx, name
func (_m *MockDatabase) CreateServiceAccount(ctx context.Context, name string) (*ServiceAccount, error) {
	ret := _m.Called(ctx, name)

	var r0 *ServiceAccount
	if rf, ok := ret.Get(0).(func(context.Context, string) *ServiceAccount); ok {
		r0 = rf(ctx, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ServiceAccount)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateSession provides a mock function with given fields: ctx, userID
func (_m *MockDatabase) CreateSession(ctx context.Context, userID uuid.UUID) (*Session, error) {
	ret := _m.Called(ctx, userID)

	var r0 *Session
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *Session); ok {
		r0 = rf(ctx, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Session)
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

// CreateTeam provides a mock function with given fields: ctx, _a1, purpose
func (_m *MockDatabase) CreateTeam(ctx context.Context, _a1 slug.Slug, purpose string) (*Team, error) {
	ret := _m.Called(ctx, _a1, purpose)

	var r0 *Team
	if rf, ok := ret.Get(0).(func(context.Context, slug.Slug, string) *Team); ok {
		r0 = rf(ctx, _a1, purpose)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Team)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, slug.Slug, string) error); ok {
		r1 = rf(ctx, _a1, purpose)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateUser provides a mock function with given fields: ctx, name, email, externalID
func (_m *MockDatabase) CreateUser(ctx context.Context, name string, email string, externalID string) (*User, error) {
	ret := _m.Called(ctx, name, email, externalID)

	var r0 *User
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) *User); ok {
		r0 = rf(ctx, name, email, externalID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, string) error); ok {
		r1 = rf(ctx, name, email, externalID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DangerousGetReconcilerConfigValues provides a mock function with given fields: ctx, reconcilerName
func (_m *MockDatabase) DangerousGetReconcilerConfigValues(ctx context.Context, reconcilerName sqlc.ReconcilerName) (*ReconcilerConfigValues, error) {
	ret := _m.Called(ctx, reconcilerName)

	var r0 *ReconcilerConfigValues
	if rf, ok := ret.Get(0).(func(context.Context, sqlc.ReconcilerName) *ReconcilerConfigValues); ok {
		r0 = rf(ctx, reconcilerName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ReconcilerConfigValues)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, sqlc.ReconcilerName) error); ok {
		r1 = rf(ctx, reconcilerName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteServiceAccount provides a mock function with given fields: ctx, serviceAccountID
func (_m *MockDatabase) DeleteServiceAccount(ctx context.Context, serviceAccountID uuid.UUID) error {
	ret := _m.Called(ctx, serviceAccountID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, serviceAccountID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteSession provides a mock function with given fields: ctx, sessionID
func (_m *MockDatabase) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	ret := _m.Called(ctx, sessionID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, sessionID)
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

// DisableReconciler provides a mock function with given fields: ctx, reconcilerName
func (_m *MockDatabase) DisableReconciler(ctx context.Context, reconcilerName sqlc.ReconcilerName) (*Reconciler, error) {
	ret := _m.Called(ctx, reconcilerName)

	var r0 *Reconciler
	if rf, ok := ret.Get(0).(func(context.Context, sqlc.ReconcilerName) *Reconciler); ok {
		r0 = rf(ctx, reconcilerName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Reconciler)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, sqlc.ReconcilerName) error); ok {
		r1 = rf(ctx, reconcilerName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DisableTeam provides a mock function with given fields: ctx, teamID
func (_m *MockDatabase) DisableTeam(ctx context.Context, teamID uuid.UUID) (*Team, error) {
	ret := _m.Called(ctx, teamID)

	var r0 *Team
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *Team); ok {
		r0 = rf(ctx, teamID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Team)
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

// EnableReconciler provides a mock function with given fields: ctx, reconcilerName
func (_m *MockDatabase) EnableReconciler(ctx context.Context, reconcilerName sqlc.ReconcilerName) (*Reconciler, error) {
	ret := _m.Called(ctx, reconcilerName)

	var r0 *Reconciler
	if rf, ok := ret.Get(0).(func(context.Context, sqlc.ReconcilerName) *Reconciler); ok {
		r0 = rf(ctx, reconcilerName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Reconciler)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, sqlc.ReconcilerName) error); ok {
		r1 = rf(ctx, reconcilerName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EnableTeam provides a mock function with given fields: ctx, teamID
func (_m *MockDatabase) EnableTeam(ctx context.Context, teamID uuid.UUID) (*Team, error) {
	ret := _m.Called(ctx, teamID)

	var r0 *Team
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *Team); ok {
		r0 = rf(ctx, teamID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Team)
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

// ExtendSession provides a mock function with given fields: ctx, sessionID
func (_m *MockDatabase) ExtendSession(ctx context.Context, sessionID uuid.UUID) (*Session, error) {
	ret := _m.Called(ctx, sessionID)

	var r0 *Session
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *Session); ok {
		r0 = rf(ctx, sessionID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Session)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, sessionID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAuditLogsForReconciler provides a mock function with given fields: ctx, reconcilerName
func (_m *MockDatabase) GetAuditLogsForReconciler(ctx context.Context, reconcilerName sqlc.ReconcilerName) ([]*AuditLog, error) {
	ret := _m.Called(ctx, reconcilerName)

	var r0 []*AuditLog
	if rf, ok := ret.Get(0).(func(context.Context, sqlc.ReconcilerName) []*AuditLog); ok {
		r0 = rf(ctx, reconcilerName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*AuditLog)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, sqlc.ReconcilerName) error); ok {
		r1 = rf(ctx, reconcilerName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
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

// GetEnabledReconcilers provides a mock function with given fields: ctx
func (_m *MockDatabase) GetEnabledReconcilers(ctx context.Context) ([]*Reconciler, error) {
	ret := _m.Called(ctx)

	var r0 []*Reconciler
	if rf, ok := ret.Get(0).(func(context.Context) []*Reconciler); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*Reconciler)
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

// GetReconciler provides a mock function with given fields: ctx, reconcilerName
func (_m *MockDatabase) GetReconciler(ctx context.Context, reconcilerName sqlc.ReconcilerName) (*Reconciler, error) {
	ret := _m.Called(ctx, reconcilerName)

	var r0 *Reconciler
	if rf, ok := ret.Get(0).(func(context.Context, sqlc.ReconcilerName) *Reconciler); ok {
		r0 = rf(ctx, reconcilerName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Reconciler)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, sqlc.ReconcilerName) error); ok {
		r1 = rf(ctx, reconcilerName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetReconcilerConfig provides a mock function with given fields: ctx, reconcilerName
func (_m *MockDatabase) GetReconcilerConfig(ctx context.Context, reconcilerName sqlc.ReconcilerName) ([]*ReconcilerConfig, error) {
	ret := _m.Called(ctx, reconcilerName)

	var r0 []*ReconcilerConfig
	if rf, ok := ret.Get(0).(func(context.Context, sqlc.ReconcilerName) []*ReconcilerConfig); ok {
		r0 = rf(ctx, reconcilerName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*ReconcilerConfig)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, sqlc.ReconcilerName) error); ok {
		r1 = rf(ctx, reconcilerName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetReconcilers provides a mock function with given fields: ctx
func (_m *MockDatabase) GetReconcilers(ctx context.Context) ([]*Reconciler, error) {
	ret := _m.Called(ctx)

	var r0 []*Reconciler
	if rf, ok := ret.Get(0).(func(context.Context) []*Reconciler); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*Reconciler)
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

// GetServiceAccountByApiKey provides a mock function with given fields: ctx, APIKey
func (_m *MockDatabase) GetServiceAccountByApiKey(ctx context.Context, APIKey string) (*ServiceAccount, error) {
	ret := _m.Called(ctx, APIKey)

	var r0 *ServiceAccount
	if rf, ok := ret.Get(0).(func(context.Context, string) *ServiceAccount); ok {
		r0 = rf(ctx, APIKey)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ServiceAccount)
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

// GetServiceAccountByName provides a mock function with given fields: ctx, name
func (_m *MockDatabase) GetServiceAccountByName(ctx context.Context, name string) (*ServiceAccount, error) {
	ret := _m.Called(ctx, name)

	var r0 *ServiceAccount
	if rf, ok := ret.Get(0).(func(context.Context, string) *ServiceAccount); ok {
		r0 = rf(ctx, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ServiceAccount)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetServiceAccountRoles provides a mock function with given fields: ctx, serviceAccountID
func (_m *MockDatabase) GetServiceAccountRoles(ctx context.Context, serviceAccountID uuid.UUID) ([]*Role, error) {
	ret := _m.Called(ctx, serviceAccountID)

	var r0 []*Role
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) []*Role); ok {
		r0 = rf(ctx, serviceAccountID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*Role)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, serviceAccountID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetServiceAccounts provides a mock function with given fields: ctx
func (_m *MockDatabase) GetServiceAccounts(ctx context.Context) ([]*ServiceAccount, error) {
	ret := _m.Called(ctx)

	var r0 []*ServiceAccount
	if rf, ok := ret.Get(0).(func(context.Context) []*ServiceAccount); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*ServiceAccount)
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

// GetSessionByID provides a mock function with given fields: ctx, sessionID
func (_m *MockDatabase) GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*Session, error) {
	ret := _m.Called(ctx, sessionID)

	var r0 *Session
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *Session); ok {
		r0 = rf(ctx, sessionID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Session)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, sessionID)
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

// GetTeamMetadata provides a mock function with given fields: ctx, _a1
func (_m *MockDatabase) GetTeamMetadata(ctx context.Context, _a1 slug.Slug) ([]*TeamMetadata, error) {
	ret := _m.Called(ctx, _a1)

	var r0 []*TeamMetadata
	if rf, ok := ret.Get(0).(func(context.Context, slug.Slug) []*TeamMetadata); ok {
		r0 = rf(ctx, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*TeamMetadata)
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

// GetTeamReconcilerErrors provides a mock function with given fields: ctx, _a1
func (_m *MockDatabase) GetTeamReconcilerErrors(ctx context.Context, _a1 slug.Slug) ([]*ReconcilerError, error) {
	ret := _m.Called(ctx, _a1)

	var r0 []*ReconcilerError
	if rf, ok := ret.Get(0).(func(context.Context, slug.Slug) []*ReconcilerError); ok {
		r0 = rf(ctx, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*ReconcilerError)
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

// GetUserByExternalID provides a mock function with given fields: ctx, externalID
func (_m *MockDatabase) GetUserByExternalID(ctx context.Context, externalID string) (*User, error) {
	ret := _m.Called(ctx, externalID)

	var r0 *User
	if rf, ok := ret.Get(0).(func(context.Context, string) *User); ok {
		r0 = rf(ctx, externalID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, externalID)
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

// LoadReconcilerStateForTeam provides a mock function with given fields: ctx, reconcilerName, _a2, state
func (_m *MockDatabase) LoadReconcilerStateForTeam(ctx context.Context, reconcilerName sqlc.ReconcilerName, _a2 slug.Slug, state interface{}) error {
	ret := _m.Called(ctx, reconcilerName, _a2, state)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, sqlc.ReconcilerName, slug.Slug, interface{}) error); ok {
		r0 = rf(ctx, reconcilerName, _a2, state)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveAllServiceAccountRoles provides a mock function with given fields: ctx, serviceAccountID
func (_m *MockDatabase) RemoveAllServiceAccountRoles(ctx context.Context, serviceAccountID uuid.UUID) error {
	ret := _m.Called(ctx, serviceAccountID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, serviceAccountID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveAllUserRoles provides a mock function with given fields: ctx, userID
func (_m *MockDatabase) RemoveAllUserRoles(ctx context.Context, userID uuid.UUID) error {
	ret := _m.Called(ctx, userID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, userID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveApiKeysFromServiceAccount provides a mock function with given fields: ctx, serviceAccountID
func (_m *MockDatabase) RemoveApiKeysFromServiceAccount(ctx context.Context, serviceAccountID uuid.UUID) error {
	ret := _m.Called(ctx, serviceAccountID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, serviceAccountID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveUserFromTeam provides a mock function with given fields: ctx, userID, teamID
func (_m *MockDatabase) RemoveUserFromTeam(ctx context.Context, userID uuid.UUID, teamID uuid.UUID) error {
	ret := _m.Called(ctx, userID, teamID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, uuid.UUID) error); ok {
		r0 = rf(ctx, userID, teamID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ResetReconcilerConfig provides a mock function with given fields: ctx, reconcilerName
func (_m *MockDatabase) ResetReconcilerConfig(ctx context.Context, reconcilerName sqlc.ReconcilerName) (*Reconciler, error) {
	ret := _m.Called(ctx, reconcilerName)

	var r0 *Reconciler
	if rf, ok := ret.Get(0).(func(context.Context, sqlc.ReconcilerName) *Reconciler); ok {
		r0 = rf(ctx, reconcilerName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Reconciler)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, sqlc.ReconcilerName) error); ok {
		r1 = rf(ctx, reconcilerName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RevokeGlobalRoleFromUser provides a mock function with given fields: ctx, userID, roleName
func (_m *MockDatabase) RevokeGlobalRoleFromUser(ctx context.Context, userID uuid.UUID, roleName sqlc.RoleName) error {
	ret := _m.Called(ctx, userID, roleName)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, sqlc.RoleName) error); ok {
		r0 = rf(ctx, userID, roleName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetReconcilerErrorForTeam provides a mock function with given fields: ctx, correlationID, _a2, reconcilerName, err
func (_m *MockDatabase) SetReconcilerErrorForTeam(ctx context.Context, correlationID uuid.UUID, _a2 slug.Slug, reconcilerName sqlc.ReconcilerName, err error) error {
	ret := _m.Called(ctx, correlationID, _a2, reconcilerName, err)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, slug.Slug, sqlc.ReconcilerName, error) error); ok {
		r0 = rf(ctx, correlationID, _a2, reconcilerName, err)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetReconcilerStateForTeam provides a mock function with given fields: ctx, reconcilerName, _a2, state
func (_m *MockDatabase) SetReconcilerStateForTeam(ctx context.Context, reconcilerName sqlc.ReconcilerName, _a2 slug.Slug, state interface{}) error {
	ret := _m.Called(ctx, reconcilerName, _a2, state)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, sqlc.ReconcilerName, slug.Slug, interface{}) error); ok {
		r0 = rf(ctx, reconcilerName, _a2, state)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetTeamMemberRole provides a mock function with given fields: ctx, userID, teamID, role
func (_m *MockDatabase) SetTeamMemberRole(ctx context.Context, userID uuid.UUID, teamID uuid.UUID, role sqlc.RoleName) error {
	ret := _m.Called(ctx, userID, teamID, role)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, uuid.UUID, sqlc.RoleName) error); ok {
		r0 = rf(ctx, userID, teamID, role)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetTeamMetadata provides a mock function with given fields: ctx, _a1, metadata
func (_m *MockDatabase) SetTeamMetadata(ctx context.Context, _a1 slug.Slug, metadata []TeamMetadata) error {
	ret := _m.Called(ctx, _a1, metadata)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, slug.Slug, []TeamMetadata) error); ok {
		r0 = rf(ctx, _a1, metadata)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Transaction provides a mock function with given fields: ctx, fn
func (_m *MockDatabase) Transaction(ctx context.Context, fn DatabaseTransactionFunc) error {
	ret := _m.Called(ctx, fn)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, DatabaseTransactionFunc) error); ok {
		r0 = rf(ctx, fn)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateTeam provides a mock function with given fields: ctx, teamID, purpose
func (_m *MockDatabase) UpdateTeam(ctx context.Context, teamID uuid.UUID, purpose *string) (*Team, error) {
	ret := _m.Called(ctx, teamID, purpose)

	var r0 *Team
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, *string) *Team); ok {
		r0 = rf(ctx, teamID, purpose)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Team)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, *string) error); ok {
		r1 = rf(ctx, teamID, purpose)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateUser provides a mock function with given fields: ctx, userID, name, email, externalID
func (_m *MockDatabase) UpdateUser(ctx context.Context, userID uuid.UUID, name string, email string, externalID string) (*User, error) {
	ret := _m.Called(ctx, userID, name, email, externalID)

	var r0 *User
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string, string, string) *User); ok {
		r0 = rf(ctx, userID, name, email, externalID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, string, string, string) error); ok {
		r1 = rf(ctx, userID, name, email, externalID)
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
