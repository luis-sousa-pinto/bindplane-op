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

package question

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/google/uuid"
	"github.com/observiq/bindplane-op/config"
)

// ServerAnswers are answers to the default server survey
type ServerAnswers struct {
	Host           string `survey:"host"`
	Port           string `survey:"port"`
	RemoteURL      string `survey:"remote-url"`
	SecretKey      string `survey:"secret-key"`
	SessionsSecret string `survey:"sessions-secret"`
	Username       string `survey:"username"`
	Password       string `survey:"password"`
}

// CreateServerQuestions creates the default set of server questions.
func CreateServerQuestions() []RelevantServerQuestion {
	return []RelevantServerQuestion{
		func(_ *ServerAnswers, cfg *config.Config) (*Question, bool) {
			return &Question{
				BeforeText: "The IP address the BindPlane server should listen on.\nSet to 0.0.0.0 to listen on all IP addresses.",
				Question: survey.Question{
					Name: "host",
					Prompt: &survey.Input{
						Message: "Server Host",
						Help:    "Bind Address for the HTTP server or 0.0.0.0 to bind to all network interfaces",
						Default: cfg.Network.Host,
					},
					Validate: survey.Required,
				},
			}, true
		},
		func(_ *ServerAnswers, cfg *config.Config) (*Question, bool) {
			return &Question{
				BeforeText: "The TCP port BindPlane should bind to.\nAll communication to the BindPlane server (HTTP, GraphQL, WebSocket) will use this port.",
				Question: survey.Question{
					Name: "port",
					Prompt: &survey.Input{
						Message: "Server Port",
						Help:    "Port for the HTTP server",
						Default: cfg.Network.Port,
					},
					Validate: survey.Required,
				},
			}, true
		},
		func(responses *ServerAnswers, cfg *config.Config) (*Question, bool) {
			return &Question{
				BeforeText: "The full HTTP URL used for communication between client and server.\nUse the IP address or hostname of the server, starting with http:// for plain text or https:// for TLS.",
				Question: survey.Question{
					Name: "remote-url",
					Prompt: &survey.Input{
						Message: "Remote URL",
						Help:    "Full HTTP URL where the server can be reached.",
						Default: getRemoteURLDefaultValue(responses, cfg.Network),
					},
					Validate: survey.Required,
				},
			}, true
		},
		func(_ *ServerAnswers, cfg *config.Config) (*Question, bool) {
			defaultSecretKey := cfg.Auth.SecretKey
			if cfg.Auth.SecretKey == config.DefaultSecretKey {
				defaultSecretKey = uuid.NewString()
			}

			return &Question{
				BeforeText: "Choose a secret key to be used for authentication between server and agents.",
				Question: survey.Question{
					Name: "secret-key",
					Prompt: &survey.Input{
						Message: "Secret Key",
						Default: defaultSecretKey,
					},
					Validate: survey.Required,
				},
			}, true
		},
		func(_ *ServerAnswers, cfg *config.Config) (*Question, bool) {
			defaultSessionSecret := cfg.Auth.SessionSecret
			if cfg.Auth.SessionSecret == config.DefaultSessionSecret {
				defaultSessionSecret = uuid.NewString()
			}
			return &Question{
				BeforeText: "Choose a secret key to be used to encode user session cookies.  Must be a uuid.",
				Question: survey.Question{
					Name: "sessions-secret",
					Prompt: &survey.Input{
						Message: "Sessions Secret",
						Default: defaultSessionSecret,
					},
					Validate: validateUUID,
				},
			}, true
		},
		func(_ *ServerAnswers, cfg *config.Config) (*Question, bool) {
			return &Question{
				BeforeText: "Specify a username and password to restrict access to the server.",
				Question: survey.Question{
					Name: "username",
					Prompt: &survey.Input{
						Message: "Username",
						Default: cfg.Auth.Username,
					},
					Validate: survey.Required,
				},
			}, true
		},
		func(_ *ServerAnswers, cfg *config.Config) (*Question, bool) {
			return CreatePasswordQuestion(cfg), true
		},
	}
}

// CreatePasswordQuestion returns the default password question.
func CreatePasswordQuestion(cfg *config.Config) *Question {
	if cfg.Auth.Password == "" {
		return &Question{
			Question: survey.Question{
				Name: "password",
				Prompt: &survey.Password{
					Message: "Password (must not be empty)",
				},
				Validate: survey.Required,
			},
		}
	}

	return &Question{
		Question: survey.Question{
			Name: "password",
			Prompt: &survey.Password{
				Message: "Password (blank will preserve the current password)",
			},
		},
	}
}

// validateUUID validates that the answer is a valid uuid
func validateUUID(answer interface{}) error {
	_, err := uuid.Parse(answer.(string))
	return err
}

// getRemoteURLDefaultValue returns the default value for the remote URL based on the current response state
func getRemoteURLDefaultValue(responses *ServerAnswers, networkCfg config.Network) string {
	hasChanged := (responses.Host != networkCfg.Host) || (responses.Port != networkCfg.Port)

	switch {
	case networkCfg.RemoteURL != "" && !hasChanged:
		return networkCfg.RemoteURL
	case responses.Host == "0.0.0.0" || responses.Host == "127.0.0.1":
		// Ignore this error as we only expect it fail on non-mainstream distributions
		hostname, _ := os.Hostname()
		return fmt.Sprintf("%s://%s:%s", networkCfg.ServerScheme(), hostname, responses.Port)
	default:
		return fmt.Sprintf("%s://%s:%s", networkCfg.ServerScheme(), responses.Host, responses.Port)
	}
}
