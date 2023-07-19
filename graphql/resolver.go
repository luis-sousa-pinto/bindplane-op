// Copyright  observIQ, Inc
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

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/observiq/bindplane-op/eventbus"
	model1 "github.com/observiq/bindplane-op/graphql/model"
	"github.com/observiq/bindplane-op/internal/server"
	"github.com/observiq/bindplane-op/model"
	"github.com/observiq/bindplane-op/model/otel"
	bpotel "github.com/observiq/bindplane-op/model/otel"
	"github.com/observiq/bindplane-op/otlp/record"
	exposedserver "github.com/observiq/bindplane-op/server"
	"github.com/observiq/bindplane-op/server/protocol"
	"github.com/observiq/bindplane-op/store"
	"github.com/observiq/bindplane-op/store/search"
	"github.com/observiq/bindplane-op/store/stats"
	"github.com/observiq/bindplane-op/util/semver"
	oteltracer "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
)

var tracer = oteltracer.Tracer("graphql")

//go:generate gqlgen generate

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// Resolver is the root resolver for the graphql server.
type Resolver struct {
	Bindplane exposedserver.BindPlane
	Updates   eventbus.Source[store.BasicEventUpdates]
}

// NewResolver returns a new Resolver and starts a go routine
// that sends agent updates to observers.
func NewResolver(bindplane exposedserver.BindPlane) *Resolver {
	resolver := &Resolver{
		Bindplane: bindplane,
		Updates:   eventbus.NewSource[store.BasicEventUpdates](),
	}

	var receiver eventbus.Receiver[store.BasicEventUpdates] = resolver.Updates

	ctx := context.Background()
	// relay events from the store to the resolver where they will be dispatched to individual graphql subscriptions
	eventbus.Relay(ctx, bindplane.Store().Updates(ctx), receiver)

	return resolver
}

// ApplySelectorToChanges applies the selector to the changes and returns the changes that match the selector.
func ApplySelectorToChanges(selector *model.Selector, changes store.Events[*model.Agent]) store.Events[*model.Agent] {
	if selector == nil {
		return changes
	}
	result := store.NewEvents[*model.Agent]()
	for _, change := range changes {
		if change.Type != store.EventTypeRemove && !selector.Matches(change.Item.Labels) {
			result.Include(change.Item, store.EventTypeRemove)
		} else {
			result.Include(change.Item, change.Type)
		}
	}
	return result
}

// ApplyQueryToChanges applies the query to the changes and returns the changes that match the query.
func ApplyQueryToChanges(query *search.Query, index search.Index, changes store.Events[*model.Agent]) store.Events[*model.Agent] {
	if query == nil {
		return changes
	}
	result := store.NewEvents[*model.Agent]()
	for _, change := range changes {
		if change.Type != store.EventTypeRemove && !index.Matches(query, change.Item.ID) {
			result.Include(change.Item, store.EventTypeRemove)
		} else {
			result.Include(change.Item, change.Type)
		}
	}
	return result
}

// ApplySelectorToEvents applies the selector to the events and returns the events that match the selector.
func ApplySelectorToEvents[T model.Resource](selector *model.Selector, events store.Events[T]) store.Events[T] {
	if selector == nil {
		return events
	}
	result := store.NewEvents[T]()
	for _, event := range events {
		if event.Type != store.EventTypeRemove && !selector.Matches(event.Item.GetLabels()) {
			result.Include(event.Item, store.EventTypeRemove)
		} else {
			result.Include(event.Item, event.Type)
		}
	}
	return result
}

// ApplyQueryToEvents applies the query to the events and returns the events that match the query.
func ApplyQueryToEvents[T model.Resource](query *search.Query, index search.Index, events store.Events[T]) store.Events[T] {
	if query == nil || index == nil {
		return events
	}
	result := store.NewEvents[T]()
	for _, event := range events {
		if event.Type != store.EventTypeRemove && !index.Matches(query, event.Item.Name()) {
			result.Include(event.Item, store.EventTypeRemove)
		} else {
			result.Include(event.Item, event.Type)
		}
	}
	return result
}

// ParseSelectorAndQuery parses the selector and query strings and returns the appropriate options
func (r *Resolver) ParseSelectorAndQuery(selector *string, query *string) (*model.Selector, *search.Query, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var parsedSelector *model.Selector
	if selector != nil {
		sel, err := model.SelectorFromString(*selector)
		if err != nil {
			return nil, nil, err
		}
		parsedSelector = &sel
	}

	// parse the parsedQuery, if any
	var parsedQuery *search.Query
	if query != nil && *query != "" {
		q := search.ParseQuery(*query)
		q.ReplaceVersionLatest(ctx, r.Bindplane.Versions())
		parsedQuery = q
	}

	return parsedSelector, parsedQuery, nil
}

