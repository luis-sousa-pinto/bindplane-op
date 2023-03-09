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
	"bytes"
	"errors"
	"io/ioutil"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	clientMocks "github.com/observiq/bindplane-op/client/mocks"
	"github.com/observiq/bindplane-op/common"
	"github.com/observiq/bindplane-op/internal/cli"
	"github.com/observiq/bindplane-op/model"
)

var tableOutput = "table"
var jsonOutput = "json"
var yamlOutput = "yaml"

func setupBindPlane(buffer *bytes.Buffer) *cli.BindPlane {
	bindplane := cli.NewBindPlane(common.InitConfig(""), buffer)
	agent1 := &model.Agent{
		ID:              "1",
		Architecture:    "amd64",
		HostName:        "local",
		Platform:        "linux",
		SecretKey:       "secret",
		Version:         "1.0.0",
		Name:            "Agent 1",
		Home:            "/stanza",
		OperatingSystem: "Ubuntu 20.10",
		MacAddress:      "00:00:ac:00:00:00",
		Type:            "stanza",
		Status:          model.Connected,
	}
	agent2 := &model.Agent{
		ID:      "2",
		Name:    "Agent 2",
		Version: "1.0.0",
		Status:  model.Disconnected,
	}
	agents := []*model.Agent{agent1, agent2}

	client := &clientMocks.MockBindPlane{}
	client.On("Agents", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(agents, nil)
	client.On("Agent", mock.Anything, "1").Return(agent1, nil)
	client.On("Agent", mock.Anything, "2").Return(agent2, nil)
	client.On("Agent", mock.Anything, "3").Return(nil, errors.New("unable to get agents, got 404 Not Found"))
	client.On("Agent", mock.Anything, "badId").Return(nil, errors.New("unable to get agents, got 404 Not Found"))

	bindplane.SetClient(client)
	return bindplane
}

func executeAndAssertOutput(t *testing.T, cmd *cobra.Command, buffer *bytes.Buffer, expected string) {
	executeErr := cmd.Execute()
	require.NoError(t, executeErr, "error while executing command")

	out, readErr := ioutil.ReadAll(buffer)
	require.NoError(t, readErr, "error while reading byte array")

	require.Equal(t, expected, string(out))
}
