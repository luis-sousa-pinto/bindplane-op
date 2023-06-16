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
	agents           Events[*model.Agent]
	agentVersions    Events[*model.AgentVersion]
	sources          Events[*model.Source]
	sourceTypes      Events[*model.SourceType]
	processors       Events[*model.Processor]
	processorTypes   Events[*model.ProcessorType]
	destinations     Events[*model.Destination]
	destinationTypes Events[*model.DestinationType]
	configurations   Events[*model.Configuration]

	// transitiveUpdates is just used to track which resources need to have their resources updated
	transitiveUpdates []model.Resource
}

// TransitiveUpdates returns a list of resources that need to have their resources updated.
func (u *EventUpdates) TransitiveUpdates() []model.Resource {
	return u.transitiveUpdates
}

// Agents returns a collection of agent events.
func (u *EventUpdates) Agents() Events[*model.Agent] {
	return u.agents
}

// AgentVersions returns a collection of agent version events.
func (u *EventUpdates) AgentVersions() Events[*model.AgentVersion] {
	return u.agentVersions
}

// Sources returns a collection of source events.
func (u *EventUpdates) Sources() Events[*model.Source] {
	return u.sources
}

// SourceTypes returns a collection of source type events.
func (u *EventUpdates) SourceTypes() Events[*model.SourceType] {
	return u.sourceTypes
}

// Processors returns a collection of processor events.
func (u *EventUpdates) Processors() Events[*model.Processor] {
	return u.processors
}

// ProcessorTypes returns a collection of processor type events.
func (u *EventUpdates) ProcessorTypes() Events[*model.ProcessorType] {
	return u.processorTypes
}

// Destinations returns a collection of destination events.
func (u *EventUpdates) Destinations() Events[*model.Destination] {
	return u.destinations
}

// DestinationTypes returns a collection of destination type events.
func (u *EventUpdates) DestinationTypes() Events[*model.DestinationType] {
	return u.destinationTypes
}

// Configurations returns a collection of configuration events.
func (u *EventUpdates) Configurations() Events[*model.Configuration] {
	return u.configurations
}

// IncludeAgent will add an agent event to Updates.
func (u *EventUpdates) IncludeAgent(agent *model.Agent, eventType EventType) {
	if u.agents == nil {
		u.agents = NewEvents[*model.Agent]()
	}

	u.agents.Include(agent, eventType)
}

// IncludeAgentVersion will add an agent version event to Updates.
func (u *EventUpdates) IncludeAgentVersion(agentVersion *model.AgentVersion, eventType EventType) {
	if u.agentVersions == nil {
		u.agentVersions = NewEvents[*model.AgentVersion]()
	}

	u.agentVersions.Include(agentVersion, eventType)
}

// IncludeSource will add a source event to Updates.
func (u *EventUpdates) IncludeSource(source *model.Source, eventType EventType) {
	if u.sources == nil {
		u.sources = NewEvents[*model.Source]()
	}

	u.sources.Include(source, eventType)
}

// IncludeSourceType will add a source type event to Updates.
func (u *EventUpdates) IncludeSourceType(sourceType *model.SourceType, eventType EventType) {
	if u.sourceTypes == nil {
		u.sourceTypes = NewEvents[*model.SourceType]()
	}

	u.sourceTypes.Include(sourceType, eventType)
}

// IncludeProcessor will add a processor event to Updates.
func (u *EventUpdates) IncludeProcessor(processor *model.Processor, eventType EventType) {
	if u.processors == nil {
		u.processors = NewEvents[*model.Processor]()
	}

	u.processors.Include(processor, eventType)
}

// IncludeProcessorType will add a processor type event to Updates.
func (u *EventUpdates) IncludeProcessorType(processorType *model.ProcessorType, eventType EventType) {
	if u.processorTypes == nil {
		u.processorTypes = NewEvents[*model.ProcessorType]()
	}

	u.processorTypes.Include(processorType, eventType)
}

// IncludeDestination will add a destination event to Updates.
func (u *EventUpdates) IncludeDestination(destination *model.Destination, eventType EventType) {
	if u.destinations == nil {
		u.destinations = NewEvents[*model.Destination]()
	}

	u.destinations.Include(destination, eventType)
}

// IncludeDestinationType will add a destination type event to Updates.
func (u *EventUpdates) IncludeDestinationType(destinationType *model.DestinationType, eventType EventType) {
	if u.destinationTypes == nil {
		u.destinationTypes = NewEvents[*model.DestinationType]()
	}

	u.destinationTypes.Include(destinationType, eventType)
}

// IncludeConfiguration will add a configuration event to Updates.
func (u *EventUpdates) IncludeConfiguration(configuration *model.Configuration, eventType EventType) {
	if u.configurations == nil {
		u.configurations = NewEvents[*model.Configuration]()
	}

	u.configurations.Include(configuration, eventType)
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
	return len(u.agents) +
		len(u.agentVersions) +
		len(u.sources) +
		len(u.sourceTypes) +
		len(u.processors) +
		len(u.processorTypes) +
		len(u.destinations) +
		len(u.destinationTypes) +
		len(u.configurations)
}

