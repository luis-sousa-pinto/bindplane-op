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
	"testing"

	"github.com/stretchr/testify/require"
)

func testIndex() *index {
	return NewInMemoryIndex("test").(*index)
}

func TestIndexFieldExistsQuery(t *testing.T) {
	runFieldExistsQueryTests(t, testIndex())
}

func TestIndexQuotedQuery(t *testing.T) {
	runQuotedQueryTests(t, testIndex())
}

func TestIndexMatches(t *testing.T) {
	runMatchesTests(t, testIndex())
}

func TestIndexRemove(t *testing.T) {
	runRemoveTests(t, testIndex())
}

func TestIndexSuggestions(t *testing.T) {
	runSuggestionsTests(t, testIndex())
}

func TestIndexTokenMatchesNilDocument(t *testing.T) {
	query := ParseQuery("test")
	result := tokenMatchesDocument(query.LastToken(), nil)
	require.False(t, result)
}

func TestIndexSearch(t *testing.T) {
	runSearchTests(t, testIndex())
}

func TestUpsertRemovesOldLabels(t *testing.T) {
	runRemovesOldLabelsTest(t, testIndex())
}

func TestIndexSelect(t *testing.T) {
	runSelectTests(t, testIndex())
}
