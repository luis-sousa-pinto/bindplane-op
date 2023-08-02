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
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/mitchellh/mapstructure"
	"github.com/observiq/bindplane-op/model/version"
	"gopkg.in/yaml.v3"
)

// AsKind either casts the Resource to the expected type or parses an AnyResource to the expected type. It returns an
// error if the parsing fails or if the resource is not of the expected type.
func AsKind[T Resource](resource Resource) (T, error) {
	if anyResource, ok := resource.(*AnyResource); ok {
		result, err := ParseOne[T](anyResource)
		return result, err
	}
	if result, ok := resource.(T); ok {
		return result, nil
	}
	var empty T
	return empty, fmt.Errorf("resource of kind %s is not the expected type", resource.GetKind())
}

// AsAny converts an arbitrary Resource to an AnyResource. It is mainly used for testing.
func AsAny(r Resource) (*AnyResource, error) {
	if r == nil {
		return nil, nil
	}
	value := reflect.ValueOf(r)
	if value.Kind() == reflect.Ptr && value.IsNil() {
		return nil, nil
	}
	if anyResource, ok := r.(*AnyResource); ok {
		return anyResource, nil
	}
	bytes, err := jsoniter.Marshal(r)
	if err != nil {
		return nil, err
	}
	res := AnyResource{}
	err = jsoniter.Unmarshal(bytes, &res)
	return &res, err
}

var reg registry

func initRegistry() {
	reg = newRegistry()
}

// RegisterDefault registers a ResourceKind as the default version if no version is specified
func RegisterDefault[T Resource](apiVersion string, kind Kind, resourceKind ResourceKind[T]) {
	reg.addKind(apiVersion, kind, wrap(resourceKind))
	reg.setDefaultVersion(kind, apiVersion)
}

// Register registers a ResourceKind so that it can be created and managed by BindPlane
func Register[T Resource](apiVersion string, kind Kind, resourceKind ResourceKind[T]) {
	reg.addKind(apiVersion, kind, wrap(resourceKind))
}

// RegisterKind registers a kind like KindAgent that doesn't implement Resource
func RegisterKind(kind Kind) {
	reg.addKindLookup(kind)
}

// ParseKind parses a kind from a specified string parameter, validating that it matches an existing kind. It ignores
// the case of the string parameter and also allows plurals, e.g. configurations => KindConfiguration. KindUnknown is
// returned if that specified kind does not match any known Kinds.
func ParseKind(kind string) Kind {
	lower := strings.ToLower(kind)
	if kind, ok := reg.kindLookup[lower]; ok {
		return kind
	}
	return KindUnknown
}

// NewEmptyResource returns a new empty resource of the specified Kind
func NewEmptyResource(kind Kind) (Resource, error) {
	k, err := reg.kind(kind)
	if err != nil {
		return nil, err
	}
	return k.NewEmptyResource(), nil
}

// NewEmptyVersionResource returns a new empty resource of the specified Kind
func NewEmptyVersionResource(version string, kind Kind) (Resource, error) {
	k, err := reg.versionKind(version, kind)
	if err != nil {
		return nil, err
	}
	return k.NewEmptyResource(), nil
}

// ParseOne parses an AnyResource and ensures that it is the correct type
func ParseOne[T Resource](r *AnyResource) (resource T, err error) {
	parsed, err := ParseResource(r)
	if err != nil {
		return resource, err
	}
	result, ok := parsed.(T)
	if !ok {
		return resource, fmt.Errorf("resource of kind %s is not the expected type", r.GetKind())
	}
	return result, nil
}

// Parse parses an AnyResource and ensures that it is the correct type
func Parse[T Resource](r []*AnyResource) ([]T, error) {
	parsed, err := ParseResources(r)
	if err != nil {
		return nil, err
	}
	result := make([]T, 0, len(parsed))
	for _, r := range parsed {
		item, ok := r.(T)
		if !ok {
			return nil, fmt.Errorf("resource of kind %s is not the expected type", r.GetKind())
		}
		result = append(result, item)
	}
	return result, nil
}

// ParseResource parses an AnyResource into a Resource
func ParseResource(r *AnyResource) (Resource, error) {
	resource, err := NewEmptyVersionResource(r.APIVersion, r.GetKind())
	if err != nil {
		return nil, err
	}
	return parseResource(r, resource, false)
}

// ParseResourceStrict maps the Spec of the provided resource to a specific type of Resource
// It will error if there are any unused keys
func ParseResourceStrict(r *AnyResource) (Resource, error) {
	resource, err := NewEmptyVersionResource(r.APIVersion, r.GetKind())
	if err != nil {
		return nil, err
	}
	return parseResource(r, resource, true)
}

// ParseResources parses a slice of AnyResource into a Resource
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

