// Copyright observIQ, Inc.
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

// Package search provides a search interface
package search

// Indexer is a function passed to the functions of Index
type Indexer func(name string, value string)

// Indexed must be implemented by resources that are indexed
type Indexed interface {
	// IndexID returns an ID used to identify the resource that is indexed
	IndexID() string

	// IndexFields should index the fields using the index function
	IndexFields(index Indexer)

	// IndexFields should index the Labels using the index function
	IndexLabels(index Indexer)
}
