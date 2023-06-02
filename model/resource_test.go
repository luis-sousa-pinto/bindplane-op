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
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

var (
	anyResourceV2 = &AnyResource{
		ResourceMeta: ResourceMeta{
			Kind: KindConfiguration,
			Metadata: Metadata{
				Name: "test",
			},
		},
		Spec: map[string]interface{}{},
		StatusType: NewStatusType(map[string]interface{}{
			"currentVersion": 2,
		}),
	}

	configurationV3 = Configuration{
		ResourceMeta: ResourceMeta{
			Kind: KindConfiguration,
			Metadata: Metadata{
				Name: "test",
				Labels: LabelsFromValidatedMap(map[string]string{
					"foo": "bar",
				}),
			},
		},
		Spec: ConfigurationSpec{
			Selector: AgentSelector{
				MatchLabels: MatchLabels{},
			},
		},
		StatusType: NewStatusType(ConfigurationStatus{
			CurrentVersion: 3,
		}),
	}
)

func TestResourcesFromFileErrors(t *testing.T) {
	var tests = []struct {
		file string
	}{
		{
			file: "source-malformed-yaml.yaml",
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("testfile:%s", test.file), func(t *testing.T) {
			_, err := ResourcesFromFile(filepath.Join("testfiles", test.file))
			assert.Error(t, err)
		})
	}
}

func TestResourceValidate(t *testing.T) {
	tests := []struct {
		name      string
		resource  Resource
		errorMsgs []string
	}{
		{
			name: "invalid name",
			resource: &Configuration{
				ResourceMeta: ResourceMeta{
					Metadata: Metadata{
						Name: "invalid=name",
					},
				},
			},
			errorMsgs: []string{
				"invalid=name is not a valid resource name",
			},
		},
		{
			name: "invalid kind unknown",
			resource: &AnyResource{
				ResourceMeta: ResourceMeta{
					Kind: KindUnknown,
					Metadata: Metadata{
						Name: "invalid-kind",
					},
				},
			},
			errorMsgs: []string{
				"Unknown is not a valid resource kind",
			},
		},
		{
			name: "invalid kind string",
			resource: &AnyResource{
				ResourceMeta: ResourceMeta{
					Kind: Kind("invalid"),
					Metadata: Metadata{
						Name: "invalid-kind",
					},
				},
			},
			errorMsgs: []string{
				"1 error occurred:\n\t* invalid is not a valid resource kind",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := test.resource.Validate()
			if len(test.errorMsgs) == 0 {
				require.NoError(t, err)
			} else {
				for _, errorMsg := range test.errorMsgs {
					require.Contains(t, err.Error(), errorMsg)
				}
			}
		})
	}
}

func TestParseSourceType(t *testing.T) {
	resources, err := ResourcesFromFile(filepath.Join("testfiles", "sourcetype-macos.yaml"))
	assert.NoError(t, err)

	parsed, err := ParseResources(resources)
	require.NoError(t, err)

	sourceType, ok := parsed[0].(*SourceType)
	require.True(t, ok)

	sourceType.EnsureHash(sourceType.Spec)

	expect := &SourceType{
		ResourceType: ResourceType{
			ResourceMeta: ResourceMeta{
				APIVersion: "bindplane.observiq.com/v1",
				Kind:       "SourceType",
				Metadata: Metadata{
					Name:        "MacOS",
					DisplayName: "Mac OS",
					Description: "Log parser for MacOS",
					Icon:        "/public/bindplane-logo.png",
					Hash:        "fdc15d3d30cb694b2af9356f8fa8c32ffc7327436298b6fb0126b8e834ef22f1",
					Version:     Version(1),
				},
			},
			Spec: ResourceTypeSpec{
				Version:            "0.0.2",
				SupportedPlatforms: []string{"macos"},
				Parameters: []ParameterDefinition{
					{
						Name:        "enable_system_log",
						Label:       "System Logs",
						Description: "Enable to collect MacOS system logs",
						Type:        "bool",
						Default:     true,
					},
					{
						Name:        "system_log_path",
						Label:       "System Log Path",
						Description: "The absolute path to the System log",
						Type:        "string",
						Default:     "/var/log/system.log",
						RelevantIf: []RelevantIfCondition{
							{
								Name:     "enable_system_log",
								Operator: "equals",
								Value:    true,
							},
						},
					},
					{
						Name:        "enable_install_log",
						Label:       "Install Logs",
						Description: "Enable to collect MacOS install logs",
						Type:        "bool",
						Default:     true,
					},
					{
						Name:        "install_log_path",
						Label:       "Install Log Path",
						Description: "The absolute path to the Install log",
						Type:        "string",
						Default:     "/var/log/install.log",
						RelevantIf: []RelevantIfCondition{
							{
								Name:     "enable_install_log",
								Operator: "equals",
								Value:    true,
							},
						},
					},
					{
						Name:    "collection_interval_seconds",
						Label:   "Collection Interval",
						Type:    "int",
						Default: "30",
					},
					{
						Name:        "start_at",
						Label:       "Start At",
						Description: "Start reading file from 'beginning' or 'end'",
						Type:        "enum",
						ValidValues: []string{"beginning", "end"},
						Default:     "end",
					},
				},
				Logs: ResourceTypeOutput{
					Receivers: ResourceTypeTemplate(strings.TrimLeft(`
- plugin/macos:
    plugin:
      name: macos
    parameters:
    - name: enable_system_log
      value: {{ .enable_system_log }}
    - name: system_log_path
      value: {{ .system_log_path }}
    - name: enable_install_log
      value: {{ .enable_install_log }}
    - name: install_log_path
      value: {{ .install_log_path }}
    - name: start_at
      value: {{ .start_at }}
- plugin/journald:
    plugin:
      name: journald
`, "\n")),
				},
				Metrics: ResourceTypeOutput{
					Receivers: ResourceTypeTemplate(strings.TrimLeft(`
- hostmetrics:
    collection_interval: 1m
    scrapers:
      load:
`, "\n")),
				},
			},
		},
	}
	require.Equal(t, expect, sourceType)
}

