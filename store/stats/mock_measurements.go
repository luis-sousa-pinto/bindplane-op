// Code generated by mockery v2.31.1. DO NOT EDIT.

package stats

import (
	context "context"

	record "github.com/observiq/bindplane-op/otlp/record"
	mock "github.com/stretchr/testify/mock"
)

// mockMeasurements is an autogenerated mock type for the Measurements type
type mockMeasurements struct {
	mock.Mock
}

// AgentMetrics provides a mock function with given fields: ctx, id, options
func (_m *mockMeasurements) AgentMetrics(ctx context.Context, id []string, options ...QueryOption) (MetricData, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, id)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 MetricData
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []string, ...QueryOption) (MetricData, error)); ok {
		return rf(ctx, id, options...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []string, ...QueryOption) MetricData); ok {
		r0 = rf(ctx, id, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(MetricData)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, []string, ...QueryOption) error); ok {
		r1 = rf(ctx, id, options...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Clear provides a mock function with given fields:
func (_m *mockMeasurements) Clear() {
	_m.Called()
}

// ConfigurationMetrics provides a mock function with given fields: ctx, name, options
func (_m *mockMeasurements) ConfigurationMetrics(ctx context.Context, name string, options ...QueryOption) (MetricData, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, name)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 MetricData
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, ...QueryOption) (MetricData, error)); ok {
		return rf(ctx, name, options...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, ...QueryOption) MetricData); ok {
		r0 = rf(ctx, name, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(MetricData)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, ...QueryOption) error); ok {
		r1 = rf(ctx, name, options...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MeasurementsSize provides a mock function with given fields: _a0
func (_m *mockMeasurements) MeasurementsSize(_a0 context.Context) (int, error) {
	ret := _m.Called(_a0)

	var r0 int
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (int, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(context.Context) int); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(int)
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OverviewMetrics provides a mock function with given fields: ctx, options
func (_m *mockMeasurements) OverviewMetrics(ctx context.Context, options ...QueryOption) (MetricData, error) {
	_va := make([]interface{}, len(options))
	for _i := range options {
		_va[_i] = options[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 MetricData
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, ...QueryOption) (MetricData, error)); ok {
		return rf(ctx, options...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, ...QueryOption) MetricData); ok {
		r0 = rf(ctx, options...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(MetricData)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, ...QueryOption) error); ok {
		r1 = rf(ctx, options...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ProcessMetrics provides a mock function with given fields: ctx
func (_m *mockMeasurements) ProcessMetrics(ctx context.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SaveAgentMetrics provides a mock function with given fields: ctx, metrics
func (_m *mockMeasurements) SaveAgentMetrics(ctx context.Context, metrics []*record.Metric) error {
	ret := _m.Called(ctx, metrics)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []*record.Metric) error); ok {
		r0 = rf(ctx, metrics)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// newMockMeasurements creates a new instance of mockMeasurements. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockMeasurements(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockMeasurements {
	mock := &mockMeasurements{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
