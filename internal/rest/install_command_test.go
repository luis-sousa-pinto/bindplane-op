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

package rest

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInstallCommand_K8sDaemonset(t *testing.T) {
	t.Run("without configuration set", func(t *testing.T) {
		params := installCommandParameters{
			platform: kubernetesDaemonset,
		}
		_, err := params.installCommand()
		require.ErrorIs(t, err, ErrConfigurationNotSet)
	})
}
