// Code generated by mockery v2.31.1. DO NOT EDIT.

package model

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MockResourceStore is an autogenerated mock type for the ResourceStore type
type MockResourceStore struct {
	mock.Mock
}

type MockResourceStore_Expecter struct {
	mock *mock.Mock
}

func (_m *MockResourceStore) EXPECT() *MockResourceStore_Expecter {
	return &MockResourceStore_Expecter{mock: &_m.Mock}
}

// Destination provides a mock function with given fields: ctx, name
func (_m *MockResourceStore) Destination(ctx context.Context, name string) (*Destination, error) {
	ret := _m.Called(ctx, name)

	var r0 *Destination
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*Destination, error)); ok {
		return rf(ctx, name)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *Destination); ok {
		r0 = rf(ctx, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Destination)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockResourceStore_Destination_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Destination'
type MockResourceStore_Destination_Call struct {
	*mock.Call
}

// Destination is a helper method to define mock.On call
//   - ctx context.Context
//   - name string
func (_e *MockResourceStore_Expecter) Destination(ctx interface{}, name interface{}) *MockResourceStore_Destination_Call {
	return &MockResourceStore_Destination_Call{Call: _e.mock.On("Destination", ctx, name)}
}

func (_c *MockResourceStore_Destination_Call) Run(run func(ctx context.Context, name string)) *MockResourceStore_Destination_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockResourceStore_Destination_Call) Return(_a0 *Destination, _a1 error) *MockResourceStore_Destination_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockResourceStore_Destination_Call) RunAndReturn(run func(context.Context, string) (*Destination, error)) *MockResourceStore_Destination_Call {
	_c.Call.Return(run)
	return _c
}

// DestinationType provides a mock function with given fields: ctx, name
func (_m *MockResourceStore) DestinationType(ctx context.Context, name string) (*DestinationType, error) {
	ret := _m.Called(ctx, name)

	var r0 *DestinationType
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*DestinationType, error)); ok {
		return rf(ctx, name)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *DestinationType); ok {
		r0 = rf(ctx, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*DestinationType)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockResourceStore_DestinationType_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DestinationType'
type MockResourceStore_DestinationType_Call struct {
	*mock.Call
}

// DestinationType is a helper method to define mock.On call
//   - ctx context.Context
//   - name string
func (_e *MockResourceStore_Expecter) DestinationType(ctx interface{}, name interface{}) *MockResourceStore_DestinationType_Call {
	return &MockResourceStore_DestinationType_Call{Call: _e.mock.On("DestinationType", ctx, name)}
}

func (_c *MockResourceStore_DestinationType_Call) Run(run func(ctx context.Context, name string)) *MockResourceStore_DestinationType_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockResourceStore_DestinationType_Call) Return(_a0 *DestinationType, _a1 error) *MockResourceStore_DestinationType_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockResourceStore_DestinationType_Call) RunAndReturn(run func(context.Context, string) (*DestinationType, error)) *MockResourceStore_DestinationType_Call {
	_c.Call.Return(run)
	return _c
}

// Processor provides a mock function with given fields: ctx, name
func (_m *MockResourceStore) Processor(ctx context.Context, name string) (*Processor, error) {
	ret := _m.Called(ctx, name)

	var r0 *Processor
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*Processor, error)); ok {
		return rf(ctx, name)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *Processor); ok {
		r0 = rf(ctx, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Processor)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockResourceStore_Processor_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Processor'
type MockResourceStore_Processor_Call struct {
	*mock.Call
}

// Processor is a helper method to define mock.On call
//   - ctx context.Context
//   - name string
func (_e *MockResourceStore_Expecter) Processor(ctx interface{}, name interface{}) *MockResourceStore_Processor_Call {
	return &MockResourceStore_Processor_Call{Call: _e.mock.On("Processor", ctx, name)}
}

func (_c *MockResourceStore_Processor_Call) Run(run func(ctx context.Context, name string)) *MockResourceStore_Processor_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockResourceStore_Processor_Call) Return(_a0 *Processor, _a1 error) *MockResourceStore_Processor_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockResourceStore_Processor_Call) RunAndReturn(run func(context.Context, string) (*Processor, error)) *MockResourceStore_Processor_Call {
	_c.Call.Return(run)
	return _c
}

