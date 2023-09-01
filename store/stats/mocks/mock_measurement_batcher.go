// Code generated by mockery v2.31.1. DO NOT EDIT.

package mocks

import (
	context "context"

	record "github.com/observiq/bindplane-op/otlp/record"
	mock "github.com/stretchr/testify/mock"
)

// MockMeasurementBatcher is an autogenerated mock type for the MeasurementBatcher type
type MockMeasurementBatcher struct {
	mock.Mock
}

// AcceptMetrics provides a mock function with given fields: ctx, metrics
func (_m *MockMeasurementBatcher) AcceptMetrics(ctx context.Context, metrics []*record.Metric) error {
	ret := _m.Called(ctx, metrics)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []*record.Metric) error); ok {
		r0 = rf(ctx, metrics)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Shutdown provides a mock function with given fields: ctx
func (_m *MockMeasurementBatcher) Shutdown(ctx context.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewMockMeasurementBatcher creates a new instance of MockMeasurementBatcher. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockMeasurementBatcher(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockMeasurementBatcher {
	mock := &MockMeasurementBatcher{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
