// Code generated by mockery. DO NOT EDIT.

package deployproxy

import (
	context "context"

	slug "github.com/nais/teams-backend/pkg/slug"
	mock "github.com/stretchr/testify/mock"
)

// MockProxy is an autogenerated mock type for the Proxy type
type MockProxy struct {
	mock.Mock
}

type MockProxy_Expecter struct {
	mock *mock.Mock
}

func (_m *MockProxy) EXPECT() *MockProxy_Expecter {
	return &MockProxy_Expecter{mock: &_m.Mock}
}

// GetApiKey provides a mock function with given fields: ctx, _a1
func (_m *MockProxy) GetApiKey(ctx context.Context, _a1 slug.Slug) (string, error) {
	ret := _m.Called(ctx, _a1)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, slug.Slug) (string, error)); ok {
		return rf(ctx, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, slug.Slug) string); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, slug.Slug) error); ok {
		r1 = rf(ctx, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockProxy_GetApiKey_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetApiKey'
type MockProxy_GetApiKey_Call struct {
	*mock.Call
}

// GetApiKey is a helper method to define mock.On call
//   - ctx context.Context
//   - _a1 slug.Slug
func (_e *MockProxy_Expecter) GetApiKey(ctx interface{}, _a1 interface{}) *MockProxy_GetApiKey_Call {
	return &MockProxy_GetApiKey_Call{Call: _e.mock.On("GetApiKey", ctx, _a1)}
}

func (_c *MockProxy_GetApiKey_Call) Run(run func(ctx context.Context, _a1 slug.Slug)) *MockProxy_GetApiKey_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(slug.Slug))
	})
	return _c
}

func (_c *MockProxy_GetApiKey_Call) Return(_a0 string, _a1 error) *MockProxy_GetApiKey_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockProxy_GetApiKey_Call) RunAndReturn(run func(context.Context, slug.Slug) (string, error)) *MockProxy_GetApiKey_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockProxy creates a new instance of MockProxy. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockProxy(t interface {
	mock.TestingT
	Cleanup(func())
},
) *MockProxy {
	mock := &MockProxy{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
