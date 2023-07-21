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

package model

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/observiq/bindplane-op/model/graph"
	"github.com/observiq/bindplane-op/model/otel"
	modelSearch "github.com/observiq/bindplane-op/model/search"
	"github.com/observiq/bindplane-op/model/validation"
	"github.com/observiq/bindplane-op/model/version"
	otelExt "go.opentelemetry.io/otel"
	"golang.org/x/exp/maps"

	"gopkg.in/yaml.v3"
)

var tracer = otelExt.Tracer("model/configuration")

type configurationKind struct{}

func (k *configurationKind) NewEmptyResource() *Configuration { return &Configuration{} }

// GetOssOtelHeaders are the Headers used to add headers to the otel configuration
func GetOssOtelHeaders() map[string]string {
	return map[string]string{}
}

// ConfigurationType indicates the kind of configuration. It is based on the presence of the Raw, Sources, and
// Destinations fields.
type ConfigurationType string

const (
	// ConfigurationTypeRaw configurations have a configuration in the Raw field that is passed directly to the agent.
	ConfigurationTypeRaw ConfigurationType = "raw"

	// ConfigurationTypeModular configurations have Sources and Destinations that are used to generate the configuration to pass to an agent.
	ConfigurationTypeModular ConfigurationType = "modular"
	// TODO(andy): Do we like Modular for configurations with Sources/Destinations?
)

// Configuration is the resource for the entire agent configuration
type Configuration struct {
	// ResourceMeta TODO(doc)
	ResourceMeta `yaml:",inline" json:",inline" mapstructure:",squash"`
	// Spec TODO(doc)
	Spec                            ConfigurationSpec `json:"spec" yaml:"spec" mapstructure:"spec"`
	StatusType[ConfigurationStatus] `yaml:",inline" json:",inline" mapstructure:",squash"`
}

var _ HasAgentSelector = (*Configuration)(nil)
var _ HasSensitiveParameters = (*Configuration)(nil)

// GetSpec returns the spec for this resource.
func (c *Configuration) GetSpec() any {
	return c.Spec
}

// NewConfiguration creates a new configuration with the specified name
func NewConfiguration(name string) *Configuration {
	return NewConfigurationWithSpec(name, ConfigurationSpec{})
}

// NewRawConfiguration creates a new configuration with the specified name and raw configuration
func NewRawConfiguration(name string, raw string) *Configuration {
	return NewConfigurationWithSpec(name, ConfigurationSpec{
		Raw: raw,
	})
}

// NewConfigurationWithSpec creates a new configuration with the specified name and spec
func NewConfigurationWithSpec(name string, spec ConfigurationSpec) *Configuration {
	return &Configuration{
		ResourceMeta: ResourceMeta{
			APIVersion: version.V1,
			Kind:       KindConfiguration,
			Metadata: Metadata{
				Name:   name,
				Labels: MakeLabels(),
			},
		},
		Spec: spec,
	}
}

// GetKind returns "Configuration"
func (c *Configuration) GetKind() Kind {
	return KindConfiguration
}

// ConfigurationSpec is the spec for a configuration resource
type ConfigurationSpec struct {
	ContentType  string                  `json:"contentType" yaml:"contentType" mapstructure:"contentType"`
	Raw          string                  `json:"raw,omitempty" yaml:"raw,omitempty" mapstructure:"raw"`
	Sources      []ResourceConfiguration `json:"sources,omitempty" yaml:"sources,omitempty" mapstructure:"sources"`
	Destinations []ResourceConfiguration `json:"destinations,omitempty" yaml:"destinations,omitempty" mapstructure:"destinations"`
	Selector     AgentSelector           `json:"selector" yaml:"selector" mapstructure:"selector"`
}

// ConfigurationStatus is the status for a configuration resource
type ConfigurationStatus struct {
	// Rollout contains status for the rollout of this configuration
	Rollout Rollout `json:"rollout,omitempty" yaml:"rollout,omitempty" mapstructure:"rollout"`

	// CurrentVersion is the version of the configuration that has most recently completed a rollout
	CurrentVersion Version `json:"currentVersion,omitempty" yaml:"currentVersion,omitempty" mapstructure:"currentVersion"`

	// PendingVersion will be set to the version of a rollout that is in progress. It will be set to 0 when the rollout
	// completes.
	PendingVersion Version `json:"pendingVersion,omitempty" yaml:"pendingVersion,omitempty" mapstructure:"pendingVersion"`

	// ----------------------------------------------------------------------
	// transient values set when the configuration is read from the store

	// Latest will be set to true on read if the configuration is the latest version
	Latest bool `json:"latest,omitempty" yaml:"latest,omitempty" mapstructure:"latest"`

	// Pending will be set to true on read if the configuration is the pending version
	Pending bool `json:"pending,omitempty" yaml:"pending,omitempty" mapstructure:"pending"`

	// Current will be set to true on read if the configuration is the current version
	Current bool `json:"current,omitempty" yaml:"current,omitempty" mapstructure:"current"`
}

// RolloutStatus is the status for a configuration rollout

// RolloutStatus is used to track the status of a rollout
type RolloutStatus int

// UnmarshalGQL implements the graphql.Unmarshaler interface.
func (s *RolloutStatus) UnmarshalGQL(i interface{}) error {
	if value, ok := i.(int); ok {
		*s = RolloutStatus(value)
		return nil
	}

	return errors.New("invalid status, must be int")
}

// MarshalGQL implements the graphql.Marshaler interface.
func (s RolloutStatus) MarshalGQL(w io.Writer) {
	bytes := []byte(strconv.Itoa(int(s)))
	_, _ = w.Write(bytes)
}

const (
	// RolloutStatusPending is created, manual start required
	RolloutStatusPending RolloutStatus = 0

	// RolloutStatusStarted is in progress
	RolloutStatusStarted RolloutStatus = 1

	// RolloutStatusPaused is paused by the user
	RolloutStatusPaused RolloutStatus = 2

	// ----------------------------------------------------------------------
	// terminal states

	// RolloutStatusError is a failed rollout because of too many errors
	RolloutStatusError RolloutStatus = 3

	// RolloutStatusStable is a completed rollout saved for labeled agents connecting
	RolloutStatusStable RolloutStatus = 4

	// RolloutStatusReplaced is an incomplete rollout replaced by another rollout
	RolloutStatusReplaced RolloutStatus = 5
)

func (s RolloutStatus) String() string {
	str, ok := rolloutStatusMap[s]
	if !ok {
		return "none"
	}
	return str
}

