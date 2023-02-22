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

// Package get provides the get command, which displays one or more resources,
// as well as subcommands for each resource type.
package get

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	"github.com/observiq/bindplane-op/client"
	"github.com/observiq/bindplane-op/internal/cli"
	"github.com/observiq/bindplane-op/internal/cli/printer"
	"github.com/observiq/bindplane-op/model"
)

// Command returns the BindPlane get cobra command.
func Command(bindplane *cli.BindPlane) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Display one or more resources",
	}

	cmd.AddCommand(
		ResourcesCommand(bindplane),
		AgentsCommand(bindplane),
		AgentVersionsCommand(bindplane),
		ConfigurationsCommand(bindplane),
		DestinationsCommand(bindplane),
		DestinationTypesCommand(bindplane),
		ProcessorsCommand(bindplane),
		ProcessorTypesCommand(bindplane),
		SourcesCommand(bindplane),
		SourceTypesCommand(bindplane),
	)

	return cmd
}

// ----------------------------------------------------------------------
// generic implementations for get

// 'some' is a function used for 'get' implementations that take arguments, e.g.
//			bindplanectl get agents 0738a9b0-7c36-46ae-800e-98b61f763654 caf962be-3950-4c13-a60b-10324f0bd304
// 'all' is a function that implements a call for all of the command type, e.g.
//			bindplanectl get agents
// will return all of the agents

type getter[T model.Printable] struct {
	some func(ctx context.Context, client client.BindPlane, name string) (T, bool, error)
	all  func(ctx context.Context, client client.BindPlane) ([]T, error)
}

// Since all the 'get' commands share common error handling and printing code,
// this function factors out that common code and uses the functions passed in the
// getter struct to execute the specific functionality, e.g. configuration, agents, sources, etc.

func getImpl[T model.Printable](bindplane *cli.BindPlane, resourceName string, g getter[T]) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		c, err := bindplane.Client()
		if err != nil {
			return fmt.Errorf("error creating client: %w", err)
		}

		var errGroup error
		if len(args) > 0 {
			items := []T{}
			for _, name := range args {
				item, exists, err := g.some(cmd.Context(), c, name)
				if err != nil {
					errGroup = multierror.Append(errGroup, err)
				}
				if !exists {
					errGroup = multierror.Append(errGroup, fmt.Errorf("no %s found with name %s", resourceName, name))
				} else {
					items = append(items, item)
				}
			}
			if len(items) == 1 {
				printer.PrintResource(bindplane.Printer(), items[0])
			} else {
				// PrintResources will print an error if there are no items
				printer.PrintResources(bindplane.Printer(), items)
			}
			return errGroup
		}

		items, err := g.all(cmd.Context(), c)
		if err != nil {
			return err
		}

		printer.PrintResources(bindplane.Printer(), items)
		return nil
	}
}
