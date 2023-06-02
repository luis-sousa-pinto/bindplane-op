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

	googletrace "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/observiq/bindplane-op/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/api/option"
)

// GoogleCloud is a tracer that uses Google Cloud Monitoring
type GoogleCloud struct {
	cfg          *config.GoogleCloudTracing
	samplingRate float64
	resource     *resource.Resource
	provider     *oteltrace.TracerProvider
}

// NewGoogleCloud creates a new GoogleCloud Tracer
func NewGoogleCloud(cfg *config.GoogleCloudTracing, samplingRate float64, resource *resource.Resource) *GoogleCloud {
	return &GoogleCloud{
		cfg:          cfg,
		samplingRate: samplingRate,
		resource:     resource,
	}
}

// Start starts the tracer
func (g *GoogleCloud) Start(_ context.Context) error {
	spanExporter, err := g.createSpanExporter()
	if err != nil {
		return fmt.Errorf("failed to create span exporter: %w", err)
	}

	sampler := oteltrace.TraceIDRatioBased(g.samplingRate)
	g.provider = oteltrace.NewTracerProvider(
		oteltrace.WithBatcher(spanExporter),
		oteltrace.WithResource(g.resource),
		oteltrace.WithSampler(sampler),
	)

	otel.SetTracerProvider(g.provider)
	return nil
}

// Shutdown shuts down the tracer
func (g *GoogleCloud) Shutdown(ctx context.Context) error {
	if g.provider == nil {
		return nil
	}
	return g.provider.Shutdown(ctx)
}

// createSpanExporter creates a new span exporter for Google Cloud
func (g *GoogleCloud) createSpanExporter() (trace.SpanExporter, error) {
	return googletrace.New(
		googletrace.WithProjectID(g.cfg.ProjectID),
		googletrace.WithTraceClientOptions([]option.ClientOption{
			option.WithCredentialsFile(g.cfg.CredentialsFile),
		}),
	)
}