// QueryOptionsAndSuggestions parses the selector and query strings and returns the appropriate options
func (r *Resolver) QueryOptionsAndSuggestions(selector *string, query *string, index search.Index) ([]store.QueryOption, []*search.Suggestion, error) {
	parsedSelector, parsedQuery, err := r.ParseSelectorAndQuery(selector, query)
	if err != nil {
		return nil, nil, err
	}

	options := []store.QueryOption{}
	if parsedSelector != nil {
		options = append(options, store.WithSelector(*parsedSelector))
	}

	var suggestions []*search.Suggestion
	if parsedQuery != nil {
		options = append(options, store.WithQuery(parsedQuery))

		s, err := index.Suggestions(parsedQuery)
		if err != nil {
			return nil, nil, err
		}

		suggestions = s
	}
	return options, suggestions, nil
}

// HasAgentConfigurationChanges determines if there is an agent update
// in updates that would affect the list of configurations
func (r *Resolver) HasAgentConfigurationChanges(updates store.BasicEventUpdates) bool {
	for _, change := range updates.Agents() {
		// Only events type Remove, Label, and Insert could affect
		// the agentCount field.
		if change.Type != store.EventTypeUpdate {
			return true
		}
	}
	return false
}

// UpgradeAvailable is the resolver for the upgradeAvailable field.
// func (r *Resolver) UpgradeAvailable(ctx context.Context, obj *model.Agent) (*string, error) {
func (r *Resolver) UpgradeAvailable(ctx context.Context, obj *model.Agent) (*string, error) {
	if !obj.SupportsUpgrade() {
		return nil, nil
	}

	latestVersion, err := r.Bindplane.Versions().LatestVersion(ctx)
	if err != nil {
		return nil, nil
	}

	if latestVersion.SemanticVersion().IsNewer(semver.Parse(obj.Version)) {
		latestVersionString := latestVersion.AgentVersion()
		return &latestVersionString, nil
	}
	return nil, nil
}

// UpdateProcessors is the resolver for the updateProcessors field.
func (r *Resolver) UpdateProcessors(ctx context.Context, input model1.UpdateProcessorsInput) (*bool, error) {
	config, err := r.Bindplane.Store().Configuration(ctx, input.Configuration)
	if err != nil {
		return nil, err
	}

	if config == nil {
		return nil, fmt.Errorf("configuration not found")
	}

	processors := make([]model.ResourceConfiguration, len(input.Processors))
	for ix, p := range input.Processors {
		processors[ix] = *p
	}

	switch input.ResourceType {
	case model1.ResourceTypeKindDestination:
		config.Spec.Destinations[input.ResourceIndex].Processors = processors
	case model1.ResourceTypeKindSource:
		config.Spec.Sources[input.ResourceIndex].Processors = processors
	default:
		return nil, fmt.Errorf("invalid resource type, should be source or destination")
	}

	// Ensure that the config can still be rendered with the added processors
	_, err = config.Render(ctx, nil, r.Bindplane.BindPlaneURL(), r.Bindplane.BindPlaneInsecureSkipVerify(), r.Bindplane.Store(), model.GetOssOtelHeaders())
	if err != nil {
		return nil, fmt.Errorf("failed  to render config: %w", err)
	}

	statuses, err := r.Bindplane.Store().ApplyResources(ctx, []model.Resource{config})
	if err != nil {
		return nil, err
	}

	if statuses[0].Status == model.StatusError || statuses[0].Status == model.StatusInvalid {
		return nil, errors.New(statuses[0].Reason)
	}

	return nil, nil
}

// RemoveAgentConfiguration sets the given agent's `configuration` label to blank
func (r *Resolver) RemoveAgentConfiguration(ctx context.Context, input *model1.RemoveAgentConfigurationInput) (*model.Agent, error) {
	ctx, span := tracer.Start(ctx, "graphql/removeAgentConfiguration",
		trace.WithAttributes(attribute.String("bindplane.agent.id", input.AgentID)),
	)
	defer span.End()

	agent, err := r.Bindplane.Store().Agent(ctx, input.AgentID)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	newAgent, err := r.Bindplane.Store().UpsertAgent(ctx, agent.ID, func(current *model.Agent) {
		current.Labels.Set["configuration"] = ""
	})

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	return newAgent, nil
}

