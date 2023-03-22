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

package serve

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/observiq/bindplane-op/common"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"gopkg.in/yaml.v3"
)

func TestEnsureSecretKey(t *testing.T) {
	t.Run("Secret key exists", func(t *testing.T) {
		s := Server{
			logger: zaptest.NewLogger(t),
		}

		bindplaneConfig := common.InitConfig("")
		bindplaneConfig.SecretKey = "A3C8DD10-2CF8-4D39-B8E4-63FD5932A169"

		require.NoError(t, s.ensureSecretKey(bindplaneConfig, ""))
	})

	t.Run("Secret key doesn't exist, file doesn't exist", func(t *testing.T) {
		s := Server{
			logger: zaptest.NewLogger(t),
		}

		bindplaneConfig := common.InitConfig("")

		require.ErrorContains(t, s.ensureSecretKey(bindplaneConfig, ""), "secret key is required")
	})

	t.Run("Secret key doesn't exist, file exists", func(t *testing.T) {
		s := Server{
			logger: zaptest.NewLogger(t),
		}
		const defaultConfig = `host: "127.0.0.1"
port: "3001"
serverURL: http://127.0.0.1:3001
logFilePath: /var/log/bindplane/bindplane.log
server:
    remoteURL: ws://127.0.0.1:3001
    storageFilePath: /var/lib/bindplane/storage/bindplane.db
`

		const expectedConfig = `host: 127.0.0.1
port: "3001"
serverURL: http://127.0.0.1:3001
logFilePath: /var/log/bindplane/bindplane.log
server:
    storageFilePath: /var/lib/bindplane/storage/bindplane.db
    secretKey: %s
    remoteURL: ws://127.0.0.1:3001
`

		bindplaneConfig := common.InitConfig("")

		err := yaml.Unmarshal([]byte(defaultConfig), bindplaneConfig)
		require.NoError(t, err)

		bindplaneConfig.Server.Host = "0.0.0.0"

		tmpDir := t.TempDir()
		confFilePath := filepath.Join(tmpDir, "test-conf.yaml")

		// Create and write default config
		confFile, err := os.Create(confFilePath)
		require.NoError(t, err)
		_, err = confFile.Write([]byte(defaultConfig))
		require.NoError(t, err)
		require.NoError(t, confFile.Close())

		require.NoError(t, s.ensureSecretKey(bindplaneConfig, confFilePath))
		require.NotEmpty(t, bindplaneConfig.SecretKey)

		confBytes, err := os.ReadFile(confFilePath)
		require.NoError(t, err)

		fullExpectedConfig := fmt.Sprintf(expectedConfig, bindplaneConfig.SecretKey)
		t.Log(string(confBytes))
		require.Equal(t, fullExpectedConfig, string(confBytes))
	})

}
