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

package config

import (
	"errors"
	"time"
)

// DefaultSyncInterval is the default interval at which agent-versions will be synchronized with GitHub.
const DefaultSyncInterval = 1 * time.Hour

// AgentVersions is the configuration for serving and checking agent versions.
type AgentVersions struct {
	SyncInterval time.Duration `mapstructure:"syncInterval,omitempty" yaml:"syncInterval,omitempty"`
}

// Validate validates the agent versions configuration.
func (c *AgentVersions) Validate() error {
	if c.SyncInterval < time.Hour {
		return errors.New("sync interval must be at least 1 hour")
	}

	return nil
}
