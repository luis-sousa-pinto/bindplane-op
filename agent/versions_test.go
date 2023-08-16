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

package agent

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	versionMocks "github.com/observiq/bindplane-op/agent/mocks"
	"github.com/observiq/bindplane-op/eventbus"
	"github.com/observiq/bindplane-op/model"
	"github.com/observiq/bindplane-op/store"
	storeMocks "github.com/observiq/bindplane-op/store/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"
)

func TestLatestVersionString(t *testing.T) {
	t.Run("versions available", func(t *testing.T) {
		agentVersions := readJSONFile[[]*model.AgentVersion](t, filepath.Join("testfiles", "agent-versions.json"))
		eventSrc := eventbus.NewSource[store.BasicEventUpdates]()

		mockClient := versionMocks.NewMockVersionClient(t)

		mockStore := storeMocks.NewMockStore(t)
		mockStore.On("Updates", mock.Anything).Return(eventSrc).Maybe()
		mockStore.On("AgentVersions", mock.Anything).Return(agentVersions, nil)

		cancelContext, cancel := context.WithCancel(context.Background())
		defer cancel()

		versions := NewVersions(cancelContext, mockClient, mockStore, VersionsSettings{
			Logger:                    zaptest.NewLogger(t),
			SyncAgentVersionsInterval: 0,
		})

		latest := versions.LatestVersionString(cancelContext)
		require.Equal(t, "v1.30.0", latest)
	})

	t.Run("version is cached", func(t *testing.T) {
		agentVersions := readJSONFile[[]*model.AgentVersion](t, filepath.Join("testfiles", "agent-versions.json"))
		eventSrc := eventbus.NewSource[store.BasicEventUpdates]()

		mockClient := versionMocks.NewMockVersionClient(t)

		mockStore := storeMocks.NewMockStore(t)
		mockStore.On("Updates", mock.Anything).Return(eventSrc).Maybe()
		mockStore.On("AgentVersions", mock.Anything).Return(agentVersions, nil).Once()

		cancelContext, cancel := context.WithCancel(context.Background())
		defer cancel()

		versions := NewVersions(cancelContext, mockClient, mockStore, VersionsSettings{
			Logger:                    zaptest.NewLogger(t),
			SyncAgentVersionsInterval: 0,
		})

		latest := versions.LatestVersionString(cancelContext)
		require.Equal(t, "v1.30.0", latest)

		latest = versions.LatestVersionString(cancelContext)
		require.Equal(t, "v1.30.0", latest)
	})

	t.Run("no versions available", func(t *testing.T) {
		eventSrc := eventbus.NewSource[store.BasicEventUpdates]()

		mockClient := versionMocks.NewMockVersionClient(t)

		mockStore := storeMocks.NewMockStore(t)
		mockStore.On("Updates", mock.Anything).Return(eventSrc).Maybe()
		mockStore.On("AgentVersions", mock.Anything).Return([]*model.AgentVersion{}, nil)

		cancelContext, cancel := context.WithCancel(context.Background())
		defer cancel()

		versions := NewVersions(cancelContext, mockClient, mockStore, VersionsSettings{
			Logger:                    zaptest.NewLogger(t),
			SyncAgentVersionsInterval: 0,
		})

		latest := versions.LatestVersionString(cancelContext)
		require.Equal(t, "", latest)
	})

	t.Run("error retrieving versions", func(t *testing.T) {
		eventSrc := eventbus.NewSource[store.BasicEventUpdates]()

		mockClient := versionMocks.NewMockVersionClient(t)

		mockStore := storeMocks.NewMockStore(t)
		mockStore.On("Updates", mock.Anything).Return(eventSrc).Maybe()
		mockStore.On("AgentVersions", mock.Anything).Return(nil, errors.New("failed to get versions"))

		cancelContext, cancel := context.WithCancel(context.Background())
		defer cancel()

		versions := NewVersions(cancelContext, mockClient, mockStore, VersionsSettings{
			Logger:                    zaptest.NewLogger(t),
			SyncAgentVersionsInterval: 0,
		})

		latest := versions.LatestVersionString(cancelContext)
		require.Equal(t, "", latest)
	})

}

