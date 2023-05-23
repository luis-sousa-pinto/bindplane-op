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

// Package search provides a search engine with indexing and suggestions for the store
package search

import (
	"strings"

	modelSearch "github.com/observiq/bindplane-op/model/search"
	"golang.org/x/exp/slices"
)

// Document is a single document in the index. It contains the id, fields, and Labels.
type Document struct {
	id     string
	fields map[string]fieldValue
	Labels map[string]string

	// originalLabels contains the Labels without ToLower and is used to exact match in Index.Select
	originalLabels map[string]string

	// values is a collection of the lowercase form of all field values, label names, and label values, separated by
	// newlines for use with general text search
	values string
}

// EmptyDocument returns a Document with the specified id and no fields or Labels
func EmptyDocument(id string) *Document {
	return &Document{
		id:             id,
		fields:         map[string]fieldValue{},
		Labels:         map[string]string{},
		originalLabels: map[string]string{},
		values:         "",
	}
}

func newDocument(indexed modelSearch.Indexed) *Document {
	id := indexed.IndexID()
	doc := EmptyDocument(id)
	indexed.IndexFields(func(name, value string) {
		name = strings.ToLower(name)
		value = strings.ToLower(value)
		doc.AddField(name, value)
	})
	indexed.IndexLabels(func(name, value string) {
		doc.originalLabels[name] = value
		name = strings.ToLower(name)
		value = strings.ToLower(value)
		doc.Labels[name] = value
	})
	doc.values = doc.BuildValues()

	return doc
}

// BuildValues builds a string containing all field values, label names, and label values, separated by newlines.
func (d *Document) BuildValues() string {
	// WriteString and WriteRune will return nil errors, so we ignore them
	var sb strings.Builder
	for _, v := range d.fields {
		v.each(func(sv string) {
			_, _ = sb.WriteString(sv)
			_, _ = sb.WriteRune('\n')
		})
	}
	for n, v := range d.Labels {
		_, _ = sb.WriteString(n)
		_, _ = sb.WriteRune('\n')
		_, _ = sb.WriteString(v)
		_, _ = sb.WriteRune('\n')
	}
	return sb.String()
}

// AddField adds a field to the document. If the field already exists, the value is appended to the existing field.
func (d *Document) AddField(name, value string) {
	if value == "" {
		return
	}
	f, ok := d.fields[name]
	if ok {
		d.fields[name] = f.add(value)
	} else {
		d.fields[name] = fieldSingleValue(value)
	}
}

// ----------------------------------------------------------------------
//
// fieldValue allows us to avoid always storing a []string when we generally have a single value.

type fieldValue interface {
	add(value string) fieldValue
	each(func(string))
	contains(value string) bool
}

type fieldSingleValue string
type fieldMultiValue []string

func (f fieldSingleValue) add(value string) fieldValue {
	return fieldMultiValue{string(f), value}
}
func (f fieldSingleValue) each(callback func(value string)) {
	callback(string(f))
}
func (f fieldSingleValue) contains(value string) bool {
	return string(f) == value
}

func (f fieldMultiValue) add(value string) fieldValue {
	return append(f, value)
}
func (f fieldMultiValue) each(callback func(value string)) {
	for _, v := range f {
		callback(v)
	}
}
func (f fieldMultiValue) contains(value string) bool {
	return slices.Contains(f, value)
}

var _ modelSearch.Indexed = (*Document)(nil)

// IndexID returns the id of the document
func (d *Document) IndexID() string { return d.id }

// IndexFields iterates over the fields of the document and calls the callback for each field
func (d *Document) IndexFields(index modelSearch.Indexer) {
	for n, v := range d.fields {
		v.each(func(sv string) {
			index(n, sv)
		})
	}
}

// IndexLabels iterates over the Labels of the document and calls the callback for each Label
func (d *Document) IndexLabels(index modelSearch.Indexer) {
	for n, v := range d.Labels {
		index(n, v)
	}
}
