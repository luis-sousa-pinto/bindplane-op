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

package exporterhelper

import "fmt"

// BPRenderOtelRetryOnFailureConfig renders the retry_on_failure config for the
// given exporter.  It does not begin with a newline and the nindent parameter is
// used to indent following lines by that number of tabs (2 spaces).
func BPRenderOtelRetryOnFailureConfig(
	enabled bool,
	// We type these as any because BindPlane validation passes in
	// ints while use in the gotemplate passes in floats.
	initialInterval,
	maxInterval,
	maxElapsedTime any,
	nindent int) string {
	indent := makeIndent(nindent)
	otelConfig := "retry_on_failure:\n"

	if !enabled {
		otelConfig += fmt.Sprintf("%senabled: false\n", indent)
		return otelConfig
	}

	otelConfig += fmt.Sprintf("%senabled: true\n", indent)
	otelConfig += fmt.Sprintf("%sinitial_interval: %v\n", indent, initialInterval)
	otelConfig += fmt.Sprintf("%smax_interval: %v\n", indent, maxInterval)
	otelConfig += fmt.Sprintf("%smax_elapsed_time: %v\n", indent, maxElapsedTime)

	return otelConfig
}

// BPRenderOtelSendingQueueConfig renders the sending_queue config for the
// given exporter.  It does not begin with a newline and the nindent parameter is
// used to indent following lines by that number of tabs (2 spaces).
func BPRenderOtelSendingQueueConfig(
	enabled bool,
	numConsumers,
	queueSize any,
	nindent int) string {

	indent := makeIndent(nindent)
	otelConfig := "sending_queue:\n"

	if !enabled {
		otelConfig += fmt.Sprintf("%senabled: false\n", indent)
		return otelConfig
	}

	otelConfig += fmt.Sprintf("%senabled: true\n", indent)
	otelConfig += fmt.Sprintf("%snum_consumers: %v\n", indent, numConsumers)
	otelConfig += fmt.Sprintf("%squeue_size: %v\n", indent, queueSize)

	return otelConfig
}

func makeIndent(n int) string {
	var indent string
	for range make([]int, n) {
		indent += "  "
	}
	return indent
}
