// Copyright  observIQ, Inc
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
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	"github.com/observiq/bindplane-op/eventbus"
	"github.com/observiq/bindplane-op/model"
	"github.com/observiq/bindplane-op/otlp/record"
	"github.com/observiq/bindplane-op/store/search"
	"github.com/observiq/bindplane-op/store/stats"
)

type mapStore struct {
	agents map[string]*model.Agent

	agentVersions    resourceStore[*model.AgentVersion]
	configurations   resourceStore[*model.Configuration]
	sources          resourceStore[*model.Source]
	sourceTypes      resourceStore[*model.SourceType]
	processors       resourceStore[*model.Processor]
	processorTypes   resourceStore[*model.ProcessorType]
	destinations     resourceStore[*model.Destination]
	destinationTypes resourceStore[*model.DestinationType]

	updates            *Updates
	agentIndex         search.Index
	configurationIndex search.Index
	logger             *zap.Logger
	sync.RWMutex

	sessionStore sessions.Store
}

var _ Store = (*mapStore)(nil)

// NewMapStore returns an in memory Store
func NewMapStore(ctx context.Context, options Options, logger *zap.Logger) Store {
	return &mapStore{
		agents:             make(map[string]*model.Agent),
		configurations:     newResourceStore[*model.Configuration](),
		sources:            newResourceStore[*model.Source](),
		sourceTypes:        newResourceStore[*model.SourceType](),
		processors:         newResourceStore[*model.Processor](),
		processorTypes:     newResourceStore[*model.ProcessorType](),
		destinations:       newResourceStore[*model.Destination](),
		destinationTypes:   newResourceStore[*model.DestinationType](),
		updates:            NewUpdates(ctx, options, logger, BuildBasicEventBroadcast(), BuildRolloutEventBroadcast(), NewRolloutUpdates),
		agentIndex:         search.NewInMemoryIndex("agent"),
		configurationIndex: search.NewInMemoryIndex("configuration"),
		logger:             logger,
		sessionStore:       NewBPCookieStore(options.SessionsSecret),
	}
}

// ----------------------------------------------------------------------
//
// resourceStore stores a single type of resource and has its own lock
type resourceStore[T model.Resource] struct {
	store map[string]T
	mtx   sync.RWMutex
}

func newResourceStore[T model.Resource]() resourceStore[T] {
	return resourceStore[T]{
		store: map[string]T{},
	}
}

func (r *resourceStore[T]) get(name string) T {
	r.mtx.RLock()
	defer r.mtx.RUnlock()
	return r.store[name]
}

func (r *resourceStore[T]) add(resource T) *model.ResourceStatus {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	// generate a uuid if none supplied
	if resource.ID() == "" {
		resource.SetID(uuid.NewString())
	}
	existing, ok := r.store[resource.Name()]
	if ok && existing.ID() != "" {
		resource.SetID(existing.ID())
	}

	r.store[resource.Name()] = resource

	var status model.UpdateStatus
	switch {
	case !ok:
		status = model.StatusCreated
	case !resourcesEqual(existing, resource):
		status = model.StatusConfigured
	default:
		status = model.StatusUnchanged
	}

	return model.NewResourceStatus(resource, status)
}

func (r *resourceStore[T]) remove(name string) (item T, exists bool) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	existing, ok := r.store[name]
	if ok {
		delete(r.store, name)
	}
	return existing, ok
}

func (r *resourceStore[T]) removeAndNotify(ctx context.Context, name string, store *mapStore) (item T, exists bool, err error) {
	r.mtx.Lock()
	existing, ok := r.store[name]

	if ok {
		dependencies, err := FindDependentResources(ctx, store.configurationIndex, existing.Name(), existing.GetKind())
		if err != nil {
			r.mtx.Unlock()
			return item, ok, err
		}

		if !dependencies.Empty() {
			r.mtx.Unlock()
			return item, ok, NewDependencyError(dependencies)
		}
		delete(r.store, name)
	}
	r.mtx.Unlock()

	if ok {
		updates := NewEventUpdates()
		updates.IncludeResource(existing, EventTypeRemove)
		store.notify(ctx, updates)
	}

	return existing, ok, nil
}

