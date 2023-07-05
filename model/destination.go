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
	"github.com/observiq/bindplane-op/model/version"
)

type destinationKind struct{}

func (k *destinationKind) NewEmptyResource() *Destination { return &Destination{} }

// Destination will generate an exporter and be at the end of a pipeline
type Destination struct {
	// ResourceMeta TODO(doc)
	ResourceMeta `yaml:",inline" json:",inline" mapstructure:",squash"`
	// Spec TODO(doc)
	Spec                      ParameterizedSpec `json:"spec" yaml:"spec" mapstructure:"spec"`
	StatusType[VersionStatus] `yaml:",inline" json:",inline" mapstructure:",squash"`
}

var _ parameterizedResource = (*Destination)(nil)
var _ HasSensitiveParameters = (*Destination)(nil)

// ValidateWithStore checks that the destination is valid, returning an error if it is not. It uses the store to
// retrieve the destination type so that parameter values can be validated against the parameter definitions.
func (d *Destination) ValidateWithStore(ctx context.Context, store ResourceStore) (warnings string, errors error) {
	errs := validation.NewErrors()

	d.validate(errs)
	d.Spec.validateTypeAndParameters(ctx, KindDestination, errs, store)

	return errs.Warnings(), errs.Result()
}

// UpdateDependencies updates the dependencies for this resource to use the latest version.
func (d *Destination) UpdateDependencies(ctx context.Context, store ResourceStore) error {
	return d.Spec.updateDependencies(ctx, KindDestination, store)
}

// GetKind returns "Destination"
func (d *Destination) GetKind() Kind { return KindDestination }

// GetSpec returns the spec for this resource.
func (d *Destination) GetSpec() any {
	return d.Spec
}

// ResourceTypeName is the name of the ResourceType that renders this resource type
func (d *Destination) ResourceTypeName() string {
	return d.Spec.Type
}

// ResourceParameters are the parameters passed to the ResourceType to generate the configuration
func (d *Destination) ResourceParameters() []Parameter {
	return d.Spec.Parameters
}

// ComponentID provides a unique component id for the specified component name
func (d *Destination) ComponentID(name string) otel.ComponentID {
	return otel.UniqueComponentID(name, d.Spec.Type, d.Name())
}

// NewDestination creates a new Destination with the specified name, type, and parameters
func NewDestination(name string, typeValue string, parameters []Parameter) *Destination {
	return NewDestinationWithSpec(name, ParameterizedSpec{
		Type:       typeValue,
		Parameters: parameters,
	})
}

// NewDestinationWithSpec creates a new Destination with the specified spec
func NewDestinationWithSpec(name string, spec ParameterizedSpec) *Destination {
	d := &Destination{
		ResourceMeta: ResourceMeta{
			APIVersion: version.V1,
			Kind:       KindDestination,
			Metadata: Metadata{
				Name:   name,
				Labels: MakeLabels(),
			},
		},
		Spec: spec,
		StatusType: StatusType[VersionStatus]{
			Status: VersionStatus{Latest: true},
		},
	}
	d.EnsureMetadata(spec)
	return d
}

// FindDestination returns a Destination from the store if it exists. If it doesn't exist, it creates a new Destination with the
// specified defaultName.
func FindDestination(ctx context.Context, destination *ResourceConfiguration, defaultName string, store ResourceStore) (*Destination, error) {
	if destination.Name == "" {
		// inline destination
		return NewDestinationWithSpec(defaultName, destination.ParameterizedSpec), nil
	}
	// named destination
	dest, err := store.Destination(ctx, destination.Name)
	if err != nil {
		return nil, err
	}
	if dest == nil {
		return nil, fmt.Errorf("unknown %s: %s", KindDestination, destination.Name)
	}
	return dest, nil
}

// ----------------------------------------------------------------------

// PrintableFieldTitles returns the list of field titles, used for printing a table of resources
func (d *Destination) PrintableFieldTitles() []string {
	return []string{"Name", "Type", "Description"}
}

// PrintableFieldValue returns the field value for a title, used for printing a table of resources
func (d *Destination) PrintableFieldValue(title string) string {
	switch title {
	case "Type":
		return d.ResourceTypeName()
	default:
		return d.ResourceMeta.PrintableFieldValue(title)
	}
}

// ----------------------------------------------------------------------

// MaskSensitiveParameters masks sensitive parameter values based on the ParameterDefinitions in the ResourceType
func (d *Destination) MaskSensitiveParameters(ctx context.Context) {
	d.Spec.maskSensitiveParameters(ctx)
}

// PreserveSensitiveParameters will replace parameters with the SensitiveParameterPlaceholder value with the value of
// the parameter from the existing resource. This does nothing if existing is nil because there is no existing
// resource.
func (d *Destination) PreserveSensitiveParameters(ctx context.Context, existing *AnyResource) error {
	return PreserveSensitiveParameters(ctx, d, existing)
}
