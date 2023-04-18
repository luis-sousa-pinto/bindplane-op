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

// Package observiq provides models and functions for interacting with managed observIQ OTel agents.
package observiq

import (
	"crypto/sha256"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
)

// Names of configuration files in the agent
const (
	CollectorFilename = "collector.yaml"
	LoggingFilename   = "logging.yaml"
	ManagerFilename   = "manager.yaml"
)

// AgentConfiguration represents the configuration files of the observiq-agent
type AgentConfiguration struct {
	// Collector is the opentelemetry configuration in collector.yaml
	Collector string `mapstructure:"collector"`

	// Logging isn't currently managed by BindPlane and is stored as a string for reference.
	Logging string `mapstructure:"logging"`

	// Manager is the agent configuration in manager.yaml
	Manager *ManagerConfig `mapstructure:"manager"`
}

// RawAgentConfiguration represents the raw configuration from the agent
type RawAgentConfiguration struct {
	Collector []byte
	Logging   []byte
	Manager   []byte
}

// ManagerConfig is the unmarshaled contents of manager.yaml
// This comes from https://github.com/observIQ/observiq-otel-collector/blob/main/opamp/config.go
type ManagerConfig struct {
	Endpoint  string            `mapstructure:"endpoint" json:"endpoint" yaml:"endpoint,omitempty"`
	SecretKey string            `mapstructure:"secret_key" json:"secret_key" yaml:"secret_key,omitempty"`
	AgentID   string            `mapstructure:"agent_id" json:"agent_id" yaml:"agent_id,omitempty"`
	Labels    string            `mapstructure:"labels" json:"labels" yaml:"labels,omitempty"`
	AgentName string            `mapstructure:"agent_name" json:"agent_name" yaml:"agent_name,omitempty"`
	TLS       *ManagerTLSConfig `mapstructure:"tls_config" json:"tls_config" yaml:"tls_config,omitempty"`
}

// ManagerTLSConfig is tls configuration for the collector's manager config
type ManagerTLSConfig struct {
	InsecureSkipVerify bool    `mapstructure:"insecure_skip_verify" json:"insecure_skip_verify" yaml:"insecure_skip_verify,omitempty"`
	KeyFile            *string `mapstructure:"key_file" json:"key_file" yaml:"key_file,omitempty"`
	CertFile           *string `mapstructure:"cert_file" json:"cert_file" yaml:"cert_file,omitempty"`
	CAFile             *string `mapstructure:"ca_file" json:"ca_file" yaml:"ca_file,omitempty"`
}

func parseManagerConfig(bytes []byte) (*ManagerConfig, error) {
	if bytes == nil {
		return nil, nil
	}
	var mc ManagerConfig
	if err := yaml.Unmarshal(bytes, &mc); err != nil {
		return nil, err
	}
	return &mc, nil
}

// DecodeAgentConfiguration will map a generic interface{} to an AgentConfiguration
func DecodeAgentConfiguration(configuration interface{}) (*AgentConfiguration, error) {
	result := &AgentConfiguration{}
	if err := mapstructure.Decode(configuration, result); err != nil {
		return nil, err
	}
	return result, nil
}

// Parse will parse a raw configuration of []byte received from the agent to a structured configuration
func (raw *RawAgentConfiguration) Parse() (*AgentConfiguration, error) {
	collector := string(raw.Collector)
	logging := string(raw.Logging)
	manager, err := parseManagerConfig(raw.Manager)
	if err != nil {
		return nil, fmt.Errorf("unable to parse manager config: %w", err)
	}
	return &AgentConfiguration{
		Collector: collector,
		Logging:   logging,
		Manager:   manager,
	}, nil
}

func computeHash(contents ...[]byte) []byte {
	h := sha256.New()
	for _, b := range contents {
		h.Write(b)
	}
	return h.Sum(nil)
}

// Hash returns a sha256 hash of the 3 configuration files in alphabetical order.
func (raw *RawAgentConfiguration) Hash() []byte {
	return computeHash(
		raw.Collector,
		raw.Logging,
		raw.Manager,
	)
}

// ApplyUpdates applies a partial configuration to a configuration, returning a new configuration and leaving the
// existing configuration unmodified.
func (raw *RawAgentConfiguration) ApplyUpdates(update *RawAgentConfiguration) RawAgentConfiguration {
	clone := *raw
	if update == nil {
		return clone
	}
	if update.Logging != nil {
		clone.Logging = update.Logging
	}
	if update.Collector != nil {
		clone.Collector = update.Collector
	}
	if update.Manager != nil {
		clone.Manager = update.Manager
	}
	return clone
}

func marshalConfig(config interface{}) []byte {
	// note that we ignore marshal errors. marshal errors occur when there are key conflicts and we control the keys in
	// the struct definition. tests ensure that we can fully marshal a configuration without an error.
	bytes, _ := yaml.Marshal(config)
	return bytes
}

// Raw will marshal a structured configuration into a raw configuration of []byte that can be sent to the agent.
func (c *AgentConfiguration) Raw() (raw RawAgentConfiguration) {
	if c.Collector != "" {
		raw.Collector = []byte(c.Collector)
	}
	if c.Logging != "" {
		raw.Logging = []byte(c.Logging)
	}
	if c.Manager != nil {
		raw.Manager = marshalConfig(c.Manager)
	}
	return raw
}

// HasLabels returns true if the existing configuration has a manager.yaml with the same labels
func (c *AgentConfiguration) HasLabels(labels string) bool {
	if c.Manager == nil {
		return labels == ""
	}
	return c.Manager.Labels == labels
}

// ReplaceLabels replaces the labels in the manager.yaml. If manager.yaml doesn't exist an empty one will be created.
func (c *AgentConfiguration) ReplaceLabels(labels string) {
	// if the labels haven't changed, keep the manager the same
	if c.HasLabels(labels) {
		return
	}
	if c.Manager == nil {
		c.Manager = &ManagerConfig{
			Labels: labels,
		}
	} else {
		clone := *c.Manager
		clone.Labels = labels
		c.Manager = &clone
	}
}

// Empty returns true if the configuration has empty collector, logging, and manager configs.
func (c AgentConfiguration) Empty() bool {
	return c.Collector == "" && c.Logging == "" && c.Manager == nil
}

// ComputeConfigurationUpdates returns the modified agent configuration if the agent's AgentConfiguration contains different settings
// for any of the configs. Only the parts of the config that are different are included in the resulting AgentConfiguration.
func ComputeConfigurationUpdates(server *AgentConfiguration, agent *AgentConfiguration) (diff AgentConfiguration) {
	// logging.yaml -- currently ignored and never included in updates

	// collector.yaml -- must match exactly or it is updated

	if server.Collector != agent.Collector {
		diff.Collector = server.Collector
	}

	// manager.yaml -- only requires that the labels be equal because this is currently the only managed portion of that configuration

	if server.Manager == nil {
		// no server manager configuration so no opinion about labels
		return diff
	}

	if agent.Manager == nil {
		// no agent manager configuration to compare so just send a config with labels
		if server.Manager.Labels != "" {
			diff.Manager = &ManagerConfig{
				Labels: server.Manager.Labels,
			}
		}
		return diff
	}

	if !agent.HasLabels(server.Manager.Labels) {
		// start with a clone of the agent manager configuration since we want to preserve the rest of the agent config
		clone := *agent.Manager
		clone.Labels = server.Manager.Labels
		diff.Manager = &clone
	}

	return diff
}
