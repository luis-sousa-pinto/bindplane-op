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

package observiq

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReplaceLabels(t *testing.T) {
	tests := []struct {
		name   string
		config *AgentConfiguration
		labels string
		verify func(t *testing.T, config *AgentConfiguration)
	}{
		{
			name: "empty, empty",
			config: &AgentConfiguration{
				Manager: &ManagerConfig{},
			},
			labels: "",
			verify: func(t *testing.T, config *AgentConfiguration) {
				require.NotNil(t, config.Manager)
				require.Equal(t, "", config.Manager.Labels)
			},
		},
		{
			name:   "nil, empty",
			config: &AgentConfiguration{},
			labels: "",
			verify: func(t *testing.T, config *AgentConfiguration) {
				// was nil, remains nil
				require.Nil(t, config.Manager)
			},
		},
		{
			name:   "nil, labels",
			config: &AgentConfiguration{},
			labels: "labels",
			verify: func(t *testing.T, config *AgentConfiguration) {
				// was nil, created
				require.NotNil(t, config.Manager)
				require.Equal(t, "labels", config.Manager.Labels)
			},
		},
		{
			name: "same, same",
			config: &AgentConfiguration{
				Manager: &ManagerConfig{Labels: "same"},
			},
			labels: "same",
			verify: func(t *testing.T, config *AgentConfiguration) {
				require.NotNil(t, config.Manager)
				require.Equal(t, "same", config.Manager.Labels)
			},
		},
		{
			name: "old, new",
			config: &AgentConfiguration{
				Manager: &ManagerConfig{Labels: "old"},
			},
			labels: "new",
			verify: func(t *testing.T, config *AgentConfiguration) {
				require.NotNil(t, config.Manager)
				require.Equal(t, "new", config.Manager.Labels)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.config.ReplaceLabels(test.labels)
			test.verify(t, test.config)
		})
	}
}

func TestAgentConfigurationParse(t *testing.T) {
	tests := []struct {
		name   string
		raw    RawAgentConfiguration
		verify func(t *testing.T, config *AgentConfiguration, err error)
	}{
		{
			name: "empty",
			raw:  RawAgentConfiguration{},
			verify: func(t *testing.T, config *AgentConfiguration, err error) {
				require.NoError(t, err)
				require.NotNil(t, config)
				require.Equal(t, "", config.Collector)
				require.Equal(t, "", config.Logging)
				require.Nil(t, config.Manager)
			},
		},
		{
			name: "complete",
			raw: RawAgentConfiguration{
				Manager: []byte(`
endpoint: endpoint
agent_name: agent_name
agent_id: agent_id
secret_key: secret_key
labels: labels
tls_config:
    ca_file: cacert
    key_file: tlskey
    cert_file: tlscert
    insecure_skip_verify: true
`),
				Collector: []byte("collector"),
				Logging:   []byte("logging"),
			},
			verify: func(t *testing.T, config *AgentConfiguration, err error) {
				require.NoError(t, err)
				require.NotNil(t, config)
				require.Equal(t, "collector", config.Collector)
				require.Equal(t, "logging", config.Logging)
				require.NotNil(t, config.Manager)

				// validate fields
				require.Equal(t, "endpoint", config.Manager.Endpoint)
				require.Equal(t, strp("cacert"), config.Manager.TLS.CAFile)
				require.Equal(t, strp("tlscert"), config.Manager.TLS.CertFile)
				require.Equal(t, strp("tlskey"), config.Manager.TLS.KeyFile)
				require.Equal(t, true, config.Manager.TLS.InsecureSkipVerify)
				require.Equal(t, "agent_name", config.Manager.AgentName)
				require.Equal(t, "agent_id", config.Manager.AgentID)
				require.Equal(t, "secret_key", config.Manager.SecretKey)
				require.Equal(t, "labels", config.Manager.Labels)
			},
		},
		{
			name: "parse yaml error",
			raw: RawAgentConfiguration{
				Manager: []byte("not yaml"),
			},
			verify: func(t *testing.T, config *AgentConfiguration, err error) {
				require.Error(t, err)
				require.Nil(t, config)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config, err := test.raw.Parse()
			test.verify(t, config, err)
		})
	}
}

