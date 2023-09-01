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

package search

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/observiq/bindplane-op/model/search"
	modelSearch "github.com/observiq/bindplane-op/model/search"
)

var tracer = otel.Tracer("store/search")

// Index provides a query interface for indexed resources. A separate index will be used for each resource type.
// Currently the functions on Index do not produce an error but some other implementation could.
//
//go:generate mockery --name=Index --filename=mock_index.go --structname=MockIndex
type Index interface {
	// Add or updates an indexed resource
	Upsert(context.Context, search.Indexed) error

	// Remove an index resource
	Remove(context.Context, search.Indexed) error

	// Query returns ids that are an exact match of the specified query
	Search(ctx context.Context, query *Query) ([]string, error)

	// Matches returns true if the specified indexID matches the query exactly
	Matches(ctx context.Context, query *Query, indexID string) bool

	// Suggestions are either names of fields or Labels or will be values if a field or label is already specified. If
	// there are no suggestions that match or there is one suggestion that is an exact match, no suggestions are returned.
	Suggestions(ctx context.Context, query *Query) ([]*Suggestion, error)

	// Select returns the matching ids
	Select(ctx context.Context, Labels map[string]string) []string
}

type index struct {
	name      string
	documents map[string]*Document
	facets    *facets
	mtx       sync.RWMutex
}

// NewInMemoryIndex returns a new implementation of the the search Index interface that stores the index in memory
func NewInMemoryIndex(name string) Index {
	return &index{
		name:      name,
		documents: map[string]*Document{},
		facets:    newFacets(),
	}
}

var _ Index = (*index)(nil)

func (i *index) Upsert(ctx context.Context, indexed modelSearch.Indexed) error {
	_, span := tracer.Start(ctx, "index/Upsert")
	defer span.End()

	i.mtx.Lock()
	defer i.mtx.Unlock()

	doc := newDocument(indexed)

	// TODO(andy): if the Document values hasn't changed, skip the facets update
	i.facets.Upsert(indexed)
	i.documents[doc.id] = doc

	return nil
}

func (i *index) Remove(ctx context.Context, indexed modelSearch.Indexed) error {
	_, span := tracer.Start(ctx, "index/Upsert")
	defer span.End()

	i.mtx.Lock()
	defer i.mtx.Unlock()

	id := indexed.IndexID()
	delete(i.documents, id)
	i.facets.Remove(id)
	return nil
}

func (i *index) Search(ctx context.Context, query *Query) ([]string, error) {
	ctx, span := tracer.Start(ctx, "index/Search")
	defer span.End()

	i.mtx.RLock()
	defer i.mtx.RUnlock()

	span.SetAttributes(
		attribute.String("bindplane.index.query", query.Original),
		attribute.String("bindplane.index.name", i.name),
		attribute.Int("bindplane.index.size", len(i.documents)),
	)

	// TODO(andy): optimize this
	var results []string
	for _, token := range query.Tokens {
		// apply each token separately, reducing the results each time
		results = i.tokenMatches(token, results)
	}

	return results, nil
}

// Matches returns true if the specified indexID matches the query
func (i *index) Matches(ctx context.Context, query *Query, indexID string) bool {
	_, span := tracer.Start(ctx, "index/Matches")
	defer span.End()

	i.mtx.RLock()
	defer i.mtx.RUnlock()

	// TODO(andy): optimize this
	// results := []string{indexID}
	// for _, token := range query.Tokens {
	// 	// apply each token separately, reducing the results each time
	// 	results = i.tokenMatches(token, results)
	// }

	// return len(results) > 0

	doc, ok := i.documents[indexID]
	if !ok {
		return false
	}

	for _, token := range query.Tokens {
		if !tokenMatchesDocument(token, doc) {
			return false
		}
	}

	return true
}

