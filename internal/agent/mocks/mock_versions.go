// Code generated by mockery v2.21.6. DO NOT EDIT.

package mocks

import (
	context "context"

	model "github.com/observiq/bindplane-op/model"
	mock "github.com/stretchr/testify/mock"
)

// MockVersions is an autogenerated mock type for the Versions type
type MockVersions struct {
	mock.Mock
}

// LatestVersion provides a mock function with given fields: ctx
func (_m *MockVersions) LatestVersion(ctx context.Context) (*model.AgentVersion, error) {
	ret := _m.Called(ctx)

	var r0 *model.AgentVersion
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*model.AgentVersion, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *model.AgentVersion); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AgentVersion)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LatestVersionString provides a mock function with given fields: ctx
func (_m *MockVersions) LatestVersionString(ctx context.Context) string {
	ret := _m.Called(ctx)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context) string); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// SyncVersion provides a mock function with given fields: version
func (_m *MockVersions) SyncVersion(version string) (*model.AgentVersion, error) {
	ret := _m.Called(version)

	var r0 *model.AgentVersion
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (*model.AgentVersion, error)); ok {
		return rf(version)
	}
	if rf, ok := ret.Get(0).(func(string) *model.AgentVersion); ok {
		r0 = rf(version)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AgentVersion)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(version)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SyncVersions provides a mock function with given fields:
func (_m *MockVersions) SyncVersions() ([]*model.AgentVersion, error) {
	ret := _m.Called()

	var r0 []*model.AgentVersion
	var r1 error
	if rf, ok := ret.Get(0).(func() ([]*model.AgentVersion, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() []*model.AgentVersion); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.AgentVersion)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Version provides a mock function with given fields: ctx, version
func (_m *MockVersions) Version(ctx context.Context, version string) (*model.AgentVersion, error) {
	ret := _m.Called(ctx, version)

	var r0 *model.AgentVersion
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*model.AgentVersion, error)); ok {
		return rf(ctx, version)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.AgentVersion); ok {
		r0 = rf(ctx, version)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AgentVersion)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, version)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewMockVersions interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockVersions creates a new instance of MockVersions. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockVersions(t mockConstructorTestingTNewMockVersions) *MockVersions {
	mock := &MockVersions{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
