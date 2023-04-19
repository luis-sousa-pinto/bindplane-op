// Code generated by mockery v2.25.1. DO NOT EDIT.

package mocks

import (
	context "context"
	net "net"

	mock "github.com/stretchr/testify/mock"

	protobufs "github.com/open-telemetry/opamp-go/protobufs"
)

// MockConnection is an autogenerated mock type for the Connection type
type MockConnection struct {
	mock.Mock
}

// RemoteAddr provides a mock function with given fields:
func (_m *MockConnection) RemoteAddr() net.Addr {
	ret := _m.Called()

	var r0 net.Addr
	if rf, ok := ret.Get(0).(func() net.Addr); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(net.Addr)
		}
	}

	return r0
}

// Send provides a mock function with given fields: ctx, message
func (_m *MockConnection) Send(ctx context.Context, message *protobufs.ServerToAgent) error {
	ret := _m.Called(ctx, message)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *protobufs.ServerToAgent) error); ok {
		r0 = rf(ctx, message)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewMockConnection interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockConnection creates a new instance of MockConnection. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockConnection(t mockConstructorTestingTNewMockConnection) *MockConnection {
	mock := &MockConnection{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
