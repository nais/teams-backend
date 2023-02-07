// Code generated by mockery v2.14.0. DO NOT EDIT.

package deployproxy

import (
	context "context"

	slug "github.com/nais/console/pkg/slug"
	mock "github.com/stretchr/testify/mock"
)

// MockProxy is an autogenerated mock type for the Proxy type
type MockProxy struct {
	mock.Mock
}

// GetApiKey provides a mock function with given fields: ctx, _a1
func (_m *MockProxy) GetApiKey(ctx context.Context, _a1 slug.Slug) (string, error) {
	ret := _m.Called(ctx, _a1)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, slug.Slug) string); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, slug.Slug) error); ok {
		r1 = rf(ctx, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewMockProxy interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockProxy creates a new instance of MockProxy. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockProxy(t mockConstructorTestingTNewMockProxy) *MockProxy {
	mock := &MockProxy{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}