// Type is the resolver for the type field.
func (r *Resolver) Type(_ context.Context, obj *model.ParameterDefinition) (model1.ParameterType, error) {
	switch obj.Type {
	case "strings":
		return model1.ParameterTypeStrings, nil

	case "string":
		return model1.ParameterTypeString, nil

	case "enum":
		return model1.ParameterTypeEnum, nil

	case "bool":
		return model1.ParameterTypeBool, nil

	case "int":
		return model1.ParameterTypeInt, nil

	case "map":
		return model1.ParameterTypeMap, nil

	case "yaml":
		return model1.ParameterTypeYaml, nil

	case "enums":
		return model1.ParameterTypeEnums, nil

	case "timezone":
		return model1.ParameterTypeTimezone, nil

	case "metrics":
		return model1.ParameterTypeMetrics, nil

	case "awsCloudwatchNamedField":
		return model1.ParameterTypeAwsCloudwatchNamedField, nil

	case "fileLogSort":
		return model1.ParameterTypeFileLogSort, nil

	default:
		return "", errors.New("unknown parameter type")
	}
}

// OverviewPage is the resolver for the overviewPage field.
func (r *Resolver) OverviewPage(ctx context.Context, configIDs []string, destinationIDs []string, period string, telemetryType string) (*model1.OverviewPage, error) {
	graph, err := OverviewGraph(ctx, r.Bindplane, configIDs, destinationIDs, period, telemetryType)
	if err != nil {
		return nil, err
	}

	return &model1.OverviewPage{
		Graph: graph,
	}, nil
}

// Agents is the resolver for the agents field.
func (r *Resolver) Agents(ctx context.Context, selector *string, query *string) (*model1.Agents, error) {
	ctx, span := tracer.Start(ctx, "graphql/Agents")
	defer span.End()

	options, suggestions, err := r.QueryOptionsAndSuggestions(selector, query, r.Bindplane.Store().AgentIndex(ctx))
	if err != nil {
		r.Bindplane.Logger().Error("error getting query options and suggestion", zap.Error(err))
		return nil, err
	}

	agents, err := r.Bindplane.Store().Agents(ctx, options...)
	if err != nil {
		r.Bindplane.Logger().Error("error in graphql Agents", zap.Error(err))
		return nil, err
	}

	return &model1.Agents{
		Agents:        agents,
		Query:         query,
		Suggestions:   suggestions,
		LatestVersion: r.Bindplane.Versions().LatestVersionString(ctx),
	}, nil
}

// Configurations is the resolver for the configurations field.
func (r *Resolver) Configurations(ctx context.Context, selector *string, query *string, onlyDeployedConfigurations *bool) (*model1.Configurations, error) {
	options, suggestions, err := r.QueryOptionsAndSuggestions(selector, query, r.Bindplane.Store().ConfigurationIndex(ctx))
	if err != nil {
		r.Bindplane.Logger().Error("error getting query options and suggestion", zap.Error(err))
		return nil, err
	}

	configurations, err := r.Bindplane.Store().Configurations(ctx, options...)
	if err != nil {
		return nil, err
	}
	// filter out configurations that are not deployed
	if onlyDeployedConfigurations != nil && *onlyDeployedConfigurations {
		filteredConfigurations := []*model.Configuration{}
		for _, configuration := range configurations {
			ids, err := r.Bindplane.Store().AgentsIDsMatchingConfiguration(ctx, configuration)
			if err != nil {
				return nil, err
			}
			if len(ids) > 0 {
				filteredConfigurations = append(filteredConfigurations, configuration)
			}
		}
		configurations = filteredConfigurations
	}
	return &model1.Configurations{
		Configurations: configurations,
		Query:          query,
		Suggestions:    suggestions,
	}, nil
}

// DestinationWithType is the resolver for the destinationWithType field.
func (r *Resolver) DestinationWithType(ctx context.Context, name string) (*model1.DestinationWithType, error) {
	resp := &model1.DestinationWithType{}

	dest, err := r.Bindplane.Store().Destination(ctx, name)
	if err != nil {
		return resp, err
	}

	if dest == nil {
		return resp, nil
	}

	destinationType, err := r.Bindplane.Store().DestinationType(ctx, dest.Spec.Type)
	if err != nil {
		return resp, err
	}

	return &model1.DestinationWithType{
		Destination:     dest,
		DestinationType: destinationType,
	}, nil
}

