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

// Package root provides the root command
package root

import (
	"fmt"

	"github.com/observiq/bindplane-op/config"
	"github.com/spf13/cobra"
)

// Command is the root command that represents the base command, in this function we add persistent flags,
// and bind them to viper.
// The persistent pre run function here is where we read the profile file and set the
// values for bindplane.Config
func Command() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "bindplane",
		Short: "Next generation agent management platform",
		// cobra.CheckErr will print the returned error with exit status,
		// so we disable errors on this and child commands so error message isn't repeated
		SilenceErrors: true,
		// This will prevent child commands from printing the help message on error.
		SilenceUsage: true,
	}

	flags := cmd.PersistentFlags()
	// Set flag and env overrides for config fields
	for _, override := range config.DefaultOverrides() {
		if err := override.Bind(flags); err != nil {
			return nil, fmt.Errorf("failed to bind override %s: %w", override.Field, err)
		}
	}
	return cmd, nil
}
