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
	"context"
	"database/sql/driver"
	"encoding/base64"
	"errors"

	jsoniter "github.com/json-iterator/go"

	"github.com/golang/protobuf/proto"
	"github.com/mitchellh/mapstructure"
	"github.com/observiq/bindplane-op/model"
	"github.com/observiq/bindplane-op/model/observiq"
	"github.com/open-telemetry/opamp-go/protobufs"
	opamp "github.com/open-telemetry/opamp-go/server/types"
	"go.uber.org/zap"
	"google.golang.org/protobuf/runtime/protoiface"
)

// These are the functions that are used to update the agent state from the OpAMP messages
var (
	SyncAgentDescription   = agentDescriptionSyncer{}
	SyncEffectiveConfig    = effectiveConfigSyncer{}
	SyncRemoteConfigStatus = remoteConfigStatusSyncer{}
	SyncPackageStatuses    = packageStatusesSyncer{}
)

// ----------------------------------------------------------------------

// SerializedAgentState is stored on the model.Agent in a partially serialized form. The status is a base64-encoded protobuf.
type SerializedAgentState struct {
	SequenceNum uint64 `json:"sequenceNum" yaml:"sequenceNum" mapstructure:"sequenceNum"`
	Status      string `json:"status,omitempty" yaml:"status,omitempty" mapstructure:"status"`
}

// Value is used to translate to a JSONB field for postgres storage
func (s SerializedAgentState) Value() (driver.Value, error) {
	return jsoniter.Marshal(s)
}

// Scan is used to translate from a JSONB field in postgres to serializedAgentState
func (s *SerializedAgentState) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return jsoniter.Unmarshal(b, &s)
}

// AgentState stores OpAMP state for the agent
type AgentState struct {
	SequenceNum uint64                  `json:"sequenceNum" yaml:"sequenceNum" mapstructure:"sequenceNum"`
	Status      protobufs.AgentToServer `json:"status,omitempty" yaml:"status,omitempty" mapstructure:"status"`
}

// EncodeState encodes the agent state into a serialized form
func EncodeState(state *AgentState) *SerializedAgentState {
	if state == nil {
		return &SerializedAgentState{}
	}
	bytes, err := proto.Marshal(&state.Status)
	if err != nil {
		bytes = nil
	}
	serialized := &SerializedAgentState{
		SequenceNum: state.SequenceNum,
		Status:      base64.StdEncoding.EncodeToString(bytes),
	}
	return serialized
}

// DecodeState decodes the agent state from a serialized form
func DecodeState(state interface{}) (*AgentState, error) {
	serialized := SerializedAgentState{}

	if err := mapstructure.Decode(state, &serialized); err != nil {
		return &AgentState{
			SequenceNum: serialized.SequenceNum,
		}, err
	}

	result := &AgentState{
		SequenceNum: serialized.SequenceNum,
	}

	bytes, err := base64.StdEncoding.DecodeString(serialized.Status)
	if err != nil {
		return result, err
	}

	// unmarshal proto
	if err := proto.Unmarshal(bytes, &result.Status); err != nil {
		return result, err
	}

	return result, nil
}

// Configuration returns the current configuration
func (s *AgentState) Configuration() *observiq.RawAgentConfiguration {
	if ec := s.Status.GetEffectiveConfig(); ec != nil {
		return agentCurrentConfiguration(ec)
	}
	return nil
}

// UpdateSequenceNumber updates the sequence number
func (s *AgentState) UpdateSequenceNumber(agentToServer *protobufs.AgentToServer) {
	s.SequenceNum = agentToServer.GetSequenceNum()
}

// IsMissingMessage returns true if the message is missing
func (s *AgentState) IsMissingMessage(agentToServer *protobufs.AgentToServer) bool {
	return agentToServer.GetSequenceNum()-s.SequenceNum != 1
}

// ----------------------------------------------------------------------

