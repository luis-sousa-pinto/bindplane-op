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

package apply

import (
	"context"
	"fmt"
	"io"

	"github.com/observiq/bindplane-op/client"
	"github.com/observiq/bindplane-op/model"
)

// Applier is an interface for applying resources.
type Applier interface {
	// ApplyResourcesFromFiles applies all resources from a list of files.
	ApplyResourcesFromFiles(ctx context.Context, filenames []string) ([]*model.AnyResourceStatus, error)

	// ApplyResourcesFromReader applies all resources from a reader.
	ApplyResourcesFromReader(ctx context.Context, reader io.Reader) ([]*model.AnyResourceStatus, error)
}

// Builder is an interface for building an Applier.
type Builder interface {
	// Build returns a new Applier.
	BuildApplier(ctx context.Context) (Applier, error)
}

// NewApplier returns a new Applier.
func NewApplier(client client.BindPlane) Applier {
	return &defaultApplier{
		client: client,
	}
}

// defaultApplier is the default implementation of Applier.
type defaultApplier struct {
	client client.BindPlane
}

// ApplyResourcesFromFiles applies all resources from a list of files.
func (a *defaultApplier) ApplyResourcesFromFiles(ctx context.Context, filenames []string) ([]*model.AnyResourceStatus, error) {
	var resources []*model.AnyResource
	for _, filename := range filenames {
		r, err := model.ResourcesFromFile(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to read resources: %w", err)
		}

		resources = append(resources, r...)
	}

	return a.client.Apply(ctx, resources)
}

// ApplyResourcesFromReader applies all resources from a reader.
func (a *defaultApplier) ApplyResourcesFromReader(ctx context.Context, reader io.Reader) ([]*model.AnyResourceStatus, error) {
	resources, err := model.ResourcesFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read resources: %w", err)
	}

	return a.client.Apply(ctx, resources)
}