func TestLatestVersion(t *testing.T) {
	t.Run("versions available", func(t *testing.T) {
		agentVersions := readJSONFile[[]*model.AgentVersion](t, filepath.Join("testfiles", "agent-versions.json"))
		expectedVersion := readJSONFile[*model.AgentVersion](t, filepath.Join("testfiles", "agent-version-v1.30.0.json"))
		eventSrc := eventbus.NewSource[store.BasicEventUpdates]()

		mockClient := versionMocks.NewMockVersionClient(t)

		mockStore := storeMocks.NewMockStore(t)
		mockStore.On("Updates", mock.Anything).Return(eventSrc).Maybe()
		mockStore.On("AgentVersions", mock.Anything).Return(agentVersions, nil)

		cancelContext, cancel := context.WithCancel(context.Background())
		defer cancel()

		versions := NewVersions(cancelContext, mockClient, mockStore, VersionsSettings{
			Logger:                    zaptest.NewLogger(t),
			SyncAgentVersionsInterval: 0,
		})

		latest, err := versions.LatestVersion(cancelContext)
		require.NoError(t, err)
		require.Equal(t, expectedVersion, latest)
	})

	t.Run("version is cached", func(t *testing.T) {
		agentVersions := readJSONFile[[]*model.AgentVersion](t, filepath.Join("testfiles", "agent-versions.json"))
		expectedVersion := readJSONFile[*model.AgentVersion](t, filepath.Join("testfiles", "agent-version-v1.30.0.json"))
		eventSrc := eventbus.NewSource[store.BasicEventUpdates]()

		mockClient := versionMocks.NewMockVersionClient(t)

		mockStore := storeMocks.NewMockStore(t)
		mockStore.On("Updates", mock.Anything).Return(eventSrc).Maybe()
		mockStore.On("AgentVersions", mock.Anything).Return(agentVersions, nil).Once()

		cancelContext, cancel := context.WithCancel(context.Background())
		defer cancel()

		versions := NewVersions(cancelContext, mockClient, mockStore, VersionsSettings{
			Logger:                    zaptest.NewLogger(t),
			SyncAgentVersionsInterval: 0,
		})

		latest, err := versions.LatestVersion(cancelContext)
		require.NoError(t, err)
		require.Equal(t, expectedVersion, latest)

		latest, err = versions.LatestVersion(cancelContext)
		require.NoError(t, err)
		require.Equal(t, expectedVersion, latest)
	})

	t.Run("no versions available", func(t *testing.T) {
		eventSrc := eventbus.NewSource[store.BasicEventUpdates]()

		mockClient := versionMocks.NewMockVersionClient(t)

		mockStore := storeMocks.NewMockStore(t)
		mockStore.On("Updates", mock.Anything).Return(eventSrc).Maybe()
		mockStore.On("AgentVersions", mock.Anything).Return([]*model.AgentVersion{}, nil)

		cancelContext, cancel := context.WithCancel(context.Background())
		defer cancel()
		versions := NewVersions(cancelContext, mockClient, mockStore, VersionsSettings{
			Logger:                    zaptest.NewLogger(t),
			SyncAgentVersionsInterval: 0,
		})

		latest, err := versions.LatestVersion(cancelContext)
		require.NoError(t, err)
		require.Equal(t, (*model.AgentVersion)(nil), latest)
	})

	t.Run("error retrieving versions", func(t *testing.T) {
		eventSrc := eventbus.NewSource[store.BasicEventUpdates]()

		mockClient := versionMocks.NewMockVersionClient(t)

		mockStore := storeMocks.NewMockStore(t)
		mockStore.On("Updates", mock.Anything).Return(eventSrc).Maybe()
		agentVersionsErr := errors.New("failed to get versions")
		mockStore.On("AgentVersions", mock.Anything).Return(nil, agentVersionsErr)

		cancelContext, cancel := context.WithCancel(context.Background())
		defer cancel()
		versions := NewVersions(cancelContext, mockClient, mockStore, VersionsSettings{
			Logger:                    zaptest.NewLogger(t),
			SyncAgentVersionsInterval: 0,
		})

		_, err := versions.LatestVersion(cancelContext)
		require.ErrorIs(t, err, agentVersionsErr)
		require.ErrorContains(t, err, "agent versions:")
	})
}

