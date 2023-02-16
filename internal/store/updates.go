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
	"context"
	"time"

	"github.com/observiq/bindplane-op/internal/eventbus"
	"github.com/observiq/bindplane-op/model"
)

// Updates is a collection of events created by a store operation.
type Updates struct {
	Agents           Events[*model.Agent]
	AgentVersions    Events[*model.AgentVersion]
	Sources          Events[*model.Source]
	SourceTypes      Events[*model.SourceType]
	Processors       Events[*model.Processor]
	ProcessorTypes   Events[*model.ProcessorType]
	Destinations     Events[*model.Destination]
	DestinationTypes Events[*model.DestinationType]
	Configurations   Events[*model.Configuration]
}

// NewUpdates returns a new Updates object.
func NewUpdates() *Updates {
	// TODO: optimize allocate as needed
	return &Updates{
		Agents:           NewEvents[*model.Agent](),
		AgentVersions:    NewEvents[*model.AgentVersion](),
		Sources:          NewEvents[*model.Source](),
		SourceTypes:      NewEvents[*model.SourceType](),
		Processors:       NewEvents[*model.Processor](),
		ProcessorTypes:   NewEvents[*model.ProcessorType](),
		Destinations:     NewEvents[*model.Destination](),
		DestinationTypes: NewEvents[*model.DestinationType](),
		Configurations:   NewEvents[*model.Configuration](),
	}
}

// IncludeAgent will add an agent event to Updates.
func (u *Updates) IncludeAgent(agent *model.Agent, eventType EventType) {
	if u.Agents == nil {
		u.Agents = NewEvents[*model.Agent]()
	}

	u.Agents.Include(agent, eventType)
}

// IncludeAgentVersion will add an agent version event to Updates.
func (u *Updates) IncludeAgentVersion(agentVersion *model.AgentVersion, eventType EventType) {
	if u.AgentVersions == nil {
		u.AgentVersions = NewEvents[*model.AgentVersion]()
	}

	u.AgentVersions.Include(agentVersion, eventType)
}

// IncludeSource will add a source event to Updates.
func (u *Updates) IncludeSource(source *model.Source, eventType EventType) {
	if u.Sources == nil {
		u.Sources = NewEvents[*model.Source]()
	}

	u.Sources.Include(source, eventType)
}

// IncludeSourceType will add a source type event to Updates.
func (u *Updates) IncludeSourceType(sourceType *model.SourceType, eventType EventType) {
	if u.SourceTypes == nil {
		u.SourceTypes = NewEvents[*model.SourceType]()
	}

	u.SourceTypes.Include(sourceType, eventType)
}

// IncludeProcessor will add a processor event to Updates.
func (u *Updates) IncludeProcessor(processor *model.Processor, eventType EventType) {
	if u.Processors == nil {
		u.Processors = NewEvents[*model.Processor]()
	}

	u.Processors.Include(processor, eventType)
}

// IncludeProcessorType will add a processor type event to Updates.
func (u *Updates) IncludeProcessorType(processorType *model.ProcessorType, eventType EventType) {
	if u.ProcessorTypes == nil {
		u.ProcessorTypes = NewEvents[*model.ProcessorType]()
	}

	u.ProcessorTypes.Include(processorType, eventType)
}

// IncludeDestination will add a destination event to Updates.
func (u *Updates) IncludeDestination(destination *model.Destination, eventType EventType) {
	if u.Destinations == nil {
		u.Destinations = NewEvents[*model.Destination]()
	}

	u.Destinations.Include(destination, eventType)
}

// IncludeDestinationType will add a destination type event to Updates.
func (u *Updates) IncludeDestinationType(destinationType *model.DestinationType, eventType EventType) {
	if u.DestinationTypes == nil {
		u.DestinationTypes = NewEvents[*model.DestinationType]()
	}

	u.DestinationTypes.Include(destinationType, eventType)
}

// IncludeConfiguration will add a configuration event to Updates.
func (u *Updates) IncludeConfiguration(configuration *model.Configuration, eventType EventType) {
	if u.Configurations == nil {
		u.Configurations = NewEvents[*model.Configuration]()
	}

	u.Configurations.Include(configuration, eventType)
}

