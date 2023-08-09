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

package config

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNetworkValidate(t *testing.T) {
	testCases := []struct {
		name     string
		network  Network
		expected error
	}{
		{
			name: "valid",
			network: Network{
				Port: "1234",
			},
		},
		{
			name: "invalid remote url",
			network: Network{
				Port:      "1234",
				RemoteURL: "invalid",
			},
			expected: errors.New("invalid remote url: scheme '' is invalid"),
		},
		{
			name: "port not an int",
			network: Network{
				Port: "test",
			},
			expected: errors.New("port must be an integer"),
		},
		{
			name: "port not in range",
			network: Network{
				Port: "65536",
			},
			expected: errors.New("port must be between 1 and 65535"),
		},
		{
			name: "invalid-valid-tls-missing-private-key",
			network: Network{
				Port: "1234",
				TLS: TLS{
					Certificate: "./testdata/tls/server.crt.test",
				},
			},
			expected: errors.New("private key must be set when tls certificate is set"),
		},
		{
			name: "invalid-valid-tls-missing-certificate-key",
			network: Network{
				Port: "1234",
				TLS: TLS{
					PrivateKey: "./testdata/tls/server.key.test",
				},
			},
			expected: errors.New("tls certificate must be set when tls private key is set"),
		},
		{
			name: "invalid-tls-mtls-missing-cert-file",
			network: Network{
				Port: "1234",
				TLS: TLS{
					Certificate: "/bad/path/testdata/tls/server.crt.test",
					PrivateKey:  "./testdata/tls/server.key.test",
				},
			},
			expected: errors.New("failed to lookup tls certificate file"),
		},
		{
			name: "invalid-tls-mtls-missing-ca-file",
			network: Network{
				Port: "1234",
				TLS: TLS{
					Certificate: "./testdata/tls/server.crt.test",
					PrivateKey:  "./testdata/tls/server.key.test",
					CertificateAuthority: []string{
						"./testdata/tls/ca.crt.test",
						"/bad/ca/path",
					},
				},
			},
			expected: errors.New("failed to lookup tls certificate authority file"),
		},
		{
			name: "invalid-tls-mtls-missing-key-file",
			network: Network{
				Port: "1234",
				TLS: TLS{
					Certificate: "./testdata/tls/server.crt.test",
					PrivateKey:  "/bad/path/testdata/tls/server.key.missing",
				},
			},
			expected: errors.New("failed to lookup tls private key file"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.network.Validate()
			switch tc.expected {
			case nil:
				require.NoError(t, err)
			default:
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expected.Error())
			}
		})
	}
}
