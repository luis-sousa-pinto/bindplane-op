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

package masksensitivedatahelper

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBPRenderMaskRules(t *testing.T) {
	type args struct {
		defaults             []any
		customRules          map[string]any
		excludeResourceKeys  []any
		excludeAttributeKeys []any
		excludeBodyKeys      []any
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"defaults",
			args{
				[]any{
					"Credit Card",
					"Date of Birth",
					"Email",
					"International Bank Account Number (IBAN)",
					"IPv4 Address",
					"IPv6 Address",
					"MAC Address",
					"Phone Number",
					"Social Security Number (SSN)",
					"US City, State",
					"US Street Address",
					"US Zipcode",
					"UUID/GUID",
				},
				map[string]any{},
				[]any{},
				[]any{},
				[]any{},
			},
			`- mask:
    rules:
      card: \b(?:(?:(?:\d{4}[- ]?){3}\d{4}|\d{15,16}))\b
      dob: \b(0?[1-9]|1[0-2])\/(0?[1-9]|[12]\d|3[01])\/(?:\d{2})?\d{2}\b
      email: \b[a-zA-Z0-9._\/\+\-—|]+@[A-Za-z0-9.\-—|]+\.?[a-zA-Z|]{0,6}\b
      iban: \b[A-Z]{2}\d{2}[A-Z\d]{1,30}\b
      ipv4: \b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b
      ipv6: \b(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}\b
      mac_address: \b([0-9A-Fa-f]{2}[:-]){5}[0-9A-Fa-f]{2}\b
      phone_number: \b((\+|\b)[1l][\-\. ])?\(?\b[\dOlZSB]{3,5}([\-\. ]|\) ?)[\dOlZSB]{3}[\-\. ][\dOlZSB]{4}\b
      ssn: \b\d{3}[- ]\d{2}[- ]\d{4}\b
      us_city_state: \b[A-Z][A-Za-z\s\.]+,\s{0,1}[A-Z]{2}\b
      us_street_address: \b\d+\s[A-z]+\s[A-z]+(\s[A-z]+)?\s*\d*\b
      us_zip_code: \b\d{5}(?:[-\s]\d{4})?\b
      uuid_guid: \b[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}\b
`,
		},
		{
			"customRules",
			args{
				[]any{},
				map[string]any{
					"custom": "custom",
				},
				[]any{},
				[]any{},
				[]any{},
			},
			`- mask:
    rules:
      custom: custom
`,
		},
		{
			"exclude",
			args{
				[]any{
					"Credit Card",
					"Date of Birth",
					"Email",
					"International Bank Account Number (IBAN)",
				},
				map[string]any{},
				[]any{
					"resource-key",
				},
				[]any{
					"attribute-key",
				},
				[]any{
					"body-key",
				},
			},
			`- mask:
    rules:
      card: \b(?:(?:(?:\d{4}[- ]?){3}\d{4}|\d{15,16}))\b
      dob: \b(0?[1-9]|1[0-2])\/(0?[1-9]|[12]\d|3[01])\/(?:\d{2})?\d{2}\b
      email: \b[a-zA-Z0-9._\/\+\-—|]+@[A-Za-z0-9.\-—|]+\.?[a-zA-Z|]{0,6}\b
      iban: \b[A-Z]{2}\d{2}[A-Z\d]{1,30}\b
    exclude:
    - resource.resource-key
    - attributes.attribute-key
    - body.body-key
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BPRenderMaskRules(
				tt.args.defaults,
				tt.args.customRules,
				tt.args.excludeResourceKeys,
				tt.args.excludeAttributeKeys,
				tt.args.excludeBodyKeys,
			)

			require.Equal(t, tt.want, got)
		})
	}
}
