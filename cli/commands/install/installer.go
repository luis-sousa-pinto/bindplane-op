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

package install

import (
	"context"

	"github.com/observiq/bindplane-op/client"
)

// Installer is an interface for getting the install command for an agent
type Installer interface {
	// GetAgentInstallCommand returns the install command for an agent
	GetAgentInstallCommand(ctx context.Context, opts client.AgentInstallOptions) (string, error)
}

// Builder is an interface for building an Installer
type Builder interface {
	// Build returns a new Installer
	BuildInstaller(ctx context.Context) (Installer, error)
}

// NewInstaller returns a new Installer
func NewInstaller(client client.BindPlane) Installer {
	return &defaultInstaller{
		client: client,
	}
}

// defaultInstaller is the default implementation of Installer
type defaultInstaller struct {
	client client.BindPlane
}

// GetAgentInstallCommand returns the install command for an agent
func (i *defaultInstaller) GetAgentInstallCommand(ctx context.Context, opts client.AgentInstallOptions) (string, error) {
	return i.client.AgentInstallCommand(ctx, opts)
}
