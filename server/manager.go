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

package server

import (
	"context"
	"errors"
	"time"

	"github.com/observiq/bindplane-op/agent"
	"github.com/observiq/bindplane-op/config"
	"github.com/observiq/bindplane-op/eventbus"
	"github.com/observiq/bindplane-op/eventbus/broadcast"
	"github.com/observiq/bindplane-op/model"
	"github.com/observiq/bindplane-op/server/protocol"
	"github.com/observiq/bindplane-op/store"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

// Manager manages agent connects and communications with them
//
//go:generate mockery --name Manager --filename mock_manager.go --structname MockManager
type Manager interface {
	// EnableProtocol adds the protocol to the manager
	EnableProtocol(protocol.Protocol)
	// Agent returns the agent with the specified agentID.
	Agent(ctx context.Context, agentID string) (*model.Agent, error)
	// UpsertAgent adds a new Agent to the Store or updates an existing one
	UpsertAgent(ctx context.Context, agentID string, updater store.AgentUpdater) (*model.Agent, error)
	// AgentUpdates returns the updates that should be applied to an agent based on the current bindplane configuration
	AgentUpdates(ctx context.Context, agent *model.Agent) (*protocol.AgentUpdates, error)
	// VerifySecretKey checks to see if the specified secretKey matches configured secretKey
	VerifySecretKey(ctx context.Context, secretKey string) (context.Context, bool)
	// ResourceStore provides access to the store to render configurations
	ResourceStore() model.ResourceStore
	// BindPlaneURL returns the URL of the BindPlane server
	BindPlaneURL() string
	// BindPlaneInsecureSkipVerify returns true if the BindPlane server should be contacted without verifying the server's certificate chain and host name
	BindPlaneInsecureSkipVerify() bool
	// RequestReport sends report configuration to the specified agent
	RequestReport(ctx context.Context, agentID string, configuration protocol.Report) error
	// AgentVersion returns information about a version of an agent
	AgentVersion(ctx context.Context, version string) (*model.AgentVersion, error)
	// Store provides access to the BindPlane Store
	Store() store.Store
	// SendAgentMessage sends a message to an agent
	SendAgentMessage(ctx context.Context, message Message) error
	// AgentMessages returns a source of messages for agents, broadcast from all nodes
	AgentMessages(ctx context.Context) eventbus.Source[Message]
	// Start starts the manager
	Start(ctx context.Context)
	// Shutdown should disconnect all agents and is called before shutdown of the server
	Shutdown(context.Context) error
}

var tracer = otel.Tracer("bindplane/manager")

const (
	// AgentCleanupTTL is the default agent cleanup time to live.
	AgentCleanupTTL = 15 * time.Minute
	// AgentCleanupInterval is the default agent cleanup interval.
	AgentCleanupInterval = time.Minute
)

// ----------------------------------------------------------------------

// DefaultManager is the default implementation of the Manager
type DefaultManager struct {
	Messages      broadcast.Broadcast[Message]
	CancelManager context.CancelCauseFunc
	ManagerCtx    context.Context
	config        *config.Config
	Storage       store.Store
	Versions      agent.Versions
	Logger        *zap.Logger
	Protocols     []protocol.Protocol
	SecretKey     string
}

var _ Manager = (*DefaultManager)(nil)

// NewManager returns a new implementation of the Manager interface
func NewManager(cfg *config.Config, store store.Store, versions agent.Versions, logger *zap.Logger) Manager {
	m := &DefaultManager{
		// agentHeartbeatTicker: time.NewTicker(AgentHeartbeatInterval),
		config:    cfg,
		Storage:   store,
		Versions:  versions,
		Logger:    logger,
		Protocols: []protocol.Protocol{},
		SecretKey: cfg.Auth.SecretKey,
	}

	return m
}

// EnableProtocol adds the protocol to the manager
func (m *DefaultManager) EnableProtocol(protocol protocol.Protocol) {
	m.Protocols = append(m.Protocols, protocol)
}

// AgentMessages returns a source of messages for agents, broadcast from all nodes
func (m *DefaultManager) AgentMessages(_ context.Context) eventbus.Source[Message] {
	return m.Messages.Consumer()
}

// Start starts the manager and concurrently running jobs
func (m *DefaultManager) Start(ctx context.Context) {
	m.ManagerCtx, m.CancelManager = context.WithCancelCause(ctx)

	m.Messages = broadcast.NewLocalBroadcast[Message](m.ManagerCtx, m.Logger)

	m.StartCleanupAgents()
}

// StartCleanupAgents starts a goroutine that will handle agent cleanup.
// It will stop once the ManagerCtx is closed
func (m *DefaultManager) StartCleanupAgents() {
	// Start a goroutine for listening for cleanups
	go func() {
		agentCleanupTicker := time.NewTicker(AgentCleanupInterval)
		defer agentCleanupTicker.Stop()
		for {
			select {
			case <-m.ManagerCtx.Done():
				return
			case <-agentCleanupTicker.C:
				m.handleAgentCleanup()
			}
		}
	}()
}

// SendAgentMessage sends a message to an agent
func (m *DefaultManager) SendAgentMessage(ctx context.Context, message Message) error {
	m.Messages.Producer().Send(ctx, message)
	return nil
}

// Store provides access to the BindPlane Store
func (m *DefaultManager) Store() store.Store {
	return m.Storage
}

// Shutdown should disconnect all agents and is called before shutdown of the server
func (m *DefaultManager) Shutdown(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "manager/Shutdown")
	defer span.End()

	m.CancelManager(errors.New("manager shutdown"))

	var errs error
	for _, p := range m.Protocols {
		if err := p.Shutdown(ctx); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	return errs
}

// Agent returns the agent with the specified agentID.
func (m *DefaultManager) Agent(ctx context.Context, agentID string) (*model.Agent, error) {
	return m.Storage.Agent(ctx, agentID)
}

// UpsertAgent adds a new Agent to the Store or updates an existing one
func (m *DefaultManager) UpsertAgent(ctx context.Context, agentID string, updater store.AgentUpdater) (*model.Agent, error) {
	return m.Storage.UpsertAgent(ctx, agentID, updater)
}

// AgentUpdates returns the updates that should be applied to an agent based on the current bindplane configuration
func (m *DefaultManager) AgentUpdates(ctx context.Context, agent *model.Agent) (*protocol.AgentUpdates, error) {
	newConfiguration, err := m.Storage.AgentConfiguration(ctx, agent)
	if err != nil {
		return nil, err
	}
	newLabels := agent.Labels.Custom()
	return &protocol.AgentUpdates{
		Labels:        &newLabels,
		Configuration: newConfiguration,
	}, nil
}

// BindPlaneURL returns the URL of the BindPlane server
func (m *DefaultManager) BindPlaneURL() string {
	return m.config.BindPlaneURL()
}

// BindPlaneInsecureSkipVerify returns true if the BindPlane server should be contacted without verifying the server's certificate chain and host name
func (m *DefaultManager) BindPlaneInsecureSkipVerify() bool {
	return m.config.BindPlaneInsecureSkipVerify()
}

// VerifySecretKey checks to see if the specified secretKey matches configured secretKey. If the BindPlane server does not
// have a configured secretKey, this returns true.
// This implementation doesn't use or modify the context, but it is included to match the interface.
func (m *DefaultManager) VerifySecretKey(ctx context.Context, secretKey string) (context.Context, bool) {
	return ctx, m.SecretKey == "" || m.SecretKey == secretKey
}

// ResourceStore provides access to the store to render configurations
func (m *DefaultManager) ResourceStore() model.ResourceStore {
	return m.Storage
}

// RequestReport sends report configuration to the specified agent
func (m *DefaultManager) RequestReport(ctx context.Context, agentID string, configuration protocol.Report) error {
	ctx, span := tracer.Start(ctx, "manager/RequestReport")
	defer span.End()

	for _, p := range m.Protocols {
		err := p.RequestReport(ctx, agentID, configuration)
		if err != nil {
			return err
		}
	}

	return nil
}

// AgentVersion returns information about a version of an agent
func (m *DefaultManager) AgentVersion(ctx context.Context, version string) (*model.AgentVersion, error) {
	_, span := tracer.Start(ctx, "manager/AgentVersion")
	defer span.End()

	span.SetAttributes(attribute.String("version", version))

	return m.Versions.Version(ctx, version)
}

// ----------------------------------------------------------------------

// handleAgentCleanup removes disconnected container agents from the store.
func (m *DefaultManager) handleAgentCleanup() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx, span := tracer.Start(ctx, "manager/handleAgentCleanup")
	defer span.End()

	cutoff := time.Now().Add(-AgentCleanupTTL)
	m.Logger.Sugar().Infof("cleaning up container agents disconnected since %s", cutoff.Format(time.RFC3339))

	// TODO: in a cluster, move this to a job
	err := m.Storage.CleanupDisconnectedAgents(ctx, cutoff)
	if err != nil {
		m.Logger.Error("error cleaning up disconnected agents", zap.Error(err))
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

func (m *DefaultManager) handleAgentHeartbeat(ctx context.Context) {
	ctx, span := tracer.Start(ctx, "manager/handleAgentHeartbeat")
	defer span.End()

	for _, p := range m.Protocols {
		ids, err := p.ConnectedAgentIDs(ctx)
		if err != nil {
			m.Logger.Error("unable to get connected agents", zap.String("protocol", p.Name()))
			continue
		}
		for _, id := range ids {
			err = p.SendHeartbeat(id)
			if err != nil {
				m.Logger.Error("unable to get send agent heartbeat", zap.String("protocol", p.Name()), zap.String("agentID", id))
				continue
			}
		}
	}
}
