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

// Package opamp implements the OpenTelemetry OpAMP protocol.
package opamp

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	bpopamp "github.com/observiq/bindplane-op/opamp"
	bpserver "github.com/observiq/bindplane-op/server"
	"github.com/open-telemetry/opamp-go/protobufs"
	opampSvr "github.com/open-telemetry/opamp-go/server"
	opamp "github.com/open-telemetry/opamp-go/server/types"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/observiq/bindplane-op/internal/opamp/connections"
	"github.com/observiq/bindplane-op/internal/opamp/legacy"
	"github.com/observiq/bindplane-op/model"
	"github.com/observiq/bindplane-op/model/observiq"
	exposedserver "github.com/observiq/bindplane-op/server"
	"github.com/observiq/bindplane-op/server/protocol"
)

var tracer = otel.Tracer("bindplane/opamp")

// ProtocolName is "opamp"
const ProtocolName = "opamp"

const (
	headerAuthorization = "Authorization"
	headerUserAgent     = "User-Agent"
	headerOpAMPVersion  = "OpAMP-Version"
	headerAgentID       = "Agent-ID"
	headerAgentVersion  = "Agent-Version"
	headerAgentHostname = "Agent-Hostname"
)

// AddRoutes adds the routes used by opamp, currently /v1/opamp
func AddRoutes(router gin.IRouter, bindplane exposedserver.BindPlane) error {
	// legacy handler is used for v0.2.0 agents
	legacyHandler, err := legacy.BuildLegacyHandler(bindplane)
	if err != nil {
		return err
	}

	// current handler is used for > v0.7.1 agents
	currentHandler, err := buildCurrentHandler(bindplane)

	multiVersionHandler := func(res http.ResponseWriter, req *http.Request) {
		opampVersion := req.Header.Get(headerOpAMPVersion)
		switch opampVersion {
		case "v0.2.0":
			legacyHandler(res, req)
		default:
			currentHandler(res, req)
		}
	}

	router.Any("/opamp", gin.WrapF(http.HandlerFunc(multiVersionHandler)))

	return nil
}

func buildCurrentHandler(bindplane exposedserver.BindPlane) (func(res http.ResponseWriter, req *http.Request), error) {
	callbacks := newServer(bindplane.Manager(), bindplane.Logger())
	server := opampSvr.New(bindplane.Logger().Sugar())
	settings := opampSvr.Settings{
		Callbacks: callbacks,
	}

	handler, _, err := server.Attach(settings)
	if err != nil {
		return nil, fmt.Errorf("error attempting to attach the OpAMP v0.7.1 server: %w", err)
	}

	bindplane.Manager().EnableProtocol(callbacks)

	return handler, nil
}

const (
	capabilities = protobufs.ServerCapabilities_ServerCapabilities_AcceptsStatus |
		protobufs.ServerCapabilities_ServerCapabilities_AcceptsEffectiveConfig |
		protobufs.ServerCapabilities_ServerCapabilities_OffersRemoteConfig
)

type opampServer struct {
	manager     bpserver.Manager
	connections bpopamp.Connections[*bpopamp.AgentConnectionState]
	logger      *zap.Logger
	updater     bpserver.Updater
}

var _ protocol.Protocol = (*opampServer)(nil)
var _ opamp.Callbacks = (*opampServer)(nil)

func newServer(manager bpserver.Manager, logger *zap.Logger) *opampServer {
	s := &opampServer{
		manager:     manager,
		connections: connections.NewConnections(),
		logger:      logger,
	}
	s.updater = newUpdater(
		s,
		s.manager,
		s.logger,
	)
	return s
}

// ----------------------------------------------------------------------
// The following callbacks will never be called concurrently for the same
// connection. They may be called concurrently for different connections.

