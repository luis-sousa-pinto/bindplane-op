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
	"context"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/observiq/bindplane-op/eventbus/broadcast"
	"github.com/observiq/bindplane-op/model"
	"go.uber.org/zap"
)

// BasicEventUpdates is a collection of events created by a store operation.
type BasicEventUpdates interface {
	// Agents returns a collection of agent events.
	Agents() Events[*model.Agent]
	// AgentVersions returns a collection of agent version events.
	AgentVersions() Events[*model.AgentVersion]
	// Sources returns a collection of source events.
	Sources() Events[*model.Source]
	// SourceTypes returns a collection of source type events.
	SourceTypes() Events[*model.SourceType]
	// Processors returns a collection of processor events.
	Processors() Events[*model.Processor]
	// ProcessorTypes returns a collection of processor type events.
	ProcessorTypes() Events[*model.ProcessorType]
	// Destinations returns a collection of destination events.
	Destinations() Events[*model.Destination]
	// DestinationTypes returns a collection of destination type events.
	DestinationTypes() Events[*model.DestinationType]
	// Configurations returns a collection of configuration events.
	Configurations() Events[*model.Configuration]

	// IncludeResource will add a resource event to Updates.
	IncludeResource(r model.Resource, eventType EventType)
	// IncludeAgent will add an agent event to Updates.
	IncludeAgent(agent *model.Agent, eventType EventType)

	// CouldAffectConfigurations returns true if the updates could affect configurations.
	CouldAffectConfigurations() bool
	// CouldAffectDestinations returns true if the updates could affect destinations.
	CouldAffectDestinations() bool
	// CouldAffectProcessors returns true if the updates could affect processors.
	CouldAffectProcessors() bool
	// CouldAffectSources returns true if the updates could affect sources.
	CouldAffectSources() bool

	// AffectsDestination returns true if the updates affect the given destination.
	AffectsDestination(destination *model.Destination) bool
	// AffectsProcessor returns true if the updates affect the given processor.
	AffectsProcessor(processor *model.Processor) bool
	// AffectsSource returns true if the updates affect the given source.
	AffectsSource(source *model.Source) bool
	// AffectsConfiguration returns true if the updates affect the given configuration.
	AffectsConfiguration(configuration *model.Configuration) bool
	// AffectsResourceProcessors returns true if the updates affect any of the given resource processors.
	AffectsResourceProcessors(processors []model.ResourceConfiguration) bool

	// AddAffectedSources will add the given sources to the updates.
	AddAffectedSources(sources []*model.Source)
	// AddAffectedProcessors will add the given processors to the updates.
	AddAffectedProcessors(processors []*model.Processor)
	// AddAffectedDestinations will add the given destinations to the updates.
	AddAffectedDestinations(destinations []*model.Destination)
	// AddAffectedConfigurations will add the given configurations to the updates.
	AddAffectedConfigurations(configurations []*model.Configuration)

	// TransitiveUpdates returns a list of resources that need to have their resources updated.
	TransitiveUpdates() []model.Resource

	// Size returns the number of events in the updates.
	Size() int
	// Empty returns true if the updates are empty.
	Empty() bool
	// Merge merges another set of updates into this one, returns true
	// if it was able to merge any updates.
	Merge(other BasicEventUpdates) bool
}

// EventUpdates is a collection of events created by a store operation.
type EventUpdates struct {
	AgentsField           Events[*model.Agent]           `json:"agents"`
	AgentVersionsField    Events[*model.AgentVersion]    `json:"agentVersions"`
	SourcesField          Events[*model.Source]          `json:"sources"`
	SourceTypesField      Events[*model.SourceType]      `json:"sourceTypes"`
	ProcessorsField       Events[*model.Processor]       `json:"processors"`
	ProcessorTypesField   Events[*model.ProcessorType]   `json:"processorTypes"`
	DestinationsField     Events[*model.Destination]     `json:"destinations"`
	DestinationTypesField Events[*model.DestinationType] `json:"destinationTypes"`
	ConfigurationsField   Events[*model.Configuration]   `json:"configurations"`

	// transitiveUpdates is just used to track which resources need to have their resources updated
	transitiveUpdates []model.Resource
}

