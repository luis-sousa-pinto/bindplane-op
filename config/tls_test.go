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
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTLSValidate(t *testing.T) {
	testCases := []struct {
		name     string
		tls      TLS
		expected error
	}{
		{
			name: "invalid-valid-tls-missing-private-key",
			tls: TLS{
				Certificate: "./testdata/tls/server.crt.test",
			},
			expected: errors.New("private key must be set when tls certificate is set"),
		},
		{
			name: "invalid-valid-tls-missing-certificate-key",
			tls: TLS{
				PrivateKey: "./testdata/tls/server.key.test",
			},
			expected: errors.New("tls certificate must be set when tls private key is set"),
		},
		{
			name: "invalid-tls-mtls-missing-cert-file",
			tls: TLS{
				Certificate: "/bad/path/testdata/tls/server.crt.test",
				PrivateKey:  "./testdata/tls/server.key.test",
			},
			expected: errors.New("failed to lookup tls certificate file"),
		},
		{
			name: "invalid-tls-mtls-missing-ca-file",
			tls: TLS{
				Certificate: "./testdata/tls/server.crt.test",
				PrivateKey:  "./testdata/tls/server.key.test",
				CertificateAuthority: []string{
					"./testdata/tls/ca.crt.test",
					"/bad/ca/path",
				},
			},
			expected: errors.New("failed to lookup tls certificate authority file"),
		},
		{
			name: "invalid-tls-mtls-missing-key-file",
			tls: TLS{
				Certificate: "./testdata/tls/server.crt.test",
				PrivateKey:  "/bad/path/testdata/tls/server.key.missing",
			},
			expected: errors.New("failed to lookup tls private key file"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.tls.Validate()
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

func TestTLSEnabled(t *testing.T) {
	testCases := []struct {
		desc     string
		tls      TLS
		expected bool
	}{
		{
			desc: "Enabled",
			tls: TLS{
				Certificate: "certificate",
				PrivateKey:  "private",
			},
			expected: true,
		},
		{
			desc:     "Disabled",
			tls:      TLS{},
			expected: false,
		},
	}

	for _, tc := range testCases {
		result := tc.tls.TLSEnabled()
		require.Equal(t, tc.expected, result)
	}
}

func TestTLSConvert(t *testing.T) {
	cases := []struct {
		name      string
		tls       TLS
		expect    *tls.Config
		expectErr bool
	}{
		{
			name: "tls",
			tls: TLS{
				Certificate:          "./testdata/tls/client.crt.test",
				PrivateKey:           "./testdata/tls/client.key.test",
				CertificateAuthority: []string{},
			},
			expect: &tls.Config{
				Certificates: func() []tls.Certificate {
					pair, err := tls.LoadX509KeyPair(
						"./testdata/tls/client.crt.test",
						"./testdata/tls/client.key.test",
					)
					if err != nil {
						t.Errorf("setup failed: %v", err)
						t.FailNow()
					}
					return []tls.Certificate{pair}
				}(),
			},
			expectErr: false,
		},
		{
			name: "mutual-tls",
			tls: TLS{
				Certificate: "./testdata/tls/client.crt.test",
				PrivateKey:  "./testdata/tls/client.key.test",
				CertificateAuthority: []string{
					"./testdata/tls/ca.crt.test",
				},
			},
			expect: &tls.Config{
				Certificates: func() []tls.Certificate {
					t, _ := tls.LoadX509KeyPair(
						"./testdata/tls/client.crt.test",
						"./testdata/tls/client.key.test",
					)
					return []tls.Certificate{t}
				}(),
				RootCAs: func() *x509.CertPool {
					path := "./testdata/tls/ca.crt.test"
					ca, err := ioutil.ReadFile(path)
					if err != nil {
						t.Errorf("setup failed: %v", err)
						t.FailNow()
					}
					var pool = x509.NewCertPool()
					pool.AppendCertsFromPEM(ca)
					return pool
				}(),
			},
			expectErr: false,
		},
		{
			name: "mutual-tls-invalid-ca",
			tls: TLS{
				Certificate: "./testdata/tls/client.crt.test",
				PrivateKey:  "./testdata/tls/client.key.test",
				CertificateAuthority: []string{
					// tls.go will never be a valid x509 pem file
					"tls.go",
				},
			},
			expect:    nil,
			expectErr: true,
		},
		{
			name: "tls-invalid-cert-path",
			tls: TLS{
				Certificate: "./testdata/tls/client.crt.test.invalid",
				PrivateKey:  "./testdata/tls/client.key.test",
				CertificateAuthority: []string{
					"./testdata/tls/ca.crt.test",
				},
			},
			expect:    nil,
			expectErr: true,
		},
		{
			name: "tls-invalid-key-path",
			tls: TLS{
				Certificate: "./testdata/tls/client.crt.test",
				PrivateKey:  "./testdata/tls/client.key.test.invalid",
				CertificateAuthority: []string{
					"./testdata/tls/ca.crt.test",
				},
			},
			expect:    nil,
			expectErr: true,
		},
		{
			name: "tls-invalid-ca-path",
			tls: TLS{
				Certificate: "./testdata/tls/client.crt.test",
				PrivateKey:  "./testdata/tls/client.key.test",
				CertificateAuthority: []string{
					"./testdata/tls/ca.crt.test.invalid",
				},
			},
			expect:    nil,
			expectErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := tc.tls.Convert()
			if tc.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, out)
			require.Equal(t, tc.expect.Certificates, out.Certificates)
			if len(tc.tls.CertificateAuthority) > 0 {
				require.NotNil(t, out.RootCAs)
			}
			require.Equal(t, tls.VersionTLS13, int(out.MinVersion))
		})
	}
}
