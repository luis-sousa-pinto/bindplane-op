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

// ProcessorTypesCommand returns the BindPlane get processor-types cobra command
func ProcessorTypesCommand(bindplane *cli.BindPlane) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "processor-types [name]",
		Aliases: []string{"processor-type"},
		Short:   "Displays the processor types",
		Long:    `A processor type is a type of service that transforms logs, metrics, and traces.`,
		RunE: getImpl(bindplane, "processor-types", getter[*model.ProcessorType]{
			some: func(ctx context.Context, client client.BindPlane, name string) (*model.ProcessorType, bool, error) {
				item, err := client.ProcessorType(ctx, name)
				return item, item != nil, err
			},
			all: func(ctx context.Context, client client.BindPlane) ([]*model.ProcessorType, error) {
				return client.ProcessorTypes(ctx)
			},
		}),
	}
	return cmd
}