// TransitiveUpdates returns a list of resources that need to have their resources updated.
func (u *EventUpdates) TransitiveUpdates() []model.Resource {
	return u.transitiveUpdates
}

// Agents returns a collection of agent events.
func (u *EventUpdates) Agents() Events[*model.Agent] {
	return u.AgentsField
}

// AgentVersions returns a collection of agent version events.
func (u *EventUpdates) AgentVersions() Events[*model.AgentVersion] {
	return u.AgentVersionsField
}

// Sources returns a collection of source events.
func (u *EventUpdates) Sources() Events[*model.Source] {
	return u.SourcesField
}

// SourceTypes returns a collection of source type events.
func (u *EventUpdates) SourceTypes() Events[*model.SourceType] {
	return u.SourceTypesField
}

// Processors returns a collection of processor events.
func (u *EventUpdates) Processors() Events[*model.Processor] {
	return u.ProcessorsField
}

// ProcessorTypes returns a collection of processor type events.
func (u *EventUpdates) ProcessorTypes() Events[*model.ProcessorType] {
	return u.ProcessorTypesField
}

// Destinations returns a collection of destination events.
func (u *EventUpdates) Destinations() Events[*model.Destination] {
	return u.DestinationsField
}

// DestinationTypes returns a collection of destination type events.
func (u *EventUpdates) DestinationTypes() Events[*model.DestinationType] {
	return u.DestinationTypesField
}

// Configurations returns a collection of configuration events.
func (u *EventUpdates) Configurations() Events[*model.Configuration] {
	return u.ConfigurationsField
}

// IncludeAgent will add an agent event to Updates.
func (u *EventUpdates) IncludeAgent(agent *model.Agent, eventType EventType) {
	if u.AgentsField == nil {
		u.AgentsField = NewEvents[*model.Agent]()
	}

	u.AgentsField.Include(agent, eventType)
}

// IncludeAgentVersion will add an agent version event to Updates.
func (u *EventUpdates) IncludeAgentVersion(agentVersion *model.AgentVersion, eventType EventType) {
	if u.AgentVersionsField == nil {
		u.AgentVersionsField = NewEvents[*model.AgentVersion]()
	}

	u.AgentVersionsField.Include(agentVersion, eventType)
}

// IncludeSource will add a source event to Updates.
func (u *EventUpdates) IncludeSource(source *model.Source, eventType EventType) {
	if u.SourcesField == nil {
		u.SourcesField = NewEvents[*model.Source]()
	}

	u.SourcesField.Include(source, eventType)
}

// IncludeSourceType will add a source type event to Updates.
func (u *EventUpdates) IncludeSourceType(sourceType *model.SourceType, eventType EventType) {
	if u.SourceTypesField == nil {
		u.SourceTypesField = NewEvents[*model.SourceType]()
	}

	u.SourceTypesField.Include(sourceType, eventType)
}

// IncludeProcessor will add a processor event to Updates.
func (u *EventUpdates) IncludeProcessor(processor *model.Processor, eventType EventType) {
	if u.ProcessorsField == nil {
		u.ProcessorsField = NewEvents[*model.Processor]()
	}

	u.ProcessorsField.Include(processor, eventType)
}

// IncludeProcessorType will add a processor type event to Updates.
func (u *EventUpdates) IncludeProcessorType(processorType *model.ProcessorType, eventType EventType) {
	if u.ProcessorTypesField == nil {
		u.ProcessorTypesField = NewEvents[*model.ProcessorType]()
	}

	u.ProcessorTypesField.Include(processorType, eventType)
}

// IncludeDestination will add a destination event to Updates.
func (u *EventUpdates) IncludeDestination(destination *model.Destination, eventType EventType) {
	if u.DestinationsField == nil {
		u.DestinationsField = NewEvents[*model.Destination]()
	}

	u.DestinationsField.Include(destination, eventType)
}

// IncludeDestinationType will add a destination type event to Updates.
func (u *EventUpdates) IncludeDestinationType(destinationType *model.DestinationType, eventType EventType) {
	if u.DestinationTypesField == nil {
		u.DestinationTypesField = NewEvents[*model.DestinationType]()
	}

	u.DestinationTypesField.Include(destinationType, eventType)
}

