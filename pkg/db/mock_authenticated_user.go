// Code generated by mockery v2.14.0. DO NOT EDIT.

package db

import (
	uuid "github.com/google/uuid"
	mock "github.com/stretchr/testify/mock"
)

// MockAuthenticatedUser is an autogenerated mock type for the AuthenticatedUser type
type MockAuthenticatedUser struct {
	mock.Mock
}

// GetID provides a mock function with given fields:
func (_m *MockAuthenticatedUser) GetID() uuid.UUID {
	ret := _m.Called()

	var r0 uuid.UUID
	if rf, ok := ret.Get(0).(func() uuid.UUID); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(uuid.UUID)
		}
	}

	return r0
}

// Identity provides a mock function with given fields:
func (_m *MockAuthenticatedUser) Identity() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// IsServiceAccount provides a mock function with given fields:
func (_m *MockAuthenticatedUser) IsServiceAccount() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

type mockConstructorTestingTNewMockAuthenticatedUser interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockAuthenticatedUser creates a new instance of MockAuthenticatedUser. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockAuthenticatedUser(t mockConstructorTestingTNewMockAuthenticatedUser) *MockAuthenticatedUser {
	mock := &MockAuthenticatedUser{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
