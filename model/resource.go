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
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/observiq/bindplane-op/internal/store/search"
	"github.com/observiq/bindplane-op/model/validation"
	"gopkg.in/yaml.v2"
)

const (
	// MetadataName TODO(doc)
	MetadataName = "name"
	// MetadataID TODO(doc)
	MetadataID = "id"
)

const (
	// V1 is the version for the initial resources defined for BindPlane
	V1 = "bindplane.observiq.com/v1"
)

// Kind indicates the kind of resource, e.g. Configuration
type Kind string

// Kind values correspond to the kinds of resources currently supported by BindPlane
const (
	KindProfile         Kind = "Profile"
	KindContext         Kind = "Context"
	KindConfiguration   Kind = "Configuration"
	KindAgent           Kind = "Agent"
	KindAgentVersion    Kind = "AgentVersion"
	KindSource          Kind = "Source"
	KindProcessor       Kind = "Processor"
	KindDestination     Kind = "Destination"
	KindSourceType      Kind = "SourceType"
	KindProcessorType   Kind = "ProcessorType"
	KindDestinationType Kind = "DestinationType"
	KindUnknown         Kind = "Unknown"
)

// createKindLookup creates a map from lowercase name => Kind, including the plural form by adding an "s" to the end of
// the name. This is used by ParseKind.
func createKindLookup() map[string]Kind {
	result := map[string]Kind{}
	for _, kind := range []Kind{
		KindProfile,
		KindContext,
		KindConfiguration,
		KindAgent,
		KindAgentVersion,
		KindSource,
		KindProcessor,
		KindDestination,
		KindSourceType,
		KindProcessorType,
		KindDestinationType,
	} {
		key := strings.ToLower(string(kind))
		plural := fmt.Sprintf("%ss", key)
		result[key] = kind
		result[plural] = kind
	}
	return result
}

var kindLookup = createKindLookup()

// Resource is implemented by all resources, e.g. SourceType, DestinationType, Configuration, etc.
type Resource interface {
	// all resources can be labeled
	Labeled

	// all resources can be indexed
	search.Indexed

	// all resources have a unique key
	HasUniqueKey

	// ID returns the uuid for this resource
	ID() string

	// SetID replaces the uuid for this resource
	SetID(id string)

	// EnsureID generates a new uuid for a resource if none exists
	EnsureID()

	// Name returns the name for this resource
	Name() string

	// GetKind returns the Kind of this resource
	GetKind() Kind

	// Description returns a description of the resource
	Description() string

	// Validate ensures that the resource is valid
	Validate() (warnings string, errors error)

	// ValidateWithStore ensures that the resource is valid and allows for extra validation given a store
	ValidateWithStore(ctx context.Context, store ResourceStore) (warnings string, errors error)
}

// AnyResource is a resource not yet fully parsed and is the common structure of all Resources. The Spec, which is
// different for each kind of resource, is represented as a map[string]interface{} and can be further parsed using
// mapstructure. Use ParseResource or ParseResources to obtain a fully parsed Resource.
type AnyResource struct {
	ResourceMeta `yaml:",inline" json:",inline" mapstructure:",squash"`
	Spec         map[string]interface{} `yaml:"spec" json:"spec" mapstructure:"spec"`
}

// ResourceMeta TODO(doc)
type ResourceMeta struct {
	APIVersion string   `yaml:"apiVersion,omitempty" json:"apiVersion"`
	Kind       Kind     `yaml:"kind,omitempty" json:"kind"`
	Metadata   Metadata `yaml:"metadata,omitempty" json:"metadata"`
}

// Metadata TODO(doc)
type Metadata struct {
	ID          string `yaml:"id,omitempty" json:"id" mapstructure:"id"`
	Name        string `yaml:"name,omitempty" json:"name" mapstructure:"name"`
	DisplayName string `yaml:"displayName,omitempty" json:"displayName,omitempty" mapstructure:"displayName"`
	Description string `yaml:"description,omitempty" json:"description,omitempty" mapstructure:"description"`
	Icon        string `yaml:"icon,omitempty" json:"icon,omitempty" mapstructure:"icon"`
	Labels      Labels `yaml:"labels,omitempty" json:"labels" mapstructure:"labels"`
}

// Parameter TODO(doc)
type Parameter struct {
	Name string `json:"name" yaml:"name" mapstructure:"name"`
	// This could be any of the following: string, bool, int, enum (string), float, []string
	Value interface{} `json:"value" yaml:"value" mapstructure:"value"`
}

var _ Resource = (*ResourceMeta)(nil)
var _ Printable = (*ResourceMeta)(nil)

// UniqueKey returns the resource Name to uniquely identify a resource
func (r *ResourceMeta) UniqueKey() string {
	return r.Metadata.Name
}

// ID returns the ID
func (r *ResourceMeta) ID() string {
	return r.Metadata.ID
}

// SetID replaces the uuid for this resource
func (r *ResourceMeta) SetID(id string) {
	r.Metadata.ID = id
}

// GetKind returns the Kind of this resource.
func (r *ResourceMeta) GetKind() Kind {
	return r.Kind
}

