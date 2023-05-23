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

package common

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

/** This file contains validation logic that is shared across various config types**/

const (
	minPort = 1
	maxPort = 65535
)

// list of valid schemes for different protocols
var (
	// ValidHTTPSchemes are the valid http schemes to be used with ValidateURL function
	ValidHTTPSchemes = []string{"http", "https"}

	// ValidWSSchemes are the valid websocket schemes to be used with ValidateURL function
	ValidWSSchemes = []string{"ws", "wss"}
)

// ValidatePort validates the port
func ValidatePort(port string) error {
	p, err := strconv.Atoi(port)
	if err != nil {
		return errors.New("port must be an integer")
	}

	if p < minPort || p > maxPort {
		return fmt.Errorf("port must be between %d and %d", minPort, maxPort)
	}

	return nil
}

// ValidateURL validates the URL and ensures the scheme matches one of the valid schemes in the list
func ValidateURL(urlString string, validSchemes []string) error {
	u, err := url.Parse(urlString)
	if err != nil {
		return fmt.Errorf("failed to parse url %s: %w", urlString, err)
	}

	// If no schemes specified then return early
	if len(validSchemes) == 0 {
		return nil
	}

	for _, scheme := range validSchemes {
		// Return early if we found the scheme
		if u.Scheme == scheme {
			return nil
		}
	}

	return fmt.Errorf("scheme '%s' is invalid: valid schemes are %v", u.Scheme, validSchemes)
}
