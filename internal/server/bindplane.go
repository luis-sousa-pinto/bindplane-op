// Copyright  observIQ, Inc
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

// Package server contains the interface and implementation of the BindPlane Server
package server

import (
	"go.uber.org/zap"

	"github.com/observiq/bindplane-op/common"
	"github.com/observiq/bindplane-op/internal/agent"
	"github.com/observiq/bindplane-op/internal/store"
)

// BindPlane is the interface for the BindPlane Server
type BindPlane interface {
	// Store TODO(doc)
	Store() store.Store
	// Manager TODO(doc)
	Manager() Manager
	// Relayer enables Live messages to flow from Agents to GraphQL subscriptions
	Relayers() *Relayers
	// Versions TODO(doc)
	Versions() agent.Versions
	// Config TODO(doc)
	Config() *common.Server
	// Logger TODO(doc)
	Logger() *zap.Logger
}

// NewBindPlane creates a new BindPlane Server using the given store and agent versions
func NewBindPlane(config *common.Server, logger *zap.Logger, s store.Store, versions agent.Versions) (BindPlane, error) {
	manager, err := NewManager(config, s, versions, logger)
	if err != nil {
		return nil, err
	}

	return &storeBindPlane{
		store: s,
		bindplane: bindplane{
			logger:   logger,
			config:   config,
			manager:  manager,
			relayers: NewRelayers(logger),
			versions: versions,
		},
	}, nil
}

// ----------------------------------------------------------------------
type bindplane struct {
	config   *common.Server
	manager  Manager
	logger   *zap.Logger
	versions agent.Versions
	relayers *Relayers
}

// Manager TODO(doc)
func (s *bindplane) Manager() Manager {
	return s.manager
}

func (s *bindplane) Relayers() *Relayers {
	return s.relayers
}

// Logger TODO(doc)
func (s *bindplane) Logger() *zap.Logger {
	return s.logger
}

// Config TODO(doc)
func (s *bindplane) Config() *common.Server {
	return s.config
}

// ----------------------------------------------------------------------

type storeBindPlane struct {
	store store.Store
	bindplane
}

var _ BindPlane = (*storeBindPlane)(nil)

// Store TODO(doc)
func (s *storeBindPlane) Store() store.Store {
	return s.store
}

// Versions TODO(doc)
func (s *storeBindPlane) Versions() agent.Versions {
	return s.versions
}

// ----------------------------------------------------------------------
