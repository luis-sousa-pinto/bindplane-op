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

package serve

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestServeCommand(t *testing.T) {
	errBuild := errors.New("new build error")
	errSeed := errors.New("seed error")

	tests := []struct {
		name      string
		builder   func(t *testing.T) *MockBuilder
		server    func(t *testing.T) *MockServer
		arg       string
		expectErr error
	}{
		{
			name: "exits on error when SupportsServer is false",
			builder: func(t *testing.T) *MockBuilder {
				builder := NewMockBuilder(t)
				builder.EXPECT().SupportsServer().Return(false)
				return builder
			},
			server:    nil,
			arg:       "",
			expectErr: ErrUnsupportedPlatform,
		},
		{
			name: "exits on error when BuildServer returns an error",
			builder: func(t *testing.T) *MockBuilder {
				builder := NewMockBuilder(t)
				builder.EXPECT().SupportsServer().Return(true)
				builder.EXPECT().BuildServer(mock.Anything).Return(nil, errBuild)
				return builder
			},
			server:    nil,
			arg:       "",
			expectErr: errBuild,
		},
		{
			name: "skips seed call with --skip-seed",
			builder: func(t *testing.T) *MockBuilder {
				builder := NewMockBuilder(t)
				builder.EXPECT().SupportsServer().Return(true)

				return builder
			},
			server: func(t *testing.T) *MockServer {
				server := NewMockServer(t)
				server.On("Serve", mock.Anything).Return(nil)
				return server
			},
			arg:       "--skip-seed",
			expectErr: nil,
		},
		{
			name: "exits when Seed errors",
			builder: func(t *testing.T) *MockBuilder {
				builder := NewMockBuilder(t)
				builder.EXPECT().SupportsServer().Return(true)

				return builder
			},
			server: func(t *testing.T) *MockServer {
				server := NewMockServer(t)
				server.On("Seed", mock.Anything).Return(errSeed)
				return server
			},
			arg:       "",
			expectErr: errSeed,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			builder := test.builder(t)
			if test.server != nil {
				builder.On("BuildServer", mock.Anything).Return(test.server(t), nil)
			}

			cmd := Command(builder)

			if test.arg != "" {
				cmd.SetArgs([]string{test.arg})
			}

			err := cmd.Execute()

			if test.expectErr != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, test.expectErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
