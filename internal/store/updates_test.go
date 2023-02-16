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

package store

import (
	"testing"

	"github.com/observiq/bindplane-op/model"
	"github.com/stretchr/testify/require"
)

func TestUpdatesIncludeAgent(t *testing.T) {
	updates := &Updates{}
	agent := &model.Agent{
		ID: "test",
	}

	updates.IncludeAgent(agent, EventTypeInsert)
	require.Equal(t, 1, len(updates.Agents))
	require.Equal(t, agent, updates.Agents[agent.UniqueKey()].Item)
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

	updates := &Updates{}
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
	require.Equal(t, agentVersion, updates.AgentVersions[agentVersion.UniqueKey()].Item)
	require.Equal(t, source, updates.Sources[source.UniqueKey()].Item)
	require.Equal(t, sourceType, updates.SourceTypes[sourceType.UniqueKey()].Item)
	require.Equal(t, processor, updates.Processors[processor.UniqueKey()].Item)
	require.Equal(t, processorType, updates.ProcessorTypes[processorType.UniqueKey()].Item)
	require.Equal(t, destination, updates.Destinations[destination.UniqueKey()].Item)
	require.Equal(t, destinationType, updates.DestinationTypes[destinationType.UniqueKey()].Item)
	require.Equal(t, configuration, updates.Configurations[configuration.UniqueKey()].Item)
}

