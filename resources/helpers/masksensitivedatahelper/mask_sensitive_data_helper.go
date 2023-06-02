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

// Package masksensitivedatahelper contains the helper functions for the mask_sensitive_data processor.
package masksensitivedatahelper

import (
	"fmt"
	"strings"
)

const (
	ccRegex              = `\b(?:(?:(?:\d{4}[- ]?){3}\d{4}|\d{15,16}))\b`
	dobRegex             = `\b(0?[1-9]|1[0-2])\/(0?[1-9]|[12]\d|3[01])\/(?:\d{2})?\d{2}\b`
	emailRegex           = `\b[a-zA-Z0-9._\/\+\-—|]+@[A-Za-z0-9.\-—|]+\.?[a-zA-Z|]{0,6}\b`
	ibanRegex            = `\b[A-Z]{2}\d{2}[A-Z\d]{1,30}\b`
	ipv4Regex            = `\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`
	ipv6Regex            = `\b(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}\b`
	macAddressRegex      = `\b([0-9A-Fa-f]{2}[:-]){5}[0-9A-Fa-f]{2}\b`
	phoneRegex           = `\b((\+|\b)[1l][\-\. ])?\(?\b[\dOlZSB]{3,5}([\-\. ]|\) ?)[\dOlZSB]{3}[\-\. ][\dOlZSB]{4}\b`
	ssnRegex             = `\b\d{3}[- ]\d{2}[- ]\d{4}\b`
	usCityStateRegex     = `\b[A-Z][A-Za-z\s\.]+,\s{0,1}[A-Z]{2}\b`
	usStreetAddressRegex = `\b\d+\s[A-z]+\s[A-z]+(\s[A-z]+)?\s*\d*\b`
	usZipcodeRegex       = `\b\d{5}(?:[-\s]\d{4})?\b`
	uuidGUIDCodeRegex    = `\b[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}\b`
)

const (
	prefixResource   = "resource"
	prefixAttributes = "attributes"
	prefixBody       = "body"
)

// BPRenderMaskRules renders the mask rules
func BPRenderMaskRules(defaults []any, customRules any, excludeResourceKeys, excludeAttributeKeys, excludeBodyKeys []any) string {
	stri := strings.Builder{}
	stri.WriteString("- mask:\n    rules:\n")

	for _, rule := range defaults {
		switch rule {

		case "Credit Card":
			stri.WriteString(fmt.Sprintf("      card: %s\n", ccRegex))
		case "Date of Birth":
			stri.WriteString(fmt.Sprintf("      dob: %s\n", dobRegex))
		case "Email":
			stri.WriteString(fmt.Sprintf("      email: %s\n", emailRegex))
		case "International Bank Account Number (IBAN)":
			stri.WriteString(fmt.Sprintf("      iban: %s\n", ibanRegex))
		case "IPv4 Address":
			stri.WriteString(fmt.Sprintf("      ipv4: %s\n", ipv4Regex))
		case "IPv6 Address":
			stri.WriteString(fmt.Sprintf("      ipv6: %s\n", ipv6Regex))
		case "MAC Address":
			stri.WriteString(fmt.Sprintf("      mac_address: %s\n", macAddressRegex))
		case "Phone Number":
			stri.WriteString(fmt.Sprintf("      phone_number: %s\n", phoneRegex))
		case "Social Security Number (SSN)":
			stri.WriteString(fmt.Sprintf("      ssn: %s\n", ssnRegex))
		case "US City, State":
			stri.WriteString(fmt.Sprintf("      us_city_state: %s\n", usCityStateRegex))
		case "US Street Address":
			stri.WriteString(fmt.Sprintf("      us_street_address: %s\n", usStreetAddressRegex))
		case "US Zipcode":
			stri.WriteString(fmt.Sprintf("      us_zip_code: %s\n", usZipcodeRegex))
		case "UUID/GUID":
			stri.WriteString(fmt.Sprintf("      uuid_guid: %s\n", uuidGUIDCodeRegex))
		}
	}

	// We expect this to be a map[string]any, but in the test for
	// TestProcessorTypes in resource_test.go it is a map[any]any.
	switch customRules.(type) {
	case map[string]any:
		for key, value := range customRules.(map[string]any) {
			stri.WriteString(fmt.Sprintf("      %s: %v\n", key, value))
		}
	case map[any]any:
		for key, value := range customRules.(map[any]any) {
			stri.WriteString(fmt.Sprintf("      %v: %v\n", key, value))
		}
	}

	if (len(excludeResourceKeys) + len(excludeAttributeKeys) + len(excludeBodyKeys)) == 0 {
		return stri.String()
	}

	stri.WriteString("    exclude:\n")
	for _, field := range excludeResourceKeys {
		stri.WriteString(fmt.Sprintf("    - %s.%s\n", prefixResource, field))
	}
	for _, field := range excludeAttributeKeys {
		stri.WriteString(fmt.Sprintf("    - %s.%s\n", prefixAttributes, field))
	}
	for _, field := range excludeBodyKeys {
		stri.WriteString(fmt.Sprintf("    - %s.%s\n", prefixBody, field))
	}
	return stri.String()
}
