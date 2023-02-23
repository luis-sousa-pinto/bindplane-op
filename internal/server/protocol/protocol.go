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

// Package protocol defines communication to agents
package protocol

import (
	"context"

	"github.com/observiq/bindplane-op/model"
)

// Protocol represents a communication protocol for managing agents
//
//go:generate mockery --name Protocol --filename mock_protocol.go --structname MockProtocol
type Protocol interface {
	// Name is the name for the protocol use mostly for logging
	Name() string

	// Connected returns true if the specified agent ID is connected
	Connected(agentID string) bool

	// ConnectedAgents should return a slice of the currently connected agent IDs
	ConnectedAgentIDs(context.Context) ([]string, error)

	// Disconnect closes the any connection to the specified agent ID
	Disconnect(agentID string) bool

	// UpdateAgent should send a message to the specified agent to apply the updates
	UpdateAgent(context.Context, *model.Agent, *AgentUpdates) error

	// SendHeartbeat sends a heartbeat to the agent to keep the websocket open
	SendHeartbeat(agentID string) error

	// RequestReport sends report configuration to the specified agent
	RequestReport(ctx context.Context, agentID string, report Report) error
}
