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

// Package model provides functions to convert models to GraphQL models
package model

import (
	"testing"

	"github.com/observiq/bindplane-op/model"
	"github.com/stretchr/testify/require"
)

func Test_clearCurrentAgentUpgradeError(t *testing.T) {
	type args struct {
		cur *model.Agent
	}
	tests := []struct {
		name string
		args args
		want *model.Agent
	}{
		{
			"clears upgrade error",
			args{
				cur: &model.Agent{Upgrade: &model.AgentUpgrade{Error: "some-error"}},
			},
			&model.Agent{Upgrade: &model.AgentUpgrade{}},
		},
		{
			"no panic for nil AgentUpgrade",
			args{
				cur: &model.Agent{},
			},
			&model.Agent{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ClearCurrentAgentUpgradeError(tt.args.cur)
			require.Equal(t, tt.want, tt.args.cur)
		})
	}
}
