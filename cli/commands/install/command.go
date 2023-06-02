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

// Package install provides the install command, which installs a new agent.
package install

import (
	"context"
	"fmt"
	"runtime"

	"github.com/observiq/bindplane-op/client"
	"github.com/spf13/cobra"
)

var (
	platformFlag  string
	versionFlag   string
	labelsFlag    string
	secretKeyFlag string
	remoteURLFlag string
)

// Command returns the BindPlane install cobra command.
func Command(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install a new agent",
	}

	cmd.AddCommand(
		AgentCommand(builder),
	)

	return cmd
}

// AgentCommand returns the BindPlane install agent cobra command
func AgentCommand(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Displays the install command for an agent managed by this server.",
		Long:  `An agent collects logs, metrics, and traces for sources and sends them to destinations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			writer := cmd.OutOrStdout()

			options := client.AgentInstallOptions{
				Version:   versionFlag,
				Labels:    labelsFlag,
				Platform:  platformFlag,
				SecretKey: secretKeyFlag,
				RemoteURL: remoteURLFlag,
			}

			installer, err := builder.BuildInstaller(ctx)
			if err != nil {
				return err
			}

			command, err := installer.GetAgentInstallCommand(ctx, options)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(writer, command)
			return err
		},
	}

	defaultPlatform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)

	cmd.Flags().StringVar(&platformFlag, "platform", defaultPlatform, `platform where the agent will be installed, one of:
	linux           [alias for linux-amd64]
	macos           [alias for darwin-arm64]
	windows         [alias for window-amd64]
	linux-amd64
	linux-arm64
	linux-arm
	darwin-arm64
	darwin-amd64
	windows-amd64
`)
	cmd.Flags().StringVar(&versionFlag, "version", "latest", "version of the agent to install")
	cmd.Flags().StringVar(&labelsFlag, "labels", "", "labels to apply to the new agent")
	cmd.Flags().StringVar(&secretKeyFlag, "secret-key", "", "secret-key to assign to the agent")
	cmd.Flags().StringVar(&remoteURLFlag, "remote-url", "", "websocket address of the BindPlane agent management platform")

	return cmd
}