// TODO make these strings nice
var rolloutStatusMap = map[RolloutStatus]string{
	RolloutStatusPending:  "pending",
	RolloutStatusPaused:   "paused",
	RolloutStatusStarted:  "started",
	RolloutStatusError:    "error",
	RolloutStatusStable:   "stable",
	RolloutStatusReplaced: "replaced",
}

// RolloutOptions are stored with a configuration and determine how rollouts for that configuration are managed.
type RolloutOptions struct {
	// StartAutomatically determines if this rollout transitions immediately from RolloutStatusPending to
	// RolloutStatusStarted without requiring that it be started manually.
	StartAutomatically bool `json:"startAutomatically" yaml:"startAutomatically" mapstructure:"startAutomatically"`

	// RollbackOnFailure determines if the rollout should be rolled back to the previous configuration if the rollout
	// fails.
	RollbackOnFailure bool `json:"rollbackOnFailure" yaml:"rollbackOnFailure" mapstructure:"rollbackOnFailure"`

	// PhaseAgentCount determines the rate at which agents will be updated during a rollout.
	PhaseAgentCount PhaseAgentCount `json:"phaseAgentCount" yaml:"phaseAgentCount" mapstructure:"phaseAgentCount"`

	// MaxErrors is the maximum number of failed agents before the rollout will be considered an error
	MaxErrors int `json:"maxErrors" yaml:"maxErrors" mapstructure:"maxErrors"`
}

// PhaseAgentCount is the number of agents that will be updated in each phase of a rollout.
type PhaseAgentCount struct {
	Initial    int     `json:"initial" yaml:"initial" mapstructure:"initial"`
	Multiplier float64 `json:"multiplier" yaml:"multiplier" mapstructure:"multiplier"`
	Maximum    int     `json:"maximum" yaml:"maximum" mapstructure:"maximum"`
}

// DefaultRolloutOptions contains the default rollout options for a configuration.
// NOTE: These options are the same as the defaults in the UI in rollouts-rest-fns.ts
var DefaultRolloutOptions = RolloutOptions{
	StartAutomatically: false,
	RollbackOnFailure:  true,
	PhaseAgentCount: PhaseAgentCount{
		Initial:    3,
		Multiplier: 5,
		Maximum:    100,
	},
	MaxErrors: 0,
}

// Rollout contains details about the rollout and its progress
type Rollout struct {
	// Name will be set to the Name of the configuration when requested via Configuration.Rollout()
	Name string `json:"name" yaml:"name" mapstructure:"name"`

	// Status is the status of the rollout
	Status RolloutStatus `json:"status" yaml:"status" mapstructure:"status"`

	// Options are set when the Rollout is created based on the rollout options specified in the configuration
	Options RolloutOptions `json:"options" yaml:"options" mapstructure:"options"`

	// Phase starts at zero and increments until all agents are updated. In each phase, initial*multiplier^phase agents will be updated.
	Phase int `json:"phase" yaml:"phase" mapstructure:"phase"`

	// Progress is the current progress of the rollout
	Progress RolloutProgress `json:"progress" yaml:"progress" mapstructure:"progress"`
}

// RolloutProgress is the current progress of the rollout
type RolloutProgress struct {
	// Completed is the number of agents with new version with Connected status
	Completed int `json:"completed" yaml:"completed" mapstructure:"completed"`

	// Errors is the number of agents with new version with Error Status
	Errors int `json:"errors" yaml:"errors" mapstructure:"errors"`

	// Pending is the number of agents that are currently being configured
	Pending int `json:"pending" yaml:"pending" mapstructure:"pending"`

	// Waiting is the number of agents that need to be scheduled for configuration
	Waiting int `json:"waiting" yaml:"waiting" mapstructure:"waiting"`
}

// AgentsPerPhase returns the number of agents that will be updated in the current phase.
func (r *Rollout) AgentsPerPhase() int {
	numAgents := r.Options.PhaseAgentCount.Initial * int(math.Pow(r.Options.PhaseAgentCount.Multiplier, float64(r.Phase)))
	if numAgents > r.Options.PhaseAgentCount.Maximum {
		return r.Options.PhaseAgentCount.Maximum
	}
	return numAgents
}

// UpdateStatus updates the status of the rollout based on the number of completed, errored, pending, and waiting agents.
func (r *Rollout) UpdateStatus(progress RolloutProgress) (newAgentsPending int) {
	r.Progress = progress
	p := &r.Progress

	if p.Errors > r.Options.MaxErrors {
		r.Status = RolloutStatusError
	} else if r.Status == RolloutStatusStarted && p.Waiting == 0 && p.Pending == 0 {
		r.Status = RolloutStatusStable
	}

	newAgentsPending = r.AgentsNextPhase()

	if newAgentsPending > p.Waiting {
		newAgentsPending = p.Waiting
	}
	if newAgentsPending > 0 {
		// optimistically update the status before the agents are actually updated
		p.Waiting -= newAgentsPending
		p.Pending += newAgentsPending
		r.Phase++
	}

	return
}

// AgentsNextPhase returns the number of agents that will be updated in the next phase. If the rollout is not in the
// started state or additional agents are pending, zero is returned.
func (r *Rollout) AgentsNextPhase() int {
	if (r.Status == RolloutStatusStarted || r.Status == RolloutStatusStable) && r.Progress.Pending == 0 {
		return r.AgentsPerPhase()
	}
	return 0
}

// ResourceConfiguration defines Sources and Destinations within a Configuration or Processors within a Source or Destination.
type ResourceConfiguration struct {
	// ID will be generated and is used to uniquely identify the resource
	ID string `json:"id,omitempty" yaml:"id,omitempty" mapstructure:"id"`

	// Name must be specified if this is a reference to another resource by name
	Name string `json:"name,omitempty" yaml:"name,omitempty" mapstructure:"name"`

	// DisplayName is a friendly name of the resource that will be displayed in the UI
	DisplayName string `json:"displayName,omitempty" yaml:"displayName,omitempty" mapstructure:"displayName"`

	// ParameterizedSpec contains the definition of an embedded resource if this is not a reference to another resource
	ParameterizedSpec `yaml:",inline" json:",inline" mapstructure:",squash"`
}

var _ HasResourceParameters = (*ResourceConfiguration)(nil)

// ResourceParameters returns the resource parameters for this resource.
func (rc *ResourceConfiguration) ResourceParameters() []Parameter {
	return rc.Parameters
}

