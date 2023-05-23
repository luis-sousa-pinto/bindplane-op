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

// HasUniqueKey is an interface that provides access to a unique key for an item. For Agents this is the ID and for
// Resources this is the Name.
type HasUniqueKey interface {
	UniqueKey() string
}

// UniqueKeys returns the list of unique keys for the specified resources
func UniqueKeys[S ~[]T, T HasUniqueKey](resources S) []string {
	keys := make([]string, 0, len(resources))
	for _, r := range resources {
		keys = append(keys, r.UniqueKey())
	}
	return keys
}
