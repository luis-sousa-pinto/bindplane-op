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

package agent

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSha256Sums(t *testing.T) {
	file, err := os.ReadFile("testfiles/observiq-otel-collector-v1.4.0-SHA256SUMS")
	require.NoError(t, err)
	parsed := ParseSha256Sums(file)
	require.Len(t, parsed, 12)
	require.Equal(t, "245638263d8755d1fd22abbca97686226161492301489d5616f0d9b7b7734e74", parsed.Sha256Sum("observiq-otel-collector_v1.4.0_linux_amd64.deb"))
	require.Equal(t, "dc133923c89ffb8ecb77e367846652a7cb5465544a8d4099f35f6fde777c5954", parsed.Sha256Sum("observiq-otel-collector_v1.4.0_linux_amd64.rpm"))
}