// OnConnecting is called when there is a new incoming connection.
// The handler can examine the request and either accept or reject the connection.
// To accept:
//
//	Return ConnectionResponse with Accept=true.
//	HTTPStatusCode and HTTPResponseHeader are ignored.
//
// To reject:
//
//	Return ConnectionResponse with Accept=false. HTTPStatusCode MUST be set to
//	non-zero value to indicate the rejection reason (typically 401, 429 or 503).
//	HTTPResponseHeader may be optionally set (e.g. "Retry-After: 30").
func (s *opampServer) OnConnecting(request *http.Request) opamp.ConnectionResponse {
	ctx, span := tracer.Start(request.Context(), "opamp/connecting")
	defer span.End()

	s.logger.Info("OnConnecting", zap.Any("headers", request.Header), zap.String("RemoteAddr", request.RemoteAddr))

	// check for compatibility
	headers := parseAgentHeaders(request)

	ctx, accept := s.manager.VerifySecretKey(ctx, headers.secretKey)
	if !accept {
		span.SetStatus(codes.Error, http.StatusText(http.StatusUnauthorized))
		return opamp.ConnectionResponse{
			Accept:         false,
			HTTPStatusCode: http.StatusUnauthorized,
		}
	}

	s.connections.OnConnecting(ctx, headers.id)

	go s.updater.Start(context.Background())

	return opamp.ConnectionResponse{
		Accept:              true,
		HTTPStatusCode:      http.StatusOK,
		ConnectionCallbacks: s,
	}
}

type agentHeaders struct {
	opampVersion string
	id           string
	version      string
	hostname     string
	secretKey    string
}

func parseAgentHeaders(request *http.Request) *agentHeaders {
	authHeader := request.Header.Get(headerAuthorization)
	secretKey := strings.Replace(authHeader, "Secret-Key ", "", 1)
	if secretKey == authHeader {
		// check for missing Secret-Key identifier
		secretKey = ""
	}
	return &agentHeaders{
		opampVersion: request.Header.Get(headerOpAMPVersion),
		id:           request.Header.Get(headerAgentID),
		version:      request.Header.Get(headerAgentVersion),
		hostname:     request.Header.Get(headerAgentHostname),
		secretKey:    secretKey,
	}
}

// OnConnected is called when the WebSocket connection is successfully established after OnConnecting() returns and the
// HTTP connection is upgraded to WebSocket.
//
// opamp.Connection doesn't have much information that we can use here
func (s *opampServer) OnConnected(_ opamp.Connection) {
	_, span := tracer.Start(context.TODO(), "opamp/connected")
	defer span.End()
}

// OnMessage is called when a message is received from the connection. Can happen
// only after OnConnected().
func (s *opampServer) OnMessage(conn opamp.Connection, message *protobufs.AgentToServer) *protobufs.ServerToAgent {
	ctx, span := tracer.Start(context.Background(), "opamp/message")
	defer span.End()

	agentID := message.InstanceUid
	response := &protobufs.ServerToAgent{
		InstanceUid:  agentID,
		Capabilities: uint64(capabilities),
	}

	if _, err := s.connections.OnMessage(agentID, conn); err != nil {
		s.logger.Error("failed to verify the agent configuration", zap.Error(err))
		response.ErrorResponse = &protobufs.ServerErrorResponse{
			Type:         protobufs.ServerErrorResponseType_ServerErrorResponseType_Unknown,
			ErrorMessage: err.Error(),
		}
		return response
	}
	hasConfiguration := message.GetEffectiveConfig().GetConfigMap() != nil

	span.SetAttributes(
		attribute.String("bindplane.agent.id", agentID),
		attribute.String("bindplane.component", "opamp"),
		attribute.Bool("bindplane.opamp.hasConfiguration", hasConfiguration),
	)

	s.logger.Info("OpAMP agent message", zap.String("agentID", agentID), zap.Strings("submessages", bpopamp.MessageComponents(message)))

	// verify the configuration and modify the response message
	err := s.verifyAgentConfig(ctx, conn, agentID, message, response)
	if err != nil {
		s.logger.Error("error verifying the agent configuration", zap.Error(err))
		// send an error response
		// TODO(andy): Ok to report the exact error?
		response.ErrorResponse = &protobufs.ServerErrorResponse{
			Type:         protobufs.ServerErrorResponseType_ServerErrorResponseType_Unknown,
			ErrorMessage: err.Error(),
		}
	}
	s.logger.Info("sending response to the agent", zap.Any("agentID", agentID), zap.Any("response", response))

	return response
}

