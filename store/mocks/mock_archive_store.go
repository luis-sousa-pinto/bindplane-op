// Code generated by mockery v2.30.16. DO NOT EDIT.

package mocks

import (
	context "context"

	model "github.com/observiq/bindplane-op/model"
	mock "github.com/stretchr/testify/mock"
)

// MockArchiveStore is an autogenerated mock type for the ArchiveStore type
type MockArchiveStore struct {
	mock.Mock
}

// ResourceHistory provides a mock function with given fields: ctx, resourceKind, resourceName
func (_m *MockArchiveStore) ResourceHistory(ctx context.Context, resourceKind model.Kind, resourceName string) ([]*model.AnyResource, error) {
	ret := _m.Called(ctx, resourceKind, resourceName)

	var r0 []*model.AnyResource
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, model.Kind, string) ([]*model.AnyResource, error)); ok {
		return rf(ctx, resourceKind, resourceName)
	}
	if rf, ok := ret.Get(0).(func(context.Context, model.Kind, string) []*model.AnyResource); ok {
		r0 = rf(ctx, resourceKind, resourceName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.AnyResource)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, model.Kind, string) error); ok {
		r1 = rf(ctx, resourceKind, resourceName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewMockArchiveStore creates a new instance of MockArchiveStore. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockArchiveStore(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockArchiveStore {
	mock := &MockArchiveStore{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
