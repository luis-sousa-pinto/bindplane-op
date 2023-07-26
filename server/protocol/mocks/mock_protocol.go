// Code generated by mockery v2.31.1. DO NOT EDIT.

package mocks

import (
	context "context"

	model "github.com/observiq/bindplane-op/model"
	mock "github.com/stretchr/testify/mock"

	protocol "github.com/observiq/bindplane-op/server/protocol"
)

// MockProtocol is an autogenerated mock type for the Protocol type
type MockProtocol struct {
	mock.Mock
}

type MockProtocol_Expecter struct {
	mock *mock.Mock
}

func (_m *MockProtocol) EXPECT() *MockProtocol_Expecter {
	return &MockProtocol_Expecter{mock: &_m.Mock}
}

// Connected provides a mock function with given fields: agentID
func (_m *MockProtocol) Connected(agentID string) bool {
	ret := _m.Called(agentID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(agentID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// MockProtocol_Connected_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Connected'
type MockProtocol_Connected_Call struct {
	*mock.Call
}

// Connected is a helper method to define mock.On call
//   - agentID string
func (_e *MockProtocol_Expecter) Connected(agentID interface{}) *MockProtocol_Connected_Call {
	return &MockProtocol_Connected_Call{Call: _e.mock.On("Connected", agentID)}
}

func (_c *MockProtocol_Connected_Call) Run(run func(agentID string)) *MockProtocol_Connected_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockProtocol_Connected_Call) Return(_a0 bool) *MockProtocol_Connected_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockProtocol_Connected_Call) RunAndReturn(run func(string) bool) *MockProtocol_Connected_Call {
	_c.Call.Return(run)
	return _c
}

// ConnectedAgentIDs provides a mock function with given fields: _a0
func (_m *MockProtocol) ConnectedAgentIDs(_a0 context.Context) ([]string, error) {
	ret := _m.Called(_a0)

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) ([]string, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(context.Context) []string); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockProtocol_ConnectedAgentIDs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ConnectedAgentIDs'
type MockProtocol_ConnectedAgentIDs_Call struct {
	*mock.Call
}

// ConnectedAgentIDs is a helper method to define mock.On call
//   - _a0 context.Context
func (_e *MockProtocol_Expecter) ConnectedAgentIDs(_a0 interface{}) *MockProtocol_ConnectedAgentIDs_Call {
	return &MockProtocol_ConnectedAgentIDs_Call{Call: _e.mock.On("ConnectedAgentIDs", _a0)}
}

func (_c *MockProtocol_ConnectedAgentIDs_Call) Run(run func(_a0 context.Context)) *MockProtocol_ConnectedAgentIDs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockProtocol_ConnectedAgentIDs_Call) Return(_a0 []string, _a1 error) *MockProtocol_ConnectedAgentIDs_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockProtocol_ConnectedAgentIDs_Call) RunAndReturn(run func(context.Context) ([]string, error)) *MockProtocol_ConnectedAgentIDs_Call {
	_c.Call.Return(run)
	return _c
}

// Disconnect provides a mock function with given fields: agentID
func (_m *MockProtocol) Disconnect(agentID string) bool {
	ret := _m.Called(agentID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(agentID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// MockProtocol_Disconnect_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Disconnect'
type MockProtocol_Disconnect_Call struct {
	*mock.Call
}

// Disconnect is a helper method to define mock.On call
//   - agentID string
func (_e *MockProtocol_Expecter) Disconnect(agentID interface{}) *MockProtocol_Disconnect_Call {
	return &MockProtocol_Disconnect_Call{Call: _e.mock.On("Disconnect", agentID)}
}

func (_c *MockProtocol_Disconnect_Call) Run(run func(agentID string)) *MockProtocol_Disconnect_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockProtocol_Disconnect_Call) Return(_a0 bool) *MockProtocol_Disconnect_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockProtocol_Disconnect_Call) RunAndReturn(run func(string) bool) *MockProtocol_Disconnect_Call {
	_c.Call.Return(run)
	return _c
}

// Name provides a mock function with given fields:
func (_m *MockProtocol) Name() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockProtocol_Name_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Name'
type MockProtocol_Name_Call struct {
	*mock.Call
}

// Name is a helper method to define mock.On call
func (_e *MockProtocol_Expecter) Name() *MockProtocol_Name_Call {
	return &MockProtocol_Name_Call{Call: _e.mock.On("Name")}
}

func (_c *MockProtocol_Name_Call) Run(run func()) *MockProtocol_Name_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockProtocol_Name_Call) Return(_a0 string) *MockProtocol_Name_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockProtocol_Name_Call) RunAndReturn(run func() string) *MockProtocol_Name_Call {
	_c.Call.Return(run)
	return _c
}

