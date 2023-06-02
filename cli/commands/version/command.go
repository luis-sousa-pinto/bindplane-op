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

// Package version provides the version command, which prints the BindPlane version.
package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Command returns a new BindPlane versions cobra command.
func Command(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Prints BindPlane version",
		Long:  `Prints BindPlane build version (commit or tag).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			writer := cmd.OutOrStdout()
			versioner, err := builder.BuildVersioner(ctx)
			if err != nil {
				return err
			}

			serverVersion, err := versioner.GetServerVersion(ctx)
			if err != nil {
				m := fmt.Sprintf("Failed to get version from server: %s\n", err.Error())
				fmt.Fprint(writer, m)
				return err
			}

			clientVersion, err := versioner.GetClientVersion(ctx)
			if err != nil {
				m := fmt.Sprintf("Failed to get version from client: %s\n", err.Error())
				fmt.Fprint(writer, m)
				return err
			}

			fmt.Fprintf(writer, "client: %s\nserver: %s\n", clientVersion.FullString(), serverVersion.FullString())
			return nil
		},
	}

	return cmd
}