func parseResource[T Resource](r *AnyResource, instance T, errorUnused bool) (T, error) {
	if r.Kind != instance.GetKind() {
		return instance, fmt.Errorf("invalid resource kind: %s", r.Kind)
	}

	if errorUnused {
		// If we are doing a "strict" unmarshal, we need to remove the "__typename" fields.
		// These fields are populated on the frontend from GraphQL, and is a pain to strip from the payload there,
		// so instead we accept those "__typename" fields and just ignore them, even in the case where
		// we are strict unmarshalling.
		stripTypenameStringMap(r.Spec)
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

// Clone returns a copy of a resource by using json marshal/unmarshal.
func Clone[T Resource](r T) (result T, err error) {
	value := reflect.ValueOf(r)
	if value.Kind() == reflect.Ptr && value.IsNil() {
		return
	}
	bytes, err := jsoniter.Marshal(r)
	if err != nil {
		return
	}
	res, err := NewEmptyResource(r.GetKind())
	if err != nil {
		return
	}
	err = jsoniter.Unmarshal(bytes, &res)
	result = res.(T)
	return
}

// IsVersionedKind returns true if the kind is versioned
func IsVersionedKind(kind Kind) bool {
	switch kind {
	case KindSource, KindSourceType, KindProcessor, KindProcessorType, KindDestination, KindDestinationType:
		return true
	case KindConfiguration:
		// Configuration is a special case. It is versioned but we don't want to automatically version it. We need to
		// manually archive configurations on rollout. This will reduce the number of versions created during editing.
		return false
	}
	return false
}

// HasVersionKind returns true if the kind has a version. This is slightly different than IsVersionedKind for the
// Configuration resource. Configuration has a version, but is not auto-versioned so is not considered true for
// IsVersionedKind.
func HasVersionKind(kind Kind) bool {
	return IsVersionedKind(kind) || kind == KindConfiguration
}

// ----------------------------------------------------------------------

type registry struct {
	// APIVersion => Kind => ResourceKind
	kinds map[string]map[Kind]ResourceKind[Resource]

	// kindLookup is a map from lowercase name => Kind, including the plural form by adding an "s" to the end of the name.
	// This is used by ParseKind.
	kindLookup map[string]Kind

	// defaultVersion is a map of Kind to the default version of that Kind
	defaultVersion map[Kind]string
}

func newRegistry() registry {
	return registry{
		kinds:          map[string]map[Kind]ResourceKind[Resource]{},
		kindLookup:     map[string]Kind{},
		defaultVersion: map[Kind]string{},
	}
}

func (r registry) addKind(apiVersion string, kind Kind, resourceKind ResourceKind[Resource]) {
	av, ok := r.kinds[apiVersion]
	if !ok {
		av = map[Kind]ResourceKind[Resource]{}
		reg.kinds[apiVersion] = av
	}

	// add for ParseResource
	av[kind] = wrap(resourceKind)

	// add for ParseKind
	r.addKindLookup(kind)
}

func (r registry) addKindLookup(kind Kind) {
	key := strings.ToLower(string(kind))
	plural := fmt.Sprintf("%ss", key)
	r.kindLookup[key] = kind
	r.kindLookup[plural] = kind
}

// normalizeApiVersion ensure that an apiVersion is specified
func (r registry) ensureAPIVersion(apiVersion string, kind Kind) string {
	switch apiVersion {
	case version.V1Alpha:
		apiVersion = version.V1
	case version.V1Beta:
		apiVersion = version.V1
	case "":
		if av, ok := reg.defaultVersion[kind]; ok {
			apiVersion = av
		} else {
			apiVersion = version.V1
		}
	}
	return apiVersion
}

func (r registry) kind(kind Kind) (ResourceKind[Resource], error) {
	return r.versionKind("", kind)
}

func (r registry) versionKind(apiVersion string, kind Kind) (ResourceKind[Resource], error) {
	if _, ok := r.kindLookup[strings.ToLower(string(kind))]; !ok {
		return nil, fmt.Errorf("unknown resource kind: %s", kind)
	}
	apiVersion = r.ensureAPIVersion(apiVersion, kind)
	av, ok := r.kinds[apiVersion]
	if !ok {
		return nil, fmt.Errorf("unknown apiVersion: %s", apiVersion)
	}
	k, ok := av[kind]
	if !ok {
		return nil, fmt.Errorf("unknown apiVersion, kind: %s, %s", apiVersion, kind)
	}
	return k, nil
}

func (r registry) setDefaultVersion(kind Kind, apiVersion string) {
	reg.defaultVersion[kind] = apiVersion
}

const typenameField = "__typename"

// stripTypenameAny strips the "__typename" field out of the provided value.
func stripTypenameAny(v any) {
	switch typedV := v.(type) {
	case map[string]any:
		stripTypenameStringMap(typedV)
	case map[any]any:
		stripTypenameAnyMap(typedV)
	case []any:
		stripTypenameArray(typedV)
	}
}

// stripTypenameStringMap strips the "__typename" field out of the map (recursively).
func stripTypenameStringMap(m map[string]any) {
	delete(m, typenameField)
	for _, v := range m {
		stripTypenameAny(v)
	}
}

// stripTypenameStringMap strips the "__typename" field out of the map (recursively).
func stripTypenameAnyMap(m map[any]any) {
	delete(m, typenameField)
	for _, v := range m {
		stripTypenameAny(v)
	}
}

// stripTypenameArray strips the "__typename" field out of submaps of the array (recursively).
func stripTypenameArray(s []any) {
	for _, v := range s {
		stripTypenameAny(v)
	}
}

// ----------------------------------------------------------------------

// genericKind allows us to store specific implementations of ResourceKind[T] in the registry which manages a
// map[Kind]ResourceKind[Resource]. If though T is a Resource, it doesn't match ResourceKind[Resource].
type genericKind[T Resource] struct {
	kind ResourceKind[T]
}

// NewEmptyResource returns a new empty resource of the specified type
func (s *genericKind[T]) NewEmptyResource() Resource { return s.kind.NewEmptyResource() }

func wrap[T Resource](kind ResourceKind[T]) ResourceKind[Resource] {
	return &genericKind[T]{kind: kind}
}