// ensureID ensures that the ID is set to a non-empty value. If the ID is already set, this does nothing. A defaultID
// can be specified to use if the ID is not already set.
func (rc *ResourceConfiguration) ensureID() {
	if rc.ID != "" {
		return
	}
	rc.ID = NewResourceID()
}

// MaskSensitiveParameters masks sensitive parameter values based on the ParameterDefinitions in the ResourceType
func (rc *ResourceConfiguration) maskSensitiveParameters(ctx context.Context) {
	maskSensitiveParameters(ctx, rc)
	for i, p := range rc.Processors {
		p := p
		maskSensitiveParameters(ctx, &p)
		rc.Processors[i] = p
	}
}

// PreserveSensitiveParameters will replace parameters with the SensitiveParameterPlaceholder value with the value of
// the parameter from the existing resource. This does nothing if existing is nil because there is no existing
// resource.
func (rc *ResourceConfiguration) preserveSensitiveParameters(ctx context.Context, existing *ResourceConfiguration) {
	preserveSensitiveParameters(ctx, rc, existing)
	for i, p := range rc.Processors {
		p := p
		existingResource := findResourceConfiguration(p.ID, existing.Processors)
		if existingResource != nil {
			preserveSensitiveParameters(ctx, &p, existingResource)
			rc.Processors[i] = p
		}
	}
}

// UpdateDependencies updates the dependencies for this resource to use the latest version.
func (c *Configuration) UpdateDependencies(ctx context.Context, store ResourceStore) error {
	// update all sources and destinations
	for i, source := range c.Spec.Sources {
		err := source.updateDependencies(ctx, KindSource, store)
		if err != nil {
			return err
		}
		c.Spec.Sources[i] = source
	}
	for i, destination := range c.Spec.Destinations {
		err := destination.updateDependencies(ctx, KindDestination, store)
		if err != nil {
			return err
		}
		c.Spec.Destinations[i] = destination
	}
	return nil
}

// Validate validates most of the configuration, but if a store is available, ValidateWithStore should be used to
// validate the sources and destinations.
func (c *Configuration) Validate() (warnings string, errors error) {
	errs := validation.NewErrors()
	c.validate(errs)
	return errs.Warnings(), errs.Result()
}

func (c *Configuration) validate(errs validation.Errors) {
	c.ResourceMeta.validate(errs)
	c.Spec.validate(errs)
}

// ValidateWithStore checks that the configuration is valid, returning an error if it is not. It uses the store to
// retrieve source types and destination types so that parameter values can be validated against the parameter
// definitions.
func (c *Configuration) ValidateWithStore(ctx context.Context, store ResourceStore) (warnings string, errors error) {
	errs := validation.NewErrors()

	c.validate(errs)
	c.Spec.validateSourcesAndDestinations(ctx, errs, store)

	return errs.Warnings(), errs.Result()
}

// Type returns the ConfigurationType. It is based on the presence of the Raw, Sources, and Destinations fields.
func (c *Configuration) Type() ConfigurationType {
	if c.Spec.Raw != "" {
		// we always prefer raw
		return ConfigurationTypeRaw
	}
	return ConfigurationTypeModular
}

// AgentSelector returns the Selector for this configuration that can be used to match this resource to agents.
func (c *Configuration) AgentSelector() Selector {
	return c.Spec.Selector.Selector()
}

// IsForAgent returns true if this configuration matches a given agent's labels.
func (c *Configuration) IsForAgent(agent *Agent) bool {
	return isResourceForAgent(c, agent)
}

// ResourceStore provides access to resources required to render configurations that use Sources and Destinations.
//
//go:generate mockery --name ResourceStore --inpackage --with-expecter --filename mock_resource_store.go --structname MockResourceStore
type ResourceStore interface {
	Source(ctx context.Context, name string) (*Source, error)
	SourceType(ctx context.Context, name string) (*SourceType, error)
	Processor(ctx context.Context, name string) (*Processor, error)
	ProcessorType(ctx context.Context, name string) (*ProcessorType, error)
	Destination(ctx context.Context, name string) (*Destination, error)
	DestinationType(ctx context.Context, name string) (*DestinationType, error)
}

// Render converts the Configuration model to a configuration yaml that can be sent to an agent. The specified Agent can
// be nil if this configuration is not being rendered for a specific agent.
func (c *Configuration) Render(ctx context.Context, agent *Agent, bindPlaneURL string, bindPlaneInsecureSkipVerify bool, store ResourceStore, headers map[string]string) (string, error) {
	ctx, span := tracer.Start(ctx, "model/Configuration/Render")
	defer span.End()

	if c.Spec.Raw != "" {
		// we always prefer raw
		return c.Spec.Raw, nil
	}
	return c.renderComponents(ctx, agent, bindPlaneURL, bindPlaneInsecureSkipVerify, store, headers)
}

func (c *Configuration) renderComponents(ctx context.Context, agent *Agent, bindPlaneURL string, bindPlaneInsecureSkipVerify bool, store ResourceStore, headers map[string]string) (string, error) {
	configuration, err := c.otelConfiguration(ctx, agent, bindPlaneURL, bindPlaneInsecureSkipVerify, store, headers)
	if err != nil {
		return "", err
	}
	return configuration.YAML()
}

type renderContext struct {
	*otel.RenderContext
	pipelineTypeUsage *PipelineTypeUsage
}

func (c *Configuration) otelConfiguration(ctx context.Context, agent *Agent, bindPlaneURL string, bindPlaneInsecureSkipVerify bool, store ResourceStore, headers map[string]string) (*otel.Configuration, error) {
	if len(c.Spec.Sources) == 0 || len(c.Spec.Destinations) == 0 {
		return nil, nil
	}

	agentID := ""
	agentFeatures := AgentFeaturesDefault
	var measurementsTLS *otel.MeasurementsTLS
	if agent != nil {
		agentID = agent.ID
		agentFeatures = agent.Features()

		// Use agent TLS options if defined
		if agent.TLS != nil {
			measurementsTLS = &otel.MeasurementsTLS{
				InsecureSkipVerify: agent.TLS.InsecureSkipVerify,
				CAFile:             agent.TLS.CAFile,
				CertFile:           agent.TLS.CertFile,
				KeyFile:            agent.TLS.KeyFile,
			}
		}
	}

	rc := &renderContext{
		RenderContext:     otel.NewRenderContext(agentID, c.Name(), bindPlaneURL, bindPlaneInsecureSkipVerify, measurementsTLS),
		pipelineTypeUsage: newPipelineTypeUsage(),
	}
	rc.IncludeSnapshotProcessor = agentFeatures.Has(AgentSupportsSnapshots)
	rc.IncludeMeasurements = agentFeatures.Has(AgentSupportsMeasurements)
	rc.IncludeRouteReceiver = agentFeatures.Has(AgentSupportsLogBasedMetrics)

	return c.otelConfigurationWithRenderContext(ctx, rc, store, headers)
}