// Below tests confirm that all seed resources are valid

func fileResource[T Resource](t *testing.T, path string) T {
	resources, err := ResourcesFromFile(path)
	require.NoError(t, err)

	parsed, err := ParseResources(resources)
	require.NoError(t, err)
	require.Len(t, parsed, 1)

	resource, ok := parsed[0].(T)
	require.True(t, ok)

	if resource.Version() == 0 {
		resource.SetVersion(1)
	}

	return resource
}

func resourcePaths(t *testing.T, folder string) []string {
	files, err := ioutil.ReadDir(folder)
	require.NoError(t, err)

	result := make([]string, len(files))
	for i, file := range files {
		result[i] = filepath.Join(folder, file.Name())
	}

	return result
}

func TestValidateSourceTypes(t *testing.T) {
	paths := resourcePaths(t, "../resources/source-types")
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			resource := fileResource[*SourceType](t, path)
			warn, err := resource.Validate()
			require.NoError(t, err)
			require.Equal(t, "", warn)
		})
	}
}

func TestValidateProcessorTypes(t *testing.T) {
	paths := resourcePaths(t, "../resources/processor-types")
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			resource := fileResource[*ProcessorType](t, path)
			warn, err := resource.Validate()
			require.NoError(t, err)
			require.Equal(t, "", warn)
		})
	}
}

func TestValidateDestinationTypes(t *testing.T) {
	paths := resourcePaths(t, "../resources/destination-types")
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			resource := fileResource[*DestinationType](t, path)
			warn, err := resource.Validate()
			require.NoError(t, err)
			require.Equal(t, "", warn)
		})
	}
}

func TestValidateAgentVersions(t *testing.T) {
	paths := resourcePaths(t, "../resources/agent-versions")
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			resource := fileResource[*AgentVersion](t, path)
			warn, err := resource.Validate()
			require.NoError(t, err)
			require.Equal(t, "", warn)
		})
	}
}

func TestResourceSetStatus(t *testing.T) {
	tests := []struct {
		name          string
		status        any
		expectVersion Version
	}{
		{
			name: "set with map",
			status: map[string]interface{}{
				"currentVersion": 2,
			},
			expectVersion: 2,
		},
		{
			name: "set with configuration status",
			status: ConfigurationStatus{
				CurrentVersion: 3,
			},
			expectVersion: 3,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clone, err := Clone(&configurationV3)
			require.NoError(t, err)
			err = clone.SetStatus(test.status)
			require.NoError(t, err)
			require.Equal(t, test.expectVersion, clone.Status.CurrentVersion)
			require.Equal(t, ConfigurationStatus{
				CurrentVersion: test.expectVersion,
			}, clone.GetStatus())
		})
	}
}

