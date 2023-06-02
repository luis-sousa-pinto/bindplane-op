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

package store

import (
	"context"
	"testing"

	"github.com/observiq/bindplane-op/model"
	"github.com/stretchr/testify/require"
)

func TestNewRolloutUpdates(t *testing.T) {
	agentEvents := NewEvents[*model.Agent]()
	agent1 := &model.Agent{
		ID: "1",
		ConfigurationStatus: model.ConfigurationVersions{
			Current: "one:0",
			Pending: "",
			Future:  "",
		},
	}
	agentEvents.Include(agent1, EventTypeUpdate)

	agent2 := &model.Agent{
		ID: "2",
		ConfigurationStatus: model.ConfigurationVersions{
			Current: "two:0",
			Pending: "two:1",
			Future:  "",
		},
	}
	agentEvents.Include(agent2, EventTypeUpdate)

	actual := NewRolloutUpdates(context.Background(), agentEvents)
	require.Len(t, actual.Updates(), 2)

	require.True(t, actual.Updates().ContainsKey(agent1.ConfigurationStatus.UniqueKey()))
	require.True(t, actual.Updates().ContainsKey(agent2.ConfigurationStatus.UniqueKey()))
}

func TestNewRolloutUpdatesMerge(t *testing.T) {
	agentEvents1 := NewEvents[*model.Agent]()
	agent1 := &model.Agent{
		ID: "1",
		ConfigurationStatus: model.ConfigurationVersions{
			Current: "one:0",
			Pending: "",
			Future:  "",
		},
	}
	agentEvents1.Include(agent1, EventTypeUpdate)

	agent2 := &model.Agent{
		ID: "2",
		ConfigurationStatus: model.ConfigurationVersions{
			Current: "two:0",
			Pending: "two:1",
			Future:  "",
		},
	}
	agentEvents1.Include(agent2, EventTypeUpdate)

	updates1 := NewRolloutUpdates(context.Background(), agentEvents1)

	agentEvents2 := NewEvents[*model.Agent]()
	agent3 := &model.Agent{
		ID: "3",
		ConfigurationStatus: model.ConfigurationVersions{
			Current: "three:0",
			Pending: "",
			Future:  "three:1",
		},
	}
	agentEvents2.Include(agent3, EventTypeUpdate)

	updates2 := NewRolloutUpdates(context.Background(), agentEvents2)

	updates1.Merge(updates2)

	require.Len(t, updates1.Updates(), 3)
}

func TestRolloutUpdatesEmpty(t *testing.T) {
	agentEvents := NewEvents[*model.Agent]()
	updates := NewRolloutUpdates(context.Background(), agentEvents)
	require.True(t, updates.Empty())
}
