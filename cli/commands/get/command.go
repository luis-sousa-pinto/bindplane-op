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
	"errors"

	"github.com/observiq/bindplane-op/client"
	"github.com/observiq/bindplane-op/model"
	"github.com/spf13/cobra"
)

// historyFlag when --history is set print the resources history
var historyFlag bool

// errHistoryNotSupported is the error
var errHistoryNotSupported = errors.New("history is not supported for this resource kind")

// Command returns the BindPlane get cobra command.
func Command(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Display one or more resources",
	}

	cmd.AddCommand(
		ResourcesCommand(builder),
		AgentsCommand(builder),
		AgentVersionsCommand(builder),
		ConfigurationsCommand(builder),
		DestinationsCommand(builder),
		DestinationTypesCommand(builder),
		ProcessorsCommand(builder),
		ProcessorTypesCommand(builder),
		SourcesCommand(builder),
		SourceTypesCommand(builder),
		RolloutsCommand(builder),
	)

	cmd.PersistentFlags().BoolVar(&historyFlag, "history", false, "If true, list the history of the resource.")
	return cmd
}

// ResourcesCommand returns the BindPlane resources cobra command
func ResourcesCommand(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resources",
		Short: "Displays all resources",
		Long:  `Use -o yaml to export all resources to yaml.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			getter, err := builder.BuildGetter(cmd.Context())
			if err != nil {
				return err
			}

			return getter.GetAllResources(cmd.Context())
		},
	}
	return cmd
}

// AgentsCommand returns the BindPlane get agents cobra command
func AgentsCommand(builder Builder) *cobra.Command {
	var (
		selector string
		query    string
		limit    int
		offset   int
	)

	cmd := &cobra.Command{
		Use:     "agents [id]",
		Aliases: []string{"agent"},
		Short:   "Displays the agents",
		Long:    `An agent collects logs, metrics, and traces for sources and sends them to destinations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			getter, err := builder.BuildGetter(ctx)
			if err != nil {
				return err
			}

			switch len(args) {
			case 0:
				queryOpts := client.QueryOptions{
					Selector: selector,
					Query:    query,
					Offset:   offset,
					Limit:    limit,
				}
				return getter.GetResourcesOfKind(ctx, model.KindAgent, queryOpts)
			case 1:
				if historyFlag {
					return errHistoryNotSupported
				}
				return getter.GetResource(ctx, model.KindAgent, args[0])
			default:
				return getter.GetResources(ctx, model.KindAgent, args)
			}
		},
	}

	cmd.Flags().StringVarP(&selector, "selector", "l", "", "label selector to filter agents by label, e.g. name=value")
	cmd.Flags().StringVarP(&query, "query", "q", "", "search query to filter agents")
	cmd.Flags().IntVar(&offset, "offset", 0, "number of agents to skip for paging")
	cmd.Flags().IntVar(&limit, "limit", 100, "maximum number of agents to return")

	return cmd
}

// AgentVersionsCommand returns the BindPlane get source-types cobra command
func AgentVersionsCommand(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "agent-versions [id]",
		Aliases: []string{"agent-version"},
		Short:   "Displays the agent versions",
		Long:    `An agent version defines a specific version of an agent with links to the release package.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Resources(cmd.Context(), builder, model.KindAgentVersion, args)
		},
	}
	return cmd
}

// ConfigurationsCommand returns the BindPlane get configurations cobra command
func ConfigurationsCommand(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "configurations [name]",
		Aliases: []string{"configuration", "configs", "config"},
		Short:   "Displays the configurations",
		Long:    "A configuration provides a complete agent configuration to ship logs, metrics, and traces",
		RunE: func(cmd *cobra.Command, args []string) error {
			return Resources(cmd.Context(), builder, model.KindConfiguration, args)
		},
	}

	return cmd
}

// DestinationTypesCommand returns the BindPlane get destination-types cobra command
func DestinationTypesCommand(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "destination-types [name]",
		Aliases: []string{"destination-type"},
		Short:   "Displays the destination types",
		Long:    `A destination type is a type of service that receives logs, metrics, and traces.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Resources(cmd.Context(), builder, model.KindDestinationType, args)
		},
	}
	return cmd
}

// DestinationsCommand returns the BindPlane get destinations cobra command
func DestinationsCommand(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "destinations [name]",
		Aliases: []string{"destination"},
		Short:   "Displays the destinations",
		Long:    `A destination is a service that receives logs, metrics, and traces.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Resources(cmd.Context(), builder, model.KindDestination, args)
		},
	}
	return cmd
}

// ProcessorTypesCommand returns the BindPlane get processor-types cobra command
func ProcessorTypesCommand(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "processor-types [name]",
		Aliases: []string{"processor-type"},
		Short:   "Displays the processor types",
		Long:    `A processor type is a type of service that transforms logs, metrics, and traces.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Resources(cmd.Context(), builder, model.KindProcessorType, args)
		},
	}
	return cmd
}

// ProcessorsCommand returns the BindPlane get processors cobra command
func ProcessorsCommand(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "processors [name]",
		Aliases: []string{"processor"},
		Short:   "Displays the processors",
		Long:    `A processor transforms logs, metrics, and traces.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Resources(cmd.Context(), builder, model.KindProcessor, args)
		},
	}
	return cmd
}

// SourceTypesCommand returns the BindPlane get source-types cobra command
func SourceTypesCommand(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "source-types [name]",
		Aliases: []string{"source-type"},
		Short:   "Displays the source types",
		Long:    `A source type is a type of source that collects logs, metrics, and traces.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Resources(cmd.Context(), builder, model.KindSourceType, args)
		},
	}
	return cmd
}

// SourcesCommand returns the BindPlane get sources cobra command
func SourcesCommand(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sources [name]",
		Aliases: []string{"source"},
		Short:   "Displays the sources",
		Long:    `A source collects logs, metrics, and traces.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Resources(cmd.Context(), builder, model.KindSource, args)
		},
	}
	return cmd
}

// RolloutsCommand returns the BindPlane get rollouts cobra command
func RolloutsCommand(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rollouts [name]",
		Aliases: []string{"source"},
		Short:   "Displays the rollouts",
		Long:    `A rollout configurations agents with a configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if historyFlag {
				// If we enable the history flag for rollouts it should pull the history for the configuration of the same name
				return Resources(cmd.Context(), builder, model.KindConfiguration, args)
			}
			return Resources(cmd.Context(), builder, model.KindRollout, args)
		},
	}
	return cmd
}

// Resources gets resources based on the kind and args
func Resources(ctx context.Context, builder Builder, kind model.Kind, args []string) error {
	getter, err := builder.BuildGetter(ctx)
	if err != nil {
		return err
	}

	switch len(args) {
	case 0:
		return getter.GetResourcesOfKind(ctx, kind, client.QueryOptions{})
	case 1:
		if historyFlag {
			// If this isn't a kind that supports history then return an error
			if !model.HasVersionKind(kind) {
				return errHistoryNotSupported
			}

			return getter.GetResourceHistory(ctx, kind, args[0])
		}
		return getter.GetResource(ctx, kind, args[0])
	default:
		return getter.GetResources(ctx, kind, args)
	}
}
