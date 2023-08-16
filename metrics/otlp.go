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

package metrics

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/observiq/bindplane-op/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	api "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// OTLP is an OTLP metrics provider
type OTLP struct {
	cfg      *config.OTLPMetrics
	mp       *metric.MeterProvider
	interval time.Duration
	resource *resource.Resource
}

// NewOTLP returns a new OTLP metrics provider
func NewOTLP(cfg *config.OTLPMetrics, interval time.Duration, resource *resource.Resource) (Provider, error) {
	cloneCfg := *cfg

	_, _, err := net.SplitHostPort(cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse gRPC endpoint: %w", err)
	}

	return &OTLP{
		cfg:      &cloneCfg,
		interval: interval,
		resource: resource,
	}, nil
}

// Start just sets the provider as the OTel global meter provider
func (p *OTLP) Start(ctx context.Context) error {
	var dialOpts []grpc.DialOption

	if p.cfg.Insecure {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.DialContext(ctx, p.cfg.Endpoint, dialOpts...)
	if err != nil {
		return fmt.Errorf("failed to dial collector: %w", err)
	}

	exp, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		return fmt.Errorf("failed to create metrics exporter: %w", err)
	}
	p.mp = api.NewMeterProvider(
		api.WithReader(
			api.NewPeriodicReader(exp, api.WithInterval(p.interval)),
		),
		api.WithResource(p.resource),
	)

	otel.SetMeterProvider(p.mp)
	return nil
}

// Shutdown shuts down the provider
func (p *OTLP) Shutdown(ctx context.Context) error {
	if p.mp == nil {
		return nil
	}
	return p.mp.Shutdown(ctx)
}
