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

package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/observiq/bindplane-op/config"
	"github.com/observiq/bindplane-op/internal/server"
	"github.com/observiq/bindplane-op/model"
	"github.com/observiq/bindplane-op/model/version"
	"github.com/observiq/bindplane-op/store"
	storeMocks "github.com/observiq/bindplane-op/store/mocks"
	statsmocks "github.com/observiq/bindplane-op/store/stats/mocks"
	"github.com/observiq/bindplane-op/store/storetest"
	"github.com/observiq/bindplane-op/util"
)

func resetStore(t *testing.T, store store.Store) {
	store.Clear()
	// add the sourceTypes
	macos := model.NewSourceType("macos", []model.ParameterDefinition{
		{
			Name: "version",
			Type: "string",
		},
		{
			Name:        "start_at",
			Type:        "enum",
			ValidValues: []string{"end", "beginning"},
		},
	}, []string{"macos"})
	nginx := model.NewSourceType("nginx", []model.ParameterDefinition{
		{
			Name: "log_format",
			Type: "string",
		},
	}, []string{"macos", "linux", "windows"})
	cabin := model.NewDestinationType("cabin", []model.ParameterDefinition{
		{
			Name: "endpoint",
			Type: "string",
		},
		{
			Name: "api_key",
			Type: "string",
		},
		{
			Name: "timeout",
			Type: "string",
		},
	})
	_, err := store.ApplyResources(context.Background(), []model.Resource{macos, nginx, cabin})
	require.NoError(t, err)
}

func addAgent(s store.Store, agent *model.Agent) (*model.Agent, error) {
	_, err := s.UpsertAgent(context.TODO(), agent.ID, func(a *model.Agent) {
		*a = *agent
	})
	return agent, err
}

func testDestination(name string, destinationType string) *model.Destination {
	return testDestinationWithParameters(name, destinationType, nil)
}

func testDestinationWithParameters(name string, destinationType string, parameters []model.Parameter) *model.Destination {
	return model.NewDestination(name, destinationType, parameters)
}

func testDestinationAsAny(t *testing.T, name string, destinationType string) *model.AnyResource {
	destination := testDestination(name, destinationType)
	spec := make(map[string]interface{})
	if err := mapstructure.Decode(destination.Spec, &spec); err != nil {
		require.NoError(t, err, "expect no error when setting up tests")
	}

	return &model.AnyResource{
		ResourceMeta: destination.ResourceMeta,
		Spec:         spec,
	}
}

func testRawConfiguration(id, name string) *model.Configuration {
	return &model.Configuration{
		ResourceMeta: model.ResourceMeta{
			Metadata: model.Metadata{
				Name:   name,
				ID:     id,
				Labels: model.MakeLabels(),
			},
			Kind: model.KindConfiguration,
		},
		Spec: model.ConfigurationSpec{
			Raw: "raw:",
		},
	}
}

func testSource(name string, sourceType string) *model.Source {
	return testSourceWithParameters(name, sourceType, nil)
}

func testSourceWithParameters(name string, sourceType string, parameters []model.Parameter) *model.Source {
	return model.NewSource(name, sourceType, parameters)
}

func testSourceAsAny(t *testing.T, name string, sourceType string) *model.AnyResource {
	source := testSource(name, sourceType)
	spec := make(map[string]interface{})
	if err := mapstructure.Decode(source.Spec, &spec); err != nil {
		require.NoError(t, err, "expect no error when setting up tests")
	}

	return &model.AnyResource{
		ResourceMeta: source.ResourceMeta,
		Spec:         spec,
	}
}

func testConfiguration(name string) *model.Configuration {
	return model.NewConfigurationWithSpec(name, model.ConfigurationSpec{
		Sources: []model.ResourceConfiguration{
			{ParameterizedSpec: model.ParameterizedSpec{Type: "macos"}},
			{ParameterizedSpec: model.ParameterizedSpec{Type: "macos"}},
		},
		Destinations: []model.ResourceConfiguration{
			{ParameterizedSpec: model.ParameterizedSpec{Type: "cabin"}},
		},
		Selector: model.AgentSelector{
			MatchLabels: map[string]string{
				"env": "production",
				"app": "bindplane-prod",
			},
		},
	})
}

func testConfigurationAsAny(t *testing.T, id string, name string) *model.AnyResource {
	configuration := testRawConfiguration(id, name)
	spec := make(map[string]interface{})
	if err := mapstructure.Decode(configuration.Spec, &spec); err != nil {
		require.NoError(t, err, "expect no error when setting up tests")
	}

	return &model.AnyResource{
		ResourceMeta: configuration.ResourceMeta,
		Spec:         spec,
	}
}