func (c *Configuration) otelConfigurationWithRenderContext(ctx context.Context, rc *renderContext, store ResourceStore, headers map[string]string) (*otel.Configuration, error) {
	configuration := otel.NewConfiguration()

	// match each source with each destination to produce a pipeline
	sources, destinations, err := c.evalComponents(ctx, store, rc)
	if err != nil {
		return nil, err
	}

	// to keep configurations consistent, iterate over the sorted keys instead of just iterating over the map directly.
	sourceNames := maps.Keys(sources)
	destinationNames := maps.Keys(destinations)
	sort.Strings(sourceNames)
	sort.Strings(destinationNames)

	for _, sourceName := range sourceNames {
		source := sources[sourceName]
		for _, destinationName := range destinationNames {
			destination := destinations[destinationName]

			name := fmt.Sprintf("%s__%s", sourceName, destinationName)
			configuration.AddPipeline(name, otel.Logs, sourceName, source, destinationName, destination, rc.RenderContext)
			configuration.AddPipeline(name, otel.Metrics, sourceName, source, destinationName, destination, rc.RenderContext)
			configuration.AddPipeline(name, otel.Traces, sourceName, source, destinationName, destination, rc.RenderContext)
		}
	}

	configuration.AddAgentMetricsPipeline(rc.RenderContext, headers)

	return configuration, nil
}

func (c *Configuration) evalComponents(ctx context.Context, store ResourceStore, rc *renderContext) (sources map[string]otel.Partials, destinations map[string]otel.Partials, err error) {
	errorHandler := func(e error) {
		if e != nil {
			err = errors.Join(err, e)
		}
	}

	sources = map[string]otel.Partials{}
	destinations = map[string]otel.Partials{}

	for i, source := range c.Spec.Sources {
		source := source // copy to local variable to securely pass a reference to a loop variable
		sourceName, srcParts := evalSource(ctx, &source, fmt.Sprintf("source%d", i), store, rc, errorHandler)
		sources[sourceName] = srcParts

		// If the route receiver is supported, check if any processor needs it
		if rc.IncludeRouteReceiver {
			for j, p := range source.Processors {
				if processorNeedsRouteReceiver(p) {
					name := fmt.Sprintf("%s__processor%d", sourceName, j)
					routeReceiver, routeParts := createRouteReceiver(ctx, name, errorHandler)
					if routeReceiver != "" {
						addMeasureProcessors(routeParts, MeasurementPositionSourceBeforeProcessors, routeReceiver, rc)
						sources[routeReceiver] = routeParts
					}
				}
			}
		}
	}

	for i, destination := range c.Spec.Destinations {
		destination := destination // copy to local variable to securely pass a reference to a loop variable
		destName, destParts := evalDestination(ctx, i, &destination, fmt.Sprintf("destination%d", i), store, rc, errorHandler)
		destinations[destName] = destParts

		// If the route receiver is supported, check if any processor needs it
		if rc.IncludeRouteReceiver {
			for j, p := range destination.Processors {
				if processorNeedsRouteReceiver(p) {
					name := fmt.Sprintf("%s__processor%d", destName, j)
					routeReceiver, routeParts := createRouteReceiver(ctx, name, errorHandler)
					if routeReceiver != "" {
						addMeasureProcessors(routeParts, MeasurementPositionSourceBeforeProcessors, routeReceiver, rc)
						sources[routeReceiver] = routeParts
					}
				}
			}
		}
	}

	return sources, destinations, err
}

func evalSource(ctx context.Context, source *ResourceConfiguration, defaultName string, store ResourceStore, rc *renderContext, errorHandler TemplateErrorHandler) (string, otel.Partials) {
	src, srcType, err := findSourceAndType(ctx, source, defaultName, store)
	if err != nil {
		errorHandler(err)
		return "", nil
	}

	srcName := src.Name()
	if src.Spec.Disabled {
		return srcName, otel.NewPartials()
	}
	partials := srcType.eval(src, errorHandler)

	if rc.pipelineTypeUsage != nil {
		rc.pipelineTypeUsage.sources.setSupported(srcName, partials)
	}

	addMeasureProcessors(partials, MeasurementPositionSourceBeforeProcessors, src.Name(), rc)
	addSnapshotProcessor(partials, MeasurementPositionSourceBeforeProcessors, src.Name(), rc)

	// evaluate the processors associated with the source
	for i, processor := range source.Processors {
		processor := processor
		_, processorParts := evalProcessor(ctx, &processor, fmt.Sprintf("%s__processor%d", srcName, i), store, rc, errorHandler)
		if processorParts == nil {
			continue
		}
		partials.Append(processorParts)
	}

	addMeasureProcessors(partials, MeasurementPositionSourceAfterProcessors, src.Name(), rc)

	return srcName, partials
}

// EvalProcessor exposes evalProcessor for observathon
func EvalProcessor(ctx context.Context, processor *ResourceConfiguration, defaultName string, store ResourceStore, rc *renderContext, errorHandler TemplateErrorHandler) (string, otel.Partials) {
	return evalProcessor(ctx, processor, defaultName, store, rc, errorHandler)
}

func evalProcessor(ctx context.Context, processor *ResourceConfiguration, defaultName string, store ResourceStore, _ *renderContext, errorHandler TemplateErrorHandler) (string, otel.Partials) {
	prc, prcType, err := findProcessorAndType(ctx, processor, defaultName, store)
	if err != nil {
		errorHandler(err)
		return "", nil
	}

	if prc.Spec.Disabled || processor.Disabled {
		return prc.Name(), otel.NewPartials()
	}

	return prc.Name(), prcType.eval(prc, errorHandler)
}

