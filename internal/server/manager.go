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

package server

import (
	"context"
	"math"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"

	"github.com/observiq/bindplane-op/common"
	"github.com/observiq/bindplane-op/internal/agent"
	"github.com/observiq/bindplane-op/internal/eventbus"
	"github.com/observiq/bindplane-op/internal/server/protocol"
	"github.com/observiq/bindplane-op/internal/store"
	"github.com/observiq/bindplane-op/model"
)

var tracer = otel.Tracer("bindplane/manager")

const (
	// AgentCleanupInterval is the default agent cleanup interval.
	AgentCleanupInterval = time.Minute
	// AgentCleanupTTL is the default agent cleanup time to live.
	AgentCleanupTTL = 15 * time.Minute
	// AgentHeartbeatInterval is the default interval for the heartbeat sent to the agent to keep the websocket live.
	AgentHeartbeatInterval = 30 * time.Second
)

// Manager manages agent connects and communications with them
//
//go:generate mockery --name Manager --filename mock_manager.go --structname MockManager
type Manager interface {
	// Start starts the manager and allows it to begin processing configuration changes
	Start(ctx context.Context)
	// EnableProtocol adds the protocol to the manager
	EnableProtocol(protocol.Protocol)
	// Agent returns the agent with the specified agentID.
	Agent(ctx context.Context, agentID string) (*model.Agent, error)
	// UpsertAgent adds a new Agent to the Store or updates an existing one
	UpsertAgent(ctx context.Context, agentID string, updater store.AgentUpdater) (*model.Agent, error)
	// AgentUpdates returns the updates that should be applied to an agent based on the current bindplane configuration
	AgentUpdates(ctx context.Context, agent *model.Agent) (*protocol.AgentUpdates, error)
	// VerifySecretKey checks to see if the specified secretKey matches configured secretKey
	VerifySecretKey(ctx context.Context, secretKey string) bool
	// ResourceStore provides access to the store to render configurations
	ResourceStore() model.ResourceStore
	// BindPlaneConfiguration provides access to the config to render configurations
	BindPlaneConfiguration() model.BindPlaneConfiguration
	// RequestReport sends report configuration to the specified agent
	RequestReport(ctx context.Context, agentID string, configuration protocol.Report) error
	// AgentVersion returns information about a version of an agent
	AgentVersion(ctx context.Context, version string) (*model.AgentVersion, error)
}

// ----------------------------------------------------------------------

type manager struct {
	// agentHeartbeatTicker *time.Ticker
	agentCleanupTicker *time.Ticker
	config             *common.Server
	store              store.Store
	versions           agent.Versions
	logger             *zap.Logger
	protocols          []protocol.Protocol
	secretKey          string
}

var _ Manager = (*manager)(nil)

// NewManager returns a new implementation of the Manager interface
func NewManager(config *common.Server, store store.Store, versions agent.Versions, logger *zap.Logger) (Manager, error) {
	return &manager{
		// agentHeartbeatTicker: time.NewTicker(AgentHeartbeatInterval),
		agentCleanupTicker: time.NewTicker(AgentCleanupInterval),
		config:             config,
		store:              store,
		versions:           versions,
		logger:             logger,
		protocols:          []protocol.Protocol{},
		secretKey:          config.SecretKey,
	}, nil
}

func (m *manager) EnableProtocol(protocol protocol.Protocol) {
	m.protocols = append(m.protocols, protocol)
}

// Start TODO(doc)
func (m *manager) Start(ctx context.Context) {
	updatesChannel, unsubscribe := eventbus.Subscribe(m.store.Updates(), eventbus.WithChannel(make(chan *store.Updates, 10_000)))
	defer unsubscribe()

	for {
		select {
		case <-ctx.Done():
			m.agentCleanupTicker.Stop()
			// m.agentHeartbeatTicker.Stop()
			return

		case updates := <-updatesChannel:
			m.logger.Info("Received configuration updates",
				zap.Int("size", updates.Size()),
				zap.Int("Agents", len(updates.Agents)),
				zap.Int("Configurations", len(updates.Configurations)),
			)
			m.handleUpdates(ctx, updates)

		case <-m.agentCleanupTicker.C:
			m.handleAgentCleanup()

			// TODO: determine if this needs to be replaced and if so, replace it
			// case <-m.agentHeartbeatTicker.C:
			// 	m.handleAgentHeartbeat()
		}
	}
}

// helper for bookkeeping during updates
type pendingAgentUpdate struct {
	agent   *model.Agent
	updates *protocol.AgentUpdates
}

type pendingAgentUpdates map[string]pendingAgentUpdate

func (p pendingAgentUpdates) agent(agent *model.Agent) pendingAgentUpdate {
	u, ok := p[agent.ID]
	if ok {
		return u
	}
	u = pendingAgentUpdate{
		agent:   agent,
		updates: &protocol.AgentUpdates{},
	}
	p[agent.ID] = u
	return u
}