// Name returns the name.
func (r *ResourceMeta) Name() string {
	return r.Metadata.Name
}

// Description returns the description.
func (r *ResourceMeta) Description() string {
	return r.Metadata.Description
}

// EnsureID sets the ID to a random uuid if not already set.
func (r *ResourceMeta) EnsureID() {
	if r.Metadata.ID == "" {
		r.Metadata.ID = uuid.NewString()
	}
}

// GetLabels implements the Labeled interface for Agents
func (r *ResourceMeta) GetLabels() Labels {
	return r.Metadata.Labels
}

// Validate checks that the resource is valid, returning an error if it is not. This provides generic validation for all
// resources. Specific resources should provide their own Validate method and call this to validate the ResourceMeta.
func (r *ResourceMeta) Validate() (warnings string, errors error) {
	errs := validation.NewErrors()
	r.validate(errs)
	return errs.Warnings(), errs.Result()
}

// ValidateWithStore allows for additional validation when a store is available.
func (r *ResourceMeta) ValidateWithStore(_ context.Context, _ ResourceStore) (warnings string, errors error) {
	return r.Validate()
}

// validate can be used from other resources to validate the Meta portion of the resource
func (r *ResourceMeta) validate(errs validation.Errors) {
	validateKind(errs, string(r.Kind))
	r.validateIcon(r.Kind, errs)
	r.Metadata.validate(errs)
}

func (r *ResourceMeta) validateIcon(kind Kind, errs validation.Errors) {
	// only currently validated for source and destination types
	switch kind {
	case KindSourceType:
	case KindDestinationType:
	default:
		return
	}

	if r.Metadata.Icon == "" {
		errs.Warn(fmt.Errorf("%s %s is missing .metadata.icon", kind, r.Name()))
		return
	}

	// find the root folder of the repo. this works because we know we're being called by something in this package
	// because this function isn't exported. we also know that this file is the model package at the root of the repo.
	_, modelFolder, _, _ := runtime.Caller(0)
	repoFolder := filepath.Join(filepath.Dir(modelFolder), "..")

	// construct the icon path from the iconParts knowing that we store icons in the ui/public folder and the icon will be a
	// relative folder inside this path
	iconParts := []string{repoFolder, "ui", "public"}
	iconParts = append(iconParts, strings.Split(r.Metadata.Icon, "/")...)
	iconPath := path.Join(iconParts...)

	// attempt to read the file to verify that it exists
	info, err := os.Stat(iconPath)
	if err != nil {
		errs.Warn(fmt.Errorf("%s %s icon cannot be read: %w", kind, r.Name(), err))
		return
	}
	if info.Size() == 0 {
		errs.Warn(fmt.Errorf("%s %s icon empty at %s", kind, r.Name(), iconPath))
	}
}

func (m *Metadata) validate(errs validation.Errors) {
	validation.IsName(errs, m.Name)
	m.Labels.validate(errs)
}

func validateKind(errors validation.Errors, kind string) {
	// it's possible for parsed to be unmarshaled to a string that isn't a valid type
	if parsed := ParseKind(string(kind)); parsed == KindUnknown {
		errors.Add(fmt.Errorf("%s is not a valid resource kind", kind))
	}
}

// ParseKind parses a kind from a specified string parameter, validating that it matches an existing kind. It ignores
// the case of the string parameter and also allows plurals, e.g. configurations => KindConfiguration. KindUnknown is
// returned if that specified kind does not match any known Kinds.
func ParseKind(kind string) Kind {
	lower := strings.ToLower(kind)
	if kind, ok := kindLookup[lower]; ok {
		return kind
	}
	return KindUnknown
}

// ParseResource maps the Spec of the provided resource to a specific type of Resource
// It will drop any unused keys
func ParseResource(r *AnyResource) (Resource, error) {
	return parseResource(r, false)
}

// ParseResourceStrict maps the Spec of the provided resource to a specific type of Resource
// It will error if there are any unused keys
func ParseResourceStrict(r *AnyResource) (Resource, error) {
	return parseResource(r, true)
}

func parseResource(r *AnyResource, strict bool) (Resource, error) {
	switch r.Kind {
	case KindProfile:
		return parseProfile(r, strict)
	case KindContext:
		return parseContext(r, strict)
	case KindConfiguration:
		return unmarshalResource(r, &Configuration{}, strict)
	case KindSource:
		return unmarshalResource(r, &Source{}, strict)
	case KindSourceType:
		return unmarshalResource(r, &SourceType{}, strict)
	case KindProcessor:
		return unmarshalResource(r, &Processor{}, strict)
	case KindProcessorType:
		return unmarshalResource(r, &ProcessorType{}, strict)
	case KindDestination:
		return unmarshalResource(r, &Destination{}, strict)
	case KindDestinationType:
		return unmarshalResource(r, &DestinationType{}, strict)
	case KindAgentVersion:
		return unmarshalResource(r, &AgentVersion{}, strict)
	}

	return nil, fmt.Errorf("unknown resource kind: %s", r.Kind)
}