func evalDestination(ctx context.Context, idx int, destination *ResourceConfiguration, defaultName string, store ResourceStore, rc *renderContext, errorHandler TemplateErrorHandler) (string, otel.Partials) {
	dest, destType, err := findDestinationAndType(ctx, destination, defaultName, store)
	if err != nil {
		errorHandler(err)
		return "", nil
	}

	destName := fmt.Sprintf("%s-%d", dest.Name(), idx)
	if dest.Spec.Disabled || destination.Disabled {
		return destName, otel.NewPartials()
	}
	partials := destType.eval(dest, errorHandler)

	if rc.pipelineTypeUsage != nil {
		rc.pipelineTypeUsage.destinations.setSupported(destName, partials)
	}

	d0partials := otel.NewPartials()
	addMeasureProcessors(d0partials, MeasurementPositionDestinationBeforeProcessors, destName, rc)
	addSnapshotProcessor(d0partials, MeasurementPositionDestinationBeforeProcessors, destName, rc)

	destProcessors := otel.NewPartials()
	// evaluate the processors associated with the destination
	for i, processor := range destination.Processors {
		processor := processor
		_, processorParts := evalProcessor(ctx, &processor, fmt.Sprintf("%s__processor%d", destName, i), store, rc, errorHandler)
		if processorParts == nil {
			continue
		}
		destProcessors.Append(processorParts)
	}

	d1partials := otel.NewPartials()
	addMeasureProcessors(d1partials, MeasurementPositionDestinationAfterProcessors, destName, rc)

	// destination processors are prepended to the destination
	partials.Prepend(d1partials)
	partials.Prepend(destProcessors)
	partials.Prepend(d0partials)

	return destName, partials
}

// createRouteReceiver renders a blank 'route' receiver with a unique name.
// processors can forward telemetry to this receiver by its unique name
// This receiver requires collector version >= 1.14.0
func createRouteReceiver(_ context.Context, name string, errorHandler TemplateErrorHandler) (string, otel.Partials) {
	srcType := SourceType{
		ResourceType: ResourceType{
			Spec: ResourceTypeSpec{
				Version:    "0.0.1",
				Parameters: []ParameterDefinition{},
				Metrics: ResourceTypeOutput{
					Receivers: "- route:",
				},
			},
			ResourceMeta: ResourceMeta{
				APIVersion: version.V1,
				Kind:       KindSourceType,
				Metadata: Metadata{
					Name:        "route",
					DisplayName: "Internal Routing",
					Description: "Used internally for log-based metrics",
					Icon:        "/icons/sources/file.svg",
				},
			},
		},
	}

	src := NewSource(name, srcType.Name(), []Parameter{})
	return src.Name(), srcType.eval(src, errorHandler)
}

func findSourceAndType(ctx context.Context, source *ResourceConfiguration, defaultName string, store ResourceStore) (*Source, *SourceType, error) {
	src, err := FindSource(ctx, source, defaultName, store)
	if err != nil {
		return nil, nil, err
	}
	if source.Name == "" && src.Spec.Type == "" {
		return src, nil, fmt.Errorf("no name or type")
	}

	srcType, err := store.SourceType(ctx, src.Spec.Type)
	if err == nil && srcType == nil {
		err = fmt.Errorf("unknown %s: %s", KindSourceType, src.Spec.Type)
	}
	if err != nil {
		return src, nil, err
	}

	return src, srcType, nil
}

func findProcessorAndType(ctx context.Context, source *ResourceConfiguration, defaultName string, store ResourceStore) (*Processor, *ProcessorType, error) {
	prc, err := FindProcessor(ctx, source, defaultName, store)
	if err != nil {
		return nil, nil, err
	}

	prcType, err := store.ProcessorType(ctx, prc.Spec.Type)
	if err == nil && prcType == nil {
		err = fmt.Errorf("unknown %s: %s", KindProcessorType, prc.Spec.Type)
	}
	if err != nil {
		return prc, nil, err
	}

	return prc, prcType, nil
}

func findDestinationAndType(ctx context.Context, destination *ResourceConfiguration, defaultName string, store ResourceStore) (*Destination, *DestinationType, error) {
	dest, err := FindDestination(ctx, destination, defaultName, store)
	if err != nil {
		return nil, nil, err
	}

	destType, err := store.DestinationType(ctx, dest.Spec.Type)
	if err == nil && destType == nil {
		err = fmt.Errorf("unknown %s: %s", KindDestinationType, dest.Spec.Type)
	}
	if err != nil {
		return dest, nil, err
	}

	return dest, destType, nil
}

func findResourceAndType(ctx context.Context, resourceKind Kind, resource *ResourceConfiguration, defaultName string, store ResourceStore) (Resource, *ResourceType, error) {
	switch resourceKind {
	case KindSource:
		src, srcType, err := findSourceAndType(ctx, resource, defaultName, store)
		if srcType == nil {
			return src, nil, err
		}
		return src, &srcType.ResourceType, err
	case KindProcessor:
		prc, prcType, err := findProcessorAndType(ctx, resource, defaultName, store)
		if prcType == nil {
			return prc, nil, err
		}
		return prc, &prcType.ResourceType, err
	case KindDestination:
		dest, destType, err := findDestinationAndType(ctx, resource, defaultName, store)
		if destType == nil {
			return dest, nil, err
		}
		return dest, &destType.ResourceType, err
	}
	return nil, nil, nil
}

// processorNeedsRouteReceiver checks if the processor needs the route receiver to be in the configuration
func processorNeedsRouteReceiver(processor ResourceConfiguration) bool {
	return strings.HasPrefix(processor.Type, "count_logs") ||
		strings.HasPrefix(processor.Type, "extract_metric") ||
		strings.HasPrefix(processor.Type, "count_telemetry")
}

// ----------------------------------------------------------------------

func (cs *ConfigurationSpec) validate(errors validation.Errors) {
	cs.validateSpecFields(errors)
	cs.validateRaw(errors)
	cs.Selector.validate(errors)
}

func (cs *ConfigurationSpec) validateSpecFields(errors validation.Errors) {
	if cs.Raw != "" {
		if len(cs.Destinations) > 0 || len(cs.Sources) > 0 {
			errors.Add(fmt.Errorf("configuration must specify raw or sources and destinations"))
		}
	}
}

func (cs *ConfigurationSpec) validateRaw(errors validation.Errors) {
	if cs.Raw == "" {
		return
	}
	parsed := map[string]any{}
	err := yaml.Unmarshal([]byte(cs.Raw), parsed)
	if err != nil {
		errors.Add(fmt.Errorf("unable to parse spec.raw as yaml: %w", err))
	}
}

func (cs *ConfigurationSpec) validateSourcesAndDestinations(ctx context.Context, errors validation.Errors, store ResourceStore) {

	for i, source := range cs.Sources {
		source.validate(ctx, KindSource, errors, store)
		// since source may be modified, we need to reassign it
		cs.Sources[i] = source
	}
	for i, destination := range cs.Destinations {
		destination.validate(ctx, KindDestination, errors, store)

		// since destination may be modified, we need to reassign it
		cs.Destinations[i] = destination
	}
}

