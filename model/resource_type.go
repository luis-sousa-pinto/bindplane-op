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

package model

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/observiq/bindplane-op/model/otel"
	"github.com/observiq/bindplane-op/model/validation"
	"github.com/observiq/bindplane-op/resources/helpers"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
)

// ResourceType is a resource that describes a type of resource including parameters for creating that resource and a
// template for formatting the resource configuration.
//
// There are separate ResourceTypes for each type of resource, e.g. SourceType for Source resources.
type ResourceType struct {
	ResourceMeta              `yaml:",inline" mapstructure:",squash"`
	Spec                      ResourceTypeSpec `json:"spec" yaml:"spec" mapstructure:"spec"`
	StatusType[VersionStatus] `yaml:",inline" mapstructure:",squash"`
}

// GetSpec returns the spec for this resource.
func (rt *ResourceType) GetSpec() any {
	return rt.Spec
}

// ResourceTypeSpec is the spec for a resourceType to
type ResourceTypeSpec struct {
	Version string `json:"version,omitempty" yaml:"version,omitempty" mapstructure:"version"`

	// Parameters currently uses the model from stanza. Eventually we will probably create a separate definition for
	// BindPlane.
	Parameters         []ParameterDefinition `json:"parameters"  yaml:"parameters"  mapstructure:"parameters"`
	SupportedPlatforms []string              `json:"supportedPlatforms,omitempty" yaml:"supportedPlatforms,omitempty" mapstructure:"supportedPlatforms"`

	// individual
	Logs    ResourceTypeOutput `json:"logs,omitempty"    yaml:"logs,omitempty"    mapstructure:"logs"`
	Metrics ResourceTypeOutput `json:"metrics,omitempty" yaml:"metrics,omitempty" mapstructure:"metrics"`
	Traces  ResourceTypeOutput `json:"traces,omitempty"  yaml:"traces,omitempty"  mapstructure:"traces"`

	// pairs (alphabetical order)
	LogsMetrics   ResourceTypeOutput `json:"logs+metrics,omitempty"   yaml:"logs+metrics,omitempty"   mapstructure:"logs+metrics"`
	LogsTraces    ResourceTypeOutput `json:"logs+traces,omitempty"    yaml:"logs+traces,omitempty"    mapstructure:"logs+traces"`
	MetricsTraces ResourceTypeOutput `json:"metrics+traces,omitempty" yaml:"metrics+traces,omitempty" mapstructure:"metrics+traces"`

	// all three (alphabetical order)
	LogsMetricsTraces ResourceTypeOutput `json:"logs+metrics+traces,omitempty" yaml:"logs+metrics+traces,omitempty" mapstructure:"logs+metrics+traces"`

	// FeatureGate is a string that is used to gate the availability of this resource type.
	FeatureGate string `json:"featureGate,omitempty" yaml:"featureGate,omitempty" mapstructure:"featureGate"`
}

// ResourceTypeOutput describes the output of the resource type
type ResourceTypeOutput struct {
	Receivers  ResourceTypeTemplate `json:"receivers,omitempty"  yaml:"receivers,omitempty"  mapstructure:"receivers"`
	Processors ResourceTypeTemplate `json:"processors,omitempty" yaml:"processors,omitempty" mapstructure:"processors"`
	Exporters  ResourceTypeTemplate `json:"exporters,omitempty"  yaml:"exporters,omitempty"  mapstructure:"exporters"`
	Extensions ResourceTypeTemplate `json:"extensions,omitempty" yaml:"extensions,omitempty" mapstructure:"extensions"`
}

// Empty returns true if Receivers, Processors, Exporters, and Extensions are the zero value ""
func (s ResourceTypeOutput) Empty() bool {
	return s.Receivers == "" && s.Processors == "" && s.Exporters == "" && s.Extensions == ""
}

// ResourceTypeTemplate is a go-template that evaluates to an array of OpenTelemetry resources
type ResourceTypeTemplate string

// TemplateErrorHandler handles errors during template evaluation. Typically these will be logged but they could be
// accumulated and reported to the user.
type TemplateErrorHandler func(error)

// ParameterDefinition returns the ParameterDefinition with the specified name or nil if no such parameter exists
func (s *ResourceTypeSpec) ParameterDefinition(name string) *ParameterDefinition {
	for _, p := range s.Parameters {
		if name == p.Name {
			return &p
		}
	}
	return nil
}

