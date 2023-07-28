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
	"path/filepath"
	"testing"
	"time"

	"github.com/observiq/bindplane-op/model"
	"github.com/observiq/bindplane-op/resources"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

func testResource[T model.Resource](t *testing.T, name string) T {
	return fileResource[T](t, filepath.Join("testfiles", name))
}

func fileResource[T model.Resource](t *testing.T, path string) T {
	resources, err := model.ResourcesFromFile(path)
	require.NoError(t, err)

	parsed, err := model.ParseResources(resources)
	require.NoError(t, err)
	require.Len(t, parsed, 1)

	resource, ok := parsed[0].(T)
	require.True(t, ok)
	return resource
}

// RunTestSeedDeprecated runs tests for the deprecated resources in the seed folder
func RunTestSeedDeprecated(ctx context.Context, t *testing.T, store Store) {
	logger := zap.NewNop()
	tests := []struct {
		name            string
		setup           func()
		expectExists    []string
		expectNotExists []string
	}{
		{
			name:  "empty store, no deprecated resources are seeded",
			setup: func() {},
			expectNotExists: []string{
				"add_attribute",
				"add_resource",
				"filter_log_record_attribute",
				"filter_resource_attribute",
			},
		},
		{
			name: "store with deprecated, update deprecated with a new version",
			setup: func() {
				// add a deprecated resource to the store (set the name to add_attribute)
				severityProcessorType := testResource[*model.ProcessorType](t, "filter_severity.yaml")
				severityProcessorType.Metadata.Name = "add_attribute"
				_, err := store.ApplyResources(ctx, []model.Resource{severityProcessorType})
				require.NoError(t, err)
			},
			expectExists: []string{
				"add_attribute",
				"add_attribute:1",
				"add_attribute:2",
			},
			expectNotExists: []string{
				"add_resource",
				"filter_log_record_attribute",
				"filter_resource_attribute",
			},
		},
	}
	for _, test := range tests {
		if test.setup != nil {
			test.setup()
		}
		err := Seed(ctx, store, logger, resources.Files, resources.SeedFolders)
		require.NoError(t, err)
		for _, name := range test.expectExists {
			pt, err := store.ProcessorType(ctx, name)
			require.NoError(t, err)
			require.NotNil(t, pt)
		}
		for _, name := range test.expectNotExists {
			pt, err := store.ProcessorType(ctx, name)
			require.NoError(t, err)
			require.Nil(t, pt)
		}

	}
}

// RunReportConnectedAgentsTests runs tests for reporting connected agents and cleaning up unreported agents
func RunReportConnectedAgentsTests(ctx context.Context, t *testing.T, store Store) {
	agentIDs := []string{
		"agent-1",
		"agent-2",
		"agent-3",
		"agent-4",
		"agent-5",
	}

	tests := []struct {
		name     string
		agentIDs []string
	}{
		{
			name: "no agents",
		},
		{
			name:     "one agent",
			agentIDs: []string{"agent-1"},
		},
		{
			name:     "two agents",
			agentIDs: []string{"agent-3", "agent-4"},
		},
	}

	oldDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	newDate := time.Date(2020, 1, 1, 0, 0, 1, 0, time.UTC)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// delete all agents
			_, err := store.DeleteAgents(ctx, agentIDs)
			require.NoError(t, err)

			// require no agents
			agents, err := store.Agents(ctx)
			require.NoError(t, err)
			require.Len(t, agents, 0)

			// add agents with an old ReportedAt date
			_, err = store.UpsertAgents(ctx, agentIDs, func(agent *model.Agent) {
				agent.Status = model.Connected
				agent.ReportedAt = &oldDate
			})
			require.NoError(t, err)

			// report with a new date
			err = store.ReportConnectedAgents(ctx, test.agentIDs, newDate)
			require.NoError(t, err)

			// require that the agents have the new ReportedAt date if they were reported
			agents, err = store.Agents(ctx)
			require.NoError(t, err)
			for _, agent := range agents {
				require.NotNil(t, agent.ReportedAt, "agent %s has nil ReportedAt", agent.ID)
				if slices.Contains(test.agentIDs, agent.ID) {
					require.Equal(t, newDate.Unix(), agent.ReportedAt.Unix())
				} else {
					require.Equal(t, oldDate.Unix(), agent.ReportedAt.Unix())
				}
			}

			// disconnect unreported agents
			err = store.DisconnectUnreportedAgents(ctx, newDate)
			require.NoError(t, err)

			// require that the agents with an old ReportedAt date are disconnected
			agents, err = store.Agents(ctx)
			require.NoError(t, err)
			for _, agent := range agents {
				if slices.Contains(test.agentIDs, agent.ID) {
					require.Equal(t, newDate.Unix(), agent.ReportedAt.Unix())
					require.Equal(t, model.Connected, agent.Status)
				} else {
					require.Equal(t, oldDate.Unix(), agent.ReportedAt.Unix())
					require.Equal(t, model.Disconnected, agent.Status)
				}
			}
		})
	}
}
