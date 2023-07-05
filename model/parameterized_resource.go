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
	"context"
	"fmt"

	"github.com/observiq/bindplane-op/model/otel"
	"github.com/observiq/bindplane-op/model/validation"
)

// HasResourceParameters returns true for any resource with parameters. This includes top-level resources like Sources,
// Processors, and Destinations and embedded resources within Configurations.
type HasResourceParameters interface {
	// Parameters returns the parameters for this resource
	ResourceParameters() []Parameter
}

// HasSensitiveParameters is an interface for resources that can have sensitive parameters
type HasSensitiveParameters interface {
	// MaskSensitiveParameters masks sensitive parameter values based on the ParameterDefinitions in the ResourceType
	MaskSensitiveParameters(ctx context.Context)

	// PreserveSensitiveParameters will replace parameters with the SensitiveParameterPlaceholder value with the value of
	// the parameter from the existing resource. This does nothing if existing is nil because there is no existing
	// resource.
	PreserveSensitiveParameters(ctx context.Context, existing *AnyResource) error
}

// ParameterizedSpec is the spec for a ParameterizedResource
type ParameterizedSpec struct {
	Type       string      `yaml:"type,omitempty" json:"type,omitempty" mapstructure:"type"`
	Parameters []Parameter `yaml:"parameters,omitempty" json:"parameters,omitempty" mapstructure:"parameters"`

	Processors []ResourceConfiguration `yaml:"processors,omitempty" json:"processors,omitempty" mapstructure:"processors"`
	Disabled   bool                    `yaml:"disabled,omitempty" json:"disabled,omitempty" mapstructure:"disabled"`
}

// SensitiveParameterPlaceholder is the value returned for sensitive parameters
const SensitiveParameterPlaceholder = "(sensitive)"

// parameterizedResource is a resource based on a resource type which provides a specific resource value via templated
type parameterizedResource interface {
	otel.ComponentIDProvider

	// Name returns the name for this resource
	Name() string

	// ResourceTypeName is the name of the ResourceType that renders this resource type
	ResourceTypeName() string

	// ResourceParameters are the parameters passed to the ResourceType to generate the configuration
	ResourceParameters() []Parameter
}

var _ HasResourceParameters = (*ParameterizedSpec)(nil)
var _ HasResourceParameters = (parameterizedResource)(nil)

// ResourceParameters returns the parameters for this resource
func (s *ParameterizedSpec) ResourceParameters() []Parameter {
	return s.Parameters
}

// validateTypeAndParameters is used by Source and Destination validation and uses methods created for Configuration
// validation.
func (s *ParameterizedSpec) validateTypeAndParameters(ctx context.Context, kind Kind, errors validation.Errors, store ResourceStore) {
	// ResourceConfiguration is a resource embedded in a Configuration, but it works equally well for Source and
	// Destination validation.
	rc := &ResourceConfiguration{
		ParameterizedSpec: ParameterizedSpec{
			Type:       s.Type,
			Parameters: s.Parameters,
			Processors: s.Processors,
		},
	}
	rc.validateParameters(ctx, kind, errors, store)
	rc.validateProcessors(ctx, kind, errors, store)

	// the type may have been modified, so copy it back
	s.Type = rc.ParameterizedSpec.Type
	// the parameters may have had their values modified, so copy them back
	s.Parameters = rc.ParameterizedSpec.Parameters
}

func (s *ParameterizedSpec) trimVersions() {
	// first remove the dependency versions
	s.Type = TrimVersion(s.Type)

	for i, p := range s.Processors {
		p.trimVersions()
		s.Processors[i] = p
	}
}

func (s *ParameterizedSpec) updateDependencies(ctx context.Context, kind Kind, store ResourceStore) error {
	s.trimVersions()

	// validate to set the latest version
	errs := validation.NewErrors()
	s.validateTypeAndParameters(ctx, kind, errs, store)
	return errs.Result()
}

func (s *ParameterizedSpec) maskSensitiveParameters(ctx context.Context) {
	maskSensitiveParameters(ctx, s)
	for i, p := range s.Processors {
		p.maskSensitiveParameters(ctx)
		s.Processors[i] = p
	}
}

// maskSensitiveParameters will mask sensitive parameters in the spec based on the Sensitive option in the
// ParameterDefinition.
func maskSensitiveParameters(ctx context.Context, resource HasResourceParameters) {
	if IsWithoutSensitiveParameterMasking(ctx) {
		return
	}
	params := resource.ResourceParameters()
	for i, param := range params {
		if param.Sensitive {
			param.Value = SensitiveParameterPlaceholder
			params[i] = param
		}
	}
}

// ParameterValue returns the value of the first Parameter with the specified name. If multiple Parameters exist with the
// specified name, only the first one will be returned.
func ParameterValue(parameters []Parameter, name string) any {
	for _, param := range parameters {
		if name == param.Name {
			return param.Value
		}
	}
	return nil
}

// PreserveSensitiveParameters will preserve sensitive parameters in the spec with the values from the existing spec.
func PreserveSensitiveParameters(ctx context.Context, resource HasResourceParameters, existing *AnyResource) error {
	if existing == nil {
		return nil
	}

	// parse the existing resource to get the parameters
	parsed, err := ParseResource(existing)
	if err != nil {
		return fmt.Errorf("unable to parse existing resource: %v %w", existing, err)
	}
	existingParameterized, ok := parsed.(HasResourceParameters)
	if !ok {
		return nil
	}
	preserveSensitiveParameters(ctx, resource, existingParameterized)
	return nil
}
func preserveSensitiveParameters(_ context.Context, resource HasResourceParameters, existing HasResourceParameters) {
	if existing == nil {
		return
	}
	existingParameters := existing.ResourceParameters()
	params := resource.ResourceParameters()
	for i, param := range params {
		// if the parameter is sensitive, replace the value with the value from the existing resource. note that we could
		// also check to ensure that the parameter value is marked Sensitive, but we assume anything with "(sensitive)" is
		// sensitive.
		if param.Value == SensitiveParameterPlaceholder {
			param.Value = ParameterValue(existingParameters, param.Name)
			params[i] = param
		}
	}
}

// ----------------------------------------------------------------------
// conditional parameter masking based on the context

type key int

var withoutSensitiveParameterMaskingKey key

// ContextWithoutSensitiveParameterMasking returns a context that will not mask sensitive parameters
func ContextWithoutSensitiveParameterMasking(ctx context.Context) context.Context {
	return context.WithValue(ctx, withoutSensitiveParameterMaskingKey, true)
}

// IsWithoutSensitiveParameterMasking returns true if the context has been set to not mask sensitive parameters
func IsWithoutSensitiveParameterMasking(ctx context.Context) bool {
	if without, ok := ctx.Value(withoutSensitiveParameterMaskingKey).(bool); ok {
		return without
	}
	return false
}
