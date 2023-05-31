// Code generated by mockery v2.28.1. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MockTracer is an autogenerated mock type for the Tracer type
type MockTracer struct {
	mock.Mock
}

// Shutdown provides a mock function with given fields: ctx
func (_m *MockTracer) Shutdown(ctx context.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Start provides a mock function with given fields: ctx
func (_m *MockTracer) Start(ctx context.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewMockTracer interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockTracer creates a new instance of MockTracer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockTracer(t mockConstructorTestingTNewMockTracer) *MockTracer {
	mock := &MockTracer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
