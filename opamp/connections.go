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

package opamp

import (
	"context"
	"errors"
	"sync"

	opamp "github.com/open-telemetry/opamp-go/server/types"
)

// Connections is an interface to manage agent connections and manage per-connection state.
// The generic type S can be specified, but [AgentConnectionState] is a good default
//
//go:generate mockery --name Connections --filename mock_connections.go --structname MockConnections
type Connections[S any] interface {
	// Connected returns true if the agent with the specified agentID is connected
	Connected(agentID string) bool

	// ConnectedAgentIDs returns the IDs of all connected agents
	ConnectedAgentIDs(context.Context) []string

	// ConnectedAgentsCount returns the number of connected agents
	ConnectedAgentsCount(context.Context) int

	// OnConnecting should be called within the opamp OnConnecting callback. It adds the association of the accountID to
	// the agentID after it has been determined. The associated opamp.Connection is not yet available. It will be added in
	// OnMessage.
	OnConnecting(ctx context.Context, agentID string) S

	// OnMessage should be called within the opamp OnMessage callback. It adds the association of the opamp.Connection to
	// the connectionState.
	OnMessage(agentID string, conn opamp.Connection) (S, error)

	// OnConnectionClose removes the connection and returns the state and the count of remaining agents connected.
	OnConnectionClose(conn opamp.Connection) (state S, count int)

	// StateForAgentID returns the current state for the specified agentID or nil if there is no connection
	StateForAgentID(agentID string) S

	// StateForConnection returns the current state for the specified [opamp.Connection] or nil if there is no connection
	StateForConnection(opamp.Connection) S
}

// AgentConnectionState is the state of an agent connection
type AgentConnectionState struct {
	AgentID  string
	Conn     opamp.Connection
	SendLock sync.Mutex
}

// ErrAgentNotRegistered is returned when an agent that is not registered sends a message
var ErrAgentNotRegistered = errors.New("agent not registered")
