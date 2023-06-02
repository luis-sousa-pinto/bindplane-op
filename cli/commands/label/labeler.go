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

package label

import (
	"context"
	"fmt"

	"github.com/observiq/bindplane-op/client"
	"github.com/observiq/bindplane-op/model"
)

// Labeler is an interface for labeling resources.
type Labeler interface {
	// GetAgentLabels returns the labels for an agent.
	GetAgentLabels(ctx context.Context, id string) (*model.Labels, error)
	// ApplyAgentLabels applies labels to an agent.
	ApplyAgentLabels(ctx context.Context, id string, labels map[string]string, overwrite bool) (*model.Labels, error)
}

// Builder is an interface for building a Labeler.
type Builder interface {
	// Build returns a new Labeler.
	BuildLabeler(ctx context.Context) (Labeler, error)
}

// NewLabeler returns a new Labeler.
func NewLabeler(client client.BindPlane) Labeler {
	return &defaultLabeler{
		client: client,
	}
}

// defaultLabeler is the default implementation of Labeler.
type defaultLabeler struct {
	client client.BindPlane
}

// GetAgentLabels returns the labels for an agent.
func (l *defaultLabeler) GetAgentLabels(ctx context.Context, id string) (*model.Labels, error) {
	return l.client.AgentLabels(ctx, id)
}

// ApplyAgentLabels applies labels to an agent.
func (l *defaultLabeler) ApplyAgentLabels(ctx context.Context, id string, labels map[string]string, overwrite bool) (*model.Labels, error) {
	agentLabels, err := model.LabelsFromMap(labels)
	if err != nil {
		return nil, fmt.Errorf("failed to create labels: %w", err)
	}

	return l.client.ApplyAgentLabels(ctx, id, &agentLabels, overwrite)
}