func (r *resourceStore[T]) list() []T {
	r.mtx.RLock()
	defer r.mtx.RUnlock()
	return maps.Values(r.store)
}

func (r *resourceStore[T]) clear() {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	r.store = map[string]T{}
}

// ----------------------------------------------------------------------

func (mapstore *mapStore) Close() error {
	mapstore.updates.Shutdown(context.Background())
	return nil
}

func (mapstore *mapStore) Clear() {
	mapstore.Lock()
	defer mapstore.Unlock()

	mapstore.agents = make(map[string]*model.Agent)

	mapstore.configurations.clear()
	mapstore.sources.clear()
	mapstore.sourceTypes.clear()
	mapstore.destinations.clear()
	mapstore.destinationTypes.clear()
}

func (mapstore *mapStore) UpdateConfiguration(_ context.Context, _ string, _ ConfigurationUpdater) (config *model.Configuration, status model.UpdateStatus, err error) {
	return nil, model.StatusConfigured, nil
}
func (mapstore *mapStore) UpdateRollouts(_ context.Context) ([]*model.Configuration, error) {
	return nil, nil
}
func (mapstore *mapStore) UpdateRollout(_ context.Context, _ string) (*model.Configuration, error) {
	return nil, nil
}
func (mapstore *mapStore) ResumeRollout(_ context.Context, _ string) (*model.Configuration, error) {
	return nil, nil
}
func (mapstore *mapStore) PauseRollout(_ context.Context, _ string) (*model.Configuration, error) {
	return nil, nil
}
func (mapstore *mapStore) StartRollout(_ context.Context, _ string, _ *model.RolloutOptions) (*model.Configuration, error) {
	return nil, nil
}
func (mapstore *mapStore) ResourceHistory(_ context.Context, _ model.Kind, _ string) ([]*model.AnyResource, error) {
	return nil, nil
}
func (mapstore *mapStore) UpdateAllRollouts(_ context.Context) error {
	return nil
}

func (mapstore *mapStore) UpsertAgent(ctx context.Context, agentID string, updater AgentUpdater) (*model.Agent, error) {
	mapstore.Lock()
	defer mapstore.Unlock()

	u := NewEventUpdates()
	agent := mapstore.upsertAgent(agentID, updater, u)

	mapstore.notify(ctx, u)

	return agent, nil
}

func (mapstore *mapStore) UpsertAgents(ctx context.Context, agentIDs []string, updater AgentUpdater) ([]*model.Agent, error) {
	mapstore.Lock()
	defer mapstore.Unlock()

	agents := make([]*model.Agent, 0, len(agentIDs))
	u := NewEventUpdates()

	for _, id := range agentIDs {
		agents = append(agents, mapstore.upsertAgent(id, updater, u))
	}

	mapstore.notify(ctx, u)

	return agents, nil
}

// UpdateAgent updates an existing Agent in the Store. If the agentID does not exist, no error is returned but the
// agent will be nil. An error is only returned if the update fails.
func (mapstore *mapStore) UpdateAgent(ctx context.Context, agentID string, updater AgentUpdater) (*model.Agent, error) {
	mapstore.Lock()
	defer mapstore.Unlock()

	u := NewEventUpdates()
	agent := mapstore.updateAgent(agentID, updater, u)

	mapstore.notify(ctx, u)

	return agent, nil
}

// UpdateAgents updates existing Agents in the Store. If an agentID does not exist, that agentID is ignored and no
// agent corresponding to that ID will be returned. An error is only returned if the update fails.
func (mapstore *mapStore) UpdateAgents(ctx context.Context, agentIDs []string, updater AgentUpdater) ([]*model.Agent, error) {
	mapstore.Lock()
	defer mapstore.Unlock()

	agents := make([]*model.Agent, 0, len(agentIDs))
	u := NewEventUpdates()

	for _, id := range agentIDs {
		if agent := mapstore.updateAgent(id, updater, u); agent != nil {
			agents = append(agents, agent)
		}
	}

	mapstore.notify(ctx, u)

	return agents, nil
}

