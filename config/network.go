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
	"fmt"
	"strings"

	"github.com/observiq/bindplane-op/common"
)

const (
	// DefaultHost is the default host to which the server will bind.
	DefaultHost = "127.0.0.1"

	// DefaultPort is the default port on which the server will serve.
	DefaultPort = "3001"
)

// Network is the configuration for networking in BindPlane
type Network struct {
	// Host is the Host to which the server will bind.
	Host string `mapstructure:"host" yaml:"host,omitempty"`

	// Port is the Port on which the server will serve.
	Port string `mapstructure:"port" yaml:"port,omitempty"`

	// RemoteURL is the remote URL of the server. If not provided, this will be constructed from the Host and Port.
	RemoteURL string `mapstructure:"remoteURL" yaml:"remoteURL,omitempty"`

	TLS `mapstructure:",squash" yaml:",inline"`
}

// BindAddress is the address (host:port) to which the server will bind
func (n *Network) BindAddress() string {
	return fmt.Sprintf("%s:%s", n.Host, n.Port)
}

// WebsocketURL returns the websocket URL of the server.
// This will be the RemoteURL if provided, otherwise it will be constructed from the Host and Port.
func (n *Network) WebsocketURL() string {
	if n.RemoteURL == "" {
		return fmt.Sprintf("%s://%s:%s", n.WebsocketScheme(), n.Host, n.Port)
	}

	// If the remote URL is provided, we need to replace the scheme with the websocket scheme.
	if strings.HasPrefix(n.RemoteURL, "https://") {
		return strings.Replace(n.RemoteURL, "https://", "wss://", 1)
	}

	return strings.Replace(n.RemoteURL, "http://", "ws://", 1)
}

// ServerURL returns the server URL of the server.
// This will be the RemoteURL if provided, otherwise it will be constructed from the Host and Port.
func (n *Network) ServerURL() string {
	if n.RemoteURL != "" {
		return n.RemoteURL
	}

	return fmt.Sprintf("%s://%s:%s", n.ServerScheme(), n.Host, n.Port)
}

// WebsocketScheme returns ws or wss
func (n *Network) WebsocketScheme() string {
	if n.TLSEnabled() {
		return "wss"
	}
	return "ws"
}

// ServerScheme returns http or https
func (n *Network) ServerScheme() string {
	if n.TLSEnabled() {
		return "https"
	}
	return "http"
}

// Validate validates the TLS configuration
func (n *Network) Validate() error {
	if err := n.validatePort(); err != nil {
		return fmt.Errorf("failed to validate port: %w", err)
	}

	if err := n.validateRemoteURL(); err != nil {
		return fmt.Errorf("failed to validate remote url: %w", err)
	}

	if err := n.TLS.Validate(); err != nil {
		return fmt.Errorf("failed to validate tls: %w", err)
	}

	return nil
}

// validatePort validates the port
func (n *Network) validatePort() error {
	return common.ValidatePort(n.Port)
}

// validateRemoteURL validates the remote URL
func (n *Network) validateRemoteURL() error {
	if n.RemoteURL == "" {
		return nil
	}

	if err := common.ValidateURL(n.RemoteURL, common.ValidHTTPSchemes); err != nil {
		return fmt.Errorf("invalid remote url: %w", err)
	}

	return nil
}
