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

// Package helpers provides helper functions for use in templates.
package helpers

import (
	"github.com/observiq/bindplane-op/resources/helpers/exporterhelper"
	"github.com/observiq/bindplane-op/resources/helpers/masksensitivedatahelper"
	"github.com/observiq/bindplane-op/resources/helpers/operatorhelper"

	"text/template"
)

// ResourceHelperFuncMap returns a map of helper functions for use in templates.
func ResourceHelperFuncMap() template.FuncMap {
	return map[string]any{
		"bpRenderOtelRetryOnFailureConfig": exporterhelper.BPRenderOtelRetryOnFailureConfig,
		"bpRenderOtelSendingQueueConfig":   exporterhelper.BPRenderOtelSendingQueueConfig,
		"bpRenderMaskRules":                masksensitivedatahelper.BPRenderMaskRules,
		"bpRenderStandardParsingOperator":  operatorhelper.BpRenderStandardParsingOperator,
		"bpMapSeverityNameToNumber":        operatorhelper.BpMapSeverityNameToNumber,
	}
}
