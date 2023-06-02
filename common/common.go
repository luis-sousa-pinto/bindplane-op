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

// Package common contains common functions and types used by BindPlane.
package common

import (
	"os"
	"path/filepath"
)

const (
	// ProfilesFolderName is the name of the profiles folder.
	ProfilesFolderName = "profiles"

	// DefaultHomeName is the default directory name for bindplane.
	DefaultHomeName = ".bindplane"

	// DefaultProfileName is the name of the default profile
	DefaultProfileName = "default"

	// HomeEnv is the environment variable for the bindplane home path.
	HomeEnv = "BINDPLANE_CONFIG_HOME"
)

// GetHome returns the bindplane home path.
func GetHome() string {
	if value, ok := os.LookupEnv(HomeEnv); ok {
		return value
	}

	// Look up the users home dir
	userHome, _ := os.UserHomeDir() //#nosec G104 -- We're ok if this fails
	return filepath.Join(userHome, DefaultHomeName)
}

// GetProfilesFolder returns the profiles folder path.
func GetProfilesFolder() string {
	return filepath.Join(GetHome(), ProfilesFolderName)
}
