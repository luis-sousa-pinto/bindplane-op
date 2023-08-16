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

package agent

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNoopClient(t *testing.T) {
	t.Run("Version errors", func(t *testing.T) {
		c := newNoopClient()
		_, err := c.Version("v1.30.0")
		require.ErrorIs(t, err, ErrVersionNotFound)
		require.ErrorContains(t, err, "noop client:")
	})

	t.Run("Latest errors", func(t *testing.T) {
		c := newNoopClient()
		_, err := c.LatestVersion()
		require.ErrorIs(t, err, ErrVersionNotFound)
		require.ErrorContains(t, err, "noop client:")
	})

	t.Run("Versions returns an empty list", func(t *testing.T) {
		c := newNoopClient()
		v, err := c.Versions()
		require.NoError(t, err)
		require.Len(t, v, 0)
	})
}
