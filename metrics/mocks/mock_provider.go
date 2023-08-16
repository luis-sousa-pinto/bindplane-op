// Code generated by mockery v2.31.1. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MockProvider is an autogenerated mock type for the Provider type
type MockProvider struct {
	mock.Mock
}

type MockProvider_Expecter struct {
	mock *mock.Mock
}

func (_m *MockProvider) EXPECT() *MockProvider_Expecter {
	return &MockProvider_Expecter{mock: &_m.Mock}
}

// Shutdown provides a mock function with given fields: _a0
func (_m *MockProvider) Shutdown(_a0 context.Context) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockProvider_Shutdown_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Shutdown'
type MockProvider_Shutdown_Call struct {
	*mock.Call
}

// Shutdown is a helper method to define mock.On call
//   - _a0 context.Context
func (_e *MockProvider_Expecter) Shutdown(_a0 interface{}) *MockProvider_Shutdown_Call {
	return &MockProvider_Shutdown_Call{Call: _e.mock.On("Shutdown", _a0)}
}

func (_c *MockProvider_Shutdown_Call) Run(run func(_a0 context.Context)) *MockProvider_Shutdown_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockProvider_Shutdown_Call) Return(_a0 error) *MockProvider_Shutdown_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockProvider_Shutdown_Call) RunAndReturn(run func(context.Context) error) *MockProvider_Shutdown_Call {
	_c.Call.Return(run)
	return _c
}

// Start provides a mock function with given fields: _a0
func (_m *MockProvider) Start(_a0 context.Context) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockProvider_Start_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Start'
type MockProvider_Start_Call struct {
	*mock.Call
}

// Start is a helper method to define mock.On call
//   - _a0 context.Context
func (_e *MockProvider_Expecter) Start(_a0 interface{}) *MockProvider_Start_Call {
	return &MockProvider_Start_Call{Call: _e.mock.On("Start", _a0)}
}

func (_c *MockProvider_Start_Call) Run(run func(_a0 context.Context)) *MockProvider_Start_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockProvider_Start_Call) Return(_a0 error) *MockProvider_Start_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockProvider_Start_Call) RunAndReturn(run func(context.Context) error) *MockProvider_Start_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockProvider creates a new instance of MockProvider. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockProvider(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockProvider {
	mock := &MockProvider{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