// ----------------------------------------------------------------------
// ResourceConfiguration

// trimVersions removes the version from the name and type of the resource configuration and any processor name and type.
func (rc *ResourceConfiguration) trimVersions() {
	rc.Name = TrimVersion(rc.Name)
	rc.ParameterizedSpec.Type = TrimVersion(rc.ParameterizedSpec.Type)
	for i, p := range rc.Processors {
		p.trimVersions()
		rc.Processors[i] = p
	}
}

func (rc *ResourceConfiguration) localName(kind Kind, index int) string {
	if rc.Name != "" {
		return rc.Name
	}
	return fmt.Sprintf("%s%d", strings.ToLower(string(kind)), index)
}

func (rc *ResourceConfiguration) validate(ctx context.Context, resourceKind Kind, errors validation.Errors, store ResourceStore) {
	rc.ensureID()
	if rc.validateHasNameOrType(resourceKind, errors) {
		rc.validateParameters(ctx, resourceKind, errors, store)
	}
	rc.validateProcessors(ctx, resourceKind, errors, store)
}

func (rc *ResourceConfiguration) validateHasNameOrType(resourceKind Kind, errors validation.Errors) bool {
	// must have name or type
	if rc.Name == "" && rc.Type == "" {
		errors.Add(fmt.Errorf("all %s must have either a name or type", resourceKind))
		return false
	}
	return true
}

// validateTypeAndParameters is used by Source and Destination validation and uses methods created for Configuration
// validation.
func (rc *ResourceConfiguration) validateTypeAndParameters(ctx context.Context, kind Kind, errors validation.Errors, store ResourceStore) {
	rc.validateParameters(ctx, kind, errors, store)
	rc.validateProcessors(ctx, kind, errors, store)

	// the type may have been modified, so copy it back
	rc.Type = rc.ParameterizedSpec.Type
}

func (rc *ResourceConfiguration) validateParameters(ctx context.Context, resourceKind Kind, errors validation.Errors, store ResourceStore) {
	// must have a name
	for _, parameter := range rc.Parameters {
		if parameter.Name == "" {
			errors.Add(fmt.Errorf("all %s parameters must have a name", resourceKind))
		}
	}

	// trim versions before attempting to find the resource and type
	rc.Name = TrimVersion(rc.Name)
	rc.Type = TrimVersion(rc.Type)

	resource, resourceType, err := findResourceAndType(ctx, resourceKind, rc, string(resourceKind), store)

	if err != nil {
		errors.Add(err)
		return
	}

	// require name or type to have a version
	if rc.Name != "" && resource != nil {
		rc.Name = JoinVersion(rc.Name, resource.Version())
	}
	if rc.Type != "" && resourceType != nil {
		rc.Type = JoinVersion(rc.Type, resourceType.Version())
	}

	// ensure parameters are valid
	for i, parameter := range rc.Parameters {
		if parameter.Name == "" {
			continue
		}
		def := resourceType.Spec.ParameterDefinition(parameter.Name)
		if def == nil {
			errors.Warn(fmt.Errorf("ignoring parameter %s not defined in type %s", parameter.Name, resourceType.Name()))
			continue
		}
		err := def.validateValue(parameter.Value)
		if err != nil {
			errors.Add(err)
		}
		parameter.Sensitive = def.Options.Sensitive
		rc.Parameters[i] = parameter
	}
}

func (rc *ResourceConfiguration) validateProcessors(ctx context.Context, _ Kind, errors validation.Errors, store ResourceStore) {
	for i, processor := range rc.Processors {
		processor.validate(ctx, KindProcessor, errors, store)
		// since processor may be modified, we need to reassign it
		rc.Processors[i] = processor
	}
}

func (rc *ResourceConfiguration) updateDependencies(ctx context.Context, kind Kind, store ResourceStore) error {
	rc.trimVersions()

	// validate to set the latest version
	errs := validation.NewErrors()
	rc.validateTypeAndParameters(ctx, kind, errs, store)
	return errs.Result()
}

// ----------------------------------------------------------------------
// Printable

// PrintableFieldTitles returns the list of field titles, used for printing a table of resources
func (c *Configuration) PrintableFieldTitles() []string {
	return []string{"Name", "Version", "Match"}
}

// PrintableVersionFieldTitles is used when printing the version history
func (c *Configuration) PrintableVersionFieldTitles() []string {
	return []string{"Name", "Version", "Date", "Match", "Current", "Pending", "Rollout"}
}

// PrintableFieldValue returns the field value for a title, used for printing a table of resources
func (c *Configuration) PrintableFieldValue(title string) string {
	switch title {
	case "Name":
		return c.Name()
	case "Match":
		return c.AgentSelector().String()
	case "Current":
		if c.Status.Current {
			return "*"
		}
		return ""
	case "Pending":
		if c.Status.Pending {
			return "*"
		}
		return ""
	case "Rollout":
		return c.Status.Rollout.Status.String()
	default:
		return c.ResourceMeta.PrintableFieldValue(title)
	}
}

// ----------------------------------------------------------------------
// Rollout

// Rollout returns the rollout status for the configuration. It also ensures that the Name field of the Rollout is set
// to match the Configuration.
func (c *Configuration) Rollout() *Rollout {
	r := c.Status.Rollout
	r.Name = c.NameAndVersion()
	return &r
}

// PrintableFieldTitles returns the list of field titles, used for printing a table of resources
func (r *Rollout) PrintableFieldTitles() []string {
	return []string{"Name", "Status", "Phase", "Completed", "Errors", "Pending", "Waiting"}
}

// PrintableFieldValue returns the field value for a title, used for printing a table of resources
func (r *Rollout) PrintableFieldValue(title string) string {
	switch title {
	case "Name":
		return r.Name
	case "Status":
		return r.Status.String()
	case "Phase":
		return fmt.Sprintf("%d", r.Phase)
	case "Completed":
		return fmt.Sprintf("%d", r.Progress.Completed)
	case "Errors":
		return fmt.Sprintf("%d", r.Progress.Errors)
	case "Pending":
		return fmt.Sprintf("%d", r.Progress.Pending)
	case "Waiting":
		return fmt.Sprintf("%d", r.Progress.Waiting)
	default:
		return "-"
	}
}

// PrintableKindSingular returns the singular form of the Kind, e.g. "Configuration"
func (r *Rollout) PrintableKindSingular() string {
	return "Rollout"
}

