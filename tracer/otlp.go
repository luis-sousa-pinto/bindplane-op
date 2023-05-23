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

package tracer

import (
	"context"
	"fmt"
	"net"

	"github.com/observiq/bindplane-op/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// OTLP is a tracer that uses OTLP
type OTLP struct {
	cfg          *config.OTLPTracing
	samplingRate float64
	resource     *resource.Resource
	provider     *oteltrace.TracerProvider
}

// NewOTLP creates a new OTLP Tracer
func NewOTLP(cfg *config.OTLPTracing, samplingRate float64, resource *resource.Resource) *OTLP {
	return &OTLP{
		cfg:          cfg,
		samplingRate: samplingRate,
		resource:     resource,
	}
}

// Start starts the tracer
func (o *OTLP) Start(ctx context.Context) error {
	spanExporter, err := o.createSpanExporter(ctx)
	if err != nil {
		return fmt.Errorf("failed to create span exporter: %w", err)
	}

	sampler := oteltrace.TraceIDRatioBased(o.samplingRate)
	o.provider = oteltrace.NewTracerProvider(
		oteltrace.WithBatcher(spanExporter),
		oteltrace.WithResource(o.resource),
		oteltrace.WithSampler(sampler),
	)

	otel.SetTracerProvider(o.provider)
	return nil
}

// Shutdown shuts down the tracer
func (o *OTLP) Shutdown(ctx context.Context) error {
	if o.provider == nil {
		return nil
	}
	return o.provider.Shutdown(ctx)
}

// createSpanExporter creates a span exporter for OTLP
func (o *OTLP) createSpanExporter(ctx context.Context) (trace.SpanExporter, error) {
	var dialOpts []grpc.DialOption

	// TODO(jsirianni): How to do we handle server side tls, mtls, etc?
	if o.cfg.Insecure {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	_, _, err := net.SplitHostPort(o.cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse gRPC endpoint: %w", err)
	}

	conn, err := grpc.DialContext(ctx, o.cfg.Endpoint, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial otlp endpoint: %w", err)
	}

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	return exporter, nil
}
