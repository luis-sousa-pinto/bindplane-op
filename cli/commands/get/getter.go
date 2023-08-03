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
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/observiq/bindplane-op/cli/printer"
	"github.com/observiq/bindplane-op/client"
	"github.com/observiq/bindplane-op/model"
)

// PrintTitle prints the titleName with a dashed underline
func PrintTitle(titleName string) {
	fmt.Println("")
	fmt.Println(titleName)
	fmt.Println(strings.Repeat("-", len(titleName)))
}

// Getter is an interface for getting and printing resources.
type Getter interface {
	// GetResource gets and prints a single resource.
	GetResource(ctx context.Context, kind model.Kind, id string) error

	// GetResources gets and prints a list of resources.
	GetResources(ctx context.Context, kind model.Kind, ids []string) error

	// GetResourcesOfKind gets and prints all resources of a given kind.
	GetResourcesOfKind(ctx context.Context, kind model.Kind, queryOpts client.QueryOptions) error

	// GetAllResources gets and prints every resource kind.
	GetAllResources(ctx context.Context) error

	// GetResourceHistory gets and prints the history for a single resource
	GetResourceHistory(ctx context.Context, kind model.Kind, id string) error
}

// Builder is an interface for building a Getter.
type Builder interface {
	// Build returns a new Getter.
	BuildGetter(ctx context.Context) (Getter, error)
}

// NewGetter returns a new Getter.
func NewGetter(client client.BindPlane, printer printer.Printer, outputFormat string) *DefaultGetter {
	return &DefaultGetter{
		client:       client,
		printer:      printer,
		outputFormat: outputFormat,
	}
}

// DefaultGetter is the default implementation of Getter.
type DefaultGetter struct {
	client       client.BindPlane
	printer      printer.Printer
	outputFormat string
}

// GetResource gets and prints a single resource.
func (g *DefaultGetter) GetResource(ctx context.Context, kind model.Kind, id string) error {
	// Special case for raw resource flag
	if g.outputFormat == "raw" {
		return g.GetRawResource(ctx, kind, id)
	}

	resource, err := g.getPrintableResource(ctx, kind, id)
	if err != nil {
		return err
	}

	g.printer.PrintResource(resource)
	return nil
}

// GetResources gets and prints a list of resources matching the supplied ids.
func (g *DefaultGetter) GetResources(ctx context.Context, kind model.Kind, ids []string) error {
	resources, err := g.getPrintableResources(ctx, kind, ids)
	if err != nil {
		return err
	}

	g.printer.PrintResources(resources)
	return nil
}

// GetResourcesOfKind gets and prints all resources of a given kind.
func (g *DefaultGetter) GetResourcesOfKind(ctx context.Context, kind model.Kind, queryOpts client.QueryOptions) error {
	resources, err := g.getAllPrintableResources(ctx, kind, queryOpts)
	if err != nil {
		return err
	}

	g.printer.PrintResources(resources)
	return nil
}

