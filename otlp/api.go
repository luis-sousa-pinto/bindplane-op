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

// Package otlp provides the HTTP handlers for receiving OTLP telemetry signals.
package otlp

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/observiq/bindplane-op/internal/server"
	"github.com/observiq/bindplane-op/otlp/record"
	exposedserver "github.com/observiq/bindplane-op/server"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	"go.opentelemetry.io/otel"
)

const (
	// HeaderSessionID is the name of the HTTP header where a BindPlane Session ID can be found. This will be used by
	// Relayer to match up the request with an eventual response via OTLP HTTP POST.
	HeaderSessionID = "X-Bindplane-Session-ID"
)

var tracer = otel.Tracer("otlp")

// AddRoutes adds endpoints for receiving OTLP formatted (compressed grpc) telemetry signals.
func AddRoutes(router gin.IRouter, bindplane exposedserver.BindPlane) {
	router.POST("/otlphttp/v1/logs", func(c *gin.Context) { Logs(c, bindplane) })
	router.POST("/otlphttp/v1/metrics", func(c *gin.Context) { Metrics(c, bindplane) })
	router.POST("/otlphttp/v1/traces", func(c *gin.Context) { Traces(c, bindplane) })
}

// Logs handles OTLP formatted log data.
func Logs(c *gin.Context, bindplane exposedserver.BindPlane) {
	traceCtx, span := tracer.Start(c.Request.Context(), "otlp/logs")
	defer span.End()

	otlpLogs := plogotlp.NewExportRequest()
	if err := parse(traceCtx, c, bindplane, otlpLogs); err != nil {
		c.Error(err)
		return
	}

	relayLogs := exposedserver.NewRelayLogs(otlpLogs.Logs())
	if err := relay(c, bindplane.Relayers().Logs(), relayLogs); err != nil {
		c.Error(err)
		return
	}

	c.Status(200)
}

// Metrics handles OTLP formatted metric data.
func Metrics(c *gin.Context, bindplane exposedserver.BindPlane) {
	traceCtx, span := tracer.Start(c.Request.Context(), "otlp/metrics")
	defer span.End()

	otlpMetrics := pmetricotlp.NewExportRequest()
	if err := parse(traceCtx, c, bindplane, otlpMetrics); err != nil {
		c.Error(err)
		return
	}

	metrics := record.ConvertMetrics(traceCtx, otlpMetrics.Metrics())

	// could be snapshot metrics or agent metrics
	if isSnapshotMetrics(c, metrics) {
		relayMetrics := exposedserver.NewRelayMetrics(otlpMetrics.Metrics())
		if err := relay(c, bindplane.Relayers().Metrics(), relayMetrics); err != nil {
			c.Error(err)
			return
		}
	}
	// agent metrics
	metrics, err := getAgentMetrics(metrics)
	if err != nil {
		c.AbortWithError(401, err)
		return
	}

	if len(metrics) > 0 {
		if measurements := bindplane.Store().Measurements(); measurements != nil {
			if err := measurements.SaveAgentMetrics(traceCtx, metrics); err != nil {
				c.Error(err)
				return
			}
		}
	}

	c.Status(200)
}

func getAgentMetrics(metrics []*record.Metric) ([]*record.Metric, error) {
	result := []*record.Metric{}

	for _, metric := range metrics {
		if strings.HasPrefix(metric.Name, "otelcol_processor_throughputmeasurement_") {
			result = append(result, metric)
		}
	}

	return result, nil
}

// Traces handles OTLP formatted trace data.
func Traces(c *gin.Context, bindplane exposedserver.BindPlane) {
	traceCtx, span := tracer.Start(c.Request.Context(), "otlp/traces")
	defer span.End()

	otlpTraces := ptraceotlp.NewExportRequest()
	if err := parse(traceCtx, c, bindplane, otlpTraces); err != nil {
		c.Error(err)
		return
	}

	relayTraces := exposedserver.NewRelayTraces(otlpTraces.Traces())
	if err := relay(c, bindplane.Relayers().Traces(), relayTraces); err != nil {
		c.Error(err)
		return
	}

	c.Status(200)
}

type unmarshalProto interface {
	UnmarshalProto(data []byte) error
}

func isSnapshotMetrics(c *gin.Context, _ []*record.Metric) bool {
	sessionID := c.Request.Header.Get(HeaderSessionID)
	return sessionID != ""
}

func parse[T unmarshalProto](traceCtx context.Context, c *gin.Context, _ exposedserver.BindPlane, result T) error {
	traceCtx, span := tracer.Start(traceCtx, "otlp/parse")
	defer span.End()
	reader := c.Request.Body

	// headers sent by the agent:
	// Agent-Id
	// Component-Id
	// Content-Encoding
	// Content-Length
	// Content-Type
	// User-Agent
	// X-Bindplane-Secret-Key
	// X-Bindplane-Session-Id

	if c.ContentType() != "application/protobuf" && c.ContentType() != "application/x-protobuf" {
		return fmt.Errorf("otlp endpoints do not support content-type %s", c.ContentType())
	}

	if c.Request.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(reader)
		if err != nil {
			return err
		}
		reader = gzipReader
	}

	bytes, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	err = result.UnmarshalProto(bytes)
	if err != nil {
		return err
	}

	return nil
}

func relay[T any](c *gin.Context, relayer exposedserver.Relayer[T], result T) error {
	// check for a recent logs session id
	if sessionID := c.Request.Header.Get(server.HeaderSessionID); sessionID != "" {
		relayer.SendResult(sessionID, result)
	}
	return nil
}
