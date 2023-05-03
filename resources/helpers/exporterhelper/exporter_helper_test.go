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

// package exporterHelper provides helper functions to use the Exporter Helper fields in
// BindPlane OP DestinationTypes.
package exporterhelper

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBPRenderOtelRetryOnFailureConfig(t *testing.T) {
	type args struct {
		enabled         bool
		initialInterval int
		maxInterval     int
		queueSize       int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"not enabled",
			args{
				false,
				0,
				0,
				0,
			},
			`retry_on_failure: { enabled: false }`,
		},
		{
			"enabled",
			args{
				true,
				1,
				2,
				3,
			},
			`retry_on_failure: { enabled: true, initial_interval: 1s, max_interval: 2s, max_elapsed_time: 3s }`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BPRenderOtelRetryOnFailureConfig(tt.args.enabled, tt.args.initialInterval, tt.args.maxInterval, tt.args.queueSize)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestBPRenderOtelSendingQueueConfig(t *testing.T) {
	type args struct {
		enabled            bool
		persistenceEnabled bool
		storageID          string
		numConsumers       int
		queueSize          int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "sending queue disabled",
			args: args{
				enabled:            false,
				persistenceEnabled: false,
				storageID:          "file_storage/something",
				numConsumers:       0,
				queueSize:          0,
			},
			want: `sending_queue: { enabled: false }`,
		},
		{
			name: "sending queue disabled, persistence enabled",
			args: args{
				enabled:            false,
				persistenceEnabled: false,
				storageID:          "file_storage/something",
				numConsumers:       0,
				queueSize:          0,
			},
			want: `sending_queue: { enabled: false }`,
		},
		{
			name: "in-memory sending queue",
			args: args{
				enabled:            true,
				persistenceEnabled: false,
				storageID:          "file_storage/something",
				numConsumers:       1,
				queueSize:          2,
			},
			want: `sending_queue: { enabled: true, num_consumers: 1, queue_size: 2 }`,
		},
		{
			name: "in-memory sending queue",
			args: args{
				enabled:            true,
				persistenceEnabled: true,
				storageID:          "file_storage/something",
				numConsumers:       1,
				queueSize:          2,
			},
			want: `sending_queue: { enabled: true, num_consumers: 1, queue_size: 2, storage: "file_storage/something" }`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BPRenderOtelSendingQueueConfig(tt.args.enabled, tt.args.persistenceEnabled, tt.args.storageID, tt.args.numConsumers, tt.args.queueSize)
			require.Equal(t, tt.want, got)
		})
	}

}
