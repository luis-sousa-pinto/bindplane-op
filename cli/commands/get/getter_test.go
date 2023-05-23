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

package get

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/observiq/bindplane-op/cli/printer"
	printermocks "github.com/observiq/bindplane-op/cli/printer/mocks"
	"github.com/observiq/bindplane-op/client"
	clientmocks "github.com/observiq/bindplane-op/client/mocks"
	"github.com/observiq/bindplane-op/model"
	"github.com/observiq/bindplane-op/model/version"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetResource(t *testing.T) {
	testCases := []struct {
		name             string
		clientFunc       func() client.BindPlane
		kind             model.Kind
		id               string
		expectedContents string
		expectedErr      error
	}{
		{
			name: "valid agent",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				agent := &model.Agent{ID: "test-id"}
				c.On("Agent", mock.Anything, "test-id").Return(agent, nil)
				return c
			},
			kind:             model.KindAgent,
			id:               "test-id",
			expectedContents: `Agent=test-id`,
		},
		{
			name: "valid agent version",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				agentVersion := &model.AgentVersion{}
				agentVersion.Metadata.ID = "0.0.0"
				c.On("AgentVersion", mock.Anything, "0.0.0").Return(agentVersion, nil)
				return c
			},
			kind:             model.KindAgentVersion,
			id:               "0.0.0",
			expectedContents: `AgentVersion=0.0.0`,
		},
		{
			name: "valid configuration",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				configuration := &model.Configuration{}
				configuration.Metadata.ID = "test-id"
				c.On("Configuration", mock.Anything, "test-id").Return(configuration, nil)
				return c
			},
			kind:             model.KindConfiguration,
			id:               "test-id",
			expectedContents: "Configuration=test-id",
		},
		{
			name: "valid rollout",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				configuration := &model.Configuration{
					ResourceMeta: model.ResourceMeta{
						Metadata: model.Metadata{
							ID:      "test-id",
							Name:    "test",
							Version: 0,
						},
					},
				}
				c.On("Configuration", mock.Anything, "test-id").Return(configuration, nil)
				return c
			},
			kind:             model.KindRollout,
			id:               "test-id",
			expectedContents: "Rollout=test:0",
		},
		{
			name: "valid destination type",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				destinationType := &model.DestinationType{}
				destinationType.Metadata.ID = "test-id"
				c.On("DestinationType", mock.Anything, "test-id").Return(destinationType, nil)
				return c
			},
			kind:             model.KindDestinationType,
			id:               "test-id",
			expectedContents: "DestinationType=test-id",
		},
		{
			name: "valid destination",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				destination := &model.Destination{}
				destination.Metadata.ID = "test-id"
				c.On("Destination", mock.Anything, "test-id").Return(destination, nil)
				return c
			},
			kind:             model.KindDestination,
			id:               "test-id",
			expectedContents: "Destination=test-id",
		},
		{
			name: "valid processor type",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				processorType := &model.ProcessorType{}
				processorType.Metadata.ID = "test-id"
				c.On("ProcessorType", mock.Anything, "test-id").Return(processorType, nil)
				return c
			},
			kind:             model.KindProcessorType,
			id:               "test-id",
			expectedContents: "ProcessorType=test-id",
		},
		{
			name: "valid processor",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				processor := &model.Processor{}
				processor.Metadata.ID = "test-id"
				c.On("Processor", mock.Anything, "test-id").Return(processor, nil)
				return c
			},
			kind:             model.KindProcessor,
			id:               "test-id",
			expectedContents: "Processor=test-id",
		},
		{
			name: "valid source type",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				sourceType := &model.SourceType{}
				sourceType.Metadata.ID = "test-id"
				c.On("SourceType", mock.Anything, "test-id").Return(sourceType, nil)
				return c
			},
			kind:             model.KindSourceType,
			id:               "test-id",
			expectedContents: "SourceType=test-id",
		},
		{
			name: "valid source",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				source := &model.Source{}
				source.Metadata.ID = "test-id"
				c.On("Source", mock.Anything, "test-id").Return(source, nil)
				return c
			},
			kind:             model.KindSource,
			id:               "test-id",
			expectedContents: "Source=test-id",
		},
		{
			name: "agent failure",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				c.On("Agent", mock.Anything, "test-id").Return(nil, errors.New("not found"))
				return c
			},
			kind:        model.KindAgent,
			id:          "test-id",
			expectedErr: errors.New("not found"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			printer := &testPrinter{}
			getter := NewGetter(tc.clientFunc(), printer)

			err := getter.GetResource(context.Background(), tc.kind, tc.id)
			switch tc.expectedErr {
			case nil:
				require.NoError(t, err)
				require.Equal(t, tc.expectedContents, printer.contents)
			default:
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr.Error())
			}
		})
	}
}

