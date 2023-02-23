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

package protocol

import "github.com/observiq/bindplane-op/model"

// AgentUpdates contains fields that can be modified on an Agent and should be sent to the agent. The model.Agent should
// not be updated directly and will be updated when the agent reports its new status after the update is complete.
type AgentUpdates struct {
	// Labels changes are only supported by OpAMP
	Labels *model.Labels

	// Configuration changes are only supported by OpAMP
	Configuration *model.Configuration

	// Version instructs the agent to install a specific version
	Version string
}

// Empty returns true if the updates are empty because no changes need to be made to the agent
func (u *AgentUpdates) Empty() bool {
	return u.Labels == nil && u.Configuration == nil && u.Version == ""
}
