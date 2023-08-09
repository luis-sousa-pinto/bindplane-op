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

package store

// This file contains shared tests for mapstore and boltstore

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/observiq/bindplane-op/eventbus"
	"github.com/observiq/bindplane-op/model"
	modelversion "github.com/observiq/bindplane-op/model/version"
	"github.com/observiq/bindplane-op/otlp/record"
	"github.com/observiq/bindplane-op/store/search"
	"github.com/observiq/bindplane-op/util"

	"github.com/observiq/bindplane-op/store/stats"
)

func addAgent(s Store, agent *model.Agent) error {
	_, err := s.UpsertAgent(context.TODO(), agent.ID, func(a *model.Agent) {
		*a = *agent
	})
	return err
}

func labels(m map[string]string) model.Labels {
	labels, _ := model.LabelsFromMap(m)
	return labels
}

func pruneDateModified(r model.Resource) {
	if r != nil && !reflect.ValueOf(r).IsZero() {
		// clear the data modified because this will vary
		r.SetDateModified(nil)
	}
}

func pruneResourceMeta(r model.Resource) {
	if r != nil && !reflect.ValueOf(r).IsZero() {
		// clear the data modified because this will vary
		r.SetDateModified(nil)

		// clear the Version because there are specific tests for Version support
		r.SetVersion(0)

		// make sure the hash is computed
		r.EnsureHash(r.GetSpec())

		// there are specific tests for Latest
		r.SetLatest(false)
	}
}
func pruneResourceMetas[T model.Resource](r []T) {
	for _, m := range r {
		pruneResourceMeta(m)
	}
}
func pruneResourceStatuses(r []model.ResourceStatus) {
	for _, m := range r {
		pruneResourceMeta(m.Resource)
	}
}

func assertResourcesMatch[T model.Resource](t *testing.T, expected, actual []T, msgAndArgs ...interface{}) bool {
	pruneResourceMetas(expected)
	pruneResourceMetas(actual)
	return assert.ElementsMatch(t, expected, actual, msgAndArgs...)
}

func assertResourceStatusesMatch(t *testing.T, expected, actual []model.ResourceStatus, msgAndArgs ...interface{}) bool {
	pruneResourceStatuses(expected)
	pruneResourceStatuses(actual)
	if len(expected) == len(actual) && len(expected) == 1 {
		// if there is only one item, it is much easier to see differences with Equal
		return assert.Equal(t, expected[0], actual[0], msgAndArgs...)
	}
	return assert.ElementsMatch(t, expected, actual, msgAndArgs...)
}

func assertResourcesEqual[T model.Resource](t *testing.T, expected, actual T, msgAndArgs ...interface{}) bool {
	pruneResourceMeta(expected)
	pruneResourceMeta(actual)
	return assert.Equal(t, expected, actual, msgAndArgs...)
}

func assertResourceVersionsEqual[T model.Resource](t *testing.T, expected, actual T, msgAndArgs ...interface{}) bool {
	pruneDateModified(expected)
	pruneDateModified(actual)
	return assert.Equal(t, expected, actual, msgAndArgs...)
}

func assertEventsEqual[T model.Resource](t *testing.T, expected, actual Events[T]) {
	for _, item := range expected {
		item.Item.SetID("")
	}
	for _, item := range actual {
		item.Item.SetID("")
	}
	require.Equal(t, expected, actual)
}

func assertUpdatesEqual(t *testing.T, expected, actual BasicEventUpdates) {
	require.Equal(t, expected.Agents, actual.Agents)
	assertEventsEqual(t, expected.AgentVersions(), actual.AgentVersions())
	assertEventsEqual(t, expected.Sources(), actual.Sources())
	assertEventsEqual(t, expected.SourceTypes(), actual.SourceTypes())
	assertEventsEqual(t, expected.Processors(), actual.Processors())
	assertEventsEqual(t, expected.ProcessorTypes(), actual.ProcessorTypes())
	assertEventsEqual(t, expected.Destinations(), actual.Destinations())
	assertEventsEqual(t, expected.DestinationTypes(), actual.DestinationTypes())
	assertEventsEqual(t, expected.Configurations(), actual.Configurations())
}

var (
	cabinDestinationType = model.NewDestinationType("cabin", []model.ParameterDefinition{
		{
			Name: "s",
			Type: "string",
		},
	})

	cabinDestination1        = model.NewDestination("cabin-1", "cabin", nil)
	cabinDestination1Changed = model.NewDestination("cabin-1", "cabin", []model.Parameter{
		{
			Name:  "s",
			Value: "1",
		},
	})
	cabinDestination2 = model.NewDestination("cabin-2", "cabin", nil)

	macosSourceType = model.NewSourceType("macos", []model.ParameterDefinition{
		{
			Name: "s",
			Type: "string",
		},
	}, []string{"macos"})

	macosSource        = model.NewSource("macos-1", "macos", nil)
	macosSourceChanged = model.NewSource("macos-1", "macos", []model.Parameter{
		{
			Name:  "s",
			Value: "1",
		},
	})

	nginxSourceType = model.NewSourceType("nginx", []model.ParameterDefinition{
		{
			Name: "s",
			Type: "string",
		},
	}, []string{"macos", "linux", "windows"})
	nginxSource        = model.NewSource("nginx", "nginx", nil)
	nginxSourceChanged = model.NewSource("nginx", "nginx", []model.Parameter{
		{
			Name:  "s",
			Value: "1",
		},
	})

	invalidSource  = model.NewSource("_production-nginx-ingress_", "macos", nil)
	invalidSource2 = model.NewSource("foo/bar/baz", "macos", nil)

	unknownResource = model.AnyResource{
		ResourceMeta: model.ResourceMeta{
			Kind: model.Kind("not-a-real-resource"),
			Metadata: model.Metadata{
				Name: "unknown",
			},
		},
	}

	testConfiguration = model.NewConfigurationWithSpec("configuration-1", model.ConfigurationSpec{
		Sources: []model.ResourceConfiguration{
			{
				Name: macosSource.Name(),
			},
		},
		Destinations: []model.ResourceConfiguration{
			{
				Name: cabinDestination1.Name(),
			},
		},
	})

	testConfigurationChanged = model.NewConfigurationWithSpec("configuration-1", model.ConfigurationSpec{
		Sources: []model.ResourceConfiguration{
			{
				Name: macosSource.Name(),
				ParameterizedSpec: model.ParameterizedSpec{
					Parameters: []model.Parameter{
						{
							Name:  "s",
							Value: "1",
						},
					},
				},
			},
		},
		Destinations: []model.ResourceConfiguration{
			{
				Name: cabinDestination1.Name(),
			},
		},
	})

	testRawConfiguration1 = model.NewRawConfiguration("Test-configuration-1", "raw:")
	testRawConfiguration2 = model.NewRawConfiguration("test-configuration-2", "raw:")
)

func applyTestTypes(t *testing.T, store Store) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	statuses, err := store.ApplyResources(ctx, cloneResources(t, []model.Resource{
		cabinDestinationType,
		macosSourceType,
		nginxSourceType,
	}))
	require.NoError(t, err)
	requireOkStatuses(t, statuses)
}

func applyTestConfiguration(t *testing.T, store Store) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	statuses, err := store.ApplyResources(ctx, cloneResources(t, []model.Resource{
		cabinDestinationType,
		cabinDestination1,
		cabinDestination2,
		macosSourceType,
		macosSource,
		nginxSourceType,
		nginxSource,
		testConfiguration,
	}))
	t.Logf("statuses %v\n", statuses)
	require.NoError(t, err)
	requireOkStatuses(t, statuses)
}

func applyAllTestResources(t *testing.T, store Store) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	statuses, err := store.ApplyResources(ctx, []model.Resource{
		cabinDestinationType,
		cabinDestination1,
		cabinDestination2,
		macosSourceType,
		macosSource,
		nginxSourceType,
		nginxSource,
		testConfiguration,
		testRawConfiguration1,
		testRawConfiguration2,
	})
	require.NoError(t, err)
	requireOkStatuses(t, statuses)
}

type configurationChanges struct {
	configurationsUpdated []string
	configurationsRemoved []string
}

func expectedUpdates(configurations ...string) configurationChanges {
	return configurationChanges{
		configurationsUpdated: configurations,
	}
}

func expectedRemoves(configurations ...string) configurationChanges {
	return configurationChanges{
		configurationsRemoved: configurations,
	}
}

func configurationChangesFromUpdates(t *testing.T, updates BasicEventUpdates) configurationChanges {
	var updated []string
	var removed []string

	for _, event := range updates.Configurations() {
		t.Logf("event[%s]: %+v\n", event.Item.Kind, event)
		if event.Type == EventTypeRemove {
			removed = append(removed, event.Item.Name())
		} else {
			updated = append(updated, event.Item.Name())
		}
	}
	changes := configurationChanges{
		configurationsUpdated: updated,
		configurationsRemoved: removed,
	}
	return changes
}

func verifyUpdates(t *testing.T, done chan bool, Updates <-chan BasicEventUpdates, expected []configurationChanges) {
	complete := func(success bool) {
		done <- success
	}
	i := 0
	for {
		select {
		case <-time.After(5 * time.Second):
			complete(false)
			t.Log("Timed out waiting for updates.")
			return
		case updates := <-Updates:
			if !assert.Less(t, i, len(expected), "more changes than expected") {
				complete(false)
				return
			}
			actual := configurationChangesFromUpdates(t, updates)

			t.Logf("actual %v\nexpected %v", actual, expected)

			if !assert.ElementsMatch(t, expected[i].configurationsRemoved, actual.configurationsRemoved, "configurationsRemoved should match") {
				complete(false)
				return
			}
			if !assert.ElementsMatch(t, expected[i].configurationsUpdated, actual.configurationsUpdated, "configurationsUpdated should match") {
				complete(false)
				return
			}
			i++
			if i == len(expected) {
				complete(true)
				return
			}
		}
	}
}

func runNotifyUpdatesTests(t *testing.T, store Store, done chan bool) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	update := func(r model.Resource) {
		status, err := store.ApplyResources(ctx, cloneResources(t, []model.Resource{r}))
		require.NoError(t, err)
		requireOkStatuses(t, status)
	}

	updates, _ := eventbus.Subscribe(ctx, store.Updates(ctx))
	applyAllTestResources(t, store)
	verifyUpdates(t, done, updates, []configurationChanges{
		expectedUpdates(testConfiguration.Name(), testRawConfiguration1.Name(), testRawConfiguration2.Name()),
	})
	ok := <-done
	require.True(t, ok)

	// these tests are dependent on each other and are expected to run in order.

	t.Run("update nginx, expect no configuration changes", func(t *testing.T) {
		go verifyUpdates(t, done, updates, []configurationChanges{
			expectedUpdates(),
		})
		update(nginxSourceChanged)
		ok := <-done
		require.True(t, ok)
	})

	t.Run("update configuration, expect configuration-1 change", func(t *testing.T) {
		update(testConfigurationChanged)
		verifyUpdates(t, done, updates, []configurationChanges{
			expectedUpdates(testConfiguration.Name()),
		})
		ok := <-done
		require.True(t, ok)
	})

	t.Run("update macos, expect configuration-1 change", func(t *testing.T) {
		update(macosSourceChanged)
		verifyUpdates(t, done, updates, []configurationChanges{
			expectedUpdates(testConfiguration.Name()),
		})
		ok := <-done
		require.True(t, ok)
	})

	t.Run("update cabin-1, expect configuration-1 change", func(t *testing.T) {
		update(cabinDestination1Changed)
		verifyUpdates(t, done, updates, []configurationChanges{
			expectedUpdates(testConfiguration.Name()),
		})
		ok := <-done
		require.True(t, ok)
	})

	t.Run("update everything, expect configuration-1 change", func(t *testing.T) {
		store.ApplyResources(ctx, []model.Resource{
			macosSource,
			macosSourceType,
			nginxSource,
			nginxSourceType,
			cabinDestination1,
			cabinDestination2,
			testConfiguration,
		})
		verifyUpdates(t, done, updates, []configurationChanges{
			expectedUpdates(testConfiguration.Name()),
		})
		ok := <-done
		require.True(t, ok)
	})

	t.Run("delete configuration, expect configuration-1 remove", func(t *testing.T) {
		// setup
		applyTestConfiguration(t, store)
		// Test batch delete here
		_, err := store.DeleteConfiguration(ctx, testConfiguration.Name())
		require.NoError(t, err)

		verifyUpdates(t, done, updates, []configurationChanges{
			expectedRemoves(testConfiguration.Name()),
		})
		ok := <-done
		require.True(t, ok)
	})
}

func runAgentConfigurationTests(ctx context.Context, t *testing.T, store Store, setupFunc func(Store)) {
	tests := []struct {
		description              string
		agentLabels              string
		configurationsLabels     []string
		expectConfigurationIndex int
		expectNil                bool
		pendingConfig            string
		currentConfig            string
		expectFutureConfig       string
	}{
		{
			description: "sets agent's future configuration using configuration= label",
			agentLabels: "configuration=c0",
			configurationsLabels: []string{
				"configuration=c0",
				"configuration=c1",
				"configuration=c2",
			},
			expectConfigurationIndex: 0,
			expectNil:                true,
			expectFutureConfig:       "c0:1",
		},
		{
			description: "doesn't select a configuration without labels or pending or current",
			agentLabels: "",
			configurationsLabels: []string{
				"configuration=c0",
				"configuration=c1",
				"configuration=c2",
			},
			expectNil: true,
		},
		{
			description: "selects pending configuration",
			agentLabels: "",
			configurationsLabels: []string{
				"configuration=c0",
				"configuration=c1",
				"configuration=c2",
			},
			pendingConfig:            "c1:1",
			expectConfigurationIndex: 1,
		},
		{
			description: "selects pending configuration when current is set",
			agentLabels: "",
			configurationsLabels: []string{
				"configuration=c0",
				"configuration=c1",
				"configuration=c2",
			},
			pendingConfig:            "c2:1",
			currentConfig:            "c1:1",
			expectConfigurationIndex: 2,
		},
		{
			description: "selects current configuration when pending isn't set",
			agentLabels: "",
			configurationsLabels: []string{
				"configuration=c0",
				"configuration=c1",
				"configuration=c2",
			},
			currentConfig:            "c2:1",
			expectConfigurationIndex: 2,
		},
		{
			description:          "no configs",
			agentLabels:          "",
			configurationsLabels: []string{},
			expectNil:            true,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			// Setup
			store.Clear()
			if setupFunc != nil {
				setupFunc(store)
			}

			// create the agent
			agentID := "id"
			a, err := store.UpsertAgent(ctx, agentID, func(current *model.Agent) {
				labels, err := model.LabelsFromSelector(test.agentLabels)
				require.NoError(t, err)
				current.Labels = labels
				current.ConfigurationStatus.Pending = test.pendingConfig
				current.ConfigurationStatus.Current = test.currentConfig
			})

			// create the configurations
			for i, configurationLabels := range test.configurationsLabels {
				labels, err := model.LabelsFromSelector(configurationLabels)
				require.NoError(t, err)

				config := model.NewRawConfiguration(fmt.Sprintf("c%d", i), "")
				config.Spec.Selector.MatchLabels = labels.AsMap()

				status, err := store.ApplyResources(ctx, []model.Resource{config})
				require.NoError(t, err)
				require.Equal(t, model.StatusCreated, status[0].Status, status[0].Reason)
			}

			// find the match
			config, err := store.AgentConfiguration(ctx, a)
			require.NoError(t, err)
			if test.expectNil {
				require.Nil(t, config)
			} else {
				require.NotNil(t, config)
				require.Equal(t, fmt.Sprintf("c%d", test.expectConfigurationIndex), config.Name())
			}
			if test.expectFutureConfig != "" {
				require.Equal(t, test.expectFutureConfig, a.ConfigurationStatus.Future)
			}
		})
	}
}

func runDeleteChannelTests(t *testing.T, store Store, done chan bool) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("delete configuration, expect configuration-1 in deleteconfigurations channel", func(t *testing.T) {
		updates, unsubscribe := eventbus.Subscribe(ctx, store.Updates(ctx))
		defer unsubscribe()
		go verifyUpdates(t, done, updates, []configurationChanges{
			expectedUpdates(testConfiguration.Name()),
			expectedRemoves(testConfiguration.Name()),
		})

		// seed
		store.Clear()
		applyTestConfiguration(t, store)
		// delete the configuration
		_, err := store.DeleteResources(ctx, []model.Resource{
			testConfiguration,
		})
		require.NoError(t, err)

		ok := <-done
		require.True(t, ok)
	})

	t.Run("batch delete a single configuration, expect configuration-1 in deleteconfigurations channel", func(t *testing.T) {
		updates, unsubscribe := eventbus.Subscribe(ctx, store.Updates(ctx))
		defer unsubscribe()
		go verifyUpdates(t, done, updates, []configurationChanges{
			expectedUpdates(testConfiguration.Name()),
			expectedRemoves(testConfiguration.Name()),
		})

		// seed
		store.Clear()
		applyTestConfiguration(t, store)
		_, err := store.DeleteResources(ctx, []model.Resource{
			testConfiguration,
		})

		require.NoError(t, err)

		ok := <-done
		require.True(t, ok)
	})

	t.Run("batch delete a source attached to a configuration expect source in-use status", func(t *testing.T) {
		updates, unsubscribe := eventbus.Subscribe(ctx, store.Updates(ctx))
		defer unsubscribe()
		go verifyUpdates(t, done, updates, []configurationChanges{
			expectedUpdates(testConfiguration.Name()),
		})

		// seed
		store.Clear()
		applyTestConfiguration(t, store)
		statuses, err := store.DeleteResources(ctx, []model.Resource{
			macosSourceChanged,
		})
		assert.NoError(t, err, "expect no error on valid delete")
		require.ElementsMatch(t, []model.ResourceStatus{
			{
				Resource: macosSourceChanged,
				Status:   model.StatusInUse,
				Reason:   "Dependent resources:\nConfiguration configuration-1\n",
			},
		}, statuses)

		ok := <-done
		require.True(t, ok)
	})

	t.Run("batch delete source and its configuration, expect configuration-1 in channel", func(t *testing.T) {
		updates, unsubscribe := eventbus.Subscribe(ctx, store.Updates(ctx))
		defer unsubscribe()
		go verifyUpdates(t, done, updates, []configurationChanges{
			expectedUpdates(testConfiguration.Name()),
			expectedRemoves(testConfiguration.Name()),
		})

		// seed
		store.Clear()
		applyTestConfiguration(t, store)
		_, err := store.DeleteResources(ctx, []model.Resource{
			testConfiguration,
			macosSource,
		})
		require.NoError(t, err)

		ok := <-done
		require.True(t, ok)
	})
}

func runAgentSubscriptionsTest(t *testing.T, store Store) {
	agent := &model.Agent{
		ID:   "1",
		Name: "agent-1",
	}
	afterStatus := &model.Agent{
		ID:     "1",
		Name:   "agent-1",
		Status: 1,
	}

	tests := []struct {
		description     string
		updaterFunction AgentUpdater
		expect          []*model.Agent
	}{
		{
			description: "agent in channel after creation",
			updaterFunction: func(current *model.Agent) {
				*current = *agent
			},
			expect: []*model.Agent{agent},
		},
		{
			description: "agent in channel after changing status",
			updaterFunction: func(current *model.Agent) {
				*current = *afterStatus
			},
			expect: []*model.Agent{afterStatus},
		},
	}

	ctx := context.Background()
	channel, _ := eventbus.Subscribe(ctx, store.Updates(ctx))

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			done := make(chan bool)
			go verifyAgentChanges(t, done, channel, test.expect)

			_, err := store.UpsertAgent(context.TODO(), agent.ID, test.updaterFunction)
			require.NoError(t, err)

			ok := <-done
			assert.True(t, ok)
		})
	}
}

func verifyAgentChanges(t *testing.T, done chan bool, agentChanges <-chan BasicEventUpdates, expectedUpdates []*model.Agent) {
	for {
		select {
		case <-time.After(5 * time.Second):
			done <- false
			return
		case changes := <-agentChanges:
			agents := []*model.Agent{}
			for _, change := range changes.Agents() {
				agents = append(agents, change.Item)
			}
			if !assert.ElementsMatch(t, expectedUpdates, agents) {
				done <- false
				return
			}

			done <- true
			return
		}
	}
}

func verifyAgentUpdates(t *testing.T, done chan bool, agentChanges <-chan BasicEventUpdates, expectedUpdates []string) {
	for {
		select {
		case <-time.After(5 * time.Second):
			done <- false
			return
		case changes := <-agentChanges:
			ids := []string{}
			for _, change := range changes.Agents() {
				ids = append(ids, change.Item.ID)
			}

			if !assert.ElementsMatch(t, ids, expectedUpdates) {
				done <- false
				return
			}

			done <- true
			return
		}
	}
}