func (mapstore *mapStore) Agents(_ context.Context, options ...QueryOption) ([]*model.Agent, error) {
	mapstore.RLock()
	defer mapstore.RUnlock()

	opts := MakeQueryOptions(options)
	result := make([]*model.Agent, 0, len(mapstore.agents))

	for _, value := range mapstore.agents {
		if opts.Selector.Matches(value.Labels) {
			result = append(result, value)
		}
	}

	if opts.Sort == "" {
		opts.Sort = "name"
	}
	return ApplySortOffsetAndLimit(result, opts, func(field string, item *model.Agent) string {
		switch field {
		case "id":
			return item.ID
		default:
			return item.Name
		}
	}), nil
}

func (mapstore *mapStore) AgentsCount(ctx context.Context, options ...QueryOption) (int, error) {
	agents, err := mapstore.Agents(ctx, options...)
	if err != nil {
		return -1, err
	}
	return len(agents), nil
}

func (mapstore *mapStore) Agent(_ context.Context, id string) (*model.Agent, error) {
	mapstore.RLock()
	defer mapstore.RUnlock()
	return mapstore.agents[id], nil
}

func (mapstore *mapStore) DeleteAgents(ctx context.Context, agentIDs []string) ([]*model.Agent, error) {
	deleted := make([]*model.Agent, 0, len(agentIDs))
	updates := NewEventUpdates()

	mapstore.Lock()
	defer mapstore.Unlock()

	for _, id := range agentIDs {
		if agent, ok := mapstore.agents[id]; ok {
			// set status deleted
			agent.Status = model.Deleted

			// save the agent to return
			deleted = append(deleted, agent)

			// delete the agent
			delete(mapstore.agents, id)

			// include in the agent updates
			updates.Agents().Include(agent, EventTypeRemove)

			// remove from the search index
			if err := mapstore.agentIndex.Remove(agent); err != nil {
				mapstore.logger.Error("failed to remove agent from the search index", zap.String("agentID", agent.ID))
			}
		}
	}

	mapstore.notify(ctx, updates)

	return deleted, nil
}

func (mapstore *mapStore) AgentVersion(_ context.Context, name string) (*model.AgentVersion, error) {
	return mapstore.agentVersions.get(name), nil
}
func (mapstore *mapStore) AgentVersions(_ context.Context) ([]*model.AgentVersion, error) {
	return mapstore.agentVersions.list(), nil
}
func (mapstore *mapStore) DeleteAgentVersion(ctx context.Context, name string) (*model.AgentVersion, error) {
	item, exists, err := mapstore.agentVersions.removeAndNotify(ctx, name, mapstore)
	if err != nil {
		return item, err
	}
	if !exists {
		return nil, nil
	}
	return item, nil
}

func (mapstore *mapStore) Configurations(_ context.Context, options ...QueryOption) ([]*model.Configuration, error) {
	opts := MakeQueryOptions(options)
	config := mapstore.configurations.list()
	if opts.Sort == "" {
		opts.Sort = "name"
	}
	return ApplySortOffsetAndLimit(config, opts, func(field string, item *model.Configuration) string {
		// we currently only support sorting by name
		return item.Name()
	}), nil
}
func (mapstore *mapStore) Configuration(_ context.Context, name string) (*model.Configuration, error) {
	return mapstore.configurations.get(name), nil
}
func (mapstore *mapStore) DeleteConfiguration(ctx context.Context, name string) (*model.Configuration, error) {
	item, exists, err := mapstore.configurations.removeAndNotify(ctx, name, mapstore)
	if err != nil {
		return item, err
	}

	if !exists {
		return nil, nil
	}
	return item, nil
}

func (mapstore *mapStore) Source(_ context.Context, name string) (*model.Source, error) {
	return mapstore.sources.get(name), nil
}
func (mapstore *mapStore) Sources(_ context.Context) ([]*model.Source, error) {
	return mapstore.sources.list(), nil
}
func (mapstore *mapStore) DeleteSource(ctx context.Context, name string) (*model.Source, error) {
	item, exists, err := mapstore.sources.removeAndNotify(ctx, name, mapstore)
	if err != nil {
		return item, err
	}

	if !exists {
		return nil, nil
	}
	return item, nil
}

