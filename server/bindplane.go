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

package server

import (
	"github.com/observiq/bindplane-op/agent"
	"github.com/observiq/bindplane-op/authenticator"
	"github.com/observiq/bindplane-op/otlp/record"
	"github.com/observiq/bindplane-op/store"
	"go.uber.org/zap"
)

// BindPlane is the interface for the BindPlane Server
//
//go:generate mockery --name BindPlane --filename mock_bindplane.go --structname MockBindPlane
type BindPlane interface {
	// Store TODO(doc)
	Store() store.Store
	// Manager TODO(doc)
	Manager() Manager
	// Relayer enables Live messages to flow from Agents to GraphQL subscriptions
	Relayers() Relayers
	// Versions TODO(doc)
	Versions() agent.Versions
	// Logger TODO(doc)
	Logger() *zap.Logger
	// BindPlaneURL returns the URL of the BindPlane server
	BindPlaneURL() string
	// BindPlaneInsecureSkipVerify returns true if the BindPlane server should be contacted without verifying the server's certificate chain and host name
	BindPlaneInsecureSkipVerify() bool
	// WebsocketURL returns the URL of the BindPlane server's websocket endpoint
	WebsocketURL() string
	// SecretKey returns the secret key used to authenticate agents with the BindPlane server
	SecretKey() string

	// Authenticator returns the authenticator for validating user credentials
	Authenticator() authenticator.Authenticator
}

// Relayers is a wrapper around multiple Relayer instances used for different types of results
type Relayers interface {
	Metrics() Relayer[[]*record.Metric]
	Logs() Relayer[[]*record.Log]
	Traces() Relayer[[]*record.Trace]
}

// Relayer forwards results to consumers awaiting the results. It is intentionally generic and is used to support cases where the request for results is decoupled from the response with the results.
type Relayer[T any] interface {
	AwaitResult() (id string, result <-chan T, cancelFunc func())
	SendResult(id string, result T)
}