// OnConnectionClose is called when the WebSocket connection is closed.
// Typically, preceded by OnDisconnect() unless the client misbehaves or the
// connection is lost.
func (s *opampServer) OnConnectionClose(conn opamp.Connection) {
	ctx, span := tracer.Start(context.Background(), "opamp/OnConnectionClose")
	defer span.End()

	state, _ := s.connections.OnConnectionClose(conn)
	if state == nil {
		return
	}

	agentID := state.AgentID
	s.logger.Info("OpAMP agent disconnected", zap.String("AgentID", agentID))

	if agentID == "" {
		return
	}

	_, err := s.manager.UpsertAgent(ctx, agentID, func(agent *model.Agent) {
		agent.Disconnect()
	})
	if err != nil {
		s.logger.Error("error trying to save disconnected state of agent", zap.String("agentID", agentID), zap.Error(err))
		return
	}
}

// ----------------------------------------------------------------------
// Protocol implementation

func (s *opampServer) Name() string {
	return ProtocolName
}

// ConnectedAgentIDs should return a slice of the currently connected agent IDs
func (s *opampServer) ConnectedAgentIDs(ctx context.Context) ([]string, error) {
	ctx, span := tracer.Start(ctx, "opamp/ConnectedAgentIDs")
	defer span.End()

	return s.connections.ConnectedAgentIDs(ctx), nil
}

func (s *opampServer) Disconnect(agentID string) bool {
	state := s.connections.StateForAgentID(agentID)
	if state == nil {
		return false
	}
	if conn := state.Conn; conn != nil {
		s.connections.OnConnectionClose(conn)
		return true
	}
	return false
}

// Connected returns true if the specified agent ID is connected
func (s *opampServer) Connected(agentID string) bool {
	return s.connections.Connected(agentID)
}

// UpdateAgent should send a message to the specified agent to update the configuration to match the specified
// configuration.
//
// This function is called when the agent configuration is updated in the Store and we want to PUSH the changes to a
// connected agent.
func (s *opampServer) UpdateAgent(ctx context.Context, agent *model.Agent, updates *protocol.AgentUpdates) error {
	state := s.connections.StateForAgentID(agent.ID)
	if state == nil {
		return fmt.Errorf("no connection state for agentID %s", agent.ID)
	}

	conn := state.Conn
	if conn == nil {
		// agent not connected, nothing to do
		return nil
	}

	ctx, span := tracer.Start(ctx, "opamp/UpdateAgent", trace.WithAttributes(
		attribute.String("bindplane.agent.id", agent.ID),
	))
	defer span.End()

	agentConfiguration, err := observiq.DecodeAgentConfiguration(agent.Configuration)
	if err != nil {
		// start with a blank configuration if the current isn't available
		agentConfiguration = &observiq.AgentConfiguration{}
	}

	newConfiguration, err := s.updatedConfiguration(ctx, agent, agentConfiguration, updates)
	if err != nil {
		return fmt.Errorf("unable to get the new configuration for agent [%s]: %w", agent.ID, err)
	}

	serverToAgent := &protobufs.ServerToAgent{
		InstanceUid:  agent.ID,
		Capabilities: uint64(capabilities),
		Flags:        uint64(protobufs.ServerToAgentFlags_ServerToAgentFlags_ReportFullState),
	}

	if newConfiguration.Empty() {
		s.logger.Info("agent already has the correct configuration")
		s.updateAgentCurrentConfiguration(ctx, agent, updates.Configuration)
	} else {
		agentRawConfiguration := agentConfiguration.Raw()
		newRawConfiguration := newConfiguration.Raw()

		serverToAgent.RemoteConfig = agentRemoteConfig(&newRawConfiguration, &agentRawConfiguration)

		// change the agent status to Configuring, but ignore any failure as this status is considered nice to have and not required to update the agent
		_, _ = s.manager.UpsertAgent(ctx, agent.ID, func(current *model.Agent) { current.Status = model.Configuring })
	}

	if updates.Version != "" {
		s.logger.Info("sending agent update to version", zap.String("version", updates.Version))
		downloadableFile, err := s.getDownloadableFile(ctx, agent, updates.Version)
		if err != nil || downloadableFile == nil {
			s.logger.Error("unable to send agent update", zap.Error(err))
			agent, _ = s.manager.UpsertAgent(ctx, agent.ID, func(current *model.Agent) {
				current.UpgradeComplete(updates.Version, err.Error())
			})
		} else {
			allPackagesHash := []byte(updates.Version)
			serverToAgent.PackagesAvailable = &protobufs.PackagesAvailable{
				AllPackagesHash: allPackagesHash,
				Packages: map[string]*protobufs.PackageAvailable{
					bpopamp.CollectorPackageName: {
						Type:    protobufs.PackageType_PackageType_TopLevel,
						Version: updates.Version,
						File:    downloadableFile,
						Hash:    []byte(updates.Version),
					},
				},
			}
			agent, _ = s.manager.UpsertAgent(ctx, agent.ID, func(current *model.Agent) {
				current.UpgradeStarted(updates.Version, allPackagesHash)
			})

			s.logger.Info("sending PackagesAvailable", zap.Any("PackagesAvailable", serverToAgent.PackagesAvailable), zap.Any("Upgrade", agent.Upgrade))
		}
	}

	// if the message doesn't have a new configuration or a new package available, do nothing
	if serverToAgent.RemoteConfig == nil && serverToAgent.PackagesAvailable == nil {
		return nil
	}

	return s.send(ctx, conn, serverToAgent)
}

