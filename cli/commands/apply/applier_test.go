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

package apply

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

func TestApplyResourcesFromFiles(t *testing.T) {
	testCases := []struct {
		name             string
		clientFunc       func() client.BindPlane
		setupFunc        func(path string) error
		expectedStatuses []*model.AnyResourceStatus
		expectedErr      error
	}{
		{
			name: "missing file",
			clientFunc: func() client.BindPlane {
				c := mocks.NewMockBindPlane(t)
				return c
			},
			setupFunc:   func(path string) error { return nil },
			expectedErr: errors.New("failed to read resources"),
		},
		{
			name: "empty file",
			clientFunc: func() client.BindPlane {
				c := mocks.NewMockBindPlane(t)
				c.On("Apply", mock.Anything, mock.Anything).Return(nil, nil)
				return c
			},
			setupFunc: func(path string) error {
				return os.WriteFile(path, []byte(""), 0644)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			applier := NewApplier(tc.clientFunc())
			filename := filepath.Join(t.TempDir(), tc.name)
			err := tc.setupFunc(filename)
			require.NoError(t, err)

			statuses, err := applier.ApplyResourcesFromFiles(context.Background(), []string{filename})
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

func TestApplyResourcesFromReader(t *testing.T) {
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
				c.On("Apply", mock.Anything, mock.Anything).Return(nil, nil)
				return c
			},
			readerFunc: func() io.Reader {
				return strings.NewReader("")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			applier := NewApplier(tc.clientFunc())
			statuses, err := applier.ApplyResourcesFromReader(context.Background(), tc.readerFunc())
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
