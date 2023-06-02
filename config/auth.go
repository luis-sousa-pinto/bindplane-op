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
)

const (
	// DefaultUsername is the default username used for communication between client and server.
	DefaultUsername = "admin"

	// DefaultPassword is the default password used for communication between client and server.
	DefaultPassword = "admin"
)

/* #nosec G101 -- these credentials are use to detect if we need to replace them and are only valid for first install */
var (
	// DefaultSecretKey is the default value for secret key.
	// Having this be hard coded allows us to determine if a user has changed from defaults.
	DefaultSecretKey = "38f6b093-ed43-457d-9564-1b55006f66b2"

	// DefaultSessionSecret is the default value for session secret.
	// Having this be hard coded allows us to determine if a user has changed from defaults.
	DefaultSessionSecret = "5cdd2530-c4ee-4294-ad8f-217a9406eaf2"
)

// Auth is the configuration for authentication
type Auth struct {
	// SecretKey is a shared secret between the server and the agent to ensure agents are authorized to communicate with the server.
	SecretKey string `mapstructure:"secretKey,omitempty" yaml:"secretKey,omitempty"`

	// Username is the basic auth username used for communication between client and server.
	Username string `mapstructure:"username" yaml:"username,omitempty"`

	// Password is the basic auth password used for communication between client and server.
	Password string `mapstructure:"password" yaml:"password,omitempty"`

	// SessionSecret is the secret used to sign the session cookie.
	SessionSecret string `mapstructure:"sessionSecret" yaml:"sessionSecret,omitempty"`
}

// Validate validates the auth configuration.
func (c *Auth) Validate() error {
	if c.Username == "" {
		return errors.New("username must be set")
	}

	if c.Password == "" {
		return errors.New("password must be set")
	}

	if c.SecretKey == "" {
		return errors.New("secret key must be set")
	}

	if c.SessionSecret == "" {
		return errors.New("session secret must be set")
	}

	return nil
}
