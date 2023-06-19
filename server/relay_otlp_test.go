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
	"testing"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

func TestRelayMetrics(t *testing.T) {
	metrics := pmetric.NewMetrics()

	rm := metrics.ResourceMetrics().AppendEmpty()
	rm.Resource().Attributes().PutBool("test", true)
	sm := rm.ScopeMetrics().AppendEmpty()
	m := sm.Metrics().AppendEmpty()
	g := m.SetEmptyGauge()
	dp := g.DataPoints().AppendEmpty()
	dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	dp.SetIntValue(10)

	marshaler := pmetric.JSONMarshaler{}
	expected, err := marshaler.MarshalMetrics(metrics)
	require.NoError(t, err)

	relayMetrics := RelayMetrics(metrics)

	actual, err := jsoniter.Marshal(&relayMetrics)
	require.NoError(t, err)
	require.Equal(t, expected, actual)

	var newMetrics RelayMetrics
	err = jsoniter.Unmarshal(actual, &newMetrics)
	require.NoError(t, err)

	require.Equal(t, metrics, newMetrics.OTLP())
}

func TestRelayLogs(t *testing.T) {
	logs := plog.NewLogs()

	rl := logs.ResourceLogs().AppendEmpty()
	rl.Resource().Attributes().PutBool("test", true)
	sl := rl.ScopeLogs().AppendEmpty()
	lr := sl.LogRecords().AppendEmpty()
	lr.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	lr.Attributes().PutStr("one", "two")

	marshaler := plog.JSONMarshaler{}
	expected, err := marshaler.MarshalLogs(logs)
	require.NoError(t, err)

	relayLogs := RelayLogs(logs)

	actual, err := jsoniter.Marshal(&relayLogs)
	require.NoError(t, err)
	require.Equal(t, expected, actual)

	var newLogs RelayLogs
	err = jsoniter.Unmarshal(actual, &newLogs)
	require.NoError(t, err)

	require.Equal(t, logs, newLogs.OTLP())
}

func TestRelayTraces(t *testing.T) {
	traces := ptrace.NewTraces()

	rs := traces.ResourceSpans().AppendEmpty()
	rs.Resource().Attributes().PutBool("test", true)
	ss := rs.ScopeSpans().AppendEmpty()
	sp := ss.Spans().AppendEmpty()
	sp.SetStartTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	sp.Attributes().PutStr("one", "two")

	marshaler := ptrace.JSONMarshaler{}
	expected, err := marshaler.MarshalTraces(traces)
	require.NoError(t, err)

	relayTraces := RelayTraces(traces)

	actual, err := jsoniter.Marshal(&relayTraces)
	require.NoError(t, err)
	require.Equal(t, expected, actual)

	var newTraces RelayTraces
	err = jsoniter.Unmarshal(actual, &newTraces)
	require.NoError(t, err)

	require.Equal(t, traces, newTraces.OTLP())
}
