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
	"fmt"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/observiq/bindplane-op/internal/otlp/record"
	"github.com/observiq/bindplane-op/internal/server"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("otlp")

// AddRoutes adds endpoints for receiving OTLP formatted (compressed grpc) telemetry signals.
func AddRoutes(router gin.IRouter, bindplane server.BindPlane) error {
	router.POST("/otlphttp/v1/logs", func(c *gin.Context) { logs(c, bindplane) })
	router.POST("/otlphttp/v1/metrics", func(c *gin.Context) { metrics(c, bindplane) })
	router.POST("/otlphttp/v1/traces", func(c *gin.Context) { traces(c, bindplane) })
	return nil
}

func logs(c *gin.Context, bindplane server.BindPlane) {
	_, span := tracer.Start(c.Request.Context(), "otlp/logs")
	defer span.End()

	otlpLogs := plogotlp.NewExportRequest()
	if err := otlpParse(c, bindplane, otlpLogs); err != nil {
		c.Error(err)
		return
	}

	logs := record.ConvertLogs(otlpLogs.Logs())
	if err := relay(c, bindplane.Relayers().Logs, logs); err != nil {
		c.Error(err)
		return
	}

	c.Status(200)
}

func metrics(c *gin.Context, bindplane server.BindPlane) {
	_, span := tracer.Start(c.Request.Context(), "otlp/metrics")
	defer span.End()

	otlpMetrics := pmetricotlp.NewExportRequest()
	if err := otlpParse(c, bindplane, otlpMetrics); err != nil {
		c.Error(err)
		return
	}

	metrics := record.ConvertMetrics(otlpMetrics.Metrics())

	// could be snapshot metrics or agent metrics
	if isSnapshotMetrics(c, metrics) {
		if err := relay(c, bindplane.Relayers().Metrics, metrics); err != nil {
			c.Error(err)
			return
		}
	} else if metrics := getAgentMetrics(c, metrics); len(metrics) > 0 {
		measurements := bindplane.Store().Measurements()
		if measurements != nil {
			if err := measurements.SaveAgentMetrics(c, metrics); err != nil {
				c.Error(err)
				return
			}
		}

	}

	c.Status(200)
}

func getAgentMetrics(_ *gin.Context, metrics []*record.Metric) []*record.Metric {
	result := []*record.Metric{}

	for _, metric := range metrics {
		if strings.HasPrefix(metric.Name, "otelcol_processor_throughputmeasurement_") {
			result = append(result, metric)
		}
	}

	return result
}

func traces(c *gin.Context, bindplane server.BindPlane) {
	_, span := tracer.Start(c.Request.Context(), "otlp/traces")
	defer span.End()

	otlpTraces := ptraceotlp.NewExportRequest()
	if err := otlpParse(c, bindplane, otlpTraces); err != nil {
		c.Error(err)
		return
	}

	traces := record.ConvertTraces(otlpTraces.Traces())
	if err := relay(c, bindplane.Relayers().Traces, traces); err != nil {
		c.Error(err)
		return
	}

	c.Status(200)
}

type unmarshalProto interface {
	UnmarshalProto(data []byte) error
}

func isSnapshotMetrics(c *gin.Context, _ []*record.Metric) bool {
	sessionID := c.Request.Header.Get("X-Bindplane-Session-Id")
	return sessionID != ""
}

func otlpParse[T unmarshalProto](c *gin.Context, _ server.BindPlane, result T) error {
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

func relay[T any](c *gin.Context, relayer *server.Relayer[T], result T) error {
	// check for a recent logs session id
	if sessionID := c.Request.Header.Get(server.HeaderSessionID); sessionID != "" {
		relayer.SendResult(sessionID, result)
	}
	return nil
}