func runUpdateAgentsTests(t *testing.T, store Store) {
	// Tests for UpsertAgent
	upsertAgentTests := []struct {
		description   string
		agent         *model.Agent
		updater       AgentUpdater
		expectUpdates []string
	}{
		{
			description:   "upsertAgent passes along updates",
			agent:         &model.Agent{ID: "1", Status: 0},
			updater:       func(current *model.Agent) { current.Status = 1 },
			expectUpdates: []string{"1"},
		},
	}

	done := make(chan bool)
	ctx := context.Background()
	channel, _ := eventbus.Subscribe(ctx, store.Updates(ctx))

	for _, test := range upsertAgentTests {
		t.Run(test.description, func(t *testing.T) {
			go verifyAgentUpdates(t, done, channel, test.expectUpdates)

			_, err := store.UpsertAgent(context.TODO(), test.agent.ID, test.updater)
			require.NoError(t, err)

			ok := <-done
			require.True(t, ok)
		})
	}

	// Tests for UpsertAgents (bulk)
	upsertAgentsTests := []struct {
		description   string
		agents        []*model.Agent
		updater       AgentUpdater
		expectUpdates []string
	}{
		{
			description: "upsertAgents passes along a single update",
			agents: []*model.Agent{
				{ID: "1"},
			},
			updater:       func(current *model.Agent) { current.Status = 1 },
			expectUpdates: []string{"1"},
		},
		{
			description: "upsertAgents passes along multiple updates in single message",
			agents: []*model.Agent{
				{ID: "1"},
				{ID: "2"},
				{ID: "3"},
				{ID: "4"},
				{ID: "5"},
				{ID: "6"},
				{ID: "7"},
				{ID: "8"},
				{ID: "9"},
				{ID: "10"},
			},
			updater:       func(current *model.Agent) { current.Status = 1 },
			expectUpdates: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
		},
	}

	for _, test := range upsertAgentsTests {
		t.Run(test.description, func(t *testing.T) {
			go verifyAgentUpdates(t, done, channel, test.expectUpdates)

			ids := make([]string, len(test.agents))
			for ix, a := range test.agents {
				ids[ix] = a.ID
			}

			_, err := store.UpsertAgents(context.TODO(), ids, test.updater)
			require.NoError(t, err)

			ok := <-done
			require.True(t, ok)
		})
	}
}

// These tests that the ApplyResources methods return the expected resources with statuses
func runApplyResourceReturnTests(t *testing.T, store Store) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// make sure these have the same ID since macosSourceChanged is just a modification of macosSource
	macosSourceChanged.SetID(macosSource.ID())

	tests := []struct {
		description string
		// initial resources to seed
		initialResources []model.Resource
		// resources to apply in the test call
		applyResources []model.Resource
		expect         []model.ResourceStatus
	}{
		{
			description:      "applies a single resource, returns created status",
			initialResources: []model.Resource{},
			applyResources:   []model.Resource{macosSource},
			expect:           []model.ResourceStatus{*model.NewResourceStatus(macosSource, model.StatusCreated)},
		},
		{
			description:      "applies a multiple new resources, returns all created statuses",
			initialResources: []model.Resource{},
			applyResources:   []model.Resource{macosSource, nginxSource, cabinDestination1},
			expect: []model.ResourceStatus{
				*model.NewResourceStatus(macosSource, model.StatusCreated),
				*model.NewResourceStatus(nginxSource, model.StatusCreated),
				*model.NewResourceStatus(cabinDestination1, model.StatusCreated),
			},
		},
		{
			description:      "applies resource to existing resource, returns status unchanged",
			initialResources: []model.Resource{macosSource},
			applyResources:   []model.Resource{macosSource},
			expect:           []model.ResourceStatus{*model.NewResourceStatus(macosSource, model.StatusUnchanged)},
		},
		{
			description:      "applies a changed resource to an existsting, returns status configured",
			initialResources: []model.Resource{macosSource},
			applyResources:   []model.Resource{macosSourceChanged},
			expect:           []model.ResourceStatus{*model.NewResourceStatus(macosSourceChanged, model.StatusConfigured)},
		},
		{
			description:      "applies mixed resource updates, returns correct statuses",
			initialResources: []model.Resource{macosSource, nginxSource},
			applyResources:   []model.Resource{macosSourceChanged, nginxSource},
			expect: []model.ResourceStatus{
				*model.NewResourceStatus(macosSourceChanged, model.StatusConfigured),
				*model.NewResourceStatus(nginxSource, model.StatusUnchanged),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			// Setup
			store.Clear()
			applyTestTypes(t, store)
			_, err := store.ApplyResources(ctx, test.initialResources)
			require.NoError(t, err, "expect no error in setup apply call")

			statuses, err := store.ApplyResources(ctx, test.applyResources)
			require.NoError(t, err, "expect no error in valid apply call")

			// apply is going to add the dependency versions to the resources, so we need to add them to the expected
			// resources before we compare. validate will do that automatically.
			for i, status := range test.expect {
				clone, err := model.Clone(status.Resource)
				require.NoError(t, err)
				clone.ValidateWithStore(ctx, store)
				status.Resource = clone
				test.expect[i] = status
			}

			assertResourceStatusesMatch(t, test.expect, statuses)
		})
	}
}

func runValidateApplyResourcesTests(t *testing.T, store Store) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tests := []struct {
		name      string
		resources []model.Resource
		reasons   []string
		statuses  []model.UpdateStatus
	}{
		{
			name:      "none",
			resources: []model.Resource{},
		},
		{
			name:      "all valid",
			resources: []model.Resource{macosSource, nginxSource},
			statuses:  []model.UpdateStatus{model.StatusCreated, model.StatusCreated},
			reasons:   []string{"", ""},
		},
		{
			name:      "one invalid",
			resources: []model.Resource{invalidSource},
			reasons:   []string{"_production-nginx-ingress_ is not a valid resource name"},
			statuses:  []model.UpdateStatus{model.StatusInvalid},
		},
		{
			name:      "two invalid of four",
			resources: []model.Resource{macosSource, invalidSource, invalidSource2, nginxSource},
			reasons:   []string{"", "_production-nginx-ingress_ is not a valid resource name", "foo/bar/baz is not a valid resource name", ""},
			statuses:  []model.UpdateStatus{model.StatusCreated, model.StatusInvalid, model.StatusInvalid, model.StatusCreated},
		},
		{
			name:      "invalid and unknown",
			resources: []model.Resource{invalidSource, &unknownResource},
			reasons:   []string{"_production-nginx-ingress_ is not a valid resource name", "not-a-real-resource is not a valid resource kind"},
			statuses:  []model.UpdateStatus{model.StatusInvalid, model.StatusInvalid},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store.Clear()
			_, err := store.ApplyResources(ctx, []model.Resource{
				macosSourceType,
				nginxSourceType,
				cabinDestinationType,
			})
			require.NoError(t, err)
			result, err := store.ApplyResources(ctx, test.resources)
			require.NoError(t, err)
			for i, status := range test.statuses {
				require.Equal(t, status, result[i].Status, result[i].Reason)
				require.Contains(t, result[i].Reason, test.reasons[i])
			}
		})
	}
}

func runDeleteResourcesReturnTests(t *testing.T, store Store) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tests := []struct {
		description      string
		initialResources []model.Resource
		deleteResources  []model.Resource
		expect           []model.ResourceStatus
	}{
		{
			description:      "calling delete on a non existent resource returns no resource status",
			initialResources: make([]model.Resource, 0),
			deleteResources:  []model.Resource{nginxSource},
			expect:           make([]model.ResourceStatus, 0),
		},
		{
			description:      "calling delete on an existing resource returns a single resource status",
			initialResources: []model.Resource{macosSource},
			deleteResources:  []model.Resource{macosSource},
			expect: []model.ResourceStatus{
				*model.NewResourceStatus(macosSource, model.StatusDeleted),
			},
		},
		{
			description:      "calling delete on one existing and one non existent resource returns single resource status",
			initialResources: []model.Resource{macosSource},
			deleteResources:  []model.Resource{macosSource, nginxSource},
			expect: []model.ResourceStatus{
				*model.NewResourceStatus(macosSource, model.StatusDeleted),
			},
		},
		{
			description:      "calling delete on multiple resources returns all resources deleted",
			initialResources: []model.Resource{macosSource, nginxSource, cabinDestination1},
			deleteResources:  []model.Resource{macosSource, nginxSource, cabinDestination1},
			expect: []model.ResourceStatus{
				*model.NewResourceStatus(macosSource, model.StatusDeleted),
				*model.NewResourceStatus(nginxSource, model.StatusDeleted),
				*model.NewResourceStatus(cabinDestination1, model.StatusDeleted),
			},
		},
		{
			description:      "calling delete on an in use resources returns update with status In Use",
			initialResources: []model.Resource{macosSource, nginxSource, cabinDestination1, testConfiguration},
			deleteResources:  []model.Resource{macosSource},
			expect: []model.ResourceStatus{
				*model.NewResourceStatusWithReason(macosSource, model.StatusInUse, "Dependent resources:\nConfiguration configuration-1\n"),
			},
		},
		{
			description:      "calling delete on an in use resources and its dependency returns all deleted",
			initialResources: []model.Resource{macosSource, nginxSource, cabinDestination1, testConfiguration},
			deleteResources:  []model.Resource{testConfiguration, cabinDestination1},
			expect: []model.ResourceStatus{
				*model.NewResourceStatus(cabinDestination1, model.StatusDeleted),
				*model.NewResourceStatus(testConfiguration, model.StatusDeleted),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			// setup
			store.Clear()
			applyTestTypes(t, store)
			_, err := store.ApplyResources(ctx, cloneResources(t, test.initialResources))
			require.NoError(t, err, "expect no error in seed apply")

			statuses, err := store.DeleteResources(ctx, test.deleteResources)
			require.NoError(t, err, "expect no error on valid delete call")

			assert.ElementsMatch(t, test.expect, statuses)
		})
	}
}

func runDependentResourcesTests(t *testing.T, s Store) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tests := []struct {
		description      string
		initialResources []model.Resource
		testResource     model.Resource
		expect           DependentResources
	}{
		{
			description: "macos source has configuration dependency",
			initialResources: []model.Resource{
				macosSourceType,
				macosSource,
				cabinDestinationType,
				cabinDestination1,
				testConfiguration,
			},
			testResource: macosSource,
			expect: DependentResources{
				{
					Name: testConfiguration.Name(),
					Kind: model.KindConfiguration,
				},
			},
		},
		{
			description: "cabin destination has configuration dependency",
			initialResources: []model.Resource{
				macosSourceType,
				macosSource,
				cabinDestinationType,
				cabinDestination1,
				testConfiguration,
			},
			testResource: cabinDestination1,
			expect: DependentResources{
				{
					Name: testConfiguration.Name(),
					Kind: model.KindConfiguration,
				},
			},
		},
	}

	for _, test := range tests {
		updates, err := s.ApplyResources(ctx, test.initialResources)
		fmt.Println("UPDATES: ", updates)

		dependencies, err := FindDependentResources(ctx, s.ConfigurationIndex(ctx), test.testResource.Name(), test.testResource.GetKind())
		require.NoError(t, err)
		assert.Equal(t, test.expect, dependencies)
	}
}

func runIndividualDeleteTests(t *testing.T, store Store) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	setup := func() {
		store.Clear()
		_, err := store.ApplyResources(ctx, []model.Resource{
			macosSourceType,
			macosSource,
			nginxSourceType,
			nginxSource,
			cabinDestinationType,
			cabinDestination1,
			cabinDestination2,
			testConfiguration})
		require.NoError(t, err)
	}

	t.Run("DeleteConfiguration", func(t *testing.T) {
		tests := []struct {
			description         string
			source              string
			expectError         error
			expectConfiguration *model.Configuration
		}{
			{
				description:         "delete configuration-1",
				source:              testConfiguration.Name(),
				expectError:         nil,
				expectConfiguration: testConfiguration,
			},
		}
		for _, test := range tests {
			setup()

			src, err := store.DeleteConfiguration(ctx, test.source)
			assertResourcesEqual(t, test.expectConfiguration, src, test.description)
			assert.Equal(t, test.expectError, err, test.description)
		}
	})

	t.Run("DeleteSource", func(t *testing.T) {
		tests := []struct {
			description  string
			source       string
			expectError  error
			expectSource *model.Source
		}{
			{
				description:  "delete nginx",
				source:       nginxSource.Name(),
				expectError:  nil,
				expectSource: nginxSource,
			},
			{
				description: "delete macos, get dependency error",
				source:      macosSource.Name(),
				expectError: NewDependencyError(DependentResources{
					Dependency{Name: testConfiguration.Name(),
						Kind: model.KindConfiguration},
				}),
				expectSource: nil,
			},
			{
				description:  "delete non existent, no resource, no error",
				source:       "foo",
				expectError:  nil,
				expectSource: nil,
			},
		}

		for _, test := range tests {
			setup()

			src, err := store.DeleteSource(ctx, test.source)
			assertResourcesEqual(t, test.expectSource, src)
			assert.Equal(t, test.expectError, err)
		}
	})

	t.Run("DeleteDestination", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		tests := []struct {
			description       string
			destination       string
			expectError       error
			expectDestination *model.Destination
		}{
			{
				description:       "delete cabinDestination2",
				destination:       cabinDestination2.Name(),
				expectError:       nil,
				expectDestination: cabinDestination2,
			},
			{
				description: "delete cabinDestination1, expect error",
				destination: cabinDestination1.Name(),
				expectError: NewDependencyError(DependentResources{
					Dependency{
						Name: testConfiguration.Name(),
						Kind: model.KindConfiguration,
					},
				}),
				expectDestination: nil,
			},
			{
				description:       "delete non existent, expect nil error and  destination",
				destination:       "foo",
				expectError:       nil,
				expectDestination: nil,
			},
		}

		for _, test := range tests {
			setup()

			dest, err := store.DeleteDestination(ctx, test.destination)
			assertResourcesEqual(t, test.expectDestination, dest)
			assert.Equal(t, test.expectError, err)
		}
	})
}

func verifyAgentsRemove(t *testing.T, done chan bool, Updates <-chan BasicEventUpdates, expectRemoves []string) {
	var val struct{}
	removesRemaining := map[string]struct{}{}
	for _, r := range expectRemoves {
		removesRemaining[r] = val
	}
	for {
		select {
		case <-time.After(5 * time.Second):
			done <- false
			t.Log("Timed out waiting for updates.")
			return
		case updates, ok := <-Updates:
			if !ok {
				done <- false
				return
			}
			agentUpdates := updates.Agents()

			// skip when we're seeding
			var skip = false
			for _, update := range agentUpdates {
				if update.Type != EventTypeRemove {
					skip = true
				}
			}
			if skip {
				continue
			}

			for _, update := range updates.Agents() {
				if update.Type == EventTypeRemove {
					delete(removesRemaining, update.Item.ID)
				}
			}

			if len(removesRemaining) == 0 {
				done <- true
				return
			}
		}
	}
}

// runDeleteAgentsTests tests store.DeleteAgents
func runDeleteAgentsTests(t *testing.T, store Store) {
	deleteTests := []struct {
		description    string
		seedAgentsIDs  []string
		deleteAgentIDs []string
		// The agents returned by the delete method
		expectDeleted []*model.Agent
		// The agents returned by the store
		expectAgents []*model.Agent
	}{
		{
			description:    "delete 1 agent",
			seedAgentsIDs:  []string{"1"},
			deleteAgentIDs: []string{"1"},
			expectDeleted: []*model.Agent{
				{ID: "1", Status: 5, Labels: model.MakeLabels()},
			},
			// The agents left in the store after delete
			expectAgents: make([]*model.Agent, 0),
		},
		{
			description:    "delete multiple agents",
			seedAgentsIDs:  []string{"1", "2", "3", "4", "5"},
			deleteAgentIDs: []string{"1", "2", "3"},
			expectDeleted: []*model.Agent{
				{ID: "1", Status: 5, Labels: model.MakeLabels()},
				{ID: "2", Status: 5, Labels: model.MakeLabels()},
				{ID: "3", Status: 5, Labels: model.MakeLabels()},
			},
			expectAgents: []*model.Agent{
				{ID: "4", Labels: model.MakeLabels()},
				{ID: "5", Labels: model.MakeLabels()},
			},
		},
		{
			description:    "delete non existing agent, no error, no delete",
			seedAgentsIDs:  []string{"1"},
			deleteAgentIDs: []string{"42"},
			expectDeleted:  make([]*model.Agent, 0),
			expectAgents:   []*model.Agent{{ID: "1", Labels: model.MakeLabels()}},
		},
	}

	// Test the delete operation
	for _, test := range deleteTests {
		// setup
		ctx := context.Background()
		store.Clear()

		// seed agents
		for _, id := range test.seedAgentsIDs {
			addAgent(store, &model.Agent{ID: id, Labels: model.MakeLabels()})
		}

		t.Run(test.description, func(t *testing.T) {
			deleted, err := store.DeleteAgents(ctx, test.deleteAgentIDs)
			require.NoError(t, err)
			assert.ElementsMatch(t, test.expectDeleted, deleted, "deleted agents do not match")

			rest, err := store.Agents(ctx)
			require.NoError(t, err)
			assert.ElementsMatch(t, test.expectAgents, rest, "remaining agents do not match")
		})
	}

	t.Run("deleting an agent removes it from the index", func(t *testing.T) {
		// setup
		store.Clear()
		ctx := context.Background()

		// seed agent
		addAgent(store, &model.Agent{ID: "1"})

		// verify its in the index
		results, err := search.Field(ctx, store.AgentIndex(ctx), "id", "1")
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"1"}, results)

		// delete it
		_, err = store.DeleteAgents(ctx, []string{"1"})
		require.NoError(t, err)

		results, err = search.Field(ctx, store.AgentIndex(ctx), "id", "1")
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{}, results)
	})

	deleteUpdatesTests := []struct {
		description    string
		seedAgentIDs   []string
		deleteAgentIDs []string
	}{
		{
			description:    "delete an agent, expect a remove update in Updates.Agents",
			seedAgentIDs:   []string{"1"},
			deleteAgentIDs: []string{"1"},
		},
	}

	for _, test := range deleteUpdatesTests {
		t.Run(test.description, func(t *testing.T) {
			// setup
			store.Clear()
			for _, id := range test.seedAgentIDs {
				addAgent(store, &model.Agent{ID: id})
			}

			ctx := context.Background()
			channel, unsubscribe := eventbus.Subscribe(ctx, store.Updates(ctx))
			defer unsubscribe()

			done := make(chan bool, 0)
			go verifyAgentsRemove(t, done, channel, test.deleteAgentIDs)

			_, err := store.DeleteAgents(ctx, test.deleteAgentIDs)
			require.NoError(t, err)

			ok := <-done
			assert.True(t, ok)
		})
	}
}

// runConfigurationsTests runs tests on Store.Configuration and Store.Configurations
func runConfigurationsTests(t *testing.T, store Store) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("lists all configurations", func(t *testing.T) {
		// Setup
		status, err := store.ApplyResources(ctx, []model.Resource{testRawConfiguration1, testRawConfiguration2})
		require.NoError(t, err)
		requireOkStatuses(t, status)

		configs, err := store.Configurations(ctx)
		assert.NoError(t, err)
		assertResourcesMatch(t, []*model.Configuration{testRawConfiguration1, testRawConfiguration2}, configs)
	})
}

func runConfigurationTests(t *testing.T, store Store) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("gets configuration by name", func(t *testing.T) {
		// Setup
		status, err := store.ApplyResources(ctx, []model.Resource{testRawConfiguration1, testRawConfiguration2})
		require.NoError(t, err)
		requireOkStatuses(t, status)

		config, err := store.Configuration(ctx, testRawConfiguration1.Name())
		assert.NoError(t, err)
		assertResourcesEqual(t, testRawConfiguration1, config)
	})
}

func runPagingTests(t *testing.T, store Store) {
	for i := 0; i < 100; i++ {
		store.UpsertAgent(context.TODO(), fmt.Sprintf("%03d", i), func(current *model.Agent) {
			current.Name = "agent-" + current.ID
		})
	}
	tests := []struct {
		name      string
		offset    int
		limit     int
		expectIDs []string
	}{
		{
			name:   "first page",
			offset: 0,
			limit:  10,
			expectIDs: []string{
				"agent-000",
				"agent-001",
				"agent-002",
				"agent-003",
				"agent-004",
				"agent-005",
				"agent-006",
				"agent-007",
				"agent-008",
				"agent-009",
			},
		},
		{
			name:   "second page",
			offset: 10,
			limit:  10,
			expectIDs: []string{
				"agent-010",
				"agent-011",
				"agent-012",
				"agent-013",
				"agent-014",
				"agent-015",
				"agent-016",
				"agent-017",
				"agent-018",
				"agent-019",
			},
		},
		{
			name:   "last few",
			offset: 95,
			limit:  10,
			expectIDs: []string{
				"agent-095",
				"agent-096",
				"agent-097",
				"agent-098",
				"agent-099",
			},
		},
		{
			name:      "page too large",
			offset:    200,
			limit:     10,
			expectIDs: []string{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			agents, err := store.Agents(context.TODO(), WithOffset(test.offset), WithLimit(test.limit))
			require.NoError(t, err)
			ids := []string{}
			for _, agent := range agents {
				ids = append(ids, agent.Name)
			}
			require.ElementsMatch(t, test.expectIDs, ids)
		})
	}
	t.Run("agents count", func(t *testing.T) {
		count, err := store.AgentsCount(context.TODO())
		require.NoError(t, err)
		require.Equal(t, 100, count)
	})
}

