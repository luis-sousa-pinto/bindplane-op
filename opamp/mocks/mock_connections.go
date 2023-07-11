// Code generated by mockery v2.31.1. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	types "github.com/open-telemetry/opamp-go/server/types"
)

// MockConnections is an autogenerated mock type for the Connections type
type MockConnections[S interface{}] struct {
	mock.Mock
}

// Connected provides a mock function with given fields: agentID
func (_m *MockConnections[S]) Connected(agentID string) bool {
	ret := _m.Called(agentID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(agentID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// ConnectedAgentIDs provides a mock function with given fields: _a0
func (_m *MockConnections[S]) ConnectedAgentIDs(_a0 context.Context) []string {
	ret := _m.Called(_a0)

	var r0 []string
	if rf, ok := ret.Get(0).(func(context.Context) []string); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// ConnectedAgentsCount provides a mock function with given fields: _a0
func (_m *MockConnections[S]) ConnectedAgentsCount(_a0 context.Context) int {
	ret := _m.Called(_a0)

	var r0 int
	if rf, ok := ret.Get(0).(func(context.Context) int); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// OnConnecting provides a mock function with given fields: ctx, agentID
func (_m *MockConnections[S]) OnConnecting(ctx context.Context, agentID string) S {
	ret := _m.Called(ctx, agentID)

	var r0 S
	if rf, ok := ret.Get(0).(func(context.Context, string) S); ok {
		r0 = rf(ctx, agentID)
	} else {
		r0 = ret.Get(0).(S)
	}

	return r0
}

// OnConnectionClose provides a mock function with given fields: conn
func (_m *MockConnections[S]) OnConnectionClose(conn types.Connection) (S, int) {
	ret := _m.Called(conn)

	var r0 S
	var r1 int
	if rf, ok := ret.Get(0).(func(types.Connection) (S, int)); ok {
		return rf(conn)
	}
	if rf, ok := ret.Get(0).(func(types.Connection) S); ok {
		r0 = rf(conn)
	} else {
		r0 = ret.Get(0).(S)
	}

	if rf, ok := ret.Get(1).(func(types.Connection) int); ok {
		r1 = rf(conn)
	} else {
		r1 = ret.Get(1).(int)
	}

	return r0, r1
}

// OnMessage provides a mock function with given fields: agentID, conn
func (_m *MockConnections[S]) OnMessage(agentID string, conn types.Connection) (S, error) {
	ret := _m.Called(agentID, conn)

	var r0 S
	var r1 error
	if rf, ok := ret.Get(0).(func(string, types.Connection) (S, error)); ok {
		return rf(agentID, conn)
	}
	if rf, ok := ret.Get(0).(func(string, types.Connection) S); ok {
		r0 = rf(agentID, conn)
	} else {
		r0 = ret.Get(0).(S)
	}

	if rf, ok := ret.Get(1).(func(string, types.Connection) error); ok {
		r1 = rf(agentID, conn)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StateForAgentID provides a mock function with given fields: agentID
func (_m *MockConnections[S]) StateForAgentID(agentID string) S {
	ret := _m.Called(agentID)

	var r0 S
	if rf, ok := ret.Get(0).(func(string) S); ok {
		r0 = rf(agentID)
	} else {
		r0 = ret.Get(0).(S)
	}

	return r0
}

// StateForConnection provides a mock function with given fields: _a0
func (_m *MockConnections[S]) StateForConnection(_a0 types.Connection) S {
	ret := _m.Called(_a0)

	var r0 S
	if rf, ok := ret.Get(0).(func(types.Connection) S); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(S)
	}

	return r0
}

// NewMockConnections creates a new instance of MockConnections. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockConnections[S interface{}](t interface {
	mock.TestingT
	Cleanup(func())
}) *MockConnections[S] {
	mock := &MockConnections[S]{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
