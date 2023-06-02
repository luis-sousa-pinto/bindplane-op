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

package version

import (
	"context"

	"github.com/observiq/bindplane-op/client"
	"github.com/observiq/bindplane-op/version"
)

// Versioner is an interface for retrieving BindPlane versions.
type Versioner interface {
	// GetServerVersion returns the server version.
	GetServerVersion(ctx context.Context) (version.Version, error)
	// GetClientVersion returns the client version.
	GetClientVersion(ctx context.Context) (version.Version, error)
}

// Builder is an interface for building a Versioner.
type Builder interface {
	// Build returns a new Versioner.
	BuildVersioner(ctx context.Context) (Versioner, error)
}

// NewVersioner returns a new Versioner.
func NewVersioner(client client.BindPlane) Versioner {
	return &defaultVersioner{
		client: client,
	}
}

// defaultVersioner is the default implementation of the Versioner interface.
type defaultVersioner struct {
	client client.BindPlane
}

// GetServerVersion returns the server version.
func (v *defaultVersioner) GetServerVersion(ctx context.Context) (version.Version, error) {
	clientVersion, err := v.client.Version(ctx)
	if err != nil {
		return clientVersion, err
	}

	return clientVersion, nil
}

// GetClientVersion returns the client version.
func (v *defaultVersioner) GetClientVersion(_ context.Context) (version.Version, error) {
	return version.NewVersion(), nil
}
