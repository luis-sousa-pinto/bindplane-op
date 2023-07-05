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
	"crypto/sha256"
	"database/sql/driver"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	modelSearch "github.com/observiq/bindplane-op/model/search"
	"github.com/observiq/bindplane-op/model/validation"
)

// Version indicates the version of a resource. It is generally incremented for each update to a resource, but
// configurations are only incremented when a rollout is started.
type Version int

// UnmarshalGQL implements the graphql.Unmarshaler interface.
func (v *Version) UnmarshalGQL(i interface{}) error {
	if value, ok := i.(int); ok {
		*v = Version(value)
		return nil
	}

	return errors.New("invalid version, must be int")
}

// MarshalGQL implements the graphql.Marshaler interface.
func (v Version) MarshalGQL(w io.Writer) {
	bytes := []byte(strconv.Itoa(int(v)))
	_, _ = w.Write(bytes)
}

const (
	// VersionPending refers to the pending Version of a resource, which is the version that is currently being rolled
	// out. This is currently only used for Configurations.
	VersionPending Version = -2

	// VersionCurrent refers to the current Version of a resource, which is the last version to be successfully rolled
	// out. This is currently only used for Configurations.
	VersionCurrent Version = -1

	// VersionLatest refers to the latest Version of a resource which is the latest version that has been created.
	VersionLatest Version = 0
)

const (
	// MetadataName TODO(doc)
	MetadataName = "name"
	// MetadataID TODO(doc)
	MetadataID = "id"
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
	KindRollout         Kind = "Rollout"
)

// Resource is implemented by all resources, e.g. SourceType, DestinationType, Configuration, etc.
type Resource interface {
	// all resources can be labeled
	Labeled

	// all resources can be indexed
	modelSearch.Indexed

	// all resources have a unique key
	HasUniqueKey

	// all resources can be printed
	Printable

	// Name returns the name for this resource
	Name() string

	// ID returns the uuid for this resource
	ID() string

	// SetID replaces the uuid for this resource
	SetID(id string)

	// EnsureID generates a new uuid for a resource if none exists
	EnsureID()

	// Version returns the version for this resource
	Version() Version

	// SetVersion replaces the version for this resource
	SetVersion(version Version)

	// Hash returns the hash of the resource spec
	Hash() string

	// EnsureHash generates a new hash for a resource if none exists. Hash is a hex formatted sha256 hash of the
	// json-encoded spec that is used to determine if the spec has changed.
	EnsureHash(spec any)

	// DateModified returns the date the resource was last modified
	DateModified() *time.Time

	// SetDateModified replaces the date the resource was last modified
	SetDateModified(date *time.Time)

	// EnsureMetadata ensures that the ID, Version, and Hash fields are set.
	EnsureMetadata(spec any)

	// GetKind returns the Kind of this resource
	GetKind() Kind

	// Description returns a description of the resource
	Description() string

	// Validate ensures that the resource is valid
	Validate() (warnings string, errors error)

	// ValidateWithStore ensures that the resource is valid and allows for extra validation given a store. It may also
	// make minor changes to the Resource, like ensuring references specify a specific version.
	ValidateWithStore(ctx context.Context, store ResourceStore) (warnings string, errors error)

	// UpdateDependencies updates the dependencies for this resource to use the latest version.
	UpdateDependencies(ctx context.Context, store ResourceStore) error

	// GetSpec returns the spec for this resource. All resources have a spec, but the format and contents of the spec will
	// be different for different resource types.
	GetSpec() any

	// GetStatus returns the status for this resource. All resources have a status, but the format and contents of the
	// status will be different for different resource types.
	// Initially, used for rollout status of a configuration.
	GetStatus() any

	// SetStatus replaces the status for this resource.
	SetStatus(status any) error

	// IsLatest returns true if the latest field on the status for this resource is true. Currently this is only used for
	// Configurations, Sources, Processors, Destinations, SourceTypes, ProcessorTypes, and DestinationTypes. For other
	// resources this always returns true.
	IsLatest() bool

	// SetLatest sets the value of the latest field on the status for this resource. Currently this is only used for
	// Configurations, Sources, Processors, Destinations, SourceTypes, ProcessorTypes, and DestinationTypes. For other
	// resources this does nothing.
	SetLatest(latest bool)

	// IsPending returns true if the pending field on the status for this resource is true. Currently this is only used
	// for Configurations. For other resources this always returns false.
	IsPending() bool

	// SetPending sets the value of the pending field on the status for this resource. Currently this is only used for
	// Configurations. For other resources this does nothing.
	SetPending(pending bool)

	// IsCurrent returns true if the current field on the status for this resource is true. Currently this is only used
	// for Configurations. For other resources this always returns false.
	IsCurrent() bool

	// SetCurrent sets the value of the current field on the status for this resource. Currently this is only used for
	// Configurations. For other resources this does nothing.
	SetCurrent(current bool)
}

