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
	"strings"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/validation"

	jsoniter "github.com/json-iterator/go"
	modelValidation "github.com/observiq/bindplane-op/model/validation"
)

const (
	// LabelBindPlaneAgentID is the label name for agent id
	LabelBindPlaneAgentID = "bindplane/agent-id"

	// LabelBindPlaneAgentName is the label name for agent name
	LabelBindPlaneAgentName = "bindplane/agent-name"

	// LabelBindPlaneAgentType is the label for the agent type
	LabelBindPlaneAgentType = "bindplane/agent-type"

	// LabelBindPlaneAgentVersion is the label name for agent version
	LabelBindPlaneAgentVersion = "bindplane/agent-version"

	// LabelBindPlaneAgentHost is the label name for agent host
	LabelBindPlaneAgentHost = "bindplane/agent-host"

	// LabelBindPlaneAgentOS is the label name for agent operating system
	LabelBindPlaneAgentOS = "bindplane/agent-os"

	// LabelBindPlaneAgentArch is the label name for agent cpu architecture
	LabelBindPlaneAgentArch = "bindplane/agent-arch"

	// LabelAgentContainerPlatform is the label name for specifying a container platform (k8s, openshift, etc...)
	LabelAgentContainerPlatform = "container-platform"
)

// Labeled TODO(doc)
type Labeled interface {
	// GetLabels TODO(doc)
	GetLabels() Labels
}

// Labels is a wrapper around Kubernetes labels.Set struct, which is just a type definition for map[string]string.
type Labels struct {
	labels.Set `json:"-" yaml:",inline"`
}

// LabelsFromMap will create a set of labels from a map of name/value pairs. It will validate that the names and values
// conform to the requirements, matching those of kubernetes labels. See
// https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/. Valid labels will be added to the Labels
// returned and any invalid labels will produce an error. This makes it possible for callers to ignore the errors and
// accept any valid values specified.
func LabelsFromMap(labels map[string]string) (Labels, error) {
	var err error

	valid := map[string]string{}
	for name, value := range labels {
		if errs := validation.IsQualifiedName(name); len(errs) > 0 {
			err = errors.Join(err, fmt.Errorf("%s is not a valid label name: %s", name, strings.Join(errs, "; ")))
			continue
		}
		if errs := validation.IsValidLabelValue(value); len(errs) > 0 {
			err = errors.Join(err, fmt.Errorf("%s is not a valid label value: %s", value, strings.Join(errs, "; ")))
			continue
		}
		valid[name] = value
	}

	return Labels{valid}, err
}

// LabelsFromValidatedMap returns a set of labels from map that is assumed to already be validated.
func LabelsFromValidatedMap(labels map[string]string) Labels {
	return Labels{labels}
}

// LabelsFromSelector TODO(doc)
func LabelsFromSelector(selector string) (Labels, error) {
	// ConvertSelectorToLabelsMap validates the labels provided
	set, err := labels.ConvertSelectorToLabelsMap(selector)
	if err != nil {
		return Labels{}, err
	}
	return Labels{set}, nil
}

// LabelsFromMerge merges new labels into existing labels.
// Any labels with blank values in the merged labels will be removed
func LabelsFromMerge(existing, new Labels) Labels {
	labels := Labels{labels.Merge(existing.Set, new.Set)}
	labels.removeEmptyValues()
	return labels
}

// MakeLabels returns a new, empty Labels object
func MakeLabels() Labels {
	return Labels{map[string]string{}}
}

func (l *Labels) removeEmptyValues() {
	for key, value := range l.Set {
		if value == "" {
			delete(l.Set, key)
		}
	}
}

// AsMap returns the labels as a map of name/value pairs
func (l *Labels) AsMap() map[string]string {
	return l.Set
}

// Conflicts returns true if the specified set of labels has a label with the same name as this set of labels but with a
// different value.
func (l Labels) Conflicts(o Labels) bool {
	return labels.Conflicts(l.Set, o.Set)
}

// Custom returns the custom labels, i.e. labels not starting with "bindplane/"
func (l Labels) Custom() Labels {
	return l.filtered(false)
}

// BindPlane returns the BindPlane labels, i.e. labels starting with "bindplane/"
func (l Labels) BindPlane() Labels {
	return l.filtered(true)
}

// filtered returns labels with or without the "bindplane/" prefix based on hasBindPlanePrefix
func (l Labels) filtered(hasBindPlanePrefix bool) Labels {
	custom := map[string]string{}
	for name, value := range l.Set {
		if hasBindPlanePrefix == strings.HasPrefix(name, "bindplane/") {
			custom[name] = value
		}
	}
	// this is a subset of labels, so we know that they are validated
	return LabelsFromValidatedMap(custom)
}

// ----------------------------------------------------------------------
//
// without custom marshalling, we end up with "labels": { "Set": {} } and we want to avoid the "Set" but json:",inline"
// isn't a thing

// MarshalJSON marshals the Labels as jsoniter. An empty Labels will be marshalled as `{}`
func (l *Labels) MarshalJSON() ([]byte, error) {
	// serialize null as empty
	if l.Set == nil {
		return []byte("{}"), nil
	}
	return jsoniter.Marshal(l.Set)
}

// UnmarshalJSON unmarshals JSON bytes into the Label's internal Set
func (l *Labels) UnmarshalJSON(b []byte) error {
	return jsoniter.Unmarshal(b, &l.Set)
}

func (l *Labels) validate(errs modelValidation.Errors) {
	_, err := LabelsFromMap(l.AsMap())
	errs.Add(err)
}