func (mapstore *mapStore) SourceType(_ context.Context, name string) (*model.SourceType, error) {
	return mapstore.sourceTypes.get(name), nil
}
func (mapstore *mapStore) SourceTypes(_ context.Context) ([]*model.SourceType, error) {
	return mapstore.sourceTypes.list(), nil
}
func (mapstore *mapStore) DeleteSourceType(ctx context.Context, name string) (*model.SourceType, error) {
	item, exists, err := mapstore.sourceTypes.removeAndNotify(ctx, name, mapstore)
	if err != nil {
		return item, err
	}

	if !exists {
		return nil, nil
	}
	return item, nil
}

func (mapstore *mapStore) Processor(_ context.Context, name string) (*model.Processor, error) {
	return mapstore.processors.get(name), nil
}
func (mapstore *mapStore) Processors(_ context.Context) ([]*model.Processor, error) {
	return mapstore.processors.list(), nil
}
func (mapstore *mapStore) DeleteProcessor(ctx context.Context, name string) (*model.Processor, error) {
	item, exists, err := mapstore.processors.removeAndNotify(ctx, name, mapstore)
	if err != nil {
		return item, err
	}

	if !exists {
		return nil, nil
	}
	return item, nil
}

func (mapstore *mapStore) ProcessorType(_ context.Context, name string) (*model.ProcessorType, error) {
	return mapstore.processorTypes.get(name), nil
}
func (mapstore *mapStore) ProcessorTypes(_ context.Context) ([]*model.ProcessorType, error) {
	return mapstore.processorTypes.list(), nil
}
func (mapstore *mapStore) DeleteProcessorType(ctx context.Context, name string) (*model.ProcessorType, error) {
	item, exists, err := mapstore.processorTypes.removeAndNotify(ctx, name, mapstore)
	if err != nil {
		return item, err
	}

	if !exists {
		return nil, nil
	}
	return item, nil
}

func (mapstore *mapStore) Destination(_ context.Context, name string) (*model.Destination, error) {
	return mapstore.destinations.get(name), nil
}
func (mapstore *mapStore) Destinations(_ context.Context) ([]*model.Destination, error) {
	return mapstore.destinations.list(), nil
}
func (mapstore *mapStore) DeleteDestination(ctx context.Context, name string) (*model.Destination, error) {
	item, exists, err := mapstore.destinations.removeAndNotify(ctx, name, mapstore)
	if err != nil {
		return item, err
	}

	if !exists {
		return nil, nil
	}
	return item, nil
}

func (mapstore *mapStore) DestinationType(_ context.Context, name string) (*model.DestinationType, error) {
	return mapstore.destinationTypes.get(name), nil
}
func (mapstore *mapStore) DestinationTypes(_ context.Context) ([]*model.DestinationType, error) {
	return mapstore.destinationTypes.list(), nil
}
func (mapstore *mapStore) DeleteDestinationType(ctx context.Context, name string) (*model.DestinationType, error) {
	item, exists, err := mapstore.destinationTypes.removeAndNotify(ctx, name, mapstore)
	if err != nil {
		return item, err
	}

	if !exists {
		return nil, nil
	}
	return item, nil
}