// AnyResource is a resource not yet fully parsed and is the common structure of all Resources. The Spec, which is
// different for each kind of resource, is represented as a map[string]interface{} and can be further parsed using
// mapstructure. Use ParseResource or ParseResources to obtain a fully parsed Resource.
type AnyResource struct {
	ResourceMeta               `yaml:",inline" mapstructure:",squash"`
	Spec                       map[string]any `yaml:"spec" json:"spec" mapstructure:"spec"`
	StatusType[map[string]any] `yaml:",inline" json:",inline" mapstructure:",squash"`
}

// treat AnyResource as having sensitive parameters because it is possible the parsed spec will contain sensitive
// parameters
var _ HasSensitiveParameters = (*AnyResource)(nil)

// GetResourceMeta returns the ResourceMeta for this resource.
func (r *AnyResource) GetResourceMeta() ResourceMeta {
	return r.ResourceMeta
}

// GetSpec returns the spec for this resource.
func (r *AnyResource) GetSpec() any {
	return r.Spec
}

// Value is used to translate to a JSONB field for postgres storage
func (r AnyResource) Value() (driver.Value, error) {
	return jsoniter.Marshal(r)
}

// Scan is used to translate from a JSONB field in postgres to AnyResource
func (r *AnyResource) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return jsoniter.Unmarshal(b, &r)
}

// MaskSensitiveParameters masks sensitive parameter values based on the ParameterDefinitions in the ResourceType
func (r *AnyResource) MaskSensitiveParameters(ctx context.Context) {
	// get the underlying resource, mask, and then make it an AnyResource again
	parsed, err := ParseResource(r)
	if err != nil {
		// if we can't parse the resource, we can't mask it
		return
	}
	resourceWithSensitiveParameters, ok := parsed.(HasSensitiveParameters)
	if !ok {
		// the underlying resource doesn't have sensitive parameters, so there's nothing to mask
		return
	}

	// mask the sensitive parameters in the parsed resource
	resourceWithSensitiveParameters.MaskSensitiveParameters(ctx)

	// make the parsed resource an AnyResource again
	if anyResource, err := AsAny(parsed); err == nil {
		// replace the spec with the masked spec
		r.Spec = anyResource.Spec
	}
}

// PreserveSensitiveParameters will replace parameters with the SensitiveParameterPlaceholder value with the value of
// the parameter from the existing resource. This does nothing if existing is nil because there is no existing
// resource.
func (r *AnyResource) PreserveSensitiveParameters(ctx context.Context, existing *AnyResource) error {
	// get the underlying resource, preserve, and then make it an AnyResource again
	parsed, err := ParseResource(r)
	if err != nil {
		// if we can't parse the resource, we can't mask it
		return nil
	}
	resourceWithSensitiveParameters, ok := parsed.(HasSensitiveParameters)
	if !ok {
		// the underlying resource doesn't have sensitive parameters, so there's nothing to preserve
		return nil
	}
	err = resourceWithSensitiveParameters.PreserveSensitiveParameters(ctx, existing)
	if err != nil {
		return err
	}
	// make the parsed resource an AnyResource again
	if anyResource, err := AsAny(parsed); err == nil {
		// replace the spec with the masked spec
		r.Spec = anyResource.Spec
	}
	return nil
}

// ResourceMeta TODO(doc)
type ResourceMeta struct {
	APIVersion string   `yaml:"apiVersion,omitempty" json:"apiVersion"`
	Kind       Kind     `yaml:"kind,omitempty" json:"kind"`
	Metadata   Metadata `yaml:"metadata,omitempty" json:"metadata"`
}

