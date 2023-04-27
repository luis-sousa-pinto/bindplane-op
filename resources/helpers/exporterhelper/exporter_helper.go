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
// given exporter. This configuration is rendered inline.
func BPRenderOtelRetryOnFailureConfig(
	enabled bool,
	// We type these as any because BindPlane validation passes in
	// ints while use in the gotemplate passes in floats.
	initialInterval,
	maxInterval,
	maxElapsedTime any) string {
	if !enabled {
		return "retry_on_failure: { enabled: false }"
	}

	return fmt.Sprintf(
		"retry_on_failure: { enabled: true, initial_interval: %ds, max_interval: %ds, max_elapsed_time: %ds }",
		anyToInt64(initialInterval),
		anyToInt64(maxInterval),
		anyToInt64(maxElapsedTime),
	)
}

// BPRenderOtelSendingQueueConfig renders the sending_queue config for the
// given exporter. This configuration is rendered inline.
func BPRenderOtelSendingQueueConfig(
	enabled,
	persistenceEnabled bool,
	storageID string,
	numConsumers,
	queueSize any) string {
	if !enabled {
		// No sending queue
		return "sending_queue: { enabled: false }"
	}

	if !persistenceEnabled {
		// In-memory buffer sending queue
		return fmt.Sprintf(
			"sending_queue: { enabled: true, num_consumers: %d, queue_size: %d }",
			anyToInt64(numConsumers),
			anyToInt64(queueSize),
		)
	}

	// Disk buffer sending queue
	return fmt.Sprintf(
		"sending_queue: { enabled: true, num_consumers: %d, queue_size: %d, storage: %q }",
		anyToInt64(numConsumers),
		anyToInt64(queueSize),
		storageID,
	)
}

// anyToInt64 converts an int or float (typed as any) to int64
func anyToInt64(floatOrInt any) int64 {
	switch v := floatOrInt.(type) {
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case float32:
		return int64(v)
	case float64:
		return int64(v)
	}

	return 0
}