func TestAgentConfigurationMarshal(t *testing.T) {
	tests := []struct {
		name          string
		configuration AgentConfiguration
		verify        func(t *testing.T, raw *RawAgentConfiguration)
	}{
		{
			name:          "empty",
			configuration: AgentConfiguration{},
			verify: func(t *testing.T, raw *RawAgentConfiguration) {
				require.Equal(t, "", string(raw.Logging))
				require.Equal(t, "", string(raw.Collector))
				require.Nil(t, raw.Manager)
			},
		},
		{
			name: "complete",
			configuration: AgentConfiguration{
				Logging:   "logging",
				Collector: "collector",
				Manager: &ManagerConfig{
					Endpoint: "endpoint",

					AgentName: "agent_name",
					AgentID:   "agent_id",
					SecretKey: "secret_key",
					Labels:    "labels",
					TLS: &ManagerTLSConfig{
						CAFile:             strp("cacert"),
						CertFile:           strp("tlscert"),
						KeyFile:            strp("tlskey"),
						InsecureSkipVerify: true,
					},
				},
			},
			verify: func(t *testing.T, raw *RawAgentConfiguration) {
				require.Equal(t, "logging", string(raw.Logging))
				require.Equal(t, "collector", string(raw.Collector))
				require.Equal(t, strings.TrimLeft(`
endpoint: endpoint
secret_key: secret_key
agent_id: agent_id
labels: labels
agent_name: agent_name
tls_config:
    insecure_skip_verify: true
    key_file: tlskey
    cert_file: tlscert
    ca_file: cacert
`, "\n"), string(raw.Manager))
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			raw := test.configuration.Raw()
			test.verify(t, &raw)
		})
	}
}

func TestAgentConfigurationComputeConfigurationUpdates(t *testing.T) {
	tests := []struct {
		name        string
		server      AgentConfiguration
		agent       AgentConfiguration
		expect      AgentConfiguration
		expectEmpty bool
	}{
		{
			name:        "empty",
			server:      AgentConfiguration{},
			agent:       AgentConfiguration{},
			expect:      AgentConfiguration{},
			expectEmpty: true,
		},
		{
			name: "same",
			server: AgentConfiguration{
				Collector: "collector",
				Logging:   "logging",
				Manager: &ManagerConfig{
					AgentName: "my-agent",
				},
			},
			agent: AgentConfiguration{
				Collector: "collector",
				Logging:   "logging",
				Manager: &ManagerConfig{
					AgentName: "my-agent",
				},
			},
			expect:      AgentConfiguration{},
			expectEmpty: true,
		},
		{
			name: "logging ignored",
			server: AgentConfiguration{
				Logging: "ignored",
			},
			agent: AgentConfiguration{
				Logging: "different",
			},
			expect:      AgentConfiguration{},
			expectEmpty: true,
		},
		{
			name: "collector different",
			server: AgentConfiguration{
				Collector: "collector",
			},
			agent: AgentConfiguration{
				Collector: "different",
			},
			expect: AgentConfiguration{
				Collector: "collector",
			},
			expectEmpty: false,
		},
		{
			name: "non-label manager changes ignored",
			server: AgentConfiguration{
				Manager: &ManagerConfig{
					AgentName: "my-agent",
				},
			},
			agent: AgentConfiguration{
				Manager: &ManagerConfig{
					AgentName: "new-name",
				},
			},
			expect:      AgentConfiguration{},
			expectEmpty: true,
		},
		{
			name: "label change, no manager",
			server: AgentConfiguration{
				Manager: &ManagerConfig{
					Labels: "foo=bar",
				},
			},
			agent: AgentConfiguration{},
			expect: AgentConfiguration{
				Manager: &ManagerConfig{
					Labels: "foo=bar",
				},
			},
			expectEmpty: false,
		},
		{
			name: "label change, different labels",
			server: AgentConfiguration{
				Manager: &ManagerConfig{
					Labels: "foo=bar",
				},
			},
			agent: AgentConfiguration{
				Manager: &ManagerConfig{
					Labels: "foo=baz",
				},
			},
			expect: AgentConfiguration{
				Manager: &ManagerConfig{
					Labels: "foo=bar",
				},
			},
			expectEmpty: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			diff := ComputeConfigurationUpdates(&test.server, &test.agent)
			require.Equal(t, test.expectEmpty, diff.Empty())
			verifyAgentConfiguration(t, test.expect, &diff)
		})
	}
}

func verifyAgentConfiguration(t *testing.T, expect AgentConfiguration, diff *AgentConfiguration) {
	require.Equal(t, expect.Logging, diff.Logging)
	require.Equal(t, expect.Collector, diff.Collector)
	if expect.Manager == nil {
		require.Nil(t, diff.Manager)
		return
	}
	require.NotNil(t, diff.Manager)

	// compare individual manager fields
	require.Equal(t, expect.Manager.Endpoint, diff.Manager.Endpoint)

	require.Equal(t, expect.Manager.AgentName, diff.Manager.AgentName)
	require.Equal(t, expect.Manager.AgentID, diff.Manager.AgentID)
	require.Equal(t, expect.Manager.SecretKey, diff.Manager.SecretKey)
	require.Equal(t, expect.Manager.Labels, diff.Manager.Labels)
	if expect.Manager.TLS != nil {
		require.NotNil(t, diff.Manager.TLS)
		require.Equal(t, expect.Manager.TLS.CertFile, diff.Manager.TLS.CertFile)
		require.Equal(t, expect.Manager.TLS.CAFile, diff.Manager.TLS.CAFile)
		require.Equal(t, expect.Manager.TLS.KeyFile, diff.Manager.TLS.KeyFile)
		require.Equal(t, expect.Manager.TLS.InsecureSkipVerify, diff.Manager.TLS.InsecureSkipVerify)
	} else {
		require.Nil(t, diff.Manager.TLS)
	}

}