func runTestUpsertAgents(t *testing.T, store Store) {
	t.Run("can insert new agents", func(t *testing.T) {
		store.Clear()
		count, err := store.AgentsCount(context.TODO())
		require.NoError(t, err)
		require.Zero(t, count)

		returnedAgents, err := store.UpsertAgents(
			context.TODO(),
			[]string{"1", "2", "3"},
			func(current *model.Agent) {
				current.Labels = model.MakeLabels()
			},
		)
		require.NoError(t, err)

		expectAgents := []*model.Agent{
			{ID: "1", Labels: model.MakeLabels()},
			{ID: "2", Labels: model.MakeLabels()},
			{ID: "3", Labels: model.MakeLabels()},
		}

		require.ElementsMatch(t, expectAgents, returnedAgents)

		gotAgents, err := store.Agents(context.TODO())
		require.NoError(t, err)
		require.ElementsMatch(t, expectAgents, gotAgents)
	})

	t.Run("upserts and updates agents correctly", func(t *testing.T) {
		tests := []struct {
			description    string
			initAgentsIDs  []string
			upsertAgentIDs []string
			updater        AgentUpdater
			expectAgents   []*model.Agent
		}{
			{
				description:    "updates existing agents and inserts",
				initAgentsIDs:  []string{"1"},
				upsertAgentIDs: []string{"1", "2"},
				updater:        func(current *model.Agent) { current.Status = 1; current.Labels = model.MakeLabels() },
				expectAgents: []*model.Agent{
					{ID: "1", Status: 1, Labels: model.MakeLabels()},
					{ID: "2", Status: 1, Labels: model.MakeLabels()},
				},
			},
		}

		for _, test := range tests {
			t.Run(test.description, func(t *testing.T) {
				// setup
				store.Clear()

				// seed agents
				for _, id := range test.initAgentsIDs {
					addAgent(store, &model.Agent{ID: id, Labels: model.MakeLabels()})
				}

				// upsert
				returnedAgents, err := store.UpsertAgents(context.TODO(), test.upsertAgentIDs, test.updater)
				require.NoError(t, err)
				require.ElementsMatch(t, test.expectAgents, returnedAgents)

				// verify
				gotAgents, err := store.Agents(context.TODO())
				require.NoError(t, err)
				require.ElementsMatch(t, test.expectAgents, gotAgents)

			})
		}
	})

}

func runTestUpsertAgent(ctx context.Context, t *testing.T, s Store) {
	// Seed with one
	a1 := &model.Agent{ID: "1", Name: "Fake Agent 1", Labels: model.Labels{Set: model.MakeLabels().Set}}
	addAgent(s, a1)

	t.Run("creates a new agent if not found", func(t *testing.T) {
		newAgentID := "3"
		s.UpsertAgent(ctx, newAgentID, testUpdater)

		got, err := s.Agent(ctx, newAgentID)
		require.NoError(t, err)

		assert.NotNil(t, got)
		assert.Equal(t, got.ID, newAgentID)
	})
	t.Run("calls updater and updates an agent if exists", func(t *testing.T) {
		updaterCalled = false
		s.UpsertAgent(context.TODO(), a1.ID, testUpdater)

		assert.True(t, updaterCalled)

		got, err := s.Agent(ctx, a1.ID)
		require.NoError(t, err)

		assert.Equal(t, got.Name, "updated")
	})
}

func runTestUpdateAgent(ctx context.Context, t *testing.T, s Store) {
	// Seed with one
	a1 := &model.Agent{ID: "1", Name: "Fake Agent 1", Labels: model.Labels{Set: model.MakeLabels().Set}}
	addAgent(s, a1)

	t.Run("update does nothing if not found", func(t *testing.T) {
		newAgentID := "3"
		s.UpdateAgent(ctx, newAgentID, testUpdater)

		got, err := s.Agent(ctx, newAgentID)
		require.NoError(t, err)
		require.Nil(t, got)
	})
	t.Run("calls updater and updates an agent if exists", func(t *testing.T) {
		updaterCalled = false
		s.UpdateAgent(context.TODO(), a1.ID, testUpdater)

		assert.True(t, updaterCalled)

		got, err := s.Agent(ctx, a1.ID)
		require.NoError(t, err)

		assert.Equal(t, got.Name, "updated")
	})
}

func runTestUpdateAgents(t *testing.T, store Store) {
	t.Run("update does not insert new agents", func(t *testing.T) {
		store.Clear()
		count, err := store.AgentsCount(context.TODO())
		require.NoError(t, err)
		require.Zero(t, count)

		returnedAgents, err := store.UpdateAgents(
			context.TODO(),
			[]string{"1", "2", "3"},
			func(current *model.Agent) {
				current.Labels = model.MakeLabels()
			},
		)
		require.NoError(t, err)

		expectAgents := []*model.Agent{}

		require.ElementsMatch(t, expectAgents, returnedAgents)

		gotAgents, err := store.Agents(context.TODO())
		require.NoError(t, err)
		require.ElementsMatch(t, expectAgents, gotAgents)
	})

	t.Run("update only updates existing agents", func(t *testing.T) {
		tests := []struct {
			description    string
			initAgentsIDs  []string
			upsertAgentIDs []string
			updater        AgentUpdater
			expectAgents   []*model.Agent
		}{
			{
				description:    "updates existing agents",
				initAgentsIDs:  []string{"1"},
				upsertAgentIDs: []string{"1", "2"},
				updater:        func(current *model.Agent) { current.Status = 1; current.Labels = model.MakeLabels() },
				expectAgents: []*model.Agent{
					{ID: "1", Status: 1, Labels: model.MakeLabels()},
				},
			},
		}

		for _, test := range tests {
			t.Run(test.description, func(t *testing.T) {
				// setup
				store.Clear()

				// seed agents
				for _, id := range test.initAgentsIDs {
					addAgent(store, &model.Agent{ID: id, Labels: model.MakeLabels()})
				}

				// upsert
				returnedAgents, err := store.UpdateAgents(context.TODO(), test.upsertAgentIDs, test.updater)
				require.NoError(t, err)
				require.ElementsMatch(t, test.expectAgents, returnedAgents)

				// verify
				gotAgents, err := store.Agents(context.TODO())
				require.NoError(t, err)
				require.ElementsMatch(t, test.expectAgents, gotAgents)

			})
		}
	})

}

func cloneResources[T model.Resource](t *testing.T, resources []T) []T {
	clones := make([]T, len(resources))
	for ix, r := range resources {
		clone, err := model.Clone(r)
		require.NoError(t, err)
		clones[ix] = clone
	}
	return clones
}

// ----------------------------------------------------------------------

func requireOkStatuses(t *testing.T, statuses []model.ResourceStatus) {
	for _, status := range statuses {
		require.Contains(t, []model.UpdateStatus{
			model.StatusUnchanged,
			model.StatusConfigured,
			model.StatusCreated,
			model.StatusDeleted,
		}, status.Status)
	}
}

// ----------------------------------------------------------------------

func runTestMeasurements(t *testing.T, store Store) {
	var ctx context.Context
	var measurements stats.Measurements

	reset := func() {
		ctx = context.TODO()

		store.Clear()
		measurements = store.Measurements()
	}

	// Mock out current time to executing the test near the end of a 10s bucket doesn't give false negatives
	// and ensure that the longer time period tests have consistent times to generate consistent rates
	now := time.Now().Truncate(24 * time.Hour).UTC()
	getCurrentTime = func() time.Time { return now }

	// Use epochStart as startTime for counters when we're not testing rollover
	epochStart := time.Unix(0, 0).UTC()

	saveMetrics := func(name string, timestamp, startTime time.Time, interval time.Duration, configuration, agent, processor string, values []float64) {
		saveTestMetrics(ctx, t, measurements, name, timestamp, startTime, interval, configuration, agent, processor, values)
	}

	t.Run("returns empty measurements with no data", func(t *testing.T) {
		reset()

		aMetrics, err := measurements.AgentMetrics(ctx, []string{"a3", "a4"})
		require.NoError(t, err)
		require.Len(t, aMetrics, 0)

		for _, configuration := range []string{"c3", "c4"} {
			metrics, err := measurements.ConfigurationMetrics(ctx, configuration)
			require.NoError(t, err)
			require.Len(t, metrics, 0)
		}
	})

	t.Run("returns measurements", func(t *testing.T) {
		reset()

		status, err := store.ApplyResources(context.Background(), []model.Resource{testRawConfiguration1})
		require.NoError(t, err)
		requireOkStatuses(t, status)

		_, err = store.UpsertAgents(context.Background(), []string{a1, a2}, func(current *model.Agent) {})
		require.NoError(t, err)

		frame := now.Truncate(10 * time.Second)
		prev := frame.Add(-10 * time.Second)
		timestamp := frame.Add(-3 * time.Minute)

		saveMetrics(stats.LogDataSizeMetricName, timestamp, epochStart, 10*time.Second, c1, a1, p1, []float64{0, 0, 0, 0, 0, 0, 10, 10, 10, 10, 10, 10, 90, 90, 90, 90, 90, 100, 110})

		saveMetrics(stats.LogDataSizeMetricName, timestamp, epochStart, 10*time.Second, c1, a1, p1, []float64{0, 0, 0, 0, 0, 0, 10, 10, 10, 10, 10, 10, 90, 90, 90, 90, 90, 100, 110})
		saveMetrics(stats.LogDataSizeMetricName, timestamp, epochStart, 10*time.Second, c1, a1, p2, []float64{0, 0, 0, 0, 0, 0, 20, 20, 20, 20, 20, 20, 200, 200, 200, 200, 200, 200, 220})
		saveMetrics(stats.LogDataSizeMetricName, timestamp, epochStart, 10*time.Second, c1, a2, p1, []float64{0, 0, 0, 0, 0, 0, 100, 100, 100, 100, 100, 100, 1000, 1000, 1000, 1000, 1000, 1000, 10000})
		saveMetrics(stats.LogDataSizeMetricName, timestamp, epochStart, 10*time.Second, c1, a2, p2, []float64{0, 0, 0, 0, 0, 0, 200, 200, 200, 200, 200, 200, 2000, 2000, 2000, 2000, 2000, 2000, 20000})

		err = measurements.ProcessMetrics(ctx)
		require.NoError(t, err)

		t.Run("returns measurements for just one agent", func(t *testing.T) {
			metrics, err := measurements.AgentMetrics(ctx, []string{a1}, stats.WithPeriod(time.Minute))
			require.NoError(t, err)
			requireMetrics(t, []*record.Metric{
				generateExpectMetric(stats.LogDataSizeMetricName, prev, epochStart, c1, a1, p1, 1.5),
				generateExpectMetric(stats.LogDataSizeMetricName, prev, epochStart, c1, a1, p2, 3),
			}, metrics)

			metrics, err = measurements.AgentMetrics(ctx, []string{a2}, stats.WithPeriod(time.Minute))
			require.NoError(t, err)
			requireMetrics(t, []*record.Metric{
				generateExpectMetric(stats.LogDataSizeMetricName, prev, epochStart, c1, a2, p1, 15),
				generateExpectMetric(stats.LogDataSizeMetricName, prev, epochStart, c1, a2, p2, 30),
			}, metrics)
		})

		t.Run("returns measurements for multiple agents", func(t *testing.T) {
			metrics, err := measurements.AgentMetrics(ctx, []string{a1, a2}, stats.WithPeriod(time.Minute))
			require.NoError(t, err)
			requireMetrics(t, []*record.Metric{
				generateExpectMetric(stats.LogDataSizeMetricName, prev, epochStart, c1, a1, p1, 1.5),
				generateExpectMetric(stats.LogDataSizeMetricName, prev, epochStart, c1, a1, p2, 3),
				generateExpectMetric(stats.LogDataSizeMetricName, prev, epochStart, c1, a2, p1, 15),
				generateExpectMetric(stats.LogDataSizeMetricName, prev, epochStart, c1, a2, p2, 30),
			}, metrics)
		})

		t.Run("returns measurements for all agents", func(t *testing.T) {
			metrics, err := measurements.AgentMetrics(ctx, []string{}, stats.WithPeriod(time.Minute))
			require.NoError(t, err)
			requireMetrics(t, []*record.Metric{
				generateExpectMetric(stats.LogDataSizeMetricName, prev, epochStart, c1, a1, p1, 1.5),
				generateExpectMetric(stats.LogDataSizeMetricName, prev, epochStart, c1, a1, p2, 3),
				generateExpectMetric(stats.LogDataSizeMetricName, prev, epochStart, c1, a2, p1, 15),
				generateExpectMetric(stats.LogDataSizeMetricName, prev, epochStart, c1, a2, p2, 30),
			}, metrics)
		})

		t.Run("returns configuration measurements", func(t *testing.T) {
			metrics, err := measurements.ConfigurationMetrics(ctx, c1, stats.WithPeriod(time.Minute))
			require.NoError(t, err)
			requireMetrics(t, []*record.Metric{
				generateExpectMetric(stats.LogDataSizeMetricName, prev, epochStart, c1, "", p1, 16.5),
				generateExpectMetric(stats.LogDataSizeMetricName, prev, epochStart, c1, "", p2, 33),
			}, metrics)
		})

		t.Run("returns overview measurements", func(t *testing.T) {
			metrics, err := measurements.OverviewMetrics(ctx, stats.WithPeriod(time.Minute))
			require.NoError(t, err)
			requireMetrics(t, []*record.Metric{
				generateExpectMetric(stats.LogDataSizeMetricName, prev, epochStart, c1, "", p1, 16.5),
				generateExpectMetric(stats.LogDataSizeMetricName, prev, epochStart, c1, "", p2, 33),
			}, metrics)
		})
	})

	t.Run("looks backwards if there is an incomplete bucket at -10s", func(t *testing.T) {
		reset()

		frame := now.Truncate(10 * time.Second)
		expectedTimestamp := frame.Add(-20 * time.Second)
		timestamp := frame.Add(-3 * time.Minute)

		saveMetrics(stats.LogDataSizeMetricName, timestamp, epochStart, 10*time.Second, c1, a1, p1, []float64{0, 0, 0, 0, 0, 0, 10, 10, 10, 10, 10, 10, 90, 90, 90, 90, 90, 100})
		saveMetrics(stats.LogDataSizeMetricName, timestamp, epochStart, 10*time.Second, c1, a1, p2, []float64{0, 0, 0, 0, 0, 0, 20, 20, 20, 20, 20, 20, 200, 200, 200, 200, 200})
		saveMetrics(stats.LogDataSizeMetricName, timestamp, epochStart, 10*time.Second, c1, a2, p1, []float64{0, 0, 0, 0, 0, 0, 100, 100, 100, 100, 100, 100, 1000, 1000, 1000, 1000, 1000})
		saveMetrics(stats.LogDataSizeMetricName, timestamp, epochStart, 10*time.Second, c1, a2, p2, []float64{0, 0, 0, 0, 0, 0, 200, 200, 200, 200, 200, 200, 2000, 2000, 2000, 2000, 2000})

		metrics, err := measurements.AgentMetrics(ctx, []string{a1}, stats.WithPeriod(time.Minute))
		require.NoError(t, err)
		requireMetrics(t, []*record.Metric{
			generateExpectMetric(stats.LogDataSizeMetricName, expectedTimestamp, epochStart, c1, a1, p1, 1.33),
			generateExpectMetric(stats.LogDataSizeMetricName, expectedTimestamp, epochStart, c1, a1, p2, 3),
		}, metrics)
	})

	t.Run("handles rollovers using the startTime of the end data point", func(t *testing.T) {
		reset()

		frame := now.Truncate(10 * time.Second)
		expectedTimestamp := frame.Add(-10 * time.Second)
		oldDataTimestamp := frame.Add(-25 * time.Hour)

		rolloverPoint := frame.Add(-1 * time.Hour)

		saveMetrics(stats.LogDataSizeMetricName, expectedTimestamp, rolloverPoint, 10*time.Second, c1, a1, p1, []float64{500})

		saveMetrics(stats.LogDataSizeMetricName, oldDataTimestamp, epochStart, 1*time.Hour, c1, a1, p1, []float64{300, 300, 300})

		metrics, err := measurements.AgentMetrics(ctx, []string{a1}, stats.WithPeriod(24*time.Hour))
		require.NoError(t, err)
		requireMetrics(t, []*record.Metric{
			generateExpectMetric(stats.LogDataSizeMetricName, expectedTimestamp, rolloverPoint, c1, a1, p1, 0.14),
		}, metrics)
	})

	t.Run("properly cleans up metrics", func(t *testing.T) {
		reset()

		// Metrics completely out of scope of measurements, should all be removed
		saveMetrics(stats.LogDataSizeMetricName, now.Add(-60*24*time.Hour).Truncate(24*time.Hour), epochStart, 1*time.Hour, c1, a1, p1, []float64{0, 0, 0, 0, 0, 0})
		// Metrics in the last 31 days, only daily metrics saved - 3 should be saved, 3 deleted
		saveMetrics(stats.LogDataSizeMetricName, now.Add(-30*24*time.Hour).Truncate(24*time.Hour), epochStart, 12*time.Hour, c1, a1, p1, []float64{0, 0, 0, 0, 0, 0})
		// Metrics in the last 1 day, only hourly metrics saved - 3 should be saved, 3 deleted
		saveMetrics(stats.LogDataSizeMetricName, now.Add(-23*time.Hour).Truncate(1*time.Hour), epochStart, 30*time.Minute, c1, a1, p1, []float64{0, 0, 0, 0, 0, 0})
		// Metrics in the last 6 hours, only 5min metrics saved - 2 should be saved, 4 deleted
		saveMetrics(stats.LogDataSizeMetricName, now.Add(-5*time.Hour).Truncate(1*time.Hour), epochStart, 2*time.Minute, c1, a1, p1, []float64{0, 0, 0, 0, 0, 0})
		// Metrics in the last 10 minutes, only 1min metrics saved - 3 should be saved, 3 deleted
		saveMetrics(stats.LogDataSizeMetricName, now.Add(-9*time.Minute).Truncate(1*time.Minute), epochStart, 30*time.Second, c1, a1, p1, []float64{0, 0, 0, 0, 0, 0})
		// Metrics in the last 100 seconds, all 6 saved
		saveMetrics(stats.LogDataSizeMetricName, now.Add(-90*time.Second).Truncate(10*time.Second), epochStart, 10*time.Second, c1, a1, p1, []float64{0, 0, 0, 0, 0, 0})

		// Each data point is written once to be looked up by Agent, once by Configuration, so we need to double the
		// number of metrics saved
		count, err := store.Measurements().MeasurementsSize(ctx)
		require.NoError(t, err)
		require.Equal(t, 72, count)

		err = measurements.ProcessMetrics(ctx)
		require.NoError(t, err)

		count, err = store.Measurements().MeasurementsSize(ctx)
		require.NoError(t, err)
		require.Equal(t, 34, count)
	})

	t.Run("handles more periods than just one minute", func(t *testing.T) {
		reset()

		saveMetrics(stats.LogDataSizeMetricName, now.Add(-25*time.Hour).Truncate(1*time.Hour), epochStart, 1*time.Hour, c1, a1, p1, []float64{1000, 1000, 1000})
		saveMetrics(stats.LogDataSizeMetricName, now.Add(-80*time.Minute).Truncate(10*time.Minute), epochStart, 10*time.Minute, c1, a1, p1, []float64{2000, 2000, 2000})
		saveMetrics(stats.LogDataSizeMetricName, now.Add(-6*time.Minute).Truncate(1*time.Minute), epochStart, 1*time.Minute, c1, a1, p1, []float64{3000, 3000, 3000})
		saveMetrics(stats.LogDataSizeMetricName, now.Add(-80*time.Second).Truncate(10*time.Second), epochStart, 10*time.Second, c1, a1, p1, []float64{4000, 4000, 4000})
		saveMetrics(stats.LogDataSizeMetricName, now.Add(-30*time.Second).Truncate(10*time.Second), epochStart, 10*time.Second, c1, a1, p1, []float64{5000, 5000, 5000})

		err := measurements.ProcessMetrics(ctx)
		require.NoError(t, err)

		frame := now.Truncate(10 * time.Second)
		prev := frame.Add(-10 * time.Second)

		metrics, err := measurements.AgentMetrics(ctx, []string{a1}, stats.WithPeriod(time.Minute))
		require.NoError(t, err)
		requireMetrics(t, []*record.Metric{
			generateExpectMetric(stats.LogDataSizeMetricName, prev, epochStart, c1, a1, p1, 16.67),
		}, metrics)

		metrics, err = measurements.AgentMetrics(ctx, []string{a1}, stats.WithPeriod(5*time.Minute))
		require.NoError(t, err)
		requireMetrics(t, []*record.Metric{
			generateExpectMetric(stats.LogDataSizeMetricName, prev, epochStart, c1, a1, p1, 5.71),
		}, metrics)

		metrics, err = measurements.AgentMetrics(ctx, []string{a1}, stats.WithPeriod(1*time.Hour))
		require.NoError(t, err)
		requireMetrics(t, []*record.Metric{
			generateExpectMetric(stats.LogDataSizeMetricName, prev, epochStart, c1, a1, p1, 0.84),
		}, metrics)

		metrics, err = measurements.AgentMetrics(ctx, []string{a1}, stats.WithPeriod(24*time.Hour))
		require.NoError(t, err)
		requireMetrics(t, []*record.Metric{
			generateExpectMetric(stats.LogDataSizeMetricName, prev, epochStart, c1, a1, p1, 0.05),
		}, metrics)
	})
}

func runTestCleanupDisconnectedAgents(t *testing.T, store Store) {
	t.Helper()

	ctx := context.Background()
	now := time.Now()
	t.Run("empty store", func(t *testing.T) {
		err := store.CleanupDisconnectedAgents(ctx, now)
		require.NoError(t, err)
	})

	t.Run("only removes containerized agents", func(t *testing.T) {
		disconnectedAt := now.Add(-1 * time.Hour)

		agentID1 := ulid.Make()
		agent1, err := store.UpsertAgent(ctx, agentID1.String(), func(current *model.Agent) {
			current.ID = agentID1.String()
			current.Labels = model.LabelsFromValidatedMap(map[string]string{
				model.LabelBindPlaneAgentArch: "amd64",
				model.LabelBindPlaneAgentName: "Agent " + agentID1.String(),
			})
			current.DisconnectedAt = &disconnectedAt
		})
		require.NoError(t, err)

		agentID2 := ulid.Make()
		_, err = store.UpsertAgent(ctx, agentID2.String(), func(current *model.Agent) {
			current.ID = agentID2.String()
			current.Labels = model.LabelsFromValidatedMap(map[string]string{
				model.LabelBindPlaneAgentArch:     "amd64",
				model.LabelBindPlaneAgentName:     "Agent " + agentID2.String(),
				model.LabelAgentContainerPlatform: "kubernetes-daemonset",
			})
			current.DisconnectedAt = &disconnectedAt
		})
		require.NoError(t, err)
		err = store.CleanupDisconnectedAgents(ctx, now)
		require.NoError(t, err)

		agents, err := store.Agents(ctx)
		require.NoError(t, err)
		require.Len(t, agents, 1)
		require.Equal(t, agent1.ID, agents[0].ID)
	})
}

