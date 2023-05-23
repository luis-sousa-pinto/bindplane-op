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

package rollout

import (
	"context"
	"errors"
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

func TestUpdateRollout(t *testing.T) {
	rolloutName := "my_rollout"
	testCases := []struct {
		name        string
		mockFunc    func(t *testing.T) (client.BindPlane, printer.Printer)
		expectedErr error
	}{
		{
			name: "Client Error",
			mockFunc: func(t *testing.T) (client.BindPlane, printer.Printer) {
				t.Helper()
				mockClient := clientmocks.NewMockBindPlane(t)
				mockClient.On("UpdateRollout", mock.Anything, rolloutName).Return(nil, errors.New("bad"))

				mockPrinter := printermocks.NewMockPrinter(t)

				return mockClient, mockPrinter
			},
			expectedErr: errors.New("bad"),
		},
		{
			name: "Success",
			mockFunc: func(t *testing.T) (client.BindPlane, printer.Printer) {
				t.Helper()
				cfg := &model.Configuration{
					ResourceMeta: model.ResourceMeta{
						APIVersion: version.V1,
						Kind:       model.KindConfiguration,
						Metadata: model.Metadata{
							Name: rolloutName,
						},
					},
				}

				mockClient := clientmocks.NewMockBindPlane(t)
				mockClient.On("UpdateRollout", mock.Anything, rolloutName).Return(cfg, nil)

				mockPrinter := printermocks.NewMockPrinter(t)
				mockPrinter.On("PrintResource", cfg.Rollout()).Return()

				return mockClient, mockPrinter
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient, mockPrinter := tc.mockFunc(t)
			rollouter := NewRollouter(mockClient, mockPrinter)
			err := rollouter.UpdateRollout(context.Background(), rolloutName)
			if tc.expectedErr != nil {
				require.ErrorContains(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUpdateRollouts(t *testing.T) {
	testCases := []struct {
		name        string
		mockFunc    func(t *testing.T) (client.BindPlane, printer.Printer)
		expectedErr error
	}{
		{
			name: "Client Error",
			mockFunc: func(t *testing.T) (client.BindPlane, printer.Printer) {
				t.Helper()
				mockClient := clientmocks.NewMockBindPlane(t)
				mockClient.On("UpdateRollouts", mock.Anything).Return(nil, errors.New("bad"))

				mockPrinter := printermocks.NewMockPrinter(t)

				return mockClient, mockPrinter
			},
			expectedErr: errors.New("bad"),
		},
		{
			name: "Success",
			mockFunc: func(t *testing.T) (client.BindPlane, printer.Printer) {
				t.Helper()
				cfg := &model.Configuration{
					ResourceMeta: model.ResourceMeta{
						APIVersion: version.V1,
						Kind:       model.KindConfiguration,
						Metadata: model.Metadata{
							Name: "rollout1",
						},
					},
				}

				mockClient := clientmocks.NewMockBindPlane(t)
				mockClient.On("UpdateRollouts", mock.Anything).Return([]*model.Configuration{cfg}, nil)

				mockPrinter := printermocks.NewMockPrinter(t)
				printables := []model.Printable{cfg.Rollout()}
				mockPrinter.On("PrintResources", printables).Return()

				return mockClient, mockPrinter
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient, mockPrinter := tc.mockFunc(t)
			rollouter := NewRollouter(mockClient, mockPrinter)
			err := rollouter.UpdateRollouts(context.Background())
			if tc.expectedErr != nil {
				require.ErrorContains(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestStartRollout(t *testing.T) {
	rolloutName := "my_rollout"
	testCases := []struct {
		name        string
		mockFunc    func(t *testing.T) (client.BindPlane, printer.Printer)
		expectedErr error
	}{
		{
			name: "Client Error",
			mockFunc: func(t *testing.T) (client.BindPlane, printer.Printer) {
				t.Helper()
				mockClient := clientmocks.NewMockBindPlane(t)
				mockClient.On("StartRollout", mock.Anything, rolloutName, mock.Anything).Return(nil, errors.New("bad"))

				mockPrinter := printermocks.NewMockPrinter(t)

				return mockClient, mockPrinter
			},
			expectedErr: errors.New("bad"),
		},
		{
			name: "Success",
			mockFunc: func(t *testing.T) (client.BindPlane, printer.Printer) {
				t.Helper()
				cfg := &model.Configuration{
					ResourceMeta: model.ResourceMeta{
						APIVersion: version.V1,
						Kind:       model.KindConfiguration,
						Metadata: model.Metadata{
							Name: rolloutName,
						},
					},
				}

				mockClient := clientmocks.NewMockBindPlane(t)
				mockClient.On("StartRollout", mock.Anything, rolloutName, mock.Anything).Return(cfg, nil)

				mockPrinter := printermocks.NewMockPrinter(t)
				mockPrinter.On("PrintResource", cfg.Rollout()).Return()

				return mockClient, mockPrinter
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient, mockPrinter := tc.mockFunc(t)
			rollouter := NewRollouter(mockClient, mockPrinter)
			err := rollouter.StartRollout(context.Background(), rolloutName)
			if tc.expectedErr != nil {
				require.ErrorContains(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestStartAllRollout(t *testing.T) {
	testCases := []struct {
		name        string
		mockFunc    func(t *testing.T) (client.BindPlane, printer.Printer)
		expectedErr error
	}{
		{
			name: "Client Error",
			mockFunc: func(t *testing.T) (client.BindPlane, printer.Printer) {
				t.Helper()
				mockClient := clientmocks.NewMockBindPlane(t)
				mockClient.On("Configurations", mock.Anything).Return(nil, errors.New("bad"))

				mockPrinter := printermocks.NewMockPrinter(t)

				return mockClient, mockPrinter
			},
			expectedErr: errors.New("bad"),
		},
		{
			name: "Partial Failure",
			mockFunc: func(t *testing.T) (client.BindPlane, printer.Printer) {
				t.Helper()
				cfg1 := &model.Configuration{
					ResourceMeta: model.ResourceMeta{
						APIVersion: version.V1,
						Kind:       model.KindConfiguration,
						Metadata: model.Metadata{
							Name: "cfg1",
						},
					},
				}
				cfg2 := &model.Configuration{
					ResourceMeta: model.ResourceMeta{
						APIVersion: version.V1,
						Kind:       model.KindConfiguration,
						Metadata: model.Metadata{
							Name: "cfg2",
						},
					},
				}
				mockClient := clientmocks.NewMockBindPlane(t)
				mockClient.On("Configurations", mock.Anything).Return([]*model.Configuration{cfg1, cfg2}, nil)

				mockClient.On("StartRollout", mock.Anything, cfg1.Name(), mock.Anything).Return(cfg1, nil)
				mockClient.On("StartRollout", mock.Anything, cfg2.Name(), mock.Anything).Return(nil, errors.New("bad"))

				mockPrinter := printermocks.NewMockPrinter(t)
				printables := []model.Printable{cfg1.Rollout()}
				mockPrinter.On("PrintResources", printables).Return()

				return mockClient, mockPrinter
			},
			expectedErr: errors.New("bad"),
		},
		{
			name: "Success",
			mockFunc: func(t *testing.T) (client.BindPlane, printer.Printer) {
				t.Helper()
				cfg1 := &model.Configuration{
					ResourceMeta: model.ResourceMeta{
						APIVersion: version.V1,
						Kind:       model.KindConfiguration,
						Metadata: model.Metadata{
							Name: "cfg1",
						},
					},
				}
				cfg2 := &model.Configuration{
					ResourceMeta: model.ResourceMeta{
						APIVersion: version.V1,
						Kind:       model.KindConfiguration,
						Metadata: model.Metadata{
							Name: "cfg2",
						},
					},
				}
				mockClient := clientmocks.NewMockBindPlane(t)
				mockClient.On("Configurations", mock.Anything).Return([]*model.Configuration{cfg1, cfg2}, nil)

				mockClient.On("StartRollout", mock.Anything, cfg1.Name(), mock.Anything).Return(cfg1, nil)
				mockClient.On("StartRollout", mock.Anything, cfg2.Name(), mock.Anything).Return(cfg2, nil)

				mockPrinter := printermocks.NewMockPrinter(t)
				printables := []model.Printable{cfg1.Rollout(), cfg2.Rollout()}
				mockPrinter.On("PrintResources", printables).Return()

				return mockClient, mockPrinter
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient, mockPrinter := tc.mockFunc(t)
			rollouter := NewRollouter(mockClient, mockPrinter)
			err := rollouter.StartAllRollouts(context.Background())
			if tc.expectedErr != nil {
				require.ErrorContains(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPauseRollout(t *testing.T) {
	rolloutName := "my_rollout"
	testCases := []struct {
		name        string
		mockFunc    func(t *testing.T) (client.BindPlane, printer.Printer)
		expectedErr error
	}{
		{
			name: "Client Error",
			mockFunc: func(t *testing.T) (client.BindPlane, printer.Printer) {
				t.Helper()
				mockClient := clientmocks.NewMockBindPlane(t)
				mockClient.On("PauseRollout", mock.Anything, rolloutName).Return(nil, errors.New("bad"))

				mockPrinter := printermocks.NewMockPrinter(t)

				return mockClient, mockPrinter
			},
			expectedErr: errors.New("bad"),
		},
		{
			name: "Success",
			mockFunc: func(t *testing.T) (client.BindPlane, printer.Printer) {
				t.Helper()
				cfg := &model.Configuration{
					ResourceMeta: model.ResourceMeta{
						APIVersion: version.V1,
						Kind:       model.KindConfiguration,
						Metadata: model.Metadata{
							Name: rolloutName,
						},
					},
				}

				mockClient := clientmocks.NewMockBindPlane(t)
				mockClient.On("PauseRollout", mock.Anything, rolloutName).Return(cfg, nil)

				mockPrinter := printermocks.NewMockPrinter(t)
				mockPrinter.On("PrintResource", cfg.Rollout()).Return()

				return mockClient, mockPrinter
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient, mockPrinter := tc.mockFunc(t)
			rollouter := NewRollouter(mockClient, mockPrinter)
			err := rollouter.PauseRollout(context.Background(), rolloutName)
			if tc.expectedErr != nil {
				require.ErrorContains(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestResumeRollout(t *testing.T) {
	rolloutName := "my_rollout"
	testCases := []struct {
		name        string
		mockFunc    func(t *testing.T) (client.BindPlane, printer.Printer)
		expectedErr error
	}{
		{
			name: "Client Error",
			mockFunc: func(t *testing.T) (client.BindPlane, printer.Printer) {
				t.Helper()
				mockClient := clientmocks.NewMockBindPlane(t)
				mockClient.On("ResumeRollout", mock.Anything, rolloutName).Return(nil, errors.New("bad"))

				mockPrinter := printermocks.NewMockPrinter(t)

				return mockClient, mockPrinter
			},
			expectedErr: errors.New("bad"),
		},
		{
			name: "Success",
			mockFunc: func(t *testing.T) (client.BindPlane, printer.Printer) {
				t.Helper()
				cfg := &model.Configuration{
					ResourceMeta: model.ResourceMeta{
						APIVersion: version.V1,
						Kind:       model.KindConfiguration,
						Metadata: model.Metadata{
							Name: rolloutName,
						},
					},
				}

				mockClient := clientmocks.NewMockBindPlane(t)
				mockClient.On("ResumeRollout", mock.Anything, rolloutName).Return(cfg, nil)

				mockPrinter := printermocks.NewMockPrinter(t)
				mockPrinter.On("PrintResource", cfg.Rollout()).Return()

				return mockClient, mockPrinter
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient, mockPrinter := tc.mockFunc(t)
			rollouter := NewRollouter(mockClient, mockPrinter)
			err := rollouter.ResumeRollout(context.Background(), rolloutName)
			if tc.expectedErr != nil {
				require.ErrorContains(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRolloutStatus(t *testing.T) {
	rolloutName := "my_rollout"
	testCases := []struct {
		name        string
		mockFunc    func(t *testing.T) (client.BindPlane, printer.Printer)
		expectedErr error
	}{
		{
			name: "Client Error",
			mockFunc: func(t *testing.T) (client.BindPlane, printer.Printer) {
				t.Helper()
				mockClient := clientmocks.NewMockBindPlane(t)
				mockClient.On("Configuration", mock.Anything, rolloutName).Return(nil, errors.New("bad"))

				mockPrinter := printermocks.NewMockPrinter(t)

				return mockClient, mockPrinter
			},
			expectedErr: errors.New("bad"),
		},
		{
			name: "Success",
			mockFunc: func(t *testing.T) (client.BindPlane, printer.Printer) {
				t.Helper()
				cfg := &model.Configuration{
					ResourceMeta: model.ResourceMeta{
						APIVersion: version.V1,
						Kind:       model.KindConfiguration,
						Metadata: model.Metadata{
							Name: rolloutName,
						},
					},
				}

				mockClient := clientmocks.NewMockBindPlane(t)
				mockClient.On("Configuration", mock.Anything, rolloutName).Return(cfg, nil)

				mockPrinter := printermocks.NewMockPrinter(t)
				mockPrinter.On("PrintResource", cfg.Rollout()).Return()

				return mockClient, mockPrinter
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient, mockPrinter := tc.mockFunc(t)
			rollouter := NewRollouter(mockClient, mockPrinter)
			err := rollouter.RolloutStatus(context.Background(), rolloutName)
			if tc.expectedErr != nil {
				require.ErrorContains(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
