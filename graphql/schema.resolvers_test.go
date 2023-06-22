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

package graphql

import (
	"context"
	"errors"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/observiq/bindplane-op/agent"
	agentMocks "github.com/observiq/bindplane-op/agent/mocks"
	"github.com/observiq/bindplane-op/config"
	sourceMocks "github.com/observiq/bindplane-op/eventbus/mocks"
	model1 "github.com/observiq/bindplane-op/graphql/model"
	"github.com/observiq/bindplane-op/internal/server"
	"github.com/observiq/bindplane-op/model"
	"github.com/observiq/bindplane-op/store"
	"github.com/observiq/bindplane-op/store/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func addAgent(s store.Store, agent *model.Agent) (*model.Agent, error) {
	_, err := s.UpsertAgent(context.TODO(), agent.ID, func(a *model.Agent) {
		*a = *agent
	})
	return agent, err
}

const mockLatestVersion = "v1.5.0"

func mockVersions() agent.Versions {
	v := &agentMocks.MockVersions{}
	v.On("LatestVersionString", mock.Anything).Return(mockLatestVersion)
	return v
}

func TestUpgradeAvailable(t *testing.T) {
	stringPointer := func(s string) *string { return &s }

	ctx := context.Background()
	testCases := []struct {
		name               string
		latestVersion      *model.AgentVersion
		upgradeableVersion *string
	}{
		{
			name: "upgrade available",
			latestVersion: &model.AgentVersion{
				Spec: model.AgentVersionSpec{
					Version: "1.6.3",
				},
			},
			upgradeableVersion: stringPointer("1.6.3"),
		},
		{
			name: "no upgrade available",
			latestVersion: &model.AgentVersion{
				Spec: model.AgentVersionSpec{
					Version: "1.5.0",
				},
			},
			upgradeableVersion: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			agentVersions := agentMocks.NewMockVersions(t)
			agentVersions.On("LatestVersion", ctx).Return(tc.latestVersion, nil)
			bindplane := server.NewBindPlane(&config.Config{},
				zaptest.NewLogger(t),
				store.NewMapStore(ctx, store.Options{}, zap.NewNop()),
				agentVersions,
			)

			resolver := agentResolver{&Resolver{
				Bindplane: bindplane,
			}}
			newVersion, err := resolver.UpgradeAvailable(ctx, &model.Agent{
				Version: "1.6.1",
			})

			require.NoError(t, err)
			if tc.upgradeableVersion != nil {
				require.Equal(t, *tc.upgradeableVersion, *newVersion)
			} else {
				require.Nil(t, newVersion)
			}
		})
	}
}

func TestQueryResolvers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mapstore := store.NewMapStore(ctx, store.Options{
		SessionsSecret:   "super-secret-key",
		MaxEventsToMerge: 1,
	}, zap.NewNop())

	bindplane := server.NewBindPlane(&config.Config{}, zaptest.NewLogger(t), mapstore, mockVersions())

	srv := NewHandler(bindplane)
	c := client.New(srv)

	s := bindplane.Store()

	t.Run("agents returns all Agents in the store", func(t *testing.T) {
		s.Clear()

		var resp map[string]model1.Agents
		var err error

		// Shouldn't get any Agents before adding to the store
		err = c.Post(`query TestQuery { agents(selector: "") { agents { id } } }`, &resp)
		require.NoError(t, err)
		require.Len(t, resp["agents"].Agents, 0)

		xy, err := model.LabelsFromSelector("x=y")
		require.NoError(t, err)

		addAgent(s, &model.Agent{ID: "1", Name: "Fake Agent 1", Labels: xy})
		addAgent(s, &model.Agent{ID: "2", Name: "Fake Agent 2"})

		// Should get the two Agents back that we added
		err = c.Post(`query TestQuery { agents(selector: "") { agents { id } } }`, &resp)
		require.NoError(t, err)
		require.Len(t, resp["agents"].Agents, 2)

		// Should get the one Agent back that matches the selector
		err = c.Post(`query TestQuery { agents(selector: "x=y") { agents { id } } }`, &resp)
		require.NoError(t, err)
		require.Len(t, resp["agents"].Agents, 1)
	})

	t.Run("agent loads a specific Agent by ID", func(t *testing.T) {
		s.Clear()

		var resp map[string]*model.Agent
		var err error

		addAgent(s, &model.Agent{ID: "1", Name: "Fake Agent 1"})
		agent, err := addAgent(s, &model.Agent{ID: "2", Name: "Fake Agent 2"})
		require.NoError(t, err)

		err = c.Post("query TestQuery($id: ID!) { agent(id: $id) { id } }", &resp, client.Var("id", "2"))
		require.NoError(t, err)
		require.Equal(t, resp["agent"].ID, agent.ID)
	})
}