// ParameterDefinitionWithLabel returns the first ParameterDefinition with the specified label or nil if no such parameter exists
func (s *ResourceTypeSpec) ParameterDefinitionWithLabel(label string) *ParameterDefinition {
	for _, p := range s.Parameters {
		if label == p.Label {
			return &p
		}
	}
	return nil
}

// ----------------------------------------------------------------------

// eval executes all of the templates associated with this resource type, returning a partial configuration for each
// telemetry type.
func (rt *ResourceType) eval(resource parameterizedResource, errorHandler TemplateErrorHandler) otel.Partials {
	result := otel.Partials{
		otel.Logs:    rt.evalOutput(&rt.Spec.Logs, resource, errorHandler),
		otel.Metrics: rt.evalOutput(&rt.Spec.Metrics, resource, errorHandler),
		otel.Traces:  rt.evalOutput(&rt.Spec.Traces, resource, errorHandler),
	}

	// add multi-pipelines components
	logsMetrics := rt.evalOutput(&rt.Spec.LogsMetrics, resource, errorHandler)
	result[otel.Logs].Append(logsMetrics)
	result[otel.Metrics].Append(logsMetrics)

	logsTraces := rt.evalOutput(&rt.Spec.LogsTraces, resource, errorHandler)
	result[otel.Logs].Append(logsTraces)
	result[otel.Traces].Append(logsTraces)

	metricsTraces := rt.evalOutput(&rt.Spec.MetricsTraces, resource, errorHandler)
	result[otel.Metrics].Append(metricsTraces)
	result[otel.Traces].Append(metricsTraces)

	logsMetricsTraces := rt.evalOutput(&rt.Spec.LogsMetricsTraces, resource, errorHandler)
	result[otel.Logs].Append(logsMetricsTraces)
	result[otel.Metrics].Append(logsMetricsTraces)
	result[otel.Traces].Append(logsMetricsTraces)

	return result
}

// evalOutput executes the templates associated with the specified output using the specified resource and errorHandler.
func (rt *ResourceType) evalOutput(output *ResourceTypeOutput, resource parameterizedResource, errorHandler TemplateErrorHandler) *otel.Partial {
	params := map[string]any{}
	// start with default parameters
	for _, p := range rt.Spec.Parameters {
		if p.Default != nil {
			params[p.Name] = p.Default
		}
	}
	// resource can overrides the parameters
	for _, p := range resource.ResourceParameters() {
		params[p.Name] = p.Value
	}
	// eval all of the components
	return &otel.Partial{
		Receivers:  rt.evalTemplate(output.Receivers, resource, params, errorHandler),
		Processors: rt.evalTemplate(output.Processors, resource, params, errorHandler),
		Exporters:  rt.evalTemplate(output.Exporters, resource, params, errorHandler),
		Extensions: rt.evalTemplate(output.Extensions, resource, params, errorHandler),
	}
}

const (
	templateFuncHasCategoryMetricsEnabled      = "bpHasCategoryMetricsEnabled"
	templateFuncDisabledCategoryMetrics        = "bpDisabledCategoryMetrics"
	templateFuncComponentID                    = "bpComponentID"
	templateFuncRouteID                        = "bpRouteID"
	templateFuncDefaultDisabledCategoryMetrics = "bpDefaultDisabledCategoryMetrics"
)

func (rt *ResourceType) templateFuncMap(nameProvider otel.ComponentIDProvider) template.FuncMap {
	return template.FuncMap{
		templateFuncHasCategoryMetricsEnabled:      rt.templateFuncHasCategoryMetricsEnabled,
		templateFuncDisabledCategoryMetrics:        rt.templateFuncDisabledCategoryMetrics,
		templateFuncComponentID:                    rt.templateFuncComponentID(nameProvider),
		templateFuncRouteID:                        rt.templateFuncRouteID(nameProvider),
		templateFuncDefaultDisabledCategoryMetrics: rt.templateFuncDefaultDisabledCategoryMetrics,
	}
}

func (rt *ResourceType) templateFuncComponentID(nameProvider otel.ComponentIDProvider) func(componentName string) (string, error) {
	return func(componentName string) (string, error) {
		return string(nameProvider.ComponentID(componentName)), nil
	}
}
func (rt *ResourceType) templateFuncRouteID(nameProvider otel.ComponentIDProvider) func() (string, error) {
	return func() (string, error) {
		componentID, err := rt.templateFuncComponentID(nameProvider)("route")
		if err != nil {
			return "", err
		}
		return strings.Replace(componentID, "route/", "", 1), nil
	}
}

