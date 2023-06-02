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

// Package apply provides the apply command, which upserts resources from a file to the store.
package apply

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/observiq/bindplane-op/model"
)

// Command returns the bindplane apply cobra command.
func Command(builder Builder) *cobra.Command {
	var fileFlag []string

	cmd := &cobra.Command{
		Use:   "apply [file]",
		Short: "Apply resources",
		Long:  `Apply resources from a file with a filepath or use 'bindplane apply -' to apply resources from stdin.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// any positional args are treated as if they were prefixed with -f/--file. this allows shell globs to be used
			// with or without -f. for example, the following two commands are the same "apply -f *.yaml" and "apply *.yaml"
			fileArgs := fileFlag
			fileArgs = append(fileArgs, args...)

			if len(fileArgs) == 0 {
				_ = cmd.Help()
				return nil
			}

			ctx := cmd.Context()
			writer := cmd.OutOrStdout()
			var resourceStatuses []*model.AnyResourceStatus

			applier, err := builder.BuildApplier(ctx)
			if err != nil {
				return err
			}

			switch fileArgs[0] {
			case "-":
				reader := cmd.InOrStdin()
				statuses, err := applier.ApplyResourcesFromReader(ctx, reader)
				if err != nil {
					return err
				}

				resourceStatuses = append(resourceStatuses, statuses...)
			default:
				statuses, err := applier.ApplyResourcesFromFiles(ctx, fileArgs)
				if err != nil {
					return err
				}

				resourceStatuses = append(resourceStatuses, statuses...)
			}

			for _, status := range resourceStatuses {
				fmt.Fprintln(writer, status.Message())
			}

			return nil
		},
	}

	cmd.Flags().StringSliceVarP(&fileFlag, "file", "f", []string{}, "path to a yaml file that specifies bindplane resources")

	return cmd
}
