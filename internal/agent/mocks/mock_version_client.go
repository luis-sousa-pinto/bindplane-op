// Code generated by mockery v2.24.0. DO NOT EDIT.

package mocks

import (
	model "github.com/observiq/bindplane-op/model"
	mock "github.com/stretchr/testify/mock"
)

// MockVersionClient is an autogenerated mock type for the VersionClient type
type MockVersionClient struct {
	mock.Mock
}

// LatestVersion provides a mock function with given fields:
func (_m *MockVersionClient) LatestVersion() (*model.AgentVersion, error) {
	ret := _m.Called()

	var r0 *model.AgentVersion
	var r1 error
	if rf, ok := ret.Get(0).(func() (*model.AgentVersion, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() *model.AgentVersion); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AgentVersion)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Version provides a mock function with given fields: version
func (_m *MockVersionClient) Version(version string) (*model.AgentVersion, error) {
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

// Versions provides a mock function with given fields:
func (_m *MockVersionClient) Versions() ([]*model.AgentVersion, error) {
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

type mockConstructorTestingTNewMockVersionClient interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockVersionClient creates a new instance of MockVersionClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockVersionClient(t mockConstructorTestingTNewMockVersionClient) *MockVersionClient {
	mock := &MockVersionClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