func (s *opampServer) getDownloadableFile(ctx context.Context, a *model.Agent, versionString string) (*protobufs.DownloadableFile, error) {
	version, err := s.manager.AgentVersion(ctx, versionString)
	if version == nil {
		return nil, fmt.Errorf("agent version %s not found", versionString)
	}
	if err != nil {
		return nil, err
	}
	platform := fmt.Sprintf("%s/%s", a.Platform, a.Architecture)
	artifact := version.Download(platform)

	if artifact == nil {
		return nil, fmt.Errorf("artifact not found for platform %s", platform)
	}

	url := artifact.URL
	hash := artifact.Hash
	if url == "" || hash == "" {
		return nil, nil
	}

	hashBytes, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	return &protobufs.DownloadableFile{
		DownloadUrl: url,
		ContentHash: hashBytes,
	}, nil
}

// SendHeartbeat sends a heartbeat to the agent to keep the websocket open
func (s *opampServer) SendHeartbeat(agentID string) error {
	state := s.connections.StateForAgentID(agentID)
	if state == nil {
		return nil
	}
	conn := state.Conn
	if conn != nil {
		return s.send(context.Background(), conn, &protobufs.ServerToAgent{})
	}
	return nil
}

// RequestReport sends report configuration to the specified agent
func (s *opampServer) RequestReport(ctx context.Context, agentID string, configuration protocol.Report) error {
	ctx, span := tracer.Start(ctx, "opamp/RequestReport", trace.WithAttributes(
		attribute.String("bindplane.agent.id", agentID)),
	)
	defer span.End()

	state := s.connections.StateForAgentID(agentID)
	if state == nil {
		return fmt.Errorf("no connection state for agentID %s", agentID)
	}

	conn := state.Conn
	if conn != nil {
		body, err := configuration.YAML()
		if err != nil {
			return err
		}
		s.logger.Info("RequestReport", zap.String(protocol.ReportName, string(body)))
		return s.send(context.Background(), conn, &protobufs.ServerToAgent{
			RemoteConfig: &protobufs.AgentRemoteConfig{
				ConfigHash: computeReportConfigurationHash(body),
				Config: &protobufs.AgentConfigMap{
					ConfigMap: map[string]*protobufs.AgentConfigFile{
						protocol.ReportName: {
							Body:        body,
							ContentType: "text/yaml",
						},
					},
				},
			},
		})
	}
	return nil
}

func (s *opampServer) Shutdown(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "opamp/Shutdown")
	defer span.End()

	connectedAgentIDs, err := s.ConnectedAgentIDs(ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("get connected agents: %w", err)
	}

	for _, id := range connectedAgentIDs {
		s.Disconnect(id)
	}

	return nil
}

func computeReportConfigurationHash(contents ...[]byte) []byte {
	h := sha256.New()
	for _, b := range contents {
		h.Write(b)
	}
	return h.Sum(nil)
}