// these are values that i happen to be testing with right now -andy
const (
	p1 = "throughputmeasurement/_s0_logs_source0"
	p2 = "throughputmeasurement/_s1_logs_source0"
	p3 = "throughputmeasurement/_d1_logs_info-logging"
	p4 = "throughputmeasurement/_d1_logs_error-logging"
	a1 = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	a2 = "01GE8Q0TFSFXYJTSHP8WKYYR2H"
	c1 = "Test-configuration-1"
	c2 = "localhost"
)

func generateExpectMetric(name string, timestamp, startTime time.Time, configuration, agent, processor string, value float64) *record.Metric {
	m := &record.Metric{
		Name:           name,
		StartTimestamp: startTime,
		Timestamp:      timestamp,
		Value:          value,
		Unit:           "B/s",
		Type:           "Rate",
		Attributes: map[string]interface{}{
			"configuration": configuration,
		},
	}

	// only add the processor if specified
	if processor != "" {
		m.Attributes["processor"] = processor
	}

	// only add the agent if specified
	if agent != "" {
		m.Attributes["agent"] = agent
	}
	return m
}

func requireMetrics(t *testing.T, expected, actual []*record.Metric) {
	require.ElementsMatch(t, expected, actual)
}

func saveTestMetrics(ctx context.Context, t *testing.T, m stats.Measurements, name string, timestamp, startTime time.Time, interval time.Duration, configuration, agent, processor string, values []float64) {
	err := m.SaveAgentMetrics(ctx, generateTestMetrics(t, name, timestamp, startTime, interval, configuration, agent, processor, values))
	require.NoError(t, err)
}

func generateTestMetric(_ *testing.T, name string, timestamp, startTime time.Time, configuration, agent, processor string, value float64) *record.Metric {
	m := &record.Metric{
		Name:           name,
		Timestamp:      timestamp,
		StartTimestamp: startTime,
		Value:          value,
		Unit:           "",
		Type:           "Sum",
		Attributes: map[string]interface{}{
			"configuration": configuration,
			"processor":     processor,
		},
	}
	// only add the agent if specified
	if agent != "" {
		m.Attributes["agent"] = agent
	}
	return m
}

// generateTestMetrics creates test metrics starting at timestamp with the specified interval. To create gaps in metrics, values < 0 are skipped.
func generateTestMetrics(t *testing.T, name string, timestamp, startTime time.Time, interval time.Duration, configuration, agent, processor string, values []float64) []*record.Metric {
	var results []*record.Metric
	for i := 0; i < len(values); i++ {
		results = append(results, generateTestMetric(
			t,
			name,
			timestamp.Add(time.Duration(i)*interval),
			startTime,
			configuration,
			agent,
			processor,
			values[i],
		))
	}
	return results
}

// ----------------------------------------------------------------------

func runTestArchive(ctx context.Context, t *testing.T, store Store) {
	archiveStore, ok := store.(ArchiveStore)
	require.True(t, ok)

	status, err := store.ApplyResources(ctx, []model.Resource{
		cabinDestinationType,
	})
	require.NoError(t, err)
	require.Equal(t, model.StatusCreated, status[0].Status)

	t.Run("two versions", func(t *testing.T) {
		// use locals so mutating them doesn't affect other tests
		var (
			v1 = model.NewDestination("cabin-1", "cabin:1", nil)
			v2 = model.NewDestination("cabin-1", "cabin:1", []model.Parameter{
				{
					Name:  "s",
					Value: "1",
				},
			})
		)

		cloneV1, err := model.Clone(v1)
		require.NoError(t, err)
		cloneV2, err := model.Clone(v2)
		require.NoError(t, err)
		status, err := store.ApplyResources(ctx, []model.Resource{cloneV1})
		require.NoError(t, err)
		require.Equal(t, model.StatusCreated, status[0].Status)

		status, err = store.ApplyResources(ctx, []model.Resource{cloneV2})
		require.NoError(t, err)
		require.Equal(t, model.StatusConfigured, status[0].Status)

		v1.SetLatest(false)
		v2.SetLatest(true)
		v2.Metadata.Version = model.Version(2)
		v2.Status.Latest = true

		// check that the new version is the default
		d, err := store.Destination(ctx, v1.Name())
		require.NoError(t, err)
		require.Equal(t, model.Version(2), d.Version())
		require.True(t, d.Status.Latest)

		v1.Status.Latest = false

		// check that the old version is archived
		tests := []struct {
			name           string
			expectResource *model.Destination
			expectVersion  model.Version
		}{
			{
				name:           "cabin-1",
				expectResource: v2,
				expectVersion:  2,
			},
			{
				name:           "cabin-1:0", // 0 is used for current
				expectResource: v2,
				expectVersion:  2,
			},
			{
				name:           "cabin-1:1",
				expectResource: v1,
				expectVersion:  1,
			},
			{
				name:           "cabin-1:2",
				expectResource: v2,
				expectVersion:  2,
			},
			{
				name:           "cabin-1:3", // 3 doesn't exist
				expectResource: nil,
				expectVersion:  3,
			},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				d, err = store.Destination(ctx, test.name)
				require.NoError(t, err)
				if d != nil {
					require.Equal(t, test.expectVersion, d.Version())
					test.expectResource.SetID(d.ID())
				}
				assertResourceVersionsEqual(t, test.expectResource, d)
			})
		}
	})

	t.Run("version history includes all versions", func(t *testing.T) {
		tests := []struct {
			name           string
			expectVersions []model.Version
		}{
			{
				name:           "cabin-1",
				expectVersions: []model.Version{2, 1}, // reverse order
			},
			{
				name:           "cabin-2", // does not exist
				expectVersions: []model.Version{},
			},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				history, err := archiveStore.ResourceHistory(ctx, model.KindDestination, test.name)
				require.NoError(t, err)
				historyVersions := []model.Version{}
				for _, item := range history {
					historyVersions = append(historyVersions, item.Version())
				}
				require.Equal(t, test.expectVersions, historyVersions)
			})
		}
	})

	t.Run("deleting a resource deletes all Versions", func(t *testing.T) {
		tests := []struct {
			name         string
			expectExists bool
		}{
			{
				name:         "cabin-1",
				expectExists: true,
			},
			{
				name:         "cabin-2", // does not exist
				expectExists: false,
			},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				dest, err := store.DeleteDestination(ctx, test.name)
				require.NoError(t, err)
				require.Equal(t, test.expectExists, dest != nil)

				// request should be nil
				dest, err = store.Destination(ctx, test.name)
				require.NoError(t, err)
				require.Nil(t, dest)

				// history should now be empty
				history, err := archiveStore.ResourceHistory(ctx, model.KindDestination, test.name)
				require.NoError(t, err)
				require.Len(t, history, 0)
			})
		}
	})

	t.Run("applying a new resource with a version ignores the version", func(t *testing.T) {
		v3 := model.NewDestination("cabin-1", "cabin", nil)
		v3.SetVersion(model.Version(3))
		statuses, err := store.ApplyResources(ctx, []model.Resource{v3})
		require.NoError(t, err)
		require.Equal(t, model.StatusCreated, statuses[0].Status)
		require.Equal(t, model.Version(1), statuses[0].Resource.Version())
	})

	t.Run("applying a new version of a resource with a version ignores the version", func(t *testing.T) {
		v2 := model.NewDestination("cabin-1", "cabin:1", []model.Parameter{
			{
				Name:  "s",
				Value: "1",
			},
		})
		v2.SetVersion(model.Version(6))

		statuses, err := store.ApplyResources(ctx, []model.Resource{v2})
		require.NoError(t, err)
		require.Equal(t, model.StatusConfigured, statuses[0].Status)
		require.Equal(t, model.Version(2), statuses[0].Resource.Version())
	})
}

func runTestStatus(ctx context.Context, t *testing.T, store Store) {
	// status changes cannot be made via ApplyResources
	t.Run("ignores status changes", func(t *testing.T) {
		applyTestConfiguration(t, store)
		_, _ = store.ApplyResources(ctx, []model.Resource{testConfiguration})
		testConfig, err := store.Configuration(ctx, "configuration-1")
		require.NoError(t, err)
		require.NotNil(t, testConfig)

		// clone config
		testConfigClone, err := model.Clone(testConfig)
		require.NoError(t, err)

		testConfigClone.SetVersion(model.Version(20))
		testConfigClone.SetStatus(model.ConfigurationStatus{
			Rollout: model.Rollout{
				Status: model.RolloutStatusStarted,
				Options: model.RolloutOptions{
					StartAutomatically: true,
					RollbackOnFailure:  true,
					MaxErrors:          10,
					PhaseAgentCount: model.PhaseAgentCount{
						Initial:    2,
						Multiplier: 2,
						Maximum:    30,
					},
				},
			},
			CurrentVersion: 17,
		})

		statuses, err := store.ApplyResources(ctx, []model.Resource{testConfigClone})
		require.NoError(t, err)
		require.Len(t, statuses, 1)
		require.Equal(t, model.StatusUnchanged, statuses[0].Status)

		testConfigAfter, err := store.Configuration(ctx, "configuration-1")
		require.NoError(t, err)
		require.NotNil(t, testConfigAfter)
		require.Equal(t, model.Version(0), testConfigAfter.Status.CurrentVersion)
		require.Equal(t, testConfig.Status, testConfigAfter.Status)
	})
}

func testUpdateRollout(ctx context.Context, t *testing.T, store Store) {
	c1 := &model.Configuration{
		ResourceMeta: model.ResourceMeta{
			APIVersion: modelversion.V1,
			Kind:       model.KindConfiguration,
			Metadata: model.Metadata{
				Name:    "config-1",
				Labels:  model.MakeLabels(),
				Version: 1,
			},
		},
		StatusType: model.StatusType[model.ConfigurationStatus]{
			Status: model.ConfigurationStatus{
				Rollout: model.Rollout{
					Status: model.RolloutStatusStarted,
					Options: model.RolloutOptions{
						StartAutomatically: true,
						RollbackOnFailure:  true,
						MaxErrors:          10,
						PhaseAgentCount: model.PhaseAgentCount{
							Initial:    2,
							Multiplier: 2,
							Maximum:    30,
						},
					},
				},
			}},
	}
	c2 := &model.Configuration{
		ResourceMeta: model.ResourceMeta{
			APIVersion: modelversion.V1,
			Kind:       model.KindConfiguration,
			Metadata: model.Metadata{
				Name:    "config-2",
				Labels:  model.MakeLabels(),
				Version: 1,
			},
		},
		StatusType: model.StatusType[model.ConfigurationStatus]{
			Status: model.ConfigurationStatus{
				Rollout: model.Rollout{
					Status: model.RolloutStatusPaused,
					Options: model.RolloutOptions{
						StartAutomatically: false,
						RollbackOnFailure:  false,
						MaxErrors:          0,
						PhaseAgentCount: model.PhaseAgentCount{
							Initial:    3,
							Multiplier: 6,
							Maximum:    12,
						},
					},
				},
			}},
	}
	c3 := &model.Configuration{
		ResourceMeta: model.ResourceMeta{
			APIVersion: modelversion.V1,
			Kind:       model.KindConfiguration,
			Metadata: model.Metadata{
				Name:    "config-3",
				Labels:  model.MakeLabels(),
				Version: 1,
			},
		},
		StatusType: model.StatusType[model.ConfigurationStatus]{
			Status: model.ConfigurationStatus{
				Rollout: model.Rollout{
					Status: model.RolloutStatusStarted,
					Options: model.RolloutOptions{
						StartAutomatically: true,
						RollbackOnFailure:  true,
						MaxErrors:          10,
						PhaseAgentCount: model.PhaseAgentCount{
							Initial:    2,
							Multiplier: 2,
							Maximum:    30,
						},
					},
				},
			}},
	}
	t.Run("update rollout for active rollout", func(t *testing.T) {
		// create 100 agents
		for i := 0; i < 100; i++ {
			a := &model.Agent{
				ID:   fmt.Sprintf("agent-%d", i),
				Name: fmt.Sprintf("Agent %d", i),
				Labels: model.Labels{
					Set: model.MakeLabels().Set,
				},
				ConfigurationStatus: model.ConfigurationVersions{
					Future: "config-1:1",
				},
				Status: model.Connected,
			}
			err := addAgent(store, a)
			require.NoError(t, err)
		}

		// clone before ApplyResources
		c1Clone, err := model.Clone(c1)
		require.NoError(t, err)
		c2Clone, err := model.Clone(c2)
		require.NoError(t, err)
		c3Clone, err := model.Clone(c3)
		require.NoError(t, err)

		// apply to add the configurations to the store
		updates, err := store.ApplyResources(ctx, []model.Resource{c1, c2, c3})
		require.NoError(t, err)

		// Test that the updates are StatusCreated
		assert.Len(t, updates, 3)
		assert.Equal(t, model.StatusCreated, updates[0].Status)
		assert.Equal(t, model.StatusCreated, updates[1].Status)
		assert.Equal(t, model.StatusCreated, updates[2].Status)

		// update to include status using the clones
		for _, c := range []*model.Configuration{c1Clone, c2Clone, c3Clone} {
			_, status, err := store.UpdateConfiguration(ctx, c.Name(), func(current *model.Configuration) {
				current.Status = c.Status
			})
			require.NoError(t, err)
			require.Equal(t, model.StatusConfigured, status, "configuration %s status", c.Name())
		}

		// These are the number of agents that should be moved to pending in each update cycle
		agentsInPending := []int{2, 4, 8, 16, 30, 30, 10}

		var config *model.Configuration
		for i, numAgents := range agentsInPending {

			config, err = store.UpdateRollout(ctx, "config-1")
			assert.Equal(t, model.RolloutStatusStarted, config.Status.Rollout.Status)
			require.NoError(t, err)
			require.NotNil(t, config)

			agents, err := store.Agents(ctx)
			require.NoError(t, err)

			// Move each pending agent to current, simulating the manager.
			// Count the number of agents in each state, and check that the number of agents in each state is correct.

			pending := 0

			for _, a := range agents {
				agentConfigBefore, err := store.AgentConfiguration(ctx, a)
				require.NoError(t, err)
				require.NotNil(t, agentConfigBefore)
				_, err = store.UpsertAgent(ctx, a.ID, func(current *model.Agent) {
					if current.ConfigurationStatus.Pending != "" {
						pending++
						current.ConfigurationStatus.Current = a.ConfigurationStatus.Pending
						current.ConfigurationStatus.Pending = ""
					}
				})
				require.NoError(t, err)
				agentConfigAfter, err := store.AgentConfiguration(ctx, a)
				require.NoError(t, err)
				require.NotNil(t, agentConfigAfter)
			}
			assert.Equal(t, numAgents, pending, "iteration %d", i)
		}

		config, err = store.UpdateRollout(ctx, "config-1")
		assert.Equal(t, model.RolloutStatusStable, config.Status.Rollout.Status)
		require.NoError(t, err)

		// repeated calls should be the same
		config, err = store.UpdateRollout(ctx, "config-1")
		dateModified := config.Metadata.DateModified.UnixMicro()
		assert.Equal(t, model.RolloutStatusStable, config.Status.Rollout.Status)
		require.NoError(t, err)

		// UpdateRollout with no changes should not modify the configuration.
		require.Equal(t, dateModified, config.Metadata.DateModified.UnixMicro())

		// confirm CurrentVersion is set
		config, err = store.Configuration(ctx, "config-1")
		require.NoError(t, err)
		// reset rollout options because we don't need to verify these
		config.Status.Rollout.Options = model.RolloutOptions{}
		require.Equal(t, model.ConfigurationStatus{
			CurrentVersion: 1,
			Rollout: model.Rollout{
				Status: model.RolloutStatusStable,
				Phase:  7,
				Progress: model.RolloutProgress{
					Completed: 100,
				},
				Options: model.RolloutOptions{},
			},
			Current: true,
			Latest:  true,
		}, config.Status)
	})
	t.Run("update rollout doesn't move more agents to pending if manager doesn't move agents to current", func(t *testing.T) {
		agents, err := store.Agents(ctx)
		require.NoError(t, err)

		agentIDs := make([]string, 0, len(agents))
		for _, a := range agents {
			agentIDs = append(agentIDs, a.ID)
		}

		_, err = store.UpsertAgents(ctx, agentIDs, func(current *model.Agent) {
			current.ConfigurationStatus.Future = "config-3:1"
			current.ConfigurationStatus.Current = ""
			current.ConfigurationStatus.Pending = ""
		})

		config, err := store.UpdateRollout(ctx, "config-3")
		assert.Equal(t, model.RolloutStatusStarted, config.Status.Rollout.Status)
		require.NoError(t, err)

		agents, err = store.Agents(ctx)
		require.NoError(t, err)

		pending := 0
		for _, a := range agents {

			if a.ConfigurationStatus.Pending != "" {
				assert.Equal(t, "config-3:1", a.ConfigurationStatus.Pending)
				pending++
			}
		}
		assert.Equal(t, 2, pending)

		// Next update should not move any agents to pending
		config, err = store.UpdateRollout(ctx, "config-3")
		assert.Equal(t, model.RolloutStatusStarted, config.Status.Rollout.Status)
		require.NoError(t, err)

		agents, err = store.Agents(ctx)
		require.NoError(t, err)

		pending = 0
		for _, a := range agents {
			if a.ConfigurationStatus.Pending != "" {
				assert.Equal(t, "config-3:1", a.ConfigurationStatus.Pending)
				pending++
			}
		}
		assert.Equal(t, 2, pending)
	})
	t.Run("update rollout for paused rollout", func(t *testing.T) {
		agents, err := store.Agents(ctx)
		require.NoError(t, err)

		agentIDs := make([]string, 0, len(agents))
		for _, a := range agents {
			agentIDs = append(agentIDs, a.ID)
		}
		_, err = store.UpsertAgents(ctx, agentIDs, func(current *model.Agent) {
			current.ConfigurationStatus.Future = "config-2:1"
			current.ConfigurationStatus.Current = ""
			current.ConfigurationStatus.Pending = ""
		})
		config, err := store.UpdateRollout(ctx, "config-2")
		require.NoError(t, err)
		assert.Equal(t, model.RolloutStatusPaused, config.Status.Rollout.Status)

		// check no agents are moved to pending
		agents, err = store.Agents(ctx)
		require.NoError(t, err)

		for _, a := range agents {
			assert.Equal(t, "", a.ConfigurationStatus.Pending)
		}
	})
	t.Run("update rollout generates errors for missing config", func(t *testing.T) {
		config, err := store.UpdateRollout(ctx, "missing-config")
		require.Error(t, err)
		require.Nil(t, config)
	})
}