func TestConfigForAgent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mapstore := store.NewMapStore(ctx, store.Options{
		SessionsSecret:   "super-secret-key",
		MaxEventsToMerge: 1,
	}, zap.NewNop())

	bindplane := server.NewBindPlane(&config.Config{}, zaptest.NewLogger(t), mapstore, mockVersions())

	srv := NewHandler(bindplane)
	c := client.New(srv)

	store := bindplane.Store()

	// SETUP
	labels := map[string]string{"env": "test", "app": "bindplane"}
	agent1Labels := model.Labels{Set: labels}

	otherLabels := map[string]string{"foo": "bar"}
	agent2labels := model.Labels{Set: otherLabels}

	addAgent(store, &model.Agent{ID: "1", Labels: agent1Labels})
	addAgent(store, &model.Agent{ID: "2", Labels: agent2labels})

	configLabels, _ := model.LabelsFromMap(map[string]string{"platform": "linux"})

	config := &model.Configuration{
		Spec: model.ConfigurationSpec{
			Raw:      "raw:",
			Selector: model.AgentSelector{MatchLabels: labels},
		},
		ResourceMeta: model.ResourceMeta{
			APIVersion: "",
			Kind:       "Configuration",
			Metadata: model.Metadata{
				Name:        "config",
				ID:          "config-123",
				Description: "should be used by agent 1",
				Labels:      configLabels,
			},
		},
	}

	_, err := bindplane.Store().ApplyResources(ctx, []model.Resource{config})
	require.NoError(t, err)

	resp := &struct {
		Agents struct {
			Agents []struct {
				ID                    string
				Name                  string
				ConfigurationResource *struct {
					Metadata struct {
						Name string
					}
				}
			}
			LatestVersion string
		}
	}{}

	agentsQuery := `
	query TestAgents {
		agents {
			agents {
				id
				name
				configurationResource {
					metadata {
						name
					}
				}
			}
			latestVersion
		}
	}
`

	err = c.Post(agentsQuery, &resp)
	require.NoError(t, err)

	for _, agent := range resp.Agents.Agents {
		switch agent.ID {
		case "1":
			require.Equal(t, "config", agent.ConfigurationResource.Metadata.Name)
		case "2":
			require.Nil(t, agent.ConfigurationResource)
		}
	}

	require.Equal(t, mockLatestVersion, resp.Agents.LatestVersion)
}

type testStore struct {
	store.Store
	store.ArchiveStore
}