// PrintableKindPlural returns the plural form of the Kind, e.g. "Configurations"
func (r *Rollout) PrintableKindPlural() string {
	return "Rollouts"
}

// ----------------------------------------------------------------------
// Indexed

// IndexFields returns a map of field name to field value to be stored in the index
func (c *Configuration) IndexFields(index modelSearch.Indexer) {
	c.ResourceMeta.IndexFields(index)

	// add the type of configuration
	index("type", string(c.Type()))

	// add source, sourceType fields
	for _, source := range c.Spec.Sources {
		source.indexFields("source", "sourceType", index)
	}

	// add destination, destinationType fields
	for _, destination := range c.Spec.Destinations {
		destination.indexFields("destination", "destinationType", index)
	}

	// add pipeline fields
	//
	// TODO(andy): I was going to add pipeline:traces, pipeline:logs, and pipeline:metrics because I thought it would be a
	// useful way to filter configurations. However, we need a ResourceStore implementation to call otelConfiguration and
	// we don't have that here, even though indexing is actually done in the store. I think the best solution is to cache
	// the output on the Spec and keep that up to date as any dependent sourceTypes and destinationTypes change. This will
	// improve performance when comparing configurations and displaying the configuration in UI.

	index("rollout-status", c.Rollout().Status.String())

	if c.IsPending() || c.Status.CurrentVersion > 0 {
		index("rollout-pending", c.Name())
	}
}

func (rc *ResourceConfiguration) indexFields(resourceName string, resourceTypeName string, index modelSearch.Indexer) {
	// don't include the resource version in the index
	n, _ := SplitVersion(rc.Name)
	t, _ := SplitVersion(rc.Type)
	index(resourceName, n)
	index(resourceTypeName, t)
}

// Duplicate copies the value of the current configuration and returns
// a duplicate with the new name.  It should be identical except for the
// Metadata.Name, Metadata.ID, and Spec.Selector fields.
func (c *Configuration) Duplicate(name string) *Configuration {
	clone := *c

	// Change the metadata values
	clone.Metadata.Name = name
	clone.Metadata.ID = uuid.NewString()

	// replace the configuration matchLabel
	matchLabels := clone.Spec.Selector.MatchLabels
	matchLabels["configuration"] = name
	return &clone
}

// ----------------------------------------------------------------------
// topology

// Graph returns a graph representing the topology of a configuration
func (c *Configuration) Graph(ctx context.Context, store ResourceStore) (*graph.Graph, error) {
	g := graph.NewGraph()

	// lastNodes is a list of the last node for each source that will be connected to the destinations
	lastNodes := make([]*graph.Node, 0, len(c.Spec.Sources))

	pipelineUsage := c.determinePipelineTypeUsage(ctx, store)
	g.Attributes["activeTypeFlags"] = pipelineUsage.ActiveFlags()

	for i, source := range c.Spec.Sources {
		sourceName := source.localName(KindSource, i)
		usage := pipelineUsage.sources.usage(sourceName)

		attributes := graph.MakeAttributes(string(KindSource), sourceName)
		attributes["activeTypeFlags"] = usage.active
		attributes["supportedTypeFlags"] = usage.supported
		attributes["sourceIndex"] = i
		s := &graph.Node{
			ID:         fmt.Sprintf("source/%s", sourceName),
			Type:       "sourceNode",
			Label:      source.Type,
			Attributes: attributes,
		}
		g.AddSource(s)

		// For now only add one intermediate node for each
		// source which represents all the processors on the source.
		p := &graph.Node{
			ID:    fmt.Sprintf("source/%s/processors", sourceName),
			Type:  "processorNode",
			Label: "Processors",
			Attributes: map[string]any{
				"activeTypeFlags":    usage.active,
				"supportedTypeFlags": usage.supported,
				"sourceIndex":        i,
			},
		}
		g.AddIntermediate(p)
		g.Connect(s, p)

		lastNodes = append(lastNodes, p)
	}

	for i, destination := range c.Spec.Destinations {
		// We don't use the name, because the same destination may be used multiple times.
		// Using the index guarantees uniqueness.
		destinationName := destination.localName(KindDestination, i)
		destinationSlug := fmt.Sprintf("%s-%d", TrimVersion(destinationName), i)
		usage := pipelineUsage.destinations.usage(destinationSlug)

		trimmedName := TrimVersion(destination.Name)

		// For now only add one intermediate node for each
		// destination which represents all the processors on the destination.
		p := &graph.Node{
			ID:    fmt.Sprintf("destination/%s/processors", destinationSlug),
			Type:  "processorNode",
			Label: "Processors",
			Attributes: map[string]any{
				"destinationIndex":   i,
				"activeTypeFlags":    usage.active,
				"supportedTypeFlags": usage.supported,
			},
		}
		g.AddIntermediate(p)
		for _, l := range lastNodes {
			g.Connect(l, p)
		}

		attributes := graph.MakeAttributes(string(KindDestination), trimmedName)
		// Needed to determine which destination card to display.
		attributes["isInline"] = destination.Name == ""
		attributes["activeTypeFlags"] = usage.active
		attributes["supportedTypeFlags"] = usage.supported
		attributes["destinationIndex"] = i

		d := &graph.Node{
			ID:         fmt.Sprintf("destination/%s", destinationSlug),
			Type:       "destinationNode",
			Label:      destination.Name,
			Attributes: attributes,
		}
		g.AddTarget(d)
		g.Connect(p, d)
	}

	return g, nil
}

// MeasurementPosition is a position within the graph of the measurements processor
type MeasurementPosition string

const (
	// MeasurementPositionSourceBeforeProcessors is the initial throughput of the source
	MeasurementPositionSourceBeforeProcessors MeasurementPosition = "s0"

	// MeasurementPositionSourceAfterProcessors is the throughput after source processors
	MeasurementPositionSourceAfterProcessors MeasurementPosition = "s1"

	// MeasurementPositionDestinationBeforeProcessors is the throughput to the destination (from all sources) before
	// destination processors
	MeasurementPositionDestinationBeforeProcessors MeasurementPosition = "d0"

	// MeasurementPositionDestinationAfterProcessors is the throughput to the destination (from all sources) after
	// destination processors
	MeasurementPositionDestinationAfterProcessors MeasurementPosition = "d1"
)

