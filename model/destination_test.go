// Copyright  observIQ, Inc.
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
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseDestination(t *testing.T) {
	resources, err := ResourcesFromFile("testfiles/destination-cabin.yaml")
	require.NoError(t, err)

	parsed, err := ParseResourcesStrict(resources)
	require.NoError(t, err)
	require.Len(t, parsed, 1)

	dt, ok := parsed[0].(*Destination)
	require.True(t, ok)
	require.Equal(t, dt.Name(), "cabin-production-logs")
	require.Equal(t, dt.Spec.Type, "observiq-cloud")
	require.Equal(t, dt.Spec.Parameters[0].Name, "endpoint")
	require.Equal(t, dt.Spec.Parameters[0].Value, "https://nozzle.app.observiq.com")
}

func TestPreserveSensitiveParameters(t *testing.T) {
	tests := []struct {
		name             string
		current          *Destination
		existing         *Destination
		expectParameters []Parameter
	}{
		{
			name:             "no parameters, no existing",
			current:          &Destination{},
			existing:         nil,
			expectParameters: []Parameter{},
		},
		{
			name: "some parameters, no existing",
			current: NewDestinationWithSpec("d0", ParameterizedSpec{
				Parameters: []Parameter{
					{
						Name:  "p0",
						Value: SensitiveParameterPlaceholder,
					},
				},
			}),
			existing: nil,
			expectParameters: []Parameter{
				{
					Name:  "p0",
					Value: SensitiveParameterPlaceholder,
				},
			},
		},
		{
			name: "some parameters, with some existing",
			current: NewDestinationWithSpec("d0", ParameterizedSpec{
				Parameters: []Parameter{
					{
						Name:  "p0",
						Value: SensitiveParameterPlaceholder,
					},
					{
						Name:  "p1",
						Value: SensitiveParameterPlaceholder,
					},
					{
						Name:  "p2",
						Value: "new value",
					},
					{
						Name:  "p3",
						Value: "old value",
					},
				},
			}),
			existing: NewDestinationWithSpec("d0", ParameterizedSpec{
				Parameters: []Parameter{
					{
						Name:  "p0",
						Value: "actual value",
					},
					{
						Name:  "p2",
						Value: "old value replaced",
					},
					{
						Name:  "p3",
						Value: "old value",
					},
				},
			}),
			expectParameters: []Parameter{
				{
					Name:  "p0",
					Value: "actual value",
				},
				{
					Name:  "p1",
					Value: nil,
				},
				{
					Name:  "p2",
					Value: "new value",
				},
				{
					Name:  "p3",
					Value: "old value",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()

			anyResource, err := AsAny(test.existing)
			require.NoError(t, err, "failed to convert existing to AnyResource")

			err = PreserveSensitiveParameters(ctx, test.current, anyResource)
			require.NoError(t, err, "failed to preserve sensitive parameters")
			require.ElementsMatch(t, test.expectParameters, test.current.Spec.Parameters)
		})
	}
}

func TestMaskSensitiveParameters(t *testing.T) {
	tests := []struct {
		name             string
		destination      *Destination
		expectParameters []Parameter
	}{
		{
			name: "no parameters",
			destination: NewDestinationWithSpec("d0", ParameterizedSpec{
				Parameters: []Parameter{},
			}),
			expectParameters: []Parameter{},
		},
		{
			name: "one sensitive",
			destination: NewDestinationWithSpec("d0", ParameterizedSpec{
				Parameters: []Parameter{
					{
						Name:  "p0",
						Value: "value",
					},
					{
						Name:      "p1",
						Value:     "value",
						Sensitive: true,
					},
				},
			}),
			expectParameters: []Parameter{
				{
					Name:  "p0",
					Value: "value",
				},
				{
					Name:      "p1",
					Value:     SensitiveParameterPlaceholder,
					Sensitive: true,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			maskSensitiveParameters(ctx, test.destination)
			require.ElementsMatch(t, test.expectParameters, test.destination.Spec.Parameters)
		})
	}
}

func TestWithoutSensitiveParameterMasking(t *testing.T) {
	ctx := context.Background()
	require.False(t, IsWithoutSensitiveParameterMasking(ctx), "false by default")

	ctx = ContextWithoutSensitiveParameterMasking(ctx)
	require.True(t, IsWithoutSensitiveParameterMasking(ctx), "true if set on context")
}
