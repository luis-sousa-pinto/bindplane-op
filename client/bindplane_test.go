// Copyright  observIQ, Inc
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

package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/observiq/bindplane-op/config"
	"github.com/observiq/bindplane-op/model"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewBindPlane(t *testing.T) {
	cases := []struct {
		name      string
		cfg       *config.Config
		logger    *zap.Logger
		expect    BindPlane
		expectErr string
	}{
		{
			name: "default",
			cfg: &config.Config{
				Network: config.Network{
					Host: "localhost",
					Port: "3001",
				},
				Auth: config.Auth{
					Username: "admin",
					Password: "admin",
				},
			},
			logger: zap.NewNop(),
			expect: &BindplaneClient{},
		},
		{
			name: "fields",
			cfg: &config.Config{
				Network: config.Network{
					Host: "10.99.4.5",
					Port: "2000",
				},
				Auth: config.Auth{
					Username: "devel",
				},
			},
			logger: zap.NewNop(),
			expect: &BindplaneClient{},
		},
		{
			name: "tls",
			cfg: &config.Config{
				Network: config.Network{
					TLS: config.TLS{
						Certificate: "../cli/commands/serve/testdata/bindplane.crt",
						PrivateKey:  "../cli/commands/serve/testdata/bindplane.key",
						CertificateAuthority: []string{
							"../cli/commands/serve/testdata/bindplane-ca.crt",
						},
					},
				},
			},
			logger: zap.NewNop(),
			expect: &BindplaneClient{},
		},
		{
			name: "tls-invalid-cert-path",
			cfg: &config.Config{
				Network: config.Network{
					TLS: config.TLS{
						Certificate: "../cli/commands/serve/testdata/bindplane.crt.invalid",
						PrivateKey:  "../cli/commands/serve/testdata/bindplane.key",
					},
				},
			},
			logger:    zap.NewNop(),
			expectErr: "failed to configure TLS client: failed to load tls certificate: open ../cli/commands/serve/testdata/bindplane.crt.invalid",
		},
		{
			name: "tls-invalid-key-path",
			cfg: &config.Config{
				Network: config.Network{
					TLS: config.TLS{
						Certificate: "../cli/commands/serve/testdata/bindplane.crt",
						PrivateKey:  "../cli/commands/serve/testdata/bindplane.key.invalid",
					},
				},
			},
			logger:    zap.NewNop(),
			expectErr: "failed to configure TLS client: failed to load tls certificate: open ../cli/commands/serve/testdata/bindplane.key.invalid",
		},
		{
			name: "tls-invalid-ca-path",
			cfg: &config.Config{
				Network: config.Network{
					TLS: config.TLS{
						Certificate: "../cli/commands/serve/testdata/bindplane.crt",
						PrivateKey:  "../cli/commands/serve/testdata/bindplane.key",
						CertificateAuthority: []string{
							"../cli/commands/serve/testdata/bindplane-ca.crt.invalid",
						},
					},
				},
			},
			logger:    zap.NewNop(),
			expectErr: "failed to configure TLS client: failed to read certificate authority file: open ../cli/commands/serve/testdata/bindplane-ca.crt.invalid",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := NewBindPlane(tc.cfg, tc.logger)
			if tc.expectErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectErr)
				return
			}
			require.NoError(t, err)

			require.NotNil(t, out)
			require.NotNil(t, out.(*BindplaneClient).Client)
			require.NotNil(t, out.(*BindplaneClient).Client)
			require.NotNil(t, out.(*BindplaneClient).Logger)
			require.Equal(t, time.Second*20, out.(*BindplaneClient).Client.GetClient().Timeout)

			if tc.cfg.Auth.Username != "" {
				require.Equal(t, tc.cfg.Auth.Username, out.(*BindplaneClient).Client.UserInfo.Username)
			}
			if tc.cfg.Auth.Password != "" {
				require.Equal(t, tc.cfg.Auth.Password, out.(*BindplaneClient).Client.UserInfo.Password)
			}

			base := fmt.Sprintf("%s/v1", tc.cfg.Network.ServerURL())
			require.Equal(t, base, out.(*BindplaneClient).Client.BaseURL)
		})
	}
}

func TestCopyConfig(t *testing.T) {
	configName, copyName := "my-config", "my-config-copy"

	testCases := []struct {
		description    string
		expectError    bool
		errMsg         string
		responseStatus int
	}{
		{
			"201 Created, no error",
			false,
			"",
			201,
		},
		{
			"409 Conflict, error",
			true,
			"a configuration with name 'my-config-copy' already exists",
			409,
		},
		{
			"409 Conflict, error",
			true,
			"failed to copy configuration, got status 400",
			400,
		},
	}

	for _, test := range testCases {
		handler := func(w http.ResponseWriter, r *http.Request) {
			// Verify the endpoint and method are expected
			require.Equal(t, r.Method, "POST")
			require.Equal(t, fmt.Sprintf("/configurations/%s/copy", configName), r.URL.Path)

			payload := &model.PostCopyConfigRequest{}
			err := jsoniter.NewDecoder(r.Body).Decode(payload)

			// Verify the expected payload
			require.NoError(t, err)
			require.Equal(t, model.PostCopyConfigRequest{
				Name: copyName,
			}, *payload)

			// Write the appropriate response status
			w.WriteHeader(test.responseStatus)
			return
		}

		url, closeFunc := newTestServer(
			handler,
		)
		defer closeFunc()

		bp, err := NewBindPlane(&config.Config{}, zap.NewNop())
		require.NoError(t, err)

		bp.(*BindplaneClient).Client.SetBaseURL(url)

		err = bp.CopyConfig(context.TODO(), "my-config", "my-config-copy")
		if test.expectError {
			require.Error(t, err)
			require.Equal(t, test.errMsg, err.Error())
		} else {
			require.NoError(t, err)
		}
	}
}

func newTestServer(handler http.HandlerFunc) (url string, closeFunc func()) {
	server := httptest.NewServer(handler)
	return server.URL, func() { server.Close() }
}
