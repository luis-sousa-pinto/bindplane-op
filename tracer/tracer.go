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

// Package tracer provides tracers for BindPlane
package tracer

import (
	"context"
	"os"
	"runtime"

	bpversion "github.com/observiq/bindplane-op/version"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// Tracer is an interface used for tracing
//
//go:generate mockery --name=Tracer --filename=mock_trace.go --structname=MockTracer
type Tracer interface {
	// Start starts the tracer
	Start(ctx context.Context) error

	// Shutdown shuts down the tracer
	Shutdown(ctx context.Context) error
}

// DefaultResource returns the default resource for the tracer
func DefaultResource() *resource.Resource {
	hostname, _ := os.Hostname()
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("bindplane"),
		semconv.ServiceVersionKey.String(bpversion.NewVersion().String()),
		semconv.HostArchKey.String(runtime.GOARCH),
		semconv.HostNameKey.String(hostname),
	)
}
