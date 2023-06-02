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

package legacy

import (
	"fmt"
	"path/filepath"
	"time"
)

// Config is a legacy BindPlane configuration.
//
// Deprecated: This is a legacy type and should not be used.
type Config struct {
	// Server is the server configuration.
	Server `mapstructure:"server" yaml:"server,omitempty"`
	// Client is the client configuration.
	Client `mapstructure:"client" yaml:"client,omitempty"`
	// Command is the command configuration.
	Command `mapstructure:"command" yaml:"command,omitempty"`
}

// LogOutput is an enum of possible values for the LogOutput configuration setting
//
// Deprecated: This is a legacy type and should not be used.
type LogOutput string

const (
	// LogOutputFile will write logs to the file specified by LogFilePath
	//
	// Deprecated: This is a legacy value and should not be used.
	LogOutputFile LogOutput = "file"

	// LogOutputStdout will write logs to stdout
	//
	// Deprecated: This is a legacy value and should not be used.
	LogOutputStdout LogOutput = "stdout"
)

// Env is an enum indicating the environment in which BindPlane is running.
//
// Deprecated: This is a legacy type and should not be used.
type Env string

const (
	// EnvDevelopment should be used for development and uses debug logging and normal gin request logging to stdout.
	//
	// Deprecated: This is a legacy value and should not be used.
	EnvDevelopment Env = "development"

	// EnvTest should be used for tests and uses debug logging with json gin request logging to the log file.
	//
	// Deprecated: This is a legacy value and should not be used.
	EnvTest Env = "test"

	// EnvProduction the the default and should be used in production and uses info logging with json gin request logging to the log file.
	//
	// Deprecated: This is a legacy value and should not be used.
	EnvProduction Env = "production"
)

// Common is a common configuration for both the client and server.
//
// Deprecated: This is a legacy type and should not be used.
type Common struct {
	// Env is one of
	Env Env `mapstructure:"env" yaml:"env,omitempty"`

	// Host is the Host to which the server will bind.
	Host string `mapstructure:"host" yaml:"host,omitempty"`

	// Port is the Port on which the server will serve.
	Port string `mapstructure:"port" yaml:"port,omitempty"`

	// ServerURL is the URL that clients should use to contact the server.
	ServerURL string `mapstructure:"serverURL" yaml:"serverURL,omitempty"`

	// Username the basic auth username used for communication between client and server.
	Username string `mapstructure:"username" yaml:"username,omitempty"`
	// The basic auth password used for communication between client and server.
	Password string `mapstructure:"password" yaml:"password,omitempty"`

	// TLSConfig is an optional TLS configuration for communication between client and server.
	TLSConfig `yaml:",inline" mapstructure:",squash"`

	// LogFilePath is the path of the bindplane log file, defaulting to $HOME/.bindplane/bindplane.log
	LogFilePath string `mapstructure:"logFilePath" yaml:"logFilePath,omitempty"`

	// LogOutput indicates where logs should be written, defaulting to "file"
	LogOutput LogOutput `mapstructure:"logOutput" yaml:"logOutput,omitempty"`

	// bindplaneHomePath is the root folder path of BindPlane home, defaulting to $HOME/.bindplane.
	// It is read-only and available via BindPlaneHomePath()
	bindplaneHomePath string

	// TraceType specifies the type of tracing to use. Valid values are "google" or "otlp".
	TraceType string `mapstructure:"traceType,omitempty" yaml:"traceType,omitempty"`

	// GoogleCloudTracing is used to send traces to Google Cloud when TraceType is set to "google".
	GoogleCloudTracing GoogleCloudTracing `mapstructure:"googleTracing,omitempty" yaml:"googleTracing,omitempty"`

	// OpenTelemetryTracing is used to send traces to an Open Telemetry OTLP receiver when
	// TraceType is set to "otlp".
	OpenTelemetryTracing OpenTelemetryTracing `mapstructure:"otlpTracing,omitempty" yaml:"otlpTracing,omitempty"`
}

// GoogleCloudTracing is configuration for tracing to Google Cloud Monitoring
//
// Deprecated: This is a legacy type and should not be used.
type GoogleCloudTracing struct {
	ProjectID       string `mapstructure:"projectID,omitempty" yaml:"projectID,omitempty"`
	CredentialsFile string `mapstructure:"credentialsFile,omitempty" yaml:"credentialsFile,omitempty"`
}