// IncludeConfiguration will add a configuration event to Updates.
func (u *EventUpdates) IncludeConfiguration(configuration *model.Configuration, eventType EventType) {
	if u.ConfigurationsField == nil {
		u.ConfigurationsField = NewEvents[*model.Configuration]()
	}

	u.ConfigurationsField.Include(configuration, eventType)
}

// IncludeResource will add a resource event to Updates.
// If the resource is not supported by Updates, this will do nothing.
func (u *EventUpdates) IncludeResource(r model.Resource, eventType EventType) {
	switch r := r.(type) {
	case *model.AgentVersion:
		u.IncludeAgentVersion(r, eventType)
	case *model.Source:
		u.IncludeSource(r, eventType)
	case *model.SourceType:
		u.IncludeSourceType(r, eventType)
	case *model.Processor:
		u.IncludeProcessor(r, eventType)
	case *model.ProcessorType:
		u.IncludeProcessorType(r, eventType)
	case *model.Destination:
		u.IncludeDestination(r, eventType)
	case *model.DestinationType:
		u.IncludeDestinationType(r, eventType)
	case *model.Configuration:
		u.IncludeConfiguration(r, eventType)
	}
}

// Empty returns true if no events exist.
func (u *EventUpdates) Empty() bool {
	return u.Size() == 0
}

// Size returns the sum of all events.
func (u *EventUpdates) Size() int {
	return len(u.AgentsField) +
		len(u.AgentVersionsField) +
		len(u.SourcesField) +
		len(u.SourceTypesField) +
		len(u.ProcessorsField) +
		len(u.ProcessorTypesField) +
		len(u.DestinationsField) +
		len(u.DestinationTypesField) +
		len(u.ConfigurationsField)
}

// Merge merges another set of updates into this one, returns true
// if it was able to merge any updates.
func (u *EventUpdates) Merge(other BasicEventUpdates) bool {
	safe :=
		u.AgentsField.CanSafelyMerge(other.Agents()) &&
			u.AgentVersionsField.CanSafelyMerge(other.AgentVersions()) &&
			u.SourcesField.CanSafelyMerge(other.Sources()) &&
			u.SourceTypesField.CanSafelyMerge(other.SourceTypes()) &&
			u.ProcessorsField.CanSafelyMerge(other.Processors()) &&
			u.ProcessorTypesField.CanSafelyMerge(other.ProcessorTypes()) &&
			u.DestinationsField.CanSafelyMerge(other.Destinations()) &&
			u.DestinationTypesField.CanSafelyMerge(other.DestinationTypes()) &&
			u.ConfigurationsField.CanSafelyMerge(other.Configurations())

	if !safe {
		return false
	}

	u.AgentsField.Merge(other.Agents())
	u.AgentVersionsField.Merge(other.AgentVersions())
	u.SourcesField.Merge(other.Sources())
	u.SourceTypesField.Merge(other.SourceTypes())
	u.ProcessorsField.Merge(other.Processors())
	u.ProcessorTypesField.Merge(other.ProcessorTypes())
	u.DestinationsField.Merge(other.Destinations())
	u.DestinationTypesField.Merge(other.DestinationTypes())
	u.ConfigurationsField.Merge(other.Configurations())
	return true
}

// HasSourceTypeEvents returns true if any source type events exist.
func (u *EventUpdates) HasSourceTypeEvents() bool {
	return !u.SourceTypesField.Empty()
}

// HasProcessorTypeEvents returns true if any processor type events exist.
func (u *EventUpdates) HasProcessorTypeEvents() bool {
	return !u.ProcessorTypesField.Empty()
}

// HasDestinationTypeEvents returns true if any destination type events exist.
func (u *EventUpdates) HasDestinationTypeEvents() bool {
	return !u.DestinationTypesField.Empty()
}

// HasSourceEvents returns true if any source events exist.
func (u *EventUpdates) HasSourceEvents() bool {
	return !u.SourcesField.Empty()
}

// HasProcessorEvents returns true if any processor events exist.
func (u *EventUpdates) HasProcessorEvents() bool {
	return !u.ProcessorsField.Empty()
}

