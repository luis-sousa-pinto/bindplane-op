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
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewResourceStatus(t *testing.T) {
	source := &Source{
		ResourceMeta: ResourceMeta{
			Metadata: Metadata{
				Name:        "TestName",
				Description: "TestDescription",
			},
		},
		Spec: ParameterizedSpec{
			Type: "TestType",
		},
	}

	status := NewResourceStatus(source, StatusCreated)
	require.Equal(t, source, status.Resource)
	require.Equal(t, StatusCreated, status.Status)
	require.Equal(t, "", status.Reason)
}

func TestNewResourceStatusWithReason(t *testing.T) {
	source := &Source{
		ResourceMeta: ResourceMeta{
			Metadata: Metadata{
				Name:        "TestName",
				Description: "TestDescription",
			},
		},
		Spec: ParameterizedSpec{
			Type: "TestType",
		},
	}
	status := NewResourceStatusWithReason(source, StatusInvalid, "Invalid format")
	require.Equal(t, source, status.Resource)
	require.Equal(t, StatusInvalid, status.Status)
	require.Equal(t, "Invalid format", status.Reason)
}

func TestNewResourceStatusWithError(t *testing.T) {
	source := &Source{
		ResourceMeta: ResourceMeta{
			Metadata: Metadata{
				Name:        "TestName",
				Description: "TestDescription",
			},
		},
		Spec: ParameterizedSpec{
			Type: "TestType",
		},
	}

	status := NewResourceStatusWithError(source, fmt.Errorf("Error message"))
	require.Equal(t, source, status.Resource)
	require.Equal(t, StatusError, status.Status)
	require.Equal(t, "Error message", status.Reason)
}
func TestAnyResourceStatusMessage(t *testing.T) {
	tests := []struct {
		name     string
		resource AnyResource
		status   UpdateStatus
		reason   string
		expected string
	}{
		{
			name: "without version and reason",
			resource: AnyResource{
				ResourceMeta: ResourceMeta{
					Kind: "TestKind",
					Metadata: Metadata{
						Name:        "TestName",
						Description: "TestDescription",
					},
				},
				Spec: map[string]any{
					"Type": "TestType",
				},
			},
			status:   StatusCreated,
			reason:   "",
			expected: "TestKind TestName created",
		},
		{
			name: "with version and reason",
			resource: AnyResource{
				ResourceMeta: ResourceMeta{
					Kind: "VersionedKind", // A kind that matches the HasVersionKind criteria
					Metadata: Metadata{
						Name:    "TestName",
						Version: 1, // A version for the resource
					},
				},
				Spec: map[string]any{
					"Type": "TestType",
				},
			},
			status:   StatusCreated,
			reason:   "TestReason",
			expected: "VersionedKind TestName created\n\tTestReason",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := &AnyResourceStatus{
				Resource: test.resource,
				Status:   test.status,
				Reason:   test.reason,
			}

			message := s.Message()
			require.Equal(t, test.expected, message)
		})
	}
}