func (s *opampServer) send(ctx context.Context, conn opamp.Connection, msg *protobufs.ServerToAgent) error {
	state := s.connections.StateForConnection(conn)
	state.SendLock.Lock()
	defer state.SendLock.Unlock()
	return conn.Send(ctx, msg)
}

// ----------------------------------------------------------------------

func (s *opampServer) verifyAgentConfig(ctx context.Context, conn opamp.Connection, agentID string, message *protobufs.AgentToServer, response *protobufs.ServerToAgent) error {
	ctx, span := tracer.Start(ctx, "opamp/verifyAgentConfig")
	defer span.End()

	// store the current configuration as reported by status
	agent, state, err := s.updateAgentState(ctx, agentID, conn, message, response)
	if err != nil {
		return fmt.Errorf("unable to update agent [%s]: %w", agentID, err)
	}

	return s.updateAgentConfig(ctx, agent, state, response)
}

// updateAgentConfig updates the current configuration by setting the RemoteConfig message if necessary
//
// This function is called when the agent connects and reports its configuration and BindPlane confirms that it is
// running the correct configuration. It gets the current Configuration for the Agent from the Manager.AgentUpdates
// method (which uses Store.AgentConfiguration) and compares it to the configuration reported by the agent. This is a
// PULL from the Agent.
func (s *opampServer) updateAgentConfig(ctx context.Context, agent *model.Agent, state *bpopamp.AgentState, response *protobufs.ServerToAgent) error {
	agentRawConfiguration := state.Configuration()
	if agentRawConfiguration == nil {
		s.logger.Info("no configuration available to verify, requesting from agent")
		response.Flags = uint64(protobufs.ServerToAgentFlags_ServerToAgentFlags_ReportFullState)
		return nil
	}

	agentConfiguration, err := agentRawConfiguration.Parse()
	if err != nil {
		// TODO(andy): ignore the current unparsable configuration and force new configuration?
		return fmt.Errorf("unable to parse the current agent configuration: %w", err)
	}

	// remove sensitive parameter masking when rendering for the agent
	ctx = model.ContextWithoutSensitiveParameterMasking(ctx)

	// check the manager for any updates that should be applied to this agent
	updates, err := s.manager.AgentUpdates(ctx, agent)
	if err != nil {
		return fmt.Errorf("unable to get agent updates [%s]: %w", agent.ID, err)
	}

	serverConfiguration, err := s.updatedConfiguration(ctx, agent, agentConfiguration, updates)
	if err != nil {
		return fmt.Errorf("unable to compute the updated agent configuration [%s]: %w", agent.ID, err)
	}

	// compare the configurations and compute a difference
	newConfiguration := observiq.ComputeConfigurationUpdates(&serverConfiguration, agentConfiguration)

	if newConfiguration.Empty() {
		// existing config is correct
		s.logger.Info("agent running with the correct config")
		s.updateAgentCurrentConfiguration(ctx, agent, updates.Configuration)
		return nil
	}

	rawNewConfiguration := newConfiguration.Raw()
	remoteConfig := agentRemoteConfig(&rawNewConfiguration, agentRawConfiguration)

	// check to see if we already tried this and received an error
	if bytes.Equal(state.Status.GetRemoteConfigStatus().GetLastRemoteConfigHash(), remoteConfig.GetConfigHash()) {
		s.logger.Info("already attempted to send this configuration")
		return nil
	}

	// change the agent status to Configuring, but ignore any failure as this status is considered nice to have and not
	// required to update the agent
	_, _ = s.manager.UpsertAgent(ctx, agent.ID, func(current *model.Agent) { current.Status = model.Configuring })

	s.logger.Info("agent running with outdated config", zap.Any("cur", agentConfiguration.Collector), zap.Any("new", serverConfiguration.Collector))
	response.RemoteConfig = remoteConfig

	return nil
}