func TestREST(t *testing.T) {
	router := gin.Default()
	svr := httptest.NewServer(router)
	defer svr.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// TODO find a better way to setup a test bbolt database outside the store package.
	db, err := storetest.InitTestBboltDB(t, []string{
		store.BucketResources,
		store.BucketAgents,
		store.BucketMeasurements,
		store.BucketArchive,
	})
	require.NoError(t, err)

	storeOpts := store.Options{
		SessionsSecret:   "super-secret-key",
		MaxEventsToMerge: 1,
	}

	store := store.NewBoltStore(ctx, db, storeOpts, zap.NewNop())

	mockBatcher := statsmocks.NewMockMeasurementBatcher(t)
	bindplane := server.NewBindPlane(&config.Config{}, zaptest.NewLogger(t), store, nil, mockBatcher)
	AddRestRoutes(router, bindplane)

	client := resty.New()
	client.SetBaseURL(svr.URL)

	s := bindplane.Store()

	seq := util.NewTestSequence(t)
	seq.Run(`Setup: Test PUT rollouts|{name}|pause and rollouts|{name}|resume`, func(t *testing.T) {
		resetStore(t, s)
		config1 := &model.Configuration{
			ResourceMeta: model.ResourceMeta{
				APIVersion: "",
				Kind:       "Configuration",
				Metadata: model.Metadata{
					Name:        "config1",
					ID:          "config-1",
					Description: "should be used by agent 1",
					Labels:      model.LabelsFromValidatedMap(map[string]string{"platform": "linux"}),
				},
			},
			StatusType: model.StatusType[model.ConfigurationStatus]{
				Status: model.ConfigurationStatus{
					Rollout: model.Rollout{
						Status:  model.RolloutStatusPending,
						Options: model.RolloutOptions{},
						Phase:   0,
						Progress: model.RolloutProgress{
							Completed: 0,
							Errors:    0,
							Waiting:   0,
						},
					},
					CurrentVersion: 0,
				}},
		}
		_, err := bindplane.Store().ApplyResources(ctx, []model.Resource{config1})
		require.NoError(t, err)
		_, err = bindplane.Store().UpsertAgents(ctx, []string{"a1", "a2", "a3", "a4", "a5"}, func(a *model.Agent) {
			a.SetFutureConfiguration(config1)
			a.Status = model.Connected
		})
		require.NoError(t, err)
	})

	seq.Run("|rollouts|config1|start produces error for missing config", func(t *testing.T) {
		result := &model.ConfigurationResponse{}
		payload := &model.StartRolloutPayload{}
		resp, err := client.R().SetBody(payload).SetResult(result).Post("/rollouts/missingname/start")
		require.Equal(t, http.StatusNotFound, resp.StatusCode())
		require.NoError(t, err)
	})
	seq.Run("|rollouts|config1|start starts rollout for config1", func(t *testing.T) {
		result := &model.ConfigurationResponse{}
		payload := &model.StartRolloutPayload{}

		resp, err := client.R().SetBody(payload).SetResult(result).Post("/rollouts/config1/start")
		require.True(t, resp.IsSuccess())
		require.NoError(t, err)
		assert.Equal(t, model.RolloutStatusStarted, result.Configuration.Status.Rollout.Status)
	})
	seq.Run("|rollouts|config1|pause pauses rollout for config1", func(t *testing.T) {
		result := &model.ConfigurationResponse{}
		resp, err := client.R().SetResult(result).Put("/rollouts/config1/pause")
		require.True(t, resp.IsSuccess())
		require.NoError(t, err)
		assert.Equal(t, model.RolloutStatusPaused, result.Configuration.Status.Rollout.Status)
	})
	seq.Run("|rollouts|config1|resume resumes rollout for config1", func(t *testing.T) {
		result := &model.ConfigurationResponse{}
		resp, err := client.R().SetResult(result).Put("/rollouts/config1/resume")
		require.True(t, resp.IsSuccess())
		require.NoError(t, err)
		assert.Equal(t, model.RolloutStatusStarted, result.Configuration.Status.Rollout.Status)
	})
	seq.Run("|rollouts|config1|resume produces error for missing config", func(t *testing.T) {
		result := &model.ConfigurationResponse{}
		resp, err := client.R().SetResult(result).Put("/rollouts/config14234/resume")
		require.Equal(t, http.StatusNotFound, resp.StatusCode())
		require.NoError(t, err)
	})
	seq.Run("|rollouts|config1|pause produces error for missing config", func(t *testing.T) {
		result := &model.ConfigurationResponse{}
		resp, err := client.R().SetResult(result).Put("/rollouts/config1324/pause")
		require.Equal(t, http.StatusNotFound, resp.StatusCode())
		require.NoError(t, err)
	})

	t.Run("Test GET |rollouts, |rollouts|{name}, and |rollouts|{name}|status", func(t *testing.T) {
		resetStore(t, s)

		endpoint := "/rollouts"
		rr := &model.ConfigurationsResponse{}

		getRequest(t, client, endpoint, rr)

		require.Len(t, rr.Configurations, 0)

		resetStore(t, bindplane.Store())

		config1 := &model.Configuration{
			ResourceMeta: model.ResourceMeta{
				APIVersion: "",
				Kind:       "Configuration",
				Metadata: model.Metadata{
					Name:        "config1",
					ID:          "config-1",
					Description: "should be used by agent 1",
					Labels:      model.LabelsFromValidatedMap(map[string]string{"platform": "linux"}),
				},
			},
			StatusType: model.StatusType[model.ConfigurationStatus]{
				Status: model.ConfigurationStatus{
					Rollout: model.Rollout{
						Status:  model.RolloutStatusPending,
						Options: model.RolloutOptions{},
						Phase:   0,
						Progress: model.RolloutProgress{
							Completed: 0,
							Errors:    0,
							Waiting:   0,
						},
					},
					CurrentVersion: 0,
				}},
		}
		config2 := &model.Configuration{
			ResourceMeta: model.ResourceMeta{
				APIVersion: "",
				Kind:       "Configuration",
				Metadata: model.Metadata{
					Name:        "config2",
					ID:          "config-2",
					Description: "should be used by agent 2",
					Labels:      model.LabelsFromValidatedMap(map[string]string{"platform": "linux"}),
				},
			},
			StatusType: model.StatusType[model.ConfigurationStatus]{
				Status: model.ConfigurationStatus{
					Rollout: model.Rollout{
						Status:  model.RolloutStatusStarted,
						Options: model.RolloutOptions{},
						Phase:   1,
						Progress: model.RolloutProgress{
							Completed: 2,
							Errors:    3,
							Waiting:   4,
						},
					},
					CurrentVersion: 0,
				}},
		}
		resources, err := bindplane.Store().ApplyResources(ctx, []model.Resource{config1, config2})
		require.NoError(t, err)

		expectedConfigurations := make([]*model.Configuration, 0)
		for _, resource := range resources {
			expectedConfig := resource.Resource.(*model.Configuration)
			expectedConfig.SetDateModified(nil)
			expectedConfigurations = append(expectedConfigurations, expectedConfig)
		}

		t.Run("|rollouts returns all rollouts", func(t *testing.T) {
			result := &model.ConfigurationsResponse{}
			resp, err := client.R().SetResult(result).Get("/rollouts")
			require.True(t, resp.IsSuccess())
			require.NoError(t, err)
			for _, config := range result.Configurations {
				config.SetDateModified(nil)
			}
			assert.Equal(t, expectedConfigurations, result.Configurations)
		})

		t.Run("|rollouts|config2 returns rollout for config with name config2", func(t *testing.T) {
			result := &model.ConfigurationResponse{}
			resp, err := client.R().SetResult(result).Get("/rollouts/config2")
			require.True(t, resp.IsSuccess())
			require.NoError(t, err)
			assert.Equal(t, config2.Rollout(), result.Configuration.Rollout())
		})

		t.Run("|rollouts|config1 returns rollout for config with name config1", func(t *testing.T) {
			result := &model.ConfigurationResponse{}
			resp, err := client.R().SetResult(result).Get("/rollouts/config1")
			require.True(t, resp.IsSuccess())
			require.NoError(t, err)
			assert.Equal(t, config1.Rollout(), result.Configuration.Rollout())
		})

		t.Run("|rollouts|config2|status returns rollout status for config with name config2", func(t *testing.T) {
			result := &model.ConfigurationResponse{}
			resp, err := client.R().SetResult(result).Get("/rollouts/config2/status")
			require.True(t, resp.IsSuccess())
			require.NoError(t, err)
			assert.Equal(t, config2.Rollout().Status, result.Configuration.Status.Rollout.Status)
		})

	})

	t.Run("GET |agents returns all Agents in the store", func(t *testing.T) {
		resetStore(t, s)

		endpoint := "/agents"
		ar := &model.AgentsResponse{}

		getRequest(t, client, endpoint, ar)

		require.Len(t, ar.Agents, 0)

		agent1, err := addAgent(s, &model.Agent{ID: "1", Name: "Fake Agent 1", Labels: model.MakeLabels()})
		require.NoError(t, err)
		agent2, err := addAgent(s, &model.Agent{ID: "2", Name: "Fake Agent 2", Labels: model.MakeLabels()})
		require.NoError(t, err)

		getRequest(t, client, endpoint, ar)

		require.Len(t, ar.Agents, 2)
		require.ElementsMatch(t, ar.Agents, []*model.Agent{agent1, agent2})
	})

	t.Run("GET |agents|:id returns a specific Agent by ID", func(t *testing.T) {
		resetStore(t, s)

		_, err := addAgent(s, &model.Agent{ID: "1", Name: "Fake Agent 1", Labels: model.MakeLabels()})
		require.NoError(t, err)
		agent, err := addAgent(s, &model.Agent{ID: "2", Name: "Fake Agent 2", Labels: model.MakeLabels()})
		require.NoError(t, err)

		ar := &model.AgentResponse{}

		getRequest(t, client, "/agents/2", ar)

		require.Equal(t, ar.Agent, agent)
	})

	t.Run("GET |destinations returns all Destinations in the store", func(t *testing.T) {
		resetStore(t, s)

		endpoint := "/destinations"
		rr := &model.DestinationsResponse{}

		getRequest(t, client, endpoint, rr)

		require.Len(t, rr.Destinations, 0)

		destination1 := testDestinationWithParameters("destination-1", "cabin",
			[]model.Parameter{{Name: "endpoint", Value: "https://nozzle.app.observiq.com"}, {Name: "timeout", Value: "10s"}})
		destination2 := testDestinationWithParameters("destination-2", "cabin",
			[]model.Parameter{{Name: "api_key", Value: "asdf"}})

		_, err := s.ApplyResources(context.Background(), []model.Resource{destination1, destination2})
		require.NoError(t, err)

		getRequest(t, client, endpoint, rr)
		require.Len(t, rr.Destinations, 2)
		for _, destination := range rr.Destinations {
			destination.SetDateModified(nil)
		}

		destination1.SetDateModified(nil)
		destination2.SetDateModified(nil)
		require.ElementsMatch(t, rr.Destinations, []*model.Destination{destination1, destination2})
	})

	t.Run("GET /destinations/:name returns a specific Destination by name", func(t *testing.T) {
		resetStore(t, s)

		destination1 := testDestinationWithParameters("destination-1", "cabin",
			[]model.Parameter{{Name: "endpoint", Value: "https://nozzle.app.observiq.com"}, {Name: "timeout", Value: "10s"}})
		destination2 := testDestinationWithParameters("destination-2", "cabin",
			[]model.Parameter{{Name: "api_key", Value: "asdf"}})

		_, err := s.ApplyResources(ctx, []model.Resource{destination1, destination2})
		require.NoError(t, err)

		rr := &model.DestinationResponse{}

		getRequest(t, client, "/destinations/destination-2", rr)
		rr.Destination.SetDateModified(nil)
		destination2.SetDateModified(nil)

		require.Equal(t, rr.Destination, destination2)
	})

	t.Run("DELETE |agents 200", func(t *testing.T) {
		resetStore(t, s)
		addAgent(s, &model.Agent{ID: "1"})

		expectBody := &model.DeleteAgentsResponse{
			Agents: []*model.Agent{
				{ID: "1", Status: 5, Labels: model.MakeLabels()},
			},
		}

		gotBody := &model.DeleteAgentsResponse{}

		_, err := client.R().SetBody(model.DeleteAgentsPayload{
			IDs: []string{"1"},
		}).SetResult(gotBody).Delete("/agents")

		require.NoError(t, err)
		require.Equal(t, expectBody, gotBody)
	})

	t.Run("DELETE |agents 400", func(t *testing.T) {
		resetStore(t, s)

		resp, err := client.R().SetBody("malformed").Delete("/agents")
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode())
	})

	t.Run("DELETE |destinations|:name 404 Not Found", func(t *testing.T) {
		resetStore(t, s)

		deleteEndpoint := fmt.Sprintf("/destinations/%s", "does-not-exist")
		resp, err := client.R().Delete(deleteEndpoint)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode())
	})

	t.Run("DELETE |destinations|:name 204 No Content", func(t *testing.T) {
		resetStore(t, s)

		destination1 := testDestination("destination-1", "cabin")
		destination2 := testDestination("destination-2", "cabin")

		_, err := s.ApplyResources(ctx, []model.Resource{destination1, destination2})
		require.NoError(t, err)

		deleteEndpoint := fmt.Sprintf("/destinations/%s", url.PathEscape(destination1.Name()))
		resp, err := client.R().Delete(deleteEndpoint)
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, resp.StatusCode())

		destinations, err := s.Destinations(ctx)
		require.NoError(t, err)

		assert.NotContains(t, destinations, destination1)
	})

	t.Run("DELETE |destinations|:name 409 Conflict", func(t *testing.T) {
		resetStore(t, s)
		dest1 := testDestination(
			"dest-1",
			"cabin",
		)

		config := model.NewConfigurationWithSpec("test-config", model.ConfigurationSpec{
			Destinations: []model.ResourceConfiguration{{Name: "dest-1"}},
		})

		_, err := s.ApplyResources(ctx, []model.Resource{dest1, config})
		require.NoError(t, err)
		deleteEndpoint := fmt.Sprintf("/destinations/%s", url.PathEscape(dest1.Name()))
		resp, err := client.R().Delete(deleteEndpoint)
		require.NoError(t, err)
		require.Equal(t, http.StatusConflict, resp.StatusCode())

		body := &ErrorResponse{}
		err = jsoniter.Unmarshal(resp.Body(), body)
		require.NoError(t, err)

		expectBody := ErrorResponse{
			Errors: []string{"Dependent resources:\nConfiguration test-config\n"},
		}

		assert.ElementsMatch(t, expectBody.Errors, body.Errors)
	},
	)

	t.Run("GET |sources returns all Sources in the store", func(t *testing.T) {
		resetStore(t, s)

		endpoint := "/sources"
		rr := &model.SourcesResponse{}

		getRequest(t, client, endpoint, rr)

		require.Len(t, rr.Sources, 0)

		source1 := testSourceWithParameters(
			"source-1",
			"nginx",
			[]model.Parameter{{Name: "log_format", Value: "default"}},
		)
		source2 := testSourceWithParameters(
			"source-2",
			"macos",
			[]model.Parameter{{Name: "version", Value: "0.0.2"}, {Name: "start_at", Value: "end"}},
		)

		_, err := s.ApplyResources(ctx, []model.Resource{source1, source2})
		require.NoError(t, err)

		getRequest(t, client, endpoint, rr)

		require.Len(t, rr.Sources, 2)
		for _, source := range rr.Sources {
			source.SetDateModified(nil)
		}
		source1.SetDateModified(nil)
		source2.SetDateModified(nil)
		require.ElementsMatch(t, rr.Sources, []*model.Source{source1, source2})
	})

	t.Run("GET |sources|:name returns a specific Source by name", func(t *testing.T) {
		resetStore(t, s)

		source1 := testSource(
			"source-1",
			"nginx",
		)
		source2 := testSource(
			"source-2",
			"macos",
		)

		_, err := s.ApplyResources(ctx, []model.Resource{
			source1,
			source2,
		})
		require.NoError(t, err)
		rr := &model.SourceResponse{}

		getRequest(t, client, "/sources/source-2", rr)

		rr.Source.SetDateModified(nil)
		source2.SetDateModified(nil)
		require.Equal(t, source2, rr.Source)
	})

	t.Run("DELETE |sources|:name 404 Not Found", func(t *testing.T) {
		resetStore(t, s)

		deleteEndpoint := fmt.Sprintf("/sources/%s", "does-not-exist")
		resp, err := client.R().Delete(deleteEndpoint)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode())
	})

	t.Run("DELETE |sources|:name 204 No Content", func(t *testing.T) {
		resetStore(t, s)

		source1 := testSource("source-1", "nginx")
		source2 := testSource("source-2", "nginx")

		_, err := s.ApplyResources(ctx, []model.Resource{
			source1,
			source2,
		})
		require.NoError(t, err)

		deleteEndpoint := fmt.Sprintf("/sources/%s", url.PathEscape(source1.Name()))
		resp, err := client.R().Delete(deleteEndpoint)
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, resp.StatusCode())

		sources, err := s.Sources(ctx)
		require.NoError(t, err)

		assert.NotContains(t, sources, source1)
	})

	t.Run("DELETE |sources|:name 409 Conflict", func(t *testing.T) {
		resetStore(t, s)

		source1 := testSource(
			"source-1",
			"nginx",
		)

		config := model.NewConfigurationWithSpec("test-config", model.ConfigurationSpec{
			Sources: []model.ResourceConfiguration{{Name: "source-1"}},
		})

		_, err := store.ApplyResources(ctx, []model.Resource{source1, config})
		require.NoError(t, err)

		deleteEndpoint := fmt.Sprintf("/sources/%s", url.PathEscape("source-1"))
		resp, err := client.R().Delete(deleteEndpoint)
		require.NoError(t, err)

		require.Equal(t, http.StatusConflict, resp.StatusCode())

		body := &ErrorResponse{}
		err = jsoniter.Unmarshal(resp.Body(), body)
		require.NoError(t, err)

		expectBody := ErrorResponse{
			Errors: []string{"Dependent resources:\nConfiguration test-config\n"},
		}

		assert.ElementsMatch(t, expectBody.Errors, body.Errors)
	})

	t.Run("POST |apply Status 200 Accepted", func(t *testing.T) {
		resetStore(t, s)

		destinationAsAny := testDestinationAsAny(t, "destination", "cabin")
		destinationAsResource := testDestination("destination", "cabin")

		configuredDestination := &model.AnyResource{}
		*configuredDestination = *destinationAsAny
		configuredSpec := make(map[string]interface{})
		configuredSpec["parameters"] = []model.Parameter{{Name: "log_format", Value: "default"}}
		configuredSpec["type"] = "cabin"
		configuredDestination.Spec = configuredSpec

		sourceAsAny := testSourceAsAny(t, "source", "macos")
		sourceAsResource := testSource("source", "macos")

		configurationID := uuid.NewString()
		configurationAsAny := testConfigurationAsAny(t, configurationID, "configuration")

		tests := []struct {
			description    string
			setupResources []model.Resource
			payload        *model.ApplyPayload
			want           []struct {
				name   string
				status model.UpdateStatus
			}
		}{
			{
				description:    "create all",
				setupResources: make([]model.Resource, 0),
				payload:        &model.ApplyPayload{Resources: []*model.AnyResource{destinationAsAny, sourceAsAny, configurationAsAny}},
				want: []struct {
					name   string
					status model.UpdateStatus
				}{
					{name: "destination", status: model.StatusCreated},
					{name: "source", status: model.StatusCreated},
					{name: "configuration", status: model.StatusCreated},
				},
			},
			{
				description:    "create and configure and unchanged",
				setupResources: []model.Resource{sourceAsResource, destinationAsResource},
				payload:        &model.ApplyPayload{Resources: []*model.AnyResource{configuredDestination, sourceAsAny, configurationAsAny}},
				want: []struct {
					name   string
					status model.UpdateStatus
				}{
					{name: "destination", status: model.StatusConfigured},
					{name: "configuration", status: model.StatusCreated},
					{name: "source", status: model.StatusUnchanged},
				},
			},
		}

		for _, test := range tests {
			t.Run(test.description, func(t *testing.T) {
				// setup
				resetStore(t, bindplane.Store())
				_, err := bindplane.Store().ApplyResources(ctx, test.setupResources)
				require.NoError(t, err, "expect no error in setup")

				result := &model.ApplyResponseClientSide{}
				resp, err := client.R().SetBody(test.payload).SetResult(result).Post("/apply")
				require.NoError(t, err, "expect no error in rest call")

				assert.Equal(t, http.StatusAccepted, resp.StatusCode())

				got := make([]struct {
					name   string
					status model.UpdateStatus
				}, 0)
				for _, r := range result.Updates {
					t.Logf("reason: %v\n", r.Reason)
					got = append(got, struct {
						name   string
						status model.UpdateStatus
					}{name: r.Resource.Name(), status: r.Status})
				}
				assert.ElementsMatch(t, test.want, got)
			})
		}
	})

	t.Run("GET |configurations", func(t *testing.T) {
		resetStore(t, bindplane.Store())

		endpoint := "/configurations"
		rr := &model.ConfigurationsResponse{}

		getRequest(t, client, endpoint, rr)

		require.Len(t, rr.Configurations, 0)

		testConfiguration1 := testRawConfiguration(uuid.NewString(), "test-configuration-1")
		testConfiguration2 := testRawConfiguration(uuid.NewString(), "test-configuration-2")

		_, err := bindplane.Store().ApplyResources(ctx, []model.Resource{
			testConfiguration1,
			testConfiguration2,
		})
		require.NoError(t, err)

		getRequest(t, client, endpoint, rr)

		require.Len(t, rr.Configurations, 2)
		for _, c := range rr.Configurations {
			c.SetDateModified(nil)
		}
		testConfiguration1.SetDateModified(nil)
		testConfiguration2.SetDateModified(nil)
		require.ElementsMatch(t, rr.Configurations, []*model.Configuration{testConfiguration1, testConfiguration2})
	})

	t.Run("GET |configurations|:name", func(t *testing.T) {
		resetStore(t, s)

		testConfiguration1 := testRawConfiguration(uuid.NewString(), "test-configuration-1")
		testConfiguration2 := testRawConfiguration(uuid.NewString(), "test-configuration-2")

		_, err := bindplane.Store().ApplyResources(ctx, []model.Resource{
			testConfiguration1,
			testConfiguration2,
		})
		require.NoError(t, err)
		pr := &model.ConfigurationResponse{}

		getRequest(t, client, "/configurations/test-configuration-2", pr)

		pr.Configuration.SetDateModified(nil)
		testConfiguration2.SetDateModified(nil)
		require.Equal(t, testConfiguration2, pr.Configuration)
	})

	t.Run("DELETE |configurations/:name 404 Not Found", func(t *testing.T) {
		resetStore(t, s)

		resp, err := client.R().Delete("/configurations/test-configuration-1")
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode())
	})

	t.Run("DELETE |configurations|:name 204 deleted", func(t *testing.T) {
		resetStore(t, s)

		testConfiguration1 := testRawConfiguration(uuid.NewString(), "test-configuration-1")
		testConfiguration2 := testRawConfiguration(uuid.NewString(), "test-configuration-2")

		_, err := bindplane.Store().ApplyResources(ctx, []model.Resource{
			testConfiguration1,
			testConfiguration2,
		})
		require.NoError(t, err)

		deleteEndpoint := fmt.Sprintf("/configurations/%s", testConfiguration1.Name())
		resp, err := client.R().Delete(deleteEndpoint)
		require.NoError(t, err)
		require.Equal(t, resp.StatusCode(), http.StatusNoContent)

		configurations, err := s.Configurations(ctx)
		require.NoError(t, err)

		assert.NotContains(t, configurations, testConfiguration1)
	})

	t.Run("POST |configurations|:name|copy", func(t *testing.T) {
		resetStore(t, s)
		originalName := "original"
		newName := "newName"
		thirdName := "third"

		original := testConfiguration(originalName)
		third := testConfiguration(thirdName)
		_, err := bindplane.Store().ApplyResources(ctx, []model.Resource{original, third})
		require.NoError(t, err)

		t.Run("404 Not Found", func(t *testing.T) {
			endpoint := "/configurations/does-not-exist/copy"

			resp, err := client.R().SetBody(&model.PostCopyConfigRequest{Name: newName}).Post(endpoint)
			require.NoError(t, err)

			require.Equal(t, http.StatusNotFound, resp.StatusCode())
		})

		t.Run("400 Bad Request", func(t *testing.T) {
			endpoint := fmt.Sprintf("/configurations/%s/copy", originalName)
			resp, err := client.R().SetBody(`{"""`).Post(endpoint)
			require.NoError(t, err)

			require.Equal(t, http.StatusBadRequest, resp.StatusCode())

		})

		t.Run("409 Conflict", func(t *testing.T) {
			endpoint := fmt.Sprintf("/configurations/%s/copy", originalName)
			resp, err := client.R().SetBody(&model.PostCopyConfigRequest{Name: thirdName}).Post(endpoint)
			require.NoError(t, err)

			require.Equal(t, http.StatusConflict, resp.StatusCode())
		})

		t.Run("201 Created", func(t *testing.T) {
			endpoint := fmt.Sprintf("/configurations/%s/copy", originalName)
			result := &model.PostCopyConfigResponse{}

			resp, err := client.R().SetBody(&model.PostCopyConfigRequest{Name: newName}).SetResult(result).Post(endpoint)
			require.NoError(t, err)

			require.Equal(t, http.StatusCreated, resp.StatusCode())
			require.Equal(t, result.Name, newName)
		})
	})

	t.Run("POST /delete Status 200 Accepted", func(t *testing.T) {
		tests := []struct {
			description   string
			seedResources []model.Resource
			payload       *model.DeletePayload
			want          []struct {
				name   string
				status model.UpdateStatus
			}
		}{
			{
				description:   "returns nothing on no op",
				seedResources: make([]model.Resource, 0),
				payload:       &model.DeletePayload{Resources: []*model.AnyResource{testDestinationAsAny(t, "destination", "cabin")}},
				want: make([]struct {
					name   string
					status model.UpdateStatus
				}, 0),
			},
			{
				description:   "single resource delete",
				seedResources: []model.Resource{testDestination("destination", "cabin")},
				payload:       &model.DeletePayload{Resources: []*model.AnyResource{testDestinationAsAny(t, "destination", "cabin")}},
				want: []struct {
					name   string
					status model.UpdateStatus
				}{
					{name: "destination", status: model.StatusDeleted},
				},
			},
			{
				description: "multi resource delete",
				seedResources: []model.Resource{
					testDestination("destination", "cabin"),
					testSource("source", "macos"),
					testRawConfiguration(uuid.NewString(), "configuration"),
				},
				payload: &model.DeletePayload{
					Resources: []*model.AnyResource{
						testDestinationAsAny(t, "destination", "cabin"),
						testConfigurationAsAny(t, uuid.NewString(), "configuration"),
						testSourceAsAny(t, "source", "macos"),
					}},
				want: []struct {
					name   string
					status model.UpdateStatus
				}{
					{name: "destination", status: model.StatusDeleted},
					{name: "source", status: model.StatusDeleted},
					{name: "configuration", status: model.StatusDeleted},
				},
			},
		}

		for _, test := range tests {
			t.Run(test.description, func(t *testing.T) {
				// Setup
				resetStore(t, bindplane.Store())
				bindplane.Store().ApplyResources(ctx, test.seedResources)

				result := &model.DeleteResponseClientSide{}
				resp, err := client.R().SetBody(test.payload).SetResult(result).Post("/delete")
				require.NoError(t, err, "expect no error on valid delete call")

				assert.Equal(t, resp.StatusCode(), http.StatusAccepted)

				got := make([]struct {
					name   string
					status model.UpdateStatus
				}, 0)
				for _, r := range result.Updates {
					got = append(got, struct {
						name   string
						status model.UpdateStatus
					}{name: r.Resource.Name(), status: r.Status})
				}

				assert.ElementsMatch(t, test.want, got)
			})
		}
	})

	t.Run("GET |agents|:id|configuration status 200", func(t *testing.T) {
		// Setup
		resetStore(t, bindplane.Store())

		labels := map[string]string{"env": "test", "app": "bindplane"}
		agent1Labels := model.Labels{Set: labels}

		otherLabels := map[string]string{"foo": "bar"}
		agent2labels := model.Labels{Set: otherLabels}

		addAgent(store, &model.Agent{ID: "1", Labels: agent1Labels})
		addAgent(store, &model.Agent{ID: "2", Labels: agent2labels})

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
					Labels:      model.LabelsFromValidatedMap(map[string]string{"platform": "linux"}),
				},
			},
		}

		_, err := s.ApplyResources(ctx, []model.Resource{config})
		require.NoError(t, err)

		config, err = s.StartRollout(ctx, "config", nil)
		require.NoError(t, err)
		require.Equal(t, "Stable", config.Status.Rollout.Status.String())

		expectConfiguration, err := s.Configuration(ctx, "config")
		require.NoError(t, err)

		t.Run("|agents|1|configuration returns config", func(t *testing.T) {
			result := &model.ConfigurationResponse{}
			_, err = client.R().SetResult(result).Get("/agents/1/configuration")
			require.NoError(t, err)
			assert.Equal(t, expectConfiguration, result.Configuration)
		})

		t.Run("|agents|2|configuration returns nil", func(t *testing.T) {
			result := &model.ConfigurationResponse{}
			_, err := client.R().SetResult(result).Get("/agents/2/configuration")
			require.NoError(t, err)

			assert.Nil(t, result.Configuration)
		})

	})

	t.Run("PATCH |agents|labels status 200", func(t *testing.T) {
		resetStore(t, bindplane.Store())

		noConflictLabels := map[string]string{"blah": "foo", "app": "test"}
		conflictingLabels := map[string]string{"test": "this"}

		addAgent(store, &model.Agent{ID: "1", Labels: model.Labels{Set: noConflictLabels}})
		addAgent(store, &model.Agent{ID: "2", Labels: model.Labels{Set: noConflictLabels}})
		addAgent(store, &model.Agent{ID: "3", Labels: model.Labels{Set: conflictingLabels}})

		tests := []struct {
			description string
			payload     *model.BulkAgentLabelsPayload
			expect      *model.BulkAgentLabelsResponse
		}{
			{
				description: "no conflicts, no errors",
				payload: &model.BulkAgentLabelsPayload{
					IDs:    []string{"1", "2"},
					Labels: map[string]string{"test": "that"},
				},
				expect: &model.BulkAgentLabelsResponse{Errors: make([]string, 0)},
			},
			{
				description: "agents not found, errors",
				payload: &model.BulkAgentLabelsPayload{
					IDs:    []string{"4", "5"},
					Labels: map[string]string{"test": "that"},
				},
				expect: &model.BulkAgentLabelsResponse{Errors: []string{
					"failed to apply labels for agent with id 4, agent not found",
					"failed to apply labels for agent with id 5, agent not found",
				}},
			},
			{
				description: "labels conflict, errors",
				payload: &model.BulkAgentLabelsPayload{
					IDs:    []string{"1", "2", "3"},
					Labels: map[string]string{"test": "that"},
				},
				expect: &model.BulkAgentLabelsResponse{
					Errors: []string{"failed to apply labels for agent with id 3, labels conflict, include overwrite: true in body to overwrite"},
				},
			},
			{
				description: "overwrite set, no errors",
				payload: &model.BulkAgentLabelsPayload{
					IDs:       []string{"1", "2", "3"},
					Labels:    map[string]string{"test": "that"},
					Overwrite: true,
				},
				expect: &model.BulkAgentLabelsResponse{Errors: make([]string, 0)},
			},
		}

		for _, test := range tests {
			resetStore(t, bindplane.Store())

			addAgent(store, &model.Agent{ID: "1", Labels: model.Labels{Set: noConflictLabels}})
			addAgent(store, &model.Agent{ID: "2", Labels: model.Labels{Set: noConflictLabels}})
			addAgent(store, &model.Agent{ID: "3", Labels: model.Labels{Set: conflictingLabels}})

			t.Run(test.description, func(t *testing.T) {
				result := &model.BulkAgentLabelsResponse{}
				_, err := client.R().SetBody(test.payload).SetResult(result).Patch("/agents/labels")
				assert.NoError(t, err)

				assert.Equal(t, test.expect, result)
			})
		}
	})
}