func TestGetResources(t *testing.T) {
	testCases := []struct {
		name             string
		clientFunc       func() client.BindPlane
		kind             model.Kind
		ids              []string
		expectedContents string
		expectedErr      error
	}{
		{
			name: "valid agent",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				agent := &model.Agent{ID: "test-id"}
				c.On("Agent", mock.Anything, "test-id").Return(agent, nil)
				return c
			},
			kind:             model.KindAgent,
			ids:              []string{"test-id"},
			expectedContents: `Agent=test-id`,
		},
		{
			name: "valid agent version",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				agentVersion := &model.AgentVersion{}
				agentVersion.Metadata.ID = "0.0.0"
				c.On("AgentVersion", mock.Anything, "0.0.0").Return(agentVersion, nil)
				return c
			},
			kind:             model.KindAgentVersion,
			ids:              []string{"0.0.0"},
			expectedContents: `AgentVersion=0.0.0`,
		},
		{
			name: "valid configuration",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				configuration := &model.Configuration{}
				configuration.Metadata.ID = "test-id"
				c.On("Configuration", mock.Anything, "test-id").Return(configuration, nil)
				return c
			},
			kind:             model.KindConfiguration,
			ids:              []string{"test-id"},
			expectedContents: "Configuration=test-id",
		},
		{
			name: "valid rollout",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				configuration := &model.Configuration{
					ResourceMeta: model.ResourceMeta{
						Metadata: model.Metadata{
							ID:      "test-id",
							Name:    "test",
							Version: 0,
						},
					},
				}
				c.On("Configuration", mock.Anything, "test-id").Return(configuration, nil)
				return c
			},
			kind:             model.KindRollout,
			ids:              []string{"test-id"},
			expectedContents: "Rollout=test:0",
		},
		{
			name: "valid destination type",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				destinationType := &model.DestinationType{}
				destinationType.Metadata.ID = "test-id"
				c.On("DestinationType", mock.Anything, "test-id").Return(destinationType, nil)
				return c
			},
			kind:             model.KindDestinationType,
			ids:              []string{"test-id"},
			expectedContents: "DestinationType=test-id",
		},
		{
			name: "valid destination",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				destination := &model.Destination{}
				destination.Metadata.ID = "test-id"
				c.On("Destination", mock.Anything, "test-id").Return(destination, nil)
				return c
			},
			kind:             model.KindDestination,
			ids:              []string{"test-id"},
			expectedContents: "Destination=test-id",
		},
		{
			name: "valid processor type",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				processorType := &model.ProcessorType{}
				processorType.Metadata.ID = "test-id"
				c.On("ProcessorType", mock.Anything, "test-id").Return(processorType, nil)
				return c
			},
			kind:             model.KindProcessorType,
			ids:              []string{"test-id"},
			expectedContents: "ProcessorType=test-id",
		},
		{
			name: "valid processor",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				processor := &model.Processor{}
				processor.Metadata.ID = "test-id"
				c.On("Processor", mock.Anything, "test-id").Return(processor, nil)
				return c
			},
			kind:             model.KindProcessor,
			ids:              []string{"test-id"},
			expectedContents: "Processor=test-id",
		},
		{
			name: "valid source type",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				sourceType := &model.SourceType{}
				sourceType.Metadata.ID = "test-id"
				c.On("SourceType", mock.Anything, "test-id").Return(sourceType, nil)
				return c
			},
			kind:             model.KindSourceType,
			ids:              []string{"test-id"},
			expectedContents: "SourceType=test-id",
		},
		{
			name: "valid source",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				source := &model.Source{}
				source.Metadata.ID = "test-id"
				c.On("Source", mock.Anything, "test-id").Return(source, nil)
				return c
			},
			kind:             model.KindSource,
			ids:              []string{"test-id"},
			expectedContents: "Source=test-id",
		},
		{
			name: "agent failure",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				c.On("Agent", mock.Anything, "test-id").Return(nil, errors.New("not found"))
				return c
			},
			kind:        model.KindAgent,
			ids:         []string{"test-id"},
			expectedErr: errors.New("not found"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			printer := &testPrinter{}
			getter := NewGetter(tc.clientFunc(), printer)

			err := getter.GetResources(context.Background(), tc.kind, tc.ids)
			switch tc.expectedErr {
			case nil:
				require.NoError(t, err)
				require.Equal(t, tc.expectedContents, printer.contents)
			default:
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr.Error())
			}
		})
	}
}

