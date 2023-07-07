// Copyright observIQ, Inc.
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

package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/observiq/bindplane-op/graphql/generated"
	model1 "github.com/observiq/bindplane-op/graphql/model"
	"github.com/observiq/bindplane-op/model"
	"github.com/observiq/bindplane-op/model/graph"
	"github.com/observiq/bindplane-op/model/otel"
	"github.com/observiq/bindplane-op/store"
)

// Labels is the resolver for the labels field.
func (r *agentResolver) Labels(ctx context.Context, obj *model.Agent) (map[string]interface{}, error) {
	labels := map[string]interface{}{}
	for k := range obj.Labels.Set {
		labels[k] = obj.Labels.Get(k)
	}
	return labels, nil
}

// Status is the resolver for the status field.
func (r *agentResolver) Status(ctx context.Context, obj *model.Agent) (int, error) {
	return int(obj.Status), nil
}

// Configuration is the resolver for the configuration field.
func (r *agentResolver) Configuration(ctx context.Context, obj *model.Agent) (*model1.AgentConfiguration, error) {
	ac := &model1.AgentConfiguration{}
	if err := mapstructure.Decode(obj.Configuration, ac); err != nil {
		return &model1.AgentConfiguration{}, err
	}

	return ac, nil
}

// ConfigurationResource is the resolver for the configurationResource field.
func (r *agentResolver) ConfigurationResource(ctx context.Context, obj *model.Agent) (*model.Configuration, error) {
	return r.Bindplane.Store().AgentConfiguration(ctx, obj)
}

// UpgradeAvailable is the resolver for the upgradeAvailable field.
func (r *agentResolver) UpgradeAvailable(ctx context.Context, obj *model.Agent) (*string, error) {
	return r.Resolver.UpgradeAvailable(ctx, obj)
}

// Features is the resolver for the features field.
func (r *agentResolver) Features(ctx context.Context, obj *model.Agent) (int, error) {
	return int(obj.Features()), nil
}

// MatchLabels is the resolver for the matchLabels field.
func (r *agentSelectorResolver) MatchLabels(ctx context.Context, obj *model.AgentSelector) (map[string]interface{}, error) {
	labels := map[string]interface{}{}
	for k := range obj.MatchLabels {
		labels[k] = obj.MatchLabels[k]
	}
	return labels, nil
}

// Status is the resolver for the status field.
func (r *agentUpgradeResolver) Status(ctx context.Context, obj *model.AgentUpgrade) (int, error) {
	return int(obj.Status), nil
}

// Kind is the resolver for the kind field.
func (r *configurationResolver) Kind(ctx context.Context, obj *model.Configuration) (string, error) {
	return string(obj.GetKind()), nil
}

// AgentCount is the resolver for the agentCount field.
func (r *configurationResolver) AgentCount(ctx context.Context, obj *model.Configuration) (*int, error) {
	ids, err := r.Bindplane.Store().AgentsIDsMatchingConfiguration(ctx, obj)
	if err != nil {
		return nil, err
	}
	count := len(ids)
	return &count, nil
}

// ActiveTypes is the resolver for the activeTypes field.
func (r *configurationResolver) ActiveTypes(ctx context.Context, obj *model.Configuration) ([]string, error) {
	activeTypes := make([]string, 0, 3)
	usage := obj.Usage(ctx, r.Bindplane.Store())
	if usage.ActiveFlags().Includes(otel.LogsFlag) {
		activeTypes = append(activeTypes, "logs")
	}
	if usage.ActiveFlags().Includes(otel.MetricsFlag) {
		activeTypes = append(activeTypes, "metrics")
	}
	if usage.ActiveFlags().Includes(otel.TracesFlag) {
		activeTypes = append(activeTypes, "traces")
	}

	return activeTypes, nil
}

// Graph is the resolver for the graph field.
func (r *configurationResolver) Graph(ctx context.Context, obj *model.Configuration) (*graph.Graph, error) {
	return obj.Graph(ctx, r.Bindplane.Store())
}

