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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetHome(t *testing.T) {
	userHome, err := os.UserHomeDir()
	require.NoError(t, err)

	home := GetHome()
	expectedPath := filepath.Join(userHome, ".bindplane")
	require.Equal(t, expectedPath, home)
}

func TestGetProfilesFolder(t *testing.T) {
	userHome, err := os.UserHomeDir()
	require.NoError(t, err)

	home := GetProfilesFolder()
	expectedPath := filepath.Join(userHome, ".bindplane/profiles")
	require.Equal(t, expectedPath, home)
}
