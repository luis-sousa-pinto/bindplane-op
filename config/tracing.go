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
)

const (
	// TracerTypeGoogleCloud is the type of tracer that sends traces to Google Cloud Monitoring
	TracerTypeGoogleCloud = "google"

	// TracerTypeOTLP is the type of tracer that sends traces to an OTLP endpoint
	TracerTypeOTLP = "otlp"

	// TracerTypeNop is the type of tracer that does nothing
	TracerTypeNop = ""
)

// Tracing is the configuration for tracing
type Tracing struct {
	// Type specifies the type of tracing to use.
	Type string `mapstructure:"type,omitempty" yaml:"type,omitempty"`

	// SamplingRate is the rate at which traces are sampled. Valid values are between 0 and 1.
	SamplingRate float64 `mapstructure:"samplingRate,omitempty" yaml:"samplingRate,omitempty"`

	// GoogleCloud is used to send traces to Google Cloud when TraceType is set to "google".
	GoogleCloud GoogleCloudTracing `mapstructure:"googleCloud,omitempty" yaml:"googleCloud,omitempty"`

	// OTLP is used to send traces to an Open Telemetry OTLP receiver when TraceType is set to "otlp".
	OTLP OTLPTracing `mapstructure:"otlp,omitempty" yaml:"otlp,omitempty"`
}

// Validate validates the tracing configuration.
func (t *Tracing) Validate() error {
	switch t.Type {
	case TracerTypeNop:
		return nil
	case TracerTypeGoogleCloud:
		if err := t.GoogleCloud.Validate(); err != nil {
			return err
		}
	case TracerTypeOTLP:
		if err := t.OTLP.Validate(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid tracing type: %s", t.Type)
	}

	if t.SamplingRate < 0 || t.SamplingRate > 1 {
		return errors.New("tracing sampling rate must be between 0 and 1")
	}

	return nil
}

// OTLPTracing is the configuration for tracing to an OTLP endpoint
type OTLPTracing struct {
	// Endpoint is the OTLP endpoint to send traces to.
	Endpoint string `mapstructure:"endpoint,omitempty" yaml:"endpoint,omitempty"`

	// Insecure disables TLS verification
	Insecure bool `mapstructure:"insecure,omitempty" yaml:"insecure,omitempty"`
}

// Validate validates the OTLP tracing configuration.
func (t *OTLPTracing) Validate() error {
	if t.Endpoint == "" {
		return errors.New("OTLP endpoint must be set for OTLP tracing")
	}

	return nil
}

// GoogleCloudTracing is the configuration for tracing to Google Cloud Monitoring
type GoogleCloudTracing struct {
	// ProjectID is the Google Cloud project ID to use when sending traces.
	ProjectID string `mapstructure:"projectID,omitempty" yaml:"projectID,omitempty"`

	// CredentialsFile is the path to the Google Cloud credentials file to use when sending traces.
	CredentialsFile string `mapstructure:"credentialsFile,omitempty" yaml:"credentialsFile,omitempty"`
}

// Validate validates the Google Cloud tracing configuration.
func (t *GoogleCloudTracing) Validate() error {
	if t.ProjectID == "" {
		return errors.New("project ID must be set for Google Cloud tracing")
	}

	return nil
}