func TestVersion(t *testing.T) {
	t.Run("version exists", func(t *testing.T) {
		agentVersion := readJSONFile[*model.AgentVersion](t, filepath.Join("testfiles", "agent-version-v1.30.0.json"))
		eventSrc := eventbus.NewSource[store.BasicEventUpdates]()

		mockClient := versionMocks.NewMockVersionClient(t)

		mockStore := storeMocks.NewMockStore(t)
		mockStore.On("Updates", mock.Anything).Return(eventSrc).Maybe()
		mockStore.On("AgentVersion", mock.Anything, "observiq-otel-collector-v1.30.0").Return(agentVersion, nil)

		cancelContext, cancel := context.WithCancel(context.Background())
		defer cancel()
		versions := NewVersions(cancelContext, mockClient, mockStore, VersionsSettings{
			Logger:                    zaptest.NewLogger(t),
			SyncAgentVersionsInterval: 0,
		})

		version, err := versions.Version(cancelContext, "v1.30.0")
		require.NoError(t, err)
		require.Equal(t, agentVersion, version)
	})

	t.Run("version does not exist", func(t *testing.T) {
		eventSrc := eventbus.NewSource[store.BasicEventUpdates]()

		mockClient := versionMocks.NewMockVersionClient(t)

		mockStore := storeMocks.NewMockStore(t)
		mockStore.On("Updates", mock.Anything).Return(eventSrc).Maybe()
		mockStore.On("AgentVersion", mock.Anything, "observiq-otel-collector-v1.30.0").Return(nil, nil)

		cancelContext, cancel := context.WithCancel(context.Background())
		defer cancel()
		versions := NewVersions(cancelContext, mockClient, mockStore, VersionsSettings{
			Logger:                    zaptest.NewLogger(t),
			SyncAgentVersionsInterval: 0,
		})

		version, err := versions.Version(cancelContext, "v1.30.0")
		require.NoError(t, err)
		require.Equal(t, (*model.AgentVersion)(nil), version)
	})

	t.Run("store error", func(t *testing.T) {
		eventSrc := eventbus.NewSource[store.BasicEventUpdates]()

		mockClient := versionMocks.NewMockVersionClient(t)

		mockStore := storeMocks.NewMockStore(t)
		agentVersionErr := errors.New("error retriving version")
		mockStore.On("Updates", mock.Anything).Return(eventSrc).Maybe()
		mockStore.On("AgentVersion", mock.Anything, "observiq-otel-collector-v1.30.0").Return(nil, agentVersionErr)

		cancelContext, cancel := context.WithCancel(context.Background())
		defer cancel()
		versions := NewVersions(cancelContext, mockClient, mockStore, VersionsSettings{
			Logger:                    zaptest.NewLogger(t),
			SyncAgentVersionsInterval: 0,
		})

		_, err := versions.Version(cancelContext, "v1.30.0")
		require.ErrorIs(t, err, agentVersionErr)
		require.ErrorContains(t, err, "agent version by name:")
	})

	t.Run("latest version exists", func(t *testing.T) {
		agentVersion := readJSONFile[*model.AgentVersion](t, filepath.Join("testfiles", "agent-version-v1.30.0.json"))
		agentVersions := readJSONFile[[]*model.AgentVersion](t, filepath.Join("testfiles", "agent-versions.json"))
		eventSrc := eventbus.NewSource[store.BasicEventUpdates]()

		mockClient := versionMocks.NewMockVersionClient(t)

		mockStore := storeMocks.NewMockStore(t)
		mockStore.On("Updates", mock.Anything).Return(eventSrc).Maybe()
		mockStore.On("AgentVersions", mock.Anything).Return(agentVersions, nil)

		cancelContext, cancel := context.WithCancel(context.Background())
		defer cancel()
		versions := NewVersions(cancelContext, mockClient, mockStore, VersionsSettings{
			Logger:                    zaptest.NewLogger(t),
			SyncAgentVersionsInterval: 0,
		})

		version, err := versions.Version(cancelContext, "latest")
		require.NoError(t, err)
		require.Equal(t, agentVersion, version)
	})

	t.Run("latest version does not exist", func(t *testing.T) {
		eventSrc := eventbus.NewSource[store.BasicEventUpdates]()

		mockClient := versionMocks.NewMockVersionClient(t)

		mockStore := storeMocks.NewMockStore(t)
		mockStore.On("Updates", mock.Anything).Return(eventSrc).Maybe()
		mockStore.On("AgentVersions", mock.Anything).Return([]*model.AgentVersion{}, nil)

		cancelContext, cancel := context.WithCancel(context.Background())
		defer cancel()
		versions := NewVersions(cancelContext, mockClient, mockStore, VersionsSettings{
			Logger:                    zaptest.NewLogger(t),
			SyncAgentVersionsInterval: 0,
		})

		version, err := versions.Version(context.Background(), "latest")
		require.NoError(t, err)
		require.Equal(t, (*model.AgentVersion)(nil), version)
	})

	t.Run("latest version store error", func(t *testing.T) {
		eventSrc := eventbus.NewSource[store.BasicEventUpdates]()

		mockClient := versionMocks.NewMockVersionClient(t)

		mockStore := storeMocks.NewMockStore(t)
		agentVersionErr := errors.New("error retriving version")
		mockStore.On("Updates", mock.Anything).Return(eventSrc).Maybe()
		mockStore.On("AgentVersions", mock.Anything).Return(nil, agentVersionErr)

		cancelContext, cancel := context.WithCancel(context.Background())
		defer cancel()
		versions := NewVersions(cancelContext, mockClient, mockStore, VersionsSettings{
			Logger:                    zaptest.NewLogger(t),
			SyncAgentVersionsInterval: 0,
		})

		_, err := versions.Version(context.Background(), "latest")
		require.ErrorIs(t, err, agentVersionErr)
		require.ErrorContains(t, err, "latest version:")
	})

}

