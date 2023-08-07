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

// Package connections provides an interface to manage agent connections and manage per-connection state
package connections

import (
	"context"
	"sync"

	bpopamp "github.com/observiq/bindplane-op/opamp"
	opamp "github.com/open-telemetry/opamp-go/server/types"
)

type connections struct {
	// maps [opamp.Connection] => agentConnectionState
	connections map[bpopamp.Connection]*bpopamp.AgentConnectionState
	// maps agentID => [agentConnectionState]
	agents map[string]*bpopamp.AgentConnectionState
	mtx    sync.RWMutex
}

// NewConnections creates a new [connections] object with empty values
func NewConnections() bpopamp.Connections[*bpopamp.AgentConnectionState] {
	return &connections{
		connections: make(map[bpopamp.Connection]*bpopamp.AgentConnectionState),
		agents:      make(map[string]*bpopamp.AgentConnectionState),
	}
}

func (c *connections) OnConnecting(_ context.Context, agentID string) *bpopamp.AgentConnectionState {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	state := &bpopamp.AgentConnectionState{
		AgentID:  agentID,
		SendLock: sync.Mutex{},
	}
	s, ok := c.agents[agentID]
	if !ok {
		c.agents[agentID] = state
		return state
	}

	return s
}

func (c *connections) OnMessage(agentID string, conn opamp.Connection) (*bpopamp.AgentConnectionState, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	state, ok := c.agents[agentID]
	if !ok {
		return nil, bpopamp.ErrAgentNotRegistered
	}

	c.connections[conn] = state
	state.Conn = conn
	return state, nil
}

func (c *connections) OnConnectionClose(conn opamp.Connection) (state *bpopamp.AgentConnectionState, count int) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	state, ok := c.connections[conn]
	if ok {
		delete(c.connections, conn)
		delete(c.agents, state.AgentID)
		count = len(c.agents)
	}
	return state, count
}

func (c *connections) Connected(agentID string) bool {
	return c.Connection(agentID) != nil
}

func (c *connections) Connection(agentID string) bpopamp.Connection {
	c.mtx.RLock()
	defer c.mtx.RUnlock()

	if state, ok := c.agents[agentID]; ok {
		return state.Conn
	}
	return nil
}

func (c *connections) ConnectedAgentIDs(_ context.Context) []string {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	ids := []string{}
	for id := range c.agents {
		ids = append(ids, id)
	}
	return ids
}

func (c *connections) ConnectedAgentsCount(_ context.Context) int {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	return len(c.agents)
}

func (c *connections) StateForAgentID(agentID string) *bpopamp.AgentConnectionState {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	return c.agents[agentID]
}

func (c *connections) StateForConnection(conn opamp.Connection) *bpopamp.AgentConnectionState {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	return c.connections[conn]
}