func (rt *ResourceType) templateFuncHasCategoryMetricsEnabled(parameterValue []any, parameterName, metricCategory string) (bool, error) {
	parameterDefinition := rt.Spec.ParameterDefinition(parameterName)
	if parameterDefinition == nil {
		return false, fmt.Errorf("unknown parameter name %s", parameterName)
	}

	if parameterDefinition.Type != metricsType {
		return false, fmt.Errorf("parameter name %s is not a metrics type", parameterName)
	}

	metricNames := parameterDefinition.metricNames(metricCategory)
	return slices.IndexFunc(metricNames, func(metricName string) bool {
		for _, val := range parameterValue {
			if val == metricName {
				// disabled
				return false
			}
		}
		// not disabled
		return true
	}) >= 0, nil
}

func (rt *ResourceType) templateFuncDisabledCategoryMetrics(parameterValue []any, parameterName, metricCategory string) ([]string, error) {
	parameterDefinition := rt.Spec.ParameterDefinition(parameterName)
	if parameterDefinition == nil {
		return nil, fmt.Errorf("unknown parameter name %s", parameterName)
	}

	if parameterDefinition.Type != metricsType {
		return nil, fmt.Errorf("parameter name %s is not a metrics type", parameterName)
	}

	metricNames := parameterDefinition.metricNames(metricCategory)

	var result []string

	for _, name := range metricNames {
		for _, val := range parameterValue {
			if name == val {
				result = append(result, name)
			}
		}
	}

	return result, nil
}

// templateFuncDefaultDisabledCategoryMetrics returns metrics that are not filtered out but are not enabled by default
func (rt *ResourceType) templateFuncDefaultDisabledCategoryMetrics(parameterValue []any, parameterName, metricCategory string) ([]string, error) {
	parameterDefinition := rt.Spec.ParameterDefinition(parameterName)
	if parameterDefinition == nil {
		return nil, fmt.Errorf("unknown parameter name %s", parameterName)
	}

	if parameterDefinition.Type != metricsType {
		return nil, fmt.Errorf("parameter name %s is not a metrics type", parameterName)
	}

	// Create a lookup to tell if a metric is in the filtered list
	parameterLookup := make(map[string]struct{}, len(parameterValue))
	for _, val := range parameterValue {
		stringVal, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("parameter value %v is not a string", val)
		}
		parameterLookup[stringVal] = struct{}{}
	}

	// Retrieve metrics for the category
	metrics := parameterDefinition.metrics(metricCategory)

	var result []string
	for _, metric := range metrics {
		// If the metric is not in the filtered list and is disabled by default, add it to the result
		if _, ok := parameterLookup[metric.Name]; !ok && metric.DefaultDisabled {
			result = append(result, metric.Name)
		}
	}

	return result, nil
}

// evalTemplate evaluates a single template with the specified paramValues. nameProvider is available to make the name
// unique and the errorHandler will accumulate errors so that they can be reported once.
func (rt *ResourceType) evalTemplate(r ResourceTypeTemplate, nameProvider otel.ComponentIDProvider, paramValues map[string]any, errorHandler TemplateErrorHandler) otel.ComponentList {
	set := otel.ComponentList{}

	// get the template for the key
	t, err := template.New(rt.Name()).
		Option("missingkey=error").
		Funcs(template.FuncMap(sprig.FuncMap())).
		Funcs(rt.templateFuncMap(nameProvider)).
		Funcs(template.FuncMap(helpers.ResourceHelperFuncMap())).
		Parse(string(r))
	if err != nil {
		errorHandler(err)
		return set
	}

	// render the template
	var writer bytes.Buffer
	if err := t.Execute(&writer, paramValues); err != nil {
		errorHandler(err)
		return set
	}

	bytes := writer.Bytes()

	// parse as yaml so that we can combine yaml fragments and render
	var parsed []map[string]any
	if err := yaml.Unmarshal(bytes, &parsed); err != nil {
		errorHandler(err)
		return set
	}

	// assemble all of the blocks after renaming them
	for _, block := range parsed {
		for key, value := range block {
			componentID := nameProvider.ComponentID(key)
			set = append(set, map[otel.ComponentID]any{
				componentID: value,
			})
		}
	}

	return set
}