func (i *index) Suggestions(ctx context.Context, query *Query) ([]*Suggestion, error) {
	_, span := tracer.Start(ctx, "index/Suggestions")
	defer span.End()

	i.mtx.RLock()
	defer i.mtx.RUnlock()

	var tokenSuggestions []*Suggestion

	// find suggestions
	lastToken := query.LastToken()
	if lastToken == nil {
		return []*Suggestion{}, nil
	}
	if lastToken.Name == "" {
		// complete against names
		tokenSuggestions = i.facets.NameSuggestions(lastToken.Operator, lastToken.Value)
	} else {
		// complete against values
		tokenSuggestions = i.facets.ValueSuggestions(lastToken.Operator, lastToken.Name, lastToken.Value)
	}

	// apply the tokenSuggestions to the query to form the final suggestions
	suggestions := []*Suggestion{}
	for _, s := range tokenSuggestions {
		suggestions = append(suggestions, &Suggestion{
			Label: s.Label,
			Query: query.ApplySuggestion(s),
			Score: s.Score,
		})
	}
	SortSuggestions(suggestions)

	return suggestions, nil
}

// Select returns the matching ids
func (i *index) Select(ctx context.Context, selector map[string]string) []string {
	_, span := tracer.Start(ctx, "index/Select")
	defer span.End()

	i.mtx.RLock()
	defer i.mtx.RUnlock()

	results := []string{}
	for _, doc := range i.documents {
		if i.selectorMatchesDocument(selector, doc) {
			results = append(results, doc.id)
		}
	}
	return results
}

// selectorMatchesDocument returns true if all of the Labels in the selector match the Document
func (i *index) selectorMatchesDocument(selector map[string]string, doc *Document) bool {
	for name, value := range selector {
		if doc.originalLabels[name] != value {
			return false
		}
	}
	return true
}

// tokenMatches returns the ids matching the specified token. If ids is specified, we only look at those ids.
func (i *index) tokenMatches(token *QueryToken, ids []string) []string {
	if token.Empty() {
		return ids
	}
	if ids == nil {
		return i.tokenMatchesAny(token)
	}
	// if we have ids, but there are none left, just return the empty list
	if len(ids) == 0 {
		return ids
	}

	results := []string{}
	for _, id := range ids {
		doc, ok := i.documents[id]
		if ok && tokenMatchesDocument(token, doc) {
			results = append(results, doc.id)
		}
	}
	return results
}

// tokenMatches returns the ids matching the specified token
func (i *index) tokenMatchesAny(token *QueryToken) []string {
	results := []string{}
	for _, doc := range i.documents {
		if tokenMatchesDocument(token, doc) {
			results = append(results, doc.id)
		}
	}
	return results
}

// tokenMatchesDocument checks to see if a single token matches the specified Document.
func tokenMatchesDocument(token *QueryToken, doc *Document) bool {
	if doc == nil {
		// highly unexpected, but not a match
		return false
	}
	// In several places we check for result != token.IsNegated(). It may not be obvious, but we are simply inverting the
	// result when IsNegated is true. If result is true, but negated is false, we also return true. If the result is
	// false, but negated is true, we return true. If both are true or both are false, we return false.
	if token.Name == "" {
		return textMatchesDocument(token.Value, doc) != token.IsNegated()
	}
	if token.Value == "" {
		return fieldExistsMatchesDocument(token.Name, doc) != token.IsNegated()
	}
	return fieldMatchesDocument(token, doc) != token.IsNegated()
}

func textMatchesDocument(query string, doc *Document) bool {
	return strings.Contains(doc.values, query)
}

func fieldExistsMatchesDocument(field string, doc *Document) bool {
	_, ok := doc.Labels[field]
	if ok {
		return true
	}
	_, ok = doc.fields[field]
	return ok
}

func fieldMatchesDocument(token *QueryToken, doc *Document) bool {
	field := token.Name

	value, ok := doc.Labels[field]
	if ok && value == token.Value {
		return true
	}
	values, ok := doc.fields[field]
	return ok && values.contains(token.Value)
}

// Field is a helper function to search the given index with a field:value pair.
func Field(ctx context.Context, index Index, field, value string) ([]string, error) {
	query := ParseQuery(fmt.Sprintf("%s:%s", field, value))
	return index.Search(ctx, query)
}