func addMeasureProcessors(partials otel.Partials, position MeasurementPosition, resourceName string, rc *renderContext) {
	if !rc.IncludeMeasurements {
		return
	}
	for _, pipelineType := range []otel.PipelineType{otel.Logs, otel.Metrics, otel.Traces} {
		processorName := otel.ComponentID(fmt.Sprintf("%s/_%s_%s_%s", otel.MeasureProcessorName, position, pipelineType, resourceName))
		addMeasureProcessor(partials[pipelineType], processorName)
	}
}

func addMeasureProcessor(partial *otel.Partial, processorName otel.ComponentID) {
	partial.Processors = append(partial.Processors, otel.ComponentMap{
		processorName: map[string]any{
			"enabled":        true,
			"sampling_ratio": 1,
		},
	})
}

// SnapshotProcessor returns the ComponentID of the snapshot processor for a given position and resource name
func SnapshotProcessor(position MeasurementPosition, resourceName string) otel.ComponentID {
	if resourceName == "" {
		return otel.SnapshotProcessorName
	}
	return otel.ComponentID(fmt.Sprintf("%s/_%s_%s", otel.SnapshotProcessorName, position, resourceName))
}

func addSnapshotProcessor(partials otel.Partials, position MeasurementPosition, resourceName string, rc *renderContext) {
	if !rc.IncludeSnapshotProcessor {
		return
	}
	for _, pipelineType := range []otel.PipelineType{otel.Logs, otel.Metrics, otel.Traces} {
		processorName := SnapshotProcessor(position, resourceName)
		partial := partials[pipelineType]
		partial.Processors = append(partial.Processors, otel.ComponentMap{processorName: nil})
	}
}

type pipelineUsage struct {
	active    otel.PipelineTypeFlags
	supported otel.PipelineTypeFlags
}

type pipelineTypeUsageMap map[string]*pipelineUsage

func (p pipelineTypeUsageMap) usage(name string) *pipelineUsage {
	splitName, _ := SplitVersion(name)
	if u, ok := p[splitName]; ok {
		return u
	}
	result := &pipelineUsage{}
	p[splitName] = result
	return result
}

func (p pipelineTypeUsageMap) setActive(name string, pipelineType otel.PipelineType) {
	p.usage(name).active.Set(pipelineType.Flag())
}

func (p pipelineTypeUsageMap) setSupported(name string, partials otel.Partials) {
	p.usage(name).supported.Set(partials.PipelineTypes())
}

// PipelineTypeUsage contains information about active telemetry on the Configuration
// and its sources and destinations.
type PipelineTypeUsage struct {
	sources      pipelineTypeUsageMap
	destinations pipelineTypeUsageMap

	// active refers to the top level configuration
	active otel.PipelineTypeFlags
}

// ActiveFlagsForDestination returns the PipelineTypeFlags that are active for a destination with given name.
func (p *PipelineTypeUsage) ActiveFlagsForDestination(name string) otel.PipelineTypeFlags {
	return p.destinations.usage(name).active
}

// ActiveFlags returns the pipeline type flags that are in use by the configuration.
func (p *PipelineTypeUsage) ActiveFlags() otel.PipelineTypeFlags {
	return p.active
}

// setActive sets the top level activeFlags for a pipelineTypeUsage
func (p *PipelineTypeUsage) setActive(t otel.PipelineType) {
	p.active.Set(t.Flag())
}

func newPipelineTypeUsage() *PipelineTypeUsage {
	return &PipelineTypeUsage{
		sources:      pipelineTypeUsageMap{},
		destinations: pipelineTypeUsageMap{},
	}
}

func (c *Configuration) determinePipelineTypeUsage(ctx context.Context, store ResourceStore) *PipelineTypeUsage {
	p := newPipelineTypeUsage()

	// the agent ID, URL, and tls values aren't important
	rc := &renderContext{
		RenderContext:     otel.NewRenderContext("AGENT_ID", c.Name(), "BINDPLANE_URL", false, nil),
		pipelineTypeUsage: p,
	}
	config, err := c.otelConfigurationWithRenderContext(ctx, rc, store, GetOssOtelHeaders())
	if err != nil {
		// pipeline type usage won't be available if there is an error rendering the configuration, but that's ok.
		return p
	}

	for _, pipeline := range config.Service.Pipelines {
		p.sources.setActive(pipeline.SourceName(), pipeline.Type())
		p.destinations.setActive(pipeline.DestinationName(), pipeline.Type())

		p.setActive(pipeline.Type())
	}

	return p
}

// Usage returns a PipelineTypeUsage struct which contains information about the active and
// supported telemetry types on a configuration.
func (c *Configuration) Usage(ctx context.Context, store ResourceStore) *PipelineTypeUsage {
	return c.determinePipelineTypeUsage(ctx, store)
}

// ----------------------------------------------------------------------

// MaskSensitiveParameters masks sensitive parameter values based on the ParameterDefinitions in the ResourceType
func (c *Configuration) MaskSensitiveParameters(ctx context.Context) {
	// masking in configuration is more complicated because we need to mask parameters in the source and destination
	for i, source := range c.Spec.Sources {
		source.maskSensitiveParameters(ctx)
		c.Spec.Sources[i] = source
	}
	for i, destination := range c.Spec.Destinations {
		destination.maskSensitiveParameters(ctx)
		c.Spec.Destinations[i] = destination
	}
}

// PreserveSensitiveParameters will replace parameters with the SensitiveParameterPlaceholder value with the value of
// the parameter from the existing resource. This does nothing if existing is nil because there is no existing
// resource.
func (c *Configuration) PreserveSensitiveParameters(ctx context.Context, existing *AnyResource) error {
	existingConfiguration, err := ParseOne[*Configuration](existing)
	if err != nil {
		return err
	}

	for i, source := range c.Spec.Sources {
		existingResource := findResourceConfiguration(source.ID, existingConfiguration.Spec.Sources)
		if existingResource != nil {
			source.preserveSensitiveParameters(ctx, existingResource)
			c.Spec.Sources[i] = source
		}
	}
	for i, destination := range c.Spec.Destinations {
		existingResource := findResourceConfiguration(destination.ID, existingConfiguration.Spec.Destinations)
		if existingResource != nil {
			destination.preserveSensitiveParameters(ctx, existingResource)
			c.Spec.Destinations[i] = destination
		}
	}

	return nil
}

// findResourceConfiguration looks through all of the resources for a resource matching the specified resourceID.
func findResourceConfiguration(resourceID string, resources []ResourceConfiguration) *ResourceConfiguration {
	for _, resource := range resources {
		if resource.ID == resourceID {
			return &resource
		}
	}
	return nil
}