// GetAllResources gets and prints every resource.
func (g *DefaultGetter) GetAllResources(ctx context.Context) error {
	var errGroup error

	blankQueryOpts := client.QueryOptions{}

	configurations, err := g.getAllPrintableResources(ctx, model.KindConfiguration, blankQueryOpts)
	if err != nil {
		errGroup = multierror.Append(errGroup, fmt.Errorf("failed to get configurations: %w", err))
	}
	PrintTitle("configurations")
	g.printer.PrintResources(configurations)

	sources, err := g.getAllPrintableResources(ctx, model.KindSource, blankQueryOpts)
	if err != nil {
		errGroup = multierror.Append(errGroup, fmt.Errorf("failed to get sources: %w", err))
	}
	PrintTitle("sources")
	g.printer.PrintResources(sources)

	processors, err := g.getAllPrintableResources(ctx, model.KindProcessor, blankQueryOpts)
	if err != nil {
		errGroup = multierror.Append(errGroup, fmt.Errorf("failed to get processors: %w", err))
	}
	PrintTitle("processors")
	g.printer.PrintResources(processors)

	destinations, err := g.getAllPrintableResources(ctx, model.KindDestination, blankQueryOpts)
	if err != nil {
		errGroup = multierror.Append(errGroup, fmt.Errorf("failed to get destinations: %w", err))
	}
	PrintTitle("destinations")
	g.printer.PrintResources(destinations)

	sourceTypes, err := g.getAllPrintableResources(ctx, model.KindSourceType, blankQueryOpts)
	if err != nil {
		errGroup = multierror.Append(errGroup, fmt.Errorf("failed to get source types: %w", err))
	}
	PrintTitle("source-types")
	g.printer.PrintResources(sourceTypes)

	processorTypes, err := g.getAllPrintableResources(ctx, model.KindProcessorType, blankQueryOpts)
	if err != nil {
		errGroup = multierror.Append(errGroup, fmt.Errorf("failed to get processor types: %w", err))
	}
	PrintTitle("processor-types")
	g.printer.PrintResources(processorTypes)

	destinationTypes, err := g.getAllPrintableResources(ctx, model.KindDestinationType, blankQueryOpts)
	if err != nil {
		errGroup = multierror.Append(errGroup, fmt.Errorf("failed to get destination types: %w", err))
	}
	PrintTitle("destination-types")
	g.printer.PrintResources(destinationTypes)

	agentVersions, err := g.getAllPrintableResources(ctx, model.KindAgentVersion, blankQueryOpts)
	if err != nil {
		errGroup = multierror.Append(errGroup, fmt.Errorf("failed to get agent versions: %w", err))
	}
	PrintTitle("agent-versions")
	g.printer.PrintResources(agentVersions)

	return errGroup
}

// GetResourceHistory gets and prints the history for a single resource
func (g *DefaultGetter) GetResourceHistory(ctx context.Context, kind model.Kind, id string) error {
	items, err := g.client.ResourceHistory(ctx, kind, id)
	if err != nil {
		return err
	}

	// convert to printables
	printables := make([]model.Printable, len(items))
	for i, item := range items {
		printable, err := model.AsKind[model.Resource](item)
		if err != nil {
			return fmt.Errorf("failed to convert %s: %w", item.ID(), err)
		}
		printables[i] = printable
	}

	g.printer.PrintResources(printables)
	return nil
}

// GetRawResource gets and prints the raw version of a resource
func (g *DefaultGetter) GetRawResource(ctx context.Context, kind model.Kind, id string) error {
	var rawConfig string
	var err error
	switch kind {
	case model.KindConfiguration:
		rawConfig, err = g.client.RawConfiguration(ctx, id)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("raw does not support resource type: %s", kind)
	}

	fmt.Println(rawConfig)
	return nil
}

// getPrintableResource gets a printable resource.
func (g *DefaultGetter) getPrintableResource(ctx context.Context, kind model.Kind, id string) (model.Printable, error) {
	var resource model.Printable
	var err error
	switch kind {
	case model.KindAgent:
		resource, err = g.client.Agent(ctx, id)
	case model.KindAgentVersion:
		resource, err = g.client.AgentVersion(ctx, id)
	case model.KindConfiguration:
		if ExportFlag {
			c := &model.Configuration{}
			c, err = g.client.Configuration(ctx, id)
			if err == nil {
				exportConfiguration(c)
				resource = c
			}
		} else {
			resource, err = g.client.Configuration(ctx, id)
		}
	case model.KindDestinationType:
		resource, err = g.client.DestinationType(ctx, id)
	case model.KindDestination:
		d := &model.Destination{}
		d, err = g.client.Destination(ctx, id)
		if err == nil && ExportFlag {
			d.Metadata = sanitizeMetadataForExport(d.Metadata)
			d.Spec.Type = model.TrimVersion(d.Spec.Type)
		}
		resource = d
	case model.KindProcessorType:
		resource, err = g.client.ProcessorType(ctx, id)
	case model.KindProcessor:
		p := &model.Processor{}
		p, err = g.client.Processor(ctx, id)
		if err == nil && ExportFlag {
			p.Metadata = sanitizeMetadataForExport(p.Metadata)
			p.Spec.Type = model.TrimVersion(p.Spec.Type)
		}
		resource = p
	case model.KindSource:
		s := &model.Source{}
		s, err = g.client.Source(ctx, id)
		if err == nil && ExportFlag {
			s.Metadata = sanitizeMetadataForExport(s.Metadata)
			s.Spec.Type = model.TrimVersion(s.Spec.Type)
		}
		resource = s
	case model.KindSourceType:
		resource, err = g.client.SourceType(ctx, id)
	case model.KindRollout:
		cfg, err := g.client.Configuration(ctx, id)
		if err != nil {
			return nil, err
		}

		resource = cfg.Rollout()
	default:
		return nil, fmt.Errorf("unknown resource type: %s", kind)
	}

	return resource, err
}

