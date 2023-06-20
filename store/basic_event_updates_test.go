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

package store

import (
	"testing"

	"github.com/observiq/bindplane-op/model"
	"github.com/stretchr/testify/require"
)

func TestUpdatesIncludeAgent(t *testing.T) {
	updates := &EventUpdates{}
	agent := &model.Agent{
		ID: "test",
	}

	updates.IncludeAgent(agent, EventTypeInsert)
	require.Equal(t, 1, len(updates.AgentsField))
	require.Equal(t, agent, updates.AgentsField[agent.UniqueKey()].Item)
}

func TestUpdatesIncludeResource(t *testing.T) {
	agentVersion := &model.AgentVersion{}
	source := &model.Source{}
	sourceType := &model.SourceType{}
	processor := &model.Processor{}
	processorType := &model.ProcessorType{}
	destination := &model.Destination{}
	destinationType := &model.DestinationType{}
	configuration := &model.Configuration{}

	updates := &EventUpdates{}
	updates.IncludeResource(agentVersion, EventTypeInsert)
	updates.IncludeResource(source, EventTypeInsert)
	updates.IncludeResource(sourceType, EventTypeInsert)
	updates.IncludeResource(processor, EventTypeInsert)
	updates.IncludeResource(processorType, EventTypeInsert)
	updates.IncludeResource(destination, EventTypeInsert)
	updates.IncludeResource(destinationType, EventTypeInsert)
	updates.IncludeResource(configuration, EventTypeInsert)
	updates.IncludeResource(nil, EventTypeInsert)

	require.Equal(t, 8, updates.Size())
	require.Equal(t, agentVersion, updates.AgentVersionsField[agentVersion.UniqueKey()].Item)
	require.Equal(t, source, updates.SourcesField[source.UniqueKey()].Item)
	require.Equal(t, sourceType, updates.SourceTypesField[sourceType.UniqueKey()].Item)
	require.Equal(t, processor, updates.ProcessorsField[processor.UniqueKey()].Item)
	require.Equal(t, processorType, updates.ProcessorTypesField[processorType.UniqueKey()].Item)
	require.Equal(t, destination, updates.DestinationsField[destination.UniqueKey()].Item)
	require.Equal(t, destinationType, updates.DestinationTypesField[destinationType.UniqueKey()].Item)
	require.Equal(t, configuration, updates.ConfigurationsField[configuration.UniqueKey()].Item)
}

func TestUpdatesEmpty(t *testing.T) {
	testCases := []struct {
		name     string
		updates  BasicEventUpdates
		expected bool
	}{
		{
			name:     "empty",
			updates:  NewEventUpdates(),
			expected: true,
		},
		{
			name: "not empty",
			updates: &EventUpdates{
				AgentsField: Events[*model.Agent]{
					"test": Event[*model.Agent]{},
				},
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.updates.Empty())
		})
	}
}

func TestUpdatesSize(t *testing.T) {
	testCases := []struct {
		name     string
		updates  BasicEventUpdates
		expected int
	}{
		{
			name:     "empty",
			updates:  NewEventUpdates(),
			expected: 0,
		},
		{
			name: "with resources",
			updates: &EventUpdates{
				AgentsField: Events[*model.Agent]{
					"test": Event[*model.Agent]{},
				},
				AgentVersionsField: Events[*model.AgentVersion]{
					"test": Event[*model.AgentVersion]{},
				},
				SourcesField: Events[*model.Source]{
					"test": Event[*model.Source]{},
				},
				SourceTypesField: Events[*model.SourceType]{
					"test": Event[*model.SourceType]{},
				},
				ProcessorsField: Events[*model.Processor]{
					"test": Event[*model.Processor]{},
				},
				ProcessorTypesField: Events[*model.ProcessorType]{
					"test": Event[*model.ProcessorType]{},
				},
				DestinationsField: Events[*model.Destination]{
					"test": Event[*model.Destination]{},
				},
				DestinationTypesField: Events[*model.DestinationType]{
					"test": Event[*model.DestinationType]{},
				},
			},
			expected: 8,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.updates.Size())
		})
	}
}

