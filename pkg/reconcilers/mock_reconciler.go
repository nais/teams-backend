// Code generated by mockery. DO NOT EDIT.

package reconcilers

import (
	context "context"

	slug "github.com/nais/teams-backend/pkg/slug"
	mock "github.com/stretchr/testify/mock"

	sqlc "github.com/nais/teams-backend/pkg/sqlc"

	uuid "github.com/google/uuid"
)

// MockReconciler is an autogenerated mock type for the Reconciler type
type MockReconciler struct {
	mock.Mock
}

type MockReconciler_Expecter struct {
	mock *mock.Mock
}

func (_m *MockReconciler) EXPECT() *MockReconciler_Expecter {
	return &MockReconciler_Expecter{mock: &_m.Mock}
}

// Delete provides a mock function with given fields: ctx, teamSlug, correlationID
func (_m *MockReconciler) Delete(ctx context.Context, teamSlug slug.Slug, correlationID uuid.UUID) error {
	ret := _m.Called(ctx, teamSlug, correlationID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, slug.Slug, uuid.UUID) error); ok {
		r0 = rf(ctx, teamSlug, correlationID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockReconciler_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type MockReconciler_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - ctx context.Context
//   - teamSlug slug.Slug
//   - correlationID uuid.UUID
func (_e *MockReconciler_Expecter) Delete(ctx interface{}, teamSlug interface{}, correlationID interface{}) *MockReconciler_Delete_Call {
	return &MockReconciler_Delete_Call{Call: _e.mock.On("Delete", ctx, teamSlug, correlationID)}
}

func (_c *MockReconciler_Delete_Call) Run(run func(ctx context.Context, teamSlug slug.Slug, correlationID uuid.UUID)) *MockReconciler_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(slug.Slug), args[2].(uuid.UUID))
	})
	return _c
}

func (_c *MockReconciler_Delete_Call) Return(_a0 error) *MockReconciler_Delete_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockReconciler_Delete_Call) RunAndReturn(run func(context.Context, slug.Slug, uuid.UUID) error) *MockReconciler_Delete_Call {
	_c.Call.Return(run)
	return _c
}

// Name provides a mock function with given fields:
func (_m *MockReconciler) Name() sqlc.ReconcilerName {
	ret := _m.Called()

	var r0 sqlc.ReconcilerName
	if rf, ok := ret.Get(0).(func() sqlc.ReconcilerName); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(sqlc.ReconcilerName)
	}

	return r0
}

// MockReconciler_Name_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Name'
type MockReconciler_Name_Call struct {
	*mock.Call
}

// Name is a helper method to define mock.On call
func (_e *MockReconciler_Expecter) Name() *MockReconciler_Name_Call {
	return &MockReconciler_Name_Call{Call: _e.mock.On("Name")}
}

func (_c *MockReconciler_Name_Call) Run(run func()) *MockReconciler_Name_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockReconciler_Name_Call) Return(_a0 sqlc.ReconcilerName) *MockReconciler_Name_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockReconciler_Name_Call) RunAndReturn(run func() sqlc.ReconcilerName) *MockReconciler_Name_Call {
	_c.Call.Return(run)
	return _c
}

// Reconcile provides a mock function with given fields: ctx, input
func (_m *MockReconciler) Reconcile(ctx context.Context, input Input) error {
	ret := _m.Called(ctx, input)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, Input) error); ok {
		r0 = rf(ctx, input)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockReconciler_Reconcile_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Reconcile'
type MockReconciler_Reconcile_Call struct {
	*mock.Call
}

// Reconcile is a helper method to define mock.On call
//   - ctx context.Context
//   - input Input
func (_e *MockReconciler_Expecter) Reconcile(ctx interface{}, input interface{}) *MockReconciler_Reconcile_Call {
	return &MockReconciler_Reconcile_Call{Call: _e.mock.On("Reconcile", ctx, input)}
}

func (_c *MockReconciler_Reconcile_Call) Run(run func(ctx context.Context, input Input)) *MockReconciler_Reconcile_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(Input))
	})
	return _c
}

func (_c *MockReconciler_Reconcile_Call) Return(_a0 error) *MockReconciler_Reconcile_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockReconciler_Reconcile_Call) RunAndReturn(run func(context.Context, Input) error) *MockReconciler_Reconcile_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockReconciler creates a new instance of MockReconciler. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockReconciler(t interface {
	mock.TestingT
	Cleanup(func())
},
) *MockReconciler {
	mock := &MockReconciler{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
