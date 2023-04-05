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

func Test_makeIndent(t *testing.T) {
	type args struct {
		n int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"0",
			args{0},
			"",
		},
		{
			"4",
			args{4},
			// 8 spaces
			"        ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeIndent(tt.args.n); got != tt.want {
				t.Errorf("makeIndent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBPRenderOtelRetryOnFailureConfig(t *testing.T) {
	type args struct {
		enabled         bool
		initialInterval int
		maxInterval     int
		queueSize       int
		nindent         int
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
				7,
			},
			`retry_on_failure:
              enabled: false
`,
		},
		{
			"enabled",
			args{
				true,
				1,
				2,
				3,
				7,
			},
			`retry_on_failure:
              enabled: true
              initial_interval: 1
              max_interval: 2
              max_elapsed_time: 3
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BPRenderOtelRetryOnFailureConfig(tt.args.enabled, tt.args.initialInterval, tt.args.maxInterval, tt.args.queueSize, tt.args.nindent)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestBPRenderOtelSendingQueueConfig(t *testing.T) {
	type args struct {
		enabled      bool
		numConsumers int
		queueSize    int
		nindent      int
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
				7,
			},
			`sending_queue:
              enabled: false
`,
		},
		{
			"enabled",
			args{
				true,
				1,
				2,
				7,
			},
			`sending_queue:
              enabled: true
              num_consumers: 1
              queue_size: 2
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BPRenderOtelSendingQueueConfig(tt.args.enabled, tt.args.numConsumers, tt.args.queueSize, tt.args.nindent)
			require.Equal(t, tt.want, got)
		})
	}

}