func TestUpdatesCouldAffectProcessors(t *testing.T) {
	testCases := []struct {
		name     string
		updates  BasicEventUpdates
		expected bool
	}{
		{
			name:     "empty",
			updates:  NewEventUpdates(),
			expected: false,
		},
		{
			name: "with processors",
			updates: &EventUpdates{
				ProcessorsField: Events[*model.Processor]{
					"test": Event[*model.Processor]{},
				},
			},
			expected: false,
		},
		{
			name: "with processor types",
			updates: &EventUpdates{
				ProcessorTypesField: Events[*model.ProcessorType]{
					"test": Event[*model.ProcessorType]{},
				},
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.updates.CouldAffectProcessors())
		})
	}
}

func TestUpdatesCouldAffectSources(t *testing.T) {
	testCases := []struct {
		name     string
		updates  BasicEventUpdates
		expected bool
	}{
		{
			name:     "empty",
			updates:  NewEventUpdates(),
			expected: false,
		},
		{
			name: "with sources",
			updates: &EventUpdates{
				SourcesField: Events[*model.Source]{
					"test": Event[*model.Source]{},
				},
			},
			expected: false,
		},
		{
			name: "with source types",
			updates: &EventUpdates{
				SourceTypesField: Events[*model.SourceType]{
					"test": Event[*model.SourceType]{},
				},
			},
			expected: true,
		},
		{
			name: "with processors",
			updates: &EventUpdates{
				ProcessorsField: Events[*model.Processor]{
					"test": Event[*model.Processor]{},
				},
			},
			expected: true,
		},
		{
			name: "with processor types",
			updates: &EventUpdates{
				ProcessorTypesField: Events[*model.ProcessorType]{
					"test": Event[*model.ProcessorType]{},
				},
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.updates.CouldAffectSources())
		})
	}
}

func TestUpdatesCouldAffectDestinations(t *testing.T) {
	testCases := []struct {
		name     string
		updates  BasicEventUpdates
		expected bool
	}{
		{
			name:     "empty",
			updates:  NewEventUpdates(),
			expected: false,
		},
		{
			name: "with destinations",
			updates: &EventUpdates{
				DestinationsField: Events[*model.Destination]{
					"test": Event[*model.Destination]{},
				},
			},
			expected: false,
		},
		{
			name: "with destination types",
			updates: &EventUpdates{
				DestinationTypesField: Events[*model.DestinationType]{
					"test": Event[*model.DestinationType]{},
				},
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.updates.CouldAffectDestinations())
		})
	}
}

func TestUpdatesCouldAffectConfigurations(t *testing.T) {
	testCases := []struct {
		name     string
		updates  BasicEventUpdates
		expected bool
	}{
		{
			name:     "empty",
			updates:  NewEventUpdates(),
			expected: false,
		},
		{
			name: "with sources",
			updates: &EventUpdates{
				SourcesField: Events[*model.Source]{
					"test": Event[*model.Source]{},
				},
			},
			expected: true,
		},
		{
			name: "with source types",
			updates: &EventUpdates{
				SourceTypesField: Events[*model.SourceType]{
					"test": Event[*model.SourceType]{},
				},
			},
			expected: true,
		},
		{
			name: "with destinations",
			updates: &EventUpdates{
				DestinationsField: Events[*model.Destination]{
					"test": Event[*model.Destination]{},
				},
			},
			expected: true,
		},
		{
			name: "with destination types",
			updates: &EventUpdates{
				DestinationTypesField: Events[*model.DestinationType]{
					"test": Event[*model.DestinationType]{},
				},
			},
			expected: true,
		},
		{
			name: "with processors",
			updates: &EventUpdates{
				ProcessorsField: Events[*model.Processor]{
					"test": Event[*model.Processor]{},
				},
			},
			expected: true,
		},
		{
			name: "with processor types",
			updates: &EventUpdates{
				ProcessorTypesField: Events[*model.ProcessorType]{
					"test": Event[*model.ProcessorType]{},
				},
			},
			expected: true,
		},
		{
			name: "with configurations",
			updates: &EventUpdates{
				ConfigurationsField: Events[*model.Configuration]{
					"test": Event[*model.Configuration]{},
				},
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.updates.CouldAffectConfigurations())
		})
	}
}