// Rendered is the resolver for the rendered field.
func (r *configurationResolver) Rendered(ctx context.Context, obj *model.Configuration) (*string, error) {
	rendered, err := obj.Render(ctx, nil, r.Bindplane.BindPlaneURL(), r.Bindplane.BindPlaneInsecureSkipVerify(), r.Bindplane.Store(), model.GetOssOtelHeaders())
	if err != nil {
		return nil, err
	}
	return &rendered, nil
}

// Kind is the resolver for the kind field.
func (r *destinationResolver) Kind(ctx context.Context, obj *model.Destination) (string, error) {
	return string(obj.GetKind()), nil
}

// Kind is the resolver for the kind field.
func (r *destinationTypeResolver) Kind(ctx context.Context, obj *model.DestinationType) (string, error) {
	return string(obj.GetKind()), nil
}

// Labels is the resolver for the labels field.
func (r *metadataResolver) Labels(ctx context.Context, obj *model.Metadata) (map[string]interface{}, error) {
	labels := map[string]interface{}{}
	for k := range obj.Labels.Set {
		labels[k] = obj.Labels.Get(k)
	}
	return labels, nil
}

// UpdateProcessors is the resolver for the updateProcessors field.
func (r *mutationResolver) UpdateProcessors(ctx context.Context, input model1.UpdateProcessorsInput) (*bool, error) {
	return r.Resolver.UpdateProcessors(ctx, input)
}

// RemoveAgentConfiguration sets the given agent's `configuration` label to blank
func (r *mutationResolver) RemoveAgentConfiguration(ctx context.Context, input *model1.RemoveAgentConfigurationInput) (*model.Agent, error) {
	return r.Resolver.RemoveAgentConfiguration(ctx, input)
}

// ClearAgentUpgradeError is the resolver for the clearAgentUpgradeError field.
func (r *mutationResolver) ClearAgentUpgradeError(ctx context.Context, input model1.ClearAgentUpgradeErrorInput) (*bool, error) {
	_, err := r.Bindplane.Store().UpsertAgent(ctx, input.AgentID, model1.ClearCurrentAgentUpgradeError)
	return nil, err
}

// EditConfigurationDescription is the resolver for the editConfigurationDescription field.
func (r *mutationResolver) EditConfigurationDescription(ctx context.Context, input model1.EditConfigurationDescriptionInput) (*bool, error) {
	_, _, err := r.Bindplane.Store().UpdateConfiguration(ctx, input.Name, func(current *model.Configuration) {
		current.Metadata.Description = input.Description
	})

	return nil, err
}

// Type is the resolver for the type field.
func (r *parameterDefinitionResolver) Type(ctx context.Context, obj *model.ParameterDefinition) (model1.ParameterType, error) {
	return r.Resolver.Type(ctx, obj)
}

// Labels is the resolver for the labels field.
func (r *parameterOptionsResolver) Labels(ctx context.Context, obj *model.ParameterOptions) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	for k, v := range obj.Labels {
		result[k] = v
	}
	return result, nil
}

// Kind is the resolver for the kind field.
func (r *processorResolver) Kind(ctx context.Context, obj *model.Processor) (string, error) {
	return string(obj.GetKind()), nil
}

// Kind is the resolver for the kind field.
func (r *processorTypeResolver) Kind(ctx context.Context, obj *model.ProcessorType) (string, error) {
	return string(obj.GetKind()), nil
}

// OverviewPage is the resolver for the overviewPage field.
func (r *queryResolver) OverviewPage(ctx context.Context, configIDs []string, destinationIDs []string, period string, telemetryType string) (*model1.OverviewPage, error) {
	return r.Resolver.OverviewPage(ctx, configIDs, destinationIDs, period, telemetryType)
}

// Agents is the resolver for the agents field.
func (r *queryResolver) Agents(ctx context.Context, selector *string, query *string) (*model1.Agents, error) {
	return r.Resolver.Agents(ctx, selector, query)
}

