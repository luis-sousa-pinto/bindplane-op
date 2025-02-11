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

	"github.com/observiq/bindplane-op/model"
	"github.com/observiq/opamp-go/protobufs"
	opamp "github.com/observiq/opamp-go/server/types"
	"go.uber.org/zap"
)

// ----------------------------------------------------------------------
// RemoteConfigStatus

type remoteConfigStatusSyncer struct{}

var _ messageSyncer[*protobufs.RemoteConfigStatus] = (*remoteConfigStatusSyncer)(nil)

func (s *remoteConfigStatusSyncer) name() string {
	return "RemoteConfigStatus"
}

func (s *remoteConfigStatusSyncer) message(msg *protobufs.AgentToServer) (result *protobufs.RemoteConfigStatus, exists bool) {
	result = msg.GetRemoteConfigStatus()
	return result, result != nil
}

func (s *remoteConfigStatusSyncer) agentCapabilitiesFlag() protobufs.AgentCapabilities {
	return protobufs.AgentCapabilities_AcceptsRemoteConfig
}

func (s *remoteConfigStatusSyncer) update(_ context.Context, _ *zap.Logger, state *AgentState, _ opamp.Connection, _ *model.Agent, value *protobufs.RemoteConfigStatus) error {
	state.Status.RemoteConfigStatus = value
	return nil
}

// UpdateAgentStatus modifies the agent status based on the RemoteConfigStatus, if available
func UpdateAgentStatus(logger *zap.Logger, agent *model.Agent, remoteStatus *protobufs.RemoteConfigStatus) {
	// if we failed the apply, enter or update an error state
	if remoteStatus.GetStatus() == protobufs.RemoteConfigStatus_FAILED {
		logger.Info("got RemoteConfigStatus_FAILED", zap.String("ErrorMessage", remoteStatus.ErrorMessage))
		agent.Status = model.Error
		agent.ErrorMessage = remoteStatus.ErrorMessage
		return
	}
	switch agent.Status {
	case model.Error:
		// only way to clear the error is to have a successful apply
		if remoteStatus.GetStatus() == protobufs.RemoteConfigStatus_APPLIED {
			agent.Status = model.Connected
			agent.ErrorMessage = ""
		}
	case model.Upgrading:
		// upgrading will be cleared by model.Agent.UpgradeComplete
	default:
		// either RemoteConfigStatus wasn't sent or wasn't failed
		agent.Status = model.Connected
		agent.ErrorMessage = ""
	}
}
