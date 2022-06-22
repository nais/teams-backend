// Code generated by mockery v2.13.0. DO NOT EDIT.

package azureclient

import (
	context "context"

	reconcilers "github.com/nais/console/pkg/reconcilers"
	mock "github.com/stretchr/testify/mock"

	uuid "github.com/google/uuid"
)

// MockClient is an autogenerated mock type for the Client type
type MockClient struct {
	mock.Mock
}

// AddMemberToGroup provides a mock function with given fields: ctx, grp, member
func (_m *MockClient) AddMemberToGroup(ctx context.Context, grp *Group, member *Member) error {
	ret := _m.Called(ctx, grp, member)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *Group, *Member) error); ok {
		r0 = rf(ctx, grp, member)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateGroup provides a mock function with given fields: ctx, grp
func (_m *MockClient) CreateGroup(ctx context.Context, grp *Group) (*Group, error) {
	ret := _m.Called(ctx, grp)

	var r0 *Group
	if rf, ok := ret.Get(0).(func(context.Context, *Group) *Group); ok {
		r0 = rf(ctx, grp)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Group)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *Group) error); ok {
		r1 = rf(ctx, grp)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetGroupById provides a mock function with given fields: ctx, id
func (_m *MockClient) GetGroupById(ctx context.Context, id uuid.UUID) (*Group, error) {
	ret := _m.Called(ctx, id)

	var r0 *Group
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *Group); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Group)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetOrCreateGroup provides a mock function with given fields: ctx, state, slug, name, description
func (_m *MockClient) GetOrCreateGroup(ctx context.Context, state reconcilers.AzureState, slug string, name string, description *string) (*Group, bool, error) {
	ret := _m.Called(ctx, state, slug, name, description)

	var r0 *Group
	if rf, ok := ret.Get(0).(func(context.Context, reconcilers.AzureState, string, string, *string) *Group); ok {
		r0 = rf(ctx, state, slug, name, description)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Group)
		}
	}

	var r1 bool
	if rf, ok := ret.Get(1).(func(context.Context, reconcilers.AzureState, string, string, *string) bool); ok {
		r1 = rf(ctx, state, slug, name, description)
	} else {
		r1 = ret.Get(1).(bool)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(context.Context, reconcilers.AzureState, string, string, *string) error); ok {
		r2 = rf(ctx, state, slug, name, description)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// GetUser provides a mock function with given fields: ctx, email
func (_m *MockClient) GetUser(ctx context.Context, email string) (*Member, error) {
	ret := _m.Called(ctx, email)

	var r0 *Member
	if rf, ok := ret.Get(0).(func(context.Context, string) *Member); ok {
		r0 = rf(ctx, email)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Member)
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

// ListGroupMembers provides a mock function with given fields: ctx, grp
func (_m *MockClient) ListGroupMembers(ctx context.Context, grp *Group) ([]*Member, error) {
	ret := _m.Called(ctx, grp)

	var r0 []*Member
	if rf, ok := ret.Get(0).(func(context.Context, *Group) []*Member); ok {
		r0 = rf(ctx, grp)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*Member)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *Group) error); ok {
		r1 = rf(ctx, grp)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListGroupOwners provides a mock function with given fields: ctx, grp
func (_m *MockClient) ListGroupOwners(ctx context.Context, grp *Group) ([]*Owner, error) {
	ret := _m.Called(ctx, grp)

	var r0 []*Owner
	if rf, ok := ret.Get(0).(func(context.Context, *Group) []*Owner); ok {
		r0 = rf(ctx, grp)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*Owner)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *Group) error); ok {
		r1 = rf(ctx, grp)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RemoveMemberFromGroup provides a mock function with given fields: ctx, grp, member
func (_m *MockClient) RemoveMemberFromGroup(ctx context.Context, grp *Group, member *Member) error {
	ret := _m.Called(ctx, grp, member)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *Group, *Member) error); ok {
		r0 = rf(ctx, grp, member)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type NewMockClientT interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockClient creates a new instance of MockClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockClient(t NewMockClientT) *MockClient {
	mock := &MockClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