func (p pendingAgentUpdates) apply(ctx context.Context, m *manager) {
	ctx, span := tracer.Start(ctx, "manager/apply")
	defer span.End()

	if len(p) == 0 {
		return
	}

	// Number of workers is a quarter of the total or 10
	// This insures for small updates we don't spin up 10 workers for 1 or 2 updates
	numWorkers := int(math.Min(float64(len(p)/4)+1, 10))

	m.logger.Info("Creating workers", zap.Int("numWorkers", numWorkers), zap.Int("pending updates", len(p)))

	startTime := time.Now()
	// Spun up worker group
	var wg sync.WaitGroup
	wg.Add(numWorkers)
	updateChan := make(chan *pendingAgentUpdate, numWorkers)
	for i := 0; i < numWorkers; i++ {
		go updateWorker(ctx, &wg, m, updateChan)
	}

	for _, pending := range p {
		if !pending.updates.Empty() {
			pendingCpy := pending
			updateChan <- &pendingCpy
		}
	}

	close(updateChan)
	wg.Wait()

	execTime := time.Since(startTime)
	m.logger.Info("Update Time", zap.String("dur", execTime.String()))
}

func updateWorker(ctx context.Context, wg *sync.WaitGroup, m *manager, updateChan <-chan *pendingAgentUpdate) {
	ctx, span := tracer.Start(ctx, "manager/updateWorker")
	defer span.End()

	defer wg.Done()

	for {
		pending, ok := <-updateChan
		if !ok {
			return
		}
		m.updateAgent(ctx, pending.agent, pending.updates)
	}
}

func (m *manager) handleUpdates(ctx context.Context, updates *store.Updates) {
	if updates.Empty() {
		return
	}
	ctx, span := tracer.Start(ctx, "manager/handleUpdates")
	defer span.End()

	pending := pendingAgentUpdates{}

	for _, change := range updates.Agents {
		// on delete, disconnect
		if change.Type == store.EventTypeRemove {
			m.disconnect(change.Item.ID)
			continue
		}
		agent := change.Item
		// otherwise, we only care able label changes
		if change.Type != store.EventTypeLabel {
			// unless there is a pending version update
			if agent.Upgrade != nil && agent.Upgrade.Status == model.UpgradePending {
				pending.agent(agent).updates.Version = agent.Upgrade.Version
			}
			continue
		}

		// only consider connected agents
		if !m.connected(agent.ID) {
			continue
		}

		// this is only triggered for label changes right now, so we can just update that field
		m.logger.Info("updating labels for agent", zap.String("agentID", agent.ID), zap.String("labels", agent.Labels.String()))
		labels := agent.Labels.Custom()
		agentUpdates := pending.agent(agent).updates
		agentUpdates.Labels = &labels
		if agent.Upgrade != nil && agent.Upgrade.Status == model.UpgradePending {
			agentUpdates.Version = agent.Upgrade.Version
		}

		// if the labels changed, there may be new configuration
		if configuration, err := m.store.AgentConfiguration(ctx, agent.ID); err != nil {
			m.logger.Error("unable to find new agent configuration", zap.String("agentID", agent.ID), zap.String("labels", agent.Labels.String()))
		} else {
			if configuration != nil {
				m.logger.Info("updating configuration for agent with new labels", zap.String("agentID", agent.ID), zap.String("labels", agent.Labels.String()), zap.String("configuration.name", configuration.Name()))
				agentUpdates.Configuration = configuration
			}
		}
	}

	for _, event := range updates.Configurations {
		configuration := event.Item
		agentIDs, err := m.store.AgentsIDsMatchingConfiguration(ctx, configuration)
		if err != nil {
			m.logger.Error("unable to apply configuration to agents", zap.String("configuration.name", configuration.Name()), zap.Error(err))
			continue
		}

		for _, agentID := range agentIDs {
			// only consider connected agents
			if !m.connected(agentID) {
				continue
			}

			agent, err := m.store.Agent(ctx, agentID)
			if err != nil {
				m.logger.Error("unable to apply configuration to agent", zap.String("agentID", agentID), zap.String("configuration.name", configuration.Name()), zap.Error(err))
				continue
			}

			// TODO(andy): support multiple matches with precedence
			if event.Type == store.EventTypeRemove {
				m.logger.Info("deleting configuration for agent", zap.String("agentID", agent.ID))

				// TODO(andy): we need a default configuration
				// https://github.com/observIQ/bindplane/issues/279
				// agentUpdates.Configuration = otel.EmptyConfig()
			} else {
				m.logger.Info("updating configuration for agent", zap.String("agentID", agent.ID))
				pending.agent(agent).updates.Configuration = configuration
			}
		}

	}

	pending.apply(ctx, m)
}

