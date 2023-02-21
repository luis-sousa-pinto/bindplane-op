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

package opamp

import (
	"sync"

	opamp "github.com/open-telemetry/opamp-go/server/types"
)

// Connection is an interface that wraps the opamp.Connection interface
//
//go:generate mockery --name Connection --filename mock_connection.go --structname MockConnection
type Connection interface {
	opamp.Connection
}

type connections struct {
	// maps connection => agentID and agentID => connection
	locks       map[Connection]*sync.Mutex
	connections map[Connection]string
	agents      map[string]Connection
	mtx         sync.RWMutex
}

func newConnections() *connections {
	return &connections{
		locks:       make(map[Connection]*sync.Mutex),
		connections: make(map[Connection]string),
		agents:      make(map[string]Connection),
	}
}

func (c *connections) connect(conn Connection, agentID string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.locks[conn] = &sync.Mutex{}
	c.connections[conn] = agentID
	c.agents[agentID] = conn
}

func (c *connections) disconnect(conn Connection) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	agentID, ok := c.connections[conn]
	if ok {
		delete(c.locks, conn)
		delete(c.connections, conn)
		delete(c.agents, agentID)
	}
}

// connected returns true if the agent with the specified agentID is connected
func (c *connections) connected(agentID string) bool {
	return c.connection(agentID) != nil
}

// connection returns the current Connection for the specified agentID or nil if there is no connection
func (c *connections) connection(agentID string) Connection {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	return c.agents[agentID]
}

func (c *connections) agentID(conn Connection) string {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	return c.connections[conn]
}

func (c *connections) sendLock(conn Connection) *sync.Mutex {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	return c.locks[conn]
}

func (c *connections) agentIDs() []string {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	ids := []string{}
	for id := range c.agents {
		ids = append(ids, id)
	}
	return ids
}