// Merge merges another set of updates into this one, returns true
// if it was able to merge any updates.
func (u *EventUpdates) Merge(other BasicEventUpdates) bool {
	safe :=
		u.agents.CanSafelyMerge(other.Agents()) &&
			u.agentVersions.CanSafelyMerge(other.AgentVersions()) &&
			u.sources.CanSafelyMerge(other.Sources()) &&
			u.sourceTypes.CanSafelyMerge(other.SourceTypes()) &&
			u.processors.CanSafelyMerge(other.Processors()) &&
			u.processorTypes.CanSafelyMerge(other.ProcessorTypes()) &&
			u.destinations.CanSafelyMerge(other.Destinations()) &&
			u.destinationTypes.CanSafelyMerge(other.DestinationTypes()) &&
			u.configurations.CanSafelyMerge(other.Configurations())

	if !safe {
		return false
	}

	u.agents.Merge(other.Agents())
	u.agentVersions.Merge(other.AgentVersions())
	u.sources.Merge(other.Sources())
	u.sourceTypes.Merge(other.SourceTypes())
	u.processors.Merge(other.Processors())
	u.processorTypes.Merge(other.ProcessorTypes())
	u.destinations.Merge(other.Destinations())
	u.destinationTypes.Merge(other.DestinationTypes())
	u.configurations.Merge(other.Configurations())
	return true
}

// HasSourceTypeEvents returns true if any source type events exist.
func (u *EventUpdates) HasSourceTypeEvents() bool {
	return !u.sourceTypes.Empty()
}

// HasProcessorTypeEvents returns true if any processor type events exist.
func (u *EventUpdates) HasProcessorTypeEvents() bool {
	return !u.processorTypes.Empty()
}

// HasDestinationTypeEvents returns true if any destination type events exist.
func (u *EventUpdates) HasDestinationTypeEvents() bool {
	return !u.destinationTypes.Empty()
}

// HasSourceEvents returns true if any source events exist.
func (u *EventUpdates) HasSourceEvents() bool {
	return !u.sources.Empty()
}

// HasProcessorEvents returns true if any processor events exist.
func (u *EventUpdates) HasProcessorEvents() bool {
	return !u.processors.Empty()
}

// HasDestinationEvents returns true if any destination events exist.
func (u *EventUpdates) HasDestinationEvents() bool {
	return !u.destinations.Empty()
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
	if u.sourceTypes.Contains(source.Spec.Type, EventTypeUpdate) {
		return true
	}
	return u.AffectsResourceProcessors(source.Spec.Processors)
}

// AffectsProcessor returns true if the updates affect the given processor.
func (u *EventUpdates) AffectsProcessor(processor *model.Processor) bool {
	return u.processorTypes.Contains(processor.Spec.Type, EventTypeUpdate)
}

// AffectsDestination returns true if the updates affect the given destination.
func (u *EventUpdates) AffectsDestination(destination *model.Destination) bool {
	// DestinationType
	if u.destinationTypes.Contains(destination.Spec.Type, EventTypeUpdate) {
		return true
	}
	return u.AffectsResourceProcessors(destination.Spec.Processors)
}

// AffectsResourceProcessors returns true if the updates affect any of the given resource processors.
func (u *EventUpdates) AffectsResourceProcessors(processors []model.ResourceConfiguration) bool {
	for _, processor := range processors {
		if u.processors.Contains(processor.Name, EventTypeUpdate) {
			return true
		}
		if u.processorTypes.Contains(processor.Type, EventTypeUpdate) {
			return true
		}
	}
	return false
}

// AffectsConfiguration returns true if the updates affect the given configuration.
func (u *EventUpdates) AffectsConfiguration(configuration *model.Configuration) bool {
	for _, source := range configuration.Spec.Sources {
		if u.sources.ContainsKey(source.Name) {
			return true
		}
		if u.sourceTypes.ContainsKey(source.Type) {
			return true
		}
	}

	for _, destination := range configuration.Spec.Destinations {
		if u.destinations.ContainsKey(destination.Name) {
			return true
		}
		if u.destinationTypes.ContainsKey(destination.Type) {
			return true
		}
	}
	return false
}

// AddAffectedSources will add updates for Sources that are affected by other resource updates.
func (u *EventUpdates) AddAffectedSources(sources []*model.Source) {
	for _, source := range sources {
		if u.sources.Contains(source.Name(), EventTypeUpdate) {
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
		if u.processors.Contains(processor.Name(), EventTypeUpdate) {
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
		if u.destinations.Contains(destination.Name(), EventTypeUpdate) {
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
		if u.configurations.Contains(configuration.Name(), EventTypeUpdate) {
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
		agents:           NewEvents[*model.Agent](),
		agentVersions:    NewEvents[*model.AgentVersion](),
		sources:          NewEvents[*model.Source](),
		sourceTypes:      NewEvents[*model.SourceType](),
		processors:       NewEvents[*model.Processor](),
		processorTypes:   NewEvents[*model.ProcessorType](),
		destinations:     NewEvents[*model.Destination](),
		destinationTypes: NewEvents[*model.DestinationType](),
		configurations:   NewEvents[*model.Configuration](),
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
