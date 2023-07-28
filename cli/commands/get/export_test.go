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

package get

import (
	"testing"

	"github.com/google/uuid"
	"github.com/observiq/bindplane-op/model"
	"github.com/stretchr/testify/require"
)

func TestExportConfiguration(t *testing.T) {
	cases := []struct {
		name   string
		input  *model.Configuration
		expect *model.Configuration
	}{
		{
			name:   "nil configuration",
			input:  nil,
			expect: nil,
		},
		{
			name: "empty configuration",
			input: &model.Configuration{
				Spec: model.ConfigurationSpec{
					Sources:      nil,
					Destinations: nil,
				},
			},
			expect: &model.Configuration{
				Spec: model.ConfigurationSpec{
					Sources:      nil,
					Destinations: nil,
				},
			},
		},
		{
			name: "full configuration",
			input: &model.Configuration{
				Spec: model.ConfigurationSpec{
					Sources: []model.ResourceConfiguration{
						{
							ID:          uuid.NewString(),
							Name:        "source-0",
							DisplayName: "display-name",
							ParameterizedSpec: model.ParameterizedSpec{
								Type:       "file:30",
								Parameters: []model.Parameter{},
								Processors: []model.ResourceConfiguration{
									{
										ID:   uuid.NewString(),
										Name: "processor-a",
										ParameterizedSpec: model.ParameterizedSpec{
											Type: "add_fields:10",
											Parameters: []model.Parameter{
												{
													Name:  "field-1",
													Value: "value-1",
												},
											},
										},
									},
								},
								Disabled: false,
							},
						},
						{
							ID:          uuid.NewString(),
							Name:        "source-1",
							DisplayName: "display-name-1",
							ParameterizedSpec: model.ParameterizedSpec{
								Type:       "source:4",
								Parameters: []model.Parameter{},
								Processors: []model.ResourceConfiguration{
									{
										ID:   uuid.NewString(),
										Name: "processor-b",
										ParameterizedSpec: model.ParameterizedSpec{
											Type:       "remove_fields:8",
											Parameters: []model.Parameter{},
										},
									},
								},
								Disabled: false,
							},
						},
					},
					Destinations: []model.ResourceConfiguration{
						{
							ID:          "",
							Name:        "dest-0",
							DisplayName: "display-name",
							ParameterizedSpec: model.ParameterizedSpec{
								Type:       "google",
								Parameters: []model.Parameter{},
								Processors: []model.ResourceConfiguration{
									{
										ID:   "",
										Name: "processor-a",
										ParameterizedSpec: model.ParameterizedSpec{
											Type:       "add_fields",
											Parameters: []model.Parameter{},
										},
									},
								},
								Disabled: false,
							},
						},
						{
							ID:          "",
							Name:        "dest-1",
							DisplayName: "display-name-1",
							ParameterizedSpec: model.ParameterizedSpec{
								Type:       "datadog",
								Parameters: []model.Parameter{},
								Processors: []model.ResourceConfiguration{
									{
										ID:   "",
										Name: "processor-b",
										ParameterizedSpec: model.ParameterizedSpec{
											Type:       "remove_fields",
											Parameters: []model.Parameter{},
										},
									},
								},
								Disabled: false,
							},
						},
					},
				},
			},
			expect: &model.Configuration{
				Spec: model.ConfigurationSpec{
					Sources: []model.ResourceConfiguration{
						{
							ID:          "",
							Name:        "source-0",
							DisplayName: "display-name",
							ParameterizedSpec: model.ParameterizedSpec{
								Type:       "file",
								Parameters: []model.Parameter{},
								Processors: []model.ResourceConfiguration{
									{
										ID:   "",
										Name: "processor-a",
										ParameterizedSpec: model.ParameterizedSpec{
											Type: "add_fields",
											Parameters: []model.Parameter{
												{
													Name:  "field-1",
													Value: "value-1",
												},
											}},
									},
								},
								Disabled: false,
							},
						},
						{
							ID:          "",
							Name:        "source-1",
							DisplayName: "display-name-1",
							ParameterizedSpec: model.ParameterizedSpec{
								Type:       "source",
								Parameters: []model.Parameter{},
								Processors: []model.ResourceConfiguration{
									{
										ID:   "",
										Name: "processor-b",
										ParameterizedSpec: model.ParameterizedSpec{
											Type:       "remove_fields",
											Parameters: []model.Parameter{},
										},
									},
								},
								Disabled: false,
							},
						},
					},
					Destinations: []model.ResourceConfiguration{
						{
							ID:          "",
							Name:        "dest-0",
							DisplayName: "display-name",
							ParameterizedSpec: model.ParameterizedSpec{
								Type:       "google",
								Parameters: []model.Parameter{},
								Processors: []model.ResourceConfiguration{
									{
										ID:   "",
										Name: "processor-a",
										ParameterizedSpec: model.ParameterizedSpec{
											Type:       "add_fields",
											Parameters: []model.Parameter{},
										},
									},
								},
								Disabled: false,
							},
						},
						{
							ID:          "",
							Name:        "dest-1",
							DisplayName: "display-name-1",
							ParameterizedSpec: model.ParameterizedSpec{
								Type:       "datadog",
								Parameters: []model.Parameter{},
								Processors: []model.ResourceConfiguration{
									{
										ID:   "",
										Name: "processor-b",
										ParameterizedSpec: model.ParameterizedSpec{
											Type:       "remove_fields",
											Parameters: []model.Parameter{},
										},
									},
								},
								Disabled: false,
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			exportConfiguration(tc.input)
			require.Equal(t, tc.expect, tc.input)
		})
	}
}