func (s *opampServer) updatedConfiguration(ctx context.Context, agent *model.Agent, agentConfiguration *observiq.AgentConfiguration, updates *protocol.AgentUpdates) (diff observiq.AgentConfiguration, err error) {
	// Configuration => collector.yaml
	if updates.Configuration != nil {
		newCollectorYAML, err := updates.Configuration.Render(ctx, agent, s.manager.BindPlaneURL(), s.manager.BindPlaneInsecureSkipVerify(), s.manager.ResourceStore(), model.GetOssOtelHeaders())
		if err != nil {
			return diff, err
		}
		diff.Collector = newCollectorYAML
	}

	// Labels => manager.yaml
	if updates.Labels != nil && !agentConfiguration.HasLabels(updates.Labels.String()) {
		diff.Manager = agentConfiguration.Manager
		diff.ReplaceLabels(updates.Labels.String())
	}

	return diff, nil
}

// agentRemoteConfig generates the protobuf for sending this Config to an agent using the OpAMP protocol
func agentRemoteConfig(updates *observiq.RawAgentConfiguration, agentRaw *observiq.RawAgentConfiguration) *protobufs.AgentRemoteConfig {
	// only store the configs that exist for the agent
	configMap := map[string]*protobufs.AgentConfigFile{}
	if updates.Collector != nil {
		configMap[observiq.CollectorFilename] = &protobufs.AgentConfigFile{Body: updates.Collector}
	}
	if updates.Logging != nil {
		configMap[observiq.LoggingFilename] = &protobufs.AgentConfigFile{Body: updates.Logging}
	}
	if updates.Manager != nil {
		configMap[observiq.ManagerFilename] = &protobufs.AgentConfigFile{Body: updates.Manager}
	}

	return &protobufs.AgentRemoteConfig{
		Config: &protobufs.AgentConfigMap{
			ConfigMap: configMap,
		},
		ConfigHash: computeHash(updates, agentRaw),
	}
}

func computeHash(updates *observiq.RawAgentConfiguration, agentRaw *observiq.RawAgentConfiguration) []byte {
	combined := agentRaw.ApplyUpdates(updates)
	return combined.Hash()
}

func (s *opampServer) updateAgentState(ctx context.Context, agentID string, conn opamp.Connection, msg *protobufs.AgentToServer, response *protobufs.ServerToAgent) (agent *model.Agent, state *bpopamp.AgentState, err error) {
	agent, err = s.manager.UpsertAgent(ctx, agentID, func(agent *model.Agent) {
		// we're using opamp
		agent.Protocol = ProtocolName

		// decode the state which we will update
		state, err = bpopamp.DecodeState(agent.State)
		if err != nil {
			s.logger.Error("error encountered while decoding agent state, starting with fresh state", zap.Error(err))
		}

		bpopamp.SyncOne[*protobufs.AgentDescription](ctx, s.logger, msg, state, conn, agent, response, &bpopamp.SyncAgentDescription)
		bpopamp.SyncOne[*protobufs.EffectiveConfig](ctx, s.logger, msg, state, conn, agent, response, &bpopamp.SyncEffectiveConfig)
		bpopamp.SyncOne[*protobufs.RemoteConfigStatus](ctx, s.logger, msg, state, conn, agent, response, &bpopamp.SyncRemoteConfigStatus)
		bpopamp.SyncOne[*protobufs.PackageStatuses](ctx, s.logger, msg, state, conn, agent, response, &bpopamp.SyncPackageStatuses)

		// after sync, update sequence number
		state.SequenceNum = msg.GetSequenceNum()

		// always update the agent status, regardless of RemoteConfigStatus message being present
		bpopamp.UpdateAgentStatus(s.logger, agent, state.Status.GetRemoteConfigStatus())

		// update ConnectedAt, etc
		if msg.GetAgentDisconnect() != nil {
			agent.Disconnect()
		} else {
			agent.Connect(agent.Version)
		}

		// the state could be new
		agent.State = bpopamp.EncodeState(state)
	})

	return agent, state, err
}

func (s *opampServer) updateAgentCurrentConfiguration(ctx context.Context, agent *model.Agent, configuration *model.Configuration) {
	_, err := s.manager.UpsertAgent(ctx, agent.ID, func(current *model.Agent) {
		current.SetCurrentConfiguration(configuration)
	})
	if err != nil {
		// if we were unable to set the Current configuration, the configuration will still be Pending and we will try again
		s.logger.Error("unable to SetCurrentConfiguration", zap.Error(err))
	}
}