// ProcessorType provides a mock function with given fields: ctx, name
func (_m *MockResourceStore) ProcessorType(ctx context.Context, name string) (*ProcessorType, error) {
	ret := _m.Called(ctx, name)

	var r0 *ProcessorType
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*ProcessorType, error)); ok {
		return rf(ctx, name)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *ProcessorType); ok {
		r0 = rf(ctx, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ProcessorType)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockResourceStore_ProcessorType_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ProcessorType'
type MockResourceStore_ProcessorType_Call struct {
	*mock.Call
}

// ProcessorType is a helper method to define mock.On call
//   - ctx context.Context
//   - name string
func (_e *MockResourceStore_Expecter) ProcessorType(ctx interface{}, name interface{}) *MockResourceStore_ProcessorType_Call {
	return &MockResourceStore_ProcessorType_Call{Call: _e.mock.On("ProcessorType", ctx, name)}
}

func (_c *MockResourceStore_ProcessorType_Call) Run(run func(ctx context.Context, name string)) *MockResourceStore_ProcessorType_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockResourceStore_ProcessorType_Call) Return(_a0 *ProcessorType, _a1 error) *MockResourceStore_ProcessorType_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockResourceStore_ProcessorType_Call) RunAndReturn(run func(context.Context, string) (*ProcessorType, error)) *MockResourceStore_ProcessorType_Call {
	_c.Call.Return(run)
	return _c
}

// Source provides a mock function with given fields: ctx, name
func (_m *MockResourceStore) Source(ctx context.Context, name string) (*Source, error) {
	ret := _m.Called(ctx, name)

	var r0 *Source
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*Source, error)); ok {
		return rf(ctx, name)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *Source); ok {
		r0 = rf(ctx, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Source)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockResourceStore_Source_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Source'
type MockResourceStore_Source_Call struct {
	*mock.Call
}

// Source is a helper method to define mock.On call
//   - ctx context.Context
//   - name string
func (_e *MockResourceStore_Expecter) Source(ctx interface{}, name interface{}) *MockResourceStore_Source_Call {
	return &MockResourceStore_Source_Call{Call: _e.mock.On("Source", ctx, name)}
}

func (_c *MockResourceStore_Source_Call) Run(run func(ctx context.Context, name string)) *MockResourceStore_Source_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockResourceStore_Source_Call) Return(_a0 *Source, _a1 error) *MockResourceStore_Source_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockResourceStore_Source_Call) RunAndReturn(run func(context.Context, string) (*Source, error)) *MockResourceStore_Source_Call {
	_c.Call.Return(run)
	return _c
}

// SourceType provides a mock function with given fields: ctx, name
func (_m *MockResourceStore) SourceType(ctx context.Context, name string) (*SourceType, error) {
	ret := _m.Called(ctx, name)

	var r0 *SourceType
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*SourceType, error)); ok {
		return rf(ctx, name)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *SourceType); ok {
		r0 = rf(ctx, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*SourceType)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockResourceStore_SourceType_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SourceType'
type MockResourceStore_SourceType_Call struct {
	*mock.Call
}

// SourceType is a helper method to define mock.On call
//   - ctx context.Context
//   - name string
func (_e *MockResourceStore_Expecter) SourceType(ctx interface{}, name interface{}) *MockResourceStore_SourceType_Call {
	return &MockResourceStore_SourceType_Call{Call: _e.mock.On("SourceType", ctx, name)}
}

func (_c *MockResourceStore_SourceType_Call) Run(run func(ctx context.Context, name string)) *MockResourceStore_SourceType_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockResourceStore_SourceType_Call) Return(_a0 *SourceType, _a1 error) *MockResourceStore_SourceType_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockResourceStore_SourceType_Call) RunAndReturn(run func(context.Context, string) (*SourceType, error)) *MockResourceStore_SourceType_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockResourceStore creates a new instance of MockResourceStore. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockResourceStore(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockResourceStore {
	mock := &MockResourceStore{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
