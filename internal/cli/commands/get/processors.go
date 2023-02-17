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

	"github.com/observiq/bindplane-op/client"
	"github.com/observiq/bindplane-op/internal/cli"
	"github.com/observiq/bindplane-op/model"
	"github.com/spf13/cobra"
)

// ProcessorsCommand returns the BindPlane get processors cobra command
func ProcessorsCommand(bindplane *cli.BindPlane) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "processors [name]",
		Aliases: []string{"processor"},
		Short:   "Displays the processors",
		Long:    `A processor transforms logs, metrics, and traces.`,
		RunE: getImpl(bindplane, "processors", getter[*model.Processor]{
			some: func(ctx context.Context, client client.BindPlane, name string) (*model.Processor, bool, error) {
				item, err := client.Processor(ctx, name)
				return item, item != nil, err
			},
			all: func(ctx context.Context, client client.BindPlane) ([]*model.Processor, error) {
				return client.Processors(ctx)
			},
		}),
	}
	return cmd
}