func TestGetResourcesOfKind(t *testing.T) {
	testCases := []struct {
		name             string
		clientFunc       func() client.BindPlane
		kind             model.Kind
		expectedContents string
		expectedErr      error
	}{
		{
			name: "valid agents",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				agent := &model.Agent{ID: "test-id"}
				c.On("Agents", mock.Anything, mock.Anything).Return([]*model.Agent{agent}, nil)
				return c
			},
			kind:             model.KindAgent,
			expectedContents: `Agent=test-id`,
		},
		{
			name: "valid agent versions",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				agentVersion := &model.AgentVersion{}
				agentVersion.Metadata.ID = "0.0.0"
				c.On("AgentVersions", mock.Anything).Return([]*model.AgentVersion{agentVersion}, nil)
				return c
			},
			kind:             model.KindAgentVersion,
			expectedContents: `AgentVersion=0.0.0`,
		},
		{
			name: "valid configurations",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				configuration := &model.Configuration{}
				configuration.Metadata.ID = "test-id"
				c.On("Configurations", mock.Anything).Return([]*model.Configuration{configuration}, nil)
				return c
			},
			kind:             model.KindConfiguration,
			expectedContents: "Configuration=test-id",
		},
		{
			name: "valid rollout",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				configuration := &model.Configuration{
					ResourceMeta: model.ResourceMeta{
						Metadata: model.Metadata{
							ID:      "test-id",
							Name:    "test",
							Version: 0,
						},
					},
				}
				c.On("Configurations", mock.Anything).Return([]*model.Configuration{configuration}, nil)
				return c
			},
			kind:             model.KindRollout,
			expectedContents: "Rollout=test:0",
		},
		{
			name: "valid destination types",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				destinationType := &model.DestinationType{}
				destinationType.Metadata.ID = "test-id"
				c.On("DestinationTypes", mock.Anything).Return([]*model.DestinationType{destinationType}, nil)
				return c
			},
			kind:             model.KindDestinationType,
			expectedContents: "DestinationType=test-id",
		},
		{
			name: "valid destinations",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				destination := &model.Destination{}
				destination.Metadata.ID = "test-id"
				c.On("Destinations", mock.Anything).Return([]*model.Destination{destination}, nil)
				return c
			},
			kind:             model.KindDestination,
			expectedContents: "Destination=test-id",
		},
		{
			name: "valid processor types",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				processorType := &model.ProcessorType{}
				processorType.Metadata.ID = "test-id"
				c.On("ProcessorTypes", mock.Anything).Return([]*model.ProcessorType{processorType}, nil)
				return c
			},
			kind:             model.KindProcessorType,
			expectedContents: "ProcessorType=test-id",
		},
		{
			name: "valid processor",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				processor := &model.Processor{}
				processor.Metadata.ID = "test-id"
				c.On("Processors", mock.Anything).Return([]*model.Processor{processor}, nil)
				return c
			},
			kind:             model.KindProcessor,
			expectedContents: "Processor=test-id",
		},
		{
			name: "valid source types",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				sourceType := &model.SourceType{}
				sourceType.Metadata.ID = "test-id"
				c.On("SourceTypes", mock.Anything).Return([]*model.SourceType{sourceType}, nil)
				return c
			},
			kind:             model.KindSourceType,
			expectedContents: "SourceType=test-id",
		},
		{
			name: "valid sources",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				source := &model.Source{}
				source.Metadata.ID = "test-id"
				c.On("Sources", mock.Anything).Return([]*model.Source{source}, nil)
				return c
			},
			kind:             model.KindSource,
			expectedContents: "Source=test-id",
		},
		{
			name: "unknown type",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				return c
			},
			kind:        model.Kind("unknown"),
			expectedErr: errors.New("unknown resource type: unknown"),
		},
		{
			name: "agents failure",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				c.On("Agents", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))
				return c
			},
			kind:        model.KindAgent,
			expectedErr: errors.New("not found"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			printer := &testPrinter{}
			getter := NewGetter(tc.clientFunc(), printer)

			err := getter.GetResourcesOfKind(context.Background(), tc.kind, client.QueryOptions{})
			switch tc.expectedErr {
			case nil:
				require.NoError(t, err)
				require.Equal(t, tc.expectedContents, printer.contents)
			default:
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr.Error())
			}
		})
	}
}

