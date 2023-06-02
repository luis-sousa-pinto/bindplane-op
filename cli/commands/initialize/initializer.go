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

package initialize

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/observiq/bindplane-op/cli/commands/initialize/question"
	"github.com/observiq/bindplane-op/cli/commands/profile"
	"github.com/observiq/bindplane-op/config"
	"gopkg.in/yaml.v3"
)

// Initializer is an interface for initializing a server or client
type Initializer interface {
	// InitializeServer initializes a server config
	InitializeServer(ctx context.Context) error

	// InitializeClient initializes a client config
	InitializeClient(ctx context.Context) error
}

// Builder is an interface for building an Initializer
type Builder interface {
	// Build returns a new Initializer
	BuildInitializer(ctx context.Context) (Initializer, error)
}

// NewInitializer returns a new Initializer
func NewInitializer(cfg *config.Config, cfgPath string, profiler profile.Profiler) Initializer {
	return &defaultInitializer{
		cfg:             cfg,
		cfgPath:         cfgPath,
		clientQuestions: question.CreateClientQuestions(),
		serverQuestions: question.CreateServerQuestions(),
		profiler:        profiler,
		stdio:           terminal.Stdio{In: os.Stdin, Out: os.Stdout, Err: os.Stderr},
	}
}

// defaultInitializer is the default implementation of Initializer
type defaultInitializer struct {
	cfg             *config.Config
	cfgPath         string
	clientQuestions []question.RelevantClientQuestion
	serverQuestions []question.RelevantServerQuestion
	profiler        profile.Profiler
	stdio           terminal.Stdio
}

// InitializeClient initializes a client config
func (i *defaultInitializer) InitializeClient(ctx context.Context) error {
	answers := &question.ClientAnswers{}
	if err := i.askClientQuestions(i.clientQuestions, answers); err != nil {
		return err
	}

	i.cfg.Network.RemoteURL = answers.ServerURL
	i.cfg.Auth.Username = answers.Username
	i.cfg.Auth.Password = answers.Password

	if err := i.writeConfig(ctx); err != nil {
		return fmt.Errorf("failed to write responses to config: %w", err)
	}

	return nil
}

// InitializeServer initializes a server config
func (i *defaultInitializer) InitializeServer(ctx context.Context) error {
	answers := &question.ServerAnswers{}
	if err := i.askServerQuestions(i.serverQuestions, answers); err != nil {
		return err
	}

	i.cfg.Network.Host = answers.Host
	i.cfg.Network.Port = answers.Port
	i.cfg.Auth.Username = answers.Username
	i.cfg.Auth.SecretKey = answers.SecretKey
	i.cfg.Auth.SessionSecret = answers.SessionsSecret
	i.cfg.Network.RemoteURL = answers.RemoteURL

	if answers.Password != "" {
		i.cfg.Auth.Password = answers.Password
	}

	if err := i.writeConfig(ctx); err != nil {
		return fmt.Errorf("failed to write responses to config: %w", err)
	}

	return nil
}

// writeConfig writes the config to disk and sets the current profile
func (i *defaultInitializer) writeConfig(ctx context.Context) error {
	bytes, err := yaml.Marshal(i.cfg)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(i.cfgPath, bytes, 0600); err != nil {
		return err
	}

	// If a profile exists by this name then set it as the current.
	// This maybe the case if using a --config flag and pointing at a file named different than the profile
	if i.profiler.ProfileExists(ctx, i.cfg.ProfileName) {
		return i.profiler.SetCurrentProfileName(ctx, i.cfg.ProfileName)
	}
	return nil
}

// askServerQuestions asks the user questions and records their answers.
func (i *defaultInitializer) askServerQuestions(questions []question.RelevantServerQuestion, response *question.ServerAnswers) error {
	for _, q := range questions {
		question, relevant := q(response, i.cfg)
		if !relevant {
			continue
		}

		if question.BeforeText != "" {
			fmt.Println()
			fmt.Println(question.BeforeText)
		}

		surveyQuestion := []*survey.Question{&question.Question}
		stdioOpt := survey.WithStdio(i.stdio.In, i.stdio.Out, i.stdio.Err)
		if err := survey.Ask(surveyQuestion, response, stdioOpt); err != nil {
			return fmt.Errorf("%s input failed: %w", question.Name, err)
		}
	}
	return nil
}

// askClientQuestions asks the user questions and records their answers.
func (i *defaultInitializer) askClientQuestions(questions []question.RelevantClientQuestion, response *question.ClientAnswers) error {
	for _, q := range questions {
		question, relevant := q(response, i.cfg)
		if !relevant {
			continue
		}

		if question.BeforeText != "" {
			fmt.Println()
			fmt.Println(question.BeforeText)
		}

		surveyQuestion := []*survey.Question{&question.Question}
		stdioOpt := survey.WithStdio(i.stdio.In, i.stdio.Out, i.stdio.Err)
		if err := survey.Ask(surveyQuestion, response, stdioOpt); err != nil {
			return fmt.Errorf("%s input failed: %w", question.Name, err)
		}
	}
	return nil
}
