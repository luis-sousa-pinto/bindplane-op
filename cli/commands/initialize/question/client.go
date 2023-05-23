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
	"github.com/AlecAivazis/survey/v2"
	"github.com/observiq/bindplane-op/config"
)

// ClientAnswers are answers to the default client survey.
type ClientAnswers struct {
	ServerURL string `survey:"server-url"`
	Username  string `survey:"username"`
	Password  string `survey:"password"`
}

// CreateClientQuestions creates the default set of client questions.
func CreateClientQuestions() []RelevantClientQuestion {
	return []RelevantClientQuestion{
		func(_ *ClientAnswers, cfg *config.Config) (*Question, bool) {
			return &Question{
				BeforeText: "URL of the BindPlane OP server",
				Question: survey.Question{
					Name: "server-url",
					Prompt: &survey.Input{
						Message: "Server URL",
						Default: cfg.Network.RemoteURL,
					},
					Validate: survey.Required,
				},
			}, true
		},
		func(_ *ClientAnswers, cfg *config.Config) (*Question, bool) {
			return &Question{
				BeforeText: "Login to access the BindPlane OP server",
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
		func(_ *ClientAnswers, cfg *config.Config) (*Question, bool) {
			return &Question{
				Question: survey.Question{
					Name: "password",
					Prompt: &survey.Password{
						Message: "Password",
					},
					Validate: survey.Required,
				},
			}, true
		},
	}
}
