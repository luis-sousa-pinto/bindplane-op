// Copyright observIQ, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"errors"

	"github.com/mitchellh/mapstructure"
)

// StatusType is a shared implementation of a Status field, GetStatus(), and SetStatus() that is used by resources that
// have a status.
type StatusType[T any] struct {
	Status T `yaml:"status,omitempty" json:"status,omitempty" mapstructure:"status,omitempty"`
}

// NewStatusType returns a new StatusType[T] with the given status.
func NewStatusType[T any](t T) StatusType[T] {
	return StatusType[T]{
		Status: t,
	}
}

// GetStatus returns the status for this resource.
func (t *StatusType[T]) GetStatus() any {
	return t.Status
}

// SetStatus replaces the status for this resource.
func (t *StatusType[T]) SetStatus(status any) error {
	if status == nil {
		var empty T
		t.Status = empty
		return nil
	}
	if s, ok := status.(T); ok {
		// matching type, assign the value
		t.Status = s
		return nil
	}
	if s, ok := status.(map[string]any); ok {
		// not a matching type, try to map it into the right format. this allows
		// Configuration.SetStatus(AnyResource.GetStatus()) to work.
		var empty T
		if err := mapstructure.Decode(s, &empty); err != nil {
			return err
		}
		t.Status = empty
		return nil
	}
	return errors.New("unable to set status")
}

// ----------------------------------------------------------------------
// helpers

// ParseResourceStatus returns the status for the given resource if the status is of the given type. Otherwise, it can
// parse it from a map[string]any.
func ParseResourceStatus[T any](resource Resource) (*T, bool) {
	if resource == nil {
		return nil, false
	}
	status := resource.GetStatus()
	if s, ok := status.(T); ok {
		return &s, true
	}
	if s, ok := status.(map[string]any); ok {
		var empty T
		if err := mapstructure.Decode(s, &empty); err != nil {
			return nil, false
		}
		return &empty, true
	}
	return nil, false
}

// ----------------------------------------------------------------------
// accessors

// IsLatest returns true if the latest field on the status for this resource is true. Currently this is only used for
// Configurations, Sources, Processors, Destinations, SourceTypes, ProcessorTypes, and DestinationTypes. For other
// resources this always returns true.
func (t *StatusType[T]) IsLatest() bool {
	return boolResourceStatus(&t.Status, func(status *ConfigurationStatus) bool { return status.Latest }) ||
		boolResourceStatus(&t.Status, func(status *VersionStatus) bool { return status.Latest }) ||
		boolAnyResourceStatus(&t.Status, "latest")
}

// IsPending returns true if the pending field on the status for this resource is true. Currently this is only used
// for Configurations. For other resources this always returns false.
func (t *StatusType[T]) IsPending() bool {
	return boolResourceStatus(&t.Status, func(status *ConfigurationStatus) bool { return status.Pending }) ||
		boolAnyResourceStatus(&t.Status, "pending")
}

// IsCurrent returns true if the current field on the status for this resource is true. Currently this is only used
// for Configurations. For other resources this always returns false.
func (t *StatusType[T]) IsCurrent() bool {
	return boolResourceStatus(&t.Status, func(status *ConfigurationStatus) bool { return status.Current }) ||
		boolAnyResourceStatus(&t.Status, "current")
}

// boolAnyResourceStatus modifies the resource status if the resource status is a map[string]any. Otherwise, it does
// nothing.
func boolAnyResourceStatus(status any, key string) bool {
	return boolResourceStatus(status, func(status *map[string]any) bool {
		if len(*status) == 0 {
			return false
		}
		if value, ok := (*status)[key].(bool); ok {
			return value
		}
		return false
	})
}

// boolResourceStatus returns the value of the getter if the resource status is of the expected type. Otherwise, it
// returns false.
func boolResourceStatus[T any](status any, getter func(*T) bool) bool {
	if s, ok := status.(*T); ok {
		return getter(s)
	}
	return false
}

// ----------------------------------------------------------------------
// mutators

// SetLatest sets the value of the latest field on the status for this resource. Currently this is only used for
// Configurations, Sources, Processors, Destinations, SourceTypes, ProcessorTypes, and DestinationTypes. For other
// resources this does nothing.
func (t *StatusType[T]) SetLatest(latest bool) {
	_ = modifyResourceStatus(&t.Status, func(status *ConfigurationStatus) { status.Latest = latest }) ||
		modifyResourceStatus(&t.Status, func(status *VersionStatus) { status.Latest = latest }) ||
		modifyAnyResourceStatus(&t.Status, "latest", latest)
}

// SetPending sets the value of the pending field on the status for this resource. Currently this is only used for
// Configurations. For other resources this does nothing.
func (t *StatusType[T]) SetPending(pending bool) {
	_ = modifyResourceStatus(&t.Status, func(status *ConfigurationStatus) { status.Pending = pending }) ||
		modifyAnyResourceStatus(&t.Status, "pending", pending)
}

// SetCurrent sets the value of the current field on the status for this resource. Currently this is only used for
// Configurations. For other resources this does nothing.
func (t *StatusType[T]) SetCurrent(current bool) {
	_ = modifyResourceStatus(&t.Status, func(status *ConfigurationStatus) { status.Current = current }) ||
		modifyAnyResourceStatus(&t.Status, "current", current)
}

// modifyAnyResourceStatus modifies the resource status if the resource status is a map[string]any. Otherwise, it does
// nothing.
func modifyAnyResourceStatus(status any, key string, value any) bool {
	return modifyResourceStatus(status, func(status *map[string]any) {
		if len(*status) == 0 {
			*status = map[string]any{}
		}
		(*status)[key] = value
	})
}

// modifyResourceStatus modifies the resource status if the resource status is of the expected type. Otherwise, it does
// nothing.
func modifyResourceStatus[T any](status any, updater func(*T)) bool {
	if s, ok := status.(*T); ok {
		updater(s)
		return true
	}
	return false
}
