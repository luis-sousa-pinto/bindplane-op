// Code generated by mockery v2.28.1. DO NOT EDIT.

package mocks

import (
	context "context"

	eventbus "github.com/observiq/bindplane-op/eventbus"
	mock "github.com/stretchr/testify/mock"
)

// MockSource is an autogenerated mock type for the Source type
type MockSource[T interface{}] struct {
	mock.Mock
}

// Send provides a mock function with given fields: ctx, event
func (_m *MockSource[T]) Send(ctx context.Context, event T) {
	_m.Called(ctx, event)
}

// Subscribe provides a mock function with given fields: ctx, subscriber, onUnsubscribe
func (_m *MockSource[T]) Subscribe(ctx context.Context, subscriber eventbus.Subscriber[T], onUnsubscribe func()) eventbus.UnsubscribeFunc {
	ret := _m.Called(ctx, subscriber, onUnsubscribe)

	var r0 eventbus.UnsubscribeFunc
	if rf, ok := ret.Get(0).(func(context.Context, eventbus.Subscriber[T], func()) eventbus.UnsubscribeFunc); ok {
		r0 = rf(ctx, subscriber, onUnsubscribe)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(eventbus.UnsubscribeFunc)
		}
	}

	return r0
}

// Subscribers provides a mock function with given fields:
func (_m *MockSource[T]) Subscribers() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

type mockConstructorTestingTNewMockSource interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockSource creates a new instance of MockSource. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockSource[T interface{}](t mockConstructorTestingTNewMockSource) *MockSource[T] {
	mock := &MockSource[T]{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}