func runTestConfigurationVersions(ctx context.Context, t *testing.T, store Store) {
	// apply test resources
	applyAllTestResources(t, store)
	c1 := &model.Configuration{
		ResourceMeta: model.ResourceMeta{
			APIVersion: modelversion.V1,
			Kind:       model.KindConfiguration,
			Metadata: model.Metadata{
				Name:        "config-1",
				Labels:      model.MakeLabels(),
				Version:     1,
				Description: "none",
				DisplayName: "none",
			},
		},
		Spec: model.ConfigurationSpec{
			Destinations: []model.ResourceConfiguration{},
		},
	}
	status := &model.StatusType[model.ConfigurationStatus]{
		Status: model.ConfigurationStatus{
			Rollout: model.Rollout{
				Status: model.RolloutStatusStarted,
				Options: model.RolloutOptions{
					StartAutomatically: true,
					RollbackOnFailure:  true,
					MaxErrors:          10,
					PhaseAgentCount: model.PhaseAgentCount{
						Initial:    2,
						Multiplier: 2,
						Maximum:    30,
					},
				},
			},
		},
	}

	// These changes should not trigger a new version
	nonSpecConfigChanges := []struct {
		name          string
		configUpdater ConfigurationUpdater
	}{
		{
			name: "Metadata description",
			configUpdater: func(c *model.Configuration) {
				c.Metadata.Description = "test description"

			},
		},
		{
			name: "Display name", // does not exist
			configUpdater: func(c *model.Configuration) {
				c.Metadata.DisplayName = "config-new-name"
			},
		},
		{
			name: "New Display name", // does not exist
			configUpdater: func(c *model.Configuration) {
				c.Metadata.DisplayName = "config-new-new-new-name"
			},
		},
	}

	// These changes should trigger a new version on an active rollout
	specConfigChanges := []struct {
		name          string
		configUpdater ConfigurationUpdater
	}{
		{
			name: "Change destination",
			configUpdater: func(c *model.Configuration) {
				c.Spec.Destinations = []model.ResourceConfiguration{
					{
						Name: cabinDestination1.Name(),
					},
				}
			},
		},
		{
			name: "Change sources",
			configUpdater: func(c *model.Configuration) {
				c.Spec.Sources = []model.ResourceConfiguration{
					{
						Name: nginxSource.Name(),
					},
				}
			},
		},
	}

	// apply config-1
	clone, err := model.Clone(c1)
	require.NoError(t, err)
	updates, err := store.ApplyResources(ctx, []model.Resource{clone})
	require.NoError(t, err)
	assert.Len(t, updates, 1)
	assert.Equal(t, model.StatusCreated, updates[0].Status)

	// rollout should not be started because apply resources erases the rollout status

	config, err := store.Configuration(ctx, "config-1")
	require.NoError(t, err)
	assert.Equal(t, model.RolloutStatusPending, config.Status.Rollout.Status)

	newVersionTest := func(t *testing.T, configUpdater ConfigurationUpdater, expectedStatus model.UpdateStatus, expectedVersions int) {
		configUpdater(clone)
		updates, err = store.ApplyResources(ctx, []model.Resource{clone})
		require.NoError(t, err)
		assert.Len(t, updates, 1)
		assert.Equal(t, expectedStatus, updates[0].Status)

		archiveStore, ok := store.(ArchiveStore)
		require.True(t, ok)

		versions, err := archiveStore.ResourceHistory(ctx, model.KindConfiguration, "config-1")
		require.NoError(t, err)
		assert.Len(t, versions, expectedVersions)
	}
	for _, change := range nonSpecConfigChanges {
		t.Run(change.name, func(t *testing.T) {
			newVersionTest(t, change.configUpdater, model.StatusConfigured, 1)
		})
	}

	for _, change := range specConfigChanges {
		t.Run(change.name, func(t *testing.T) {
			newVersionTest(t, change.configUpdater, model.StatusConfigured, 1)
		})
	}

	// reapply config-1 so the test changes will all be changes
	clone, err = model.Clone(c1)
	require.NoError(t, err)
	updates, err = store.ApplyResources(ctx, []model.Resource{clone})
	require.NoError(t, err)
	assert.Len(t, updates, 1)
	assert.Equal(t, model.StatusConfigured, updates[0].Status)

	// start rollout
	_, err = store.StartRollout(ctx, "config-1", &status.Status.Rollout.Options)
	require.NoError(t, err)

	// expect no new versions to be created.
	for _, change := range nonSpecConfigChanges {
		t.Run(change.name, func(t *testing.T) {
			newVersionTest(t, change.configUpdater, model.StatusConfigured, 1)
		})
	}

	versionCount := 1
	for _, change := range specConfigChanges {
		t.Run(change.name, func(t *testing.T) {
			versionCount++
			newVersionTest(t, change.configUpdater, model.StatusConfigured, versionCount)
			// new version should not have a rollout started
			configWithVersion := fmt.Sprintf("config-1:%d", versionCount)
			config, err = store.Configuration(ctx, configWithVersion)
			require.NoError(t, err)
			require.NotNil(t, config)
			assert.Equal(t, model.RolloutStatusPending, config.Status.Rollout.Status)
			// start rollout
			_, err = store.StartRollout(ctx, "config-1", &status.Status.Rollout.Options)
			require.NoError(t, err)
		})
	}

	// versions 1 & 2 should have status rollout replaced
	config, err = store.Configuration(ctx, "config-1:1")
	require.NoError(t, err)
	assert.Equal(t, model.RolloutStatusStable, config.Status.Rollout.Status)
	assert.Equal(t, false, config.Status.Latest)
	assert.Equal(t, false, config.Status.Current)

	config, err = store.Configuration(ctx, "config-1:2")
	require.NoError(t, err)
	assert.Equal(t, model.RolloutStatusStable, config.Status.Rollout.Status)
	assert.Equal(t, false, config.Status.Latest)
	assert.Equal(t, false, config.Status.Current)

	// version 3 should have a rollout started
	config, err = store.Configuration(ctx, "config-1:3")
	require.NoError(t, err)
	assert.Equal(t, model.RolloutStatusStable, config.Status.Rollout.Status)
	assert.Equal(t, true, config.Status.Latest)

	// also check the config coming from store.Configuration
	config, err = store.Configuration(ctx, "config-1:3")
	require.NoError(t, err)
	assert.Equal(t, model.RolloutStatusStable, config.Status.Rollout.Status)
	assert.Equal(t, true, config.Status.Latest)
	assert.Equal(t, true, config.Status.Current)
}

func TestIsNewConfigurationVersion(t *testing.T) {
	v1 := model.NewConfigurationWithSpec("config", model.ConfigurationSpec{
		Raw: "old",
	})
	v1.EnsureMetadata(v1.Spec)
	v1Any, err := model.AsAny(v1)
	require.NoError(t, err)

	v1StatusChange, err := model.Clone(v1)
	require.NoError(t, err)
	v1StatusChange.Status = model.ConfigurationStatus{
		CurrentVersion: 1,
	}

	v1Started, err := model.Clone(v1)
	require.NoError(t, err)
	v1Started.Status = model.ConfigurationStatus{
		Rollout: model.Rollout{
			Status: model.RolloutStatusStarted,
		},
	}
	v1StartedAny, err := model.AsAny(v1Started)
	fmt.Println(v1StartedAny)

	require.NoError(t, err)

	v1SpecChange, err := model.Clone(v1)
	require.NoError(t, err)
	v1SpecChange.Spec.Raw = "new"
	v1SpecChange.EnsureHash(v1SpecChange.Spec)

	tests := []struct {
		name        string
		curResource *model.AnyResource
		newResource model.Resource
		expectValue bool
		expectError error
	}{
		{
			name:        "no current resource - pending",
			curResource: nil,
			newResource: v1,
			expectValue: false,
			expectError: nil,
		},
		{
			name:        "no current resource - started",
			curResource: nil,
			newResource: v1Started,
			expectValue: false,
			expectError: nil,
		},
		{
			name:        "not a configuration",
			curResource: v1Any,
			newResource: model.NewSourceType("source", []model.ParameterDefinition{}, []string{}),
			expectValue: false,
			expectError: nil,
		},
		{
			name:        "not started - status change",
			curResource: v1Any,
			newResource: v1StatusChange,
			expectValue: false,
			expectError: nil,
		},
		{
			name:        "not started - started",
			curResource: v1Any,
			newResource: v1Started,
			expectValue: false,
			expectError: nil,
		},
		{
			name:        "not started - spec change",
			curResource: v1Any,
			newResource: v1SpecChange,
			expectValue: false,
			expectError: nil,
		},
		{
			name:        "started - status change",
			curResource: v1StartedAny,
			newResource: v1StatusChange,
			expectValue: false,
			expectError: nil,
		},
		{
			name:        "started - noop",
			curResource: v1StartedAny,
			newResource: v1Started,
			expectValue: false,
			expectError: nil,
		},
		{
			name:        "started - spec change",
			curResource: v1StartedAny,
			newResource: v1SpecChange,
			expectValue: true,
			expectError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, _, err := IsNewConfigurationVersion(test.curResource, test.newResource)
			require.Equal(t, test.expectValue, value)
			require.Equal(t, test.expectError, err)
		})
	}
}

func runTestUpdateRollouts(ctx context.Context, t *testing.T, store Store) {

	store.Clear()
	midRolloutConfig := &model.Configuration{
		ResourceMeta: model.ResourceMeta{
			APIVersion: modelversion.V1,
			Kind:       model.KindConfiguration,
			Metadata: model.Metadata{
				Name:    "mid-rollout-config",
				Labels:  model.MakeLabels(),
				Version: 1,
			},
		},
		StatusType: model.StatusType[model.ConfigurationStatus]{
			Status: model.ConfigurationStatus{
				Rollout: model.Rollout{
					Status: model.RolloutStatusStarted, Options: model.RolloutOptions{
						StartAutomatically: true,
						RollbackOnFailure:  true,
						MaxErrors:          10,
						PhaseAgentCount: model.PhaseAgentCount{
							Initial:    2,
							Multiplier: 2,
							Maximum:    30,
						},
					},
					Phase: 2,
					Progress: model.RolloutProgress{
						Completed: 6,
						Errors:    0,
						Pending:   8,
						Waiting:   20,
					},
				},
			},
		},
	}
	pausedConfig := &model.Configuration{
		ResourceMeta: model.ResourceMeta{
			APIVersion: modelversion.V1,
			Kind:       model.KindConfiguration,
			Metadata: model.Metadata{
				Name:    "paused-config",
				Labels:  model.MakeLabels(),
				Version: 1,
			},
		},
		StatusType: model.StatusType[model.ConfigurationStatus]{
			Status: model.ConfigurationStatus{
				Rollout: model.Rollout{
					Status: model.RolloutStatusPaused,
					Options: model.RolloutOptions{
						StartAutomatically: false, // TODO test
						RollbackOnFailure:  false, // TODO test
						MaxErrors:          0,     // TODO test
						PhaseAgentCount: model.PhaseAgentCount{
							Initial:    2,
							Multiplier: 2,
							Maximum:    30,
						},
					},
					Progress: model.RolloutProgress{
						Completed: 0,
						Errors:    0,
						Pending:   0,
						Waiting:   10,
					},
				},
			}},
	}
	pendingConfig := &model.Configuration{
		ResourceMeta: model.ResourceMeta{
			APIVersion: modelversion.V1,
			Kind:       model.KindConfiguration,
			Metadata: model.Metadata{
				Name:    "pending-config",
				Labels:  model.MakeLabels(),
				Version: 1,
			},
		},
		StatusType: model.StatusType[model.ConfigurationStatus]{
			Status: model.ConfigurationStatus{
				Rollout: model.Rollout{
					Status: model.RolloutStatusStarted,
					Options: model.RolloutOptions{
						StartAutomatically: true,
						RollbackOnFailure:  true,
						MaxErrors:          10,
						PhaseAgentCount: model.PhaseAgentCount{
							Initial:    2,
							Multiplier: 2,
							Maximum:    30,
						},
					},
					Progress: model.RolloutProgress{
						Completed: 0,
						Errors:    0,
						Pending:   0,
						Waiting:   30,
					},
				},
			}},
	}

	allConfigs := []struct {
		config         *model.Configuration
		expectedAgents map[int]int // map from phase to expected number of agents pending
	}{
		{
			midRolloutConfig,
			map[int]int{
				2: 8,
				3: 8,
				4: 12,
			},
		},
		{
			pausedConfig,
			map[int]int{0: 0},
		},
		{
			pendingConfig,
			map[int]int{1: 2, 2: 4, 3: 8, 4: 16},
		},
	}
	seq := util.NewTestSequence(t)

	for _, c := range allConfigs {
		seq.Run(fmt.Sprintf("applying config %s", c.config.Name()), func(t *testing.T) {
			applyTestConfigurationAndAgents(ctx, t, store, c.config)
		})
	}

	for {
		var configs []*model.Configuration
		var err error
		seq.Run("updating rollouts", func(t *testing.T) {
			configs, err = store.UpdateRollouts(ctx)
			require.NoError(t, err)
		})

		agentsPendingCount := map[string]int{}
		seq.Run("Move each pending agent to current", func(t *testing.T) {
			agents, err := store.Agents(ctx)
			require.NoError(t, err)

			// Move each pending agent to current, simulating the manager.
			// Count the number of agents in each state, and check that the number of agents in each state is correct.

			for _, a := range agents {
				_, err = store.UpsertAgent(ctx, a.ID, func(current *model.Agent) {
					if current.ConfigurationStatus.Pending != "" {
						agentsPendingCount[current.ConfigurationStatus.Pending]++
						current.ConfigurationStatus.Current = a.ConfigurationStatus.Pending
						current.ConfigurationStatus.Pending = ""
					}
				})
				require.NoError(t, err)
			}
		})

		seq.Run("Check that the number of agents in each state is correct", func(t *testing.T) {
			for _, c := range allConfigs {
				configInStore, err := store.Configuration(ctx, c.config.NameAndVersion())
				require.NoError(t, err)
				require.NotNil(t, configInStore)

				if configInStore.Status.Rollout.Status == model.RolloutStatusStarted {
					phase := configInStore.Rollout().Phase
					numAgents := agentsPendingCount[c.config.NameAndVersion()]

					require.Equal(t, c.expectedAgents[phase], numAgents, "config %s, phase %d", c.config.NameAndVersion(), phase)
				}
			}
		})

		rolloutsFinished := true
		for _, c := range configs {
			if c.Status.Rollout.Status == model.RolloutStatusStarted {
				rolloutsFinished = false
			}

		}
		if rolloutsFinished {
			break
		}
	}

}

// I need a function that takes a configuration and inserts it into the store
// and then inserts all the appropriate agents into the store
func applyTestConfigurationAndAgents(ctx context.Context, t *testing.T, store Store, config *model.Configuration) {

	// insert the configuration
	clone, err := model.Clone(config)
	require.NoError(t, err)
	statuses, err := store.ApplyResources(ctx, []model.Resource{clone})
	require.NoError(t, err)
	require.Len(t, statuses, 1)
	require.Equal(t, model.StatusCreated, statuses[0].Status)

	_, status, err := store.UpdateConfiguration(ctx, config.Name(), func(current *model.Configuration) {
		current.Status = config.Status
	})
	require.NoError(t, err)
	require.Equal(t, model.StatusConfigured, status, "configuration %s status", config.Name())

	// insert the agents
	// create agents in state specified in config
	initAgent := func(ctx context.Context, store Store, id string, status model.ConfigurationVersions) {
		a := &model.Agent{
			ID:                  id,
			ConfigurationStatus: status,
			Status:              model.Connected,
		}
		err := addAgent(store, a)
		require.NoError(t, err)
	}
	// pending agents
	for i := 0; i < config.Status.Rollout.Progress.Pending; i++ {
		initAgent(ctx, store, fmt.Sprintf("%s-pending-%d", config.Name(), i),
			model.ConfigurationVersions{
				Pending: config.NameAndVersion(),
			})
	}

	// completed agents
	for i := 0; i < config.Status.Rollout.Progress.Completed; i++ {
		initAgent(ctx, store, fmt.Sprintf("%s-completed-%d", config.Name(), i),
			model.ConfigurationVersions{
				Current: config.NameAndVersion(),
			})
	}

	// errored agents
	for i := 0; i < config.Status.Rollout.Progress.Errors; i++ {
		initAgent(ctx, store, fmt.Sprintf("%s-errored-%d", config.Name(), i),
			model.ConfigurationVersions{
				Future: config.NameAndVersion(),
			})
	}

	// future agents
	for i := 0; i < config.Status.Rollout.Progress.Waiting; i++ {
		initAgent(ctx, store, fmt.Sprintf("%s-waiting-%d", config.Name(), i),
			model.ConfigurationVersions{
				Future: config.NameAndVersion(),
			})
	}

}

func testResumeErroredRollout(ctx context.Context, t *testing.T, store Store) {
	// setup
	c1 := model.NewConfigurationWithSpec("c1", model.ConfigurationSpec{
		Raw: "service:",
		Selector: model.AgentSelector{
			MatchLabels: model.MatchLabels{
				"configuration": "c1",
			},
		},
	})

	agentIDs := []string{}
	for i := 0; i < 10; i++ {
		agentIDs = append(agentIDs, fmt.Sprintf("c1agent-%d", i))
	}

	simulateAgentConfigurationError := func(configurationName string, expectAgents int) {
		agents, err := store.Agents(ctx, WithQuery(search.ParseQuery("configuration-pending:"+configurationName)))
		require.NoError(t, err)
		require.Len(t, agents, expectAgents)
		for _, agent := range agents {
			_, err = store.UpsertAgent(ctx, agent.ID, func(agent *model.Agent) {
				agent.Status = model.Error
			})
		}
	}

	seq := util.NewTestSequence(t)

	seq.Run("setup: new configuration", func(t *testing.T) {
		config, err := model.Clone(c1)
		require.NoError(t, err)
		status, err := store.ApplyResources(ctx, []model.Resource{config})
		require.NoError(t, err)
		require.Len(t, status, 1)
		require.Equal(t, model.StatusCreated, status[0].Status)
		require.Equal(t, status[0].Resource.(*model.Configuration).Status.Rollout.Status, model.RolloutStatusPending)
	})

	seq.Run("setup: create 10 agents for config", func(t *testing.T) {
		for _, id := range agentIDs {
			agent, err := store.UpsertAgent(ctx, id, func(current *model.Agent) { current.Status = model.Connected })
			require.NoError(t, err)
			require.Equal(t, id, agent.ID)
		}
	})

	seq.Run("assign configuration to agents", func(t *testing.T) {
		_, err := store.UpsertAgents(ctx, agentIDs, func(agent *model.Agent) {
			agent.Labels = model.LabelsFromValidatedMap(map[string]string{
				"configuration": "c1",
			})

			agent.SetFutureConfiguration(c1)
		})
		require.NoError(t, err)
	})

	seq.Run("agents are assigned a future configuration", func(t *testing.T) {
		agents, err := store.Agents(ctx)
		require.NoError(t, err)
		for _, agent := range agents {
			if strings.HasPrefix(agent.ID, "c1") {
				require.Equal(t, "c1", agent.Labels.Set["configuration"])
				require.Equal(t, "", agent.ConfigurationStatus.Current)
				require.Equal(t, "", agent.ConfigurationStatus.Pending)
				require.Equal(t, "c1:1", agent.ConfigurationStatus.Future)
			}
		}
	})

	seq.Run("start the rollout", func(t *testing.T) {
		configuration, err := store.StartRollout(ctx, "c1", &model.RolloutOptions{
			PhaseAgentCount: model.PhaseAgentCount{
				Initial:    3,
				Multiplier: 2,
				Maximum:    5,
			},
			MaxErrors: 1,
		})
		require.NoError(t, err)
		require.NotNil(t, configuration)
		require.Equal(t, model.RolloutStatusStarted, configuration.Status.Rollout.Status)
		require.Equal(t, model.RolloutProgress{
			Pending:   3,
			Completed: 0,
			Errors:    0,
			Waiting:   7,
		}, configuration.Status.Rollout.Progress)
	})

	seq.Run("UpdateRollout, get errored status", func(t *testing.T) {
		simulateAgentConfigurationError("c1:1", 3)
		config, err := store.UpdateRollout(ctx, "c1")
		require.NoError(t, err)

		require.Equal(t, model.RolloutStatusError, config.Status.Rollout.Status)
		require.Equal(t, 3, config.Rollout().Progress.Errors)
	})

	seq.Run("resume rollout, verify new options were set", func(t *testing.T) {
		config, err := store.ResumeRollout(ctx, "c1:1")
		require.NoError(t, err)
		require.NotNil(t, config)

		require.Equal(t, model.RolloutStatusStarted, config.Status.Rollout.Status)
		require.Equal(t, 4, config.Rollout().Options.MaxErrors)
		require.Equal(t, model.RolloutProgress{
			Completed: 0,
			Errors:    3,
			Pending:   5,
			Waiting:   2,
		}, config.Status.Rollout.Progress)
	})

	seq.Run("UpdateRollout, verify new Progress", func(t *testing.T) {
		config, err := store.UpdateRollout(ctx, "c1")
		require.NoError(t, err)

		require.Equal(t, model.RolloutStatusStarted, config.Status.Rollout.Status)
		require.Equal(t, model.RolloutProgress{
			Completed: 0,
			Errors:    3,
			Pending:   5,
			Waiting:   2,
		}, config.Status.Rollout.Progress)
	})

	seq.Run("simulate configuration, get status errored", func(t *testing.T) {
		simulateAgentConfigurationError("c1:1", 8)
		config, err := store.UpdateRollout(ctx, "c1")
		require.NoError(t, err)

		require.Equal(t, model.RolloutStatusError, config.Status.Rollout.Status)
	})

	seq.Run("resume rollout, verify new options were set", func(t *testing.T) {
		config, err := store.ResumeRollout(ctx, "c1:1")
		require.NoError(t, err)
		require.NotNil(t, config)

		require.Equal(t, model.RolloutStatusStarted, config.Status.Rollout.Status)
		// 8 agents have errors (3 from first phase, 5 from second) - MaxErrors should be 9
		require.Equal(t, 9, config.Rollout().Options.MaxErrors)
	})

	seq.Run("UpdateRollout, verify new Progress", func(t *testing.T) {
		config, err := store.UpdateRollout(ctx, "c1")
		require.NoError(t, err)

		require.Equal(t, model.RolloutStatusStarted, config.Status.Rollout.Status)
		require.Equal(t, model.RolloutProgress{
			Completed: 0,
			Errors:    8,
			Pending:   2,
			Waiting:   0,
		}, config.Status.Rollout.Progress)
	})

	seq.Run("simulate configuration, get status errored", func(t *testing.T) {
		simulateAgentConfigurationError("c1:1", 10)
		config, err := store.UpdateRollout(ctx, "c1")
		require.NoError(t, err)

		require.Equal(t, model.RolloutProgress{
			Completed: 0,
			Errors:    10,
			Pending:   0,
			Waiting:   0,
		}, config.Status.Rollout.Progress)
		require.Equal(t, model.RolloutStatusError, config.Status.Rollout.Status)
	})

	seq.Run("resume rollout, verify new options were set", func(t *testing.T) {
		config, err := store.ResumeRollout(ctx, "c1:1")
		require.NoError(t, err)
		require.NotNil(t, config)

		require.Equal(t, model.RolloutStatusStable, config.Status.Rollout.Status)
		require.Equal(t, model.RolloutProgress{
			Completed: 0,
			Errors:    10,
			Pending:   0,
			Waiting:   0,
		}, config.Status.Rollout.Progress)
		require.Equal(t, 11, config.Rollout().Options.MaxErrors)
	})

	seq.Run("UpdateRollout, rollout completed and set as stable", func(t *testing.T) {
		config, err := store.UpdateRollout(ctx, "c1")
		require.NoError(t, err)

		require.Equal(t, model.RolloutStatusStable, config.Status.Rollout.Status)
		require.Equal(t, model.RolloutProgress{
			Completed: 0,
			Errors:    10,
			Pending:   0,
			Waiting:   0,
		}, config.Status.Rollout.Progress)
	})
}

