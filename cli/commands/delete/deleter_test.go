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

package delete

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/observiq/bindplane-op/client"
	"github.com/observiq/bindplane-op/client/mocks"
	"github.com/observiq/bindplane-op/model"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDeleteResources(t *testing.T) {
	testCases := []struct {
		name        string
		clientFunc  func() client.BindPlane
		kind        model.Kind
		ids         []string
		expectedErr error
	}{
		{
			name: "delete configuration",
			clientFunc: func() client.BindPlane {
				c := mocks.NewMockBindPlane(t)
				c.On("DeleteConfiguration", mock.Anything, "123").Return(nil)
				return c
			},
			kind: model.KindConfiguration,
			ids:  []string{"123"},
		},
		{
			name: "delete source",
			clientFunc: func() client.BindPlane {
				c := mocks.NewMockBindPlane(t)
				c.On("DeleteSource", mock.Anything, "123").Return(nil)
				return c
			},
			kind: model.KindSource,
			ids:  []string{"123"},
		},
		{
			name: "delete processor",
			clientFunc: func() client.BindPlane {
				c := mocks.NewMockBindPlane(t)
				c.On("DeleteProcessor", mock.Anything, "123").Return(nil)
				return c
			},
			kind: model.KindProcessor,
			ids:  []string{"123"},
		},
		{
			name: "delete destination",
			clientFunc: func() client.BindPlane {
				c := mocks.NewMockBindPlane(t)
				c.On("DeleteDestination", mock.Anything, "123").Return(nil)
				return c
			},
			kind: model.KindDestination,
			ids:  []string{"123"},
		},
		{
			name: "delete source type",
			clientFunc: func() client.BindPlane {
				c := mocks.NewMockBindPlane(t)
				c.On("DeleteSourceType", mock.Anything, "123").Return(nil)
				return c
			},
			kind: model.KindSourceType,
			ids:  []string{"123"},
		},
		{
			name: "delete processor type",
			clientFunc: func() client.BindPlane {
				c := mocks.NewMockBindPlane(t)
				c.On("DeleteProcessorType", mock.Anything, "123").Return(nil)
				return c
			},
			kind: model.KindProcessorType,
			ids:  []string{"123"},
		},
		{
			name: "delete destination type",
			clientFunc: func() client.BindPlane {
				c := mocks.NewMockBindPlane(t)
				c.On("DeleteDestinationType", mock.Anything, "123").Return(nil)
				return c
			},
			kind: model.KindDestinationType,
			ids:  []string{"123"},
		},
		{
			name: "delete agent version",
			clientFunc: func() client.BindPlane {
				c := mocks.NewMockBindPlane(t)
				c.On("DeleteAgentVersion", mock.Anything, "123").Return(nil)
				return c
			},
			kind: model.KindAgentVersion,
			ids:  []string{"123"},
		},
		{
			name: "delete agents",
			clientFunc: func() client.BindPlane {
				c := mocks.NewMockBindPlane(t)
				c.On("DeleteAgents", mock.Anything, []string{"123"}).Return(nil, nil)
				return c
			},
			kind: model.KindAgent,
			ids:  []string{"123"},
		},
		{
			name: "delete unknown kind",
			clientFunc: func() client.BindPlane {
				c := mocks.NewMockBindPlane(t)
				return c
			},
			kind:        model.Kind("unknown"),
			ids:         []string{"123"},
			expectedErr: errors.New("unsupported resource kind: unknown"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := NewDeleter(tc.clientFunc())
			err := d.DeleteResources(context.Background(), tc.kind, tc.ids)
			switch tc.expectedErr {
			case nil:
				require.NoError(t, err)
			default:
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr.Error())
			}
		})
	}
}

func TestDeleteSomeResources(t *testing.T) {
	testCases := []struct {
		name        string
		clientFunc  func() client.BindPlane
		kind        model.Kind
		ids         []string
		expectedErr error
	}{
		{
			name: "delete agents",
			clientFunc: func() client.BindPlane {
				c := mocks.NewMockBindPlane(t)
				c.On("DeleteAgents", mock.Anything, []string{"test-id"}).Return(nil, nil)
				return c
			},
			kind: model.KindAgent,
			ids:  []string{"test-id"},
		},
		{
			name: "delete unknown kind",
			clientFunc: func() client.BindPlane {
				c := mocks.NewMockBindPlane(t)
				return c
			},
			kind:        model.Kind("unknown"),
			ids:         []string{"test-id"},
			expectedErr: errors.New("unsupported resource kind: unknown"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := NewDeleter(tc.clientFunc())
			err := d.DeleteResources(context.Background(), tc.kind, tc.ids)
			switch tc.expectedErr {
			case nil:
				require.NoError(t, err)
			default:
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr.Error())
			}
		})
	}
}

func TestDeleteResourcesFromFile(t *testing.T) {
	testCases := []struct {
		name             string
		clientFunc       func() client.BindPlane
		setupFunc        func(path string)
		expectedStatuses []*model.AnyResourceStatus
		expectedErr      error
	}{
		{
			name: "missing file",
			clientFunc: func() client.BindPlane {
				c := mocks.NewMockBindPlane(t)
				return c
			},
			setupFunc:   func(path string) {},
			expectedErr: errors.New("failed to read resources"),
		},
		{
			name: "empty file",
			clientFunc: func() client.BindPlane {
				c := mocks.NewMockBindPlane(t)
				c.On("Delete", mock.Anything, mock.Anything).Return(nil, nil)
				return c
			},
			setupFunc: func(path string) {
				os.WriteFile(path, []byte(""), 0644)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := NewDeleter(tc.clientFunc())
			filepath := filepath.Join(t.TempDir(), tc.name)
			tc.setupFunc(filepath)

			statuses, err := d.DeleteResourcesFromFiles(context.Background(), []string{filepath})
			switch tc.expectedErr {
			case nil:
				require.NoError(t, err)
				require.Equal(t, tc.expectedStatuses, statuses)
			default:
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr.Error())
			}
		})
	}
}

func TestDeleteResourcesFromReader(t *testing.T) {
	testCases := []struct {
		name             string
		clientFunc       func() client.BindPlane
		readerFunc       func() io.Reader
		expectedStatuses []*model.AnyResourceStatus
		expectedErr      error
	}{
		{
			name: "failed reader",
			clientFunc: func() client.BindPlane {
				c := mocks.NewMockBindPlane(t)
				return c
			},
			readerFunc: func() io.Reader {
				return strings.NewReader("invalid")
			},
			expectedErr: errors.New("failed to read resources"),
		},
		{
			name: "empty reader",
			clientFunc: func() client.BindPlane {
				c := mocks.NewMockBindPlane(t)
				c.On("Delete", mock.Anything, mock.Anything).Return(nil, nil)
				return c
			},
			readerFunc: func() io.Reader {
				return strings.NewReader("")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := NewDeleter(tc.clientFunc())
			statuses, err := d.DeleteResourcesFromReader(context.Background(), tc.readerFunc())
			switch tc.expectedErr {
			case nil:
				require.NoError(t, err)
				require.Equal(t, tc.expectedStatuses, statuses)
			default:
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr.Error())
			}
		})
	}
}
