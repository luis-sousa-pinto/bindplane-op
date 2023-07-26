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

package connections

import (
	"context"
	"testing"

	"github.com/observiq/bindplane-op/opamp/mocks"
	"github.com/stretchr/testify/require"
)

func TestConnect(t *testing.T) {
	ctx := context.Background()

	agentID := "1"
	c := NewConnections()
	conn := mocks.NewMockConnection(t)

	c.OnConnecting(ctx, agentID)
	c.OnMessage(agentID, conn)
	require.Equal(t, []string{agentID}, c.ConnectedAgentIDs(ctx), "should have agentID 1 connected")
	require.Equal(t, conn, c.StateForAgentID(agentID).Conn, "should be able to lookup connection by agentID")
	require.Equal(t, agentID, c.StateForConnection(conn).AgentID, "should be able to lookup agentID by connection")
}

func TestDisconnect(t *testing.T) {
	ctx := context.Background()
	agentID := "1"
	c := NewConnections()

	conn := mocks.NewMockConnection(t)
	c.OnConnecting(ctx, agentID)
	c.OnMessage(agentID, conn)
	require.Equal(t, []string{agentID}, c.ConnectedAgentIDs(ctx), "should have agentID 1 connected")
	_, count := c.OnConnectionClose(conn)
	require.Equal(t, 0, count)
	require.Equal(t, []string{}, c.ConnectedAgentIDs(ctx), "should have no connections")
	require.Equal(t, 0, c.ConnectedAgentsCount(ctx), "should have no connected agents")
	require.Nil(t, c.StateForAgentID(agentID), "should have no connection by agentID")
	require.Nil(t, c.StateForConnection(conn), "should have no agentID by connection")
}

func TestConnected(t *testing.T) {
	ctx := context.Background()
	c := NewConnections()
	conn := mocks.NewMockConnection(t)

	c.OnConnecting(ctx, "1")
	c.OnMessage("1", conn)
	require.Equal(t, []string{"1"}, c.ConnectedAgentIDs(ctx), "should have agentID 1 connected")
	require.True(t, c.Connected("1"), "should have agentID 1 connected")
}

func TestErrorOnNotConnectedAgent(t *testing.T) {
	agentID := "1"
	c := NewConnections()
	conn := mocks.NewMockConnection(t)

	state, err := c.OnMessage(agentID, conn)
	require.Error(t, err)
	require.Nil(t, state)
}

func TestNoChangeOnConnectionClose(t *testing.T) {
	agentIDs := []string{"1", "2"}
	ctx := context.Background()

	c := NewConnections()
	conn0 := mocks.NewMockConnection(t)
	conn1 := mocks.NewMockConnection(t)
	c.OnConnecting(ctx, agentIDs[0])
	state, err := c.OnMessage(agentIDs[0], conn0)
	require.NoError(t, err)
	require.NotNil(t, state)

	// conn1 isn't connected, so it should return 0, nil
	state, count := c.OnConnectionClose(conn1)
	require.Nil(t, state)
	require.Equal(t, 0, count)
}