func TestAgentConfigurationApplyUpdates(t *testing.T) {
	tests := []struct {
		name    string
		current RawAgentConfiguration
		updates *RawAgentConfiguration
		expect  RawAgentConfiguration
	}{
		{
			name:    "empty",
			current: RawAgentConfiguration{},
			updates: &RawAgentConfiguration{},
			expect:  RawAgentConfiguration{},
		},
		{
			name:    "nil",
			current: RawAgentConfiguration{},
			updates: nil,
			expect:  RawAgentConfiguration{},
		},
		{
			name: "partial",
			current: RawAgentConfiguration{
				Logging:   []byte("logging"),
				Collector: nil,
				Manager:   nil,
			},
			updates: &RawAgentConfiguration{
				Logging:   []byte("different"),
				Collector: []byte("collector"),
			},
			expect: RawAgentConfiguration{
				Logging:   []byte("different"),
				Collector: []byte("collector"),
			},
		},
		{
			name: "complete",
			current: RawAgentConfiguration{
				Logging:   []byte("logging"),
				Collector: []byte("collector"),
				Manager:   []byte("labels: foo=bar"),
			},
			updates: &RawAgentConfiguration{
				Logging:   []byte("logging2"),
				Collector: []byte("collector2"),
				Manager:   []byte("labels: foo=baz"),
			},
			expect: RawAgentConfiguration{
				Logging:   []byte("logging2"),
				Collector: []byte("collector2"),
				Manager:   []byte("labels: foo=baz"),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.current.ApplyUpdates(test.updates)
			require.Equal(t, string(test.expect.Collector), string(result.Collector), "Collector")
			require.Equal(t, string(test.expect.Logging), string(result.Logging), "Logging")
			require.Equal(t, string(test.expect.Manager), string(result.Manager), "Manager")
		})
	}
}

func TestAgentConfigurationHash(t *testing.T) {
	raw := RawAgentConfiguration{
		Logging:   []byte("logging"),
		Collector: []byte("collector"),
	}
	require.ElementsMatch(t, []byte{
		2, 159, 2, 3, 209, 111, 244, 145, 150, 81, 82, 158, 180, 134, 62, 7, 151, 74, 32, 121, 111, 156, 37, 229, 4, 136, 76, 189, 127, 213, 126, 210,
	}, raw.Hash())
}

func TestDecodeAgentConfiguration(t *testing.T) {
	tests := []struct {
		name          string
		configuration interface{}
		expect        AgentConfiguration
		expectError   bool
	}{
		{
			name:          "nil",
			configuration: nil,
			expect:        AgentConfiguration{},
		},
		{
			name:          "empty",
			configuration: map[string]interface{}{},
			expect:        AgentConfiguration{},
		},
		{
			name: "malformed",
			configuration: map[string]interface{}{
				"Collector": map[string]interface{}{
					"something": "Collector should be a string",
				},
			},
			expectError: true,
		},
		{
			name: "complete",
			configuration: map[string]any{
				"collector": "collector contents",
				"logging":   "logging contents",
				"manager": map[string]any{
					"endpoint":   "endpoint",
					"agent_name": "agent_name",
					"agent_id":   "agent_id",
					"secret_key": "secret_key",
					"labels":     "labels",
					"tls_config": map[string]any{
						"key_file":             "keyfile",
						"cert_file":            "certfile",
						"ca_file":              "cafile",
						"insecure_skip_verify": true,
					},
				},
			},
			expect: AgentConfiguration{
				Logging:   "logging contents",
				Collector: "collector contents",
				Manager: &ManagerConfig{
					Endpoint:  "endpoint",
					AgentName: "agent_name",
					AgentID:   "agent_id",
					SecretKey: "secret_key",
					Labels:    "labels",
					TLS: &ManagerTLSConfig{
						CAFile:             strp("cafile"),
						KeyFile:            strp("keyfile"),
						CertFile:           strp("certfile"),
						InsecureSkipVerify: true,
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config, err := DecodeAgentConfiguration(test.configuration)
			if test.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			verifyAgentConfiguration(t, test.expect, config)
		})
	}
}

func strp(s string) *string {
	return &s
}