// Snapshot returns a snapshot of the agent with the specified id and pipeline type
func (r *Resolver) Snapshot(ctx context.Context, agentID string, pipelineType otel.PipelineType, position *string, resourceName *string) (*model1.Snapshot, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	signals := &model1.Snapshot{}
	store := r.Bindplane.Store()

	// get the agent
	agent, err := store.Agent(ctx, agentID)
	if err != nil {
		return signals, err
	}

	// construct a reporting config for this agent
	config, err := store.AgentConfiguration(ctx, agent)
	if err != nil {
		return signals, err
	}
	if config == nil {
		return signals, fmt.Errorf("no configuration available for agent %s", agentID)
	}

	snapshotProcessorName := otel.SnapshotProcessorName
	if position != nil && resourceName != nil {
		snapshotProcessorName = model.SnapshotProcessor(model.MeasurementPosition(*position), *resourceName)
	}

	reportRequest := func(id string) protocol.Report {
		rc := protocol.Report{
			Snapshot: protocol.Snapshot{
				Processor:    string(snapshotProcessorName),
				PipelineType: pipelineType,
				Endpoint: protocol.ReportEndpoint{
					URL: fmt.Sprintf("%s/v1/otlphttp/v1/%s", r.Bindplane.BindPlaneURL(), pipelineType),
					Header: http.Header{
						server.HeaderSessionID: []string{id},
					},
				},
			},
		}
		r.Bindplane.Logger().Info("Requesting report", zap.Any("config", rc))
		return rc
	}

	// all three cases follow the same pattern:
	//
	// 1) receive a channel to await results from relayer
	//
	// 2) send a message to the agent with the specified session id
	//
	// 3) wait for results or timeout

	switch pipelineType {
	case otel.Logs:
		id, result, cancel := r.Bindplane.Relayers().Logs().AwaitResult()
		defer cancel()

		if err := r.Bindplane.Manager().RequestReport(ctx, agentID, reportRequest(id)); err != nil {
			return signals, err
		}

		select {
		case <-ctx.Done():
		case logs := <-result:
			signals.Logs = record.ConvertLogs(logs.OTLP())
		}
	case otel.Metrics:
		id, result, cancel := r.Bindplane.Relayers().Metrics().AwaitResult()
		defer cancel()

		if err := r.Bindplane.Manager().RequestReport(ctx, agentID, reportRequest(id)); err != nil {
			return signals, err
		}

		select {
		case <-ctx.Done():
		case metrics := <-result:
			signals.Metrics = record.ConvertMetrics(ctx, metrics.OTLP())
		}
	case otel.Traces:
		id, result, cancel := r.Bindplane.Relayers().Traces().AwaitResult()
		defer cancel()

		if err := r.Bindplane.Manager().RequestReport(ctx, agentID, reportRequest(id)); err != nil {
			return signals, err
		}

		select {
		case <-ctx.Done():
		case traces := <-result:
			signals.Traces = record.ConvertTraces(traces.OTLP())
		}
	}

	return signals, nil
}

// AgentChanges returns a channel of agent changes
func (r *Resolver) AgentChanges(ctx context.Context, selector *string, query *string) (<-chan []*model1.AgentChange, error) {
	parsedSelector, parsedQuery, err := r.ParseSelectorAndQuery(selector, query)
	if err != nil {
		return nil, err
	}

	// we can ignore the unsubscribe function because this will automatically unsubscribe when the context is done. we
	// could subscribe directly to store.AgentChanges, but the resolver is setup to relay events and the filter and
	// dispatch will happen in a separate goroutine.
	channel, _ := eventbus.SubscribeWithFilter(ctx, r.Updates, func(updates store.BasicEventUpdates) (result []*model1.AgentChange, accept bool) {
		// if the observer is using a selector or query, we want to change Update to Remove if it no longer matches the
		// selector or query
		events := ApplySelectorToChanges(parsedSelector, updates.Agents())
		events = ApplyQueryToChanges(parsedQuery, r.Bindplane.Store().AgentIndex(ctx), events)

		return model1.ToAgentChangeArray(events), !events.Empty()
	})

	return channel, nil
}

