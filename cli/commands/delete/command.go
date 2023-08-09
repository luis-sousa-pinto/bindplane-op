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

// Package delete provides the delete command, as well as subcommands for deleting
// specific resources and resource types.
package delete

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/observiq/bindplane-op/model"
)

// Command returns the bindplane delete cobra command
func Command(builder Builder) *cobra.Command {
	var fileFlag []string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete bindplane resources",
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

			deleter, err := builder.BuildDeleter(ctx)
			if err != nil {
				return err
			}

			switch fileArgs[0] {
			case "-":
				reader := cmd.InOrStdin()
				statuses, err := deleter.DeleteResourcesFromReader(ctx, reader)
				if err != nil {
					return err
				}

				resourceStatuses = append(resourceStatuses, statuses...)
			default:
				statuses, err := deleter.DeleteResourcesFromFiles(ctx, fileArgs)
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

	cmd.AddCommand(
		deleteResourceCommand(builder, "agent", model.KindAgent, []string{"agents"}),
		deleteResourceCommand(builder, "agent-version", model.KindAgentVersion, []string{"agent-versions"}),
		deleteResourceCommand(builder, "configuration", model.KindConfiguration, []string{"configurations", "configs", "config"}),
		deleteResourceCommand(builder, "source", model.KindSource, []string{"sources"}),
		deleteResourceCommand(builder, "source-type", model.KindSourceType, []string{"source-types", "sourceType", "sourceTypes"}),
		deleteResourceCommand(builder, "processor", model.KindProcessor, []string{"processors"}),
		deleteResourceCommand(builder, "processor-type", model.KindProcessorType, []string{"processor-types", "processorType", "processorTypes"}),
		deleteResourceCommand(builder, "destination", model.KindDestination, []string{"destinations"}),
		deleteResourceCommand(builder, "destination-type", model.KindDestinationType, []string{"destination-types", "destinationType", "destinationTypes"}),
	)

	return cmd
}

func deleteResourceCommand(builder Builder, resourceType string, kind model.Kind, aliases []string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     fmt.Sprintf("%s <name>", resourceType),
		Aliases: aliases,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("missing required argument <name>")
			}

			ctx := cmd.Context()
			writer := cmd.OutOrStdout()
			deleter, err := builder.BuildDeleter(ctx)
			if err != nil {
				return err
			}

			if err := deleter.DeleteResources(ctx, kind, args); err != nil {
				return err
			}

			fmt.Fprintln(writer, "Successfully deleted resources")
			return nil
		},
	}

	return cmd
}