func TestSyncVersion(t *testing.T) {
	t.Run("version exists", func(t *testing.T) {
		expectedVersion := readJSONFile[*model.AgentVersion](t, filepath.Join("testfiles", "agent-version-v1.30.0.json"))
		eventSrc := eventbus.NewSource[store.BasicEventUpdates]()

		mockClient := versionMocks.NewMockVersionClient(t)
		mockClient.On("Version", "v1.30.0").Return(expectedVersion, nil)

		mockStore := storeMocks.NewMockStore(t)
		mockStore.On("Updates", mock.Anything).Return(eventSrc).Maybe()

		cancelContext, cancel := context.WithCancel(context.Background())
		defer cancel()
		versions := NewVersions(cancelContext, mockClient, mockStore, VersionsSettings{
			Logger:                    zaptest.NewLogger(t),
			SyncAgentVersionsInterval: 0,
		})

		version, err := versions.SyncVersion("v1.30.0")
		require.NoError(t, err)
		require.Equal(t, expectedVersion, version)
	})

	t.Run("fail to fetch version", func(t *testing.T) {
		eventSrc := eventbus.NewSource[store.BasicEventUpdates]()

		mockClient := versionMocks.NewMockVersionClient(t)
		versionErr := errors.New("failed to fetch version endpoint")
		mockClient.On("Version", "v1.30.0").Return(nil, versionErr)

		mockStore := storeMocks.NewMockStore(t)
		mockStore.On("Updates", mock.Anything).Return(eventSrc).Maybe()

		cancelContext, cancel := context.WithCancel(context.Background())
		defer cancel()
		versions := NewVersions(cancelContext, mockClient, mockStore, VersionsSettings{
			Logger:                    zaptest.NewLogger(t),
			SyncAgentVersionsInterval: 0,
		})

		_, err := versions.SyncVersion("v1.30.0")
		require.ErrorIs(t, err, versionErr)
	})

}

