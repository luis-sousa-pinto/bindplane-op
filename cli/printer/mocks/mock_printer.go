// Code generated by mockery v2.28.1. DO NOT EDIT.

package mocks

import (
	model "github.com/observiq/bindplane-op/model"
	mock "github.com/stretchr/testify/mock"
)

// MockPrinter is an autogenerated mock type for the Printer type
type MockPrinter struct {
	mock.Mock
}

// PrintResource provides a mock function with given fields: _a0
func (_m *MockPrinter) PrintResource(_a0 model.Printable) {
	_m.Called(_a0)
}

// PrintResources provides a mock function with given fields: _a0
func (_m *MockPrinter) PrintResources(_a0 []model.Printable) {
	_m.Called(_a0)
}

type mockConstructorTestingTNewMockPrinter interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockPrinter creates a new instance of MockPrinter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockPrinter(t mockConstructorTestingTNewMockPrinter) *MockPrinter {
	mock := &MockPrinter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