// ----------------------------------------------------------------------
// featureGate
// ----------------------------------------------------------------------

// FeatureGate returns the feature flag for this resource type
func (rt *ResourceType) FeatureGate() string {
	return rt.Spec.FeatureGate
}

// SetFeatureGate sets the feature flag for this resource type
func (rt *ResourceType) SetFeatureGate(featureGate string) {
	rt.Spec.FeatureGate = featureGate
}

// ----------------------------------------------------------------------

// PrintableFieldTitles returns the list of field titles, used for printing a table of resources
func (rt *ResourceType) PrintableFieldTitles() []string {
	return []string{"Name", "Display", "Version"}
}

// ----------------------------------------------------------------------
// validation

// Validate returns an error if any part of the ResourceType is invalid
func (rt *ResourceType) Validate() (warnings string, errors error) {
	errs := validation.NewErrors()

	rt.ResourceMeta.validate(errs)
	rt.Spec.validate(rt.Kind, errs)
	if rt.Kind == KindSourceType {
		rt.validateRequiredSourceTypeTLS(errs)
	}
	if rt.Kind == KindDestinationType {
		rt.validateRequiredDestinationTypeTLS(errs)
	}

	return errs.Warnings(), errs.Result()
}

// ValidateWithStore returns an error if any part of the ResourceType is invalid
func (rt *ResourceType) ValidateWithStore(_ context.Context, _ ResourceStore) (warnings string, errors error) {
	return rt.Validate()
}

var errMissingTLSParam = errors.New("tls parameter missing from resource type")

func (rt *ResourceType) validateRequiredSourceTypeTLS(errs validation.Errors) {
	if exempt, ok := sourceTypeTLSExemptions[rt.Metadata.Name]; ok && exempt {
		return
	}

	if p := rt.Spec.ParameterDefinitionWithLabel("Enable TLS"); p == nil {
		errs.Warn(fmt.Errorf("%w: %s: %s", errMissingTLSParam, rt.Metadata.Name, "Enable TLS"))
	}
	if p := rt.Spec.ParameterDefinitionWithLabel("TLS Certificate Authority File"); p == nil {
		errs.Warn(fmt.Errorf("%w: %s: %s", errMissingTLSParam, rt.Metadata.Name, "TLS Certificate Authority File"))
	}
	if p := rt.Spec.ParameterDefinitionWithLabel("Skip TLS Certificate Verification"); p == nil {
		// Some existing SourceTypes invert the `insecure_skip_verify` parameter, that's allowed
		if p := rt.Spec.ParameterDefinitionWithLabel("Strict TLS Certificate Verification"); p == nil {
			errs.Warn(fmt.Errorf("%w: %s: %s", errMissingTLSParam, rt.Metadata.Name, "Skip TLS Certificate Verification"))
		}
	}
	if p := rt.Spec.ParameterDefinitionWithLabel("TLS Client Certificate File"); p == nil {
		errs.Warn(fmt.Errorf("%w: %s: %s", errMissingTLSParam, rt.Metadata.Name, "TLS Client Certificate File"))
	}
	if p := rt.Spec.ParameterDefinitionWithLabel("TLS Client Private Key File"); p == nil {
		errs.Warn(fmt.Errorf("%w: %s: %s", errMissingTLSParam, rt.Metadata.Name, "TLS Client Private Key File"))
	}
}

func (rt *ResourceType) validateRequiredDestinationTypeTLS(errs validation.Errors) {
	if exempt, ok := destinationTypeTLSExemptions[rt.Metadata.Name]; ok && exempt {
		return
	}

	if p := rt.Spec.ParameterDefinitionWithLabel("Enable TLS"); p == nil {
		errs.Warn(fmt.Errorf("%w: %s: %s", errMissingTLSParam, rt.Metadata.Name, "Enable TLS"))
	}
	if p := rt.Spec.ParameterDefinitionWithLabel("TLS Certificate Authority File"); p == nil {
		errs.Warn(fmt.Errorf("%w: %s: %s", errMissingTLSParam, rt.Metadata.Name, "TLS Certificate Authority File"))
	}
	if p := rt.Spec.ParameterDefinitionWithLabel("Skip TLS Certificate Verification"); p == nil {
		// Some existing DestinationTypes invert the `insecure_skip_verify` parameter, that's allowed
		if p := rt.Spec.ParameterDefinitionWithLabel("Strict TLS Certificate Verification"); p == nil {
			errs.Warn(fmt.Errorf("%w: %s: %s", errMissingTLSParam, rt.Metadata.Name, "Skip TLS Certificate Verification"))
		}
	}
}

