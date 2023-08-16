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

package config

import (
	"errors"
	"fmt"
	"net"
	"time"
)

const (
	// MetricsTypeOTLP is the OTLP metrics type
	MetricsTypeOTLP = "otlp"
	// MetricsTypeNop is the metrics type that does nothing
	MetricsTypeNop = ""
)

// Metrics is the config for sending APM metrics
type Metrics struct {
	// Type is the type of metrics to send
	Type string `mapstructure:"type,omitempty" yaml:"type,omitempty"`

	// Interval is the interval to send metrics at
	Interval time.Duration `mapstructure:"interval,omitempty" yaml:"interval,omitempty"`

	// OTLP is the config for sending OTLP metrics
	OTLP OTLPMetrics `mapstructure:"otlp,omitempty" yaml:"otlp,omitempty"`
}

// OTLPMetrics is the config for sending OTLP metrics
type OTLPMetrics struct {
	// Endpoint is the gRPC endpoint to send metrics to
	Endpoint string `mapstructure:"endpoint,omitempty" yaml:"endpoint,omitempty"`

	// Insecure is whether to use an insecure connection
	Insecure bool `mapstructure:"insecure,omitempty" yaml:"insecure,omitempty"`
}

// Validate validates the metrics config
func (m *Metrics) Validate() error {
	switch m.Type {
	case MetricsTypeOTLP:
		return m.OTLP.validate()
	case MetricsTypeNop:
	default:
		return fmt.Errorf("unknown metrics type: %s", m.Type)
	}

	return nil
}

// validate checks that the gRPC endpoint is set and valid
func (m *OTLPMetrics) validate() error {
	if m.Endpoint == "" {
		return errors.New("metrics endpoint is required")
	}
	_, _, err := net.SplitHostPort(m.Endpoint)
	if err != nil {
		return fmt.Errorf("failed to parse gRPC endpoint: %w", err)
	}

	return nil
}