// IncludeResource will add a resource event to Updates.
// If the resource is not supported by Updates, this will do nothing.
func (u *Updates) IncludeResource(r model.Resource, eventType EventType) {
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
func (u *Updates) Empty() bool {
	return u.Size() == 0
}

// Size returns the sum of all events.
func (u *Updates) Size() int {
	return len(u.Agents) +
		len(u.AgentVersions) +
		len(u.Sources) +
		len(u.SourceTypes) +
		len(u.Processors) +
		len(u.ProcessorTypes) +
		len(u.Destinations) +
		len(u.DestinationTypes) +
		len(u.Configurations)
}

// HasSourceTypeEvents returns true if any source type events exist.
func (u *Updates) HasSourceTypeEvents() bool {
	return !u.SourceTypes.Empty()
}

// HasProcessorTypeEvents returns true if any processor type events exist.
func (u *Updates) HasProcessorTypeEvents() bool {
	return !u.ProcessorTypes.Empty()
}

// HasDestinationTypeEvents returns true if any destination type events exist.
func (u *Updates) HasDestinationTypeEvents() bool {
	return !u.DestinationTypes.Empty()
}

// HasSourceEvents returns true if any source events exist.
func (u *Updates) HasSourceEvents() bool {
	return !u.Sources.Empty()
}

// HasProcessorEvents returns true if any processor events exist.
func (u *Updates) HasProcessorEvents() bool {
	return !u.Processors.Empty()
}

// HasDestinationEvents returns true if any destination events exist.
func (u *Updates) HasDestinationEvents() bool {
	return !u.Destinations.Empty()
}

// CouldAffectProcessors returns true if the updates could affect processors.
func (u *Updates) CouldAffectProcessors() bool {
	return u.HasProcessorTypeEvents()
}

// CouldAffectSources returns true if the updates could affect sources.
func (u *Updates) CouldAffectSources() bool {
	return u.HasSourceTypeEvents() ||
		u.HasProcessorTypeEvents() ||
		u.HasProcessorEvents()
}

// CouldAffectDestinations returns true if the updates could affect destinations.
func (u *Updates) CouldAffectDestinations() bool {
	return u.HasDestinationTypeEvents()
}

// CouldAffectConfigurations returns true if the updates could affect configurations.
func (u *Updates) CouldAffectConfigurations() bool {
	return u.HasSourceTypeEvents() ||
		u.HasSourceEvents() ||
		u.HasProcessorTypeEvents() ||
		u.HasProcessorEvents() ||
		u.HasDestinationTypeEvents() ||
		u.HasDestinationEvents()
}

// AffectsSource returns true if the updates affect the given source.
func (u *Updates) AffectsSource(source *model.Source) bool {
	for _, sourceTypeEvent := range u.SourceTypes.Updates() {
		sourceTypeName := sourceTypeEvent.Item.Name()
		if source.Spec.Type == sourceTypeName {
			return true
		}
	}

	for _, processorTypeEvent := range u.ProcessorTypes.Updates() {
		processorTypeName := processorTypeEvent.Item.Name()
		for _, processor := range source.Spec.Processors {
			if processor.Type == processorTypeName {
				return true
			}
		}
	}

	for _, processorEvent := range u.Processors.Updates() {
		processorName := processorEvent.Item.Name()
		for _, processor := range source.Spec.Processors {
			if processor.Name == processorName {
				return true
			}
		}
	}

	return false
}

// AffectsProcessor returns true if the updates affect the given processor.
func (u *Updates) AffectsProcessor(processor *model.Processor) bool {
	for _, processorTypeEvent := range u.ProcessorTypes {
		if processorTypeEvent.Type == EventTypeUpdate {
			processorTypeName := processorTypeEvent.Item.Name()
			if processor.Spec.Type == processorTypeName {
				return true
			}
		}
	}

	return false
}

// AffectsDestination returns true if the updates affect the given destination.
func (u *Updates) AffectsDestination(destination *model.Destination) bool {
	for _, destinationTypeEvent := range u.DestinationTypes {
		if destinationTypeEvent.Type == EventTypeUpdate {
			destinationTypeName := destinationTypeEvent.Item.Name()
			if destination.Spec.Type == destinationTypeName {
				return true
			}
		}
	}

	return false
}

// AffectsConfiguration returns true if the updates affect the given configuration.
func (u *Updates) AffectsConfiguration(configuration *model.Configuration) bool {
	for _, source := range configuration.Spec.Sources {
		if _, ok := u.Sources[source.Name]; ok {
			return true
		}

		if _, ok := u.SourceTypes[source.Type]; ok {
			return true
		}
	}

	for _, destination := range configuration.Spec.Destinations {
		if _, ok := u.Destinations[destination.Name]; ok {
			return true
		}

		if _, ok := u.DestinationTypes[destination.Type]; ok {
			return true
		}
	}

	return false
}

// AddAffectedSources will add updates for Sources that are affected by other resource updates.
func (u *Updates) AddAffectedSources(sources []*model.Source) {
	for _, source := range sources {
		if u.Sources.Contains(source.Name(), EventTypeUpdate) {
			continue
		}

		if u.AffectsSource(source) {
			u.IncludeSource(source, EventTypeUpdate)
		}
	}
}

// AddAffectedProcessors will add updates for Processors that are affected by other resource updates.
func (u *Updates) AddAffectedProcessors(processors []*model.Processor) {
	for _, processor := range processors {
		if u.Processors.Contains(processor.Name(), EventTypeUpdate) {
			continue
		}

		if u.AffectsProcessor(processor) {
			u.IncludeProcessor(processor, EventTypeUpdate)
		}
	}
}

// AddAffectedDestinations will add updates for Destinations that are affected by other resource updates.
func (u *Updates) AddAffectedDestinations(destinations []*model.Destination) {
	for _, destination := range destinations {
		if u.Destinations.Contains(destination.Name(), EventTypeUpdate) {
			continue
		}

		if u.AffectsDestination(destination) {
			u.IncludeDestination(destination, EventTypeUpdate)
		}
	}
}

// AddAffectedConfigurations will add updates for Configurations that are affected by other resource updates.
func (u *Updates) AddAffectedConfigurations(configurations []*model.Configuration) {
	for _, configuration := range configurations {
		if u.Configurations.Contains(configuration.Name(), EventTypeUpdate) {
			continue
		}

		if u.AffectsConfiguration(configuration) {
			u.IncludeConfiguration(configuration, EventTypeUpdate)
		}
	}
}

// MergeUpdates merges the updates from the given Updates into the current Updates.
func MergeUpdates(into, from *Updates) bool {
	safe := into.Agents.CanSafelyMerge(from.Agents) &&
		into.AgentVersions.CanSafelyMerge(from.AgentVersions) &&
		into.Sources.CanSafelyMerge(from.Sources) &&
		into.SourceTypes.CanSafelyMerge(from.SourceTypes) &&
		into.Processors.CanSafelyMerge(from.Processors) &&
		into.ProcessorTypes.CanSafelyMerge(from.ProcessorTypes) &&
		into.Destinations.CanSafelyMerge(from.Destinations) &&
		into.DestinationTypes.CanSafelyMerge(from.DestinationTypes) &&
		into.Configurations.CanSafelyMerge(from.Configurations)

	if !safe {
		return false
	}

	into.Agents.Merge(from.Agents)
	into.AgentVersions.Merge(from.AgentVersions)
	into.Sources.Merge(from.Sources)
	into.SourceTypes.Merge(from.SourceTypes)
	into.Processors.Merge(from.Processors)
	into.ProcessorTypes.Merge(from.ProcessorTypes)
	into.Destinations.Merge(from.Destinations)
	into.DestinationTypes.Merge(from.DestinationTypes)
	into.Configurations.Merge(from.Configurations)

	return true
}

// UpdatesEventBus is a wrapped event bus for store updates.
type UpdatesEventBus struct {
	// external is an external channel used by external clients.
	external eventbus.Source[*Updates]

	// internal is an internal channel used for merging and relaying.
	internal eventbus.Source[*Updates]
}

// NewUpdatesEventBus creates a new UpdatesEventBus.
func NewUpdatesEventBus(ctx context.Context, maxEventsToMerge int) *UpdatesEventBus {
	external := eventbus.NewSource[*Updates]()
	internal := eventbus.NewSource[*Updates]()

	if maxEventsToMerge == 0 {
		maxEventsToMerge = 100
	}

	// introduce a separate relay with a large buffer to avoid blocking on changes
	eventbus.RelayWithMerge(
		ctx,
		internal,
		MergeUpdates,
		external,
		200*time.Millisecond,
		maxEventsToMerge,
		eventbus.WithUnboundedChannel[*Updates](100*time.Millisecond),
	)

	return &UpdatesEventBus{
		external: external,
		internal: internal,
	}
}

// Updates returns the external channel that can be provided to external clients.
func (s *UpdatesEventBus) Updates() eventbus.Source[*Updates] {
	return s.external
}

// Send adds an Updates event to the internal channel where it can be merged and relayed to the external channel.
func (s *UpdatesEventBus) Send(updates *Updates) {
	s.internal.Send(updates)
}
