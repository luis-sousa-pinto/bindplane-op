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
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/sessions"
	"github.com/hashicorp/go-multierror"
	jsoniter "github.com/json-iterator/go"
	"github.com/observiq/bindplane-op/eventbus"
	"github.com/observiq/bindplane-op/model"
	"github.com/observiq/bindplane-op/otlp/record"
	"github.com/observiq/bindplane-op/store/search"
	"github.com/observiq/bindplane-op/store/stats"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"
)

const (
	// cleanupMeasurementsInterval is the interval at which old measurements are cleaned up
	cleanupMeasurementsInterval = time.Minute
)

// BoltstoreCore is an implementation of the store interface that uses BoltDB as the underlying storage mechanism
type BoltstoreCore struct {
	StoreUpdates   *Updates
	DB             *bbolt.DB
	Logger         *zap.Logger
	SessionStorage sessions.Store
	RolloutBatcher RolloutBatcher
	sync.RWMutex
	BoltstoreCommon
}

// BoltstoreCommon is an interface for common implementation details between different boltstore implementations
type BoltstoreCommon interface {
	Database() *bbolt.DB
	AgentsBucket(ctx context.Context, tx *bbolt.Tx) (*bbolt.Bucket, error)
	MeasurementsBucket(ctx context.Context, tx *bbolt.Tx, metric string) (*bbolt.Bucket, error)
	ResourcesBucket(ctx context.Context, tx *bbolt.Tx, kind model.Kind) (*bbolt.Bucket, error)
	ArchiveBucket(ctx context.Context, tx *bbolt.Tx) (*bbolt.Bucket, error)
	ResourceKey(r model.Resource) []byte
	AgentsIndex(ctx context.Context) search.Index
	ConfigurationsIndex(ctx context.Context) search.Index
	ZapLogger() *zap.Logger
	Notify(ctx context.Context, updates BasicEventUpdates)
	CreateEventUpdate() BasicEventUpdates
}

// AgentConfiguration returns the configuration that should be applied to an agent.
func (s *BoltstoreCore) AgentConfiguration(ctx context.Context, agent *model.Agent) (*model.Configuration, error) {
	if agent == nil {
		return nil, fmt.Errorf("cannot return configuration for nil agent")
	}

	// if Pending is specified, this is the new configuration we expect to have
	if agent.ConfigurationStatus.Pending != "" {
		return s.Configuration(ctx, agent.ConfigurationStatus.Pending)
	}

	// if Current is specified, this is the configuration that is currently applied
	if agent.ConfigurationStatus.Current != "" {
		return s.Configuration(ctx, agent.ConfigurationStatus.Current)
	}

	// ConfigurationStatus is not set, findAgentConfiguration
	return s.FindAgentConfiguration(ctx, agent)
}

