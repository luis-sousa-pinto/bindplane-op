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

package serve

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/observiq/bindplane-op/common"
	"github.com/observiq/bindplane-op/internal/cli"
)

func TestServe(t *testing.T) {
	bindplaneConfig := common.InitConfig("")
	bindplaneConfig.SessionsSecret = "super-secret-key"
	bindplaneConfig.SecretKey = "A3C8DD10-2CF8-4D39-B8E4-63FD5932A169"
	bindplane := cli.NewBindPlane(bindplaneConfig, os.Stdout)

	t.Run("default server", func(t *testing.T) {
		defer func() {
			_ = os.Remove(bindplaneConfig.BoltDatabasePath())
		}()
		serve := Command(bindplane)
		var err error
		go func() {
			err = serve.Execute()
		}()
		time.Sleep(time.Millisecond * 500)
		require.NoError(t, err, "expected server to startup without returning an error")
	})

	// TODO(jsirianni): We need a way to stop gin between tests https://github.com/observiq/bindplane/issues/249
}
