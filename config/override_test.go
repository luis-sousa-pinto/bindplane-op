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
	"os"
	"testing"
	"time"

	modelversion "github.com/observiq/bindplane-op/model/version"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestOverrideDefaults(t *testing.T) {
	flagSet := pflag.NewFlagSet("test", pflag.PanicOnError)
	overrides := DefaultOverrides()
	for _, override := range overrides {
		override.Bind(flagSet)
	}

	cfg := NewConfig()
	err := viper.Unmarshal(cfg)
	require.NoError(t, err)

	expectedCfg := &Config{
		APIVersion:       modelversion.V1,
		Env:              EnvProduction,
		Output:           DefaultOutput,
		RolloutsInterval: DefaultRolloutsInterval,
		Logging: Logging{
			Output:   LoggingOutputFile,
			FilePath: DefaultLoggingFilePath,
		},
		Network: Network{
			Host: DefaultHost,
			Port: DefaultPort,
			TLS: TLS{
				CertificateAuthority: []string{},
			},
		},
		Auth: Auth{
			Username:      DefaultUsername,
			Password:      DefaultPassword,
			SecretKey:     DefaultSecretKey,
			SessionSecret: DefaultSessionSecret,
		},
		Store: Store{
			Type:      StoreTypeBBolt,
			MaxEvents: DefaultMaxEvents,
			BBolt: BBolt{
				Path: DefaultBBoltPath,
			},
		},
		Metrics: Metrics{
			Interval: DefaultMetricsInterval,
		},
		AgentVersions: AgentVersions{
			SyncInterval: DefaultSyncInterval,
		},
	}
	require.Equal(t, expectedCfg, cfg)
}

func TestOverrideFlags(t *testing.T) {
	flagSet := pflag.NewFlagSet("test", pflag.PanicOnError)
	args := []string{
		"--env", "production",
		"--output", "json",
		"-o", "json",
		"--logging-output", "file",
		"--logging-file-path", "/tmp/test.log",
		"--host", "localhost",
		"--port", "8080",
		"--remote-url", "http://localhost:8080",
		"--offline", "true",
		"--tls-cert", "/tmp/cert.pem",
		"--tls-key", "/tmp/key.pem",
		"--tls-ca", "/tmp/ca.pem",
		"--tls-skip-verify", "true",
		"--rollouts-interval", "50s",
		"--username", "user",
		"--password", "password",
		"--secret-key", "secret",
		"--session-secret", "session",
		"--tracing-type", "otlp",
		"--tracing-otlp-endpoint", "localhost:4317",
		"--tracing-otlp-insecure", "true",
		"--tracing-sampling-rate", "0.5",
		"--metrics-type", "otlp",
		"--metrics-otlp-endpoint", "localhost:4317",
		"--metrics-otlp-insecure", "true",
		"--metrics-interval", "2m",
		"--store-type", "bbolt",
		"--store-bbolt-path", "/tmp/store.db",
		"--store-max-events", "200",
		"--agent-versions-sync-interval", "2h",
	}

	overrides := DefaultOverrides()
	for _, override := range overrides {
		override.Bind(flagSet)
	}

	err := flagSet.Parse(args)
	require.NoError(t, err)

	cfg := NewConfig()
	err = viper.Unmarshal(cfg)
	require.NoError(t, err)

	expectedCfg := &Config{
		APIVersion:       modelversion.V1,
		Env:              EnvProduction,
		Output:           "json",
		Offline:          true,
		RolloutsInterval: time.Second * 50,
		Logging: Logging{
			Output:   "file",
			FilePath: "/tmp/test.log",
		},
		Network: Network{
			Host:      "localhost",
			Port:      "8080",
			RemoteURL: "http://localhost:8080",
			TLS: TLS{
				Certificate:          "/tmp/cert.pem",
				PrivateKey:           "/tmp/key.pem",
				CertificateAuthority: []string{"/tmp/ca.pem"},
				InsecureSkipVerify:   true,
			},
		},
		Auth: Auth{
			Username:      "user",
			Password:      "password",
			SecretKey:     "secret",
			SessionSecret: "session",
		},
		Store: Store{
			Type:      StoreTypeBBolt,
			MaxEvents: 200,
			BBolt: BBolt{
				Path: "/tmp/store.db",
			},
		},
		Tracing: Tracing{
			Type:         "otlp",
			SamplingRate: float64(0.5),
			OTLP: OTLPTracing{
				Endpoint: "localhost:4317",
				Insecure: true,
			},
		},
		Metrics: Metrics{
			Type: "otlp",
			OTLP: OTLPMetrics{
				Endpoint: "localhost:4317",
				Insecure: true,
			},
			Interval: time.Minute * 2,
		},
		AgentVersions: AgentVersions{
			SyncInterval: time.Hour * 2,
		},
	}
	require.Equal(t, expectedCfg, cfg)
}

