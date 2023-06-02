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

package agent

import "strings"

// Sha256sums is a map of file name to sha256sum
type Sha256sums map[string]string

// ParseSha256Sums is an incomplete parser of a .sha256 or -SHA256SUMS file. Technically there are separate text and
// binary modes, but we only use the text mode. See
// https://www.gnu.org/software/coreutils/manual/html_node/md5sum-invocation.html for more details on the file format
func ParseSha256Sums(contents []byte) Sha256sums {
	result := Sha256sums{}
	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		parts := strings.Split(line, " ")
		if len(parts) == 3 {
			result[parts[2]] = parts[0]
		}
	}
	return result
}

// Sha256Sum returns the sha256sum for a given file
func (s Sha256sums) Sha256Sum(file string) string {
	return s[file]
}