func TestSyncVersions(t *testing.T) {
	t.Run("version exists", func(t *testing.T) {
		agentVersions := readJSONFile[[]*model.AgentVersion](t, filepath.Join("testfiles", "agent-versions.json"))
		eventSrc := eventbus.NewSource[store.BasicEventUpdates]()

		mockClient := versionMocks.NewMockVersionClient(t)
		mockClient.On("Versions").Return(agentVersions, nil)

		mockStore := storeMocks.NewMockStore(t)
		mockStore.On("Updates", mock.Anything).Return(eventSrc).Maybe()

		cancelContext, cancel := context.WithCancel(context.Background())
		defer cancel()
		versions := NewVersions(cancelContext, mockClient, mockStore, VersionsSettings{
			Logger:                    zaptest.NewLogger(t),
			SyncAgentVersionsInterval: 0,
		})

		syncVersions, err := versions.SyncVersions()
		require.NoError(t, err)
		require.Equal(t, agentVersions, syncVersions)
	})

	t.Run("fail to fetch version", func(t *testing.T) {
		eventSrc := eventbus.NewSource[store.BasicEventUpdates]()

		mockClient := versionMocks.NewMockVersionClient(t)
		versionErr := errors.New("failed to fetch version endpoint")
		mockClient.On("Versions").Return(nil, versionErr)

		mockStore := storeMocks.NewMockStore(t)
		mockStore.On("Updates", mock.Anything).Return(eventSrc).Maybe()

		cancelContext, cancel := context.WithCancel(context.Background())
		defer cancel()
		versions := NewVersions(cancelContext, mockClient, mockStore, VersionsSettings{
			Logger:                    zaptest.NewLogger(t),
			SyncAgentVersionsInterval: 0,
		})

		_, err := versions.SyncVersions()
		require.ErrorIs(t, err, versionErr)
	})
}

func TestSyncAgentVersions(t *testing.T) {
	t.Run("syncs immediately", func(t *testing.T) {
		agentVersions := readJSONFile[[]*model.AgentVersion](t, filepath.Join("testfiles", "agent-versions.json"))
		eventSrc := eventbus.NewSource[store.BasicEventUpdates]()

		mockClient := versionMocks.NewMockVersionClient(t)
		mockClient.On("Versions").Return(agentVersions, nil)

		applyCalledChan := make(chan struct{})

		mockStore := storeMocks.NewMockStore(t)
		mockStore.On("Updates", mock.Anything).Return(eventSrc).Maybe()
		mockStore.On("ApplyResources", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			resources := args.Get(1).([]model.Resource)
			require.Equal(t, resourcesAsResourceSlice(agentVersions), resources)
			close(applyCalledChan)
		}).Return(resourceAsCreatedStatuses(agentVersions), nil)

		cancelContext, cancel := context.WithCancel(context.Background())
		defer cancel()
		NewVersions(cancelContext, mockClient, mockStore, VersionsSettings{
			Logger:                    zaptest.NewLogger(t),
			SyncAgentVersionsInterval: 1 * time.Hour,
		})

		select {
		case <-time.After(1 * time.Second):
			require.Fail(t, "Timed out while waiting for initial sync call")
		case <-applyCalledChan: // OK
		}
	})

	t.Run("syncs on timer", func(t *testing.T) {
		agentVersions := readJSONFile[[]*model.AgentVersion](t, filepath.Join("testfiles", "agent-versions.json"))
		eventSrc := eventbus.NewSource[store.BasicEventUpdates]()

		mockClient := versionMocks.NewMockVersionClient(t)
		mockClient.On("Versions").Return(agentVersions, nil)

		applyCalledChan := make(chan struct{})
		applyCalledCount := 0

		mockStore := storeMocks.NewMockStore(t)
		mockStore.On("Updates", mock.Anything).Return(eventSrc).Maybe()
		mockStore.On("ApplyResources", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			applyCalledCount++
			resources := args.Get(1).([]model.Resource)
			require.Equal(t, resourcesAsResourceSlice(agentVersions), resources)
			// We expect this function to be called twice, so we'll only close
			// the signalling channel after the second call
			if applyCalledCount == 2 {
				close(applyCalledChan)
			}
		}).Return(resourceAsCreatedStatuses(agentVersions), nil)

		cancelContext, cancel := context.WithCancel(context.Background())
		defer cancel()
		vers := NewVersions(cancelContext, mockClient, mockStore, VersionsSettings{
			Logger:                    zaptest.NewLogger(t),
			SyncAgentVersionsInterval: 0,
		})

		actualVersions := vers.(*versions)
		go actualVersions.syncAgentVersions(cancelContext, 1*time.Millisecond)

		select {
		case <-time.After(1 * time.Second):
			require.Fail(t, "Timed out while waiting for 2 sync calls")
		case <-applyCalledChan: // OK
		}
	})

	t.Run("exits on context cancel", func(t *testing.T) {
		agentVersions := readJSONFile[[]*model.AgentVersion](t, filepath.Join("testfiles", "agent-versions.json"))
		eventSrc := eventbus.NewSource[store.BasicEventUpdates]()

		mockClient := versionMocks.NewMockVersionClient(t)
		mockClient.On("Versions").Return(agentVersions, nil)

		mockStore := storeMocks.NewMockStore(t)
		mockStore.On("Updates", mock.Anything).Return(eventSrc).Maybe()
		mockStore.On("ApplyResources", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			resources := args.Get(1).([]model.Resource)
			require.Equal(t, resourcesAsResourceSlice(agentVersions), resources)
		}).Return(resourceAsCreatedStatuses(agentVersions), nil).Once()

		cancelContext, cancel := context.WithCancel(context.Background())
		// Cancel context immediately to create cancelled context
		cancel()

		vers := NewVersions(cancelContext, mockClient, mockStore, VersionsSettings{
			Logger:                    zaptest.NewLogger(t),
			SyncAgentVersionsInterval: 0,
		})

		actualVersions := vers.(*versions)
		doneChan := make(chan struct{})
		go func() {
			actualVersions.syncAgentVersions(cancelContext, 1*time.Millisecond)
			close(doneChan)
		}()

		select {
		case <-time.After(1 * time.Second):
			require.Fail(t, "Timed out while waiting for syncAgentVersions to return")
		case <-doneChan: // OK
		}

	})
}