// Agent is the resolver for the agent field.
func (r *queryResolver) Agent(ctx context.Context, id string) (*model.Agent, error) {
	return r.Bindplane.Store().Agent(ctx, id)
}

// Configurations is the resolver for the configurations field.
func (r *queryResolver) Configurations(ctx context.Context, selector *string, query *string, onlyDeployedConfigurations *bool) (*model1.Configurations, error) {
	return r.Resolver.Configurations(ctx, selector, query, onlyDeployedConfigurations)
}

// Configuration is the resolver for the configuration field.
func (r *queryResolver) Configuration(ctx context.Context, name string) (*model.Configuration, error) {
	return r.Bindplane.Store().Configuration(ctx, name)
}

// ConfigurationHistory is the resolver for the configurationHistory field.
func (r *queryResolver) ConfigurationHistory(ctx context.Context, name string) ([]*model.Configuration, error) {
	archive, ok := r.Bindplane.Store().(store.ArchiveStore)
	if !ok {
		return nil, errors.New("cannot get configuration history from non-archive store")
	}

	history, err := archive.ResourceHistory(ctx, model.KindConfiguration, name)
	if err != nil {
		return nil, fmt.Errorf("configurationHistory resolver, archive: %w", err)
	}

	configurationHistory, err := model.Parse[*model.Configuration](history)
	if err != nil {
		return nil, fmt.Errorf("configurationHistory resolver, parsing history: %w", err)
	}

	return configurationHistory, nil
}

// Sources is the resolver for the sources field.
func (r *queryResolver) Sources(ctx context.Context) ([]*model.Source, error) {
	return r.Bindplane.Store().Sources(ctx)
}

// Source is the resolver for the source field.
func (r *queryResolver) Source(ctx context.Context, name string) (*model.Source, error) {
	return r.Bindplane.Store().Source(ctx, name)
}

// SourceTypes is the resolver for the sourceTypes field.
func (r *queryResolver) SourceTypes(ctx context.Context) ([]*model.SourceType, error) {
	return r.Bindplane.Store().SourceTypes(ctx)
}

// SourceType is the resolver for the sourceType field.
func (r *queryResolver) SourceType(ctx context.Context, name string) (*model.SourceType, error) {
	return r.Bindplane.Store().SourceType(ctx, name)
}

// Processors is the resolver for the processors field.
func (r *queryResolver) Processors(ctx context.Context) ([]*model.Processor, error) {
	return r.Bindplane.Store().Processors(ctx)
}

// Processor is the resolver for the processor field.
func (r *queryResolver) Processor(ctx context.Context, name string) (*model.Processor, error) {
	return r.Bindplane.Store().Processor(ctx, name)
}

// ProcessorTypes is the resolver for the processorTypes field.
func (r *queryResolver) ProcessorTypes(ctx context.Context) ([]*model.ProcessorType, error) {
	return r.Bindplane.Store().ProcessorTypes(ctx)
}

// ProcessorType is the resolver for the processorType field.
func (r *queryResolver) ProcessorType(ctx context.Context, name string) (*model.ProcessorType, error) {
	return r.Bindplane.Store().ProcessorType(ctx, name)
}

// Destinations is the resolver for the destinations field.
func (r *queryResolver) Destinations(ctx context.Context) ([]*model.Destination, error) {
	return r.Bindplane.Store().Destinations(ctx)
}

// Destination is the resolver for the destination field.
func (r *queryResolver) Destination(ctx context.Context, name string) (*model.Destination, error) {
	return r.Bindplane.Store().Destination(ctx, name)
}

// DestinationWithType is the resolver for the destinationWithType field.
func (r *queryResolver) DestinationWithType(ctx context.Context, name string) (*model1.DestinationWithType, error) {
	return r.Resolver.DestinationWithType(ctx, name)
}

// DestinationsInConfigs is the resolver for the destinationsInConfigs field.
func (r *queryResolver) DestinationsInConfigs(ctx context.Context) ([]*model.Destination, error) {
	return r.Resolver.DestinationsInConfigs(ctx)
}