func getRequest(t *testing.T, client *resty.Client, endpoint string, result interface{}) {
	_, err := client.R().SetResult(result).Get(endpoint)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRESTMock(t *testing.T) {
	source1 := testSource("source1", "macos")
	source2 := testSource("source2", "macos")
	source1AsAny := testSourceAsAny(t, "source1", "macos")
	agent1 := &model.Agent{ID: "1", Name: "agent1", Labels: model.MakeLabels()}
	agent2 := &model.Agent{ID: "2", Name: "agent2", Labels: model.MakeLabels()}
	configuration1 := testRawConfiguration("1", "configuration1")
	configuration2 := testRawConfiguration("2", "configuration2")
	destination1 := testDestination("destination1", "cabin")
	destination1AsAny := testDestinationAsAny(t, "destination", "cabin")
	destination2 := testDestination("destination2", "cabin")
	testConfig1 := testRawConfiguration("1", "config1")
	testConfig2 := testRawConfiguration("2", "config2")

	malformedDestination := &model.AnyResource{}
	*malformedDestination = *destination1AsAny
	malformedDestination.Kind = "unknown"

	installCommandParams := installCommandParameters{
		platform:  "windows-amd64",
		version:   "2.1.1",
		labels:    "app=bindplane,env=test",
		secretKey: "uuid",
		remoteURL: "localhost:3001",
	}
	expectInstallText, err := installCommandParams.installCommand()
	require.NoError(t, err)

	tests := []struct {
		method       string
		endpoint     string
		requestBody  interface{}
		resultPtr    interface{}
		expectStatus int
		expectResult interface{}
		mockFunction string
		mockArgs     []interface{}
		mockReturn   []interface{}
	}{
		/* ----------------------------- Apply Resources ---------------------------- */
		{
			method:   "POST",
			endpoint: "/apply",
			requestBody: model.ApplyPayload{
				Resources: []*model.AnyResource{
					destination1AsAny,
					source1AsAny,
				},
			},
			resultPtr:    &model.ApplyResponseClientSide{},
			expectStatus: 202,
			expectResult: nil,

			mockFunction: "ApplyResources",
			mockArgs:     []interface{}{mock.Anything, mock.Anything, mock.Anything},
			mockReturn: []interface{}{
				[]model.ResourceStatus{
					{
						Status: model.StatusCreated,
						Resource: &model.AnyResource{
							ResourceMeta: model.ResourceMeta{
								APIVersion: version.V1,
								Kind:       model.KindDestination,
								Metadata: model.Metadata{
									ID:   "1",
									Name: "destination1",
								},
							},
							Spec: map[string]interface{}{
								"parameters": []interface{}{},
							},
						},
					},
					{
						Status:   model.StatusCreated,
						Resource: source1,
					},
				},
				nil,
			},
		},
		{
			method:       "POST",
			endpoint:     "/apply",
			requestBody:  `{"This":"is","malformed"":"json"}`,
			resultPtr:    &ErrorResponse{},
			expectStatus: 400,
		},
		{
			method:   "POST",
			endpoint: "/apply",
			requestBody: &model.ApplyPayload{
				Resources: []*model.AnyResource{
					malformedDestination,
				},
			},
			resultPtr: &ErrorResponse{},
			expectResult: &ErrorResponse{
				Errors: []string{"unknown resource kind: unknown"},
			},
			expectStatus: 400,
		},
		{
			method:   "POST",
			endpoint: "/apply",
			requestBody: &model.ApplyPayload{
				Resources: []*model.AnyResource{
					source1AsAny,
					destination1AsAny,
				},
			},
			resultPtr: &ErrorResponse{},
			expectResult: &ErrorResponse{
				Errors: []string{"internal server error"},
			},
			expectStatus: 500,

			mockFunction: "ApplyResources",
			mockArgs:     []interface{}{mock.Anything, mock.Anything},
			mockReturn:   []interface{}{[]model.ResourceStatus{}, errors.New("internal server error")},
		},

		/* ----------------------------- Delete Endpoint ---------------------------- */
		{
			method:   "POST",
			endpoint: "/delete",
			requestBody: model.DeletePayload{
				Resources: []*model.AnyResource{
					destination1AsAny,
					source1AsAny,
				},
			},
			expectStatus: 202,
			expectResult: nil,

			mockFunction: "DeleteResources",
			mockArgs:     []interface{}{mock.Anything, mock.Anything},
			mockReturn: []interface{}{
				[]model.ResourceStatus{
					{Status: model.StatusDeleted, Resource: destination1},
				},
				nil,
			},
		},
		{
			method:       "POST",
			endpoint:     "/delete",
			requestBody:  `{"some": ""malformed","json":2}`,
			expectStatus: 400,
		},
		{
			method:   "POST",
			endpoint: "/delete",
			requestBody: &model.DeletePayload{
				Resources: []*model.AnyResource{
					malformedDestination,
				},
			},
			expectStatus: 400,
		},
		{
			method:   "POST",
			endpoint: "/delete",
			requestBody: model.DeletePayload{
				Resources: []*model.AnyResource{
					destination1AsAny,
					source1AsAny,
				},
			},
			resultPtr:    &ErrorResponse{},
			expectStatus: 500,
			expectResult: &ErrorResponse{
				Errors: []string{"internal server error"},
			},

			mockFunction: "DeleteResources",
			mockArgs:     []interface{}{mock.Anything, mock.Anything},
			mockReturn:   []interface{}{[]model.ResourceStatus{}, errors.New("internal server error")},
		},
		/* --------------------------- Source Endpoints --------------------------- */
		{
			method:       "GET",
			endpoint:     "/sources",
			requestBody:  nil,
			resultPtr:    &model.SourcesResponse{},
			expectStatus: 200,
			expectResult: &model.SourcesResponse{
				Sources: []*model.Source{source1, source2},
			},

			mockFunction: "Sources",
			mockArgs:     []interface{}{mock.Anything},
			mockReturn:   []interface{}{[]*model.Source{source1, source2}, nil},
		},
		{
			method:       "GET",
			endpoint:     "/sources",
			requestBody:  nil,
			resultPtr:    &ErrorResponse{},
			expectStatus: 500,
			expectResult: &ErrorResponse{
				Errors: []string{"internal server error"},
			},

			mockFunction: "Sources",
			mockArgs:     []interface{}{mock.Anything},
			mockReturn:   []interface{}{[]*model.Source{}, errors.New("internal server error")},
		},
		{
			method:       "GET",
			endpoint:     "/sources/name",
			requestBody:  nil,
			resultPtr:    &model.SourceResponse{},
			expectStatus: 200,
			expectResult: &model.SourceResponse{
				Source: source1,
			},

			mockFunction: "Source",
			mockArgs:     []interface{}{mock.Anything, "name"},
			mockReturn:   []interface{}{source1, nil},
		},
		{
			method:       "GET",
			endpoint:     "/sources/does-not-exist",
			requestBody:  nil,
			resultPtr:    &ErrorResponse{},
			expectStatus: 404,
			expectResult: &ErrorResponse{
				Errors: []string{ErrResourceNotFound.Error()},
			},

			mockFunction: "Source",
			mockArgs:     []interface{}{mock.Anything, "does-not-exist"},
			mockReturn:   []interface{}{nil, nil},
		},
		{
			method:       "GET",
			endpoint:     "/sources/name",
			requestBody:  nil,
			resultPtr:    &ErrorResponse{},
			expectStatus: 500,
			expectResult: &ErrorResponse{
				Errors: []string{"internal server error"},
			},

			mockFunction: "Source",
			mockArgs:     []interface{}{mock.Anything, "name"},
			mockReturn:   []interface{}{nil, errors.New("internal server error")},
		},
		{
			method:       "DELETE",
			endpoint:     "/sources/name",
			requestBody:  nil,
			resultPtr:    nil,
			expectStatus: 204, // no content
			expectResult: nil,

			mockFunction: "DeleteSource",
			mockArgs:     []interface{}{mock.Anything, "name"},
			mockReturn:   []interface{}{source1, nil},
		},
		{
			method:       "DELETE",
			endpoint:     "/sources/does-not-exist",
			requestBody:  nil,
			resultPtr:    &ErrorResponse{},
			expectStatus: 404,
			expectResult: &ErrorResponse{
				Errors: []string{ErrResourceNotFound.Error()},
			},

			mockFunction: "DeleteSource",
			mockArgs:     []interface{}{mock.Anything, "does-not-exist"},
			mockReturn:   []interface{}{nil, nil},
		},
		{
			method:       "DELETE",
			endpoint:     "/sources/name",
			requestBody:  nil,
			resultPtr:    &ErrorResponse{},
			expectStatus: 500,
			expectResult: &ErrorResponse{
				Errors: []string{"internal server error"},
			},

			mockFunction: "DeleteSource",
			mockArgs:     []interface{}{mock.Anything, "name"},
			mockReturn:   []interface{}{nil, errors.New("internal server error")},
		},

		/* ------------------------------ bindplane version ------------------------------ */
		{
			method:       "GET",
			endpoint:     "/version",
			expectStatus: 200,
		},

		/* ----------------------------- Install Command ---------------------------- */
		{
			method:       "GET",
			endpoint:     "/agent-versions/2.1.1/install-command",
			expectStatus: 200,
		},
		{
			method:       "GET",
			endpoint:     "/agent-versions/2.1.1/install-command?platform=windows-amd64&labels=app%3Dbindplane%2Cenv%3Dtest&secret-key=uuid&remote-url=localhost%3A3001",
			resultPtr:    &model.InstallCommandResponse{},
			expectStatus: 200,
			expectResult: &model.InstallCommandResponse{
				Command: expectInstallText,
			},
		},
		/* ---------------------------- Agents Endpoints ---------------------------- */
		{
			method:       "GET",
			endpoint:     "/agents",
			requestBody:  nil,
			resultPtr:    &model.AgentsResponse{},
			expectStatus: 200,
			expectResult: &model.AgentsResponse{
				Agents: []*model.Agent{agent1, agent2},
			},

			mockFunction: "Agents",
			mockArgs:     []interface{}{mock.Anything, mock.Anything, mock.Anything, mock.Anything},
			mockReturn:   []interface{}{[]*model.Agent{agent1, agent2}, nil},
		},
		{
			method:       "GET",
			endpoint:     "/agents",
			requestBody:  nil,
			resultPtr:    &ErrorResponse{},
			expectStatus: 500,
			expectResult: &ErrorResponse{
				Errors: []string{"internal server error"},
			},

			mockFunction: "Agents",
			mockArgs:     []interface{}{mock.Anything, mock.Anything, mock.Anything, mock.Anything},
			mockReturn:   []interface{}{[]*model.Agent{}, errors.New("internal server error")},
		},
		{
			method:       "GET",
			endpoint:     "/agents/id",
			requestBody:  nil,
			resultPtr:    &model.AgentResponse{},
			expectStatus: 200,
			expectResult: &model.AgentResponse{
				Agent: agent1,
			},

			mockFunction: "Agent",
			mockArgs:     []interface{}{mock.Anything, "id"},
			mockReturn:   []interface{}{agent1, nil},
		},
		{
			method:   "POST",
			endpoint: "/agents/does-not-exist/version",
			requestBody: &model.PostAgentVersionRequest{
				Version: "v1.5.0",
			},
			resultPtr:    &ErrorResponse{},
			expectStatus: 404,
			expectResult: &ErrorResponse{
				Errors: []string{ErrResourceNotFound.Error()},
			},

			mockFunction: "Agent",
			mockArgs:     []interface{}{mock.Anything, "does-not-exist"},
			mockReturn:   []interface{}{nil, nil},
		},
		{
			method:       "GET",
			endpoint:     "/agents/does-not-exist",
			requestBody:  nil,
			resultPtr:    &ErrorResponse{},
			expectStatus: 404,
			expectResult: &ErrorResponse{
				Errors: []string{ErrResourceNotFound.Error()},
			},

			mockFunction: "Agent",
			mockArgs:     []interface{}{mock.Anything, "does-not-exist"},
			mockReturn:   []interface{}{nil, nil},
		},
		{
			method:       "GET",
			endpoint:     "/agents/name",
			requestBody:  nil,
			resultPtr:    &ErrorResponse{},
			expectStatus: 500,
			expectResult: &ErrorResponse{
				Errors: []string{"internal server error"},
			},

			mockFunction: "Agent",
			mockArgs:     []interface{}{mock.Anything, "name"},
			mockReturn:   []interface{}{nil, errors.New("internal server error")},
		},

		/* --------------------------- Configuration Endpoints --------------------------- */
		{
			method:       "GET",
			endpoint:     "/configurations",
			requestBody:  nil,
			resultPtr:    &model.ConfigurationsResponse{},
			expectStatus: 200,
			expectResult: &model.ConfigurationsResponse{
				Configurations: []*model.Configuration{configuration1, configuration2},
			},

			mockFunction: "Configurations",
			mockArgs:     []interface{}{mock.Anything},
			mockReturn:   []interface{}{[]*model.Configuration{configuration1, configuration2}, nil},
		},
		{
			method:       "GET",
			endpoint:     "/configurations",
			requestBody:  nil,
			resultPtr:    &ErrorResponse{},
			expectStatus: 500,
			expectResult: &ErrorResponse{
				Errors: []string{"internal server error"},
			},

			mockFunction: "Configurations",
			mockArgs:     []interface{}{mock.Anything},
			mockReturn:   []interface{}{[]*model.Configuration{}, errors.New("internal server error")},
		},
		{
			method:       "GET",
			endpoint:     "/configurations/name",
			requestBody:  nil,
			resultPtr:    &model.ConfigurationResponse{},
			expectStatus: 200,
			expectResult: &model.ConfigurationResponse{
				Configuration: configuration1,
				Raw:           "raw:",
			},

			mockFunction: "Configuration",
			mockArgs:     []interface{}{mock.Anything, "name"},
			mockReturn:   []interface{}{configuration1, nil},
		},
		{
			method:       "GET",
			endpoint:     "/configurations/does-not-exist",
			requestBody:  nil,
			resultPtr:    &ErrorResponse{},
			expectStatus: 404,
			expectResult: &ErrorResponse{
				Errors: []string{ErrResourceNotFound.Error()},
			},

			mockFunction: "Configuration",
			mockArgs:     []interface{}{mock.Anything, "does-not-exist"},
			mockReturn:   []interface{}{nil, nil},
		},
		{
			method:       "GET",
			endpoint:     "/configurations/name",
			requestBody:  nil,
			resultPtr:    &ErrorResponse{},
			expectStatus: 500,
			expectResult: &ErrorResponse{
				Errors: []string{"internal server error"},
			},

			mockFunction: "Configuration",
			mockArgs:     []interface{}{mock.Anything, "name"},
			mockReturn:   []interface{}{nil, errors.New("internal server error")},
		},
		{
			method:       "DELETE",
			endpoint:     "/configurations/name",
			requestBody:  nil,
			resultPtr:    nil,
			expectStatus: 204, // no content
			expectResult: nil,

			mockFunction: "DeleteConfiguration",
			mockArgs:     []interface{}{mock.Anything, "name"},
			mockReturn:   []interface{}{configuration1, nil},
		},
		{
			method:       "DELETE",
			endpoint:     "/configurations/does-not-exist",
			requestBody:  nil,
			resultPtr:    &ErrorResponse{},
			expectStatus: 404,
			expectResult: &ErrorResponse{
				Errors: []string{ErrResourceNotFound.Error()},
			},

			mockFunction: "DeleteConfiguration",
			mockArgs:     []interface{}{mock.Anything, "does-not-exist"},
			mockReturn:   []interface{}{nil, nil},
		},
		{
			method:       "DELETE",
			endpoint:     "/configurations/name",
			requestBody:  nil,
			resultPtr:    &ErrorResponse{},
			expectStatus: 500,
			expectResult: &ErrorResponse{
				Errors: []string{"internal server error"},
			},

			mockFunction: "DeleteConfiguration",
			mockArgs:     []interface{}{mock.Anything, "name"},
			mockReturn:   []interface{}{nil, errors.New("internal server error")},
		},

		/* --------------------------- Destination Endpoints --------------------------- */
		{
			method:       "GET",
			endpoint:     "/destinations",
			requestBody:  nil,
			resultPtr:    &model.DestinationsResponse{},
			expectStatus: 200,
			expectResult: &model.DestinationsResponse{
				Destinations: []*model.Destination{destination1, destination2},
			},

			mockFunction: "Destinations",
			mockArgs:     []interface{}{mock.Anything},
			mockReturn:   []interface{}{[]*model.Destination{destination1, destination2}, nil},
		},
		{
			method:       "GET",
			endpoint:     "/destinations",
			requestBody:  nil,
			resultPtr:    &ErrorResponse{},
			expectStatus: 500,
			expectResult: &ErrorResponse{
				Errors: []string{"internal server error"},
			},

			mockFunction: "Destinations",
			mockArgs:     []interface{}{mock.Anything},
			mockReturn:   []interface{}{[]*model.Destination{}, errors.New("internal server error")},
		},
		{
			method:       "GET",
			endpoint:     "/destinations/name",
			requestBody:  nil,
			resultPtr:    &model.DestinationResponse{},
			expectStatus: 200,
			expectResult: &model.DestinationResponse{
				Destination: destination1,
			},

			mockFunction: "Destination",
			mockArgs:     []interface{}{mock.Anything, "name"},
			mockReturn:   []interface{}{destination1, nil},
		},
		{
			method:       "GET",
			endpoint:     "/destinations/does-not-exist",
			requestBody:  nil,
			resultPtr:    &ErrorResponse{},
			expectStatus: 404,
			expectResult: &ErrorResponse{
				Errors: []string{ErrResourceNotFound.Error()},
			},

			mockFunction: "Destination",
			mockArgs:     []interface{}{mock.Anything, "does-not-exist"},
			mockReturn:   []interface{}{nil, nil},
		},
		{
			method:       "GET",
			endpoint:     "/destinations/name",
			requestBody:  nil,
			resultPtr:    &ErrorResponse{},
			expectStatus: 500,
			expectResult: &ErrorResponse{
				Errors: []string{"internal server error"},
			},

			mockFunction: "Destination",
			mockArgs:     []interface{}{mock.Anything, "name"},
			mockReturn:   []interface{}{nil, errors.New("internal server error")},
		},
		{
			method:       "DELETE",
			endpoint:     "/destinations/name",
			requestBody:  nil,
			resultPtr:    nil,
			expectStatus: 204, // no content
			expectResult: nil,

			mockFunction: "DeleteDestination",
			mockArgs:     []interface{}{mock.Anything, "name"},
			mockReturn:   []interface{}{destination1, nil},
		},
		{
			method:       "DELETE",
			endpoint:     "/destinations/does-not-exist",
			requestBody:  nil,
			resultPtr:    &ErrorResponse{},
			expectStatus: 404,
			expectResult: &ErrorResponse{
				Errors: []string{ErrResourceNotFound.Error()},
			},

			mockFunction: "DeleteDestination",
			mockArgs:     []interface{}{mock.Anything, "does-not-exist"},
			mockReturn:   []interface{}{nil, nil},
		},
		{
			method:       "DELETE",
			endpoint:     "/destinations/name",
			requestBody:  nil,
			resultPtr:    &ErrorResponse{},
			expectStatus: 500,
			expectResult: &ErrorResponse{
				Errors: []string{"internal server error"},
			},

			mockFunction: "DeleteDestination",
			mockArgs:     []interface{}{mock.Anything, "name"},
			mockReturn:   []interface{}{nil, errors.New("internal server error")},
		},
		/* --------------------------- Configuration Endpoints --------------------------- */
		{
			method:      "GET",
			endpoint:    "/configurations",
			requestBody: nil,
			resultPtr:   &model.ConfigurationsResponse{},
			expectResult: &model.ConfigurationsResponse{
				Configurations: []*model.Configuration{testConfig1, testConfig2},
			},
			expectStatus: 200,

			mockFunction: "Configurations",
			mockArgs:     []interface{}{mock.Anything},
			mockReturn:   []interface{}{[]*model.Configuration{testConfig1, testConfig2}, nil},
		},
		{
			method:       "GET",
			endpoint:     "/configurations",
			requestBody:  nil,
			resultPtr:    &ErrorResponse{},
			expectResult: &ErrorResponse{Errors: []string{"internal server error"}},
			expectStatus: 500,

			mockFunction: "Configurations",
			mockArgs:     []interface{}{mock.Anything},
			mockReturn:   []interface{}{[]*model.Configuration{}, errors.New("internal server error")},
		},
		{
			method:       "GET",
			endpoint:     "/configurations/name",
			requestBody:  nil,
			resultPtr:    &model.ConfigurationResponse{},
			expectStatus: 200,
			expectResult: &model.ConfigurationResponse{
				Configuration: testConfig1,
				Raw:           "raw:",
			},

			mockFunction: "Configuration",
			mockArgs:     []interface{}{mock.Anything, "name"},
			mockReturn:   []interface{}{testConfig1, nil},
		},
		{
			method:       "GET",
			endpoint:     "/configurations/does-not-exist",
			requestBody:  nil,
			resultPtr:    &ErrorResponse{},
			expectStatus: 404,
			expectResult: &ErrorResponse{
				Errors: []string{ErrResourceNotFound.Error()},
			},

			mockFunction: "Configuration",
			mockArgs:     []interface{}{mock.Anything, "does-not-exist"},
			mockReturn:   []interface{}{nil, nil},
		},
		{
			method:       "GET",
			endpoint:     "/configurations/name",
			requestBody:  nil,
			resultPtr:    &ErrorResponse{},
			expectStatus: 500,
			expectResult: &ErrorResponse{
				Errors: []string{"internal server error"},
			},

			mockFunction: "Configuration",
			mockArgs:     []interface{}{mock.Anything, "name"},
			mockReturn:   []interface{}{nil, errors.New("internal server error")},
		},
		{
			method:       "DELETE",
			endpoint:     "/configurations/name",
			requestBody:  nil,
			resultPtr:    nil,
			expectStatus: 204, // no content
			expectResult: nil,

			mockFunction: "DeleteConfiguration",
			mockArgs:     []interface{}{mock.Anything, "name"},
			mockReturn:   []interface{}{configuration1, nil},
		},
		{
			method:       "DELETE",
			endpoint:     "/configurations/does-not-exist",
			requestBody:  nil,
			resultPtr:    &ErrorResponse{},
			expectStatus: 404,
			expectResult: &ErrorResponse{
				Errors: []string{ErrResourceNotFound.Error()},
			},

			mockFunction: "DeleteConfiguration",
			mockArgs:     []interface{}{mock.Anything, "does-not-exist"},
			mockReturn:   []interface{}{nil, nil},
		},
		{
			method:       "DELETE",
			endpoint:     "/configurations/name",
			requestBody:  nil,
			resultPtr:    &ErrorResponse{},
			expectStatus: 500,
			expectResult: &ErrorResponse{
				Errors: []string{"internal server error"},
			},

			mockFunction: "DeleteConfiguration",
			mockArgs:     []interface{}{mock.Anything, "name"},
			mockReturn:   []interface{}{nil, errors.New("internal server error")},
		},
	}

	for _, test := range tests {
		t.Run(strings.Join([]string{test.method, test.endpoint, fmt.Sprint(test.expectStatus)}, " "), func(t *testing.T) {
			router := gin.Default()
			svr := httptest.NewServer(router)
			defer svr.Close()

			store := &storeMocks.MockStore{}
			mockBatcher := statsmocks.NewMockMeasurementBatcher(t)
			bindplane := server.NewBindPlane(&config.Config{}, zaptest.NewLogger(t), store, nil, mockBatcher)
			AddRestRoutes(router, bindplane)

			client := resty.New()
			client.SetBaseURL(svr.URL)

			// Set the return for the mocked store method
			if len(test.mockArgs) > 0 {
				store.On(test.mockFunction, test.mockArgs...).Return(test.mockReturn...)
			} else {
				store.On(test.mockFunction).Return(test.mockReturn...)
			}

			request := client.R()

			if test.requestBody != nil {
				request.SetBody(test.requestBody)
			}

			resp, err := request.Execute(test.method, test.endpoint)
			require.NoError(t, err)

			if test.resultPtr != nil {
				// parse the body directly because SetResult only works for status codes 200-299
				err = jsoniter.Unmarshal(resp.Body(), test.resultPtr)
				require.NoError(t, err)
			}

			if test.expectResult != nil {
				assert.Equal(t, test.expectResult, test.resultPtr)
			}
			assert.Equal(t, test.expectStatus, resp.StatusCode())
		})
	}
}