// HasDestinationEvents returns true if any destination events exist.
func (u *EventUpdates) HasDestinationEvents() bool {
	return !u.DestinationsField.Empty()
}

// CouldAffectProcessors returns true if the updates could affect processors.
func (u *EventUpdates) CouldAffectProcessors() bool {
	return u.HasProcessorTypeEvents()
}

// CouldAffectSources returns true if the updates could affect sources.
func (u *EventUpdates) CouldAffectSources() bool {
	return u.HasSourceTypeEvents() ||
		u.HasProcessorTypeEvents() ||
		u.HasProcessorEvents()
}

// CouldAffectDestinations returns true if the updates could affect destinations.
func (u *EventUpdates) CouldAffectDestinations() bool {
	return u.HasDestinationTypeEvents()
}

// CouldAffectConfigurations returns true if the updates could affect configurations.
func (u *EventUpdates) CouldAffectConfigurations() bool {
	return u.HasSourceTypeEvents() ||
		u.HasSourceEvents() ||
		u.HasProcessorTypeEvents() ||
		u.HasProcessorEvents() ||
		u.HasDestinationTypeEvents() ||
		u.HasDestinationEvents()
}

// AffectsSource returns true if the updates affect the given source.
func (u *EventUpdates) AffectsSource(source *model.Source) bool {
	return u.SourceTypesField.Contains(source.Spec.Type, EventTypeUpdate) ||
		u.AffectsResourceProcessors(source.Spec.Processors)
}

// AffectsProcessor returns true if the updates affect the given processor.
func (u *EventUpdates) AffectsProcessor(processor *model.Processor) bool {
	return u.ProcessorTypesField.Contains(processor.Spec.Type, EventTypeUpdate)
}

// AffectsDestination returns true if the updates affect the given destination.
func (u *EventUpdates) AffectsDestination(destination *model.Destination) bool {
	// DestinationType
	return u.DestinationTypesField.Contains(destination.Spec.Type, EventTypeUpdate) ||
		u.AffectsResourceProcessors(destination.Spec.Processors)
}

// AffectsResourceProcessors returns true if the updates affect any of the given resource processors.
func (u *EventUpdates) AffectsResourceProcessors(processors []model.ResourceConfiguration) bool {
	for _, processor := range processors {
		if u.ProcessorsField.Contains(processor.Name, EventTypeUpdate) ||
			u.ProcessorTypesField.Contains(processor.Type, EventTypeUpdate) {
			return true
		}
	}
	return false
}

// AffectsConfiguration returns true if the updates affect the given configuration.
func (u *EventUpdates) AffectsConfiguration(configuration *model.Configuration) bool {
	for _, source := range configuration.Spec.Sources {
		if u.SourcesField.ContainsKey(source.Name) ||
			u.SourceTypesField.ContainsKey(source.Type) ||
			u.AffectsResourceProcessors(source.Processors) {
			return true
		}
	}

	for _, destination := range configuration.Spec.Destinations {
		if u.DestinationsField.ContainsKey(destination.Name) ||
			u.DestinationTypesField.ContainsKey(destination.Type) ||
			u.AffectsResourceProcessors(destination.Processors) {
			return true
		}
	}
	return false
}

// AddAffectedSources will add updates for Sources that are affected by other resource updates.
func (u *EventUpdates) AddAffectedSources(sources []*model.Source) {
	for _, source := range sources {
		if u.SourcesField.Contains(source.Name(), EventTypeUpdate) {
			continue
		}

		if u.AffectsSource(source) {
			u.IncludeSource(source, EventTypeUpdate)
			u.transitiveUpdates = append(u.transitiveUpdates, source)
		}
	}
}

// AddAffectedProcessors will add updates for Processors that are affected by other resource updates.
func (u *EventUpdates) AddAffectedProcessors(processors []*model.Processor) {
	for _, processor := range processors {
		if u.ProcessorsField.Contains(processor.Name(), EventTypeUpdate) {
			continue
		}

		if u.AffectsProcessor(processor) {
			u.IncludeProcessor(processor, EventTypeUpdate)
			u.transitiveUpdates = append(u.transitiveUpdates, processor)
		}
	}
}