func (mapstore *mapStore) ApplyResources(ctx context.Context, resources []model.Resource) ([]model.ResourceStatus, error) {
	mapstore.Lock()
	defer mapstore.Unlock()
	var result error

	updates := NewEventUpdates()
	resourceStatuses := make([]model.ResourceStatus, 0)

	for _, resource := range resources {
		_, err := resource.ValidateWithStore(ctx, mapstore)
		if err != nil {
			resourceStatuses = append(resourceStatuses, *model.NewResourceStatusWithReason(resource, model.StatusInvalid, err.Error()))
			continue
		}

		var resourceStatus *model.ResourceStatus
		switch r := resource.(type) {
		case *model.AgentVersion:
			resourceStatus = mapstore.agentVersions.add(r)
		case *model.Configuration:
			resourceStatus = mapstore.configurations.add(r)
			if err := mapstore.configurationIndex.Upsert(resourceStatus.Resource); err != nil {
				mapstore.logger.Error("error updating configuration in the search index", zap.Error(err))
			}
		case *model.Source:
			resourceStatus = mapstore.sources.add(r)
		case *model.SourceType:
			resourceStatus = mapstore.sourceTypes.add(r)
		case *model.Processor:
			resourceStatus = mapstore.processors.add(r)
		case *model.ProcessorType:
			resourceStatus = mapstore.processorTypes.add(r)
		case *model.Destination:
			resourceStatus = mapstore.destinations.add(r)
		case *model.DestinationType:
			resourceStatus = mapstore.destinationTypes.add(r)
		default:
			resourceStatus = model.NewResourceStatusWithReason(resource, model.StatusInvalid, fmt.Sprintf("unknown resource type in apply: %s", r.Name()))
		}

		if resourceStatus != nil {
			resourceStatuses = append(resourceStatuses, *resourceStatus)

			switch resourceStatus.Status {
			case model.StatusCreated:
				updates.IncludeResource(resource, EventTypeInsert)
			case model.StatusConfigured:
				updates.IncludeResource(resource, EventTypeUpdate)
			}
		}
	}

	mapstore.notify(ctx, updates)
	return resourceStatuses, result
}

func (mapstore *mapStore) DeleteResources(ctx context.Context, resources []model.Resource) ([]model.ResourceStatus, error) {
	mapstore.Lock()
	defer mapstore.Unlock()

	// Send the pipeline deletes, even if its an empty map to satisfy tests
	updates := NewEventUpdates()

	resourceStatuses := make([]model.ResourceStatus, 0)

	for _, r := range resources {
		dependencies, err := FindDependentResources(ctx, mapstore.configurationIndex, r.Name(), r.GetKind())
		if err != nil {
			mapstore.logger.Error("failed to get dependent resources", zap.Error(err))
			continue
		}

		if !dependencies.Empty() {
			resourceStatuses = append(resourceStatuses,
				*model.NewResourceStatusWithReason(
					r,
					model.StatusInUse,
					dependencies.Message(),
				),
			)
			continue
		}

		var exists bool
		switch r := r.(type) {
		case *model.AgentVersion:
			_, exists = mapstore.agentVersions.remove(r.Name())

		case *model.Configuration:
			c, e := mapstore.configurations.remove(r.Name())
			if err := mapstore.configurationIndex.Remove(c); err != nil {
				mapstore.logger.Error("error removing configuration from the search index", zap.Error(err))
			}
			exists = e

		case *model.Source:
			_, exists = mapstore.sources.remove(r.Name())

		case *model.SourceType:
			_, exists = mapstore.sourceTypes.remove(r.Name())

		case *model.Processor:
			_, exists = mapstore.processors.remove(r.Name())

		case *model.ProcessorType:
			_, exists = mapstore.processorTypes.remove(r.Name())

		case *model.Destination:
			_, exists = mapstore.destinations.remove(r.Name())

		case *model.DestinationType:
			_, exists = mapstore.destinationTypes.remove(r.Name())

		default:
			continue
		}
		if exists {
			resourceStatuses = append(resourceStatuses, *model.NewResourceStatus(r, model.StatusDeleted))
			updates.IncludeResource(r, EventTypeRemove)
		}
	}

	mapstore.notify(ctx, updates)
	return resourceStatuses, nil
}

// AgentConfiguration returns the configuration that should be applied to an agent.
func (mapstore *mapStore) AgentConfiguration(_ context.Context, agent *model.Agent) (*model.Configuration, error) {
	mapstore.RLock()
	defer mapstore.RUnlock()
	if agent == nil {
		return nil, fmt.Errorf("cannot return configuration for unknown agent")
	}

	labels := agent.Labels

	// look through all of the configurations and check their selector to see if they match this agent. there are more
	// efficient implementations, but this is fine for mapstore.
	for _, c := range mapstore.configurations.store {
		if c.AgentSelector().Matches(labels) {
			return c, nil
		}
	}

	return nil, nil
}

// AgentsIDsMatchingConfiguration returns the list of agent IDs that are using the specified configuration
func (mapstore *mapStore) AgentsIDsMatchingConfiguration(_ context.Context, configuration *model.Configuration) ([]string, error) {
	ids := mapstore.agentIndex.Select(configuration.Spec.Selector.MatchLabels)
	return ids, nil
}