// DestinationTypes is the resolver for the destinationTypes field.
func (r *queryResolver) DestinationTypes(ctx context.Context) ([]*model.DestinationType, error) {
	return r.Bindplane.Store().DestinationTypes(ctx)
}

// DestinationType is the resolver for the destinationType field.
func (r *queryResolver) DestinationType(ctx context.Context, name string) (*model.DestinationType, error) {
	return r.Bindplane.Store().DestinationType(ctx, name)
}

// Snapshot is the resolver for the snapshot field.
func (r *queryResolver) Snapshot(ctx context.Context, agentID string, pipelineType otel.PipelineType, position *string, resourceName *string) (*model1.Snapshot, error) {
	return r.Resolver.Snapshot(ctx, agentID, pipelineType, position, resourceName)
}

// AgentMetrics is the resolver for the agentMetrics field.
func (r *queryResolver) AgentMetrics(ctx context.Context, period string, ids []string) (*model1.GraphMetrics, error) {
	return AgentMetrics(ctx, r.Bindplane, period, ids)
}

// ConfigurationMetrics is the resolver for the configurationMetrics field.
func (r *queryResolver) ConfigurationMetrics(ctx context.Context, period string, name *string) (*model1.GraphMetrics, error) {
	return ConfigurationMetrics(ctx, r.Bindplane, period, name)
}

// OverviewMetrics is the resolver for the overviewMetrics field.
func (r *queryResolver) OverviewMetrics(ctx context.Context, period string, configIDs []string, destinationIDs []string) (*model1.GraphMetrics, error) {
	return OverviewMetrics(ctx, r.Bindplane, period, configIDs, destinationIDs)
}

// Operator is the resolver for the operator field.
func (r *relevantIfConditionResolver) Operator(ctx context.Context, obj *model.RelevantIfCondition) (model1.RelevantIfOperatorType, error) {
	return model1.RelevantIfOperatorType(obj.Operator), nil
}

// Completed is the resolver for the completed field.
func (r *rolloutResolver) Completed(ctx context.Context, obj *model.Rollout) (int, error) {
	return obj.Progress.Completed, nil
}

// Errors is the resolver for the errors field.
func (r *rolloutResolver) Errors(ctx context.Context, obj *model.Rollout) (int, error) {
	return obj.Progress.Errors, nil
}

// Pending is the resolver for the pending field.
func (r *rolloutResolver) Pending(ctx context.Context, obj *model.Rollout) (int, error) {
	return obj.Progress.Pending, nil
}

// Waiting is the resolver for the waiting field.
func (r *rolloutResolver) Waiting(ctx context.Context, obj *model.Rollout) (int, error) {
	return obj.Progress.Waiting, nil
}

// Kind is the resolver for the kind field.
func (r *sourceResolver) Kind(ctx context.Context, obj *model.Source) (string, error) {
	return string(obj.GetKind()), nil
}

// Kind is the resolver for the kind field.
func (r *sourceTypeResolver) Kind(ctx context.Context, obj *model.SourceType) (string, error) {
	return string(obj.GetKind()), nil
}

// AgentChanges is the resolver for the agentChanges field.
func (r *subscriptionResolver) AgentChanges(ctx context.Context, selector *string, query *string) (<-chan []*model1.AgentChange, error) {
	return r.Resolver.AgentChanges(ctx, selector, query)
}

// ConfigurationChanges is the resolver for the configurationChanges field.
func (r *subscriptionResolver) ConfigurationChanges(ctx context.Context, selector *string, query *string) (<-chan []*model1.ConfigurationChange, error) {
	return r.Resolver.ConfigurationChanges(ctx, selector, query)
}

// AgentMetricsSubscription is the resolver for the agentMetricsSubscription field.
func (r *subscriptionResolver) AgentMetrics(ctx context.Context, period string, ids []string) (<-chan *model1.GraphMetrics, error) {
	return r.Resolver.AgentMetrics(ctx, period, ids)
}