// FindAgentConfiguration uses label matching to find the appropriate configuration for this agent. If a configuration
// is found that does not match the Current configuration, the configuration will be assigned to Future.
func (s *BoltstoreCore) FindAgentConfiguration(ctx context.Context, agent *model.Agent) (*model.Configuration, error) {
	// check for configuration= label and use that
	if configurationName, ok := agent.Labels.Set["configuration"]; ok {
		// if there is a configuration label, this takes precedence and we don't need to look any further
		configuration, err := s.Configuration(ctx, model.JoinVersion(configurationName, model.VersionCurrent))
		if configuration == nil {
			configuration, err = s.Configuration(ctx, configurationName)
		}
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve agent configuration: %w", err)
		}
		agent.SetFutureConfiguration(configuration)
		if configuration != nil && configuration.Status.Rollout.Status == model.RolloutStatusStable {
			return configuration, nil
		}
		return nil, nil
	}

	var result *model.Configuration

	err := s.DB.View(func(tx *bbolt.Tx) error {
		// iterate over the configurations looking for one that applies
		prefix := []byte(model.KindConfiguration)
		bucket, err := s.ResourcesBucket(ctx, tx, model.KindConfiguration)
		if err != nil {
			return err
		}
		cursor := bucket.Cursor()

		for k, v := cursor.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = cursor.Next() {
			configuration := &model.Configuration{}
			if err := jsoniter.Unmarshal(v, configuration); err != nil {
				s.ZapLogger().Error("unable to unmarshal configuration, ignoring", zap.Error(err))
				continue
			}
			if configuration.IsForAgent(agent) {
				result = configuration
				break
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("unable to retrieve agent configuration: %w", err)
	}
	agent.SetFutureConfiguration(result)
	return result, nil
}

// AgentsIDsMatchingConfiguration returns the list of agent IDs that are using the specified configuration
func (s *BoltstoreCore) AgentsIDsMatchingConfiguration(ctx context.Context, configuration *model.Configuration) ([]string, error) {
	ids := s.AgentIndex(ctx).Select(ctx, configuration.Spec.Selector.MatchLabels)
	return ids, nil
}

// Updates returns a channel that will receive updates when resources are added, updated, or deleted.
func (s *BoltstoreCore) Updates(_ context.Context) eventbus.Source[BasicEventUpdates] {
	return s.StoreUpdates.Updates()
}

// DeleteResources iterates threw a slice of resources, and removes them from storage by name.
// Sends any successful pipeline deletes to the pipelineDeletes channel, to be handled by the manager.
// Exporter and receiver deletes are sent to the manager via notifyUpdates.
func (s *boltstore) DeleteResources(ctx context.Context, resources []model.Resource) ([]model.ResourceStatus, error) {
	return s.DeleteResourcesCore(ctx, resources)
}

// DeleteResourcesCore iterates threw a slice of resources, and removes them from storage by name.
func (s *BoltstoreCore) DeleteResourcesCore(ctx context.Context, resources []model.Resource) ([]model.ResourceStatus, error) {
	updates := s.CreateEventUpdate()

	// track deleteStatuses to return
	deleteStatuses := make([]model.ResourceStatus, 0)

	for _, r := range resources {
		empty, err := model.NewEmptyResource(r.GetKind())
		if err != nil {
			deleteStatuses = append(deleteStatuses, *model.NewResourceStatusWithReason(r, model.StatusError, err.Error()))
			continue
		}

		deleted, exists, err := DeleteResource(ctx, s, r.GetKind(), r.UniqueKey(), empty)

		switch err.(type) {
		case *DependencyError:
			deleteStatuses = append(
				deleteStatuses,
				*model.NewResourceStatusWithReason(r, model.StatusInUse, err.Error()))
			continue

		case nil:
			break

		default:
			deleteStatuses = append(deleteStatuses, *model.NewResourceStatusWithReason(r, model.StatusError, err.Error()))
			continue
		}

		if !exists {
			continue
		}

		deleteStatuses = append(deleteStatuses, *model.NewResourceStatus(r, model.StatusDeleted))
		updates.IncludeResource(deleted, EventTypeRemove)
	}

	s.Notify(ctx, updates)
	return deleteStatuses, nil
}

// UpsertAgents upserts the agents with the given IDs.  If the agent does not exist, it will be created.
func (s *BoltstoreCore) UpsertAgents(ctx context.Context, agentIDs []string, updater AgentUpdater) ([]*model.Agent, error) {
	ctx, span := tracer.Start(ctx, "store/UpsertAgents")
	defer span.End()
	return s.updateOrUpsertAgents(ctx, false, agentIDs, updater)
}

// UpdateAgents updates existing Agents in the Store. If an agentID does not exist, that agentID is ignored and no
// agent corresponding to that ID will be returned. An error is only returned if the update fails.
func (s *BoltstoreCore) UpdateAgents(ctx context.Context, agentIDs []string, updater AgentUpdater) ([]*model.Agent, error) {
	ctx, span := tracer.Start(ctx, "store/UpdateAgents")
	defer span.End()
	return s.updateOrUpsertAgents(ctx, true, agentIDs, updater)
}

func (s *BoltstoreCore) updateOrUpsertAgents(ctx context.Context, requireExists bool, agentIDs []string, updater AgentUpdater) ([]*model.Agent, error) {
	agents := make([]*model.Agent, 0, len(agentIDs))
	updates := s.CreateEventUpdate()

	err := s.DB.Update(func(tx *bbolt.Tx) error {
		bucket, err := s.AgentsBucket(ctx, tx)
		if err != nil {
			return err
		}

		for _, agentID := range agentIDs {
			agent, err := s.updateOrUpsertAgentTx(ctx, requireExists, bucket, agentID, updater, updates)
			if err != nil {
				return err
			}
			if agent != nil {
				agents = append(agents, agent)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// update the search index with changes
	for _, a := range agents {
		if err := s.AgentsIndex(ctx).Upsert(ctx, a); err != nil {
			s.ZapLogger().Error("failed to update the search index", zap.String("agentID", a.ID))
		}
	}

	// notify results
	s.Notify(ctx, updates)
	return agents, nil
}

// UpsertAgent creates or updates the given agent and calls the updater method on it.
func (s *BoltstoreCore) UpsertAgent(ctx context.Context, agentID string, updater AgentUpdater) (*model.Agent, error) {
	ctx, span := tracer.Start(ctx, "store/UpsertAgent")
	defer span.End()
	return s.updateOrUpsertAgent(ctx, false, agentID, updater)
}

// UpdateAgent updates an existing Agent in the Store. If the agentID does not exist, no error is returned but the
// agent will be nil. An error is only returned if the update fails.
func (s *BoltstoreCore) UpdateAgent(ctx context.Context, agentID string, updater AgentUpdater) (*model.Agent, error) {
	ctx, span := tracer.Start(ctx, "store/UpsertAgent")
	defer span.End()
	return s.updateOrUpsertAgent(ctx, true, agentID, updater)
}

// UpdateAgentStatus will update the status of an existing agent. If the agentID does not exist, this does nothing. An
// error is only returned if updating the status of the agent fails.
//
// In boltstore, this uses UpdateAgent directly.
func (s *BoltstoreCore) UpdateAgentStatus(ctx context.Context, agentID string, status model.AgentStatus) error {
	_, err := s.UpdateAgent(ctx, agentID, func(current *model.Agent) {
		current.Status = status
	})
	return err
}

func (s *BoltstoreCore) updateOrUpsertAgent(ctx context.Context, requireExists bool, agentID string, updater AgentUpdater) (*model.Agent, error) {
	var updatedAgent *model.Agent
	updates := s.CreateEventUpdate()

	err := s.DB.Update(func(tx *bbolt.Tx) error {
		bucket, err := s.AgentsBucket(ctx, tx)
		if err != nil {
			return err
		}

		agent, err := s.updateOrUpsertAgentTx(ctx, requireExists, bucket, agentID, updater, updates)
		updatedAgent = agent
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	if updatedAgent == nil {
		return nil, nil
	}

	// update the index
	err = s.AgentsIndex(ctx).Upsert(ctx, updatedAgent)
	if err != nil {
		s.ZapLogger().Error("failed to update the search index", zap.String("agentID", updatedAgent.ID))
	}

	s.Notify(ctx, updates)

	return updatedAgent, nil
}

// Agents returns all agents in the store with the given options.
func (s *BoltstoreCore) Agents(ctx context.Context, options ...QueryOption) ([]*model.Agent, error) {
	opts := MakeQueryOptions(options)

	// search is implemented using the search index
	if opts.Query != nil {
		ids, err := s.AgentsIndex(ctx).Search(ctx, opts.Query)
		if err != nil {
			return nil, err
		}
		return s.agentsByID(ctx, ids, opts)
	}

	agents := []*model.Agent{}

	err := s.DB.View(func(tx *bbolt.Tx) error {
		bucket, err := s.AgentsBucket(ctx, tx)
		if err != nil {
			return err
		}
		cursor := bucket.Cursor()
		prefix := AgentPrefix()

		for k, v := cursor.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = cursor.Next() {
			agent := &model.Agent{}
			if err := jsoniter.Unmarshal(v, agent); err != nil {
				return fmt.Errorf("agents: %w", err)
			}

			if opts.Selector.Matches(agent.Labels) {
				agents = append(agents, agent)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if opts.Sort == "" {
		opts.Sort = "name"
	}
	return ApplySortOffsetAndLimit(agents, opts, func(field string, item *model.Agent) string {
		switch field {
		case "id":
			return item.ID
		default:
			return item.Name
		}
	}), nil
}

func (s *BoltstoreCore) agentsByID(ctx context.Context, ids []string, opts QueryOptions) ([]*model.Agent, error) {
	var agents []*model.Agent

	err := s.DB.View(func(tx *bbolt.Tx) error {
		bucket, err := s.AgentsBucket(ctx, tx)
		if err != nil {
			return err
		}

		for _, id := range ids {
			data := bucket.Get(AgentKey(id))
			if data == nil {
				return nil
			}
			agent := &model.Agent{}
			if err := jsoniter.Unmarshal(data, agent); err != nil {
				return fmt.Errorf("agents: %w", err)
			}

			if opts.Selector.Matches(agent.Labels) {
				agents = append(agents, agent)
			}
		}
		return nil
	})

	return agents, err
}

// AgentsCount returns the number of agents in the store with the given options.
func (s *BoltstoreCore) AgentsCount(ctx context.Context, options ...QueryOption) (int, error) {
	agents, err := s.Agents(ctx, options...)
	if err != nil {
		return -1, err
	}
	return len(agents), nil
}

// Agent returns the agent with the given ID.
func (s *BoltstoreCore) Agent(ctx context.Context, id string) (*model.Agent, error) {
	var agent *model.Agent

	err := s.DB.View(func(tx *bbolt.Tx) error {
		bucket, err := s.AgentsBucket(ctx, tx)
		if err != nil {
			return err
		}
		data := bucket.Get(AgentKey(id))
		if data == nil {
			return nil
		}
		agent = &model.Agent{}
		return jsoniter.Unmarshal(data, agent)
	})

	return agent, err
}

// AgentVersion returns the agent version with the given name.
func (s *BoltstoreCore) AgentVersion(ctx context.Context, name string) (*model.AgentVersion, error) {
	item, exists, err := Resource[*model.AgentVersion](ctx, s, model.KindAgentVersion, name)
	if !exists {
		item = nil
	}
	return item, err
}

// AgentVersions returns all agent versions in the store with the given options.
func (s *BoltstoreCore) AgentVersions(ctx context.Context) ([]*model.AgentVersion, error) {
	result, err := Resources[*model.AgentVersion](ctx, s, model.KindAgentVersion)
	if err == nil {
		model.SortAgentVersionsLatestFirst(result)
	}
	return result, err
}

// DeleteAgentVersion deletes the agent version with the given name.
func (s *BoltstoreCore) DeleteAgentVersion(ctx context.Context, name string) (*model.AgentVersion, error) {
	item, exists, err := DeleteResourceAndNotify(ctx, s, model.KindAgentVersion, name, &model.AgentVersion{})
	if !exists {
		return nil, err
	}
	return item, err
}

// Configurations returns the configurations in the store with the given options.
func (s *BoltstoreCore) Configurations(ctx context.Context, options ...QueryOption) ([]*model.Configuration, error) {
	opts := MakeQueryOptions(options)
	// search is implemented using the search index
	if opts.Query != nil {
		names, err := s.ConfigurationsIndex(ctx).Search(ctx, opts.Query)
		if err != nil {
			return nil, err
		}
		return ResourcesByUniqueKeys[*model.Configuration](ctx, s, model.KindConfiguration, names, opts)
	}

	return resourcesWithFilter(ctx, s, model.KindConfiguration, func(c *model.Configuration) bool {
		return opts.Selector.Matches(c.GetLabels())
	})
}

// Configuration returns the configuration with the given name.
func (s *BoltstoreCore) Configuration(ctx context.Context, name string) (*model.Configuration, error) {
	item, exists, err := Resource[*model.Configuration](ctx, s, model.KindConfiguration, name)
	if !exists {
		item = nil
	}
	return item, err
}

// UpdateConfiguration updates the configuration with the given name using the updater function.
func (s *BoltstoreCore) UpdateConfiguration(ctx context.Context, name string, updater ConfigurationUpdater) (config *model.Configuration, status model.UpdateStatus, err error) {
	ctx, span := tracer.Start(ctx, "store/UpdateConfiguration")
	defer span.End()

	updates := s.CreateEventUpdate()

	err = s.DB.Update(func(tx *bbolt.Tx) error {
		config, status, err = UpdateResource(ctx, s, tx, model.KindConfiguration, name, func(config *model.Configuration) error {
			updater(config)
			return nil
		})
		if err != nil {
			return err
		}
		if config == nil {
			return nil
		}

		switch status {
		case model.StatusCreated:
			updates.IncludeResource(config, EventTypeInsert)
		case model.StatusConfigured:
			updates.IncludeResource(config, EventTypeUpdate)
		}
		err = s.ConfigurationsIndex(ctx).Upsert(ctx, config)
		if err != nil {
			s.Logger.Error("failed to update the search index", zap.String("configuration", name), zap.Error(err))
		}
		return nil
	})

	s.Notify(ctx, updates)

	return config, status, err
}

// DeleteConfiguration deletes the configuration with the given name.
func (s *BoltstoreCore) DeleteConfiguration(ctx context.Context, name string) (*model.Configuration, error) {
	item, exists, err := DeleteResourceAndNotify(ctx, s, model.KindConfiguration, name, &model.Configuration{})
	if !exists {
		return nil, err
	}
	return item, err
}

// Source returns the source with the given name.
func (s *BoltstoreCore) Source(ctx context.Context, name string) (*model.Source, error) {
	item, exists, err := Resource[*model.Source](ctx, s, model.KindSource, name)
	if !exists {
		item = nil
	}
	return item, err
}

// Sources returns the sources in the store.
func (s *BoltstoreCore) Sources(ctx context.Context) ([]*model.Source, error) {
	return Resources[*model.Source](ctx, s, model.KindSource)
}

// DeleteSource deletes the source with the given name.
func (s *BoltstoreCore) DeleteSource(ctx context.Context, name string) (*model.Source, error) {
	item, exists, err := DeleteResourceAndNotify(ctx, s, model.KindSource, name, &model.Source{})
	if !exists {
		return nil, err
	}
	return item, err
}

// SourceType returns the source type with the given name.
func (s *BoltstoreCore) SourceType(ctx context.Context, name string) (*model.SourceType, error) {
	item, exists, err := Resource[*model.SourceType](ctx, s, model.KindSourceType, name)
	if !exists {
		item = nil
	}
	return item, err
}

// SourceTypes returns the source types in the store.
func (s *BoltstoreCore) SourceTypes(ctx context.Context) ([]*model.SourceType, error) {
	return Resources[*model.SourceType](ctx, s, model.KindSourceType)
}

// DeleteSourceType deletes the source type with the given name.
func (s *BoltstoreCore) DeleteSourceType(ctx context.Context, name string) (*model.SourceType, error) {
	item, exists, err := DeleteResourceAndNotify(ctx, s, model.KindSourceType, name, &model.SourceType{})
	if !exists {
		return nil, err
	}
	return item, err
}

// Processor returns the processor with the given name.
func (s *BoltstoreCore) Processor(ctx context.Context, name string) (*model.Processor, error) {
	item, exists, err := Resource[*model.Processor](ctx, s, model.KindProcessor, name)
	if !exists {
		item = nil
	}
	return item, err
}

// Processors returns the processors in the store.
func (s *BoltstoreCore) Processors(ctx context.Context) ([]*model.Processor, error) {
	return Resources[*model.Processor](ctx, s, model.KindProcessor)
}

// DeleteProcessor deletes the processor with the given name.
func (s *BoltstoreCore) DeleteProcessor(ctx context.Context, name string) (*model.Processor, error) {
	item, exists, err := DeleteResourceAndNotify(ctx, s, model.KindProcessor, name, &model.Processor{})
	if !exists {
		return nil, err
	}
	return item, err
}

// ProcessorType returns the processor type with the given name.
func (s *BoltstoreCore) ProcessorType(ctx context.Context, name string) (*model.ProcessorType, error) {
	item, exists, err := Resource[*model.ProcessorType](ctx, s, model.KindProcessorType, name)
	if !exists {
		item = nil
	}
	return item, err
}

// ProcessorTypes returns the processor types in the store.
func (s *BoltstoreCore) ProcessorTypes(ctx context.Context) ([]*model.ProcessorType, error) {
	return Resources[*model.ProcessorType](ctx, s, model.KindProcessorType)
}

// DeleteProcessorType deletes the processor type with the given name.
func (s *BoltstoreCore) DeleteProcessorType(ctx context.Context, name string) (*model.ProcessorType, error) {
	item, exists, err := DeleteResourceAndNotify(ctx, s, model.KindProcessorType, name, &model.ProcessorType{})
	if !exists {
		return nil, err
	}
	return item, err
}

// Destination returns the destination with the given name.
func (s *BoltstoreCore) Destination(ctx context.Context, name string) (*model.Destination, error) {
	item, exists, err := Resource[*model.Destination](ctx, s, model.KindDestination, name)
	if !exists {
		item = nil
	}
	return item, err
}

// Destinations returns the destinations in the store.
func (s *BoltstoreCore) Destinations(ctx context.Context) ([]*model.Destination, error) {
	return Resources[*model.Destination](ctx, s, model.KindDestination)
}

// DeleteDestination deletes the destination with the given name.
func (s *BoltstoreCore) DeleteDestination(ctx context.Context, name string) (*model.Destination, error) {
	item, exists, err := DeleteResourceAndNotify(ctx, s, model.KindDestination, name, &model.Destination{})
	if !exists {
		return nil, err
	}
	return item, err
}

// DestinationType returns the destination type with the given name.
func (s *BoltstoreCore) DestinationType(ctx context.Context, name string) (*model.DestinationType, error) {
	item, exists, err := Resource[*model.DestinationType](ctx, s, model.KindDestinationType, name)
	if !exists {
		item = nil
	}
	return item, err
}

// DestinationTypes returns the destination types in the store.
func (s *BoltstoreCore) DestinationTypes(ctx context.Context) ([]*model.DestinationType, error) {
	return Resources[*model.DestinationType](ctx, s, model.KindDestinationType)
}

// DeleteDestinationType deletes the destination type with the given name.
func (s *BoltstoreCore) DeleteDestinationType(ctx context.Context, name string) (*model.DestinationType, error) {
	item, exists, err := DeleteResourceAndNotify(ctx, s, model.KindDestinationType, name, &model.DestinationType{})
	if !exists {
		return nil, err
	}
	return item, err
}

// ReportConnectedAgents sets the ReportedAt time for the specified agents to the specified time. This update should
// not fire an update event for the agents on the Updates eventbus.
func (s *BoltstoreCore) ReportConnectedAgents(ctx context.Context, agentIDs []string, time time.Time) error {
	ctx, span := tracer.Start(ctx, "store/ReportConnectedAgents")
	defer span.End()

	// these updates will not be reported to the eventbus
	updates := s.CreateEventUpdate()

	err := s.DB.Update(func(tx *bbolt.Tx) error {
		bucket, err := s.AgentsBucket(ctx, tx)
		if err != nil {
			return fmt.Errorf("failed to get agents bucket: %w", err)
		}

		var errs error

		for _, agentID := range agentIDs {
			_, err := s.updateOrUpsertAgentTx(ctx, true, bucket, agentID, func(current *model.Agent) {
				current.ReportedAt = &time
			}, updates)
			errs = errors.Join(errs, err)
		}

		return errs
	})

	// ignore updates gathered above to avoid filling the eventbus with frequent ReportedAt updates

	return err
}

// DisconnectUnreportedAgents sets the Status of agents to Disconnected if the agent ReportedAt time is before the
// specified time.
func (s *BoltstoreCore) DisconnectUnreportedAgents(ctx context.Context, since time.Time) error {
	ctx, span := tracer.Start(ctx, "store/DisconnectUnreportedAgents")
	defer span.End()

	changes := s.CreateEventUpdate()

	err := s.DB.Update(func(tx *bbolt.Tx) error {
		bucket, err := s.AgentsBucket(ctx, tx)
		if err != nil {
			return fmt.Errorf("failed to get agents bucket: %w", err)
		}
		cursor := bucket.Cursor()
		prefix := AgentPrefix()

		for k, v := cursor.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = cursor.Next() {
			agent := &model.Agent{}
			if err := jsoniter.Unmarshal(v, agent); err != nil {
				// unable to unmarshal, ignore this agent
				s.Logger.Info("unable to unmarshal agent, ignoring", zap.Error(err))
				continue
			}
			if agent.ReportedSince(since) {
				// agent reported recently, nothing to do
				continue
			}
			if agent.Status == model.Disconnected {
				// agent already disconnected, nothing to do
				continue
			}
			agent.Disconnect()

			// marshal it back to to json
			data, err := jsoniter.Marshal(agent)
			if err != nil {
				return fmt.Errorf("failed to marshal agent: %w", err)
			}

			// update the agent in the store
			err = bucket.Put(k, data)
			if err != nil {
				return fmt.Errorf("failed to update agent: %w", err)
			}

			changes.IncludeAgent(agent, EventTypeUpdate)
		}

		return nil
	})
	s.Notify(ctx, changes)

	return err
}

// CleanupDisconnectedAgents removes all containerized agents that have been disconnected since the given time
func (s *BoltstoreCore) CleanupDisconnectedAgents(ctx context.Context, since time.Time) error {
	// get all agents with the container-platform label
	agents, err := s.Agents(ctx, WithQuery(search.ParseQuery(fmt.Sprintf("%s:", model.LabelAgentContainerPlatform))))
	if err != nil {
		return err
	}

	var errs error
	changes := s.CreateEventUpdate()

	for _, agent := range agents {
		if agent.DisconnectedSince(since) {
			err := s.DB.Update(func(tx *bbolt.Tx) error {
				bucket, err := s.AgentsBucket(ctx, tx)
				if err != nil {
					return err
				}
				return bucket.Delete(AgentKey(agent.ID))
			})
			if err != nil {
				errs = errors.Join(errs, err)
				continue
			}
			changes.IncludeAgent(agent, EventTypeRemove)

			// update the index
			if err := s.AgentsIndex(ctx).Remove(ctx, agent); err != nil {
				s.Logger.Error("failed to remove from the search index", zap.String("agentID", agent.ID))
			}
		}
	}

	s.Notify(ctx, changes)
	return errs
}

// AgentIndex provides access to the search Index implementation managed by the Store
func (s *BoltstoreCore) AgentIndex(ctx context.Context) search.Index {
	return s.AgentsIndex(ctx)
}

// ConfigurationIndex provides access to the search Index for Configurations
func (s *BoltstoreCore) ConfigurationIndex(ctx context.Context) search.Index {
	return s.ConfigurationsIndex(ctx)
}

// Database returns the underlying bbolt database
func (s *BoltstoreCore) Database() *bbolt.DB {
	return s.DB
}

// StartRollout will start a rollout for the specified configuration with the specified options. If nil is passed for
// options, any existing rollout options on the configuration status will be used. If there are no rollout options in
// the configuration status, default values will be used for the rollout. If there is an existing rollout a different
// version of this configuration, it will be replaced. Does nothing if the rollout does not have a RolloutStatusPending
// status. Returns the current Configuration with its Rollout status.
func (s *BoltstoreCore) StartRollout(ctx context.Context, configurationName string, options *model.RolloutOptions) (*model.Configuration, error) {
	config, err := s.Configuration(ctx, configurationName)
	if err != nil || config == nil {
		return config, err
	}

	switch config.Status.Rollout.Status {
	case model.RolloutStatusStarted:
		// if the rollout is already started, we don't need to do anything
		return config, nil

	case model.RolloutStatusPaused:
		// if the rollout is paused, we need to resume it
		return s.ResumeRollout(ctx, configurationName)
	}

	nameAndVersion := config.NameAndVersion()

	// if there is already a rollout in progress, we need to replace it
	rollouts, err := CurrentRolloutsForConfiguration(ctx, s.AgentIndex(ctx), config.Name())
	if err != nil {
		return nil, err
	}
	for _, rollout := range rollouts {
		if rollout == nameAndVersion {
			// this exact rollout already exists in some state, so we need to restart it
			continue
		}
		// a different rollout exists for this configuration, replace it
		_, _, err = editResource(ctx, s, nil, model.KindConfiguration, rollout, func(config *model.Configuration) error {
			config.Status.Rollout.Status = model.RolloutStatusReplaced
			return nil
		})
	}

	agentIDs, err := s.AgentsIDsMatchingConfiguration(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("agentIDs matching configuration: %w", err)
	}
	// set the rollout options and start the rollout
	if options == nil {
		defaultOptions := model.RolloutOptionsForAgentCount(len(agentIDs))
		options = &defaultOptions
	}
	config, _, err = editResource(ctx, s, nil, model.KindConfiguration, configurationName, func(config *model.Configuration) error {
		config.Status.Rollout.Status = model.RolloutStatusStarted
		config.Status.Rollout.Phase = 0
		config.Status.Rollout.Options = *options
		return nil
	})
	if err != nil {
		return nil, err
	}
	// track the pending version on the latest configuration
	_, _, err = editResource(ctx, s, nil, model.KindConfiguration, model.JoinVersion(config.Name(), model.VersionLatest), func(latest *model.Configuration) error {
		latest.Status.PendingVersion = config.Version()
		return nil
	})
	if err != nil {
		return nil, err
	}

	// set future configuration for all agents using this configuration
	_, err = s.UpdateAgents(ctx, agentIDs, func(agent *model.Agent) {
		agent.SetFutureConfiguration(config)
	})
	if err != nil {
		return nil, fmt.Errorf("UpdateAgents to SetFutureConfiguration: %w", err)
	}

	// run the first phase of the rollout
	return s.UpdateRollout(ctx, configurationName)
}

// PauseRollout will pause a rollout for the specified configuration. Does nothing if the rollout does not have a
// RolloutStatusStarted status. Returns the current Configuration with its Rollout status.
func (s *BoltstoreCore) PauseRollout(ctx context.Context, configurationName string) (*model.Configuration, error) {
	config, err := s.Configuration(ctx, configurationName)
	if err != nil || config == nil {
		return nil, err
	}

	// pause does nothing if its not started
	if config.Status.Rollout.Status != model.RolloutStatusStarted {
		return config, nil
	}

	// set the rollout status to paused
	_, _, err = editResource(ctx, s, nil, model.KindConfiguration, configurationName, func(config *model.Configuration) error {
		config.Status.Rollout.Status = model.RolloutStatusPaused
		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.UpdateRollout(ctx, configurationName)
}

// ResumeRollout will resume a rollout for the specified configuration.
// Does nothing if the Rollout status is not RolloutStatusStarted or RolloutStatusStarted.
// For RolloutStatusError - it will increase the maxErrors of the
// rollout by the current number of errors + 1.
// For RolloutStatusStarted - it will pause the rollout.
func (s *BoltstoreCore) ResumeRollout(ctx context.Context, configurationName string) (*model.Configuration, error) {
	config, err := s.Configuration(ctx, configurationName)
	if err != nil || config == nil {
		return nil, err
	}

	if config.Status.Rollout.Status == model.RolloutStatusPaused {
		// set the rollout status to started
		config, _, err = editResource(ctx, s, nil, model.KindConfiguration, configurationName, func(config *model.Configuration) error {
			config.Status.Rollout.Status = model.RolloutStatusStarted
			return nil
		})
	}

	if config.Status.Rollout.Status == model.RolloutStatusError {
		// increase the maxErrors of the rollout to the current number of errors + the rollout.MaxErrors value
		config, _, err = editResource(ctx, s, nil, model.KindConfiguration, configurationName, func(config *model.Configuration) error {
			config.Status.Rollout.Options.MaxErrors = config.Status.Rollout.Progress.Errors + 1
			config.Status.Rollout.Status = model.RolloutStatusStarted
			return nil
		})
	}

	if err != nil || config == nil {
		return nil, err
	}

	return s.UpdateRollout(ctx, configurationName)
}

// UpdateRollout updates a rollout in progress. Does nothing if the rollout does not have a RolloutStatusStarted
// status. Returns the current Configuration with its Rollout status.
func (s *BoltstoreCore) UpdateRollout(ctx context.Context, configuration string) (updatedConfig *model.Configuration, err error) {
	updates := s.CreateEventUpdate()

	// get the configuration
	err = s.DB.Update(func(tx *bbolt.Tx) error {
		var (
			agentsWaiting []string
			agentsNext    int
			rolloutStatus model.RolloutStatus
		)
		oldRolloutStatus := model.RolloutStatusPending
		config, wasModified, err := editResource(ctx, s, tx, model.KindConfiguration, configuration, func(r *model.Configuration) error {
			oldRolloutStatus = r.Status.Rollout.Status

			agentsWaiting, agentsNext, err = UpdateRolloutMetrics(ctx, s.AgentIndex(ctx), r)
			if err != nil {
				return err
			}
			rolloutStatus = r.Status.Rollout.Status
			if rolloutStatus == model.RolloutStatusStable {
				r.Status.CurrentVersion = r.Version()
			}

			return nil
		})
		if err != nil || config == nil {
			return err
		}
		if rolloutStatus != oldRolloutStatus && rolloutStatus == model.RolloutStatusStable {
			// update current version
			err := s.updateCurrentVersion(ctx, tx, config)
			if err != nil {
				return err
			}
			config.SetCurrent(true)
		}

		updatedConfig = config

		if !wasModified {
			return nil
		}

		updates.IncludeResource(config, EventTypeUpdate)

		if agentsNext > 0 {
			var agents []*model.Agent

			// update the next batch of agents
			agentsBucket, err := s.AgentsBucket(ctx, tx)
			if err != nil {
				return fmt.Errorf("failed to get agents bucket: %w", err)
			}
			// ensure we don't try to update more agents than we have
			if agentsNext > len(agentsWaiting) {
				agentsNext = len(agentsWaiting)
			}
			agentIDs := agentsWaiting[:agentsNext]
			for _, agentID := range agentIDs {
				agent, err := s.updateOrUpsertAgentTx(ctx, true, agentsBucket, agentID, func(agent *model.Agent) {
					agent.SetPendingConfiguration(config)
				}, updates)
				if err != nil {
					return err
				}
				agents = append(agents, agent)
			}

			// update the search index with changes
			for _, a := range agents {
				if err := s.AgentsIndex(ctx).Upsert(ctx, a); err != nil {
					s.Logger.Error("failed to update the search index", zap.String("agentID", a.ID))
				}
			}
		}

		return err
	})

	// config may have changed
	s.Notify(ctx, updates)

	return
}

// UpdateRollouts updates all rollouts in progress. It returns each of the Configurations that contains an active
// rollout.
func (s *BoltstoreCore) UpdateRollouts(ctx context.Context) ([]*model.Configuration, error) {
	activeRollouts, err := s.currentRollouts(ctx)
	if err != nil {
		return nil, err
	}

	present := struct{}{}
	updated := map[string]struct{}{}

	var errs error

	result := make([]*model.Configuration, 0, len(activeRollouts))
	for _, activeRollout := range activeRollouts {
		name, _ := model.SplitVersion(activeRollout)
		if _, ok := updated[name]; ok {
			continue
		}
		configuration, err := s.UpdateRollout(ctx, name)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		if configuration != nil {
			result = append(result, configuration)
		}
		// assume we updated it so we don't try to request the same configuration again
		updated[name] = present
	}

	return result, errs
}

func (s *BoltstoreCore) currentRollouts(ctx context.Context) ([]string, error) {
	return StartedRolloutsFromIndex(ctx, s.ConfigurationsIndex(ctx))
}

// updateCurrentVersion updates the CurrentVersion of a Configuration resource which is maintained by the most recent
// version. If a Rollout completes for a specific version, that version becomes the current version.
func (s *BoltstoreCore) updateCurrentVersion(ctx context.Context, tx *bbolt.Tx, configVersion *model.Configuration) error {
	// pass configVersion.Name() to edit the latest version of the configuration
	_, _, err := editResource(ctx, s, tx, model.KindConfiguration, configVersion.Name(), func(r *model.Configuration) error {
		r.Status.CurrentVersion = configVersion.Version()
		return nil
	})
	return err
}

// ----------------------------------------------------------------------
// ArchiveStore

// ResourceHistory returns the history of a resource given its kind and name.
func (s *BoltstoreCore) ResourceHistory(ctx context.Context, resourceKind model.Kind, resourceName string) ([]*model.AnyResource, error) {
	return resourceHistory[*model.AnyResource](ctx, s, resourceKind, resourceName)
}

// -----------------------------------------------------------------------------

// getObjectIds will retrieve identifiers for all objects in a bucket where the keys are formatted KIND|IDENTIFIER
func (s *BoltstoreCore) getObjectIds(bucketFunc func(tx *bbolt.Tx) (*bbolt.Bucket, error), kind model.Kind) ([]string, error) {
	ids := []string{}
	prefix := []byte(fmt.Sprintf("%s|", kind))
	err := s.Database().View(func(tx *bbolt.Tx) error {
		bucket, err := bucketFunc(tx)
		if err != nil {
			return nil
		}
		cursor := bucket.Cursor()
		for k, _ := cursor.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = cursor.Next() {
			if _, id, found := strings.Cut(string(k), "|"); found {
				ids = append(ids, id)
			}
		}
		return nil
	})
	return ids, err
}

// ----------------------------------------------------------------------
// Measurements implementation

const measurementsDateFormat = "2006-01-02T15:04:05"

// AgentMetrics provides metrics for an individual agents. They are essentially configuration metrics filtered to a
// list of agents.
//
// Note: While the same record.Metric struct is used to return the metrics, these are not the same metrics provided to
// Store. They will be aggregated and counter metrics will be converted into rates.
func (s *BoltstoreCore) AgentMetrics(ctx context.Context, ids []string, options ...stats.QueryOption) (stats.MetricData, error) {
	// Empty string single key or empty array of ids is a request for all Agents
	if len(ids) == 0 || (len(ids) == 1 && ids[0] == "") {
		var err error
		ids, err = s.getObjectIds(func(tx *bbolt.Tx) (*bbolt.Bucket, error) {
			return s.AgentsBucket(ctx, tx)
		}, model.KindAgent)
		if err != nil {
			return nil, err
		}
	}
	return s.retrieveMetrics(ctx, stats.SupportedMetricNames, string(model.KindAgent), ids, options...)
}

// ConfigurationMetrics provides all metrics associated with a configuration aggregated from all agents using the
// configuration.
//
// Note: While the same record.Metric struct is used to return the metrics, these are not the same metrics provided to
// Store. They will be aggregated and counter metrics will be converted into rates.
func (s *BoltstoreCore) ConfigurationMetrics(ctx context.Context, name string, options ...stats.QueryOption) (stats.MetricData, error) {
	names := []string{name}
	var err error
	// Empty name is a request for all configurations
	if name == "" {
		names, err = s.getObjectIds(func(tx *bbolt.Tx) (*bbolt.Bucket, error) {
			return s.ResourcesBucket(ctx, tx, model.KindConfiguration)
		}, model.KindConfiguration)
		if err != nil {
			return nil, err
		}
	}

	baseMetrics, err := s.retrieveMetrics(ctx, stats.SupportedMetricNames, string(model.KindConfiguration), names, options...)
	if err != nil {
		return nil, err
	}

	groupedMetrics := map[string]stats.MetricData{}
	for _, m := range baseMetrics {
		// since multiple configurations may be returned for the overview page, group by configuration and processor
		key := fmt.Sprintf("%s|%s", stats.Configuration(m), stats.Processor(m))
		groupedMetrics[key] = append(groupedMetrics[key], m)
	}

	finalMetrics := stats.MetricData{}
	for _, metrics := range groupedMetrics {
		sum := 0.0
		for _, m := range metrics {
			val, _ := stats.Value(m)
			sum += val
		}

		attributes := map[string]interface{}{
			stats.ConfigurationAttributeName: metrics[0].Attributes[stats.ConfigurationAttributeName],
			stats.ProcessorAttributeName:     metrics[0].Attributes[stats.ProcessorAttributeName],
		}

		m := generateRecord(metrics[0], metrics[0].StartTimestamp, sum, attributes)

		finalMetrics = append(finalMetrics, m)
	}

	return finalMetrics, nil
}

// OverviewMetrics provides all metrics needed for the overview page. This page shows configs and destinations.
func (s *BoltstoreCore) OverviewMetrics(ctx context.Context, options ...stats.QueryOption) (stats.MetricData, error) {
	return s.ConfigurationMetrics(ctx, "", options...)
}

// MeasurementsSize returns the count of keys in the store, and is used only for testing
func (s *BoltstoreCore) MeasurementsSize(ctx context.Context) (int, error) {
	count := 0
	err := s.Database().View(func(tx *bbolt.Tx) error {
		for _, metricName := range stats.SupportedMetricNames {
			bucket, err := s.MeasurementsBucket(ctx, tx, metricName)
			if err != nil {
				return err
			}
			count += bucket.Stats().KeyN
		}

		return nil
	})

	return count, err
}

// available as a function so it can be mocked in tests
var getCurrentTime = func() time.Time {
	return time.Now().UTC()
}

func (s *BoltstoreCore) retrieveMetrics(ctx context.Context, metricNames []string, objectType string, ids []string, options ...stats.QueryOption) (stats.MetricData, error) {
	result := stats.MetricData{}
	opts := stats.MakeQueryOptions(options)
	rollup := stats.GetDurationFromPeriod(opts)

	endDate := getCurrentTime().Add(-10 * time.Second).Truncate(10 * time.Second)
	startDate := endDate.Add(-1 * opts.Period).Truncate(rollup)

	endDateString := endDate.Format(measurementsDateFormat)
	startDateString := startDate.Format(measurementsDateFormat)

	err := s.Database().View(func(tx *bbolt.Tx) error {
		var errs error
		for metricIndex, metricName := range metricNames {
			mBucket, err := s.MeasurementsBucket(ctx, tx, metricName)
			if err != nil || mBucket == nil {
				return err
			}
			cursor := mBucket.Cursor()

			for idIndex, id := range ids {
				endMetrics, err := findEndMetrics(cursor, endDateString, objectType, id)
				if err != nil {
					errs = multierror.Append(errs, err)
					continue
				}

				// On the first metric & agent, check if a previous time would offer more complete data
				if metricIndex == 0 && idIndex == 0 {
					prevDate := endDate.Add(-10 * time.Second)
					prevDateString := prevDate.Format(measurementsDateFormat)
					// Ignore this error and just act on endMetrics if an error occurred
					prevMetrics, _ := findEndMetrics(cursor, prevDateString, objectType, id)

					// If there were _more_ metrics in the latest bucket, assume that the current bucket
					// has not been filled yet. This could also mean that there were configuration/agent
					// changes that led to less resources having metrics in the latest bucket, but we
					// would rather show stale data for 1 extra period (10s usually) than show data that was incorrect
					if len(endMetrics) < len(prevMetrics) {
						endMetrics = prevMetrics
						endDate = prevDate
						endDateString = prevDateString

						startDate = endDate.Add(-1 * opts.Period).Truncate(rollup)
						startDateString = startDate.Format(measurementsDateFormat)
					}
				}

				if len(endMetrics) == 0 {
					continue
				}

				// List the keys for which we want to find matching start points. This will be modified as
				// data is found in findStartMetrics
				desiredKeys := map[string]interface{}{}
				for k := range endMetrics {
					desiredKeys[k] = true
				}

				startMetrics, err := findStartMetrics(cursor, startDateString, endDate, objectType, id, desiredKeys)
				if err != nil {
					errs = multierror.Append(errs, err)
					continue
				}

				// Only calculate rates if we found a start point to match the end points desired
				for key, first := range startMetrics {
					if last, ok := endMetrics[key]; ok {
						if metric := s.calculateRateMetric(first, last); metric != nil {
							result = append(result, metric)
						}
					}
				}
			}
		}

		return errs
	})
	return result, err
}

func findEndMetrics(c *bbolt.Cursor, time, objectType, id string) (map[string]*record.Metric, error) {
	identifier := fmt.Sprintf("%s|%s", objectType, sanitizeKey(id))
	prefix := []byte(fmt.Sprintf("%s|%s|", identifier, time))
	metrics := map[string]*record.Metric{}

	var k, v []byte
	for k, v = c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
		var m *record.Metric
		m = &record.Metric{}
		if err := jsoniter.Unmarshal(v, m); err != nil {
			return nil, err
		}
		metrics[removeTimestampFromKey(k)] = m
	}

	return metrics, nil
}

func findStartMetrics(c *bbolt.Cursor, time string, endTime time.Time, objectType, id string, desiredKeys map[string]interface{}) (map[string]*record.Metric, error) {
	identifier := fmt.Sprintf("%s|%s", objectType, sanitizeKey(id))
	startingPrefix := []byte(fmt.Sprintf("%s|%s|", identifier, time))
	metrics := map[string]*record.Metric{}

	continueToSeek := func(k []byte) bool {
		// Stop searching if we've satisfied all of the keys we were looking for
		if len(desiredKeys) == 0 {
			return false
		}
		// Stop searching if we've hit the end of the data
		if k == nil {
			return false
		}

		// Stop searching if we're no longer looking at data for the right object
		if !bytes.HasPrefix(k, []byte(identifier)) {
			return false
		}

		return true
	}

	var k, v []byte
	for k, v = c.Seek(startingPrefix); continueToSeek(k); k, v = c.Next() {
		// If we've been passed a desiredKeys list, only unmarshal & process
		// data points on the requested list
		keyWithoutTimestamp := removeTimestampFromKey(k)
		if _, ok := desiredKeys[keyWithoutTimestamp]; !ok {
			continue
		}
		var m *record.Metric
		m = &record.Metric{}
		if err := jsoniter.Unmarshal(v, m); err != nil {
			return nil, err
		}

		// If we've reached or passed the endTime for this query, stop searching
		if m.Timestamp.Sub(endTime) >= 0 {
			break
		}

		metrics[keyWithoutTimestamp] = m
		delete(desiredKeys, keyWithoutTimestamp)
	}

	return metrics, nil
}

func removeTimestampFromKey(k []byte) string {
	keyParts := strings.Split(string(k), "|")
	keyParts = append(keyParts[:2], keyParts[3:]...)
	return strings.Join(keyParts, "|")
}

// SaveAgentMetrics saves new metrics. These metrics will be aggregated to determine metrics associated with agents and configurations.
func (s *BoltstoreCore) SaveAgentMetrics(ctx context.Context, metrics []*record.Metric) error {
	groupedMetrics := map[string][]*record.Metric{}
	for _, m := range metrics {
		groupedMetrics[m.Name] = append(groupedMetrics[m.Name], m)
	}

	var errs error
	for _, metricName := range stats.SupportedMetricNames {
		if group, ok := groupedMetrics[metricName]; ok {
			err := s.storeMeasurements(ctx, metricName, group)
			if err != nil {
				errs = multierror.Append(errs, err)
			}
		}
	}

	return errs
}

// ProcessMetrics is called in the background at regular intervals and performs metric roll-up and removes old data
func (s *BoltstoreCore) ProcessMetrics(ctx context.Context) error {
	var errs error
	for _, m := range stats.SupportedMetricNames {
		if err := s.cleanupMeasurements(ctx, m); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	if errs != nil {
		s.ZapLogger().Error("error cleaning up measurements", zap.Error(errs))
	}
	return errs
}

// StartMeasurements starts the background process for writing and rolling up metrics
func (s *BoltstoreCore) StartMeasurements(ctx context.Context) {
	// start the measurements timer for writing & rolling up
	s.Logger.Info("starting measurements cleanup", zap.Duration("interval", cleanupMeasurementsInterval))
	go func() {
		measurementsTicker := time.NewTicker(cleanupMeasurementsInterval)
		defer measurementsTicker.Stop()

		for {
			select {
			case <-measurementsTicker.C:
				// periodically clean up old measurements
				_ = s.ProcessMetrics(ctx)
			case <-ctx.Done():
				// send anything left in the buffer before stopping
				return
			}
		}
	}()
}

func (s *BoltstoreCore) storeMeasurements(ctx context.Context, metricName string, metrics stats.MetricData) error {
	if len(metrics) == 0 {
		return nil
	}
	return s.Database().Update(func(tx *bbolt.Tx) error {
		var errs error
		bucket, err := s.MeasurementsBucket(ctx, tx, metricName)
		if err != nil {
			return err
		}
		if bucket == nil {
			return fmt.Errorf("measurementsBucket does not exist: %s", metricName)
		}

		for _, m := range metrics {
			data, err := jsoniter.Marshal(m)
			if err != nil {
				errs = multierror.Append(errs, err)
				continue
			}
			for _, key := range metricsKeys(m) {
				if err = bucket.Put(key, data); err != nil {
					errs = multierror.Append(errs, err)
				}
			}
		}
		return errs
	})
}

func (s *BoltstoreCore) cleanupMeasurements(ctx context.Context, metricName string) error {
	return s.Database().Update(func(tx *bbolt.Tx) error {
		var errs error
		bucket, err := s.MeasurementsBucket(ctx, tx, metricName)
		if err != nil {
			return err
		}
		c := bucket.Cursor()

		// Capture now to re-use for all the date math
		now := getCurrentTime()

		// Only look at data points older than 100 seconds, and we keep all the latest 10s intervals
		endCleanupDate := now.Add(-100 * time.Second)

		// Iterate through all points in the bucket, due to ordering we can't just scan by date
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			if ts, err := time.Parse(measurementsDateFormat, keyTimestamp(k)); err != nil {
				errs = multierror.Append(errs, err)
			} else {
				// Assume anything older than the cutoff is being deleted, unless the
				// normalized timestamp matches one of the rollup times
				keep := ts.Sub(endCleanupDate) > 0
				// For last 10 minutes keep the minute data
				keep = keep || ts.Truncate(time.Minute*1).Equal(ts) && now.Sub(ts) <= 10*time.Minute
				// for the last 6 hours keep the 5 minute data
				keep = keep || ts.Truncate(time.Minute*5).Equal(ts) && now.Sub(ts) <= 6*time.Hour
				// for the last 24 hours keep the hourly data
				keep = keep || ts.Truncate(time.Hour*1).Equal(ts) && now.Sub(ts) <= 24*time.Hour
				// for the last 31 days keep the daily data
				keep = keep || ts.Truncate(time.Hour*24).Equal(ts) && now.Sub(ts) <= 24*31*time.Hour

				if !keep {
					if err := c.Delete(); err != nil {
						errs = multierror.Append(errs, err)
					}
				}
			}
		}

		return errs
	})
}

func (s *BoltstoreCore) calculateRateMetric(first, last *record.Metric) *record.Metric {
	var lastValue float64
	var pass bool
	if lastValue, pass = stats.Value(last); !pass {
		return nil
	}

	var firstValue float64
	if firstValue, pass = stats.Value(first); !pass {
		return nil
	}

	rate, err := stats.Rate(
		stats.RateMetric{
			Timestamp:      first.Timestamp,
			StartTimestamp: first.StartTimestamp,
			Value:          firstValue,
		},
		stats.RateMetric{
			Timestamp:      last.Timestamp,
			StartTimestamp: last.StartTimestamp,
			Value:          lastValue,
		},
	)

	if err != nil {
		return nil
	}

	return generateRecord(last, rate.StartTimestamp, rate.Value, last.Attributes)
}

func generateRecord(source *record.Metric, startTime time.Time, value interface{}, attributes map[string]interface{}) *record.Metric {
	return &record.Metric{
		Name:           source.Name,
		Timestamp:      source.Timestamp.UTC(),
		StartTimestamp: startTime.UTC(),
		Value:          value,
		Unit:           "B/s",
		Type:           "Rate",
		Attributes:     attributes,
		Resource:       source.Resource,
	}
}

func keyTimestamp(k []byte) string {
	return strings.Split(string(k), "|")[2]
}

func archivePrefix(kind model.Kind, name string) []byte {
	return []byte(fmt.Sprintf("%s|%s|", kind, name))
}

func archiveKey(kind model.Kind, name string, version model.Version) []byte {
	// 6 digits max which allows for 999,999 ordered Versions beyond which the sort order is undefined
	return []byte(fmt.Sprintf("%s|%s|%06d", kind, name, version))
}

func archiveKeyFromResource(r model.Resource) []byte {
	if r == nil || r.GetKind() == model.KindUnknown {
		return make([]byte, 0)
	}
	return archiveKey(r.GetKind(), r.UniqueKey(), r.Version())
}

// metricsKey provides a key to store *record.Metric based on the agentID, processor and configuration of a single metric. It is
// assumed that all metrics come from a single agent using a single configuration.
func metricsKeys(m *record.Metric) [][]byte {
	normalizedTime := m.Timestamp.UTC().Truncate(10 * time.Second).Format(measurementsDateFormat)
	return [][]byte{
		[]byte(fmt.Sprintf("%s|%s|%s|%s|%s", model.KindAgent, stats.Agent(m), normalizedTime, stats.Configuration(m), stats.Processor(m))),
		[]byte(fmt.Sprintf("%s|%s|%s|%s|%s", model.KindConfiguration, stats.Configuration(m), normalizedTime, stats.Agent(m), stats.Processor(m))),
	}
}

func sanitizeKey(key string) string {
	return strings.ReplaceAll(key, "|", "")
}