func (s *ResourceTypeSpec) validate(kind Kind, errs validation.Errors) {
	s.validateSupportedPlatforms(kind, errs)
	s.validateParameterDefinitions(kind, errs)

	// assemble default parameter values for validation
	params := map[string]any{}
	for _, p := range s.Parameters {
		if p.Default != nil {
			params[p.Name] = p.Default
		} else {
			// for template validation, just provide a reasonable default based on the type
			switch p.Type {
			case boolType:
				params[p.Name] = false
			case enumType:
				params[p.Name] = "" // p.ValidValues[0] // cannot guarantee this is valid and "" is fine
			case enumsType:
				params[p.Name] = []string{}
			case intType:
				params[p.Name] = 0
			case mapType:
				params[p.Name] = make(map[string]string)
			case stringType:
				params[p.Name] = ""
			case stringsType:
				params[p.Name] = []string{}
			case yamlType:
				params[p.Name] = ""
			case mapToEnumType:
				params[p.Name] = make(map[string]string)
			}
		}
	}

	s.Logs.validateTemplates(errs, "logs", params)
	s.Metrics.validateTemplates(errs, "metrics", params)
	s.Traces.validateTemplates(errs, "traces", params)
}

const (
	platformWindows             = "windows"
	platformLinux               = "linux"
	platformMacOS               = "macos"
	platformK8sDaemonset        = "kubernetes-daemonset"
	platformK8sDeployment       = "kubernetes-deployment"
	platformOpenshiftDaemonset  = "openshift-daemonset"
	platformOpenshiftDeployment = "openshift-deployment"
)

type supportedPlatforms struct {
	platforms []string
}

// SupportedPlatforms contains the list of supported platforms
// in the ResourceType Spec
var SupportedPlatforms = &supportedPlatforms{
	platforms: []string{
		platformWindows,
		platformLinux,
		platformMacOS,
		platformK8sDaemonset,
		platformK8sDeployment,
		platformOpenshiftDaemonset,
		platformOpenshiftDeployment,
	},
}

func (s *supportedPlatforms) contains(platform string) bool {
	for _, p := range s.platforms {
		if p == platform {
			return true
		}
	}
	return false
}

// ErrMissingSupportedPlatforms is returned when the supportedPlatforms field is missing
// or empty
var ErrMissingSupportedPlatforms = errors.New("supportedPlatforms must be specified")

// ErrInvalidPlatform is returned when a platform name is not valid
var ErrInvalidPlatform = errors.New("invalid platform name")

func (s *ResourceTypeSpec) validateSupportedPlatforms(kind Kind, errs validation.Errors) {
	// Only run for SourceTypes
	if kind != KindSourceType {
		return
	}

	if s.SupportedPlatforms == nil || len(s.SupportedPlatforms) == 0 {
		errs.Add(ErrMissingSupportedPlatforms)
	}

	for _, platform := range s.SupportedPlatforms {
		if !SupportedPlatforms.contains(platform) {
			errs.Add(ErrInvalidPlatform)
		}
	}
}

func (s *ResourceTypeSpec) validateParameterDefinitions(kind Kind, errs validation.Errors) {
	for _, parameter := range s.Parameters {
		parameter.validateDefinition(kind, errs)
		s.validateParameterRelevantIf(parameter, errs)

		// indicate that password is deprecated
		if parameter.Options.Password {
			errs.Warn(fmt.Errorf("parameter '%s' uses deprecated 'password' option, use 'sensitive' instead", parameter.Name))
		}
		if parameter.Options.Sensitive && parameter.Options.Multiline {
			errs.Warn(fmt.Errorf("parameter '%s' cannot be 'sensitive' and 'multiline', ignoring 'multiline'", parameter.Name))
		}
	}

	s.validateNoDuplicateParameterNames(errs)
}

