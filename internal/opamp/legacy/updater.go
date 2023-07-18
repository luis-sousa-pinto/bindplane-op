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

package legacy

import (
	"context"
	"math"
	"sync"

	"github.com/mitchellh/mapstructure"
	"github.com/observiq/bindplane-op/eventbus"
	"github.com/observiq/bindplane-op/model"
	bpserver "github.com/observiq/bindplane-op/server"
	"github.com/observiq/bindplane-op/server/protocol"
	"github.com/observiq/bindplane-op/store"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type updater struct {
	protocol protocol.Protocol
	manager  bpserver.Manager
	stop     chan struct{}
	stopOnce *sync.Once
	logger   *zap.Logger
}

// newUpdater returns a new updater with the specified components
func newUpdater(protocol protocol.Protocol, manager bpserver.Manager, logger *zap.Logger) bpserver.Updater {
	return &updater{
		protocol: protocol,
		manager:  manager,
		stop:     make(chan struct{}),
		stopOnce: &sync.Once{},
		logger:   logger,
	}
}

func (u *updater) store() store.Store {
	return u.manager.Store()
}

// Start subscribes to store updates and agent messages.
// It will handle updates and messages until the context is canceled or Stop is called.
func (u *updater) Start(ctx context.Context) {
	u.logger.Info("updater subscribing to updates")
	updatesChannel, unsubscribe := eventbus.Subscribe(ctx, u.store().Updates(ctx), eventbus.WithChannel(make(chan store.BasicEventUpdates, 10_000)))
	defer unsubscribe()

	messagesChannel, unsubscribeMessages := eventbus.Subscribe(ctx, u.manager.AgentMessages(ctx))
	defer unsubscribeMessages()

	for {
		select {
		case <-ctx.Done():
			u.logger.Info("Context canceled", zap.Error(ctx.Err()))
			// m.agentCleanupTicker.Stop()
			// m.agentHeartbeatTicker.Stop()
			return
		case <-u.stop:
			u.logger.Info("Stop requested")
			return
		case updates := <-updatesChannel:
			u.logger.Info("Received configuration updates",
				zap.Int("size", updates.Size()),
				zap.Int("Agents", len(updates.Agents())),
				zap.Int("Configurations", len(updates.Configurations())),
			)
			u.handleUpdates(ctx, updates)

		case message := <-messagesChannel:
			u.logger.Info("Received agent message",
				zap.String("agentID", message.AgentID()),
				zap.String("type", message.Type()))
			u.handleMessage(ctx, message)

			// TODO: determine if these need to be replaced and if so, replace them
			// case <-m.agentCleanupTicker.C:
			// 	m.handleAgentCleanup()

			// case <-m.agentHeartbeatTicker.C:
			// 	m.handleAgentHeartbeat()
		}
	}
}

// Stop calls the stop function.
// Concurrent calls to Stop will only call the stop function once.
func (u *updater) Stop(_ context.Context) {
	u.stopOnce.Do(func() {
		close(u.stop)
	})
}

// TODO: maybe pass in pendingAgentUpdates so the contents can be tested after
func (u *updater) handleUpdates(ctx context.Context, updates store.BasicEventUpdates) {
	if updates.Empty() {
		return
	}
	ctx, span := tracer.Start(ctx, "updater/handleUpdates")
	defer span.End()

	pending := pendingAgentUpdates{}

	for _, change := range updates.Agents() {
		// on delete, disconnect
		if change.Type == store.EventTypeRemove {
			u.protocol.Disconnect(change.Item.ID)
			continue
		}
		agent := change.Item
		// otherwise, we only care about rollouts
		if change.Type != store.EventTypeRollout {
			// unless there is a pending version update
			if agent.Upgrade != nil && agent.Upgrade.Status == model.UpgradePending {
				pending.agent(agent).updates.Version = agent.Upgrade.Version
			}
			continue
		}

		// only consider connected agents
		if !u.protocol.Connected(agent.ID) {
			continue
		}

		u.logger.Info("rollout to agent", zap.String("agentID", agent.ID), zap.String("labels", agent.Labels.String()))
		labels := agent.Labels.Custom()
		agentUpdates := pending.agent(agent).updates
		agentUpdates.Labels = &labels
		if agent.Upgrade != nil && agent.Upgrade.Status == model.UpgradePending {
			agentUpdates.Version = agent.Upgrade.Version
		}

		// during a rollout, there should be a new configuration
		configuration, err := u.store().AgentConfiguration(ctx, agent)
		switch {
		case err != nil:
			u.logger.Error("unable to find new agent configuration", zap.String("agentID", agent.ID),
				zap.String("labels", agent.Labels.String()))
		case configuration != nil:
			u.logger.Info("updating configuration for agent", zap.String("agentID", agent.ID),
				zap.String("labels", agent.Labels.String()), zap.String("configuration.name", configuration.Name()))
			agentUpdates.Configuration = configuration
		}
	}

	pending.apply(ctx, u.updateAgent)
}

func (u *updater) updateAgent(ctx context.Context, agent *model.Agent, updates *protocol.AgentUpdates) {
	// if not connected to this agent, ignore the request
	if !u.protocol.Connected(agent.ID) {
		return
	}

	ctx, span := tracer.Start(ctx, "updater/updateAgent", trace.WithAttributes(attribute.String("agentID", agent.ID)))
	defer span.End()

	err := u.protocol.UpdateAgent(ctx, agent, updates)
	if err != nil {
		u.logger.Error("unable to update agent", zap.String("agentID", agent.ID), zap.Error(err))
	}
}

func (u *updater) handleMessage(ctx context.Context, message bpserver.Message) {
	// if not connected to this agent, ignore the message
	if !u.protocol.Connected(message.AgentID()) {
		return
	}

	ctx, span := tracer.Start(ctx, "updater/handleMessage", trace.WithAttributes(
		attribute.String("agentID", message.AgentID()), attribute.String("type", message.Type())))
	defer span.End()

	switch message.Type() {
	case bpserver.AgentMessageTypeSnapshot:
		body, err := parseAgentMessageBody[bpserver.SnapshotBody](message)
		if err != nil {
			u.logger.Error("unable to parse snapshot message body", zap.String("agentID", message.AgentID()),
				zap.Error(err))
			return
		}
		err = u.protocol.RequestReport(ctx, message.AgentID(), body.Configuration)
		if err != nil {
			u.logger.Error("unable to request report from agent", zap.String("agentID", message.AgentID()),
				zap.Error(err))
			return
		}
	}
}

// ParseAgentMessageBody parses the body of an agent message into the specified type
func parseAgentMessageBody[T any](agentMessage bpserver.Message) (T, error) {
	var body T
	err := mapstructure.Decode(agentMessage.Body(), &body)
	return body, err
}

// helper for bookkeeping during updates
type pendingAgentUpdate struct {
	agent   *model.Agent
	updates *protocol.AgentUpdates
}

type agentUpdater func(context.Context, *model.Agent, *protocol.AgentUpdates)

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

func (p pendingAgentUpdates) apply(ctx context.Context, updater agentUpdater) {
	ctx, span := tracer.Start(ctx, "updater/apply")
	defer span.End()

	if len(p) == 0 {
		return
	}

	// Number of workers is a quarter of the total or 10
	// This insures for small updates we don't spin up 10 workers for 1 or 2 updates
	numWorkers := int(math.Min(float64(len(p)/4)+1, 10))

	span.SetAttributes(
		attribute.Int("numWorkers", numWorkers),
		attribute.Int("pendingUpdates", len(p)),
	)

	// Spun up worker group
	var wg sync.WaitGroup
	wg.Add(numWorkers)
	updateChan := make(chan *pendingAgentUpdate, numWorkers)
	for i := 0; i < numWorkers; i++ {
		go updateWorker(ctx, &wg, updater, updateChan)
	}

	for _, pending := range p {
		if !pending.updates.Empty() {
			pendingCpy := pending
			updateChan <- &pendingCpy
		}
	}

	close(updateChan)
	wg.Wait()
}

func updateWorker(ctx context.Context, wg *sync.WaitGroup, updater agentUpdater, updateChan <-chan *pendingAgentUpdate) {
	ctx, span := tracer.Start(ctx, "updater/updateWorker")
	defer span.End()

	defer wg.Done()

	for {
		pending, ok := <-updateChan
		if !ok {
			return
		}
		updater(ctx, pending.agent, pending.updates)
	}
}