func TestGetAllResources(t *testing.T) {
	testCases := []struct {
		name            string
		clientFunc      func() client.BindPlane
		expectedErr     error
		expectedContent string
	}{
		{
			name: "valid resources",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				agentVersion := &model.AgentVersion{}
				agentVersion.Metadata.ID = "version-id"
				c.On("AgentVersions", mock.Anything).Return([]*model.AgentVersion{agentVersion}, nil)

				configuration := &model.Configuration{}
				configuration.Metadata.ID = "configuration-id"
				c.On("Configurations", mock.Anything).Return([]*model.Configuration{configuration}, nil)

				destinationType := &model.DestinationType{}
				destinationType.Metadata.ID = "destination-type-id"
				c.On("DestinationTypes", mock.Anything).Return([]*model.DestinationType{destinationType}, nil)

				destination := &model.Destination{}
				destination.Metadata.ID = "destination-id"
				c.On("Destinations", mock.Anything).Return([]*model.Destination{destination}, nil)

				processorType := &model.ProcessorType{}
				processorType.Metadata.ID = "processor-type-id"
				c.On("ProcessorTypes", mock.Anything).Return([]*model.ProcessorType{processorType}, nil)

				processor := &model.Processor{}
				processor.Metadata.ID = "processor-id"
				c.On("Processors", mock.Anything).Return([]*model.Processor{processor}, nil)

				sourceType := &model.SourceType{}
				sourceType.Metadata.ID = "source-type-id"
				c.On("SourceTypes", mock.Anything).Return([]*model.SourceType{sourceType}, nil)

				source := &model.Source{}
				source.Metadata.ID = "source-id"
				c.On("Sources", mock.Anything).Return([]*model.Source{source}, nil)

				return c
			},
			expectedContent: "Configuration=configuration-id|Source=source-id|Processor=processor-id|Destination=destination-id|SourceType=source-type-id|ProcessorType=processor-type-id|DestinationType=destination-type-id|AgentVersion=version-id",
		},
		{
			name: "failed resources",
			clientFunc: func() client.BindPlane {
				c := clientmocks.NewMockBindPlane(t)
				c.On("AgentVersions", mock.Anything).Return(nil, errors.New("not found"))
				c.On("Configurations", mock.Anything).Return(nil, errors.New("not found"))
				c.On("DestinationTypes", mock.Anything).Return(nil, errors.New("not found"))
				c.On("Destinations", mock.Anything).Return(nil, errors.New("not found"))
				c.On("ProcessorTypes", mock.Anything).Return(nil, errors.New("not found"))
				c.On("Processors", mock.Anything).Return(nil, errors.New("not found"))
				c.On("SourceTypes", mock.Anything).Return(nil, errors.New("not found"))
				c.On("Sources", mock.Anything).Return(nil, errors.New("not found"))
				return c
			},
			expectedErr: errors.New("8 errors occurred"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			printer := &testPrinter{}
			getter := NewGetter(tc.clientFunc(), printer)

			err := getter.GetAllResources(context.Background())
			switch tc.expectedErr {
			case nil:
				require.NoError(t, err)
				require.Equal(t, tc.expectedContent, printer.contents)
			default:
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr.Error())
			}
		})
	}
}