func TestUpdatesAffectsSource(t *testing.T) {
	testCases := []struct {
		name     string
		updates  BasicEventUpdates
		source   *model.Source
		expected bool
	}{
		{
			name:     "empty",
			updates:  NewEventUpdates(),
			source:   &model.Source{},
			expected: false,
		},
		{
			name: "with source type update",
			updates: &EventUpdates{
				SourceTypesField: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil, []string{"macos", "linux", "windows"}),
						Type: EventTypeUpdate,
					},
				},
			},
			source:   model.NewSource("test-name", "test-source-type", nil),
			expected: true,
		},
		{
			name: "with source type insert",
			updates: &EventUpdates{
				SourceTypesField: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil, []string{"macos", "linux", "windows"}),
						Type: EventTypeInsert,
					},
				},
			},
			source:   model.NewSource("test-name", "test-source-type", nil),
			expected: false,
		},
		{
			name: "with processor update",
			updates: &EventUpdates{
				ProcessorsField: Events[*model.Processor]{
					"test-processor": Event[*model.Processor]{
						Item: model.NewProcessor("test-processor", "test-processor-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			source:   createSourceWithProcessor("test-source-type", "test-name", "test-processor-type", "test-processor"),
			expected: true,
		},
		{
			name: "with processor insert",
			updates: &EventUpdates{
				ProcessorsField: Events[*model.Processor]{
					"test-processor": Event[*model.Processor]{
						Item: model.NewProcessor("test-processor", "test-processor-type", nil),
						Type: EventTypeInsert,
					},
				},
			},
			source:   createSourceWithProcessor("test-source-type", "test-name", "test-processor-type", "test-processor"),
			expected: false,
		},
		{
			name: "with processor type update",
			updates: &EventUpdates{
				ProcessorTypesField: Events[*model.ProcessorType]{
					"test-processor-type": Event[*model.ProcessorType]{
						Item: model.NewProcessorType("test-processor-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			source:   createSourceWithProcessor("test-source-type", "test-name", "test-processor-type", "test-processor"),
			expected: true,
		},
		{
			name: "with processor type insert",
			updates: &EventUpdates{
				ProcessorTypesField: Events[*model.ProcessorType]{
					"test-processor-type": Event[*model.ProcessorType]{
						Item: model.NewProcessorType("test-processor-type", nil),
						Type: EventTypeInsert,
					},
				},
			},
			source:   createSourceWithProcessor("test-source-type", "test-name", "test-processor-type", "test-processor"),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.updates.AffectsSource(tc.source))
		})
	}
}

func TestUpdatesAffectsProcessor(t *testing.T) {
	testCase := []struct {
		name      string
		updates   BasicEventUpdates
		processor *model.Processor
		expected  bool
	}{
		{
			name:      "empty",
			updates:   NewEventUpdates(),
			processor: &model.Processor{},
			expected:  false,
		},
		{
			name: "with processor type update",
			updates: &EventUpdates{
				ProcessorTypesField: Events[*model.ProcessorType]{
					"test-processor-type": Event[*model.ProcessorType]{
						Item: model.NewProcessorType("test-processor-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			processor: model.NewProcessor("test-name", "test-processor-type", nil),
			expected:  true,
		},
		{
			name: "with processor type insert",
			updates: &EventUpdates{
				ProcessorTypesField: Events[*model.ProcessorType]{
					"test-processor-type": Event[*model.ProcessorType]{
						Item: model.NewProcessorType("test-processor-type", nil),
						Type: EventTypeInsert,
					},
				},
			},
			processor: model.NewProcessor("test-name", "test-processor-type", nil),
			expected:  false,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.updates.AffectsProcessor(tc.processor))
		})
	}
}

func TestUpdatesAffectsDestination(t *testing.T) {
	testCases := []struct {
		name        string
		updates     BasicEventUpdates
		destination *model.Destination
		expected    bool
	}{
		{
			name:        "empty",
			updates:     NewEventUpdates(),
			destination: &model.Destination{},
			expected:    false,
		},
		{
			name: "with destination type update",
			updates: &EventUpdates{
				DestinationTypesField: Events[*model.DestinationType]{
					"test-destination-type": Event[*model.DestinationType]{
						Item: model.NewDestinationType("test-destination-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			destination: model.NewDestination("test-name", "test-destination-type", nil),
			expected:    true,
		},
		{
			name: "with destination type insert",
			updates: &EventUpdates{
				DestinationTypesField: Events[*model.DestinationType]{
					"test-destination-type": Event[*model.DestinationType]{
						Item: model.NewDestinationType("test-destination-type", nil),
						Type: EventTypeInsert,
					},
				},
			},
			destination: model.NewDestination("test-name", "test-destination-type", nil),
			expected:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.updates.AffectsDestination(tc.destination))
		})
	}
}

func TestUpdatesAffectsConfiguration(t *testing.T) {
	testCases := []struct {
		name          string
		updates       BasicEventUpdates
		configuration *model.Configuration
		expected      bool
	}{
		{
			name:          "empty",
			updates:       NewEventUpdates(),
			configuration: &model.Configuration{},
			expected:      false,
		},
		{
			name: "with source update",
			updates: &EventUpdates{
				SourcesField: Events[*model.Source]{
					"test-source": Event[*model.Source]{
						Item: model.NewSource("test-source", "test-source-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			configuration: createConfigurationWithSource("test-config", "test-source-type", "test-source"),
			expected:      true,
		},
		{
			name: "with unrelated source update",
			updates: &EventUpdates{
				SourcesField: Events[*model.Source]{
					"unrelated-source": Event[*model.Source]{
						Item: model.NewSource("unrelated-source", "unrelated-source-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			configuration: createConfigurationWithSource("test-config", "test-source-type", "test-source"),
			expected:      false,
		},
		{
			name: "with source type update",
			updates: &EventUpdates{
				SourceTypesField: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil, []string{"macos", "linux", "windows"}),
						Type: EventTypeUpdate,
					},
				},
			},
			configuration: createConfigurationWithSource("test-config", "test-source-type", "test-source"),
			expected:      true,
		},
		{
			name: "with unrelated source type update",
			updates: &EventUpdates{
				SourceTypesField: Events[*model.SourceType]{
					"unrelated-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("unrelated-source-type", nil, []string{"macos", "linux", "windows"}),
						Type: EventTypeUpdate,
					},
				},
			},
			configuration: createConfigurationWithSource("test-config", "test-source-type", "test-source"),
			expected:      false,
		},
		{
			name: "with destination update",
			updates: &EventUpdates{
				DestinationsField: Events[*model.Destination]{
					"test-destination": Event[*model.Destination]{
						Item: model.NewDestination("test-destination", "test-destination-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			configuration: createConfigurationWithDestination("test-config", "test-destination-type", "test-destination"),
			expected:      true,
		},
		{
			name: "with unrelated destination update",
			updates: &EventUpdates{
				DestinationsField: Events[*model.Destination]{
					"unrelated-destination": Event[*model.Destination]{
						Item: model.NewDestination("unrelated-destination", "unrelated-destination-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			configuration: createConfigurationWithDestination("test-config", "test-destination-type", "test-destination"),
			expected:      false,
		},
		{
			name: "with destination type update",
			updates: &EventUpdates{
				DestinationTypesField: Events[*model.DestinationType]{
					"test-destination-type": Event[*model.DestinationType]{
						Item: model.NewDestinationType("test-destination-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			configuration: createConfigurationWithDestination("test-config", "test-destination-type", "test-destination"),
			expected:      true,
		},
		{
			name: "with unrelated destination type update",
			updates: &EventUpdates{
				DestinationTypesField: Events[*model.DestinationType]{
					"unrelated-destination-type": Event[*model.DestinationType]{
						Item: model.NewDestinationType("unrelated-destination-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			configuration: createConfigurationWithDestination("test-config", "test-destination-type", "test-destination"),
			expected:      false,
		},
		{
			name: "with destination processor type update",
			updates: &EventUpdates{
				ProcessorTypesField: Events[*model.ProcessorType]{
					"test-processor-type": Event[*model.ProcessorType]{
						Item: model.NewProcessorType("test-processor-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			configuration: createConfigurationWithDestinationProcessors("test-config", "test-destination-type", "test-processor-type"),
			expected:      true,
		},
		{
			name: "with unrelated destination processor type update",
			updates: &EventUpdates{
				ProcessorTypesField: Events[*model.ProcessorType]{
					"unrelated-processor-type": Event[*model.ProcessorType]{
						Item: model.NewProcessorType("unrelated-processor-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			configuration: createConfigurationWithDestinationProcessors("test-config", "test-destination-type", "test-processor-type"),
			expected:      false,
		},
		{
			name: "with source processor type update",
			updates: &EventUpdates{
				ProcessorTypesField: Events[*model.ProcessorType]{
					"test-processor-type": Event[*model.ProcessorType]{
						Item: model.NewProcessorType("test-processor-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			configuration: createConfigurationWithSourceProcessors("test-config", "test-source-type", "test-processor-type"),
			expected:      true,
		},
		{
			name: "with unrelated source processor type update",
			updates: &EventUpdates{
				ProcessorTypesField: Events[*model.ProcessorType]{
					"unrelated-processor-type": Event[*model.ProcessorType]{
						Item: model.NewProcessorType("unrelated-processor-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			configuration: createConfigurationWithSourceProcessors("test-config", "test-source-type", "test-processor-type"),
			expected:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.updates.AffectsConfiguration(tc.configuration))
		})
	}
}

func TestUpdatesAddAffectedSources(t *testing.T) {
	testCases := []struct {
		name     string
		updates  BasicEventUpdates
		sources  []*model.Source
		expected BasicEventUpdates
	}{
		{
			name:     "empty",
			updates:  NewEventUpdates(),
			sources:  []*model.Source{},
			expected: NewEventUpdates(),
		},
		{
			name: "with affected source",
			updates: &EventUpdates{
				SourceTypesField: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil, []string{"macos", "linux", "windows"}),
						Type: EventTypeUpdate,
					},
				},
			},
			sources: []*model.Source{model.NewSource("test-source", "test-source-type", nil)},
			expected: &EventUpdates{
				SourceTypesField: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil, []string{"macos", "linux", "windows"}),
						Type: EventTypeUpdate,
					},
				},
				SourcesField: Events[*model.Source]{
					"test-source": Event[*model.Source]{
						Item: model.NewSource("test-source", "test-source-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
		},
		{
			name: "with unrelated source",
			updates: &EventUpdates{
				SourceTypesField: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil, []string{"macos", "linux", "windows"}),
						Type: EventTypeUpdate,
					},
				},
			},
			sources: []*model.Source{model.NewSource("unrelated-source", "unrelated-source-type", nil)},
			expected: &EventUpdates{
				SourceTypesField: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil, []string{"macos", "linux", "windows"}),
						Type: EventTypeUpdate,
					},
				},
			},
		},
		{
			name: "with existing source update",
			updates: &EventUpdates{
				SourceTypesField: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil, []string{"macos", "linux", "windows"}),
						Type: EventTypeUpdate,
					},
				},
				SourcesField: Events[*model.Source]{
					"test-source": Event[*model.Source]{
						Item: model.NewSource("test-source", "test-source-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			sources: []*model.Source{model.NewSource("test-source", "test-source-type", nil)},
			expected: &EventUpdates{
				SourceTypesField: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil, []string{"macos", "linux", "windows"}),
						Type: EventTypeUpdate,
					},
				},
				SourcesField: Events[*model.Source]{
					"test-source": Event[*model.Source]{
						Item: model.NewSource("test-source", "test-source-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.updates.AddAffectedSources(tc.sources)

			assertEventsEqual(t, tc.expected.Sources(), tc.updates.Sources())
			assertEventsEqual(t, tc.expected.SourceTypes(), tc.updates.SourceTypes())
		})
	}
}

func TestUpdatesAddAffectedProcessors(t *testing.T) {
	testCases := []struct {
		name       string
		updates    BasicEventUpdates
		processors []*model.Processor
		expected   BasicEventUpdates
	}{
		{
			name:       "empty",
			updates:    NewEventUpdates(),
			processors: []*model.Processor{},
			expected:   NewEventUpdates(),
		},
		{
			name: "with affected processor",
			updates: &EventUpdates{
				ProcessorTypesField: Events[*model.ProcessorType]{
					"test-processor-type": Event[*model.ProcessorType]{
						Item: model.NewProcessorType("test-processor-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			processors: []*model.Processor{model.NewProcessor("test-processor", "test-processor-type", nil)},
			expected: &EventUpdates{
				ProcessorTypesField: Events[*model.ProcessorType]{
					"test-processor-type": Event[*model.ProcessorType]{
						Item: model.NewProcessorType("test-processor-type", nil),
						Type: EventTypeUpdate,
					},
				},
				ProcessorsField: Events[*model.Processor]{
					"test-processor": Event[*model.Processor]{
						Item: model.NewProcessor("test-processor", "test-processor-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
		},
		{
			name: "with unrelated processor",
			updates: &EventUpdates{
				ProcessorTypesField: Events[*model.ProcessorType]{
					"test-processor-type": Event[*model.ProcessorType]{
						Item: model.NewProcessorType("test-processor-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			processors: []*model.Processor{model.NewProcessor("unrelated-processor", "unrelated-processor-type", nil)},
			expected: &EventUpdates{
				ProcessorTypesField: Events[*model.ProcessorType]{
					"test-processor-type": Event[*model.ProcessorType]{
						Item: model.NewProcessorType("test-processor-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
		},
		{
			name: "with existing processor update",
			updates: &EventUpdates{
				ProcessorTypesField: Events[*model.ProcessorType]{
					"test-processor-type": Event[*model.ProcessorType]{
						Item: model.NewProcessorType("test-processor-type", nil),
						Type: EventTypeUpdate,
					},
				},
				ProcessorsField: Events[*model.Processor]{
					"test-processor": Event[*model.Processor]{
						Item: model.NewProcessor("test-processor", "test-processor-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			processors: []*model.Processor{model.NewProcessor("test-processor", "test-processor-type", nil)},
			expected: &EventUpdates{
				ProcessorTypesField: Events[*model.ProcessorType]{
					"test-processor-type": Event[*model.ProcessorType]{
						Item: model.NewProcessorType("test-processor-type", nil),
						Type: EventTypeUpdate,
					},
				},
				ProcessorsField: Events[*model.Processor]{
					"test-processor": Event[*model.Processor]{
						Item: model.NewProcessor("test-processor", "test-processor-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.updates.AddAffectedProcessors(tc.processors)
			assertEventsEqual(t, tc.expected.Processors(), tc.updates.Processors())
			assertEventsEqual(t, tc.expected.ProcessorTypes(), tc.updates.ProcessorTypes())
		})
	}
}

func TestUpdatesAddAffectedDestinations(t *testing.T) {
	testCases := []struct {
		name         string
		updates      BasicEventUpdates
		destinations []*model.Destination
		expected     BasicEventUpdates
	}{
		{
			name:         "empty",
			updates:      NewEventUpdates(),
			destinations: []*model.Destination{},
			expected:     NewEventUpdates(),
		},
		{
			name: "with affected destination",
			updates: &EventUpdates{
				DestinationTypesField: Events[*model.DestinationType]{
					"test-destination-type": Event[*model.DestinationType]{
						Item: model.NewDestinationType("test-destination-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			destinations: []*model.Destination{model.NewDestination("test-destination", "test-destination-type", nil)},
			expected: &EventUpdates{
				DestinationTypesField: Events[*model.DestinationType]{
					"test-destination-type": Event[*model.DestinationType]{
						Item: model.NewDestinationType("test-destination-type", nil),
						Type: EventTypeUpdate,
					},
				},
				DestinationsField: Events[*model.Destination]{
					"test-destination": Event[*model.Destination]{
						Item: model.NewDestination("test-destination", "test-destination-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
		},
		{
			name: "with unrelated destination",
			updates: &EventUpdates{
				DestinationTypesField: Events[*model.DestinationType]{
					"test-destination-type": Event[*model.DestinationType]{
						Item: model.NewDestinationType("test-destination-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			destinations: []*model.Destination{model.NewDestination("unrelated-destination", "unrelated-destination-type", nil)},
			expected: &EventUpdates{
				DestinationTypesField: Events[*model.DestinationType]{
					"test-destination-type": Event[*model.DestinationType]{
						Item: model.NewDestinationType("test-destination-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
		},
		{
			name: "with existing destination update",
			updates: &EventUpdates{
				DestinationTypesField: Events[*model.DestinationType]{
					"test-destination-type": Event[*model.DestinationType]{
						Item: model.NewDestinationType("test-destination-type", nil),
						Type: EventTypeUpdate,
					},
				},
				DestinationsField: Events[*model.Destination]{
					"test-destination": Event[*model.Destination]{
						Item: model.NewDestination("test-destination", "test-destination-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			destinations: []*model.Destination{model.NewDestination("test-destination", "test-destination-type", nil)},
			expected: &EventUpdates{
				DestinationTypesField: Events[*model.DestinationType]{
					"test-destination-type": Event[*model.DestinationType]{
						Item: model.NewDestinationType("test-destination-type", nil),
						Type: EventTypeUpdate,
					},
				},
				DestinationsField: Events[*model.Destination]{
					"test-destination": Event[*model.Destination]{
						Item: model.NewDestination("test-destination", "test-destination-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.updates.AddAffectedDestinations(tc.destinations)
			assertEventsEqual(t, tc.expected.Destinations(), tc.updates.Destinations())
			assertEventsEqual(t, tc.expected.DestinationTypes(), tc.updates.DestinationTypes())
		})
	}
}

func TestUpdatesAddAffectedConfigurations(t *testing.T) {
	testCases := []struct {
		name           string
		updates        BasicEventUpdates
		configurations []*model.Configuration
		expected       BasicEventUpdates
	}{
		{
			name:           "empty",
			updates:        NewEventUpdates(),
			configurations: []*model.Configuration{},
			expected:       NewEventUpdates(),
		},
		{
			name: "with affected configuration",
			updates: &EventUpdates{
				SourceTypesField: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil, []string{"macos", "linux", "windows"}),
						Type: EventTypeUpdate,
					},
				},
			},
			configurations: []*model.Configuration{
				createConfigurationWithSource("test-configuration", "test-source-type", "test-source-name"),
			},
			expected: &EventUpdates{
				SourceTypesField: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil, []string{"macos", "linux", "windows"}),
						Type: EventTypeUpdate,
					},
				},
				ConfigurationsField: Events[*model.Configuration]{
					"test-configuration": Event[*model.Configuration]{
						Item: createConfigurationWithSource("test-configuration", "test-source-type", "test-source-name"),
						Type: EventTypeUpdate,
					},
				},
			},
		},
		{
			name: "with unrelated configuration",
			updates: &EventUpdates{
				SourceTypesField: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil, []string{"macos", "linux", "windows"}),
						Type: EventTypeUpdate,
					},
				},
			},
			configurations: []*model.Configuration{
				createConfigurationWithSource("unrelated-configuration", "unrelated-source-type", "unrelated-source-name"),
			},
			expected: &EventUpdates{
				SourceTypesField: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil, []string{"macos", "linux", "windows"}),
						Type: EventTypeUpdate,
					},
				},
			},
		},
		{
			name: "with existing configuration update",
			updates: &EventUpdates{
				SourceTypesField: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil, []string{"macos", "linux", "windows"}),
						Type: EventTypeUpdate,
					},
				},
				ConfigurationsField: Events[*model.Configuration]{
					"test-configuration": Event[*model.Configuration]{
						Item: createConfigurationWithSource("test-configuration", "test-source-type", "test-source-name"),
						Type: EventTypeUpdate,
					},
				},
			},
			configurations: []*model.Configuration{
				createConfigurationWithSource("test-configuration", "test-source-type", "test-source-name"),
			},
			expected: &EventUpdates{
				SourceTypesField: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil, []string{"macos", "linux", "windows"}),
						Type: EventTypeUpdate,
					},
				},
				ConfigurationsField: Events[*model.Configuration]{
					"test-configuration": Event[*model.Configuration]{
						Item: createConfigurationWithSource("test-configuration", "test-source-type", "test-source-name"),
						Type: EventTypeUpdate,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.updates.AddAffectedConfigurations(tc.configurations)
			assertEventsEqual(t, tc.expected.Configurations(), tc.updates.Configurations())
			assertEventsEqual(t, tc.expected.SourceTypes(), tc.updates.SourceTypes())
		})
	}
}

func TestMergeUpdates(t *testing.T) {
	testCases := []struct {
		name       string
		into       BasicEventUpdates
		from       BasicEventUpdates
		expected   BasicEventUpdates
		successful bool
	}{
		{
			name:       "empty",
			into:       NewEventUpdates(),
			from:       NewEventUpdates(),
			expected:   NewEventUpdates(),
			successful: true,
		},
		{
			name: "with no conflicts",
			into: NewEventUpdates(),
			from: &EventUpdates{
				SourceTypesField: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil, []string{"macos", "linux", "windows"}),
						Type: EventTypeUpdate,
					},
				},
			},
			expected: &EventUpdates{
				AgentsField:        Events[*model.Agent]{},
				AgentVersionsField: Events[*model.AgentVersion]{},
				SourceTypesField: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil, []string{"macos", "linux", "windows"}),
						Type: EventTypeUpdate,
					},
				},
				SourcesField:          Events[*model.Source]{},
				ProcessorTypesField:   Events[*model.ProcessorType]{},
				ProcessorsField:       Events[*model.Processor]{},
				DestinationTypesField: Events[*model.DestinationType]{},
				DestinationsField:     Events[*model.Destination]{},
				ConfigurationsField:   Events[*model.Configuration]{},
			},
			successful: true,
		},
		{
			name: "with conflicts",
			into: &EventUpdates{
				SourceTypesField: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil, []string{"macos", "linux", "windows"}),
						Type: EventTypeUpdate,
					},
				},
			},
			from: &EventUpdates{
				SourceTypesField: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil, []string{"macos", "linux", "windows"}),
						Type: EventTypeInsert,
					},
				},
			},
			expected: &EventUpdates{
				SourceTypesField: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil, []string{"macos", "linux", "windows"}),
						Type: EventTypeUpdate,
					},
				},
			},
			successful: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			successful := MergeUpdates(tc.into, tc.from)
			require.Equal(t, tc.successful, successful)
			assertEventsEqual(t, tc.expected.SourceTypes(), tc.into.SourceTypes())
		})
	}
}

// createSourceWithProcessor creates a source with a processor for testing.
func createSourceWithProcessor(sourceType, sourceName, processorType, processorName string) *model.Source {
	return model.NewSourceWithSpec(sourceName, model.ParameterizedSpec{
		Type: sourceType,
		Processors: []model.ResourceConfiguration{
			{
				Name: processorName,
				ParameterizedSpec: model.ParameterizedSpec{
					Type: processorType,
				},
			},
		},
	})
}

// createConfigurationWithSource creates a configuration with a source for testing.
func createConfigurationWithSource(configName, sourceType, sourceName string) *model.Configuration {
	return model.NewConfigurationWithSpec(configName, model.ConfigurationSpec{
		Sources: []model.ResourceConfiguration{
			{
				Name: sourceName,
				ParameterizedSpec: model.ParameterizedSpec{
					Type: sourceType,
				},
			},
		},
	})
}

// createConfigurationWithSource creates a configuration with a source for testing.
func createConfigurationWithSourceProcessors(configName, sourceType, processorType string) *model.Configuration {
	return model.NewConfigurationWithSpec(configName, model.ConfigurationSpec{
		Sources: []model.ResourceConfiguration{
			{
				ParameterizedSpec: model.ParameterizedSpec{
					Type: sourceType,
					Processors: []model.ResourceConfiguration{
						{
							ParameterizedSpec: model.ParameterizedSpec{
								Type: processorType,
							},
						},
					},
				},
			},
		},
	})
}

// createConfigurationWithDestination creates a configuration with a destination for testing.
func createConfigurationWithDestination(configName, destinationType, destinationName string) *model.Configuration {
	return model.NewConfigurationWithSpec(configName, model.ConfigurationSpec{
		Destinations: []model.ResourceConfiguration{
			{
				Name: destinationName,
				ParameterizedSpec: model.ParameterizedSpec{
					Type: destinationType,
				},
			},
		},
	})
}

// createConfigurationWithDestination creates a configuration with a destination for testing.
func createConfigurationWithDestinationProcessors(configName, destinationType, processorType string) *model.Configuration {
	return model.NewConfigurationWithSpec(configName, model.ConfigurationSpec{
		Destinations: []model.ResourceConfiguration{
			{
				ParameterizedSpec: model.ParameterizedSpec{
					Type: destinationType,
					Processors: []model.ResourceConfiguration{
						{
							ParameterizedSpec: model.ParameterizedSpec{
								Type: processorType,
							},
						},
					},
				},
			},
		},
	})
}