// ConfigurationMetrics returns a channel of configuration metrics
func (r *Resolver) ConfigurationMetrics(ctx context.Context, period string, name *string, agent *string) (<-chan *model1.GraphMetrics, error) {
	channel := make(chan *model1.GraphMetrics)

	updateTicker := time.NewTicker(ConfigurationMetricsUpdateInterval)

	sendMetrics := func() {
		if agent != nil && *agent != "" {
			ids := []string{*agent}
			if metrics, err := AgentMetrics(ctx, r.Bindplane, period, ids); err != nil {
				r.Bindplane.Logger().Error("failed to get agentMetrics", zap.Error(err))
			} else {
				channel <- metrics
			}
		} else {
			if metrics, err := ConfigurationMetrics(ctx, r.Bindplane, period, name); err != nil {
				r.Bindplane.Logger().Error("failed to get configurationMetrics", zap.Error(err))
			} else {
				channel <- metrics
			}
		}
	}

	go MetricSubscriber(ctx, sendMetrics, updateTicker)

	return channel, nil
}

// OverviewMetrics returns a channel of overview metrics
func (r *Resolver) OverviewMetrics(ctx context.Context, period string, configIDs []string, destinationIDs []string) (<-chan *model1.GraphMetrics, error) {
	channel := make(chan *model1.GraphMetrics)

	updateTicker := time.NewTicker(OverviewMetricsUpdateInterval)

	sendMetrics := func() {
		if metrics, err := OverviewMetrics(ctx, r.Bindplane, period, configIDs, destinationIDs); err != nil {
			r.Bindplane.Logger().Error("failed to get overviewMetrics", zap.Error(err))
		} else {
			channel <- metrics
		}
	}

	go MetricSubscriber(ctx, sendMetrics, updateTicker)

	return channel, nil
}

// AgentMetrics returns a channel of agent metrics
func (r *Resolver) AgentMetrics(ctx context.Context, period string, ids []string) (<-chan *model1.GraphMetrics, error) {
	channel := make(chan *model1.GraphMetrics)

	updateTicker := time.NewTicker(AgentMetricsUpdateInterval)

	sendMetrics := func() {
		if metrics, err := AgentMetrics(ctx, r.Bindplane, period, ids); err != nil {
			r.Bindplane.Logger().Error("failed to get agentMetrics", zap.Error(err))
		} else {
			channel <- metrics
		}
	}

	go MetricSubscriber(ctx, sendMetrics, updateTicker)

	return channel, nil
}

// ConfigurationChanges returns a channel of configuration changes
func (r *Resolver) ConfigurationChanges(ctx context.Context, selector *string, query *string) (<-chan []*model1.ConfigurationChange, error) {
	parsedSelector, parsedQuery, err := r.ParseSelectorAndQuery(selector, query)
	if err != nil {
		return nil, err
	}

	// we can ignore the unsubscribe function because this will automatically unsubscribe when the context is done.
	channel, _ := eventbus.SubscribeWithFilter(ctx, r.Updates, func(updates store.BasicEventUpdates) (result []*model1.ConfigurationChange, accept bool) {
		// if the observer is using a selector or query, we want to change Update to Remove if it no longer matches the
		// selector or query

		configUpdates := updates.Configurations()
		if r.HasAgentConfigurationChanges(updates) {
			configUpdates = configUpdates.Clone()
			// add all configurations here as updates since we don't know what agent counts could be affected
			if configs, err := r.Bindplane.Store().Configurations(ctx); err == nil {
				for _, config := range configs {
					// don't add configuration pseudo-updates that already have updates associated with them
					if _, ok := configUpdates[config.UniqueKey()]; !ok {
						configUpdates.Include(config, store.EventTypeUpdate)
					}
				}
			} else {
				r.Bindplane.Logger().Error("unable to get configurations to include in agent changes", zap.Error(err))
			}
		}

		events := ApplySelectorToEvents(parsedSelector, configUpdates)
		events = ApplyQueryToEvents(parsedQuery, r.Bindplane.Store().ConfigurationIndex(ctx), events)

		return model1.ToConfigurationChanges(events), len(events) > 0
	})

	return channel, nil
}

// ConfigurationNodeIDResolver assigns the appropriate NodeID to a GraphMetric based on its position,
func ConfigurationNodeIDResolver(_ *record.Metric, position model.MeasurementPosition, _ bpotel.PipelineType, resourceName string) string {
	switch position {
	case model.MeasurementPositionSourceBeforeProcessors:
		return fmt.Sprintf("source/%s", resourceName)
	case model.MeasurementPositionSourceAfterProcessors:
		return fmt.Sprintf("source/%s/processors", resourceName)
	case model.MeasurementPositionDestinationBeforeProcessors:
		return fmt.Sprintf("destination/%s/processors", resourceName)
	case model.MeasurementPositionDestinationAfterProcessors:
		return fmt.Sprintf("destination/%s", resourceName)
	}
	return resourceName
}

