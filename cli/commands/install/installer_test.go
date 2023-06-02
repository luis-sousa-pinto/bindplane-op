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

package install

import (
	"context"
	"testing"

	"github.com/observiq/bindplane-op/client"
	"github.com/observiq/bindplane-op/client/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetAgentInstallCommand(t *testing.T) {
	c := mocks.NewMockBindPlane(t)
	c.On("AgentInstallCommand", mock.Anything, mock.Anything).Return("test", nil)
	i := NewInstaller(c)
	result, err := i.GetAgentInstallCommand(context.Background(), client.AgentInstallOptions{})
	require.NoError(t, err)
	require.Equal(t, "test", result)
}