func TestSyncAgentVersionsOnce(t *testing.T) {
	t.Run("successful sync", func(t *testing.T) {
		agentVersions := readJSONFile[[]*model.AgentVersion](t, filepath.Join("testfiles", "agent-versions.json"))
		eventSrc := eventbus.NewSource[store.BasicEventUpdates]()

		core, logs := observer.New(zap.DebugLevel)
		logger := zap.New(core)

		mockClient := versionMocks.NewMockVersionClient(t)
		mockClient.On("Versions").Return(agentVersions, nil)

		mockStore := storeMocks.NewMockStore(t)
		mockStore.On("Updates", mock.Anything).Return(eventSrc).Maybe()
		mockStore.On("ApplyResources", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			resources := args.Get(1).([]model.Resource)
			require.Equal(t, resourcesAsResourceSlice(agentVersions), resources)
		}).Return(resourceAsCreatedStatuses(agentVersions), nil)

		cancelContext, cancel := context.WithCancel(context.Background())
		defer cancel()
		vers := NewVersions(cancelContext, mockClient, mockStore, VersionsSettings{
			Logger:                    logger,
			SyncAgentVersionsInterval: 0,
		})

		actualVersions := vers.(*versions)

		actualVersions.syncAgentVersionsOnce(cancelContext)

		// Make sure we logged the agent version statuses
		filteredLogs := logs.Filter(func(le observer.LoggedEntry) bool {
			return le.Level == zap.DebugLevel && le.Message == "syncAgentVersions"
		})

		require.Equal(t, 1, logs.Len())
		require.Equal(t, statusesAsStrings(resourceAsCreatedStatuses(agentVersions)), filteredLogs.All()[0].ContextMap()["statuses"])
	})

	t.Run("SyncVersions fails", func(t *testing.T) {
		agentVersions := readJSONFile[[]*model.AgentVersion](t, filepath.Join("testfiles", "agent-versions.json"))
		eventSrc := eventbus.NewSource[store.BasicEventUpdates]()

		core, logs := observer.New(zap.DebugLevel)
		logger := zap.New(core)

		mockClient := versionMocks.NewMockVersionClient(t)
		mockClient.On("Versions").Return(agentVersions, nil)

		mockStore := storeMocks.NewMockStore(t)
		applyResourcesErr := errors.New("failed to apply resources")
		mockStore.On("Updates", mock.Anything).Return(eventSrc).Maybe()
		mockStore.On("ApplyResources", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			resources := args.Get(1).([]model.Resource)
			require.Equal(t, resourcesAsResourceSlice(agentVersions), resources)
		}).Return(nil, applyResourcesErr)

		cancelContext, cancel := context.WithCancel(context.Background())
		defer cancel()
		vers := NewVersions(cancelContext, mockClient, mockStore, VersionsSettings{
			Logger:                    logger,
			SyncAgentVersionsInterval: 0,
		})

		actualVersions := vers.(*versions)

		actualVersions.syncAgentVersionsOnce(cancelContext)

		// Make sure we logged the error message
		filteredLogs := logs.Filter(func(le observer.LoggedEntry) bool {
			return le.Level == zap.ErrorLevel && le.Message == "Error during syncAgentVersions ApplyResources"
		})

		require.Equal(t, 1, logs.Len())
		require.Equal(t, applyResourcesErr.Error(), filteredLogs.All()[0].ContextMap()["error"])
	})

	t.Run("ApplyResources fails", func(t *testing.T) {
		eventSrc := eventbus.NewSource[store.BasicEventUpdates]()

		core, logs := observer.New(zap.DebugLevel)
		logger := zap.New(core)

		mockClient := versionMocks.NewMockVersionClient(t)
		versionsErr := errors.New("failed to get versions")
		mockClient.On("Versions").Return(nil, versionsErr)

		mockStore := storeMocks.NewMockStore(t)
		mockStore.On("Updates", mock.Anything).Return(eventSrc).Maybe()

		cancelContext, cancel := context.WithCancel(context.Background())
		defer cancel()
		vers := NewVersions(cancelContext, mockClient, mockStore, VersionsSettings{
			Logger:                    logger,
			SyncAgentVersionsInterval: 0,
		})

		actualVersions := vers.(*versions)

		actualVersions.syncAgentVersionsOnce(cancelContext)

		// Make sure we logged the error message
		filteredLogs := logs.Filter(func(le observer.LoggedEntry) bool {
			return le.Level == zap.ErrorLevel && le.Message == "Error during syncAgentVersions SyncVersions"
		})

		require.Equal(t, 1, logs.Len())
		require.Equal(t, versionsErr.Error(), filteredLogs.All()[0].ContextMap()["error"])
	})
}

