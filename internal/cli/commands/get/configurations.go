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
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/observiq/bindplane-op/client"
	"github.com/observiq/bindplane-op/internal/cli"
	"github.com/observiq/bindplane-op/model"
)

// ConfigurationsCommand returns the BindPlane get configurations cobra command
func ConfigurationsCommand(bindplane *cli.BindPlane) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "configurations [name]",
		Aliases: []string{"configuration", "configs", "config"},
		Short:   "Displays the configurations",
		Long:    "A configuration provides a complete agent configuration to ship logs, metrics, and traces",
		RunE: func(cmd *cobra.Command, args []string) error {
			// special case for raw configurations
			if bindplane.Config.Output == "raw" && len(args) > 0 {
				name := args[0]
				c, err := bindplane.Client()
				if err != nil {
					return fmt.Errorf("error creating client: %w", err)
				}
				raw, err := c.RawConfiguration(cmd.Context(), name)
				if err != nil {
					return err
				}
				if _, err := cmd.OutOrStdout().Write([]byte(raw)); err != nil {
					return err
				}
				return nil
			}

			// normal get case
			return getImpl(bindplane, "configurations", getter[*model.Configuration]{
				some: func(ctx context.Context, client client.BindPlane, name string) (*model.Configuration, bool, error) {
					item, err := client.Configuration(ctx, name)
					return item, item != nil, err
				},
				all: func(ctx context.Context, client client.BindPlane) ([]*model.Configuration, error) {
					return client.Configurations(ctx)
				},
			})(cmd, args)
		},
	}

	return cmd
}