func TestResourceStatusAfterClone(t *testing.T) {
	tests := []struct {
		name          string
		resource      Resource
		expectVersion Version
	}{
		{
			name:          "any resource",
			resource:      anyResourceV2,
			expectVersion: 2,
		},
		{
			name:          "configuration resource",
			resource:      &configurationV3,
			expectVersion: 3,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clone, err := Clone(test.resource)
			require.NoError(t, err)

			configuration, ok := clone.(*Configuration)
			require.True(t, ok)
			require.Equal(t, test.expectVersion, configuration.Status.CurrentVersion)
		})
	}
}

func TestResourceStatusYAML(t *testing.T) {
	tests := []struct {
		name     string
		resource Resource
	}{
		{
			name:     "configuration resource",
			resource: &configurationV3,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// marshal to YAML
			data, err := yaml.Marshal(test.resource)
			require.NoError(t, err)

			// unmarshal from YAML
			var unmarshaled Configuration
			err = yaml.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)

			require.Equal(t, test.resource, &unmarshaled)
		})
	}
}

func TestResourceStatusJSON(t *testing.T) {
	tests := []struct {
		name     string
		resource Resource
	}{
		{
			name:     "configuration resource",
			resource: &configurationV3,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// marshal to JSON
			data, err := json.Marshal(test.resource)
			require.NoError(t, err)

			// unmarshal from JSON
			var unmarshaled Configuration
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)

			require.Equal(t, test.resource, &unmarshaled)
		})
	}
}

func TestSplitVersion(t *testing.T) {
	tests := []struct {
		resourceKey   string
		expectName    string
		expectVersion Version
	}{
		{
			resourceKey:   "foo",
			expectName:    "foo",
			expectVersion: VersionLatest,
		},
		{
			resourceKey:   "foo:1",
			expectName:    "foo",
			expectVersion: 1,
		},
		{
			resourceKey:   "foo:",
			expectName:    "foo",
			expectVersion: VersionLatest,
		},
		{
			resourceKey:   "foo:unknown",
			expectName:    "foo",
			expectVersion: VersionLatest,
		},
		{
			resourceKey:   "foo:1:2",
			expectName:    "foo",
			expectVersion: VersionLatest,
		},
		{
			resourceKey:   "foo:current",
			expectName:    "foo",
			expectVersion: VersionCurrent,
		},
	}
	for _, tt := range tests {
		t.Run(tt.resourceKey, func(t *testing.T) {
			got, got1 := SplitVersion(tt.resourceKey)
			if got != tt.expectName {
				t.Errorf("SplitVersion() got = %v, want %v", got, tt.expectName)
			}
			if got1 != tt.expectVersion {
				t.Errorf("SplitVersion() got1 = %v, want %v", got1, tt.expectVersion)
			}
		})
	}
}

func TestJoinVersion(t *testing.T) {
	tests := []struct {
		resourceKey string
		version     Version
		expect      string
	}{
		{
			resourceKey: "foo",
			version:     VersionLatest,
			expect:      "foo",
		},
		{
			resourceKey: "foo",
			version:     1,
			expect:      "foo:1",
		},
		{
			resourceKey: "foo",
			version:     VersionCurrent,
			expect:      "foo:current",
		},
		{
			resourceKey: "foo:1",
			version:     VersionLatest,
			expect:      "foo",
		},
		{
			resourceKey: "foo:1",
			version:     2,
			expect:      "foo:2",
		},
	}
	for _, test := range tests {
		t.Run(test.resourceKey, func(t *testing.T) {
			got := JoinVersion(test.resourceKey, test.version)
			if got != test.expect {
				t.Errorf("JoinVersion() got = %v, want %v", got, test.expect)
			}
		})
	}
}

func TestTrimVersion(t *testing.T) {
	tests := []struct {
		resourceKey string
		expect      string
	}{
		{
			resourceKey: "",
			expect:      "",
		},
		{
			resourceKey: "foo",
			expect:      "foo",
		},
		{
			resourceKey: "foo:1",
			expect:      "foo",
		},
		{
			resourceKey: "foo:2",
			expect:      "foo",
		},
		{
			resourceKey: "foo:2:1",
			expect:      "foo",
		},
	}
	for _, test := range tests {
		t.Run(test.resourceKey, func(t *testing.T) {
			got := TrimVersion(test.resourceKey)
			if got != test.expect {
				t.Errorf("TrimVersion() got = %v, want %v", got, test.expect)
			}
		})
	}
}

