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

package delete

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/observiq/bindplane-op/client"
	"github.com/observiq/bindplane-op/model"
)

// Deleter is an interface for deleting resources.
type Deleter interface {
	// DeleteResources deletes resources.
	DeleteResources(ctx context.Context, kind model.Kind, ids []string) error
	// DeleteResourcesFromFile deletes resources from a file.
	DeleteResourcesFromFile(ctx context.Context, filename string) ([]*model.AnyResourceStatus, error)
}

// Builder is an interface for building a Deleter.
type Builder interface {
	// Build returns a new Deleter.
	BuildDeleter(ctx context.Context) (Deleter, error)
}

// NewDeleter returns a new Deleter.
func NewDeleter(client client.BindPlane) *DefaultDeleter {
	return &DefaultDeleter{
		client: client,
	}
}

// DefaultDeleter is the default implementation of Deleter.
type DefaultDeleter struct {
	client client.BindPlane
}

// DeleteResources deletes resources.
func (d *DefaultDeleter) DeleteResources(ctx context.Context, kind model.Kind, ids []string) error {
	if kind == model.KindAgent {
		_, err := d.client.DeleteAgents(ctx, ids)
		return err
	}

	var err error
	for _, id := range ids {
		if err = d.deleteResource(ctx, kind, id); err != nil {
			err = multierror.Append(err, fmt.Errorf("failed to delete: %s", id))
		}
	}

	return err
}

// deleteResource deletes a resource.
func (d *DefaultDeleter) deleteResource(ctx context.Context, kind model.Kind, id string) error {
	switch kind {
	case model.KindConfiguration:
		return d.client.DeleteConfiguration(ctx, id)
	case model.KindSource:
		return d.client.DeleteSource(ctx, id)
	case model.KindProcessor:
		return d.client.DeleteProcessor(ctx, id)
	case model.KindDestination:
		return d.client.DeleteDestination(ctx, id)
	case model.KindSourceType:
		return d.client.DeleteSourceType(ctx, id)
	case model.KindProcessorType:
		return d.client.DeleteProcessorType(ctx, id)
	case model.KindDestinationType:
		return d.client.DeleteDestinationType(ctx, id)
	case model.KindAgentVersion:
		return d.client.DeleteAgentVersion(ctx, id)
	default:
		return fmt.Errorf("unsupported resource kind: %s", kind)
	}
}

// DeleteResourcesFromFile deletes all resources from a file.
func (d *DefaultDeleter) DeleteResourcesFromFile(ctx context.Context, filename string) ([]*model.AnyResourceStatus, error) {
	resources, err := model.ResourcesFromFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read resources: %w", err)
	}

	return d.client.Delete(ctx, resources)
}