func configIsDeployed(ctx context.Context, bindplane exposedserver.BindPlane, configurationName string) bool {
	config, err := bindplane.Store().Configuration(ctx, configurationName)
	if err != nil || config == nil {
		bindplane.Logger().Debug("unable to get configuration", zap.String("configuration", configurationName), zap.Error(err))
		return false
	}
	agentIDs, err := bindplane.Store().AgentsIDsMatchingConfiguration(ctx, config)
	if err != nil {
		bindplane.Logger().Debug("unable to get agent IDs matching configuration", zap.String("configuration", configurationName), zap.Error(err))
		return false
	}
	if len(agentIDs) == 0 {
		return false
	}
	return true
}

func matchSubsequence(query string, target string) bool {
	query = strings.ToLower(query)
	target = strings.ToLower(target)
	j := 0
	for i := 0; i < len(target) && j < len(query); i++ {
		if query[j] == target[i] {
			j++
		}
	}
	return j == len(query)
}

func destinationsInConfigs(ctx context.Context, store store.Store, query *string) ([]*model.Destination, error) {
	// returns only destinations that are in non-raw (managed?) configs and deployed to agents
	configs, err := store.Configurations(ctx)
	if err != nil {
		return nil, err
	}

	if query == nil {
		emptyQuery := ""
		query = &emptyQuery
	}
	// create a map from destination name to destination
	destinationsMap := make(map[string]string)
	destinations := make([]*model.Destination, 0, len(destinationsMap))

	// loop through configs, collect all destinations from configs that aren't raw
	for _, config := range configs {
		if config.Spec.Raw != "" {
			continue
		}
		ids, err := store.AgentsIDsMatchingConfiguration(ctx, config)
		if err != nil {
			return nil, err
		}
		if len(ids) > 0 {
			for _, destination := range config.Spec.Destinations {
				_, ok := destinationsMap[destination.Name]
				if !ok {
					dest, err := store.Destination(ctx, destination.Name)
					if err != nil {
						return destinations, err
					}
					destinationsMap[destination.Name] = "Remember that we've already seen this destination!"
					if matchSubsequence(*query, destination.Name) {
						destinations = append(destinations, dest)
					}
				}
			}
		}
	}

	return destinations, nil
}

// Destinations is the resolver for the destinations field.
func Destinations(ctx context.Context, store store.Store, query *string, filterUnused *bool) ([]*model.Destination, error) {
	if filterUnused != nil && *filterUnused {
		return destinationsInConfigs(ctx, store, query)
	}

	dests, err := store.Destinations(ctx)
	if err != nil {
		return dests, errors.Join(errors.New("queryResolver.Destinations failed to get Destinations from store"), err)
	}
	if query == nil {
		return dests, nil
	}
	destinations := []*model.Destination{}
	for _, dest := range dests {
		if matchSubsequence(*query, dest.Name()) {
			destinations = append(destinations, dest)
		}
	}
	return destinations, nil
}