func Test_queryResolver_ConfigurationHistory(t *testing.T) {
	updates := sourceMocks.NewMockSource[store.BasicEventUpdates](t)
	updates.On("Subscribe", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	tests := []struct {
		name           string
		store          func(t *testing.T) any
		want           []*model.Configuration
		wantErr        bool
		wantErrMessage string
	}{
		{
			"not an archive store",
			func(t *testing.T) any {
				s := mocks.NewMockStore(t)
				s.On("Updates", mock.Anything).Return(updates)
				return s
			},
			nil,
			true,
			"cannot get configuration history from non-archive store",
		},
		{
			"ResourceHistory fails",
			func(t *testing.T) any {
				s := &mocks.MockStore{}
				s.On("Updates", mock.Anything).Return(updates)

				as := mocks.NewMockArchiveStore(t)
				as.On("ResourceHistory", mock.Anything, model.KindConfiguration, "name").Return(nil, errors.New("error"))
				store := &testStore{
					Store:        s,
					ArchiveStore: as,
				}
				return store
			},
			nil,
			true,
			"configurationHistory resolver, archive: error",
		},
		{
			"error parsing",
			func(t *testing.T) any {
				s := &mocks.MockStore{}
				s.On("Updates", mock.Anything).Return(updates)

				as := mocks.NewMockArchiveStore(t)
				as.On("ResourceHistory", mock.Anything, model.KindConfiguration, "name").Return(
					[]*model.AnyResource{
						{
							ResourceMeta: model.ResourceMeta{
								Kind:     model.KindUnknown,
								Metadata: model.Metadata{},
							},
						},
					}, nil)
				store := &testStore{
					Store:        s,
					ArchiveStore: as,
				}
				return store
			},
			nil,
			true,
			"configurationHistory resolver, parsing history: unknown resource kind: Unknown",
		},
		{
			"error not a configuration",
			func(t *testing.T) any {
				s := &mocks.MockStore{}
				s.On("Updates", mock.Anything).Return(updates)

				as := mocks.NewMockArchiveStore(t)
				as.On("ResourceHistory", mock.Anything, model.KindConfiguration, "name").Return(
					[]*model.AnyResource{
						{
							ResourceMeta: model.ResourceMeta{
								Kind:     model.KindDestination,
								Metadata: model.Metadata{},
							},
						},
					}, nil)
				store := &testStore{
					Store:        s,
					ArchiveStore: as,
				}
				return store
			},
			nil,
			true,
			"configurationHistory resolver, parsing history: resource of kind Destination is not the expected type",
		},
		{
			"returns configurations",
			func(t *testing.T) any {
				s := &mocks.MockStore{}
				s.On("Updates", mock.Anything).Return(updates)

				as := mocks.NewMockArchiveStore(t)
				as.On("ResourceHistory", mock.Anything, model.KindConfiguration, "name").Return(
					[]*model.AnyResource{
						{
							ResourceMeta: model.ResourceMeta{
								Kind: model.KindConfiguration,
								Metadata: model.Metadata{
									Version: 1,
								},
							},
						},
						{
							ResourceMeta: model.ResourceMeta{
								Kind: model.KindConfiguration,
								Metadata: model.Metadata{
									Version: 2,
								},
							},
						},
					}, nil)
				store := &testStore{
					Store:        s,
					ArchiveStore: as,
				}
				return store
			},
			[]*model.Configuration{
				{
					ResourceMeta: model.ResourceMeta{
						Kind: model.KindConfiguration,
						Metadata: model.Metadata{
							Version: 1,
						},
					},
					Spec: model.ConfigurationSpec{},
				},
				{
					ResourceMeta: model.ResourceMeta{
						Kind: model.KindConfiguration,
						Metadata: model.Metadata{
							Version: 2,
						},
					},
					Spec: model.ConfigurationSpec{},
				},
			},
			false,
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bindplane := server.NewBindPlane(
				&config.Config{},
				zap.NewNop(),
				tt.store(t).(store.Store),
				mockVersions(),
			)

			resolver := NewResolver(bindplane)
			r := &queryResolver{
				Resolver: resolver,
			}
			got, err := r.ConfigurationHistory(context.Background(), "name")

			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.wantErrMessage, err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func Test_mutationResolver_ClearAgentUpgradeError(t *testing.T) {
	tests := []struct {
		name    string
		store   func(t *testing.T) store.Store
		input   *model1.ClearAgentUpgradeErrorInput
		wantErr bool
	}{
		{
			"error when upsert fails",
			func(t *testing.T) store.Store {
				s := mocks.NewMockStore(t)
				s.On("UpsertAgent", mock.Anything, "1", mock.AnythingOfType("store.AgentUpdater")).Return(nil, errors.New("error"))
				return s
			},
			&model1.ClearAgentUpgradeErrorInput{
				AgentID: "1",
			},
			true,
		},
		{
			"upsert succeeds",
			func(t *testing.T) store.Store {
				s := mocks.NewMockStore(t)
				s.On("UpsertAgent", mock.Anything, "1", mock.AnythingOfType("store.AgentUpdater")).Return(&model.Agent{}, nil)
				return s
			},
			&model1.ClearAgentUpgradeErrorInput{
				AgentID: "1",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bindplane := server.NewBindPlane(&config.Config{}, zaptest.NewLogger(t), tt.store(t), mockVersions())
			resolver := &Resolver{Bindplane: bindplane}
			r := &mutationResolver{
				Resolver: resolver,
			}

			_, err := r.ClearAgentUpgradeError(context.Background(), *tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("mutationResolver.ClearAgentUpgradeError() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_mutationResolver_EditConfigurationDescription(t *testing.T) {
	configName := "config-name"
	storeErr := errors.New("store error")

	updates := sourceMocks.NewMockSource[store.BasicEventUpdates](t)
	updates.On("Subscribe", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	tests := []struct {
		name    string
		store   func(t *testing.T) store.Store
		wantErr error
	}{
		{
			"error when UpdateConfiguration errors",
			func(t *testing.T) store.Store {
				s := mocks.NewMockStore(t)
				s.On("Updates", mock.Anything).Return(updates)
				s.On("UpdateConfiguration", mock.Anything, configName, mock.Anything).Return(nil, model.StatusConfigured, storeErr)
				return s
			},
			storeErr,
		},
		{
			"success",
			func(t *testing.T) store.Store {
				s := mocks.NewMockStore(t)
				s.On("Updates", mock.Anything).Return(updates)
				s.On("UpdateConfiguration", mock.Anything, configName, mock.Anything).Return(&model.Configuration{}, model.StatusConfigured, nil)

				return s
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bindplane := server.NewBindPlane(
				&config.Config{},
				zap.NewNop(),
				tt.store(t),
				mockVersions(),
			)

			resolver := NewResolver(bindplane)
			r := &mutationResolver{
				Resolver: resolver,
			}
			_, err := r.EditConfigurationDescription(context.Background(), model1.EditConfigurationDescriptionInput{
				Name:        configName,
				Description: "new-description",
			})

			if tt.wantErr != nil {
				require.ErrorAs(t, err, &tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}