// AddAffectedDestinations will add updates for Destinations that are affected by other resource updates.
func (u *EventUpdates) AddAffectedDestinations(destinations []*model.Destination) {
	for _, destination := range destinations {
		if u.DestinationsField.Contains(destination.Name(), EventTypeUpdate) {
			continue
		}

		if u.AffectsDestination(destination) {
			u.IncludeDestination(destination, EventTypeUpdate)
			u.transitiveUpdates = append(u.transitiveUpdates, destination)
		}
	}
}

// AddAffectedConfigurations will add updates for Configurations that are affected by other resource updates.
func (u *EventUpdates) AddAffectedConfigurations(configurations []*model.Configuration) {
	for _, configuration := range configurations {
		if u.ConfigurationsField.Contains(configuration.Name(), EventTypeUpdate) {
			continue
		}

		if u.AffectsConfiguration(configuration) {
			u.IncludeConfiguration(configuration, EventTypeUpdate)
			u.transitiveUpdates = append(u.transitiveUpdates, configuration)
		}
	}
}

// MergeUpdates merges the updates from the given Updates into the current Updates.
func MergeUpdates(into, from BasicEventUpdates) bool {
	safe := into.Agents().CanSafelyMerge(from.Agents()) &&
		into.AgentVersions().CanSafelyMerge(from.AgentVersions()) &&
		into.Sources().CanSafelyMerge(from.Sources()) &&
		into.SourceTypes().CanSafelyMerge(from.SourceTypes()) &&
		into.Processors().CanSafelyMerge(from.Processors()) &&
		into.ProcessorTypes().CanSafelyMerge(from.ProcessorTypes()) &&
		into.Destinations().CanSafelyMerge(from.Destinations()) &&
		into.DestinationTypes().CanSafelyMerge(from.DestinationTypes()) &&
		into.Configurations().CanSafelyMerge(from.Configurations())

	if !safe {
		return false
	}

	into.Agents().Merge(from.Agents())
	into.AgentVersions().Merge(from.AgentVersions())
	into.Sources().Merge(from.Sources())
	into.SourceTypes().Merge(from.SourceTypes())
	into.Processors().Merge(from.Processors())
	into.ProcessorTypes().Merge(from.ProcessorTypes())
	into.Destinations().Merge(from.Destinations())
	into.DestinationTypes().Merge(from.DestinationTypes())
	into.Configurations().Merge(from.Configurations())

	return true
}

// NewEventUpdates returns a new Updates object.
func NewEventUpdates() BasicEventUpdates {
	// TODO: optimize allocate as needed
	return &EventUpdates{
		AgentsField:           NewEvents[*model.Agent](),
		AgentVersionsField:    NewEvents[*model.AgentVersion](),
		SourcesField:          NewEvents[*model.Source](),
		SourceTypesField:      NewEvents[*model.SourceType](),
		ProcessorsField:       NewEvents[*model.Processor](),
		ProcessorTypesField:   NewEvents[*model.ProcessorType](),
		DestinationsField:     NewEvents[*model.Destination](),
		DestinationTypesField: NewEvents[*model.DestinationType](),
		ConfigurationsField:   NewEvents[*model.Configuration](),
	}
}

// BuildBasicEventBroadcast returns a BroadCastBuilder that builds a broadcast.Broadcast[BasicUpdates] using routing and broadcast options for oss.
func BuildBasicEventBroadcast() BroadCastBuilder[BasicEventUpdates] {
	return func(ctx context.Context, options Options, logger *zap.Logger, maxEventsToMerge int) broadcast.Broadcast[BasicEventUpdates] {
		return broadcast.NewLocalBroadcast(ctx, logger,
			broadcast.WithUnboundedChannel[BasicEventUpdates](100*time.Millisecond),
			broadcast.WithParseFunc(func(data []byte) (BasicEventUpdates, error) {
				var updates EventUpdates
				err := jsoniter.Unmarshal(data, &updates)
				return &updates, err
			}),
			broadcast.WithMerge(func(into, single BasicEventUpdates) bool {
				return into.Merge(single)
			}, 100*time.Millisecond, maxEventsToMerge),
		)
	}
}