func (m *manager) Agent(ctx context.Context, agentID string) (*model.Agent, error) {
	return m.store.Agent(ctx, agentID)
}

func (m *manager) UpsertAgent(ctx context.Context, agentID string, updater store.AgentUpdater) (*model.Agent, error) {
	return m.store.UpsertAgent(ctx, agentID, updater)
}

// AgentUpdates returns the updates that should be applied to an agent based on the current bindplane configuration
func (m *manager) AgentUpdates(ctx context.Context, agent *model.Agent) (*protocol.AgentUpdates, error) {
	newConfiguration, err := m.store.AgentConfiguration(ctx, agent.ID)
	if err != nil {
		return nil, err
	}
	newLabels := agent.Labels.Custom()
	return &protocol.AgentUpdates{
		Labels:        &newLabels,
		Configuration: newConfiguration,
	}, nil
}

// VerifySecretKey checks to see if the specified secretKey matches configured secretKey. If the BindPlane server does not
// have a configured secretKey, this returns true.
func (m *manager) VerifySecretKey(_ context.Context, secretKey string) bool {
	return m.secretKey == "" || m.secretKey == secretKey
}

// ResourceStore provides access to the store to render configurations
func (m *manager) ResourceStore() model.ResourceStore {
	return m.store
}

// BindPlaneConfiguration provides access to the config to render configurations
func (m *manager) BindPlaneConfiguration() model.BindPlaneConfiguration {
	return m.config
}

// RequestReport sends report configuration to the specified agent
func (m *manager) RequestReport(ctx context.Context, agentID string, configuration protocol.Report) error {
	ctx, span := tracer.Start(ctx, "manager/RequestReport")
	defer span.End()

	for _, p := range m.protocols {
		err := p.RequestReport(ctx, agentID, configuration)
		if err != nil {
			return err
		}
	}

	return nil
}

// AgentVersion returns information about a version of an agent
func (m *manager) AgentVersion(ctx context.Context, version string) (*model.AgentVersion, error) {
	_, span := tracer.Start(ctx, "manager/AgentVersion")
	defer span.End()

	span.SetAttributes(attribute.String("version", version))

	return m.versions.Version(ctx, version)
}

// ----------------------------------------------------------------------

// handleAgentCleanup removes disconnected container agents from the store.
func (m *manager) handleAgentCleanup() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx, span := tracer.Start(ctx, "manager/handleAgentCleanup")
	defer span.End()

	cutoff := time.Now().Add(-AgentCleanupTTL)
	m.logger.Sugar().Infof("cleaning up container agents disconnected since %s", cutoff.Format(time.RFC3339))

	// TODO: in a cluster, move this to a job
	err := m.store.CleanupDisconnectedAgents(ctx, cutoff)
	if err != nil {
		m.logger.Error("error cleaning up disconnected agents", zap.Error(err))
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

func (m *manager) handleAgentHeartbeat(ctx context.Context) {
	ctx, span := tracer.Start(ctx, "manager/handleAgentHeartbeat")
	defer span.End()

	for _, p := range m.protocols {
		ids, err := p.ConnectedAgentIDs(ctx)
		if err != nil {
			m.logger.Error("unable to get connected agents", zap.String("protocol", p.Name()))
			continue
		}
		for _, id := range ids {
			err = p.SendHeartbeat(id)
			if err != nil {
				m.logger.Error("unable to get send agent heartbeat", zap.String("protocol", p.Name()), zap.String("agentID", id))
				continue
			}
		}
	}
}

// ----------------------------------------------------------------------
// Protocol usage

func (m *manager) disconnect(agentID string) bool {
	for _, p := range m.protocols {
		if p.Disconnect(agentID) {
			return true
		}
	}
	return false
}

func (m *manager) connected(agentID string) bool {
	for _, p := range m.protocols {
		if p.Connected(agentID) {
			return true
		}
	}
	return false
}

// connectedAgentIDs returns the list of agents connected using any protocol
func (m *manager) connectedAgentIDs(ctx context.Context) []string {
	ids := []string{}
	for _, p := range m.protocols {
		list, err := p.ConnectedAgentIDs(ctx)
		if err != nil {
			m.logger.Error("unable to get connected agents", zap.String("protocol", p.Name()))
			continue
		}
		ids = append(ids, list...)
	}
	return ids
}

func (m *manager) updateAgent(ctx context.Context, agent *model.Agent, updates *protocol.AgentUpdates) {
	for _, p := range m.protocols {
		err := p.UpdateAgent(ctx, agent, updates)
		if err != nil {
			m.logger.Error("unable to update agent", zap.String("agentID", agent.ID))
		}
	}
}