// OverviewMetrics returns a list of metrics for the overview page
func OverviewMetrics(ctx context.Context, bindplane exposedserver.BindPlane, period string, configIDs []string, destinationIDs []string) (*model1.GraphMetrics, error) {
	if period == "" {
		period = "1m"
	}
	d, err := time.ParseDuration(period)
	if err != nil {
		return nil, fmt.Errorf("failed to parse period %s", period)
	}
	metrics, err := bindplane.Store().Measurements().OverviewMetrics(ctx, stats.WithPeriod(d))
	if err != nil {
		return nil, err
	}

	var maxMetricValue float64
	var maxLogValue float64
	var maxTraceValue float64

	everythingOrSelected := func(resourceKey, resourceType string) string {
		resourcesSelected := []string{}
		switch resourceType {
		case "configuration":
			if configIDs == nil {
				return fmt.Sprintf("%s/%s", resourceType, resourceKey)
			}
			resourcesSelected = configIDs
		case "destination":
			if destinationIDs == nil {
				return fmt.Sprintf("%s/%s", resourceType, resourceKey)
			}
			resourcesSelected = destinationIDs
		}

		inEverything := true
		for _, resourceID := range resourcesSelected {
			if strings.HasSuffix(resourceID, resourceKey) {
				inEverything = false
			}
		}

		if inEverything {
			return fmt.Sprintf("everything/%s", resourceType)
		}
		return fmt.Sprintf("%s/%s", resourceType, resourceKey)
	}

	includeMetric := func(metricMap map[string]*model1.GraphMetric, pipelineType string, nodeID string, metric *record.Metric) {
		// separate metric per pipelineType
		key := fmt.Sprintf("%s_%s", pipelineType, nodeID)
		if cur, ok := metricMap[key]; ok {
			// already exists, include in sum
			if value, ok := stats.Value(metric); ok {
				cur.Value += value
			} else {
				bindplane.Logger().Debug("unable to parse value as float", zap.Any("value", metric.Value))
			}
		} else {
			// doesn't exist, create a metric
			m, err := model1.ToGraphMetric(metric)
			if err != nil {
				bindplane.Logger().Debug("unable to convert record.Metric to GraphMetric", zap.Error(err))
				return
			}
			m.NodeID = nodeID
			metricMap[key] = m
		}
	}

	destinations, err := destinationsInConfigs(ctx, bindplane.Store(), nil)
	if err != nil {
		return nil, errors.Join(errors.New("Failed to get destinations in OverviewMetrics"), err)
	}
	// map of processor (includes type and name) => metric
	destinationMetrics := map[string]*model1.GraphMetric{}
	includeDestination := func(metric *record.Metric, pipelineType, destinationName string) {
		destinationFound := false
		for _, destination := range destinations {
			if destination.Name() == destinationName {
				destinationFound = true
				break
			}
		}
		if !destinationFound {
			return
		}
		// need to eliminate destination metrics that are from undeployed configs

		if !configIsDeployed(ctx, bindplane, stats.Configuration(metric)) {
			return
		}
		nodeID := everythingOrSelected(destinationName, "destination")
		includeMetric(destinationMetrics, pipelineType, nodeID, metric)
	}

	// map of configuration name => metric
	configurationMetrics := map[string]*model1.GraphMetric{}
	includeConfiguration := func(metric *record.Metric, pipelineType string) {
		configurationName := stats.Configuration(metric)
		if !configIsDeployed(ctx, bindplane, configurationName) {
			return
		}
		nodeID := everythingOrSelected(configurationName, "configuration")
		includeMetric(configurationMetrics, pipelineType, nodeID, metric)
	}

	for _, metric := range metrics {
		position, pipelineType, resourceName := stats.ProcessorParsed(metric)
		if position != string(model.MeasurementPositionDestinationAfterProcessors) {
			continue
		}

		splitStrs := strings.Split(resourceName, "-")
		if len(splitStrs) > 1 {
			splitStrs = splitStrs[0 : len(splitStrs)-1]
		}
		resourceName = strings.Join(splitStrs, "-")

		includeDestination(metric, pipelineType, resourceName)
		includeConfiguration(metric, pipelineType)

	}

	var graphMetrics []*model1.GraphMetric

	// add all of the totals for destinations and configurations
	graphMetrics = append(graphMetrics, maps.Values(destinationMetrics)...)
	graphMetrics = append(graphMetrics, maps.Values(configurationMetrics)...)

	// Go through the metrics and find the highest value for each telemetry type
	for _, metric := range graphMetrics {
		switch metric.Name {
		case "metric_data_size":
			if metric.Value > maxMetricValue {
				maxMetricValue = metric.Value
			}
		case "log_data_size":
			if metric.Value > maxLogValue {
				maxLogValue = metric.Value
			}
		case "trace_data_size":
			if metric.Value > maxTraceValue {
				maxTraceValue = metric.Value
			}
		}
	}

	return &model1.GraphMetrics{
		Metrics:        graphMetrics,
		MaxLogValue:    maxLogValue,
		MaxMetricValue: maxMetricValue,
		MaxTraceValue:  maxTraceValue,
	}, nil
}

// ConfigurationMetrics returns a list of metrics for the configuration page
func ConfigurationMetrics(ctx context.Context, bindplane exposedserver.BindPlane, period string, name *string) (*model1.GraphMetrics, error) {
	if period == "" {
		period = "1m"
	}
	configurationName := ""
	if name != nil {
		configurationName = *name
	}
	d, err := time.ParseDuration(period)
	if err != nil {
		return nil, fmt.Errorf("failed to parse period %s with duration %s", period, d)
	}
	metrics, err := bindplane.Store().Measurements().ConfigurationMetrics(ctx, configurationName, stats.WithPeriod(d))
	if err != nil {
		return nil, err
	}

	return assignMetricsToGraph(metrics, ConfigurationNodeIDResolver, bindplane), nil
}

