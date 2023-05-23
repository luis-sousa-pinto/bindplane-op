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

	"github.com/observiq/bindplane-op/agent"
	"github.com/observiq/bindplane-op/authenticator"
	"github.com/observiq/bindplane-op/config"
	bpserver "github.com/observiq/bindplane-op/server"
	"github.com/observiq/bindplane-op/store"
)

// NewBindPlane creates a new BindPlane Server using the given store and agent versions
func NewBindPlane(cfg *config.Config, logger *zap.Logger, s store.Store, versions agent.Versions) bpserver.BindPlane {
	return &storeBindPlane{
		store: s,
		bindplane: bindplane{
			logger:        logger,
			config:        cfg,
			manager:       bpserver.NewManager(cfg, s, versions, logger),
			relayers:      NewRelayers(logger),
			versions:      versions,
			authenticator: authenticator.NewBasicAuthenticator(cfg.Auth.Username, cfg.Auth.Password),
		},
	}
}

// ----------------------------------------------------------------------
type bindplane struct {
	config        *config.Config
	manager       bpserver.Manager
	logger        *zap.Logger
	versions      agent.Versions
	relayers      *Relayers
	authenticator authenticator.Authenticator
}

// Manager TODO(doc)
func (s *bindplane) Manager() bpserver.Manager {
	return s.manager
}

func (s *bindplane) Relayers() bpserver.Relayers {
	return s.relayers
}

// Logger TODO(doc)
func (s *bindplane) Logger() *zap.Logger {
	return s.logger
}

// Config TODO(doc)
func (s *bindplane) Config() *config.Config {
	return s.config
}

func (s *storeBindPlane) Authenticator() authenticator.Authenticator {
	return s.authenticator
}

// ----------------------------------------------------------------------

type storeBindPlane struct {
	store store.Store
	bindplane
}

var _ bpserver.BindPlane = (*storeBindPlane)(nil)

// Store TODO(doc)
func (s *storeBindPlane) Store() store.Store {
	return s.store
}

// Versions TODO(doc)
func (s *storeBindPlane) Versions() agent.Versions {
	return s.versions
}

func (s *storeBindPlane) BindPlaneURL() string {
	return s.config.BindPlaneURL()
}

func (s *storeBindPlane) WebsocketURL() string {
	return s.config.Network.WebsocketURL()
}

func (s *storeBindPlane) SecretKey() string {
	return s.config.Auth.SecretKey
}

func (s *storeBindPlane) BindPlaneInsecureSkipVerify() bool {
	return s.config.BindPlaneInsecureSkipVerify()
}
