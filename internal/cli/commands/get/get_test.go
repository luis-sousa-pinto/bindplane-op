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
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/observiq/bindplane-op/common"
	"github.com/observiq/bindplane-op/internal/cli"
	"github.com/observiq/bindplane-op/internal/cli/commands"
	"github.com/observiq/bindplane-op/internal/cli/commands/profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testArgs struct {
	pluralCmd      string
	singularCmd    string
	expectedOutput string
}

func TestGetCommand(t *testing.T) {
	tests := []testArgs{
		{
			pluralCmd:      "agents",
			singularCmd:    "agent",
			expectedOutput: "ID\tNAME   \tVERSION\tSTATUS      \tCONNECTED\tDISCONNECTED\tLABELS \n1 \tAgent 1\t1.0.0  \tConnected   \t-        \t-           \t      \t\n2 \tAgent 2\t1.0.0  \tDisconnected\t-        \t-           \t      \t\n",
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("supports both singular and plural commands for %s", test.pluralCmd), func(t *testing.T) {
			buffer := bytes.NewBufferString("")
			bindplane := setupBindPlane(buffer)
			bindplane.Config.Output = tableOutput

			cmd := Command(bindplane)
			cmd.SetOut(buffer)

			cmd.SetArgs([]string{test.pluralCmd})
			executeAndAssertOutput(t, cmd, buffer, test.expectedOutput)
			buffer.Reset()

			cmd.SetArgs([]string{test.singularCmd})
			executeAndAssertOutput(t, cmd, buffer, test.expectedOutput)
		})
	}
}

func TestGetIndividualCommand(t *testing.T) {
	var tests = []struct {
		description   string
		args          []string
		expectOutput  string
		expectedError error
	}{
		{
			description:   "get agent 1",
			args:          []string{"get", "agent", "1"},
			expectOutput:  "ID\tNAME   \tVERSION\tSTATUS   \tCONNECTED\tDISCONNECTED\tLABELS \n1 \tAgent 1\t1.0.0  \tConnected\t-        \t-           \t      \t\n",
			expectedError: nil,
		},
		{
			description:   "get agent 2",
			args:          []string{"get", "agent", "2"},
			expectOutput:  "ID\tNAME   \tVERSION\tSTATUS      \tCONNECTED\tDISCONNECTED\tLABELS \n2 \tAgent 2\t1.0.0  \tDisconnected\t-        \t-           \t      \t\n",
			expectedError: nil,
		},
		{
			description:   "get agent 1 2",
			args:          []string{"get", "agent", "1", "2"},
			expectOutput:  "ID\tNAME   \tVERSION\tSTATUS      \tCONNECTED\tDISCONNECTED\tLABELS \n1 \tAgent 1\t1.0.0  \tConnected   \t-        \t-           \t      \t\n2 \tAgent 2\t1.0.0  \tDisconnected\t-        \t-           \t      \t\n",
			expectedError: nil,
		},
		{
			description:   "get agent 2 1",
			args:          []string{"get", "agent", "2", "1"},
			expectOutput:  "ID\tNAME   \tVERSION\tSTATUS      \tCONNECTED\tDISCONNECTED\tLABELS \n2 \tAgent 2\t1.0.0  \tDisconnected\t-        \t-           \t      \t\n1 \tAgent 1\t1.0.0  \tConnected   \t-        \t-           \t      \t\n",
			expectedError: nil,
		},
		{
			description:   "get agent 3",
			args:          []string{"get", "agent", "3"},
			expectOutput:  "No matching resources found.\n",
			expectedError: multierror.Append(errors.New("unable to get agents, got 404 Not Found\t"), errors.New("no agents found with name 3")),
		},
	}

	home := commands.BindplaneHome()

	var h = profile.NewHelper(home)

	// We need to perform this before creating a new bindplane cli because bindplane cli
	// creates a new logger with a file in ~/.bindplane
	err := h.HomeFolderSetup()
	if err != nil {
		fmt.Printf("error while trying to set up BindPlane home directory %s, %s\n", home, err.Error())
		os.Exit(1)
	}

	buffer := bytes.NewBufferString("")

	// Initialize the BindPlane CLI
	bindplane := cli.NewBindPlane(common.InitConfig(home), buffer)
	bindplane.SetClient(&mockClient{})

	// root command is neccessary to inherit error handling behavior
	rootCmd := commands.Command(bindplane, "bindplanectl")

	rootCmd.SetOut(buffer)
	cmd := Command(bindplane)
	cmd.SetOut(buffer)
	rootCmd.AddCommand(cmd)

	// The following should replace the above setup code to better reflect reality.
	// buffer := bytes.NewBufferString("")
	// rootCmd, bindplane := cmd.bindplanectl.main.SetupCobra(buffer)

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {

			rootCmd.SetArgs(test.args)
			cmdError := rootCmd.Execute()

			out, err := ioutil.ReadAll(buffer)
			require.NoError(t, err)

			assert.Equal(t, test.expectOutput, string(out))
			assert.Equal(t, test.expectedError, cmdError)
		})
	}
}