// interface that defines how to sync each message
type messageSyncer[T protoiface.MessageV1] interface {
	// name is useful for debugging
	name() string

	// message returns the message within the AgentToServer the is being synced
	message(msg *protobufs.AgentToServer) (T, bool)

	// apply applies the updated message to the specified AgentToServer
	update(ctx context.Context, logger *zap.Logger, state *AgentState, conn opamp.Connection, agent *model.Agent, value T) error

	// agentCapabilitiesFlag returns the flag to check on the agent to determine if it supports this message. If
	// unsupported, the reportFlag will not be specified.
	agentCapabilitiesFlag() protobufs.AgentCapabilities
}

// ----------------------------------------------------------------------

// SyncOne syncs a single message
func SyncOne[T protoiface.MessageV1](ctx context.Context, logger *zap.Logger, agentToServer *protobufs.AgentToServer, state *AgentState, conn opamp.Connection, agent *model.Agent, response *protobufs.ServerToAgent, syncer messageSyncer[T]) (updated bool) {
	agentMessage, agentMessageExists := syncer.message(agentToServer)
	localMessage, localMessageExists := syncer.message(&state.Status)

	initialSyncRequired := !localMessageExists && !agentMessageExists
	serverSkippedMessage := state.IsMissingMessage(agentToServer)

	// make sure we have a message
	if initialSyncRequired || serverSkippedMessage {
		// Either:
		//
		// 1) agent doesn't have the message at all => request contents
		//
		// 2) we missed a messages in sequence => request contents
		//
		logger.Debug("not synced or missed message => ReportFullState",
			zap.String("syncer", syncer.name()),
			zap.Bool("serverSkippedMessage", serverSkippedMessage),
			zap.Bool("initialSyncRequired", initialSyncRequired),
		)
		if hasCapability(agentToServer, syncer.agentCapabilitiesFlag()) {
			response.Flags = uint64(protobufs.ServerToAgentFlags_ServerToAgentFlags_ReportFullState)
		}
		return false
	}

	if localMessageExists {
		if !agentMessageExists || proto.Equal(agentMessage, localMessage) {
			// data on the server is present and matches content => do nothing
			logger.Debug("exists locally and unchanged => do nothing", zap.String("syncer", syncer.name()))
			return false
		}
	}

	// before attempting to store, make sure we clone the message
	agentMessage = proto.Clone(agentMessage).(T)

	// update
	if err := syncer.update(ctx, logger, state, conn, agent, agentMessage); err != nil {
		logger.Debug("message different => update error", zap.String("syncer", syncer.name()), zap.Error(err))
		errorMessage := err.Error()
		if response.ErrorResponse != nil {
			errorMessage = response.ErrorResponse.ErrorMessage + ", " + errorMessage
		}
		response.ErrorResponse = &protobufs.ServerErrorResponse{
			Type:         protobufs.ServerErrorResponseType_ServerErrorResponseType_Unknown,
			ErrorMessage: errorMessage,
		}
	} else {
		logger.Debug("message different => update", zap.String("syncer", syncer.name()))
	}

	return true
}

func hasCapability(agentToServer *protobufs.AgentToServer, capability protobufs.AgentCapabilities) bool {
	return uint64(capability)&agentToServer.GetCapabilities() != 0
}

// ----------------------------------------------------------------------
// misc utils

// MessageComponents returns the names of the components in the message
func MessageComponents(agentToServer *protobufs.AgentToServer) []string {
	var components []string
	components = includeComponent(components, agentToServer.AgentDescription, "AgentDescription")
	components = includeComponent(components, agentToServer.EffectiveConfig, "EffectiveConfig")
	components = includeComponent(components, agentToServer.RemoteConfigStatus, "RemoteConfigStatus")
	components = includeComponent(components, agentToServer.PackageStatuses, "PackageStatuses")
	return components
}

func includeComponent(components []string, msg any, name string) []string {
	if msg != nil {
		components = append(components, name)
	}
	return components
}