func TestUpdateDependencies(t *testing.T) {
	ctx := context.Background()

	sourceWithType := func(typeName string) *Source {
		return &Source{
			ResourceMeta: ResourceMeta{
				Metadata: Metadata{
					Name:    "source1",
					Version: 2,
				},
			},
			Spec: ParameterizedSpec{
				Type: typeName,
			},
		}
	}
	processorWithType := func(typeName string) *Processor {
		return &Processor{
			ResourceMeta: ResourceMeta{
				Metadata: Metadata{
					Name: "processor1",
				},
			},
			Spec: ParameterizedSpec{
				Type: typeName,
			},
		}
	}
	destinationWithType := func(typeName string) *Destination {
		return &Destination{
			ResourceMeta: ResourceMeta{
				Metadata: Metadata{
					Name: "destination1",
				},
			},
			Spec: ParameterizedSpec{
				Type: typeName,
			},
		}
	}

	sourceTypeV2 := &SourceType{
		ResourceType: ResourceType{
			ResourceMeta: ResourceMeta{
				Metadata: Metadata{
					Name:    "source-type",
					Version: 2,
				},
			},
		},
	}
	processorTypeV2 := &ProcessorType{
		ResourceType: ResourceType{
			ResourceMeta: ResourceMeta{
				Metadata: Metadata{
					Name:    "processor-type",
					Version: 2,
				},
			},
		},
	}
	destinationTypeV2 := &DestinationType{
		ResourceType: ResourceType{
			ResourceMeta: ResourceMeta{
				Metadata: Metadata{
					Name:    "destination-type",
					Version: 2,
				},
			},
		},
	}

	tests := []struct {
		name     string
		setup    func(*testing.T, *MockResourceStore)
		resource Resource
		expect   Resource
	}{
		{
			name: "update source-type reference in source",
			setup: func(t *testing.T, store *MockResourceStore) {
				store.EXPECT().SourceType(ctx, "source-type").Return(sourceTypeV2, nil)
			},
			resource: sourceWithType("source-type:1"),
			expect:   sourceWithType("source-type:2"),
		},
		{
			name: "no change, already using latest",
			setup: func(t *testing.T, store *MockResourceStore) {
				store.EXPECT().SourceType(ctx, "source-type").Return(sourceTypeV2, nil)
			},
			resource: sourceWithType("source-type:2"),
			expect:   sourceWithType("source-type:2"),
		},
		{
			name: "no version, add version",
			setup: func(t *testing.T, store *MockResourceStore) {
				store.EXPECT().SourceType(ctx, "source-type").Return(sourceTypeV2, nil)
			},
			resource: sourceWithType("source-type"),
			expect:   sourceWithType("source-type:2"),
		},
		{
			name: "update processor-type reference in processor",
			setup: func(t *testing.T, store *MockResourceStore) {
				store.EXPECT().ProcessorType(ctx, "processor-type").Return(processorTypeV2, nil)
			},
			resource: processorWithType("processor-type:1"),
			expect:   processorWithType("processor-type:2"),
		},
		{
			name: "update destination-type reference in destination",
			setup: func(t *testing.T, store *MockResourceStore) {
				store.EXPECT().DestinationType(ctx, "destination-type").Return(destinationTypeV2, nil)
			},
			resource: destinationWithType("destination-type:1"),
			expect:   destinationWithType("destination-type:2"),
		},
		{
			name: "update configuration with sources, processors, and destinations",
			setup: func(t *testing.T, store *MockResourceStore) {
				store.EXPECT().SourceType(ctx, "source-type").Return(sourceTypeV2, nil)
				store.EXPECT().DestinationType(ctx, "destination-type").Return(destinationTypeV2, nil)
				store.EXPECT().ProcessorType(ctx, "processor-type").Return(processorTypeV2, nil).Times(2)
			},
			resource: &Configuration{
				ResourceMeta: ResourceMeta{
					Metadata: Metadata{
						Name: "config1",
					},
				},
				Spec: ConfigurationSpec{
					Sources: []ResourceConfiguration{
						{
							ParameterizedSpec: ParameterizedSpec{
								Type: "source-type:1",
								Processors: []ResourceConfiguration{
									{
										ParameterizedSpec: ParameterizedSpec{
											Type: "processor-type:1",
										},
									},
								},
							},
						},
					},
					Destinations: []ResourceConfiguration{
						{
							ParameterizedSpec: ParameterizedSpec{
								Type: "destination-type:1",
								Processors: []ResourceConfiguration{
									{
										ParameterizedSpec: ParameterizedSpec{
											Type: "processor-type:1",
										},
									},
								},
							},
						},
					},
				},
			},
			expect: &Configuration{
				ResourceMeta: ResourceMeta{
					Metadata: Metadata{
						Name: "config1",
					},
				},
				Spec: ConfigurationSpec{
					Sources: []ResourceConfiguration{
						{
							ParameterizedSpec: ParameterizedSpec{
								Type: "source-type:2",
								Processors: []ResourceConfiguration{
									{
										ParameterizedSpec: ParameterizedSpec{
											Type: "processor-type:2",
										},
									},
								},
							},
						},
					},
					Destinations: []ResourceConfiguration{
						{
							ParameterizedSpec: ParameterizedSpec{
								Type: "destination-type:2",
								Processors: []ResourceConfiguration{
									{
										ParameterizedSpec: ParameterizedSpec{
											Type: "processor-type:2",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := NewMockResourceStore(t)

			if test.setup != nil {
				test.setup(t, store)
			}

			err := test.resource.UpdateDependencies(ctx, store)
			require.NoError(t, err)

			require.Equal(t, test.expect, test.resource)
		})
	}
}

func TestParseSourceTypeStrict_TypenameKey(t *testing.T) {
	resources, err := ResourcesFromFile(filepath.Join("testfiles", "sourcetype-macos-typename-key.yaml"))
	assert.NoError(t, err)

	parsed, err := ParseResourcesStrict(resources)
	require.NoError(t, err)

	sourceType, ok := parsed[0].(*SourceType)
	require.True(t, ok)

	expect := &SourceType{
		ResourceType: ResourceType{
			ResourceMeta: ResourceMeta{
				APIVersion: "bindplane.observiq.com/v1",
				Kind:       "SourceType",
				Metadata: Metadata{
					Name:        "MacOS",
					DisplayName: "Mac OS",
					Description: "Log parser for MacOS",
					Icon:        "/public/bindplane-logo.png",
					Version:     Version(1),
				},
			},
			Spec: ResourceTypeSpec{
				Version:            "0.0.2",
				SupportedPlatforms: []string{"macos"},
				Parameters: []ParameterDefinition{
					{
						Name:        "enable_system_log",
						Label:       "System Logs",
						Description: "Enable to collect MacOS system logs",
						Type:        "bool",
						Default:     true,
					},
					{
						Name:        "system_log_path",
						Label:       "System Log Path",
						Description: "The absolute path to the System log",
						Type:        "string",
						Default:     "/var/log/system.log",
						RelevantIf: []RelevantIfCondition{
							{
								Name:     "enable_system_log",
								Operator: "equals",
								Value:    true,
							},
						},
					},
					{
						Name:        "enable_install_log",
						Label:       "Install Logs",
						Description: "Enable to collect MacOS install logs",
						Type:        "bool",
						Default:     true,
					},
					{
						Name:        "install_log_path",
						Label:       "Install Log Path",
						Description: "The absolute path to the Install log",
						Type:        "string",
						Default:     "/var/log/install.log",
						RelevantIf: []RelevantIfCondition{
							{
								Name:     "enable_install_log",
								Operator: "equals",
								Value:    true,
							},
						},
					},
					{
						Name:    "collection_interval_seconds",
						Label:   "Collection Interval",
						Type:    "int",
						Default: "30",
					},
					{
						Name:        "start_at",
						Label:       "Start At",
						Description: "Start reading file from 'beginning' or 'end'",
						Type:        "enum",
						ValidValues: []string{"beginning", "end"},
						Default:     "end",
					},
				},
				Logs: ResourceTypeOutput{
					Receivers: ResourceTypeTemplate(strings.TrimLeft(`
- plugin/macos:
    plugin:
      name: macos
    parameters:
    - name: enable_system_log
      value: {{ .enable_system_log }}
    - name: system_log_path
      value: {{ .system_log_path }}
    - name: enable_install_log
      value: {{ .enable_install_log }}
    - name: install_log_path
      value: {{ .install_log_path }}
    - name: start_at
      value: {{ .start_at }}
- plugin/journald:
    plugin:
      name: journald
`, "\n")),
				},
				Metrics: ResourceTypeOutput{
					Receivers: ResourceTypeTemplate(strings.TrimLeft(`
- hostmetrics:
    collection_interval: 1m
    scrapers:
      load:
`, "\n")),
				},
			},
		},
	}
	require.Equal(t, expect, sourceType)
}
