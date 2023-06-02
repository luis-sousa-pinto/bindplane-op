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

package config

import (
	"testing"

	modelversion "github.com/observiq/bindplane-op/model/version"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	c := NewConfig()
	require.Equal(t, modelversion.V1, c.APIVersion)
}
