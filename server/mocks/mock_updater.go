// Code generated by mockery v2.28.1. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MockUpdater is an autogenerated mock type for the Updater type
type MockUpdater struct {
	mock.Mock
}

// Start provides a mock function with given fields: _a0
func (_m *MockUpdater) Start(_a0 context.Context) {
	_m.Called(_a0)
}

// Stop provides a mock function with given fields: _a0
func (_m *MockUpdater) Stop(_a0 context.Context) {
	_m.Called(_a0)
}

type mockConstructorTestingTNewMockUpdater interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockUpdater creates a new instance of MockUpdater. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockUpdater(t mockConstructorTestingTNewMockUpdater) *MockUpdater {
	mock := &MockUpdater{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
