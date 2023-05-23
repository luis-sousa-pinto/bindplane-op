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

// Package config contains the top level configuration structures and logic
package config

import (
	"fmt"
	"time"

	modelversion "github.com/observiq/bindplane-op/model/version"
)

const (
	// DefaultOutput is the default output format of the CLI
	DefaultOutput = "table"

	// EnvDevelopment should be used for development and uses debug logging and normal gin request logging to stdout.
	EnvDevelopment = "development"

	// EnvTest should be used for tests and uses debug logging with json gin request logging to the log file.
	EnvTest = "test"

	// EnvProduction the the default and should be used in production and uses info logging with json gin request logging to the log file.
	EnvProduction = "production"

	// DefaultRolloutsInterval is the interval at which rollouts are updated
	DefaultRolloutsInterval = 5 * time.Second
)

// Config is the configuration of BindPlane
type Config struct {
	// ProfileName is the name of the profile associated with this config file if there is one
	ProfileName string `mapstructure:"name,omitempty" yaml:"name,omitempty"`

	// APIVersion is the version of the configuration file
	APIVersion string `mapstructure:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`

	// Env is the environment of the service
	Env string `mapstructure:"env,omitempty" yaml:"env,omitempty"`

	// Output is the output format of the CLI
	Output string `mapstructure:"output,omitempty" yaml:"output,omitempty"`

	// Offline mode indicates if the server should be considered offline. An offline server will not attempt to contact
	// any other services. It will still allow agents to connect and serve api requests.
	Offline bool `mapstructure:"offline,omitempty" yaml:"offline,omitempty"`

	// RolloutsInterval is the interval at which rollouts' progress is updated.
	RolloutsInterval time.Duration `mapstructure:"rolloutsInterval" yaml:"rolloutsInterval,omitempty"`

	// Auth is the configuration for authentication
	Auth Auth `mapstructure:"auth,omitempty" yaml:"auth,omitempty"`

	// Network is the configuration for networking
	Network Network `mapstructure:"network" yaml:"network,omitempty"`

	// AgentVersions is the configuration for agent versions
	AgentVersions AgentVersions `mapstructure:"agentVersions,omitempty" yaml:"agentVersions,omitempty"`

	// Store is the configuration for storage
	Store Store `mapstructure:"store,omitempty" yaml:"store,omitempty"`

	// Tracing is the tracer configuration for the server
	Tracing Tracing `mapstructure:"tracing,omitempty" yaml:"tracing,omitempty"`

	// Logging configuration for the logger
	Logging Logging `yaml:"logging,omitempty" mapstructure:"logging,omitempty"`
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if err := c.Auth.Validate(); err != nil {
		return fmt.Errorf("failed to validate auth: %w", err)
	}

	if err := c.Network.Validate(); err != nil {
		return fmt.Errorf("failed to validate network: %w", err)
	}

	if err := c.AgentVersions.Validate(); err != nil {
		return fmt.Errorf("failed to validate agent versions: %w", err)
	}

	if err := c.Store.Validate(); err != nil {
		return fmt.Errorf("failed to validate store: %w", err)
	}

	if err := c.Tracing.Validate(); err != nil {
		return fmt.Errorf("failed to validate tracing: %w", err)
	}

	if err := c.Logging.Validate(); err != nil {
		return fmt.Errorf("failed to validate logging: %w", err)
	}

	return nil
}

// BindPlaneURL returns the BindPlane URL
func (c *Config) BindPlaneURL() string {
	return c.Network.ServerURL()
}

// BindPlaneInsecureSkipVerify returns the BindPlane InsecureSkipVerify
func (c *Config) BindPlaneInsecureSkipVerify() bool {
	return c.Network.InsecureSkipVerify
}

// NewConfig returns a new config
func NewConfig() *Config {
	return &Config{
		APIVersion: modelversion.V1,
	}
}