// ConfigurationMetricsSubscription is the resolver for the configurationMetricsSubscription field.
func (r *subscriptionResolver) ConfigurationMetrics(ctx context.Context, period string, name *string, agent *string) (<-chan *model1.GraphMetrics, error) {
	return r.Resolver.ConfigurationMetrics(ctx, period, name, agent)
}

// OverviewMetricsSubscription is the resolver for the overviewMetricsSubscription field.
func (r *subscriptionResolver) OverviewMetrics(ctx context.Context, period string, configIDs []string, destinationIDs []string) (<-chan *model1.GraphMetrics, error) {
	return r.Resolver.OverviewMetrics(ctx, period, configIDs, destinationIDs)
}

// Agent returns generated.AgentResolver implementation.
func (r *Resolver) Agent() generated.AgentResolver { return &agentResolver{r} }

// AgentSelector returns generated.AgentSelectorResolver implementation.
func (r *Resolver) AgentSelector() generated.AgentSelectorResolver { return &agentSelectorResolver{r} }

// AgentUpgrade returns generated.AgentUpgradeResolver implementation.
func (r *Resolver) AgentUpgrade() generated.AgentUpgradeResolver { return &agentUpgradeResolver{r} }

// Configuration returns generated.ConfigurationResolver implementation.
func (r *Resolver) Configuration() generated.ConfigurationResolver { return &configurationResolver{r} }

// Destination returns generated.DestinationResolver implementation.
func (r *Resolver) Destination() generated.DestinationResolver { return &destinationResolver{r} }

// DestinationType returns generated.DestinationTypeResolver implementation.
func (r *Resolver) DestinationType() generated.DestinationTypeResolver {
	return &destinationTypeResolver{r}
}

// Metadata returns generated.MetadataResolver implementation.
func (r *Resolver) Metadata() generated.MetadataResolver { return &metadataResolver{r} }

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// ParameterDefinition returns generated.ParameterDefinitionResolver implementation.
func (r *Resolver) ParameterDefinition() generated.ParameterDefinitionResolver {
	return &parameterDefinitionResolver{r}
}

// ParameterOptions returns generated.ParameterOptionsResolver implementation.
func (r *Resolver) ParameterOptions() generated.ParameterOptionsResolver {
	return &parameterOptionsResolver{r}
}

// Processor returns generated.ProcessorResolver implementation.
func (r *Resolver) Processor() generated.ProcessorResolver { return &processorResolver{r} }

// ProcessorType returns generated.ProcessorTypeResolver implementation.
func (r *Resolver) ProcessorType() generated.ProcessorTypeResolver { return &processorTypeResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// RelevantIfCondition returns generated.RelevantIfConditionResolver implementation.
func (r *Resolver) RelevantIfCondition() generated.RelevantIfConditionResolver {
	return &relevantIfConditionResolver{r}
}

// Rollout returns generated.RolloutResolver implementation.
func (r *Resolver) Rollout() generated.RolloutResolver { return &rolloutResolver{r} }

// Source returns generated.SourceResolver implementation.
func (r *Resolver) Source() generated.SourceResolver { return &sourceResolver{r} }

// SourceType returns generated.SourceTypeResolver implementation.
func (r *Resolver) SourceType() generated.SourceTypeResolver { return &sourceTypeResolver{r} }

// Subscription returns generated.SubscriptionResolver implementation.
func (r *Resolver) Subscription() generated.SubscriptionResolver { return &subscriptionResolver{r} }

type agentResolver struct{ *Resolver }
type agentSelectorResolver struct{ *Resolver }
type agentUpgradeResolver struct{ *Resolver }
type configurationResolver struct{ *Resolver }
type destinationResolver struct{ *Resolver }
type destinationTypeResolver struct{ *Resolver }
type metadataResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type parameterDefinitionResolver struct{ *Resolver }
type parameterOptionsResolver struct{ *Resolver }
type processorResolver struct{ *Resolver }
type processorTypeResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type relevantIfConditionResolver struct{ *Resolver }
type rolloutResolver struct{ *Resolver }
type sourceResolver struct{ *Resolver }
type sourceTypeResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
