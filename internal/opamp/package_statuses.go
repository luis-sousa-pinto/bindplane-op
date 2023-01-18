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
	"bytes"
	"context"

	"github.com/observiq/bindplane-op/model"
	"github.com/open-telemetry/opamp-go/protobufs"
	opamp "github.com/open-telemetry/opamp-go/server/types"
	"go.uber.org/zap"
)

// ----------------------------------------------------------------------
// RemoteConfigStatus

type packageStatusesSyncer struct{}

var _ messageSyncer[*protobufs.PackageStatuses] = (*packageStatusesSyncer)(nil)

func (s *packageStatusesSyncer) name() string {
	return "PackageStatuses"
}

func (s *packageStatusesSyncer) message(msg *protobufs.AgentToServer) (result *protobufs.PackageStatuses, exists bool) {
	result = msg.GetPackageStatuses()
	return result, result != nil
}

func (s *packageStatusesSyncer) agentCapabilitiesFlag() protobufs.AgentCapabilities {
	return protobufs.AgentCapabilities_ReportsPackageStatuses
}

func (s *packageStatusesSyncer) update(_ context.Context, _ *zap.Logger, state *agentState, _ opamp.Connection, agent *model.Agent, value *protobufs.PackageStatuses) error {
	// if an upgrade is in progress but the hash doesn't match, ignore the message. this could happen if two upgrades
	// happen in quick succession and we will get another update soon.
	if agent.Upgrade != nil && !bytes.Equal(agent.Upgrade.AllPackagesHash, value.ServerProvidedAllPackagesHash) {
		return nil
	}

	state.Status.PackageStatuses = value

	upgradeComplete := false

	errorMessage := value.ErrorMessage
	var agentVersion string
	if agent.Upgrade != nil {
		agentVersion = agent.Upgrade.Version
	}

	if packages := value.GetPackages(); packages != nil {
		if collector := packages[CollectorPackageName]; collector != nil {
			upgradeComplete = collector.Status == protobufs.PackageStatus_InstallFailed || collector.Status == protobufs.PackageStatus_Installed
			agentVersion = collector.AgentHasVersion
			if collector.ErrorMessage != "" {
				errorMessage = collector.ErrorMessage
			}
		}
	}

	if upgradeComplete {
		agent.UpgradeComplete(agentVersion, errorMessage)
	}

	return nil
}