// validateParameterRelevantIf in ResourceTypeSpec because we need to check against other parameter names
func (s *ResourceTypeSpec) validateParameterRelevantIf(parameter ParameterDefinition, errs validation.Errors) {
	for _, relevantIf := range parameter.RelevantIf {
		if relevantIf.Name == "" {
			errs.Add(fmt.Errorf("relevantIf for '%s' must have a name", parameter.Name))
			continue
		}
		ref := s.ParameterDefinition(relevantIf.Name)
		if ref == nil {
			errs.Add(fmt.Errorf("relevantIf for '%s' refers to nonexistant parameter '%s'", parameter.Name, relevantIf.Name))
			continue
		}
		if relevantIf.Operator == "" {
			errs.Add(fmt.Errorf("relevantIf '%s' for '%s' must have an operator", ref.Name, parameter.Name))
		}
		if relevantIf.Value == nil {
			errs.Add(fmt.Errorf("relevantIf '%s' for '%s' must have a value", ref.Name, parameter.Name))
			continue
		}
		err := ref.validateValueType(parameterFieldRelevantIf, relevantIf.Value)
		if err != nil {
			errs.Add(fmt.Errorf("relevantIf '%s' for '%s': %w", ref.Name, parameter.Name, err))
		}
	}
}

func (s ResourceTypeOutput) validateTemplates(errs validation.Errors, name string, params map[string]any) {
	s.Receivers.validate(errs, fmt.Sprintf("%s.receivers", name), params)
	s.Processors.validate(errs, fmt.Sprintf("%s.processors", name), params)
	s.Exporters.validate(errs, fmt.Sprintf("%s.exporters", name), params)
	s.Extensions.validate(errs, fmt.Sprintf("%s.extensions", name), params)
}

func (s ResourceTypeTemplate) validate(errs validation.Errors, name string, params map[string]any) {
	if s == "" {
		// no validation for empty templates
		return
	}
	// ensure the template is valid
	t, err := template.New(name).
		Option("missingkey=error").
		Funcs(template.FuncMap(sprig.FuncMap())).
		Funcs(bpTemplateFuncMap()).
		Funcs(helpers.ResourceHelperFuncMap()).
		Parse(string(s))
	if err != nil {
		errs.Add(err)
		return
	}
	// ensure that it can be executed with default values
	if err := t.Execute(io.Discard, params); err != nil {
		errs.Add(err)
	}
}

func bpTemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		templateFuncHasCategoryMetricsEnabled: func(parameterValue []any, parameterName, metricCategory string) (bool, error) {
			return false, nil
		},
		templateFuncDisabledCategoryMetrics: func(parameterValue []any, parameterName, metricCategory string) ([]string, error) {
			return nil, nil
		},
		templateFuncComponentID: func(name string) (string, error) {
			return name, nil
		},
		templateFuncRouteID: func() (string, error) {
			return "", nil
		},
		templateFuncDefaultDisabledCategoryMetrics: func(parameterValue []any, parameterName, metricCategory string) ([]string, error) {
			return nil, nil
		},
	}
}

// TelemetryTypes returns the supported telemetry types (logs, metrics, or traces).
// Only applicable to SourceTypes.
func (s *ResourceTypeSpec) TelemetryTypes() []otel.PipelineType {
	telemetryTypes := make([]otel.PipelineType, 0, 3)

	if !s.Logs.Empty() || !s.LogsMetrics.Empty() || !s.LogsTraces.Empty() || !s.LogsMetricsTraces.Empty() {
		telemetryTypes = append(telemetryTypes, otel.Logs)
	}

	if !s.Metrics.Empty() || !s.LogsMetrics.Empty() || !s.MetricsTraces.Empty() || !s.LogsMetricsTraces.Empty() {
		telemetryTypes = append(telemetryTypes, otel.Metrics)
	}

	if !s.Traces.Empty() || !s.LogsTraces.Empty() || !s.MetricsTraces.Empty() || !s.LogsMetricsTraces.Empty() {
		telemetryTypes = append(telemetryTypes, otel.Traces)
	}

	return telemetryTypes
}

func (s *ResourceTypeSpec) validateNoDuplicateParameterNames(errs validation.Errors) {
	// visited is a map of parameter names to bool
	visited := make(map[string]bool, 0)

	for _, p := range s.Parameters {
		if visited[p.Name] {
			errs.Add(
				fmt.Errorf("found multiple parameters with name %s", p.Name),
			)
		} else {
			visited[p.Name] = true
		}
	}
}
