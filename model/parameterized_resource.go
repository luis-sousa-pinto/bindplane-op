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

	"github.com/observiq/bindplane-op/model/otel"
	"github.com/observiq/bindplane-op/model/validation"
)

// ParameterizedSpec is the spec for a ParameterizedResource
type ParameterizedSpec struct {
	Type       string      `yaml:"type,omitempty" json:"type,omitempty" mapstructure:"type"`
	Parameters []Parameter `yaml:"parameters,omitempty" json:"parameters,omitempty" mapstructure:"parameters"`

	Processors []ResourceConfiguration `yaml:"processors,omitempty" json:"processors,omitempty" mapstructure:"processors"`
	Disabled   bool                    `yaml:"disabled,omitempty" json:"disabled,omitempty" mapstructure:"disabled"`
}

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
