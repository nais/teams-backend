// Code generated by mockery v2.14.0. DO NOT EDIT.

package teamsync

import (
	context "context"

	db "github.com/nais/console/pkg/db"
	mock "github.com/stretchr/testify/mock"

	slug "github.com/nais/console/pkg/slug"

	sqlc "github.com/nais/console/pkg/sqlc"

	uuid "github.com/google/uuid"
)

// MockHandler is an autogenerated mock type for the Handler type
type MockHandler struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *MockHandler) Close() {
	_m.Called()
}

// DeleteTeam provides a mock function with given fields: teamSlug, correlationID
func (_m *MockHandler) DeleteTeam(teamSlug slug.Slug, correlationID uuid.UUID) error {
	ret := _m.Called(teamSlug, correlationID)

	var r0 error
	if rf, ok := ret.Get(0).(func(slug.Slug, uuid.UUID) error); ok {
		r0 = rf(teamSlug, correlationID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// InitReconcilers provides a mock function with given fields: ctx
func (_m *MockHandler) InitReconcilers(ctx context.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveReconciler provides a mock function with given fields: reconcilerName
func (_m *MockHandler) RemoveReconciler(reconcilerName sqlc.ReconcilerName) {
	_m.Called(reconcilerName)
}

// Schedule provides a mock function with given fields: input
func (_m *MockHandler) Schedule(input Input) error {
	ret := _m.Called(input)

	var r0 error
	if rf, ok := ret.Get(0).(func(Input) error); ok {
		r0 = rf(input)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ScheduleAllTeams provides a mock function with given fields: ctx, correlationID
func (_m *MockHandler) ScheduleAllTeams(ctx context.Context, correlationID uuid.UUID) ([]*db.Team, error) {
	ret := _m.Called(ctx, correlationID)

	var r0 []*db.Team
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) []*db.Team); ok {
		r0 = rf(ctx, correlationID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*db.Team)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, correlationID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetReconcilerFactories provides a mock function with given fields: factories
func (_m *MockHandler) SetReconcilerFactories(factories ReconcilerFactories) {
	_m.Called(factories)
}

// SyncTeams provides a mock function with given fields: ctx
func (_m *MockHandler) SyncTeams(ctx context.Context) {
	_m.Called(ctx)
}

// UpdateMetrics provides a mock function with given fields: ctx
func (_m *MockHandler) UpdateMetrics(ctx context.Context) {
	_m.Called(ctx)
}

// UseReconciler provides a mock function with given fields: reconciler
func (_m *MockHandler) UseReconciler(reconciler db.Reconciler) error {
	ret := _m.Called(reconciler)

	var r0 error
	if rf, ok := ret.Get(0).(func(db.Reconciler) error); ok {
		r0 = rf(reconciler)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewMockHandler interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockHandler creates a new instance of MockHandler. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockHandler(t mockConstructorTestingTNewMockHandler) *MockHandler {
	mock := &MockHandler{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
