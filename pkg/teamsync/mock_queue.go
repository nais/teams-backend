// Code generated by mockery v2.14.0. DO NOT EDIT.

package teamsync

import (
	reconcilers "github.com/nais/console/pkg/reconcilers"
	mock "github.com/stretchr/testify/mock"
)

// MockQueue is an autogenerated mock type for the Queue type
type MockQueue struct {
	mock.Mock
}

// Add provides a mock function with given fields: input
func (_m *MockQueue) Add(input reconcilers.Input) error {
	ret := _m.Called(input)

	var r0 error
	if rf, ok := ret.Get(0).(func(reconcilers.Input) error); ok {
		r0 = rf(input)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Close provides a mock function with given fields:
func (_m *MockQueue) Close() {
	_m.Called()
}

type mockConstructorTestingTNewMockQueue interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockQueue creates a new instance of MockQueue. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockQueue(t mockConstructorTestingTNewMockQueue) *MockQueue {
	mock := &MockQueue{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