func TestUpdatesEmpty(t *testing.T) {
	testCases := []struct {
		name     string
		updates  *Updates
		expected bool
	}{
		{
			name:     "empty",
			updates:  NewUpdates(),
			expected: true,
		},
		{
			name: "not empty",
			updates: &Updates{
				Agents: Events[*model.Agent]{
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
		updates  *Updates
		expected int
	}{
		{
			name:     "empty",
			updates:  NewUpdates(),
			expected: 0,
		},
		{
			name: "with resources",
			updates: &Updates{
				Agents: Events[*model.Agent]{
					"test": Event[*model.Agent]{},
				},
				AgentVersions: Events[*model.AgentVersion]{
					"test": Event[*model.AgentVersion]{},
				},
				Sources: Events[*model.Source]{
					"test": Event[*model.Source]{},
				},
				SourceTypes: Events[*model.SourceType]{
					"test": Event[*model.SourceType]{},
				},
				Processors: Events[*model.Processor]{
					"test": Event[*model.Processor]{},
				},
				ProcessorTypes: Events[*model.ProcessorType]{
					"test": Event[*model.ProcessorType]{},
				},
				Destinations: Events[*model.Destination]{
					"test": Event[*model.Destination]{},
				},
				DestinationTypes: Events[*model.DestinationType]{
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
		updates  *Updates
		expected bool
	}{
		{
			name:     "empty",
			updates:  NewUpdates(),
			expected: false,
		},
		{
			name: "with processors",
			updates: &Updates{
				Processors: Events[*model.Processor]{
					"test": Event[*model.Processor]{},
				},
			},
			expected: false,
		},
		{
			name: "with processor types",
			updates: &Updates{
				ProcessorTypes: Events[*model.ProcessorType]{
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
		updates  *Updates
		expected bool
	}{
		{
			name:     "empty",
			updates:  NewUpdates(),
			expected: false,
		},
		{
			name: "with sources",
			updates: &Updates{
				Sources: Events[*model.Source]{
					"test": Event[*model.Source]{},
				},
			},
			expected: false,
		},
		{
			name: "with source types",
			updates: &Updates{
				SourceTypes: Events[*model.SourceType]{
					"test": Event[*model.SourceType]{},
				},
			},
			expected: true,
		},
		{
			name: "with processors",
			updates: &Updates{
				Processors: Events[*model.Processor]{
					"test": Event[*model.Processor]{},
				},
			},
			expected: true,
		},
		{
			name: "with processor types",
			updates: &Updates{
				ProcessorTypes: Events[*model.ProcessorType]{
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
		updates  *Updates
		expected bool
	}{
		{
			name:     "empty",
			updates:  NewUpdates(),
			expected: false,
		},
		{
			name: "with destinations",
			updates: &Updates{
				Destinations: Events[*model.Destination]{
					"test": Event[*model.Destination]{},
				},
			},
			expected: false,
		},
		{
			name: "with destination types",
			updates: &Updates{
				DestinationTypes: Events[*model.DestinationType]{
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
		updates  *Updates
		expected bool
	}{
		{
			name:     "empty",
			updates:  NewUpdates(),
			expected: false,
		},
		{
			name: "with sources",
			updates: &Updates{
				Sources: Events[*model.Source]{
					"test": Event[*model.Source]{},
				},
			},
			expected: true,
		},
		{
			name: "with source types",
			updates: &Updates{
				SourceTypes: Events[*model.SourceType]{
					"test": Event[*model.SourceType]{},
				},
			},
			expected: true,
		},
		{
			name: "with destinations",
			updates: &Updates{
				Destinations: Events[*model.Destination]{
					"test": Event[*model.Destination]{},
				},
			},
			expected: true,
		},
		{
			name: "with destination types",
			updates: &Updates{
				DestinationTypes: Events[*model.DestinationType]{
					"test": Event[*model.DestinationType]{},
				},
			},
			expected: true,
		},
		{
			name: "with processors",
			updates: &Updates{
				Processors: Events[*model.Processor]{
					"test": Event[*model.Processor]{},
				},
			},
			expected: true,
		},
		{
			name: "with processor types",
			updates: &Updates{
				ProcessorTypes: Events[*model.ProcessorType]{
					"test": Event[*model.ProcessorType]{},
				},
			},
			expected: true,
		},
		{
			name: "with configurations",
			updates: &Updates{
				Configurations: Events[*model.Configuration]{
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
		updates  *Updates
		source   *model.Source
		expected bool
	}{
		{
			name:     "empty",
			updates:  NewUpdates(),
			source:   &model.Source{},
			expected: false,
		},
		{
			name: "with source type update",
			updates: &Updates{
				SourceTypes: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			source:   model.NewSource("test-name", "test-source-type", nil),
			expected: true,
		},
		{
			name: "with source type insert",
			updates: &Updates{
				SourceTypes: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil),
						Type: EventTypeInsert,
					},
				},
			},
			source:   model.NewSource("test-name", "test-source-type", nil),
			expected: false,
		},
		{
			name: "with processor update",
			updates: &Updates{
				Processors: Events[*model.Processor]{
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
			updates: &Updates{
				Processors: Events[*model.Processor]{
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
			updates: &Updates{
				ProcessorTypes: Events[*model.ProcessorType]{
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
			updates: &Updates{
				ProcessorTypes: Events[*model.ProcessorType]{
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
		updates   *Updates
		processor *model.Processor
		expected  bool
	}{
		{
			name:      "empty",
			updates:   NewUpdates(),
			processor: &model.Processor{},
			expected:  false,
		},
		{
			name: "with processor type update",
			updates: &Updates{
				ProcessorTypes: Events[*model.ProcessorType]{
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
			updates: &Updates{
				ProcessorTypes: Events[*model.ProcessorType]{
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
		updates     *Updates
		destination *model.Destination
		expected    bool
	}{
		{
			name:        "empty",
			updates:     NewUpdates(),
			destination: &model.Destination{},
			expected:    false,
		},
		{
			name: "with destination type update",
			updates: &Updates{
				DestinationTypes: Events[*model.DestinationType]{
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
			updates: &Updates{
				DestinationTypes: Events[*model.DestinationType]{
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
		updates       *Updates
		configuration *model.Configuration
		expected      bool
	}{
		{
			name:          "empty",
			updates:       NewUpdates(),
			configuration: &model.Configuration{},
			expected:      false,
		},
		{
			name: "with source update",
			updates: &Updates{
				Sources: Events[*model.Source]{
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
			updates: &Updates{
				Sources: Events[*model.Source]{
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
			updates: &Updates{
				SourceTypes: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			configuration: createConfigurationWithSource("test-config", "test-source-type", "test-source"),
			expected:      true,
		},
		{
			name: "with unrelated source type update",
			updates: &Updates{
				SourceTypes: Events[*model.SourceType]{
					"unrelated-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("unrelated-source-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			configuration: createConfigurationWithSource("test-config", "test-source-type", "test-source"),
			expected:      false,
		},
		{
			name: "with destination update",
			updates: &Updates{
				Destinations: Events[*model.Destination]{
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
			updates: &Updates{
				Destinations: Events[*model.Destination]{
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
			updates: &Updates{
				DestinationTypes: Events[*model.DestinationType]{
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
			updates: &Updates{
				DestinationTypes: Events[*model.DestinationType]{
					"unrelated-destination-type": Event[*model.DestinationType]{
						Item: model.NewDestinationType("unrelated-destination-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			configuration: createConfigurationWithDestination("test-config", "test-destination-type", "test-destination"),
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
		updates  *Updates
		sources  []*model.Source
		expected *Updates
	}{
		{
			name:     "empty",
			updates:  NewUpdates(),
			sources:  []*model.Source{},
			expected: NewUpdates(),
		},
		{
			name: "with affected source",
			updates: &Updates{
				SourceTypes: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			sources: []*model.Source{model.NewSource("test-source", "test-source-type", nil)},
			expected: &Updates{
				SourceTypes: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil),
						Type: EventTypeUpdate,
					},
				},
				Sources: Events[*model.Source]{
					"test-source": Event[*model.Source]{
						Item: model.NewSource("test-source", "test-source-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
		},
		{
			name: "with unrelated source",
			updates: &Updates{
				SourceTypes: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			sources: []*model.Source{model.NewSource("unrelated-source", "unrelated-source-type", nil)},
			expected: &Updates{
				SourceTypes: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
		},
		{
			name: "with existing source update",
			updates: &Updates{
				SourceTypes: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil),
						Type: EventTypeUpdate,
					},
				},
				Sources: Events[*model.Source]{
					"test-source": Event[*model.Source]{
						Item: model.NewSource("test-source", "test-source-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			sources: []*model.Source{model.NewSource("test-source", "test-source-type", nil)},
			expected: &Updates{
				SourceTypes: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil),
						Type: EventTypeUpdate,
					},
				},
				Sources: Events[*model.Source]{
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
			require.Equal(t, tc.expected, tc.updates)
		})
	}
}

func TestUpdatesAddAffectedProcessors(t *testing.T) {
	testCases := []struct {
		name       string
		updates    *Updates
		processors []*model.Processor
		expected   *Updates
	}{
		{
			name:       "empty",
			updates:    NewUpdates(),
			processors: []*model.Processor{},
			expected:   NewUpdates(),
		},
		{
			name: "with affected processor",
			updates: &Updates{
				ProcessorTypes: Events[*model.ProcessorType]{
					"test-processor-type": Event[*model.ProcessorType]{
						Item: model.NewProcessorType("test-processor-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			processors: []*model.Processor{model.NewProcessor("test-processor", "test-processor-type", nil)},
			expected: &Updates{
				ProcessorTypes: Events[*model.ProcessorType]{
					"test-processor-type": Event[*model.ProcessorType]{
						Item: model.NewProcessorType("test-processor-type", nil),
						Type: EventTypeUpdate,
					},
				},
				Processors: Events[*model.Processor]{
					"test-processor": Event[*model.Processor]{
						Item: model.NewProcessor("test-processor", "test-processor-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
		},
		{
			name: "with unrelated processor",
			updates: &Updates{
				ProcessorTypes: Events[*model.ProcessorType]{
					"test-processor-type": Event[*model.ProcessorType]{
						Item: model.NewProcessorType("test-processor-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			processors: []*model.Processor{model.NewProcessor("unrelated-processor", "unrelated-processor-type", nil)},
			expected: &Updates{
				ProcessorTypes: Events[*model.ProcessorType]{
					"test-processor-type": Event[*model.ProcessorType]{
						Item: model.NewProcessorType("test-processor-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
		},
		{
			name: "with existing processor update",
			updates: &Updates{
				ProcessorTypes: Events[*model.ProcessorType]{
					"test-processor-type": Event[*model.ProcessorType]{
						Item: model.NewProcessorType("test-processor-type", nil),
						Type: EventTypeUpdate,
					},
				},
				Processors: Events[*model.Processor]{
					"test-processor": Event[*model.Processor]{
						Item: model.NewProcessor("test-processor", "test-processor-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			processors: []*model.Processor{model.NewProcessor("test-processor", "test-processor-type", nil)},
			expected: &Updates{
				ProcessorTypes: Events[*model.ProcessorType]{
					"test-processor-type": Event[*model.ProcessorType]{
						Item: model.NewProcessorType("test-processor-type", nil),
						Type: EventTypeUpdate,
					},
				},
				Processors: Events[*model.Processor]{
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
			require.Equal(t, tc.expected, tc.updates)
		})
	}
}

func TestUpdatesAddAffectedDestinations(t *testing.T) {
	testCases := []struct {
		name         string
		updates      *Updates
		destinations []*model.Destination
		expected     *Updates
	}{
		{
			name:         "empty",
			updates:      NewUpdates(),
			destinations: []*model.Destination{},
			expected:     NewUpdates(),
		},
		{
			name: "with affected destination",
			updates: &Updates{
				DestinationTypes: Events[*model.DestinationType]{
					"test-destination-type": Event[*model.DestinationType]{
						Item: model.NewDestinationType("test-destination-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			destinations: []*model.Destination{model.NewDestination("test-destination", "test-destination-type", nil)},
			expected: &Updates{
				DestinationTypes: Events[*model.DestinationType]{
					"test-destination-type": Event[*model.DestinationType]{
						Item: model.NewDestinationType("test-destination-type", nil),
						Type: EventTypeUpdate,
					},
				},
				Destinations: Events[*model.Destination]{
					"test-destination": Event[*model.Destination]{
						Item: model.NewDestination("test-destination", "test-destination-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
		},
		{
			name: "with unrelated destination",
			updates: &Updates{
				DestinationTypes: Events[*model.DestinationType]{
					"test-destination-type": Event[*model.DestinationType]{
						Item: model.NewDestinationType("test-destination-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			destinations: []*model.Destination{model.NewDestination("unrelated-destination", "unrelated-destination-type", nil)},
			expected: &Updates{
				DestinationTypes: Events[*model.DestinationType]{
					"test-destination-type": Event[*model.DestinationType]{
						Item: model.NewDestinationType("test-destination-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
		},
		{
			name: "with existing destination update",
			updates: &Updates{
				DestinationTypes: Events[*model.DestinationType]{
					"test-destination-type": Event[*model.DestinationType]{
						Item: model.NewDestinationType("test-destination-type", nil),
						Type: EventTypeUpdate,
					},
				},
				Destinations: Events[*model.Destination]{
					"test-destination": Event[*model.Destination]{
						Item: model.NewDestination("test-destination", "test-destination-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			destinations: []*model.Destination{model.NewDestination("test-destination", "test-destination-type", nil)},
			expected: &Updates{
				DestinationTypes: Events[*model.DestinationType]{
					"test-destination-type": Event[*model.DestinationType]{
						Item: model.NewDestinationType("test-destination-type", nil),
						Type: EventTypeUpdate,
					},
				},
				Destinations: Events[*model.Destination]{
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
			require.Equal(t, tc.expected, tc.updates)
		})
	}
}

func TestUpdatesAddAffectedConfigurations(t *testing.T) {
	testCases := []struct {
		name           string
		updates        *Updates
		configurations []*model.Configuration
		expected       *Updates
	}{
		{
			name:           "empty",
			updates:        NewUpdates(),
			configurations: []*model.Configuration{},
			expected:       NewUpdates(),
		},
		{
			name: "with affected configuration",
			updates: &Updates{
				SourceTypes: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			configurations: []*model.Configuration{
				createConfigurationWithSource("test-configuration", "test-source-type", "test-source-name"),
			},
			expected: &Updates{
				SourceTypes: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil),
						Type: EventTypeUpdate,
					},
				},
				Configurations: Events[*model.Configuration]{
					"test-configuration": Event[*model.Configuration]{
						Item: createConfigurationWithSource("test-configuration", "test-source-type", "test-source-name"),
						Type: EventTypeUpdate,
					},
				},
			},
		},
		{
			name: "with unrelated configuration",
			updates: &Updates{
				SourceTypes: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			configurations: []*model.Configuration{
				createConfigurationWithSource("unrelated-configuration", "unrelated-source-type", "unrelated-source-name"),
			},
			expected: &Updates{
				SourceTypes: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
		},
		{
			name: "with existing configuration update",
			updates: &Updates{
				SourceTypes: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil),
						Type: EventTypeUpdate,
					},
				},
				Configurations: Events[*model.Configuration]{
					"test-configuration": Event[*model.Configuration]{
						Item: createConfigurationWithSource("test-configuration", "test-source-type", "test-source-name"),
						Type: EventTypeUpdate,
					},
				},
			},
			configurations: []*model.Configuration{
				createConfigurationWithSource("test-configuration", "test-source-type", "test-source-name"),
			},
			expected: &Updates{
				SourceTypes: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil),
						Type: EventTypeUpdate,
					},
				},
				Configurations: Events[*model.Configuration]{
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
			require.Equal(t, tc.expected, tc.updates)
		})
	}
}

func TestMergeUpdates(t *testing.T) {
	testCases := []struct {
		name       string
		into       *Updates
		from       *Updates
		expected   *Updates
		successful bool
	}{
		{
			name:       "empty",
			into:       NewUpdates(),
			from:       NewUpdates(),
			expected:   NewUpdates(),
			successful: true,
		},
		{
			name: "with no conflicts",
			into: NewUpdates(),
			from: &Updates{
				SourceTypes: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			expected: &Updates{
				Agents:        Events[*model.Agent]{},
				AgentVersions: Events[*model.AgentVersion]{},
				SourceTypes: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil),
						Type: EventTypeUpdate,
					},
				},
				Sources:          Events[*model.Source]{},
				ProcessorTypes:   Events[*model.ProcessorType]{},
				Processors:       Events[*model.Processor]{},
				DestinationTypes: Events[*model.DestinationType]{},
				Destinations:     Events[*model.Destination]{},
				Configurations:   Events[*model.Configuration]{},
			},
			successful: true,
		},
		{
			name: "with conflicts",
			into: &Updates{
				SourceTypes: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil),
						Type: EventTypeUpdate,
					},
				},
			},
			from: &Updates{
				SourceTypes: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil),
						Type: EventTypeInsert,
					},
				},
			},
			expected: &Updates{
				SourceTypes: Events[*model.SourceType]{
					"test-source-type": Event[*model.SourceType]{
						Item: model.NewSourceType("test-source-type", nil),
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
			require.Equal(t, tc.expected, tc.into)
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