// OpenTelemetryTracing is configuration for tracing to an Open Telemetry Collector
//
// Deprecated: This is a legacy type and should not be used.
type OpenTelemetryTracing struct {
	Endpoint string `mapstructure:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	TLS      struct {
		Insecure bool `mapstructure:"insecure,omitempty" yaml:"insecure,omitempty"`
	} `mapstructure:"tls,omitempty" yaml:"tls,omitempty"`
}

// TLSConfig contains configuration for connecting over TLS and mTLS.
//
// Deprecated: This is a legacy type and should not be used.
type TLSConfig struct {
	// Certificate is the path to the x509 PEM encoded certificate file that will be used to
	// establish TLS connections.
	//
	// When operating in server mode, this certificate is presented to clients.
	// When operating in client mode with mTLS, this certificate is used for authentication
	// against the server.
	Certificate string `mapstructure:"tlsCert" yaml:"tlsCert,omitempty"`

	// PrivateKey is the matching x509 PEM encoded private key for the Certificate.
	PrivateKey string `mapstructure:"tlsKey" yaml:"tlsKey,omitempty"`

	// CertificateAuthority is one or more file paths to x509 PEM encoded certificate authority chains.
	// These certificate authorities are used for trusting incoming client mTLS connections.
	CertificateAuthority []string `mapstructure:"tlsCa" yaml:"tlsCa,omitempty"`

	// InsecureSkipVerify controls whether a client verifies the server's certificate chain and host name. If
	// InsecureSkipVerify is true, crypto/tls accepts any certificate presented by the server and any host name in that
	// certificate.
	//
	// It is also used to signal that clients, like the agent metrics pipeline, should connect to the server with
	// tls.insecure set to true.
	//
	// In this mode, TLS is susceptible to machine-in-the-middle attacks. This should be used only for testing only.
	InsecureSkipVerify bool `mapstructure:"tlsSkipVerify" yaml:"tlsSkipVerify,omitempty"`
}

const (
	// StoreTypeMap uses an in-memory store.
	//
	// Deprecated: This is a legacy value and should not be used.
	StoreTypeMap = "map"
	// StoreTypeBbolt uses go.etcd.io/bbolt for storage.
	//
	// Deprecated: This is a legacy value and should not be used.
	StoreTypeBbolt = "bbolt"
)

// Server is the server configuration.
//
// Deprecated: This is a legacy type and should not be used.
type Server struct {
	// StoreType indicates the type of store to use. "map", "bbolt", and "googlecloud" are currently supported.
	StoreType string `mapstructure:"storeType,omitempty" yaml:"storeType,omitempty"`

	// StorageFilePath is the path of the bindplane storage file.
	StorageFilePath string `mapstructure:"storageFilePath,omitempty" yaml:"storageFilePath,omitempty"`

	// SecretKey is a shared secret between the server and the agent to ensure agents are authorized to communicate with the server.
	SecretKey string `mapstructure:"secretKey,omitempty" yaml:"secretKey,omitempty"`

	// RemoteURL is the URL that agents should use to contact the server
	RemoteURL string `mapstructure:"remoteURL,omitempty" yaml:"remoteURL,omitempty"`

	// Offline mode indicates if the server should be considered offline. An offline server will not attempt to contact
	// any other services. It will still allow agents to connect and serve api requests.
	Offline bool `mapstructure:"offline,omitempty" yaml:"offline,omitempty"`

	// SessionSecret is used to encode the user sessions cookies.  It should be a uuid.
	SessionsSecret string `mapstructure:"sessionsSecret,omitempty" yaml:"sessionsSecret,omitempty"`

	Common `yaml:",inline" mapstructure:",squash"`

	// SyncAgentVersionsInterval is the interval at which agent-versions will be synchronized with GitHub. Set to 0 to
	// turn off synchronization. Disabled if Offline is true.
	SyncAgentVersionsInterval time.Duration `mapstructure:"syncAgentVersionsInterval,omitempty" yaml:"syncAgentVersionsInterval,omitempty"`
}

// Client is the client configuration.
//
// Deprecated: This is a legacy type and should not be used.
type Client struct {
	Common
}

// Command is the command configuration.
//
// Deprecated: This is a legacy type and should not be used.
type Command struct {
	// Output is the output format for the command.
	Output string `mapstructure:"output" yaml:"output"`
}

// BindAddress is the address (host:port) to which the server will bind
func (c *Server) BindAddress() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// WebsocketURL is the URL that should be used for agents connecting to the server
func (c *Server) WebsocketURL() string {
	if c.RemoteURL != "" {
		return c.RemoteURL
	}
	if c.Host == "" && c.Port == "" {
		return ""
	}
	return fmt.Sprintf("%s://%s:%s", c.WebsocketScheme(), c.Host, c.Port)
}

// BoltDatabasePath returns the path to the bolt database file
func (c *Server) BoltDatabasePath() string {
	if c.StorageFilePath != "" {
		return c.StorageFilePath
	}
	return filepath.Join(c.BindPlaneHomePath(), "storage")
}

// BindPlaneEnv ensures that Env has a valid value and defaults to EnvProduction
func (c *Common) BindPlaneEnv() Env {
	switch c.Env {
	case EnvDevelopment:
		return EnvDevelopment
	case EnvTest:
		return EnvTest
	default:
		return EnvProduction
	}
}

// BindPlaneHomePath returns the path to the BindPlane home where files are stored by default
func (c *Common) BindPlaneHomePath() string {
	return c.bindplaneHomePath
}

// BindPlaneLogFilePath returns the path to the log file for bindplane
func (c *Common) BindPlaneLogFilePath() string {
	if c.LogFilePath != "" {
		return c.LogFilePath
	}
	return filepath.Join(c.BindPlaneHomePath(), "bindplane.log")
}

// EnableTLS returns true if TLS is enabled
func (c *Common) EnableTLS() bool {
	return c.Certificate != "" && c.PrivateKey != ""
}

// WebsocketScheme returns ws or wss
func (c *Common) WebsocketScheme() string {
	if c.EnableTLS() {
		return "wss"
	}
	return "ws"
}

// ServerScheme returns http or https
func (c *Common) ServerScheme() string {
	if c.EnableTLS() {
		return "https"
	}
	return "http"
}

// BindPlaneURL returns the configured server url. If one is not configured,
// a url derived from the configured host and port is used.
func (c *Common) BindPlaneURL() string {
	if c.ServerURL != "" {
		return c.ServerURL
	}
	if c.Host == "" && c.Port == "" {
		return ""
	}
	return fmt.Sprintf("%s://%s:%s", c.ServerScheme(), c.Host, c.Port)
}

// BindPlaneInsecureSkipVerify returns the value of InsecureSkipVerify from the TLSConfig
func (c *Common) BindPlaneInsecureSkipVerify() bool {
	return c.InsecureSkipVerify
}