func TestWatchAgentVersionUpdates(t *testing.T) {
	eventSrc := eventbus.NewSource[store.BasicEventUpdates]()

	mockClient := versionMocks.NewMockVersionClient(t)

	mockStore := storeMocks.NewMockStore(t)
	mockStore.On("Updates", mock.Anything).Return(eventSrc).Maybe()

	cancelContext, cancel := context.WithCancel(context.Background())
	defer cancel()

	vers := NewVersions(cancelContext, mockClient, mockStore, VersionsSettings{
		Logger:                    zaptest.NewLogger(t),
		SyncAgentVersionsInterval: 0,
	})
	actualVersions := vers.(*versions)

	agentUpdate := model.NewAgentVersion(model.AgentVersionSpec{})
	actualVersions.latestVersion.Update(agentUpdate)

	updates := store.NewEventUpdates()
	updates.IncludeResource(agentUpdate, store.EventTypeUpdate)

	eventSrc.Send(context.Background(), updates)

	require.Eventually(t, func() bool {
		return actualVersions.latestVersion.Get() == nil
	}, 10*time.Minute, 10*time.Millisecond)
}

// Converts a slice of raw values that implement model.Resource
// to a slice of model.Resource
func resourcesAsResourceSlice[T model.Resource](slice []T) []model.Resource {
	var resources []model.Resource
	for _, sliceElem := range slice {
		resources = append(resources, sliceElem)
	}
	return resources
}

// Converts the slice of model.Resource (or an implementor) into ResourceStatus,
// where the status is StatusCreated.
func resourceAsCreatedStatuses[T model.Resource](slice []T) []model.ResourceStatus {
	var statuses []model.ResourceStatus
	for _, sliceElem := range slice {
		statuses = append(statuses, *model.NewResourceStatus(sliceElem, model.StatusCreated))
	}
	return statuses
}

// statusesAsStrings returns []any instead of []string because this is the type
// they are cast as by zap when logged in the observer.
func statusesAsStrings(statuses []model.ResourceStatus) []any {
	var strings []any
	for _, status := range statuses {
		strings = append(strings, status.String())
	}
	return strings
}