func testStartRollout(ctx context.Context, t *testing.T, store Store) {
	c1 := model.NewConfigurationWithSpec("c1", model.ConfigurationSpec{
		Raw: "service:",
		Selector: model.AgentSelector{
			MatchLabels: model.MatchLabels{
				"configuration": "c1",
			},
		},
	})
	c2 := model.NewConfigurationWithSpec("c2", model.ConfigurationSpec{
		Raw: "service:",
		Selector: model.AgentSelector{
			MatchLabels: model.MatchLabels{
				"configuration": "c2",
			},
		},
	})
	c1agentIDs := []string{}
	for i := 0; i < 10; i++ {
		c1agentIDs = append(c1agentIDs, fmt.Sprintf("c1agent-%d", i))
	}
	c2agentIDs := []string{}
	for i := 0; i < 10; i++ {
		c2agentIDs = append(c2agentIDs, fmt.Sprintf("c2agent-%d", i))
	}
	seq := util.NewTestSequence(t)

	c1c2 := func(configurations []*model.Configuration) (c1 *model.Configuration, c2 *model.Configuration) {
		for _, c := range configurations {
			switch c.Name() {
			case "c1":
				c1 = c
			case "c2":
				c2 = c
			}
		}
		return
	}

	simulateAgentConfiguration := func(configurationName string, expectAgents int) {
		agents, err := store.Agents(ctx, WithQuery(search.ParseQuery("rollout-pending:"+configurationName)))
		require.NoError(t, err)
		require.Len(t, agents, expectAgents)
		configuration, err := store.Configuration(ctx, configurationName)
		require.NoError(t, err)
		for _, agent := range agents {
			_, err = store.UpsertAgent(ctx, agent.ID, func(agent *model.Agent) {
				agent.SetCurrentConfiguration(configuration)
				agent.Status = model.Connected
			})
		}
	}

	// these tests are run in order, so we can use the same config for each test
	seq.Run("setup: new configurations", func(t *testing.T) {
		for _, c := range []*model.Configuration{c1, c2} {
			config, err := model.Clone(c)
			require.NoError(t, err)
			status, err := store.ApplyResources(ctx, []model.Resource{config})
			require.NoError(t, err)
			require.Len(t, status, 1)
			require.Equal(t, model.StatusCreated, status[0].Status)

			config, err = store.Configuration(ctx, c.Name())
			require.NoError(t, err)
			require.Equal(t, config.Status.Rollout.Status, model.RolloutStatusPending)
			require.False(t, config.IsCurrent())
			require.False(t, config.IsPending())
			require.True(t, config.IsLatest())
		}
	})
	seq.Run("setup: create 10 agents for each config", func(t *testing.T) {
		for _, ids := range [][]string{c1agentIDs, c2agentIDs} {
			for _, id := range ids {
				agent, err := store.UpsertAgent(ctx, id, func(current *model.Agent) {
					current.Status = model.Connected
				})
				require.NoError(t, err)
				require.Equal(t, id, agent.ID)
			}
		}
	})
	seq.Run("assign configuration to agents", func(t *testing.T) {
		_, err := store.UpsertAgents(ctx, c1agentIDs, func(agent *model.Agent) {
			agent.Labels = model.LabelsFromValidatedMap(map[string]string{
				"configuration": "c1",
			})
		})
		require.NoError(t, err)
	})
	seq.Run("agents are assigned a future configuration", func(t *testing.T) {
		agents, err := store.Agents(ctx)
		require.NoError(t, err)
		for _, agent := range agents {
			if strings.HasPrefix(agent.ID, "c1") {
				require.Equal(t, "c1", agent.Labels.Set["configuration"])
				require.Equal(t, "", agent.ConfigurationStatus.Current)
				require.Equal(t, "", agent.ConfigurationStatus.Pending)
				require.Equal(t, "c1:1", agent.ConfigurationStatus.Future)
			}
		}
	})
	seq.Run("update rollout does nothing because the rollout is pending", func(t *testing.T) {
		configuration, err := store.UpdateRollout(ctx, "c1")
		require.NoError(t, err)
		require.NotNil(t, configuration)
		require.Equal(t, model.RolloutStatusPending, configuration.Status.Rollout.Status)
		require.False(t, configuration.IsCurrent())
		require.False(t, configuration.IsPending())
		require.True(t, configuration.IsLatest())
	})
	seq.Run("start the rollout", func(t *testing.T) {
		configuration, err := store.StartRollout(ctx, "c1", &model.RolloutOptions{
			PhaseAgentCount: model.PhaseAgentCount{
				Initial:    3,
				Multiplier: 2,
				Maximum:    5,
			},
			MaxErrors: 1,
		})
		require.NoError(t, err)
		require.NotNil(t, configuration)
		require.Equal(t, model.RolloutStatusStarted, configuration.Status.Rollout.Status)
		require.Equal(t, model.RolloutProgress{
			Pending:   3,
			Completed: 0,
			Errors:    0,
			Waiting:   7,
		}, configuration.Status.Rollout.Progress)
		require.False(t, configuration.IsCurrent())
		require.True(t, configuration.IsPending())
		require.True(t, configuration.IsLatest())
	})
	seq.Run("get the rollout status", func(t *testing.T) {
		configuration, err := store.Configuration(ctx, "c1")
		require.NoError(t, err)
		require.NotNil(t, configuration)
		require.Equal(t, model.RolloutStatusStarted, configuration.Status.Rollout.Status)
		require.False(t, configuration.IsCurrent())
		require.True(t, configuration.IsPending())
		require.True(t, configuration.IsLatest())
	})
	seq.Run("create c1:2 by modifying the configuration", func(t *testing.T) {
		config, err := model.Clone(c1)
		require.NoError(t, err)
		config.Spec.Raw = "receivers:\nservice:\n"
		status, err := store.ApplyResources(ctx, []model.Resource{config})
		require.NoError(t, err)
		require.Len(t, status, 1)
		require.Equal(t, model.StatusConfigured, status[0].Status)
	})
	seq.Run("the new configuration is version 2 and has a pending rollout", func(t *testing.T) {
		configuration, err := store.Configuration(ctx, "c1")
		require.NoError(t, err)
		require.NotNil(t, configuration)
		require.Equal(t, model.Version(2), configuration.Version())
		require.Equal(t, model.RolloutStatusPending, configuration.Status.Rollout.Status)
		require.False(t, configuration.IsCurrent())
		require.False(t, configuration.IsPending())
		require.True(t, configuration.IsLatest())
	})
	seq.Run("verify the history with 2 versions", func(t *testing.T) {
		s, ok := store.(ArchiveStore)
		if !ok {
			t.Skip("store does not implement ArchiveStore")
			return
		}
		history, err := s.ResourceHistory(ctx, model.KindConfiguration, "c1")
		require.NoError(t, err)
		require.Len(t, history, 2)

		expect := []struct {
			Version model.Version
			Status  model.RolloutStatus
			Current bool
			Pending bool
			Latest  bool
		}{
			{Version: 2, Status: model.RolloutStatusPending, Latest: true},
			{Version: 1, Status: model.RolloutStatusStarted, Pending: true},
		}

		for i, item := range history {
			require.Equal(t, expect[i].Version, item.Version())
			require.Equal(t, expect[i].Current, item.IsCurrent())
			require.Equal(t, expect[i].Pending, item.IsPending())
			require.Equal(t, expect[i].Latest, item.IsLatest())
		}

		// convert to configuration to verify the status
		configurations, err := model.Parse[*model.Configuration](history)
		require.NoError(t, err)

		for i, item := range configurations {
			require.Equal(t, expect[i].Version, item.Version())
			require.Equal(t, expect[i].Status, item.Status.Rollout.Status)
			require.Equal(t, expect[i].Current, item.IsCurrent())
			require.Equal(t, expect[i].Pending, item.IsPending())
			require.Equal(t, expect[i].Latest, item.IsLatest())
		}
	})
	seq.Run("update the rollout again but no progress because the agents haven't been configured", func(t *testing.T) {
		c1, err := store.Configuration(ctx, "c1:1")
		require.NoError(t, err)
		require.NotNil(t, c1)
		require.Equal(t, model.RolloutStatusStarted, c1.Status.Rollout.Status)
		require.False(t, c1.IsCurrent())
		require.True(t, c1.IsPending())
		require.False(t, c1.IsLatest())

		configuration, err := store.UpdateRollout(ctx, "c1:1")
		require.NoError(t, err)
		require.NotNil(t, configuration)
		require.Equal(t, model.RolloutStatusStarted, configuration.Status.Rollout.Status)
		require.Equal(t, model.RolloutProgress{
			Pending:   3,
			Completed: 0,
			Errors:    0,
			Waiting:   7,
		}, configuration.Status.Rollout.Progress)
		require.Equal(t, model.Version(1), configuration.Version())
		require.False(t, configuration.IsCurrent())
		require.True(t, configuration.IsPending())
		require.False(t, configuration.IsLatest())
	})
	seq.Run("find the agents pending and move them to current, simulating successful configuration", func(t *testing.T) {
		simulateAgentConfiguration("c1:1", 3)
	})
	seq.Run("update the rollout again and make progress", func(t *testing.T) {
		configuration, err := store.UpdateRollout(ctx, "c1:1")
		require.NoError(t, err)
		require.NotNil(t, configuration)
		require.Equal(t, model.RolloutStatusStarted, configuration.Status.Rollout.Status)
		require.Equal(t, model.RolloutProgress{
			Pending:   5,
			Completed: 3,
			Errors:    0,
			Waiting:   2,
		}, configuration.Status.Rollout.Progress)
		require.Equal(t, 2, configuration.Status.Rollout.Phase)
		require.Equal(t, model.Version(1), configuration.Version())
		require.False(t, configuration.IsCurrent())
		require.True(t, configuration.IsPending())
		require.False(t, configuration.IsLatest())
	})
	seq.Run("start a rollout of c1:2", func(t *testing.T) {
		configuration, err := store.StartRollout(ctx, "c1:2", &model.RolloutOptions{
			PhaseAgentCount: model.PhaseAgentCount{
				Initial:    3,
				Multiplier: 2,
				Maximum:    5,
			},
			MaxErrors: 1,
		})
		require.NoError(t, err)
		require.NotNil(t, configuration)
		require.Equal(t, model.RolloutStatusStarted, configuration.Status.Rollout.Status)
		require.Equal(t, model.RolloutProgress{
			Pending:   3,
			Completed: 0,
			Errors:    0,
			Waiting:   7,
		}, configuration.Status.Rollout.Progress)
		require.Equal(t, model.Version(2), configuration.Version())
		require.False(t, configuration.IsCurrent())
		require.True(t, configuration.IsPending())
		require.True(t, configuration.IsLatest())
	})
	seq.Run("rollout c1:1 is now replaced", func(t *testing.T) {
		config, err := store.Configuration(ctx, "c1:1")
		require.NoError(t, err)
		require.Equal(t, model.RolloutStatusReplaced, config.Status.Rollout.Status)
		require.Equal(t, model.Version(1), config.Version())
		require.False(t, config.IsCurrent())
		require.False(t, config.IsPending())
		require.False(t, config.IsLatest())
	})
	seq.Run("verify that all agents are future or pending c1:2", func(t *testing.T) {
		agents, err := store.Agents(ctx, WithQuery(search.ParseQuery("configuration:c1")))
		require.NoError(t, err)
		require.Len(t, agents, 10)
		pending := 0
		for _, agent := range agents {
			if agent.ConfigurationStatus.Future == "" {
				require.Equal(t, "c1:2", agent.ConfigurationStatus.Pending)
				pending++
			} else {
				require.Equal(t, "c1:2", agent.ConfigurationStatus.Future)
			}
		}
		require.Equal(t, 3, pending)
	})
	seq.Run("simulate an error on an agent", func(t *testing.T) {
		config, err := store.Configuration(ctx, "c1:2")
		require.NoError(t, err)

		agents, err := store.Agents(ctx, WithQuery(search.ParseQuery("configuration-pending:c1:2")))
		require.NoError(t, err)
		require.Len(t, agents, 3)

		for i, agent := range agents {
			_, err := store.UpsertAgent(ctx, agent.ID, func(agent *model.Agent) {
				// make two of them errors
				if i == 0 || i == 1 {
					agent.Status = model.Error
				}
			})
			require.NoError(t, err)
		}
		require.Equal(t, model.Version(2), config.Version())
		require.False(t, config.IsCurrent())
		require.True(t, config.IsPending())
		require.True(t, config.IsLatest())
	})
	seq.Run("assign c2 to the c2 agents and start a c2 rollout", func(t *testing.T) {
		_, err := store.UpsertAgents(ctx, c2agentIDs, func(agent *model.Agent) {
			agent.Labels = model.LabelsFromValidatedMap(map[string]string{
				"configuration": "c2",
			})
		})
		require.NoError(t, err)
		configuration, err := store.StartRollout(ctx, "c2", &model.RolloutOptions{
			PhaseAgentCount: model.PhaseAgentCount{
				Initial:    1,
				Multiplier: 2,
				Maximum:    5,
			},
			MaxErrors: 1,
		})
		require.NoError(t, err)
		require.NotNil(t, configuration)
		require.Equal(t, model.RolloutStatusStarted, configuration.Status.Rollout.Status)
		require.Equal(t, model.RolloutProgress{
			Pending:   1,
			Completed: 0,
			Errors:    0,
			Waiting:   9,
		}, configuration.Status.Rollout.Progress)
		require.False(t, configuration.IsCurrent())
		require.True(t, configuration.IsPending())
		require.True(t, configuration.IsLatest())
	})
	seq.Run("simulate configuration of c1 and c2", func(t *testing.T) {
		simulateAgentConfiguration("c2:1", 1)
		simulateAgentConfiguration("c1:2", 1) // 2 errored agents don't count as pending
	})
	seq.Run("update rollouts expect an error for c1:2", func(t *testing.T) {
		configs, err := store.UpdateRollouts(ctx)
		require.NoError(t, err)

		// find c1 and c2 in the results
		c1, c2 := c1c2(configs)

		// c1
		require.NotNil(t, c1)
		require.Equal(t, model.RolloutProgress{
			Pending:   0,
			Completed: 1,
			Errors:    2,
			Waiting:   7,
		}, c1.Status.Rollout.Progress)
		require.Equal(t, model.RolloutStatusError, c1.Status.Rollout.Status)
		require.Equal(t, model.Version(2), c1.Version())
		require.False(t, c1.IsCurrent())
		require.True(t, c1.IsPending())
		require.True(t, c1.IsLatest())

		// c2
		require.NotNil(t, c2)
		require.Equal(t, model.RolloutProgress{
			Pending:   2,
			Completed: 1,
			Errors:    0,
			Waiting:   7,
		}, c2.Status.Rollout.Progress)
		require.Equal(t, model.RolloutStatusStarted, c2.Status.Rollout.Status)
		require.Equal(t, model.Version(1), c2.Version())
		require.False(t, c2.IsCurrent())
		require.True(t, c2.IsPending())
		require.True(t, c2.IsLatest())
	})
	seq.Run("create c1:3 by modifying the configuration", func(t *testing.T) {
		config, err := model.Clone(c1)
		require.NoError(t, err)
		config.Spec.Raw = "receivers:\nprocessors:\nservice:\n"
		status, err := store.ApplyResources(ctx, []model.Resource{config})
		require.NoError(t, err)
		require.Len(t, status, 1)
		require.Equal(t, model.StatusConfigured, status[0].Status)
	})
	seq.Run("the new configuration is version 3 and has a pending rollout", func(t *testing.T) {
		config, err := store.Configuration(ctx, "c1")
		require.NoError(t, err)
		require.NotNil(t, config)
		require.Equal(t, model.RolloutStatusPending, config.Status.Rollout.Status)
		require.Equal(t, model.Version(3), config.Version())
		require.False(t, config.IsCurrent())
		require.False(t, config.IsPending())
		require.True(t, config.IsLatest())
	})
	seq.Run("verify the history with 3 versions", func(t *testing.T) {
		s, ok := store.(ArchiveStore)
		if !ok {
			t.Skip("store does not implement ArchiveStore")
			return
		}
		history, err := s.ResourceHistory(ctx, model.KindConfiguration, "c1")
		require.NoError(t, err)
		require.Len(t, history, 3)

		expect := []struct {
			Version model.Version
			Status  model.RolloutStatus
			Current bool
			Pending bool
			Latest  bool
		}{
			{Version: 3, Status: model.RolloutStatusPending, Latest: true},
			{Version: 2, Status: model.RolloutStatusError, Pending: true},
			{Version: 1, Status: model.RolloutStatusReplaced},
		}

		for i, item := range history {
			require.Equal(t, expect[i].Version, item.Version())
			require.Equal(t, expect[i].Current, item.IsCurrent())
			require.Equal(t, expect[i].Pending, item.IsPending())
			require.Equal(t, expect[i].Latest, item.IsLatest())
		}

		// convert to configuration to verify the status
		configurations, err := model.Parse[*model.Configuration](history)
		require.NoError(t, err)

		for i, item := range configurations {
			require.Equal(t, expect[i].Version, item.Version())
			require.Equal(t, expect[i].Status, item.Status.Rollout.Status)
			require.Equal(t, expect[i].Current, item.IsCurrent())
			require.Equal(t, expect[i].Pending, item.IsPending())
			require.Equal(t, expect[i].Latest, item.IsLatest())
		}
	})
	seq.Run("start a rollout of c1:3", func(t *testing.T) {
		configuration, err := store.StartRollout(ctx, "c1:3", &model.RolloutOptions{
			PhaseAgentCount: model.PhaseAgentCount{
				Initial:    3,
				Multiplier: 2,
				Maximum:    5,
			},
			MaxErrors: 0,
		})
		require.NoError(t, err)
		require.NotNil(t, configuration)
		require.Equal(t, model.RolloutStatusStarted, configuration.Status.Rollout.Status)
		require.Equal(t, model.RolloutProgress{
			Pending:   3,
			Completed: 0,
			Errors:    0,
			Waiting:   7,
		}, configuration.Status.Rollout.Progress)
		require.Equal(t, model.Version(3), configuration.Version())
		require.False(t, configuration.IsCurrent())
		require.True(t, configuration.IsPending())
		require.True(t, configuration.IsLatest())
	})
	seq.Run("clear the error state on all agents", func(t *testing.T) {
		_, err := store.UpsertAgents(ctx, c1agentIDs, func(agent *model.Agent) {
			agent.Status = model.Connected
		})
		require.NoError(t, err)
	})
	seq.Run("simulate completion of phase 1 of c1:3", func(t *testing.T) {
		simulateAgentConfiguration("c1:3", 3)
	})
	seq.Run("simulate completion of phase 2 of c2:1", func(t *testing.T) {
		simulateAgentConfiguration("c2:1", 2)
	})
	seq.Run("phase 2 of the rollout of c1:3", func(t *testing.T) {
		configuration, err := store.UpdateRollout(ctx, "c1:3")
		require.NoError(t, err)
		require.NotNil(t, configuration)
		require.Equal(t, model.RolloutProgress{
			Pending:   5,
			Completed: 3,
			Errors:    0,
			Waiting:   2,
		}, configuration.Status.Rollout.Progress)
		require.Equal(t, model.RolloutStatusStarted, configuration.Status.Rollout.Status)
		require.Equal(t, 2, configuration.Status.Rollout.Phase)
		require.Equal(t, model.Version(3), configuration.Version())
		require.False(t, configuration.IsCurrent())
		require.True(t, configuration.IsPending())
		require.True(t, configuration.IsLatest())

		// simulate the completion of phase 2
		simulateAgentConfiguration("c1:3", 5)
	})
	seq.Run("confirm that all agents are using the config per store.AgentIDsMatchingConfiguration", func(t *testing.T) {
		config, err := store.Configuration(ctx, "c1")
		require.NoError(t, err)
		agentIDs, err := store.AgentsIDsMatchingConfiguration(ctx, config)
		require.NoError(t, err)
		require.ElementsMatch(t, c1agentIDs, agentIDs)
	})
	seq.Run("pause the rollout of c1:3", func(t *testing.T) {
		configuration, err := store.PauseRollout(ctx, "c1")
		require.NoError(t, err)
		require.NotNil(t, configuration)

		// it may seem like these numbers are old since the agents were just updated, but Pause does not do an Update.
		require.Equal(t, model.RolloutProgress{
			Pending:   0,
			Completed: 8,
			Errors:    0,
			Waiting:   2,
		}, configuration.Status.Rollout.Progress)
		require.Equal(t, model.RolloutStatusPaused, configuration.Status.Rollout.Status)
		require.Equal(t, 2, configuration.Status.Rollout.Phase)
		require.Equal(t, model.Version(3), configuration.Version())
		require.False(t, configuration.IsCurrent())
		require.True(t, configuration.IsPending())
		require.True(t, configuration.IsLatest())
	})
	seq.Run("update does nothing while paused", func(t *testing.T) {
		configurations, err := store.UpdateRollouts(ctx)
		c1, c2 := c1c2(configurations)

		// c1
		require.NoError(t, err)
		require.NotNil(t, c1)
		require.Equal(t, model.RolloutProgress{
			Pending:   0,
			Completed: 8,
			Errors:    0,
			Waiting:   2,
		}, c1.Status.Rollout.Progress)
		require.Equal(t, model.RolloutStatusPaused, c1.Status.Rollout.Status)
		require.Equal(t, 2, c1.Status.Rollout.Phase)
		require.Equal(t, model.Version(3), c1.Version())
		require.False(t, c1.IsCurrent())
		require.True(t, c1.IsPending())
		require.True(t, c1.IsLatest())

		// c2
		require.NotNil(t, c2)
		require.Equal(t, model.RolloutProgress{
			Pending:   4,
			Completed: 3,
			Errors:    0,
			Waiting:   3,
		}, c2.Status.Rollout.Progress)
		require.Equal(t, model.RolloutStatusStarted, c2.Status.Rollout.Status)
		require.Equal(t, model.Version(1), c2.Version())
		require.False(t, c2.IsCurrent())
		require.True(t, c2.IsPending())
		require.True(t, c2.IsLatest())
	})
	seq.Run("resume the rollout of c1:3", func(t *testing.T) {
		configuration, err := store.ResumeRollout(ctx, "c1")
		require.NoError(t, err)
		require.NotNil(t, configuration)
		require.Equal(t, model.RolloutProgress{
			Pending:   2,
			Completed: 8,
			Errors:    0,
			Waiting:   0,
		}, configuration.Status.Rollout.Progress)
		require.Equal(t, model.RolloutStatusStarted, configuration.Status.Rollout.Status)
		require.Equal(t, 3, configuration.Status.Rollout.Phase)
		require.Equal(t, model.Version(3), configuration.Version())
		require.False(t, configuration.IsCurrent())
		require.True(t, configuration.IsPending())
		require.True(t, configuration.IsLatest())
	})
	seq.Run("phase 3 of the rollout of c1:3", func(t *testing.T) {
		configuration, err := store.UpdateRollout(ctx, "c1:3")
		require.NoError(t, err)
		require.NotNil(t, configuration)
		require.Equal(t, model.RolloutProgress{
			Pending:   2,
			Completed: 8,
			Errors:    0,
			Waiting:   0,
		}, configuration.Status.Rollout.Progress)
		require.Equal(t, model.RolloutStatusStarted, configuration.Status.Rollout.Status)
		require.Equal(t, 3, configuration.Status.Rollout.Phase)
		require.Equal(t, model.Version(3), configuration.Version())
		require.False(t, configuration.IsCurrent())
		require.True(t, configuration.IsPending())
		require.True(t, configuration.IsLatest())
	})
	seq.Run("pending agents disconnects, but rollout still completes.", func(t *testing.T) {
		agents, err := FindAgents(ctx, store.AgentIndex(ctx), model.FieldConfigurationPending, "c1:3")
		require.NoError(t, err)
		require.Len(t, agents, 2)
		store.UpsertAgents(ctx, agents, func(agent *model.Agent) {
			agent.Status = model.Disconnected
		})
		// simulate the completion of phase 3
		simulateAgentConfiguration("c1:3", 0)
	})
	seq.Run("completed rollout of c1:3", func(t *testing.T) {
		configuration, err := store.UpdateRollout(ctx, "c1:3")
		require.NoError(t, err)
		require.NotNil(t, configuration)
		require.Equal(t, model.RolloutProgress{
			Pending:   0,
			Completed: 8,
			Errors:    0,
			Waiting:   0,
		}, configuration.Status.Rollout.Progress)
		require.Equal(t, model.RolloutStatusStable, configuration.Status.Rollout.Status)
		require.Equal(t, 3, configuration.Status.Rollout.Phase)
		require.Equal(t, model.Version(3), configuration.Version())
		require.True(t, configuration.IsCurrent())
		require.True(t, configuration.IsPending())
		require.True(t, configuration.IsLatest())
	})
}