func (mapstore *mapStore) Updates(_ context.Context) eventbus.Source[BasicEventUpdates] {
	return mapstore.updates.Updates()
}

func (mapstore *mapStore) AgentRolloutUpdates(_ context.Context) eventbus.Source[RolloutEventUpdates] {
	return mapstore.updates.RolloutUpdates()
}

// ReportConnectedAgents sets the ReportedAt time for the specified agents to the specified time. This update should
// not fire an update event for the agents on the Updates eventbus.
func (mapstore *mapStore) ReportConnectedAgents(_ context.Context, agentIDs []string, time time.Time) error {
	mapstore.Lock()
	defer mapstore.Unlock()

	for _, agent := range mapstore.agents {
		if slices.Contains(agentIDs, agent.ID) {
			agent.ReportedAt = &time
		}
	}

	return nil
}

// DisconnectUnreportedAgents sets the Status of agents to Disconnected if the agent ReportedAt time is before the
// specified time.
func (mapstore *mapStore) DisconnectUnreportedAgents(ctx context.Context, since time.Time) error {
	mapstore.Lock()
	defer mapstore.Unlock()

	updates := NewEventUpdates()

	for _, agent := range mapstore.agents {
		if !agent.ReportedSince(since) {
			agent.Status = model.Disconnected
			updates.IncludeAgent(agent, EventTypeUpdate)
		}
	}

	mapstore.notify(ctx, updates)

	return nil
}

// CleanupDisconnectedAgents removes agents that have disconnected before the specified time
func (mapstore *mapStore) CleanupDisconnectedAgents(ctx context.Context, since time.Time) error {
	mapstore.Lock()
	defer mapstore.Unlock()

	updates := NewEventUpdates()

	for _, agent := range mapstore.agents {
		if agent.GetLabels().Has(model.LabelAgentContainerPlatform) &&
			agent.DisconnectedSince(since) {
			delete(mapstore.agents, agent.ID)
			updates.IncludeAgent(agent, EventTypeRemove)
		}
	}

	mapstore.notify(ctx, updates)

	return nil
}

// Index provides access to the search Index implementation managed by the Store
func (mapstore *mapStore) AgentIndex(_ context.Context) search.Index {
	return mapstore.agentIndex
}

// ConfigurationIndex provides access to the search Index for Configurations
func (mapstore *mapStore) ConfigurationIndex(_ context.Context) search.Index {
	return mapstore.configurationIndex
}

func (mapstore *mapStore) UserSessions() sessions.Store {
	return mapstore.sessionStore
}

// Measurements stores stats for agents and configurations
func (mapstore *mapStore) Measurements() stats.Measurements {
	return mapstore
}

// ----------------------------------------------------------------------
// these functions require that the mapstore is already locked

// ----------------------------------------------------------------------

// upsertAgent updates the agent with given id while the mapstore is locked.
// it updates the passed *Updates to include the change.
func (mapstore *mapStore) updateAgent(agentID string, updater AgentUpdater, updates BasicEventUpdates) *model.Agent {
	_, ok := mapstore.agents[agentID]
	if !ok {
		return nil
	}
	return mapstore.upsertAgent(agentID, updater, updates)
}

// upsertAgent updates the agent with given id while the mapstore is locked.
// it updates the passed *Updates to include the change.
func (mapstore *mapStore) upsertAgent(agentID string, updater AgentUpdater, updates BasicEventUpdates) *model.Agent {
	agentEventType := EventTypeInsert

	agent := &model.Agent{ID: agentID}
	if curAgent, ok := mapstore.agents[agentID]; ok {
		agent = curAgent
		agentEventType = EventTypeUpdate
	}

	// compare labels before/after and notify if they change
	labelsBefore := agent.Labels.String()

	updater(agent)
	mapstore.agents[agentID] = agent

	// update the index
	err := mapstore.agentIndex.Upsert(agent)
	if err != nil {
		mapstore.logger.Error("failed to update the search index", zap.String("agentID", agent.ID))
	}

	labelsAfter := agent.Labels.String()
	if labelsBefore != "" && labelsAfter != labelsBefore {
		agentEventType = EventTypeLabel
	}

	updates.IncludeAgent(agent, agentEventType)
	return agent
}