// RequestReport provides a mock function with given fields: ctx, agentID, report
func (_m *MockProtocol) RequestReport(ctx context.Context, agentID string, report protocol.Report) error {
	ret := _m.Called(ctx, agentID, report)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, protocol.Report) error); ok {
		r0 = rf(ctx, agentID, report)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockProtocol_RequestReport_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RequestReport'
type MockProtocol_RequestReport_Call struct {
	*mock.Call
}

// RequestReport is a helper method to define mock.On call
//   - ctx context.Context
//   - agentID string
//   - report protocol.Report
func (_e *MockProtocol_Expecter) RequestReport(ctx interface{}, agentID interface{}, report interface{}) *MockProtocol_RequestReport_Call {
	return &MockProtocol_RequestReport_Call{Call: _e.mock.On("RequestReport", ctx, agentID, report)}
}

func (_c *MockProtocol_RequestReport_Call) Run(run func(ctx context.Context, agentID string, report protocol.Report)) *MockProtocol_RequestReport_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(protocol.Report))
	})
	return _c
}

func (_c *MockProtocol_RequestReport_Call) Return(_a0 error) *MockProtocol_RequestReport_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockProtocol_RequestReport_Call) RunAndReturn(run func(context.Context, string, protocol.Report) error) *MockProtocol_RequestReport_Call {
	_c.Call.Return(run)
	return _c
}

// SendHeartbeat provides a mock function with given fields: agentID
func (_m *MockProtocol) SendHeartbeat(agentID string) error {
	ret := _m.Called(agentID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(agentID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockProtocol_SendHeartbeat_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SendHeartbeat'
type MockProtocol_SendHeartbeat_Call struct {
	*mock.Call
}

// SendHeartbeat is a helper method to define mock.On call
//   - agentID string
func (_e *MockProtocol_Expecter) SendHeartbeat(agentID interface{}) *MockProtocol_SendHeartbeat_Call {
	return &MockProtocol_SendHeartbeat_Call{Call: _e.mock.On("SendHeartbeat", agentID)}
}

func (_c *MockProtocol_SendHeartbeat_Call) Run(run func(agentID string)) *MockProtocol_SendHeartbeat_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockProtocol_SendHeartbeat_Call) Return(_a0 error) *MockProtocol_SendHeartbeat_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockProtocol_SendHeartbeat_Call) RunAndReturn(run func(string) error) *MockProtocol_SendHeartbeat_Call {
	_c.Call.Return(run)
	return _c
}

// Shutdown provides a mock function with given fields: ctx
func (_m *MockProtocol) Shutdown(ctx context.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockProtocol_Shutdown_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Shutdown'
type MockProtocol_Shutdown_Call struct {
	*mock.Call
}

// Shutdown is a helper method to define mock.On call
//   - ctx context.Context
func (_e *MockProtocol_Expecter) Shutdown(ctx interface{}) *MockProtocol_Shutdown_Call {
	return &MockProtocol_Shutdown_Call{Call: _e.mock.On("Shutdown", ctx)}
}

func (_c *MockProtocol_Shutdown_Call) Run(run func(ctx context.Context)) *MockProtocol_Shutdown_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockProtocol_Shutdown_Call) Return(_a0 error) *MockProtocol_Shutdown_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockProtocol_Shutdown_Call) RunAndReturn(run func(context.Context) error) *MockProtocol_Shutdown_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateAgent provides a mock function with given fields: _a0, _a1, _a2
func (_m *MockProtocol) UpdateAgent(_a0 context.Context, _a1 *model.Agent, _a2 *protocol.AgentUpdates) error {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.Agent, *protocol.AgentUpdates) error); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockProtocol_UpdateAgent_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateAgent'
type MockProtocol_UpdateAgent_Call struct {
	*mock.Call
}

// UpdateAgent is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 *model.Agent
//   - _a2 *protocol.AgentUpdates
func (_e *MockProtocol_Expecter) UpdateAgent(_a0 interface{}, _a1 interface{}, _a2 interface{}) *MockProtocol_UpdateAgent_Call {
	return &MockProtocol_UpdateAgent_Call{Call: _e.mock.On("UpdateAgent", _a0, _a1, _a2)}
}

func (_c *MockProtocol_UpdateAgent_Call) Run(run func(_a0 context.Context, _a1 *model.Agent, _a2 *protocol.AgentUpdates)) *MockProtocol_UpdateAgent_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*model.Agent), args[2].(*protocol.AgentUpdates))
	})
	return _c
}

func (_c *MockProtocol_UpdateAgent_Call) Return(_a0 error) *MockProtocol_UpdateAgent_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockProtocol_UpdateAgent_Call) RunAndReturn(run func(context.Context, *model.Agent, *protocol.AgentUpdates) error) *MockProtocol_UpdateAgent_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockProtocol creates a new instance of MockProtocol. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockProtocol(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockProtocol {
	mock := &MockProtocol{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