func runTestDependencyUpdates(ctx context.Context, t *testing.T, store Store, clearStore func(t *testing.T)) {

	// create a processor type to use in the configuration
	severityProcessorType := testResource[*model.ProcessorType](t, "filter_severity.yaml")

	// create another configuration with more complex dependencies
	c2 := model.NewConfigurationWithSpec("c2", model.ConfigurationSpec{
		Sources: []model.ResourceConfiguration{
			{
				ParameterizedSpec: model.ParameterizedSpec{
					Type:       "macos",
					Parameters: []model.Parameter{},
					Processors: []model.ResourceConfiguration{
						{
							ParameterizedSpec: model.ParameterizedSpec{
								Type: severityProcessorType.Name(),
								Parameters: []model.Parameter{
									{
										Name:  "severity",
										Value: "INFO",
									},
								},
							},
						},
					},
				},
			},
			{
				Name: nginxSource.Name(),
			},
		},
		Destinations: []model.ResourceConfiguration{
			{
				ParameterizedSpec: model.ParameterizedSpec{
					Type: cabinDestinationType.Name(),
				},
			},
		},
	})

	type statusResult struct {
		kind    model.Kind
		name    string
		version model.Version
		status  model.UpdateStatus
	}

	tests := []struct {
		name                  string
		applyResources        func(t *testing.T) []model.Resource
		expectedStatusResults []statusResult
		expectedError         error
		verifyResults         func(t *testing.T, statuses []model.ResourceStatus, err error)
	}{
		{
			name: "no changes",
			applyResources: func(t *testing.T) []model.Resource {
				c, err := model.Clone(macosSourceType)
				require.NoError(t, err)
				return []model.Resource{c}
			},
			expectedStatusResults: []statusResult{
				{
					kind:    model.KindSourceType,
					name:    macosSourceType.Name(),
					version: 1,
					status:  model.StatusUnchanged,
				},
			},
		},
		{
			name: "no dependency changes",
			applyResources: func(t *testing.T) []model.Resource {
				c, err := model.Clone(cabinDestination2)
				require.NoError(t, err)
				c.Spec.Parameters = []model.Parameter{
					{
						Name:  "port",
						Value: 5,
					},
				}
				return []model.Resource{c}
			},
			expectedStatusResults: []statusResult{
				{
					kind:    model.KindDestination,
					name:    cabinDestination2.Name(),
					version: 2,
					status:  model.StatusConfigured,
				},
			},
		},
		{
			name: "cabin DestinationType change should update all cabin destinations and configurations that use them",
			applyResources: func(t *testing.T) []model.Resource {
				c, err := model.Clone(cabinDestinationType)
				require.NoError(t, err)
				c.Spec.Parameters = []model.ParameterDefinition{
					{
						Name: "port",
						Type: "string",
					},
				}
				return []model.Resource{c}
			},
			expectedStatusResults: []statusResult{
				{
					kind:    model.KindDestinationType,
					name:    cabinDestinationType.Name(),
					version: 2,
					status:  model.StatusConfigured,
				},
				{
					kind:    model.KindDestination,
					name:    cabinDestination1.Name(),
					version: 2,
					status:  model.StatusConfigured,
				},
				{
					kind:    model.KindDestination,
					name:    cabinDestination2.Name(),
					version: 2,
					status:  model.StatusConfigured,
				},
				{
					kind:    model.KindConfiguration,
					name:    c2.Name(),
					version: 2,
					status:  model.StatusConfigured,
				},
				{
					kind:    model.KindConfiguration,
					name:    testConfiguration.Name(),
					version: 2,
					status:  model.StatusConfigured,
				},
			},
		},
		{
			name: "cabin-1 Destination change should update all configurations that use it",
			applyResources: func(t *testing.T) []model.Resource {
				c, err := model.Clone(cabinDestination1)
				require.NoError(t, err)
				c.Spec.Parameters = []model.Parameter{
					{
						Name:  "port",
						Value: 5,
					},
				}
				return []model.Resource{c}
			},
			expectedStatusResults: []statusResult{
				{
					kind:    model.KindDestination,
					name:    cabinDestination1.Name(),
					version: 2,
					status:  model.StatusConfigured,
				},
				{
					kind:    model.KindConfiguration,
					name:    testConfiguration.Name(),
					version: 2,
					status:  model.StatusConfigured,
				},
			},
		},
		{
			name: "multiple dependency updates should still only trigger a single configuration update for each configuration",
			applyResources: func(t *testing.T) []model.Resource {
				ct, err := model.Clone(cabinDestinationType)
				require.NoError(t, err)
				ct.Spec.Parameters = []model.ParameterDefinition{
					{
						Name: "port",
						Type: "string",
					},
				}

				cd1, err := model.Clone(cabinDestination1)
				require.NoError(t, err)
				cd1.Spec.Parameters = []model.Parameter{
					{
						Name:  "port",
						Value: "5",
					},
				}

				cd2, err := model.Clone(cabinDestination2)
				require.NoError(t, err)
				cd2.Spec.Parameters = []model.Parameter{
					{
						Name:  "port",
						Value: "6",
					},
				}

				c1, err := model.Clone(testConfiguration)
				require.NoError(t, err)
				c1.Spec.Destinations = []model.ResourceConfiguration{
					{
						Name: cabinDestination2.Name(),
					},
				}

				return []model.Resource{ct, cd1, cd2, c1}
			},
			expectedStatusResults: []statusResult{
				{
					kind:    model.KindDestinationType,
					name:    cabinDestinationType.Name(),
					version: 2,
					status:  model.StatusConfigured,
				},
				{
					kind:    model.KindDestination,
					name:    cabinDestination1.Name(),
					version: 2,
					status:  model.StatusConfigured,
				},
				{
					kind:    model.KindDestination,
					name:    cabinDestination2.Name(),
					version: 2,
					status:  model.StatusConfigured,
				},
				{
					kind:    model.KindConfiguration,
					name:    c2.Name(),
					version: 2,
					status:  model.StatusConfigured,
				},
				{
					kind:    model.KindConfiguration,
					name:    testConfiguration.Name(),
					version: 2,
					status:  model.StatusConfigured,
				},
			},
		},
		{
			name: "configuration dependency versions ignored",
			applyResources: func(t *testing.T) []model.Resource {
				c3, err := model.Clone(c2)
				require.NoError(t, err)

				c3.Metadata.Name = "c3"
				c3.Spec.Sources[0].Type = "macos:7"
				c3.Spec.Sources[0].Processors[0].Type = "filter_severity:4"
				c3.Spec.Sources[1].Name = "nginx:5"
				c3.Spec.Destinations[0].Type = "cabin:12"

				return []model.Resource{c3}
			},
			expectedStatusResults: []statusResult{
				{
					kind:    model.KindConfiguration,
					name:    "c3",
					version: 1,
					status:  model.StatusCreated,
				},
			},
			verifyResults: func(t *testing.T, statuses []model.ResourceStatus, err error) {
				// make sure the dependencies are updated
				c3 := statuses[0].Resource.(*model.Configuration)
				require.Equal(t, "macos:1", c3.Spec.Sources[0].Type)
				require.Equal(t, "filter_severity:1", c3.Spec.Sources[0].Processors[0].Type)
				require.Equal(t, "nginx:1", c3.Spec.Sources[1].Name)
				require.Equal(t, "cabin:1", c3.Spec.Destinations[0].Type)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clearStore(t)
			applyAllTestResources(t, store)
			_, err := store.ApplyResources(ctx, []model.Resource{severityProcessorType, c2})
			require.NoError(t, err)

			// rollout configurations to force another version if they change
			_, err = store.StartRollout(ctx, "c2", nil)
			require.NoError(t, err)
			_, err = store.StartRollout(ctx, testConfiguration.Name(), nil)
			require.NoError(t, err)

			statuses, err := store.ApplyResources(ctx, test.applyResources(t))
			require.Equal(t, test.expectedError, err)

			actualStatusResults := []statusResult{}
			for _, status := range statuses {
				actualStatusResults = append(actualStatusResults, statusResult{
					kind:    status.Resource.GetKind(),
					name:    status.Resource.Name(),
					version: status.Resource.Version(),
					status:  status.Status,
				})
			}

			// order doesn't really matter
			require.ElementsMatch(t, test.expectedStatusResults, actualStatusResults, "A=expected, B=actual")

			if test.verifyResults != nil {
				test.verifyResults(t, statuses, err)
			}
		})
	}
}

func runTestCurrentRolloutsForConfiguration(ctx context.Context, t *testing.T, store Store) {

	tests := []struct {
		name              string
		configurationName string
		setupAgent        func(i int, agent *model.Agent)
		expect            []string
	}{
		{
			name:              "no agents",
			configurationName: "c1",
			expect:            nil,
		},
		{
			name: "pending and future",
			setupAgent: func(i int, agent *model.Agent) {
				switch i % 4 {
				case 0:
					agent.ConfigurationStatus.Future = "c1:2"
				case 1:
					agent.ConfigurationStatus.Pending = "c1:1"
				case 2:
					agent.ConfigurationStatus.Future = "c2:2"
				case 3:
					agent.ConfigurationStatus.Pending = "c2:1"
				}
			},
			configurationName: "c1",
			expect:            []string{"c1:1", "c1:2"},
		},
		{
			name: "future only",
			setupAgent: func(i int, agent *model.Agent) {
				switch i % 4 {
				case 0:
					agent.ConfigurationStatus.Future = "c1:2"
				}
			},
			configurationName: "c1",
			expect:            []string{"c1:2"},
		},
		{
			name: "pending only",
			setupAgent: func(i int, agent *model.Agent) {
				switch i % 4 {
				case 0:
					agent.ConfigurationStatus.Pending = "c1:2"
				}
			},
			configurationName: "c1",
			expect:            []string{"c1:2"},
		},
		{
			name: "no matches only",
			setupAgent: func(i int, agent *model.Agent) {
				switch i % 4 {
				case 0:
					agent.ConfigurationStatus.Future = "c1:2"
				}
			},
			configurationName: "c2",
			expect:            []string{},
		},
	}

	// seed 10 agents
	for i := 0; i < 10; i++ {
		store.UpsertAgent(ctx, fmt.Sprintf("agent-%d", i), func(a *model.Agent) {
			a.Status = model.Connected
		})
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			agents, err := store.Agents(ctx)
			require.NoError(t, err)
			for i, agent := range agents {
				_, err := store.UpsertAgent(ctx, agent.ID, func(a *model.Agent) {
					a.ConfigurationStatus = model.ConfigurationVersions{}
					if test.setupAgent != nil {
						test.setupAgent(i, a)
					}
				})
				require.NoError(t, err)
			}

			rollouts, err := CurrentRolloutsForConfiguration(store.AgentIndex(ctx), test.configurationName)
			require.NoError(t, err)
			require.ElementsMatch(t, test.expect, rollouts)
		})
	}
}

func runTestCountAgents(ctx context.Context, t *testing.T, store Store) {
	store.Clear()
	for i := 0; i < 10; i++ {
		_, err := store.UpsertAgent(ctx, fmt.Sprintf("agent-%d", i), func(a *model.Agent) {
			a.Status = model.Connected
			if i%3 == 0 {
				t.Logf(`agent-%d: "c:1", "c:2", "c:3"`, i)
				a.ConfigurationStatus.Set("c:1", "c:2", "c:3")
			} else {
				a.ConfigurationStatus.Set("c:1", "c:3", "")
				t.Logf(`agent-%d: "c:1", "c:3", ""`, i)

			}
		})
		require.NoError(t, err)
	}

	// test capitalization as well
	agentIDs, err := FindAgents(ctx, store.AgentIndex(ctx), model.FieldConfigurationPending, "C:2")
	require.NoError(t, err)
	require.Equal(t, 4, len(agentIDs))

	agentIDs, err = FindAgents(ctx, store.AgentIndex(ctx), model.FieldConfigurationFuture, "C:2")
	require.NoError(t, err)
	require.Equal(t, 0, len(agentIDs))

	agentIDs, err = FindAgents(ctx, store.AgentIndex(ctx), model.FieldConfigurationPending, "c:3")
	require.NoError(t, err)
	require.Equal(t, 6, len(agentIDs))

	agentIDs, err = FindAgents(ctx, store.AgentIndex(ctx), model.FieldConfigurationFuture, "c:3")
	require.NoError(t, err)
	require.Equal(t, 4, len(agentIDs))

	agentIDs, err = FindAgents(ctx, store.AgentIndex(ctx), model.FieldConfigurationCurrent, "c:1")
	require.NoError(t, err)
	require.Equal(t, 10, len(agentIDs))

	agentIDs, err = FindAgents(ctx, store.AgentIndex(ctx), model.FieldConfigurationFuture, "")
	require.NoError(t, err)
	require.Equal(t, 4, len(agentIDs))
}