// Metadata is the metadata about a resource
type Metadata struct {
	ID          string `yaml:"id,omitempty" json:"id" mapstructure:"id"`
	Name        string `yaml:"name,omitempty" json:"name" mapstructure:"name"`
	DisplayName string `yaml:"displayName,omitempty" json:"displayName,omitempty" mapstructure:"displayName"`
	Description string `yaml:"description,omitempty" json:"description,omitempty" mapstructure:"description"`
	Icon        string `yaml:"icon,omitempty" json:"icon,omitempty" mapstructure:"icon"`
	Labels      Labels `yaml:"labels,omitempty" json:"labels" mapstructure:"labels"`

	// Hash is a hex formatted sha256 Hash of the json-encoded spec that is used to determine if the spec has changed.
	Hash string `yaml:"hash,omitempty" json:"hash,omitempty" mapstructure:"hash"`

	// Version is a 1-based integer that is incremented each time the spec is changed.
	Version      Version    `yaml:"version,omitempty" json:"version,omitempty" mapstructure:"version"`
	DateModified *time.Time `yaml:"dateModified,omitempty" json:"dateModified,omitempty" mapstructure:"dateModified"`
}

// VersionStatus indicates if the resource is the latest version
type VersionStatus struct {
	Latest bool `json:"latest" yaml:"latest" mapstructure:"latest"`
}

// NoStatus is a placeholder for resources that do not have a status
type NoStatus struct {
}

// Parameter TODO(doc)
type Parameter struct {
	// Name is the name of the parameter
	Name string `json:"name" yaml:"name" mapstructure:"name"`

	// Value could be any of the following: string, bool, int, enum (string), float, []string, map
	Value interface{} `json:"value" yaml:"value" mapstructure:"value"`

	// Sensitive will be true if the value is sensitive and should be masked when printed.
	Sensitive bool `json:"sensitive,omitempty" yaml:"sensitive,omitempty" mapstructure:"sensitive"`
}

var _ Printable = (*ResourceMeta)(nil)

// GetResourceMeta returns the ResourceMeta for this resource.
func (r *ResourceMeta) GetResourceMeta() ResourceMeta {
	return *r
}

// UniqueKey returns the resource Name to uniquely identify a resource. Some Resource implementations use the ID instead.
func (r *ResourceMeta) UniqueKey() string {
	return r.Metadata.Name
}

// Version returns the version
func (r *ResourceMeta) Version() Version {
	if r == nil {
		return VersionLatest
	}
	return r.Metadata.Version
}

// SetVersion replaces the version for this resource
func (r *ResourceMeta) SetVersion(version Version) {
	r.Metadata.Version = version
}

// NameAndVersion returns the Resource name:version for this Resource
func (r *ResourceMeta) NameAndVersion() string {
	return NameAndVersion(r.Name(), r.Version())
}

// NameAndVersion returns a Resource name:version for the specified name and version
func NameAndVersion(name string, version Version) string {
	return fmt.Sprintf("%s:%d", name, version)
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
		r.Metadata.ID = NewResourceID()
	}
}

// Hash returns the hash of the resource spec
func (r *ResourceMeta) Hash() string {
	return r.Metadata.Hash
}

// EnsureHash computes the hash of the resource spec and sets it on the resource
func (r *ResourceMeta) EnsureHash(spec any) {
	r.Metadata.Hash = computeHash(spec)
}

// ComputeHash computes the hash of the resource spec
func computeHash(spec any) string {
	bytes, err := jsoniter.Marshal(spec)
	if err != nil {
		return ""
	}
	sha := sha256.Sum256(bytes)
	return hex.EncodeToString(sha[:])
}

// DateModified returns the date the resource was last modified
func (r *ResourceMeta) DateModified() *time.Time {
	return r.Metadata.DateModified
}

// SetDateModified replaces the date the resource was last modified
func (r *ResourceMeta) SetDateModified(date *time.Time) {
	r.Metadata.DateModified = date
}

// EnsureMetadata ensures that the metadata is set to reasonable defaults
func (r *ResourceMeta) EnsureMetadata(spec any) {
	r.EnsureID()
	r.EnsureHash(spec)
	if HasVersionKind(r.Kind) {
		if r.Metadata.Version == 0 {
			r.Metadata.Version = 1
		}
	}
}

