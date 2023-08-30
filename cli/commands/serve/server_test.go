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

package serve

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/observiq/bindplane-op/config"
	"github.com/observiq/bindplane-op/eventbus"
	metricsmocks "github.com/observiq/bindplane-op/metrics/mocks"
	"github.com/observiq/bindplane-op/model"
	serverMocks "github.com/observiq/bindplane-op/server/mocks"
	"github.com/observiq/bindplane-op/store"
	storeMocks "github.com/observiq/bindplane-op/store/mocks"
	searchMocks "github.com/observiq/bindplane-op/store/search/mocks"
	statsmocks "github.com/observiq/bindplane-op/store/stats/mocks"
	traceMocks "github.com/observiq/bindplane-op/tracer/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestSeed(t *testing.T) {
	testCases := []struct {
		name      string
		storeFunc func() store.Store
		expected  error
	}{
		{
			name: "success",
			storeFunc: func() store.Store {
				testConfig := model.NewConfiguration("test")
				configIndex := searchMocks.NewMockIndex(t)
				configIndex.On("Upsert", testConfig).Return(nil)

				testAgent := &model.Agent{ID: "test"}
				agentIndex := searchMocks.NewMockIndex(t)
				agentIndex.On("Upsert", testAgent).Return(nil)

				s := storeMocks.NewMockStore(t)
				s.On("ApplyResources", mock.Anything, mock.Anything).Return(nil, nil)
				s.On("Configurations", mock.Anything).Return([]*model.Configuration{testConfig}, nil)
				s.On("ConfigurationIndex", mock.Anything).Return(configIndex)
				s.On("Agents", mock.Anything).Return([]*model.Agent{testAgent}, nil)
				s.On("AgentIndex", mock.Anything).Return(agentIndex)
				s.EXPECT().ProcessorType(mock.Anything, mock.Anything).Return(nil, nil)
				return s
			},
			expected: nil,
		},
		{
			name: "error applying resources does not fail startup",
			storeFunc: func() store.Store {
				testConfig := model.NewConfiguration("test")
				configIndex := searchMocks.NewMockIndex(t)
				configIndex.On("Upsert", testConfig).Return(nil)

				testAgent := &model.Agent{ID: "test"}
				agentIndex := searchMocks.NewMockIndex(t)
				agentIndex.On("Upsert", testAgent).Return(nil)

				s := storeMocks.NewMockStore(t)
				s.On("ApplyResources", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error"))
				s.On("Configurations", mock.Anything).Return([]*model.Configuration{testConfig}, nil)
				s.On("ConfigurationIndex", mock.Anything).Return(configIndex)
				s.On("Agents", mock.Anything).Return([]*model.Agent{testAgent}, nil)
				s.On("AgentIndex", mock.Anything).Return(agentIndex)
				return s
			},
			expected: nil,
		},
		{
			name: "error getting configurations",
			storeFunc: func() store.Store {
				s := storeMocks.NewMockStore(t)
				s.On("ApplyResources", mock.Anything, mock.Anything).Return(nil, nil)
				s.On("Configurations", mock.Anything).Return(nil, fmt.Errorf("error"))
				s.EXPECT().ProcessorType(mock.Anything, mock.Anything).Return(nil, nil)
				return s
			},
			expected: errors.New("failed to get configurations"),
		},
		{
			name: "error seeding configuration index",
			storeFunc: func() store.Store {
				testConfig := model.NewConfiguration("test")
				configIndex := searchMocks.NewMockIndex(t)
				configIndex.On("Upsert", testConfig).Return(fmt.Errorf("error"))

				s := storeMocks.NewMockStore(t)
				s.On("ApplyResources", mock.Anything, mock.Anything).Return(nil, nil)
				s.On("Configurations", mock.Anything).Return([]*model.Configuration{testConfig}, nil)
				s.On("ConfigurationIndex", mock.Anything).Return(configIndex)
				s.EXPECT().ProcessorType(mock.Anything, mock.Anything).Return(nil, nil)
				return s
			},
			expected: errors.New("failed to seed configuration index"),
		},
		{
			name: "error getting agents",
			storeFunc: func() store.Store {
				testConfig := model.NewConfiguration("test")
				configIndex := searchMocks.NewMockIndex(t)
				configIndex.On("Upsert", testConfig).Return(nil)

				s := storeMocks.NewMockStore(t)
				s.On("ApplyResources", mock.Anything, mock.Anything).Return(nil, nil)
				s.On("Configurations", mock.Anything).Return([]*model.Configuration{testConfig}, nil)
				s.On("ConfigurationIndex", mock.Anything).Return(configIndex)
				s.On("Agents", mock.Anything).Return(nil, fmt.Errorf("error"))
				s.EXPECT().ProcessorType(mock.Anything, mock.Anything).Return(nil, nil)
				return s
			},
			expected: errors.New("failed to get agents"),
		},
		{
			name: "error seeding agent index",
			storeFunc: func() store.Store {
				testConfig := model.NewConfiguration("test")
				configIndex := searchMocks.NewMockIndex(t)
				configIndex.On("Upsert", testConfig).Return(nil)

				testAgent := &model.Agent{ID: "test"}
				agentIndex := searchMocks.NewMockIndex(t)
				agentIndex.On("Upsert", testAgent).Return(fmt.Errorf("error"))

				s := storeMocks.NewMockStore(t)
				s.On("ApplyResources", mock.Anything, mock.Anything).Return(nil, nil)
				s.On("Configurations", mock.Anything).Return([]*model.Configuration{testConfig}, nil)
				s.On("ConfigurationIndex", mock.Anything).Return(configIndex)
				s.On("Agents", mock.Anything).Return([]*model.Agent{testAgent}, nil)
				s.On("AgentIndex", mock.Anything).Return(agentIndex)
				s.EXPECT().ProcessorType(mock.Anything, mock.Anything).Return(nil, nil)
				return s
			},
			expected: errors.New("failed to seed agent index"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewServer(nil, tc.storeFunc(), nil, zap.NewNop(), nil, nil)
			err := s.Seed(context.Background())
			switch tc.expected {
			case nil:
				require.NoError(t, err)
			default:
				require.Contains(t, err.Error(), tc.expected.Error())
			}
		})
	}
}

