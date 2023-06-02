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

package tracer

import (
	"os"
	"runtime"
	"testing"

	bpversion "github.com/observiq/bindplane-op/version"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
)

func TestDefaultResource(t *testing.T) {
	hostname, _ := os.Hostname()
	resource := DefaultResource()
	attributes := resource.Attributes()
	expected := []attribute.KeyValue{
		{
			Key:   attribute.Key("host.arch"),
			Value: attribute.StringValue(runtime.GOARCH),
		},
		{
			Key:   attribute.Key("host.name"),
			Value: attribute.StringValue(hostname),
		},
		{
			Key:   attribute.Key("service.name"),
			Value: attribute.StringValue("bindplane"),
		},
		{
			Key:   attribute.Key("service.version"),
			Value: attribute.StringValue(bpversion.NewVersion().String()),
		},
	}

	require.Equal(t, expected, attributes)
}
