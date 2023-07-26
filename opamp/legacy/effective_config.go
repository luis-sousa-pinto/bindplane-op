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
	"fmt"

	"github.com/observiq/bindplane-op/model"
	"github.com/observiq/bindplane-op/model/observiq"
	"github.com/observiq/opamp-go/protobufs"
	opamp "github.com/observiq/opamp-go/server/types"
	"go.uber.org/zap"
)

// ----------------------------------------------------------------------
// EffectiveConfig

type effectiveConfigSyncer struct{}

func (s *effectiveConfigSyncer) name() string {
	return "EffectiveConfig"
}

func (s *effectiveConfigSyncer) message(msg *protobufs.AgentToServer) (result *protobufs.EffectiveConfig, exists bool) {
	result = msg.GetEffectiveConfig()
	return result, result != nil
}

func (s *effectiveConfigSyncer) agentCapabilitiesFlag() protobufs.AgentCapabilities {
	return protobufs.AgentCapabilities_ReportsEffectiveConfig
}

func (s *effectiveConfigSyncer) update(_ context.Context, _ *zap.Logger, state *AgentState, _ opamp.Connection, agent *model.Agent, value *protobufs.EffectiveConfig) error {
	state.Status.EffectiveConfig = value

	// parse the configuration
	agentRawConfiguration := agentCurrentConfiguration(value)
	agentConfiguration, err := agentRawConfiguration.Parse()
	if err != nil {
		return fmt.Errorf("unable to parse the current agent configuration: %w", err)
	}

	if agentConfiguration.Manager != nil && agentConfiguration.Manager.TLS != nil {
		// Synchronize TLS settings

		managerTLS := agentConfiguration.Manager.TLS
		agent.TLS = &model.ManagerTLS{
			InsecureSkipVerify: managerTLS.InsecureSkipVerify,
			CAFile:             managerTLS.CAFile,
			CertFile:           managerTLS.CertFile,
			KeyFile:            managerTLS.KeyFile,
		}
	}

	// save the entire configuration so we have it to report
	agent.Configuration = agentConfiguration

	return nil
}

// agentCurrentConfiguration parses the AgentConfiguration from the EffectiveConfig
func agentCurrentConfiguration(effectiveConfig *protobufs.EffectiveConfig) *observiq.RawAgentConfiguration {
	configMap := effectiveConfig.GetConfigMap().GetConfigMap()
	raw := &observiq.RawAgentConfiguration{
		Collector: configMap[observiq.CollectorFilename].GetBody(),
		Logging:   configMap[observiq.LoggingFilename].GetBody(),
		Manager:   configMap[observiq.ManagerFilename].GetBody(),
	}
	return raw
}
