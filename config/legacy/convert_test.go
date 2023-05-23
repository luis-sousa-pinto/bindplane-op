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

package legacy

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/observiq/bindplane-op/config"
	modelversion "github.com/observiq/bindplane-op/model/version"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestConvert(t *testing.T) {
	testCases := []struct {
		name     string
		home     string
		legacy   Config
		expected config.Config
	}{
		{
			name: "default",
			legacy: Config{
				Server: Server{
					Common: Common{
						Env:      "production",
						Host:     "localhost",
						Port:     "9000",
						Username: "admin",
						Password: "admin",
						TLSConfig: TLSConfig{
							Certificate:          "cert",
							PrivateKey:           "private",
							CertificateAuthority: []string{"ca"},
							InsecureSkipVerify:   true,
						},
						TraceType: "google",
						GoogleCloudTracing: GoogleCloudTracing{
							ProjectID:       "project",
							CredentialsFile: "creds",
						},
						LogFilePath: "logfile",
						LogOutput:   "file",
					},
					StoreType:                 "bbolt",
					SecretKey:                 "secret",
					SessionsSecret:            "sessions",
					StorageFilePath:           "storage",
					SyncAgentVersionsInterval: time.Hour,
					Offline:                   true,
				},
				Client: Client{
					Common: Common{
						ServerURL: "https://localhost:9000",
						Username:  "admin",
						Password:  "admin",
						TLSConfig: TLSConfig{
							Certificate:          "cert",
							PrivateKey:           "private",
							CertificateAuthority: []string{"ca"},
							InsecureSkipVerify:   true,
						},
					},
				},
				Command: Command{
					Output: "json",
				},
			},
			expected: config.Config{
				APIVersion: modelversion.V1,
				Env:        "production",
				Output:     "json",
				Offline:    true,
				Auth: config.Auth{
					Username:      "admin",
					Password:      "admin",
					SecretKey:     "secret",
					SessionSecret: "sessions",
				},
				Network: config.Network{
					Host: "localhost",
					Port: "9000",
					TLS: config.TLS{
						Certificate:          "cert",
						PrivateKey:           "private",
						CertificateAuthority: []string{"ca"},
						InsecureSkipVerify:   true,
					},
				},
				AgentVersions: config.AgentVersions{
					SyncInterval: time.Hour,
				},
				Store: config.Store{
					Type: "bbolt",
					BBolt: config.BBolt{
						Path: "storage",
					},
				},
				Tracing: config.Tracing{
					Type: "google",
					GoogleCloud: config.GoogleCloudTracing{
						ProjectID:       "project",
						CredentialsFile: "creds",
					},
				},
				Logging: config.Logging{
					FilePath: "logfile",
					Output:   "file",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, Convert(tc.legacy))
		})
	}
}

func TestConvertFile(t *testing.T) {
	filepath := filepath.Join(t.TempDir(), "config.yaml")
	legacyContents := `host: 127.0.0.1
port: "3001"
serverURL: http://127.0.0.1:3001
username: admin
password: adminpass
server:
  secretKey: b1a71608-e80e-46dd-bc51-59c5a5634d25
  remoteURL: ws://127.0.0.1:3001
  sessionsSecret: 5d83d350-288b-4da6-b237-2f8e9b6b42a5
  authType: system
`

	expectedCfg := config.Config{
		ProfileName: "config",
		APIVersion:  modelversion.V1,
		Auth: config.Auth{
			Username:      "admin",
			Password:      "adminpass",
			SecretKey:     "b1a71608-e80e-46dd-bc51-59c5a5634d25",
			SessionSecret: "5d83d350-288b-4da6-b237-2f8e9b6b42a5",
		},
		Network: config.Network{
			Host:      "127.0.0.1",
			Port:      "3001",
			RemoteURL: "http://127.0.0.1:3001",
		},
	}

	legacyBytes := []byte(legacyContents)

	err := os.WriteFile(filepath, legacyBytes, 0644)
	require.NoError(t, err)

	err = ConvertFile(filepath)
	require.NoError(t, err)

	newBytes, err := os.ReadFile(filepath)
	require.NoError(t, err)

	var newCfg config.Config
	err = yaml.Unmarshal(newBytes, &newCfg)
	require.NoError(t, err)

	require.Equal(t, expectedCfg, newCfg)

	// Verify backup
	backupCfg := fmt.Sprintf("%s.backup", filepath)
	require.FileExists(t, backupCfg)
	backupData, err := os.ReadFile(backupCfg)
	require.NoError(t, err)
	require.Equal(t, legacyBytes, backupData)
}
