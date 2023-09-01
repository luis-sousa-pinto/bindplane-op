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

type sourceKind struct{}

func (k *sourceKind) NewEmptyResource() *Source { return &Source{} }

// Source will generate an exporter and be at the end of a pipeline
type Source struct {
	// ResourceMeta is the metadata for the Source
	ResourceMeta `yaml:",inline" mapstructure:",squash"`
	// Spec is the specification for the Source that contains the type and parameters
	Spec                      ParameterizedSpec `json:"spec" yaml:"spec" mapstructure:"spec"`
	StatusType[VersionStatus] `yaml:",inline" mapstructure:",squash"`
}

var _ parameterizedResource = (*Source)(nil)
var _ HasSensitiveParameters = (*Source)(nil)

// ValidateWithStore checks that the source is valid, returning an error if it is not. It uses the store to retrieve the
// source type so that parameter values can be validated against the parameter definitions.
func (s *Source) ValidateWithStore(ctx context.Context, store ResourceStore) (warnings string, errors error) {
	errs := validation.NewErrors()

	s.validate(errs)
	s.Spec.validateTypeAndParameters(ctx, KindSource, errs, store)

	return errs.Warnings(), errs.Result()
}

// UpdateDependencies updates the dependencies for this resource to use the latest version.
func (s *Source) UpdateDependencies(ctx context.Context, store ResourceStore) error {
	return s.Spec.updateDependencies(ctx, KindSource, store)
}

// GetKind returns "Source"
func (s *Source) GetKind() Kind { return KindSource }

// GetSpec returns the spec for this resource.
func (s *Source) GetSpec() any {
	return s.Spec
}

// ResourceTypeName is the name of the ResourceType that renders this resource type
func (s *Source) ResourceTypeName() string {
	return s.Spec.Type
}

// ResourceParameters are the parameters passed to the ResourceType to generate the configuration
func (s *Source) ResourceParameters() []Parameter {
	return s.Spec.Parameters
}

// ComponentID provides a unique component id for the specified component name
func (s *Source) ComponentID(name string) otel.ComponentID {
	return otel.UniqueComponentID(name, s.Spec.Type, s.Name())
}

// NewSource creates a new Source with the specified name, type, and parameters
func NewSource(name string, sourceTypeName string, parameters []Parameter) *Source {
	return NewSourceWithSpec(name, ParameterizedSpec{
		Type:       sourceTypeName,
		Parameters: parameters,
	})
}

// NewSourceWithSpec creates a new Source with the specified spec
func NewSourceWithSpec(name string, spec ParameterizedSpec) *Source {
	s := &Source{
		ResourceMeta: ResourceMeta{
			APIVersion: "bindplane.observiq.com/v1",
			Kind:       KindSource,
			Metadata: Metadata{
				Name:   name,
				Labels: MakeLabels(),
			},
		},
		Spec: spec,
	}
	s.EnsureMetadata(spec)
	return s
}

// FindSource returns a Source from the store if it exists. If it doesn't exist, it creates a new Source with the
// specified defaultName.
func FindSource(ctx context.Context, source *ResourceConfiguration, defaultName string, store ResourceStore) (*Source, error) {
	if source.Name == "" {
		// inline source
		src := NewSourceWithSpec(defaultName, source.ParameterizedSpec)
		return src, nil
	}
	// named source
	src, err := store.Source(ctx, source.Name)
	if err != nil {
		return nil, err
	}
	if src == nil {
		return nil, fmt.Errorf("unknown %s: %s", KindSource, source.Name)
	}
	return src, nil
}

// ----------------------------------------------------------------------

// PrintableFieldTitles returns the list of field titles, used for printing a table of resources
func (s *Source) PrintableFieldTitles() []string {
	return []string{"Name", "Type", "Description"}
}

// PrintableFieldValue returns the field value for a title, used for printing a table of resources
func (s *Source) PrintableFieldValue(title string) string {
	switch title {
	case "Type":
		return s.ResourceTypeName()
	default:
		return s.ResourceMeta.PrintableFieldValue(title)
	}
}

// ----------------------------------------------------------------------

// MaskSensitiveParameters masks sensitive parameter values based on the ParameterDefinitions in the ResourceType
func (s *Source) MaskSensitiveParameters(ctx context.Context) {
	s.Spec.maskSensitiveParameters(ctx)
}

// PreserveSensitiveParameters will replace parameters with the SensitiveParameterPlaceholder value with the value of
// the parameter from the existing resource. This does nothing if existing is nil because there is no existing
// resource.
func (s *Source) PreserveSensitiveParameters(ctx context.Context, existing *AnyResource) error {
	return PreserveSensitiveParameters(ctx, s, existing)
}
