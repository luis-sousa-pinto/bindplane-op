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

package common

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	cases := []struct {
		name        string
		config      Config
		expectedErr error
	}{
		{
			"empty",
			Config{},
			nil,
		},
		{
			"valid-directory",
			Config{
				Server: Server{
					Common: Common{
						bindplaneHomePath: "./",
					},
				},
				Client: Client{
					Common: Common{
						bindplaneHomePath: "./",
					},
				},
			},
			nil,
		},
		{
			"valid-port",
			Config{
				Server: Server{
					Common: Common{
						Port: "5000",
					},
				},
				Client: Client{
					Common: Common{
						Port: "1000",
					},
				},
			},
			nil,
		},
		{
			"valid-server-address",
			Config{
				Server: Server{
					Common: Common{
						ServerURL: "http://localhost:3000",
					},
				},
				Client: Client{
					Common: Common{
						ServerURL: "http://localhost:3000",
					},
				},
			},
			nil,
		},
		{
			"valid-server-address-tls",
			Config{
				Server: Server{
					Common: Common{
						ServerURL: "https://localhost:3000",
					},
				},
				Client: Client{
					Common: Common{
						ServerURL: "https://localhost:9999",
					},
				},
			},
			nil,
		},
		{
			"valid-secret-key-uuid-v1",
			Config{
				Server: Server{
					SecretKey: "5696de96-95ab-11ec-b909-0242ac120002",
				},
			},
			nil,
		},
		{
			"valid-secret-key-uuid-v1",
			Config{
				Server: Server{
					SecretKey: "603ecef5-32ef-4e78-9e84-beef8a96cdb9",
				},
			},
			nil,
		},
		{
			"valid-remote-url",
			Config{
				Server: Server{
					RemoteURL: "ws://github.com:5555",
				},
			},
			nil,
		},
		{
			"valid-remote-url-tls",
			Config{
				Server: Server{
					RemoteURL: "wss://github.com:5555",
				},
			},
			nil,
		},
		{
			"valid-tls",
			Config{
				Server: Server{
					Common: Common{
						TLSConfig: TLSConfig{
							Certificate: "./testdata/tls/server.crt.test",
							PrivateKey:  "./testdata/tls/server.key.test",
						},
					},
				},
				Client: Client{
					Common: Common{
						TLSConfig: TLSConfig{
							Certificate: "./testdata/tls/client.crt.test",
							PrivateKey:  "./testdata/tls/client.key.test",
						},
					},
				},
			},
			nil,
		},
		{
			"valid-tls-mtls",
			Config{
				Server: Server{
					Common: Common{
						TLSConfig: TLSConfig{
							Certificate:          "./testdata/tls/server.crt.test",
							PrivateKey:           "./testdata/tls/server.key.test",
							CertificateAuthority: []string{"./testdata/tls/ca.crt.test"},
						},
					},
				},
				Client: Client{
					Common: Common{
						TLSConfig: TLSConfig{
							Certificate:          "./testdata/tls/client.crt.test",
							PrivateKey:           "./testdata/tls/client.key.test",
							CertificateAuthority: []string{"./testdata/tls/ca.crt.test"},
						},
					},
				},
			},
			nil,
		},
		{
			"valid-storage-file-path",
			Config{
				Server: Server{
					StorageFilePath: "./testdata",
				},
				Client: Client{},
			},
			nil,
		},
		{
			"invalid-directory",
			Config{
				Server: Server{
					Common: Common{
						bindplaneHomePath: "/badpath/root",
					},
				},
				Client: Client{
					Common: Common{
						bindplaneHomePath: "./badrel/path",
					},
				},
			},
			errors.Join(
				errors.New("failed to lookup directory /badpath/root: stat /badpath/root: no such file or directory"),
				errors.New("failed to lookup directory ./badrel/path: stat ./badrel/path: no such file or directory"),
			),
		},
		{
			"invalid-port",
			Config{
				Server: Server{
					Common: Common{
						Port: "ten",
					},
				},
			},
			errors.New("failed to convert port ten to an"),
		},
		{
			"invalid-server-port",
			Config{
				Server: Server{
					Common: Common{
						Port: "100000",
					},
				},
			},
			errors.New("port must be between"),
		},
		{
			"invalid-client-port",
			Config{
				Server: Server{
					Common: Common{
						Port: "5000",
					},
				},
				Client: Client{
					Common: Common{
						Port: "0",
					},
				},
			},
			errors.New("port must be between"),
		},
		{
			"invalid-server-address",
			Config{
				Server: Server{
					Common: Common{
						ServerURL: "localhost:3000",
					},
				},
				Client: Client{
					Common: Common{
						ServerURL: "ws://localhost:3000",
					},
				},
			},
			errors.New("failed to validate server address localhost:3000: scheme localhost is invalid: valid schemes are [http https]"),
		},
		{
			"invalid-server-address-malformed-url",
			Config{
				Server: Server{
					Common: Common{
						ServerURL: "6:3000",
					},
				},
				Client: Client{
					Common: Common{
						ServerURL: "4:3000",
					},
				},
			},
			errors.New("failed to validate server address 6:3000: failed to parse url 6:3000: parse \"6:3000\": first path segment in URL cannot contain"),
		},
		{
			"invalid-secret-key-uuid",
			Config{
				Server: Server{
					SecretKey: "603ecef5",
				},
			},
			errors.New("failed to validate secret key: invalid UUID"),
		},
		{
			"invalid-remote-url",
			Config{
				Server: Server{
					RemoteURL: "github.com:5555",
				},
			},
			errors.New("failed to validate remote address github.com:5555: scheme github.com is invalid: valid schemes are [ws wss]"),
		},
		{
			"invalid-remote-url-scheme",
			Config{
				Server: Server{
					RemoteURL: "http://github.com:5555",
				},
			},
			errors.New("scheme http is invalid: valid schemes are [ws wss]"),
		},
		{
			"missing-scheme",
			Config{
				Server: Server{
					RemoteURL: "github.com",
				},
			},
			errors.New("scheme is not set"),
		},
		{
			"invalid-remote-url-malformed-url",
			Config{
				Server: Server{
					RemoteURL: "5:github.com",
				},
			},
			errors.New("first path segment in URL cannot contain colon"),
		},
		{
			"invalid-valid-tls-missing-private-key",
			Config{
				Server: Server{
					Common: Common{
						TLSConfig: TLSConfig{
							Certificate: "./testdata/tls/server.crt.test",
						},
					},
				},
				Client: Client{
					Common: Common{
						TLSConfig: TLSConfig{
							Certificate: "./testdata/tls/client.crt.test",
						},
					},
				},
			},
			errors.New("private key must be set when tls certificate is set"),
		},
		{
			"invalid-valid-tls-missing-certificate-key",
			Config{
				Server: Server{
					Common: Common{
						TLSConfig: TLSConfig{
							PrivateKey: "./testdata/tls/server.key.test",
						},
					},
				},
				Client: Client{
					Common: Common{
						TLSConfig: TLSConfig{
							PrivateKey: "./testdata/tls/client.key.test",
						},
					},
				},
			},
			errors.New("tls certificate must be set when tls private key is set"),
		},
		{
			"invalid-tls-mtls-missing-keypair",
			Config{
				Server: Server{
					Common: Common{
						TLSConfig: TLSConfig{
							CertificateAuthority: []string{"./testdata/tls/ca.crt.test"},
						},
					},
				},
				Client: Client{
					Common: Common{
						TLSConfig: TLSConfig{
							CertificateAuthority: []string{"./testdata/tls/ca.crt.test"},
						},
					},
				},
			},
			errors.New("certificate and private key must be set when tls certificate authority is set"),
		},
		{
			"invalid-tls-mtls-missing-cert-file",
			Config{
				Server: Server{
					Common: Common{
						TLSConfig: TLSConfig{
							Certificate: "/bad/path/testdata/tls/server.crt.test",
							PrivateKey:  "./testdata/tls/server.key.test",
						},
					},
				},
				Client: Client{
					Common: Common{
						TLSConfig: TLSConfig{
							Certificate: "./testdata/tls/client.crt.test",
							PrivateKey:  "./testdata/tls/client.key.test",
						},
					},
				},
			},
			errors.New("failed to lookup tls certificate file"),
		},
		{
			"invalid-tls-mtls-missing-key-file",
			Config{
				Server: Server{
					Common: Common{
						TLSConfig: TLSConfig{
							Certificate: "./testdata/tls/client.crt.test",
							PrivateKey:  "./testdata/tls/server.key.test",
						},
					},
				},
				Client: Client{
					Common: Common{
						TLSConfig: TLSConfig{
							Certificate: "./testdata/tls/client.crt.test",
							PrivateKey:  "/bad/path/testdata/tls/client.key.test",
						},
					},
				},
			},
			errors.New("failed to lookup tls private key file"),
		},
		{
			"invalid-tls-mtls-missing-ca-file",
			Config{
				Server: Server{
					Common: Common{
						TLSConfig: TLSConfig{
							Certificate: "./testdata/tls/server.crt.test",
							PrivateKey:  "./testdata/tls/server.key.test",
							CertificateAuthority: []string{
								"./testdata/tls/ca.crt.test",
								"/bad/ca/path",
							},
						},
					},
				},
				Client: Client{
					Common: Common{
						TLSConfig: TLSConfig{
							Certificate:          "./testdata/tls/client.crt.test",
							PrivateKey:           "./testdata/tls/client.key.test",
							CertificateAuthority: []string{"./testdata/tls/ca.crt.test"},
						},
					},
				},
			},
			errors.New("failed to lookup tls certificate authority file"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
			if tc.expectedErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.expectedErr.Error())
			}
		})
	}
}
