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

func Test_queryResolver_ConfigurationHistory(t *testing.T) {
	updates := sourceMocks.NewMockSource[store.BasicEventUpdates](t)
	updates.On("Subscribe", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	tests := []struct {
		name           string
		store          func(t *testing.T) store.Store
		want           []*model.Configuration
		wantErr        bool
		wantErrMessage string
	}{
		{
			"ResourceHistory fails",
			func(t *testing.T) store.Store {
				s := mocks.NewMockStore(t)
				s.On("Updates", mock.Anything).Return(updates)

				s.On("ResourceHistory", mock.Anything, model.KindConfiguration, "name").Return(nil, errors.New("error"))
				return s
			},
			nil,
			true,
			"configurationHistory resolver, archive: error",
		},
		{
			"error parsing",
			func(t *testing.T) store.Store {
				s := mocks.NewMockStore(t)
				s.On("Updates", mock.Anything).Return(updates)

				s.On("ResourceHistory", mock.Anything, model.KindConfiguration, "name").Return(
					[]*model.AnyResource{
						{
							ResourceMeta: model.ResourceMeta{
								Kind:     model.KindUnknown,
								Metadata: model.Metadata{},
							},
						},
					}, nil)
				return s
			},
			nil,
			true,
			"configurationHistory resolver, parsing history: unknown resource kind: Unknown",
		},
		{
			"error not a configuration",
			func(t *testing.T) store.Store {
				s := mocks.NewMockStore(t)
				s.On("Updates", mock.Anything).Return(updates)

				s.On("ResourceHistory", mock.Anything, model.KindConfiguration, "name").Return(
					[]*model.AnyResource{
						{
							ResourceMeta: model.ResourceMeta{
								Kind:     model.KindDestination,
								Metadata: model.Metadata{},
							},
						},
					}, nil)
				return s
			},
			nil,
			true,
			"configurationHistory resolver, parsing history: resource of kind Destination is not the expected type",
		},
		{
			"returns configurations",
			func(t *testing.T) store.Store {
				s := mocks.NewMockStore(t)
				s.On("Updates", mock.Anything).Return(updates)

				s.On("ResourceHistory", mock.Anything, model.KindConfiguration, "name").Return(
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
				return s
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
				tt.store(t),
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

func Test_queryResolver_Destination(t *testing.T) {
	updates := sourceMocks.NewMockSource[store.BasicEventUpdates](t)
	updates.On("Subscribe", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	destinations := []*model.Destination{
		{
			ResourceMeta: model.ResourceMeta{
				Kind: model.KindDestination,
				Metadata: model.Metadata{
					Name: "dest_1",
				},
			},
		},
		{
			ResourceMeta: model.ResourceMeta{
				Kind: model.KindDestination,
				Metadata: model.Metadata{
					Name: "dest_2",
				},
			},
		},
	}

	// setup defaultStore:
	defaultStore := &mocks.MockStore{}
	defaultStore.On("Updates", mock.Anything).Return(updates)

	defaultStore.On("Destinations", mock.Anything).Return(
		destinations, nil)
	defaultStore.On("AgentsIDsMatchingConfiguration", mock.Anything, mock.Anything).Return([]string{"agent_id"}, nil)
	for _, d := range destinations {
		defaultStore.On("Destination", mock.Anything, d.Metadata.Name).Return(d, nil)
	}
	defaultStore.On("Configurations", mock.Anything).Return([]*model.Configuration{
		{
			ResourceMeta: model.ResourceMeta{
				Kind: model.KindConfiguration,
				Metadata: model.Metadata{
					Name: "config_1",
				},
			},
			Spec: model.ConfigurationSpec{
				Destinations: []model.ResourceConfiguration{
					{
						Name: "dest_1",
						ID:   "dest_1",
					},
				},
			},
		},
	}, nil)
	pointerToString := func(s string) *string {
		return &s
	}
	pointerToBool := func(b bool) *bool {
		return &b
	}

	type args struct {
		query         *string
		onlyInConfigs *bool
	}
	tests := []struct {
		name           string
		store          func(t *testing.T) any
		args           args
		want           []*model.Destination
		wantErr        bool
		wantErrMessage string
	}{

		{
			"Destinations fails",
			func(t *testing.T) any {
				store := mocks.NewMockStore(t)
				store.On("Updates", mock.Anything).Return(updates)
				store.On("Destinations", mock.Anything).Return(nil, errors.New("error"))

				return store
			},
			args{
				query:         nil,
				onlyInConfigs: nil,
			},
			nil,
			true,
			"queryResolver.Destinations failed to get Destinations from store\nerror",
		},

		{
			"returns destinations",
			func(t *testing.T) any {

				return defaultStore
			},
			args{
				query:         nil,
				onlyInConfigs: nil,
			},
			destinations,
			false,
			"",
		},
		{
			"returns destinations with query",
			func(t *testing.T) any {

				return defaultStore
			},
			args{
				query:         pointerToString("dest"),
				onlyInConfigs: nil,
			},
			destinations,
			false,
			"",
		},
		{
			"returns no destinations with query",
			func(t *testing.T) any {
				return defaultStore
			},
			args{
				query:         pointerToString("xxxxxxx"),
				onlyInConfigs: nil,
			},
			[]*model.Destination{},
			false,
			"",
		},
		{
			"returns one destination with query",
			func(t *testing.T) any {
				return defaultStore
			},
			args{
				query:         pointerToString("dest_2"),
				onlyInConfigs: nil,
			},
			[]*model.Destination{
				{
					ResourceMeta: model.ResourceMeta{
						Kind: model.KindDestination,
						Metadata: model.Metadata{
							Name: "dest_2",
						},
					},
				},
			},
			false,
			"",
		},
		{
			"handles weird characters in query",
			func(t *testing.T) any {
				return defaultStore
			},
			args{
				query:         pointerToString(`3%'	2\\]\;" 	$|@>!#<^&*()_+{}[]:;?/.,~\ + "'`),
				onlyInConfigs: nil,
			},
			[]*model.Destination{},
			false,
			"",
		},
		{
			"returns destinations with onlyInConfigs",
			func(t *testing.T) any {
				return defaultStore
			},
			args{
				query:         nil,
				onlyInConfigs: pointerToBool(true),
			},
			[]*model.Destination{{
				ResourceMeta: model.ResourceMeta{
					Kind: model.KindDestination,
					Metadata: model.Metadata{
						Name: "dest_1",
					},
				},
			}},
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
			got, err := r.Destinations(context.Background(), tt.args.query, tt.args.onlyInConfigs)

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

func Test_queryResolver_DestinationWithType(t *testing.T) {
	destinationErr := errors.New("destination error")
	destinationTypeErr := errors.New("destination type error")

	destinationName := "destination-name"
	destinationTypeName := "custom"

	destination := model.NewDestination(destinationName, destinationTypeName, nil)
	destinationType := model.NewDestinationType(destinationTypeName, nil)

	updates := sourceMocks.NewMockSource[store.BasicEventUpdates](t)
	updates.On("Subscribe", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	tests := []struct {
		name      string
		store     func(t *testing.T) any
		expect    *model1.DestinationWithType
		expectErr error
	}{
		{
			name: "error on Destination error",
			store: func(t *testing.T) any {
				store := mocks.NewMockStore(t)
				store.EXPECT().Updates(mock.Anything).Return(updates)
				store.EXPECT().Destination(mock.Anything, destinationName).Return(nil, destinationErr)

				return store
			},
			expect:    &model1.DestinationWithType{},
			expectErr: destinationErr,
		},
		{
			name: "destination not found",
			store: func(t *testing.T) any {
				store := mocks.NewMockStore(t)
				store.EXPECT().Updates(mock.Anything).Return(updates)
				store.EXPECT().Destination(mock.Anything, destinationName).Return(nil, nil)

				return store
			},
			expect:    &model1.DestinationWithType{},
			expectErr: nil,
		},
		{
			name: "error on DestinationType error",
			store: func(t *testing.T) any {
				store := mocks.NewMockStore(t)
				store.EXPECT().Updates(mock.Anything).Return(updates)
				store.EXPECT().Destination(mock.Anything, destinationName).Return(destination, nil)
				store.EXPECT().DestinationType(mock.Anything, destinationTypeName).Return(nil, destinationTypeErr)

				return store
			},
			expect:    &model1.DestinationWithType{},
			expectErr: destinationTypeErr,
		},
		{
			name: "destination type not found",
			store: func(t *testing.T) any {
				store := mocks.NewMockStore(t)
				store.EXPECT().Updates(mock.Anything).Return(updates)
				store.EXPECT().Destination(mock.Anything, destinationName).Return(destination, nil)
				store.EXPECT().DestinationType(mock.Anything, destinationTypeName).Return(nil, nil)

				return store
			},
			expect:    &model1.DestinationWithType{},
			expectErr: nil,
		},
		{
			name: "success",
			store: func(t *testing.T) any {
				store := mocks.NewMockStore(t)
				store.EXPECT().Updates(mock.Anything).Return(updates)
				store.EXPECT().Destination(mock.Anything, destinationName).Return(destination, nil)
				store.EXPECT().DestinationType(mock.Anything, destinationTypeName).Return(destinationType, nil)

				return store
			},
			expect: &model1.DestinationWithType{
				Destination:     destination,
				DestinationType: destinationType,
			},
			expectErr: nil,
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
			got, err := r.DestinationWithType(context.Background(), destinationName)

			if tt.expectErr != nil {
				require.Error(t, err)
				require.Equal(t, tt.expectErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.expect, got)
		})
	}

}

func Test_queryResolver_SourceWithType(t *testing.T) {
	sourceErr := errors.New("source error")
	sourceTypeErr := errors.New("source type error")

	sourceName := "source-name"
	sourceTypeName := "host"

	source := model.NewSource(sourceName, sourceTypeName, nil)
	sourceType := model.NewSourceType(sourceTypeName, nil, nil)

	updates := sourceMocks.NewMockSource[store.BasicEventUpdates](t)
	updates.On("Subscribe", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	tests := []struct {
		name      string
		store     func(t *testing.T) any
		expect    *model1.SourceWithType
		expectErr error
	}{
		{
			name: "error on Source error",
			store: func(t *testing.T) any {
				store := mocks.NewMockStore(t)
				store.EXPECT().Updates(mock.Anything).Return(updates)
				store.EXPECT().Source(mock.Anything, sourceName).Return(nil, sourceErr)

				return store
			},
			expect:    &model1.SourceWithType{},
			expectErr: sourceErr,
		},
		{
			name: "source not found",
			store: func(t *testing.T) any {
				store := mocks.NewMockStore(t)
				store.EXPECT().Updates(mock.Anything).Return(updates)
				store.EXPECT().Source(mock.Anything, sourceName).Return(nil, nil)

				return store
			},
			expect:    &model1.SourceWithType{},
			expectErr: nil,
		},
		{
			name: "error on SourceType error",
			store: func(t *testing.T) any {
				store := mocks.NewMockStore(t)
				store.EXPECT().Updates(mock.Anything).Return(updates)
				store.EXPECT().Source(mock.Anything, sourceName).Return(source, nil)
				store.EXPECT().SourceType(mock.Anything, sourceTypeName).Return(nil, sourceTypeErr)

				return store
			},
			expect:    &model1.SourceWithType{},
			expectErr: sourceTypeErr,
		},
		{
			name: "source type not found",
			store: func(t *testing.T) any {
				store := mocks.NewMockStore(t)
				store.EXPECT().Updates(mock.Anything).Return(updates)
				store.EXPECT().Source(mock.Anything, sourceName).Return(source, nil)
				store.EXPECT().SourceType(mock.Anything, sourceTypeName).Return(nil, nil)

				return store
			},
			expect:    &model1.SourceWithType{},
			expectErr: nil,
		},
		{
			name: "success",
			store: func(t *testing.T) any {
				store := mocks.NewMockStore(t)
				store.EXPECT().Updates(mock.Anything).Return(updates)
				store.EXPECT().Source(mock.Anything, sourceName).Return(source, nil)
				store.EXPECT().SourceType(mock.Anything, sourceTypeName).Return(sourceType, nil)

				return store
			},
			expect: &model1.SourceWithType{
				Source:     source,
				SourceType: sourceType,
			},
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
			got, err := r.SourceWithType(context.Background(), sourceName)

			if tt.expectErr != nil {
				require.Error(t, err)
				require.Equal(t, tt.expectErr, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.expect, got)
		})
	}
}

func Test_queryResolver_ProcessorWithType(t *testing.T) {
	processorErr := errors.New("processor error")
	processorTypeErr := errors.New("processor type error")

	processorName := "processor-name"
	processorTypeName := "batch"

	processor := model.NewProcessor(processorName, processorTypeName, nil)
	processorType := model.NewProcessorType(processorTypeName, nil)

	updates := sourceMocks.NewMockSource[store.BasicEventUpdates](t)
	updates.On("Subscribe", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	tests := []struct {
		name      string
		store     func(t *testing.T) any
		expect    *model1.ProcessorWithType
		expectErr error
	}{
		{
			name: "error on Processor error",
			store: func(t *testing.T) any {
				store := mocks.NewMockStore(t)
				store.EXPECT().Updates(mock.Anything).Return(updates)
				store.EXPECT().Processor(mock.Anything, processorName).Return(nil, processorErr)

				return store
			},
			expect:    &model1.ProcessorWithType{},
			expectErr: processorErr,
		},
		{
			name: "processor not found",
			store: func(t *testing.T) any {
				store := mocks.NewMockStore(t)
				store.EXPECT().Updates(mock.Anything).Return(updates)
				store.EXPECT().Processor(mock.Anything, processorName).Return(nil, nil)

				return store
			},
			expect:    &model1.ProcessorWithType{},
			expectErr: nil,
		},
		{
			name: "error on ProcessorType error",
			store: func(t *testing.T) any {
				store := mocks.NewMockStore(t)
				store.EXPECT().Updates(mock.Anything).Return(updates)
				store.EXPECT().Processor(mock.Anything, processorName).Return(processor, nil)
				store.EXPECT().ProcessorType(mock.Anything, processorTypeName).Return(nil, processorTypeErr)

				return store
			},
			expect:    &model1.ProcessorWithType{},
			expectErr: processorTypeErr,
		},
		{
			name: "source type not found",
			store: func(t *testing.T) any {
				store := mocks.NewMockStore(t)
				store.EXPECT().Updates(mock.Anything).Return(updates)
				store.EXPECT().Processor(mock.Anything, processorName).Return(processor, nil)
				store.EXPECT().ProcessorType(mock.Anything, processorTypeName).Return(nil, nil)

				return store
			},
			expect:    &model1.ProcessorWithType{},
			expectErr: nil,
		},
		{
			name: "success",
			store: func(t *testing.T) any {
				store := mocks.NewMockStore(t)
				store.EXPECT().Updates(mock.Anything).Return(updates)
				store.EXPECT().Processor(mock.Anything, processorName).Return(processor, nil)
				store.EXPECT().ProcessorType(mock.Anything, processorTypeName).Return(processorType, nil)

				return store
			},
			expect: &model1.ProcessorWithType{
				Processor:     processor,
				ProcessorType: processorType,
			},
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
			got, err := r.ProcessorWithType(context.Background(), processorName)

			if tt.expectErr != nil {
				require.Error(t, err)
				require.Equal(t, tt.expectErr, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.expect, got)
		})
	}
}