// GetLabels implements the Labeled interface for Resources
func (r *ResourceMeta) GetLabels() Labels {
	return r.Metadata.Labels
}

// UpdateDependencies updates the dependencies for this resource to use the latest version.
func (r *ResourceMeta) UpdateDependencies(_ context.Context, _ ResourceStore) error {
	// Generic resources don't have dependencies.
	return nil
}

// SetLabels implements the Labeled interface for Resources
func (r *ResourceMeta) SetLabels(l Labels) {
	r.Metadata.Labels.Set = l.Set
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
	iconPath := filepath.Join(iconParts...)

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

// ResourceDateFormat is used when formatting dates for Resources.
const ResourceDateFormat = "2006-01-02 15:04:05"

// PrintableFieldValue returns the field value for a title, used for printing a table of resources. The implementation
// for ResourceMeta contains fields common to all resources. Resources for defer to this implementation for anything not
// specific to or overridden by that resource kind.
func (r *ResourceMeta) PrintableFieldValue(title string) string {
	switch title {
	case "ID":
		return r.ID()
	case "Name":
		return r.Name()
	case "Hash":
		return r.Hash()
	case "Display":
		return r.Metadata.DisplayName
	case "Description":
		return r.Metadata.Description
	case "Version":
		return strconv.Itoa(int(r.Metadata.Version))
	case "Date":
		if r.DateModified() == nil {
			return "-"
		}
		return r.DateModified().Format(ResourceDateFormat)
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
func (r *ResourceMeta) IndexFields(index modelSearch.Indexer) {
	index("kind", string(r.Kind))
	r.Metadata.indexFields(index)
}

// IndexLabels returns a map of label name to label value to be stored in the index
func (r *ResourceMeta) IndexLabels(index modelSearch.Indexer) {
	r.Metadata.indexLabels(index)
}

// ----------------------------------------------------------------------
// convenience

// SetVersion sets the Version on the Resource and returns it. It does not return a modified copy. It modifies the
// existing resource and returns it as a convenience.
func SetVersion[R Resource](resource R, Version Version) R {
	resource.SetVersion(Version)
	return resource
}

// SplitVersionDefault splits a resource key into the resource key and version and allows a default version to be
// specified if none is specified or the specified version cannot be parsed.
func SplitVersionDefault(resourceKey string, defaultVersion Version) (string, Version) {
	parts := strings.SplitN(resourceKey, ":", 2)
	name := parts[0]
	if len(parts) == 1 {
		return name, defaultVersion
	}
	switch parts[1] {
	case "":
		return name, defaultVersion
	case "latest":
		return name, VersionLatest
	case "stable", "current":
		return name, VersionCurrent
	case "pending":
		return name, VersionPending
	}
	version, err := strconv.Atoi(parts[1])
	if err != nil {
		return name, defaultVersion
	}
	return name, Version(version)
}

// TrimVersion removes a version from a resource key. It does nothing if there is no version.
func TrimVersion(resourceKey string) string {
	key, _ := SplitVersion(resourceKey)
	return key
}

// SplitVersion splits a resource key into the resource key and version.
func SplitVersion(resourceKey string) (string, Version) {
	return SplitVersionDefault(resourceKey, VersionLatest)
}

// JoinVersion joins a resource key and version into a resource key for the specified version.
func JoinVersion(resourceKey string, version Version) string {
	// make sure there isn't already a version
	resourceKey, _ = SplitVersion(resourceKey)
	switch version {
	case VersionLatest:
		return resourceKey

	case VersionCurrent:
		return fmt.Sprintf("%s:current", resourceKey)

	case VersionPending:
		return fmt.Sprintf("%s:pending", resourceKey)

	default:
		return fmt.Sprintf("%s:%d", resourceKey, version)
	}
}

// indexFields returns a map of field name to field value to be stored in the index
func (m *Metadata) indexFields(index modelSearch.Indexer) {
	index("id", m.ID)
	index("name", m.Name)
	index("displayName", m.DisplayName)
	index("description", m.Description)
}

// indexLabels returns a map of label name to label value to be stored in the index
func (m *Metadata) indexLabels(index modelSearch.Indexer) {
	for n, v := range m.Labels.Set {
		index(n, v)
	}
}