func TestOverrideEnvs(t *testing.T) {
	envs := map[string]string{
		"BINDPLANE_ENV":                          "production",
		"BINDPLANE_OUTPUT":                       "json",
		"BINDPLANE_LOGGING_OUTPUT":               "file",
		"BINDPLANE_LOGGING_FILE_PATH":            "/tmp/test.log",
		"BINDPLANE_HOST":                         "localhost",
		"BINDPLANE_PORT":                         "8080",
		"BINDPLANE_REMOTE_URL":                   "http://localhost:8080",
		"BINDPLANE_OFFLINE":                      "true",
		"BINDPLANE_TLS_CERT":                     "/tmp/cert.pem",
		"BINDPLANE_TLS_KEY":                      "/tmp/key.pem",
		"BINDPLANE_TLS_CA":                       "/tmp/ca.pem",
		"BINDPLANE_TLS_SKIP_VERIFY":              "true",
		"BINDPLANE_ROLLOUTS_INTERVAL":            "50s",
		"BINDPLANE_USERNAME":                     "user",
		"BINDPLANE_PASSWORD":                     "password",
		"BINDPLANE_SECRET_KEY":                   "secret",
		"BINDPLANE_SESSION_SECRET":               "session",
		"BINDPLANE_TRACING_TYPE":                 "otlp",
		"BINDPLANE_TRACING_OTLP_ENDPOINT":        "localhost:4317",
		"BINDPLANE_TRACING_OTLP_INSECURE":        "true",
		"BINDPLANE_TRACING_SAMPLING_RATE":        "0.5",
		"BINDPLANE_METRICS_TYPE":                 "otlp",
		"BINDPLANE_METRICS_OTLP_ENDPOINT":        "localhost:4317",
		"BINDPLANE_METRICS_OTLP_INSECURE":        "true",
		"BINDPLANE_STORE_TYPE":                   "bbolt",
		"BINDPLANE_STORE_BBOLT_PATH":             "/tmp/store.db",
		"BINDPLANE_STORE_MAX_EVENTS":             "200",
		"BINDPLANE_AGENT_VERSIONS_SYNC_INTERVAL": "2h",
	}
	setEnvs(t, envs)
	defer unsetEnvs(t, envs)

	flagSet := pflag.NewFlagSet("test", pflag.PanicOnError)
	overrides := DefaultOverrides()
	for _, override := range overrides {
		override.Bind(flagSet)
	}

	cfg := NewConfig()
	err := viper.Unmarshal(cfg)
	require.NoError(t, err)

	expectedCfg := &Config{
		APIVersion:       modelversion.V1,
		Env:              EnvProduction,
		Output:           "json",
		Offline:          true,
		RolloutsInterval: time.Second * 50,

		Logging: Logging{
			Output:   "file",
			FilePath: "/tmp/test.log",
		},
		Network: Network{
			Host:      "localhost",
			Port:      "8080",
			RemoteURL: "http://localhost:8080",
			TLS: TLS{
				Certificate:          "/tmp/cert.pem",
				PrivateKey:           "/tmp/key.pem",
				CertificateAuthority: []string{"/tmp/ca.pem"},
				InsecureSkipVerify:   true,
			},
		},
		Auth: Auth{
			Username:      "user",
			Password:      "password",
			SecretKey:     "secret",
			SessionSecret: "session",
		},
		Store: Store{
			Type:      StoreTypeBBolt,
			MaxEvents: 200,
			BBolt: BBolt{
				Path: "/tmp/store.db",
			},
		},
		Tracing: Tracing{
			Type:         "otlp",
			SamplingRate: float64(0.5),
			OTLP: OTLPTracing{
				Endpoint: "localhost:4317",
				Insecure: true,
			},
		},
		Metrics: Metrics{
			Type:     "otlp",
			Interval: DefaultMetricsInterval,
			OTLP: OTLPMetrics{
				Endpoint: "localhost:4317",
				Insecure: true,
			},
		},
		AgentVersions: AgentVersions{
			SyncInterval: time.Hour * 2,
		},
	}
	require.Equal(t, expectedCfg, cfg)
}

// setEnvs sets the given environment variables.
func setEnvs(t *testing.T, envs map[string]string) {
	for k, v := range envs {
		err := os.Setenv(k, v)
		require.NoError(t, err)
	}
}

// unsetEnvs unsets the given environment variables.
func unsetEnvs(t *testing.T, envs map[string]string) {
	for k := range envs {
		err := os.Unsetenv(k)
		require.NoError(t, err)
	}
}