// addTransitiveUpdates adds all the transitive updates based on the resource type.
func (mapstore *mapStore) addTransitiveUpdates(ctx context.Context, updates BasicEventUpdates) error {
	if updates.CouldAffectProcessors() {
		processors, err := mapstore.Processors(ctx)
		if err != nil {
			return fmt.Errorf("failed to get processors: %w", err)
		}

		updates.AddAffectedProcessors(processors)
	}

	if updates.CouldAffectSources() {
		sources, err := mapstore.Sources(ctx)
		if err != nil {
			return fmt.Errorf("failed to get sources: %w", err)
		}

		updates.AddAffectedSources(sources)
	}

	if updates.CouldAffectDestinations() {
		destinations, err := mapstore.Destinations(ctx)
		if err != nil {
			return fmt.Errorf("failed to get destinations: %w", err)
		}

		updates.AddAffectedDestinations(destinations)
	}

	if updates.CouldAffectConfigurations() {
		configurations, err := mapstore.Configurations(ctx)
		if err != nil {
			return fmt.Errorf("failed to get configurations: %w", err)
		}

		updates.AddAffectedConfigurations(configurations)
	}

	return nil
}

func (mapstore *mapStore) notify(ctx context.Context, updates BasicEventUpdates) {
	err := mapstore.addTransitiveUpdates(ctx, updates)
	if err != nil {
		// TODO: if we can't notify about all updates, what do we do?
		mapstore.logger.Error("unable to add transitive updates", zap.Any("updates", updates), zap.Error(err))
	}
	if !updates.Empty() {
		mapstore.updates.Send(ctx, updates)
	}
}

func resourcesEqual(r1 model.Resource, r2 model.Resource) bool {
	r1Any := &model.AnyResource{}
	r2Any := &model.AnyResource{}
	err := mapstructure.Decode(r1, r1Any)
	if err != nil {
		return false
	}
	err = mapstructure.Decode(r2, r2Any)
	if err != nil {
		return false
	}

	r1Any.Metadata.ID = ""
	r2Any.Metadata.ID = ""
	return reflect.DeepEqual(r1Any, r2Any)
}

// ----------------------------------------------------------------------
// Measurements implementation

// AgentMetrics provides metrics for an individual agents. They are essentially configuration metrics filtered to a
// list of agents.
//
// Note: While the same record.Metric struct is used to return the metrics, these are not the same metrics provided to
// Store. They will be aggregated and counter metrics will be converted into rates.
func (mapstore *mapStore) AgentMetrics(_ context.Context, _ []string, _ ...stats.QueryOption) (stats.MetricData, error) {
	return nil, nil
}

// ConfigurationMetrics provides all metrics associated with a configuration aggregated from all agents using the
// configuration.
//
// Note: While the same record.Metric struct is used to return the metrics, these are not the same metrics provided to
// Store. They will be aggregated and counter metrics will be converted into rates.
func (mapstore *mapStore) ConfigurationMetrics(_ context.Context, _ string, _ ...stats.QueryOption) (stats.MetricData, error) {
	return nil, nil
}

// OverviewMetrics provides all metrics needed for the overview page. This page shows agents and destinations.
func (mapstore *mapStore) OverviewMetrics(_ context.Context, _ ...stats.QueryOption) (stats.MetricData, error) {
	return nil, nil
}

// SaveAgentMetrics saves new metrics. These metrics will be aggregated to determine metrics associated with agents and configurations.
func (mapstore *mapStore) SaveAgentMetrics(_ context.Context, _ []*record.Metric) error {
	return nil
}

// MeasurementsSize returns the count of keys in the store, and is used only for testing
func (mapstore *mapStore) MeasurementsSize(_ context.Context) (int, error) {
	return 0, nil
}

// ProcessMetrics is called in the background at regular intervals and performs metric roll-up and removes old data
func (mapstore *mapStore) ProcessMetrics(_ context.Context) error {
	return nil
}
