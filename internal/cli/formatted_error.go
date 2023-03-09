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

package cli

import (
	"fmt"
	"strings"
)

// formattedError is an error that formats its Error() string in a user friendly way
type formattedError struct {
	errs []error
}

// Error returns the formatted error string
func (f *formattedError) Error() string {
	if f == nil || len(f.errs) == 0 {
		return ""
	}

	if len(f.errs) == 1 {
		return fmt.Sprintf("1 error occurred:\n\t* %s\n\n", f.errs[0])
	}

	points := make([]string, len(f.errs))
	for i, err := range f.errs {
		points[i] = fmt.Sprintf("* %s", err)
	}

	return fmt.Sprintf(
		"%d errors occurred:\n\t%s\n\n",
		len(f.errs), strings.Join(points, "\n\t"))
}

// FormatError takes in an error and formats it in a user friendly way
func FormatError(err error) error {
	if err == nil {
		return nil
	}

	// Check if this is a errors.joinedError
	u, ok := err.(interface {
		Unwrap() []error
	})

	// Not a errors.joinedError so return the single wrapped errors
	if !ok {
		return &formattedError{
			errs: []error{err},
		}
	}

	return &formattedError{
		errs: u.Unwrap(),
	}
}
