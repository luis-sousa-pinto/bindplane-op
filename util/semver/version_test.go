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

package semver

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func v(str string) *Version {
	return Parse(str)
}

func TestVersionParsing(t *testing.T) {
	// normal versions
	require.Equal(t, Version{1, 2, 3}, *v("1.2.3"))
	require.Equal(t, Version{1, 2, 0}, *v("1.2"))
	require.Equal(t, Version{1, 0, 0}, *v("1"))

	// multiple digits
	require.Equal(t, Version{11, 12, 13}, *v("11.12.13"))
	require.Equal(t, Version{11, 12, 0}, *v("11.12"))
	require.Equal(t, Version{11, 0, 0}, *v("11"))

	// empty string and no version
	require.Equal(t, Version{0, 0, 0}, *v(""))
	require.Equal(t, Version{0, 0, 0}, *v("barfoo"))

	// embedded version because we're really lenient
	require.Equal(t, Version{0, 1, 2}, *v("bar0.1.2foo"))
	require.Equal(t, Version{0, 1, 0}, *v("bar0.1foo"))
	require.Equal(t, Version{0, 0, 0}, *v("bar0foo"))

	// strange unlikely cases
	require.Equal(t, Version{0, 0, 0}, *v("a.b.c"))
	require.Equal(t, Version{1, 0, 0}, *v("1..."))
	require.Equal(t, Version{1, 0, 0}, *v("...1"))
	require.Equal(t, Version{1, 0, 0}, *v("1...2.3"))
}

func TestVersionCompare(t *testing.T) {
	require.True(t, v("1.0.0").IsOlder(v("2.0.0")))
	require.True(t, v("1.0.0").IsOlder(v("1.1.0")))
	require.True(t, v("1.0.0").IsOlder(v("1.0.1")))

	require.False(t, v("2.0.0").IsOlder(v("1.0.0")))
	require.False(t, v("1.1.0").IsOlder(v("1.0.0")))
	require.False(t, v("1.0.1").IsOlder(v("1.0.0")))

	require.False(t, v("1.0.0").IsNewer(v("2.0.0")))
	require.False(t, v("1.0.0").IsNewer(v("1.1.0")))
	require.False(t, v("1.0.0").IsNewer(v("1.0.1")))

	require.True(t, v("2.0.0").IsNewer(v("1.0.0")))
	require.True(t, v("1.1.0").IsNewer(v("1.0.0")))
	require.True(t, v("1.0.1").IsNewer(v("1.0.0")))

	require.True(t, v("2.0.0").Equals(v("2.0.0")))
	require.False(t, v("1.0.0").Equals(v("2.0.0")))
	require.False(t, v("0.0.0").Equals(v("0.0.1")))
}

func TestVersionNew(t *testing.T) {
	require.Equal(t, v("1.2.3"), New(1, 2, 3))
}
func TestVersionString(t *testing.T) {
	require.Equal(t, "1.2.3", v("1.2.3").String())
}