// AgentMetrics returns a list of metrics for the agent page
func AgentMetrics(ctx context.Context, bindplane exposedserver.BindPlane, period string, ids []string) (*model1.GraphMetrics, error) {
	if period == "" {
		period = "1m"
	}

	d, err := time.ParseDuration(period)
	if err != nil {
		return nil, fmt.Errorf("failed to parse period %s with duration %s", period, d)
	}
	if len(ids) == 0 {
		agents, _ := bindplane.Store().Agents(ctx)
		for _, a := range agents {
			ids = append(ids, a.ID)
		}
	}

	var graphMetrics []*model1.GraphMetric

	returnMetrics := &model1.GraphMetrics{
		Metrics: graphMetrics,
	}

	for _, id := range ids {
		metrics, err := bindplane.Store().Measurements().AgentMetrics(ctx, []string{id}, stats.WithPeriod(d))
		if err != nil {
			return nil, err
		}

		singleIDMetrics := assignMetricsToGraph(metrics, ConfigurationNodeIDResolver, bindplane)

		for _, m := range singleIDMetrics.Metrics {
			// This prevents each AgentID being pointed at the same string
			aID := id
			m.AgentID = &aID
		}
		returnMetrics.Metrics = append(returnMetrics.Metrics, singleIDMetrics.Metrics...)
		returnMetrics.MaxMetricValue = singleIDMetrics.MaxMetricValue
		returnMetrics.MaxLogValue = singleIDMetrics.MaxLogValue
		returnMetrics.MaxTraceValue = singleIDMetrics.MaxTraceValue
	}

	return returnMetrics, nil
}

// NodeIDResolver is a function that assigns the appropriate NodeID to a GraphMetric based on its position,
// pipelineType, and resourceName parsed out of the metric processor name. If an empty string is returned, this metric
// will be ignored.
type NodeIDResolver func(metric *record.Metric, position model.MeasurementPosition, pipelineType bpotel.PipelineType, resourceName string) string

func assignMetricsToGraph(metrics []*record.Metric, resolver NodeIDResolver, bindplane exposedserver.BindPlane) *model1.GraphMetrics {
	var graphMetrics []*model1.GraphMetric
	var maxMetricValue float64
	var maxLogValue float64
	var maxTraceValue float64

	for _, m := range metrics {
		graphMetric, err := model1.ToGraphMetric(m)
		if err != nil {
			bindplane.Logger().Debug("unable to convert record.Metric to GraphMetric", zap.Error(err))
			continue
		}

		// figure out what node this is. this must be sync'd with model.Configuration::Graph()
		position, pipelineType, resourceName := stats.ProcessorParsed(m)
		graphMetric.NodeID = resolver(m, model.MeasurementPosition(position), bpotel.PipelineType(pipelineType), resourceName)

		graphMetric.PipelineType = pipelineType
		graphMetrics = append(graphMetrics, graphMetric)

		// keep track of running max value
		switch pipelineType {
		case "metrics":
			if graphMetric.Value > maxMetricValue {
				maxMetricValue = graphMetric.Value
			}
		case "logs":
			if graphMetric.Value > maxLogValue {
				maxLogValue = graphMetric.Value
			}
		case "traces":
			if graphMetric.Value > maxTraceValue {
				maxTraceValue = graphMetric.Value
			}
		}
	}

	return &model1.GraphMetrics{
		Metrics:        graphMetrics,
		MaxMetricValue: maxMetricValue,
		MaxLogValue:    maxLogValue,
		MaxTraceValue:  maxTraceValue,
	}
}

// ConfigurationMetricsUpdateInterval is the interval at which the configuration metrics are updated on the configuration page
const ConfigurationMetricsUpdateInterval = 10 * time.Second

// OverviewMetricsUpdateInterval is the interval at which the overview metrics are updated on the overview page
const OverviewMetricsUpdateInterval = 10 * time.Second

// AgentMetricsUpdateInterval is the interval at which the agent metrics are updated on the agent page
const AgentMetricsUpdateInterval = 10 * time.Second

// MetricSubscriber is a goroutine that sends metrics on a ticker
func MetricSubscriber(ctx context.Context, sendMetrics func(), updateTicker *time.Ticker) {
	defer updateTicker.Stop()

	sendMetrics()
	for {
		select {
		case <-updateTicker.C:
			// tick, send data
			sendMetrics()

		case <-ctx.Done():
			return
		}
	}
}
