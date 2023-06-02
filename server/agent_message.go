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

// Package server contains the [Manager] interface
package server

import (
	"github.com/observiq/bindplane-op/server/protocol"
)

const (
	// AgentMessageTypeSnapshot is the type of message that is sent to an agent to request a snapshot
	AgentMessageTypeSnapshot = "snapshot"
)

// Message is a message that is sent to an agent
type Message interface {
	AgentID() string
	Type() string
	Body() map[string]interface{}
}

// AgentMessage is a message that is sent to an agent
type AgentMessage struct {
	AgentIDField string         `json:"agent_id"`
	TypeField    string         `json:"type"`
	BodyField    map[string]any `json:"body"`
}

// AgentID returns the agent ID of the message
func (m AgentMessage) AgentID() string {
	return m.AgentIDField
}

// Type returns the type of the message
func (m AgentMessage) Type() string {
	return m.TypeField
}

// Body returns the body of the message
func (m AgentMessage) Body() map[string]interface{} {
	return m.BodyField
}

// SnapshotBody is the body of a snapshot message
type SnapshotBody struct {
	Configuration protocol.Report `json:"configuration" mapstructure:"configuration"`
}
