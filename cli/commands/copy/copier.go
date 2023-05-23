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

package copy

import (
	"context"

	"github.com/observiq/bindplane-op/client"
)

// Copier is an interface for copying resources.
type Copier interface {
	// CopyConfig copies a configuration.
	CopyConfig(ctx context.Context, name, newName string) error
}

// Builder is an interface for building a Copier.
type Builder interface {
	// Build returns a new Copier.
	BuildCopier(ctx context.Context) (Copier, error)
}

// NewCopier returns a new Copier.
func NewCopier(client client.BindPlane) Copier {
	return &defaultCopier{
		client: client,
	}
}

// defaultCopier is the default implementation of Copier.
type defaultCopier struct {
	client client.BindPlane
}

// CopyConfig copies a configuration.
func (c *defaultCopier) CopyConfig(ctx context.Context, name, newName string) error {
	return c.client.CopyConfig(ctx, name, newName)
}