func TestServeWithClient(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{
		Network: config.Network{
			Host: "localhost",
			Port: "5555",
		},
		RolloutsInterval: config.DefaultRolloutsInterval,
		Auth: config.Auth{
			Username: "username",
			Password: "password",
		},
	}

	st := storeMocks.NewMockStore(t)
	st.On("Updates", mock.Anything).Return(eventbus.NewSource[store.BasicEventUpdates]())

	mockMeasurements := statsmocks.NewMockMeasurements(t)
	st.On("Measurements").Return(mockMeasurements)

	tracer := traceMocks.NewMockTracer(t)
	tracer.On("Start", mock.Anything).Return(nil)
	tracer.On("Shutdown", mock.Anything).Return(nil)

	mp := metricsmocks.NewMockProvider(t)
	mp.On("Start", mock.Anything).Return(nil)
	mp.On("Shutdown", mock.Anything).Return(nil)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	mockRouteBuilder := serverMocks.NewMockRouteBuilder(t)
	mockRouteBuilder.On("AddRoutes", mock.Anything, mock.Anything).Return(nil)

	errChan := make(chan error, 1)
	go func() {
		errChan <- NewServer(cfg, st, tracer, logger, mp, mockRouteBuilder).Serve(ctx)
	}()

	var resultStatus int
	resultsWithClient := func() bool {
		client := resty.New()
		client.SetBaseURL(cfg.Network.ServerURL())

		resp, err := client.R().Get("/health")
		resultStatus = resp.StatusCode()
		t.Log(err)
		return err == nil
	}

	require.Eventually(t, resultsWithClient, time.Second*5, time.Millisecond*100)
	require.Equal(t, http.StatusOK, resultStatus)
	cancel()

	err := <-errChan
	require.NoError(t, err)
}