// ParseResources parses all the generic AnyResources into their concrete resource structs.
func ParseResources(resources []*AnyResource) ([]Resource, error) {
	result := []Resource{}

	for _, resource := range resources {
		parsed, err := ParseResource(resource)
		if err != nil {
			return result, err
		}
		result = append(result, parsed)
	}

	return result, nil
}

// ParseResourcesStrict parses all the generic AnyResources into their concrete resource structs.
// Any extra fields on any of the resources will cause an error.
func ParseResourcesStrict(resources []*AnyResource) ([]Resource, error) {
	result := []Resource{}

	for _, resource := range resources {
		parsed, err := ParseResourceStrict(resource)
		if err != nil {
			return result, err
		}
		result = append(result, parsed)
	}

	return result, nil
}

// ResourcesFromFile creates an io.Reader from reading the given file and uses unmarshalResources
// to return a slice of *AnyResource read from the file.
func ResourcesFromFile(filename string) ([]*AnyResource, error) {
	file, err := os.Open(filepath.Clean(filename))
	if err != nil {
		return nil, err
	}

	resources, err := ResourcesFromReader(file)
	if err != nil {
		return nil, err
	}
	return resources, file.Close()
}

// ResourcesFromReader creates a yaml decoder from an io.Reader and returns a slice of *AnyResource and an error.
// If the decoder is able to reach the end of the reader with no error, err will be nil.
func ResourcesFromReader(reader io.Reader) ([]*AnyResource, error) {
	resources := []*AnyResource{}
	dec := yaml.NewDecoder(reader)

	for {
		resource := &AnyResource{}
		if err := dec.Decode(resource); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		resources = append(resources, resource)
	}

	return resources, nil
}

// unmarshalResource unmarshals the *AnyResource into the provided instance.
// If errorUnused is true, the unmarshal will fail if any keys are not mapped to a field on the instance.
func unmarshalResource[T Resource](r *AnyResource, instance T, errorUnused bool) (T, error) {
	if r.Kind != instance.GetKind() {
		return instance, fmt.Errorf("invalid resource kind: %s", r.Kind)
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ErrorUnused: errorUnused,
		Result:      instance,
	})
	if err != nil {
		return instance, fmt.Errorf("failed to create decoder: %w", err)
	}

	err = decoder.Decode(r)
	if err != nil {
		return instance, fmt.Errorf("failed to decode definition: %w", err)
	}
	return instance, nil
}

// ----------------------------------------------------------------------
// Printable

// PrintableKindSingular returns the singular form of the Kind, e.g. "Configuration"
func (r *ResourceMeta) PrintableKindSingular() string {
	return string(r.Kind)
}

// PrintableKindPlural returns the plural form of the Kind, e.g. "Configurations"
func (r *ResourceMeta) PrintableKindPlural() string {
	// the default implementation assumes we can add "s"
	return fmt.Sprintf("%ss", r.Kind)
}

// PrintableFieldTitles returns the list of field titles, used for printing a table of resources
func (r *ResourceMeta) PrintableFieldTitles() []string {
	return []string{"Name"}
}

// PrintableFieldValue returns the field value for a title, used for printing a table of resources
func (r *ResourceMeta) PrintableFieldValue(title string) string {
	switch title {
	case "ID":
		return r.ID()
	case "Name":
		return r.Name()
	case "Display":
		return r.Metadata.DisplayName
	default:
		return "-"
	}
}

// ----------------------------------------------------------------------
// Indexed

// IndexID returns an ID used to identify the resource that is indexed
func (r *ResourceMeta) IndexID() string {
	return r.Metadata.Name
}

// IndexFields returns a map of field name to field value to be stored in the index
func (r *ResourceMeta) IndexFields(index search.Indexer) {
	index("kind", string(r.Kind))
	r.Metadata.indexFields(index)
}

// IndexLabels returns a map of label name to label value to be stored in the index
func (r *ResourceMeta) IndexLabels(index search.Indexer) {
	r.Metadata.indexLabels(index)
}

// indexFields returns a map of field name to field value to be stored in the index
func (m *Metadata) indexFields(index search.Indexer) {
	index("id", m.ID)
	index("name", m.Name)
	index("displayName", m.DisplayName)
	index("description", m.Description)
}

// indexLabels returns a map of label name to label value to be stored in the index
func (m *Metadata) indexLabels(index search.Indexer) {
	for n, v := range m.Labels.Set {
		index(n, v)
	}
}

// NewEmptyResource will return a zero value struct for the given resource kind.
func NewEmptyResource(kind Kind) (Resource, error) {
	switch kind {
	case KindAgentVersion:
		return &AgentVersion{}, nil
	case KindConfiguration:
		return &Configuration{}, nil
	case KindSource:
		return &Source{}, nil
	case KindProcessor:
		return &Processor{}, nil
	case KindDestination:
		return &Destination{}, nil
	case KindSourceType:
		return &SourceType{}, nil
	case KindProcessorType:
		return &ProcessorType{}, nil
	case KindDestinationType:
		return &DestinationType{}, nil
	default:
		return nil, fmt.Errorf("cannot make empty resource for unexpected kind: %s", kind)
	}
}
