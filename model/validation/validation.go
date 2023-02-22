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

// Package validation contains functions for validating constraints are met for given strings
package validation

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"
)

// IsName validates the name and adds to Errors if the name is invalid
func IsName(err Errors, name string) {
	if errors := validation.IsValidLabelValue(name); len(errors) > 0 {
		err.Add(fmt.Errorf("%s is not a valid resource name: %s", name, strings.Join(errors, "; ")))
	}
}