func TestGetResourceHistory(t *testing.T) {
	testCases := []struct {
		name        string
		mockSetup   func(t *testing.T) (client.BindPlane, printer.Printer)
		kind        model.Kind
		resourceID  string
		expectedErr error
	}{
		{
			name: "Client Error",
			mockSetup: func(t *testing.T) (client.BindPlane, printer.Printer) {
				t.Helper()
				mockClient := clientmocks.NewMockBindPlane(t)
				mockClient.On("ResourceHistory", mock.Anything, model.KindConfiguration, "config1").Return(nil, errors.New("bad"))

				mockPrinter := printermocks.NewMockPrinter(t)

				return mockClient, mockPrinter
			},
			kind:        model.KindConfiguration,
			resourceID:  "config1",
			expectedErr: errors.New("bad"),
		},
		{
			name: "Success",
			mockSetup: func(t *testing.T) (client.BindPlane, printer.Printer) {
				t.Helper()
				resource := &model.AnyResource{
					ResourceMeta: model.ResourceMeta{
						APIVersion: version.V1,
						Kind:       model.KindConfiguration,
						Metadata: model.Metadata{
							Name: "config1",
						},
					},
				}
				mockClient := clientmocks.NewMockBindPlane(t)
				mockClient.On("ResourceHistory", mock.Anything, model.KindConfiguration, "config1").Return([]*model.AnyResource{resource}, nil)

				mockPrinter := printermocks.NewMockPrinter(t)

				printable, err := model.AsKind[model.Resource](resource)
				require.NoError(t, err)
				mockPrinter.On("PrintResources", []model.Printable{printable})

				return mockClient, mockPrinter
			},
			kind:        model.KindConfiguration,
			resourceID:  "config1",
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient, mockPrinter := tc.mockSetup(t)
			getter := NewGetter(mockClient, mockPrinter)
			err := getter.GetResourceHistory(context.Background(), tc.kind, tc.resourceID)
			switch tc.expectedErr {
			case nil:
				require.NoError(t, err)
			default:
				require.ErrorContains(t, err, tc.expectedErr.Error())
			}
		})
	}
}

// testPrinter is a printer used for testing
type testPrinter struct {
	contents string
}

// PrintResource prints a generic model that implements the printable interface
func (t *testPrinter) PrintResource(item model.Printable) {
	res, ok := item.(model.Resource)
	if ok {
		t.contents += fmt.Sprintf("%s=%s", res.GetKind(), res.ID())
		return
	}

	agent, ok := item.(*model.Agent)
	if ok {
		t.contents += fmt.Sprintf("Agent=%s", agent.UniqueKey())
		return
	}

	rollout, ok := item.(*model.Rollout)
	if ok {
		t.contents += fmt.Sprintf("Rollout=%s", rollout.Name)
		return
	}
}

// PrintResources prints a generic model that implements the model.Printable interface
func (t *testPrinter) PrintResources(list []model.Printable) {
	for _, item := range list {
		if len(t.contents) > 0 {
			t.contents += "|"
		}
		t.PrintResource(item)
	}
}