// getPrintableResources gets a list of printable resources matching the supplied ids.
func (g *DefaultGetter) getPrintableResources(ctx context.Context, kind model.Kind, ids []string) ([]model.Printable, error) {
	var resources []model.Printable
	for _, id := range ids {
		resource, err := g.getPrintableResource(ctx, kind, id)
		if err != nil {
			return nil, err
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// getAllPrintableResources gets all printable resources of a given type.
func (g *DefaultGetter) getAllPrintableResources(ctx context.Context, kind model.Kind, queryOpts client.QueryOptions) ([]model.Printable, error) {
	var resources []model.Printable
	switch kind {
	case model.KindAgent:
		agents, err := g.client.Agents(ctx, queryOpts)
		for _, agent := range agents {
			resources = append(resources, agent)
		}
		return resources, err
	case model.KindAgentVersion:
		agentVersions, err := g.client.AgentVersions(ctx)
		for _, agentVersion := range agentVersions {
			resources = append(resources, agentVersion)
		}
		return resources, err
	case model.KindConfiguration:
		configuration, err := g.client.Configurations(ctx)
		for _, configuration := range configuration {
			if ExportFlag {
				exportConfiguration(configuration)
			}
			resources = append(resources, configuration)
		}
		return resources, err

	case model.KindDestinationType:
		destinationTypes, err := g.client.DestinationTypes(ctx)
		for _, destinationType := range destinationTypes {
			resources = append(resources, destinationType)
		}
		return resources, err
	case model.KindDestination:
		destination, err := g.client.Destinations(ctx)
		for _, destination := range destination {
			if ExportFlag {
				destination.Metadata = sanitizeMetadataForExport(destination.Metadata)
				destination.Spec.Type = model.TrimVersion(destination.Spec.Type)
			}
			resources = append(resources, destination)
		}
		return resources, err
	case model.KindProcessorType:
		processorTypes, err := g.client.ProcessorTypes(ctx)
		for _, processorType := range processorTypes {
			resources = append(resources, processorType)
		}
		return resources, err
	case model.KindProcessor:
		processor, err := g.client.Processors(ctx)
		for _, processor := range processor {
			if ExportFlag {
				processor.Metadata = sanitizeMetadataForExport(processor.Metadata)
				processor.Spec.Type = model.TrimVersion(processor.Spec.Type)
			}
			resources = append(resources, processor)
		}
		return resources, err
	case model.KindSource:
		source, err := g.client.Sources(ctx)
		for _, source := range source {
			if ExportFlag {
				source.Metadata = sanitizeMetadataForExport(source.Metadata)
				source.Spec.Type = model.TrimVersion(source.Spec.Type)
			}
			resources = append(resources, source)
		}
		return resources, err
	case model.KindSourceType:
		sourceTypes, err := g.client.SourceTypes(ctx)
		for _, sourceType := range sourceTypes {
			resources = append(resources, sourceType)
		}
		return resources, err
	case model.KindRollout:
		cfgs, err := g.client.Configurations(ctx)
		if err != nil {
			return nil, err
		}

		for _, cfg := range cfgs {
			resources = append(resources, cfg.Rollout())
		}

		return resources, nil
	default:
		return nil, fmt.Errorf("unknown resource type: %s", kind)
	}
}