func runTestMaskSensitiveParameters(ctx context.Context, t *testing.T, store Store) {
	store.Clear()

	sensitiveSourceType := model.NewSourceType("source-type-1", []model.ParameterDefinition{
		{
			Name: "username",
			Type: "string",
		},
		{
			Name: "password",
			Type: "string",
			Options: model.ParameterOptions{
				Sensitive: true,
			},
		},
	}, []string{"macos"})
	sensitiveProcessorType := model.NewProcessorType("processor-type-1", []model.ParameterDefinition{
		{
			Name: "username",
			Type: "string",
		},
		{
			Name: "password",
			Type: "string",
			Options: model.ParameterOptions{
				Sensitive: true,
			},
		},
	})
	sensitiveDestination := model.NewDestination("destination-1", "destination-type-1", []model.Parameter{
		{
			Name:  "username",
			Value: "user1",
		},
		{
			Name:  "password",
			Value: "pw1",
		},
	})
	sensitiveDestinationType := model.NewDestinationType("destination-type-1", []model.ParameterDefinition{
		{
			Name: "username",
			Type: "string",
		},
		{
			Name: "password",
			Type: "string",
			Options: model.ParameterOptions{
				Sensitive: true,
			},
		},
	})
	sensitiveDestination.Spec.Processors = []model.ResourceConfiguration{
		{
			ParameterizedSpec: model.ParameterizedSpec{
				Type: "processor-type-1",
				Parameters: []model.Parameter{
					{
						Name:  "username",
						Value: "user2",
					},
					{
						Name:  "password",
						Value: "pw2",
					},
				},
			},
		},
	}
	sensitiveConfiguration := model.NewConfigurationWithSpec("configuration-1", model.ConfigurationSpec{
		Sources: []model.ResourceConfiguration{
			{
				ParameterizedSpec: model.ParameterizedSpec{
					Type: "source-type-1",
					Parameters: []model.Parameter{
						{
							Name:  "username",
							Value: "user3",
						},
						{
							Name:  "password",
							Value: "pw3",
						},
					},
					Processors: []model.ResourceConfiguration{
						{
							ParameterizedSpec: model.ParameterizedSpec{
								Type: "processor-type-1",
								Parameters: []model.Parameter{
									{
										Name:  "username",
										Value: "user4",
									},
									{
										Name:  "password",
										Value: "pw4",
									},
								},
							},
						},
					},
				},
			},
		},
		Destinations: []model.ResourceConfiguration{
			{
				Name: "destination-1",
				ParameterizedSpec: model.ParameterizedSpec{
					Processors: []model.ResourceConfiguration{
						{
							ParameterizedSpec: model.ParameterizedSpec{
								Type: "processor-type-1",
								Parameters: []model.Parameter{
									{
										Name:  "username",
										Value: "user5",
									},
									{
										Name:  "password",
										Value: "pw5",
									},
								},
							},
						},
					},
				},
			},
		},
	})

	seq := util.NewTestSequence(t)

	seq.Run("create the resources", func(t *testing.T) {
		_, err := store.ApplyResources(ctx, []model.Resource{
			sensitiveSourceType,
			sensitiveProcessorType,
			sensitiveDestinationType,
			sensitiveDestination,
			sensitiveConfiguration,
		})
		require.NoError(t, err)
	})

	seq.Run("test that the sensitive parameters are not returned by accessor", func(t *testing.T) {
		destination, err := store.Destination(ctx, "destination-1")
		require.NoError(t, err)
		require.Equal(t, "user1", destination.Spec.Parameters[0].Value)
		require.Equal(t, model.SensitiveParameterPlaceholder, destination.Spec.Parameters[1].Value)
		require.True(t, destination.Spec.Parameters[1].Sensitive)
		require.Equal(t, "user2", destination.Spec.Processors[0].Parameters[0].Value)
		require.Equal(t, model.SensitiveParameterPlaceholder, destination.Spec.Processors[0].Parameters[1].Value)
		require.True(t, destination.Spec.Parameters[1].Sensitive)

		config, err := store.Configuration(ctx, "configuration-1")
		require.NoError(t, err)
		require.Equal(t, "user3", config.Spec.Sources[0].Parameters[0].Value)
		require.Equal(t, model.SensitiveParameterPlaceholder, config.Spec.Sources[0].Parameters[1].Value)
		require.True(t, config.Spec.Sources[0].Parameters[1].Sensitive)
		require.Equal(t, "user4", config.Spec.Sources[0].Processors[0].Parameters[0].Value)
		require.Equal(t, model.SensitiveParameterPlaceholder, config.Spec.Sources[0].Processors[0].Parameters[1].Value)
		require.True(t, config.Spec.Sources[0].Processors[0].Parameters[1].Sensitive)
	})

	seq.Run("test that the sensitive parameters are not returned by list", func(t *testing.T) {
		destinations, err := store.Destinations(ctx)
		require.NoError(t, err)
		require.Equal(t, 1, len(destinations))
		destination := destinations[0]
		require.Equal(t, "user1", destination.Spec.Parameters[0].Value)
		require.Equal(t, model.SensitiveParameterPlaceholder, destination.Spec.Parameters[1].Value)
		require.True(t, destination.Spec.Parameters[1].Sensitive)
		require.Equal(t, "user2", destination.Spec.Processors[0].Parameters[0].Value)
		require.Equal(t, model.SensitiveParameterPlaceholder, destination.Spec.Processors[0].Parameters[1].Value)
		require.True(t, destination.Spec.Processors[0].Parameters[1].Sensitive)

		configs, err := store.Configurations(ctx)
		require.NoError(t, err)
		require.Equal(t, 1, len(configs))
		config := configs[0]
		require.NoError(t, err)
		require.Equal(t, "user3", config.Spec.Sources[0].Parameters[0].Value)
		require.Equal(t, model.SensitiveParameterPlaceholder, config.Spec.Sources[0].Parameters[1].Value)
		require.True(t, config.Spec.Sources[0].Parameters[1].Sensitive)
		require.Equal(t, "user4", config.Spec.Sources[0].Processors[0].Parameters[0].Value)
		require.Equal(t, model.SensitiveParameterPlaceholder, config.Spec.Sources[0].Processors[0].Parameters[1].Value)
		require.True(t, config.Spec.Sources[0].Processors[0].Parameters[1].Sensitive)
	})

	seq.Run("test that the sensitive parameters are not returned by delete", func(t *testing.T) {
		config, err := store.DeleteConfiguration(ctx, "configuration-1")
		require.NoError(t, err)
		require.Equal(t, "user3", config.Spec.Sources[0].Parameters[0].Value)
		require.Equal(t, model.SensitiveParameterPlaceholder, config.Spec.Sources[0].Parameters[1].Value)
		require.True(t, config.Spec.Sources[0].Parameters[1].Sensitive)
		require.Equal(t, "user4", config.Spec.Sources[0].Processors[0].Parameters[0].Value)
		require.Equal(t, model.SensitiveParameterPlaceholder, config.Spec.Sources[0].Processors[0].Parameters[1].Value)
		require.True(t, config.Spec.Sources[0].Processors[0].Parameters[1].Sensitive)

		destination, err := store.DeleteDestination(ctx, "destination-1")
		require.NoError(t, err)
		require.Equal(t, "user1", destination.Spec.Parameters[0].Value)
		require.Equal(t, model.SensitiveParameterPlaceholder, destination.Spec.Parameters[1].Value)
		require.Equal(t, "user2", destination.Spec.Processors[0].Parameters[0].Value)
		require.Equal(t, model.SensitiveParameterPlaceholder, destination.Spec.Processors[0].Parameters[1].Value)
		require.True(t, destination.Spec.Processors[0].Parameters[1].Sensitive)
	})

	seq.Run("add the destination again", func(t *testing.T) {
		_, err := store.ApplyResources(ctx, []model.Resource{
			sensitiveDestination,
			sensitiveConfiguration,
		})
		require.NoError(t, err)
	})

	seq.Run("edit the destination and confirm that the sensitive parameters are preserved", func(t *testing.T) {
		modified, err := store.Destination(ctx, "destination-1")
		require.NoError(t, err)
		modified.Spec.Parameters[0].Value = "user3"
		statuses, err := store.ApplyResources(ctx, []model.Resource{modified})
		require.NoError(t, err)
		require.Equal(t, model.StatusConfigured, statuses[0].Status)

		// confirm that the sensitive parameters are masked
		destination, err := store.Destination(ctx, "destination-1")
		require.NoError(t, err)
		require.Equal(t, "user3", destination.Spec.Parameters[0].Value)
		require.Equal(t, model.SensitiveParameterPlaceholder, destination.Spec.Parameters[1].Value)
		require.True(t, destination.Spec.Parameters[1].Sensitive)

		// confirm that the sensitive parameters are preserved
		noMaskingCtx := model.ContextWithoutSensitiveParameterMasking(ctx)
		destination, err = store.Destination(noMaskingCtx, "destination-1")
		require.NoError(t, err)
		require.Equal(t, "user3", destination.Spec.Parameters[0].Value)
		require.Equal(t, "pw1", destination.Spec.Parameters[1].Value)
		require.True(t, destination.Spec.Parameters[1].Sensitive)
	})

	seq.Run("edit the configuration and confirm at the sensitive parameters are perserved", func(t *testing.T) {
		modified, err := store.Configuration(ctx, "configuration-1")
		require.NoError(t, err)
		modified.Spec.Sources[0].Parameters[0].Value = "user5"
		modified.Spec.Sources[0].Processors[0].Parameters[0].Value = "user6"
		statuses, err := store.ApplyResources(ctx, []model.Resource{modified})
		require.NoError(t, err)
		require.Equal(t, model.StatusConfigured, statuses[0].Status)

		// confirm that the sensitive parameters are masked
		config, err := store.Configuration(ctx, "configuration-1")
		require.NoError(t, err)
		require.Equal(t, "user5", config.Spec.Sources[0].Parameters[0].Value)
		require.Equal(t, model.SensitiveParameterPlaceholder, config.Spec.Sources[0].Parameters[1].Value)
		require.True(t, config.Spec.Sources[0].Parameters[1].Sensitive)
		require.Equal(t, "user6", config.Spec.Sources[0].Processors[0].Parameters[0].Value)
		require.Equal(t, model.SensitiveParameterPlaceholder, config.Spec.Sources[0].Processors[0].Parameters[1].Value)
		require.True(t, config.Spec.Sources[0].Processors[0].Parameters[1].Sensitive)

		// confirm that the sensitive parameters are preserved
		noMaskingCtx := model.ContextWithoutSensitiveParameterMasking(ctx)
		config, err = store.Configuration(noMaskingCtx, "configuration-1")
		require.NoError(t, err)
		require.Equal(t, "user5", config.Spec.Sources[0].Parameters[0].Value)
		require.Equal(t, "pw3", config.Spec.Sources[0].Parameters[1].Value)
		require.True(t, config.Spec.Sources[0].Parameters[1].Sensitive)
		require.Equal(t, "user6", config.Spec.Sources[0].Processors[0].Parameters[0].Value)
		require.Equal(t, "pw4", config.Spec.Sources[0].Processors[0].Parameters[1].Value)
		require.True(t, config.Spec.Sources[0].Processors[0].Parameters[1].Sensitive)
	})

	seq.Run("edit the destination and change the sensitive parameter", func(t *testing.T) {
		modified, err := store.Destination(ctx, "destination-1")
		require.NoError(t, err)
		modified.Spec.Parameters[1].Value = "pw3"
		statuses, err := store.ApplyResources(ctx, []model.Resource{modified})
		require.NoError(t, err)
		require.Equal(t, model.StatusConfigured, statuses[0].Status)

		// confirm that the sensitive parameters are masked
		destination, err := store.Destination(ctx, "destination-1")
		require.NoError(t, err)
		require.Equal(t, "user3", destination.Spec.Parameters[0].Value)
		require.Equal(t, model.SensitiveParameterPlaceholder, destination.Spec.Parameters[1].Value)
		require.True(t, destination.Spec.Parameters[1].Sensitive)

		// confirm that the sensitive parameters are preserved
		noMaskingCtx := model.ContextWithoutSensitiveParameterMasking(ctx)
		destination, err = store.Destination(noMaskingCtx, "destination-1")
		require.NoError(t, err)
		require.Equal(t, "user3", destination.Spec.Parameters[0].Value)
		require.Equal(t, "pw3", destination.Spec.Parameters[1].Value)
		require.True(t, destination.Spec.Parameters[1].Sensitive)
	})

	seq.Run("edit the configuration and change the sensitive parameters", func(t *testing.T) {
		modified, err := store.Configuration(ctx, "configuration-1")
		require.NoError(t, err)
		modified.Spec.Sources[0].Parameters[1].Value = "pw10"
		modified.Spec.Sources[0].Processors[0].Parameters[1].Value = "pw11"
		statuses, err := store.ApplyResources(ctx, []model.Resource{modified})
		require.NoError(t, err)
		require.Equal(t, model.StatusConfigured, statuses[0].Status)

		// confirm that the sensitive parameters are masked
		config, err := store.Configuration(ctx, "configuration-1")
		require.NoError(t, err)
		require.Equal(t, "user5", config.Spec.Sources[0].Parameters[0].Value)
		require.Equal(t, model.SensitiveParameterPlaceholder, config.Spec.Sources[0].Parameters[1].Value)
		require.True(t, config.Spec.Sources[0].Parameters[1].Sensitive)
		require.Equal(t, "user6", config.Spec.Sources[0].Processors[0].Parameters[0].Value)
		require.Equal(t, model.SensitiveParameterPlaceholder, config.Spec.Sources[0].Processors[0].Parameters[1].Value)
		require.True(t, config.Spec.Sources[0].Processors[0].Parameters[1].Sensitive)

		// confirm that the sensitive parameters are preserved
		noMaskingCtx := model.ContextWithoutSensitiveParameterMasking(ctx)
		config, err = store.Configuration(noMaskingCtx, "configuration-1")
		require.NoError(t, err)
		require.Equal(t, "user5", config.Spec.Sources[0].Parameters[0].Value)
		require.Equal(t, "pw10", config.Spec.Sources[0].Parameters[1].Value)
		require.True(t, config.Spec.Sources[0].Parameters[1].Sensitive)
		require.Equal(t, "user6", config.Spec.Sources[0].Processors[0].Parameters[0].Value)
		require.Equal(t, "pw11", config.Spec.Sources[0].Processors[0].Parameters[1].Value)
		require.True(t, config.Spec.Sources[0].Processors[0].Parameters[1].Sensitive)
	})

	seq.Run("edit a dependency and confirm that the sensitive parameters are preserved", func(t *testing.T) {
		config, err := store.Configuration(ctx, "configuration-1")
		require.NoError(t, err)
		oldVersion := config.Version()

		_, err = store.StartRollout(ctx, "configuration-1", nil)
		require.NoError(t, err)
		_, err = store.UpdateRollout(ctx, "configuration-1")
		require.NoError(t, err)

		modified, err := model.Clone(sensitiveSourceType)
		require.NoError(t, err)
		modified.Spec.Parameters = append(modified.Spec.Parameters, model.ParameterDefinition{
			Name: "param3",
			Type: "string",
		})
		statuses, err := store.ApplyResources(ctx, []model.Resource{modified})
		require.NoError(t, err)
		require.Equal(t, model.StatusConfigured, statuses[0].Status)

		// confirm the new version of the processor-type
		sourceType, err := store.SourceType(ctx, "source-type-1")
		require.NoError(t, err)
		require.Equal(t, model.Version(2), sourceType.Version())

		// confirm that the configuration is updated
		noMaskingCtx := model.ContextWithoutSensitiveParameterMasking(ctx)
		config, err = store.Configuration(noMaskingCtx, "configuration-1")
		require.NoError(t, err)
		require.Equal(t, "source-type-1:2", config.Spec.Sources[0].Type)
		require.Equal(t, "user5", config.Spec.Sources[0].Parameters[0].Value)
		require.Equal(t, "pw10", config.Spec.Sources[0].Parameters[1].Value)
		require.True(t, config.Spec.Sources[0].Parameters[1].Sensitive)
		require.Equal(t, "user6", config.Spec.Sources[0].Processors[0].Parameters[0].Value)
		require.Equal(t, "pw11", config.Spec.Sources[0].Processors[0].Parameters[1].Value)
		require.True(t, config.Spec.Sources[0].Processors[0].Parameters[1].Sensitive)
		require.Equal(t, model.Version(oldVersion+1), config.Version())

		// rollout these changes
		_, err = store.StartRollout(ctx, "configuration-1", nil)
		require.NoError(t, err)
		_, err = store.UpdateRollout(ctx, "configuration-1")
		require.NoError(t, err)
	})

	seq.Run("edit a processor type dependency and confirm that the sensitive parameters are also preserved", func(t *testing.T) {
		config, err := store.Configuration(ctx, "configuration-1")
		require.NoError(t, err)
		oldVersion := config.Version()

		modified, err := model.Clone(sensitiveProcessorType)
		require.NoError(t, err)
		modified.Spec.Parameters = append(modified.Spec.Parameters, model.ParameterDefinition{
			Name: "param4",
			Type: "string",
		})
		statuses, err := store.ApplyResources(ctx, []model.Resource{modified})
		require.NoError(t, err)
		require.Equal(t, model.StatusConfigured, statuses[0].Status)

		// confirm the new version of the processor-type
		processorType, err := store.ProcessorType(ctx, "processor-type-1")
		require.NoError(t, err)
		require.Equal(t, model.Version(2), processorType.Version())

		// confirm that the configuration is updated
		noMaskingCtx := model.ContextWithoutSensitiveParameterMasking(ctx)
		config, err = store.Configuration(noMaskingCtx, "configuration-1")
		require.NoError(t, err)
		require.Equal(t, "source-type-1:2", config.Spec.Sources[0].Type)
		require.Equal(t, "user5", config.Spec.Sources[0].Parameters[0].Value)
		require.Equal(t, "pw10", config.Spec.Sources[0].Parameters[1].Value)
		require.Equal(t, "processor-type-1:2", config.Spec.Sources[0].Processors[0].Type)
		require.True(t, config.Spec.Sources[0].Parameters[1].Sensitive)
		require.Equal(t, "user6", config.Spec.Sources[0].Processors[0].Parameters[0].Value)
		require.Equal(t, "pw11", config.Spec.Sources[0].Processors[0].Parameters[1].Value)
		require.True(t, config.Spec.Sources[0].Processors[0].Parameters[1].Sensitive)
		require.Equal(t, "processor-type-1:2", config.Spec.Destinations[0].Processors[0].Type)
		require.Equal(t, model.Version(oldVersion+1), config.Version())
	})

	seq.Run("confirm that all versions of the resource have masked parameters", func(t *testing.T) {
		s, ok := store.(ArchiveStore)
		if !ok {
			t.Skip("store does not implement ArchiveStore")
			return
		}
		history, err := s.ResourceHistory(ctx, model.KindConfiguration, "configuration-1")
		require.NoError(t, err)

		configs, err := model.Parse[*model.Configuration](history)
		require.NoError(t, err)

		for _, config := range configs {
			require.Equal(t, model.SensitiveParameterPlaceholder, config.Spec.Sources[0].Parameters[1].Value)
			require.Equal(t, model.SensitiveParameterPlaceholder, config.Spec.Sources[0].Processors[0].Parameters[1].Value)
			require.Equal(t, model.SensitiveParameterPlaceholder, config.Spec.Destinations[0].Processors[0].Parameters[1].Value)
		}
	})

	seq.Run("get an old version and confirm that the sensitive parameters are masked", func(t *testing.T) {
		config, err := store.Configuration(ctx, "configuration-1:1")
		require.NoError(t, err)
		require.Equal(t, "source-type-1:1", config.Spec.Sources[0].Type)
		require.Equal(t, "user5", config.Spec.Sources[0].Parameters[0].Value)
		require.Equal(t, model.SensitiveParameterPlaceholder, config.Spec.Sources[0].Parameters[1].Value)
	})
}
func runRolloutToDisconnectedAgentsTest(ctx context.Context, t *testing.T, store Store) {
	c1 := model.NewConfigurationWithSpec("c1", model.ConfigurationSpec{
		Raw: "service:",
		Selector: model.AgentSelector{
			MatchLabels: model.MatchLabels{
				"configuration": "c1",
			},
		},
	})

	c1agentIDs := []string{}
	for i := 0; i < 10; i++ {
		c1agentIDs = append(c1agentIDs, fmt.Sprintf("c1agent-%d", i))
	}

	simulateAgentConfiguration := func(configurationName string, expectAgents int) {
		agents, err := store.Agents(ctx, WithQuery(search.ParseQuery("rollout-pending:"+configurationName)))
		require.NoError(t, err)
		require.Len(t, agents, expectAgents)
		configuration, err := store.Configuration(ctx, configurationName)
		require.NoError(t, err)
		for _, agent := range agents {
			_, err = store.UpsertAgent(ctx, agent.ID, func(agent *model.Agent) {
				agent.SetCurrentConfiguration(configuration)
			})
		}
	}

	seq := util.NewTestSequence(t)
	// these tests are run in order, so we can use the same config for each test
	seq.Run("setup: new configurations", func(t *testing.T) {
		config, err := model.Clone(c1)
		require.NoError(t, err)
		status, err := store.ApplyResources(ctx, []model.Resource{config})
		require.NoError(t, err)
		require.Len(t, status, 1)
		require.Equal(t, model.StatusCreated, status[0].Status)

		config, err = store.Configuration(ctx, c1.Name())
		require.NoError(t, err)
		require.Equal(t, config.Status.Rollout.Status, model.RolloutStatusPending)
		require.False(t, config.IsCurrent())
		require.False(t, config.IsPending())
		require.True(t, config.IsLatest())
	})
	seq.Run("setup: create 10 agents for each config", func(t *testing.T) {
		for i, id := range c1agentIDs {
			agent, err := store.UpsertAgent(ctx, id, func(current *model.Agent) {
				current.Status = model.Connected
				if i%2 == 0 {
					current.Status = model.Disconnected
				}
			})
			require.NoError(t, err)
			require.Equal(t, id, agent.ID)
		}
	})
	seq.Run("assign configuration to agents", func(t *testing.T) {
		_, err := store.UpsertAgents(ctx, c1agentIDs, func(agent *model.Agent) {
			agent.Labels = model.LabelsFromValidatedMap(map[string]string{
				"configuration": "c1",
			})
		})
		require.NoError(t, err)
	})
	seq.Run("agents are assigned a future configuration", func(t *testing.T) {
		agents, err := store.Agents(ctx)
		require.NoError(t, err)
		for _, agent := range agents {
			if strings.HasPrefix(agent.ID, "c1") {
				require.Equal(t, "c1", agent.Labels.Set["configuration"])
				require.Equal(t, "", agent.ConfigurationStatus.Current)
				require.Equal(t, "", agent.ConfigurationStatus.Pending)
				require.Equal(t, "c1:1", agent.ConfigurationStatus.Future)
			}
		}
	})
	seq.Run("update rollout does nothing because the rollout is pending", func(t *testing.T) {
		configuration, err := store.UpdateRollout(ctx, "c1")
		require.NoError(t, err)
		require.NotNil(t, configuration)
		require.Equal(t, model.RolloutStatusPending, configuration.Status.Rollout.Status)
		require.False(t, configuration.IsCurrent())
		require.False(t, configuration.IsPending())
		require.True(t, configuration.IsLatest())
	})
	seq.Run("start the rollout", func(t *testing.T) {
		configuration, err := store.StartRollout(ctx, "c1", &model.RolloutOptions{
			PhaseAgentCount: model.PhaseAgentCount{
				Initial:    3,
				Multiplier: 2,
				Maximum:    5,
			},
			MaxErrors: 1,
		})
		require.NoError(t, err)
		require.NotNil(t, configuration)
		require.Equal(t, model.RolloutStatusStarted, configuration.Status.Rollout.Status)
		require.Equal(t, model.RolloutProgress{
			Pending:   3,
			Completed: 0,
			Errors:    0,
			Waiting:   2,
		}, configuration.Status.Rollout.Progress)
		require.False(t, configuration.IsCurrent())
		require.True(t, configuration.IsPending())
		require.True(t, configuration.IsLatest())
	})
	seq.Run("get the rollout status", func(t *testing.T) {
		configuration, err := store.Configuration(ctx, "c1")
		require.NoError(t, err)
		require.NotNil(t, configuration)
		require.Equal(t, model.RolloutStatusStarted, configuration.Status.Rollout.Status)
		require.False(t, configuration.IsCurrent())
		require.True(t, configuration.IsPending())
		require.True(t, configuration.IsLatest())
	})
	seq.Run("update the rollout again but no progress because the agents haven't been configured", func(t *testing.T) {
		c1, err := store.Configuration(ctx, "c1:1")
		require.NoError(t, err)
		require.NotNil(t, c1)
		require.Equal(t, model.RolloutStatusStarted, c1.Status.Rollout.Status)
		require.False(t, c1.IsCurrent())
		require.True(t, c1.IsPending())
		require.True(t, c1.IsLatest())

		configuration, err := store.UpdateRollout(ctx, "c1:1")
		require.NoError(t, err)
		require.NotNil(t, configuration)
		require.Equal(t, model.RolloutStatusStarted, configuration.Status.Rollout.Status)
		require.Equal(t, model.RolloutProgress{
			Pending:   3,
			Completed: 0,
			Errors:    0,
			Waiting:   2,
		}, configuration.Status.Rollout.Progress)
		require.Equal(t, model.Version(1), configuration.Version())
		require.False(t, configuration.IsCurrent())
		require.True(t, configuration.IsPending())
		require.True(t, configuration.IsLatest())
	})
	seq.Run("find the agents pending and move them to current, simulating successful configuration", func(t *testing.T) {
		simulateAgentConfiguration("c1:1", 3)
	})
	seq.Run("update the rollout again and make progress", func(t *testing.T) {
		configuration, err := store.UpdateRollout(ctx, "c1:1")
		require.NoError(t, err)
		require.NotNil(t, configuration)
		require.Equal(t, model.RolloutStatusStarted, configuration.Status.Rollout.Status)
		require.Equal(t, model.RolloutProgress{
			Pending:   2,
			Completed: 3,
			Errors:    0,
			Waiting:   0,
		}, configuration.Status.Rollout.Progress)
		require.Equal(t, 2, configuration.Status.Rollout.Phase)
		require.Equal(t, model.Version(1), configuration.Version())
		require.False(t, configuration.IsCurrent())
		require.True(t, configuration.IsPending())
		require.True(t, configuration.IsLatest())
	})
	seq.Run("find the agents pending and move them to current, simulating successful configuration", func(t *testing.T) {
		simulateAgentConfiguration("c1:1", 2)
	})
	seq.Run("update the rollout again and make progress", func(t *testing.T) {
		configuration, err := store.UpdateRollout(ctx, "c1:1")
		require.NoError(t, err)
		require.NotNil(t, configuration)
		require.Equal(t, model.RolloutStatusStable, configuration.Status.Rollout.Status)
		require.Equal(t, model.RolloutProgress{
			Pending:   0,
			Completed: 5,
			Errors:    0,
			Waiting:   0,
		}, configuration.Status.Rollout.Progress)
		require.Equal(t, 2, configuration.Status.Rollout.Phase)
		require.Equal(t, model.Version(1), configuration.Version())
		require.True(t, configuration.IsCurrent())
		require.True(t, configuration.IsPending())
		require.True(t, configuration.IsLatest())
	})

}
