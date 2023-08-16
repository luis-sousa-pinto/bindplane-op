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

// Package metrics provides the Provider interface for sending APM metrics
package metrics

import (
	"context"
)

// Provider is a metrics provider
//
//go:generate mockery --name=Provider --filename=mock_provider.go --structname=MockProvider --with-expecter
type Provider interface {
	// Start starts the provider
	Start(context.Context) error
	// Shutdown shuts down the provider
	Shutdown(context.Context) error
}
