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

// Package serve provides the `serve` command for the CLI.
// The command starts the server.
package serve

import (
	"context"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/observiq/bindplane-op/internal/cli"
	"github.com/observiq/bindplane-op/internal/cli/flags"
)

// Command returns the BindPlane serve cobra command
func Command(bindplane *cli.BindPlane) *cobra.Command {
	var forceConsoleColor bool
	var skipSeed bool

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Starts the server",
		Long:  `Serves websockets for agents, REST for cli, and GraphQL.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s := Server{
				logger: bindplane.Logger(),
			}
			if err := s.Start(context.Background(), bindplane, forceConsoleColor, skipSeed); err != nil {
				bindplane.Logger().Error("unable to Start the server", zap.Error(err))
				return cli.FormatError(err)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&forceConsoleColor, "force-console-color", false, "If true, gin.ForceConsoleColor() will be called.")
	cmd.Flags().BoolVar(&skipSeed, "skip-seed", false, "If true, store will not seed ResourceTypes present in /resources")
	_ = cmd.Flags().MarkHidden("force-console-color")

	flags.Serve(cmd)

	return cmd
}
