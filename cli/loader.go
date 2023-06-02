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

package cli

import (
	"context"

	"github.com/spf13/cobra"
)

// Loader is an interface for loading the BindPlane configuration
type Loader interface {
	// LoadConfig loads the configuration from the given path
	LoadConfig(ctx context.Context, path string) error

	// LoadProfile loads the profile with the given name
	LoadProfile(ctx context.Context, name string) error

	// ValidateConfig validates the configuration.
	ValidateConfig(ctx context.Context) error
}

// preRun is a function that runs before the command
type preRun func(cmd *cobra.Command, args []string) error

// addPrerunToExistingCmd adds a pre run to an existing command
func addPrerunToExistingCmd(cmd *cobra.Command, new preRun) *cobra.Command {
	if cmd.PersistentPreRunE != nil {
		// If there is already a persistent pre run, we need to wrap it
		// so that it runs after the config load pre run
		oldPreRun := cmd.PersistentPreRunE
		cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
			if err := new(cmd, args); err != nil {
				return err
			}
			return oldPreRun(cmd, args)
		}
	} else {
		cmd.PersistentPreRunE = new
	}
	return cmd
}

// AddValidationPrerun adds the validation pre run to the command
func AddValidationPrerun(cmd *cobra.Command, loader Loader) *cobra.Command {
	validatePreRun := func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		return loader.ValidateConfig(ctx)
	}

	cmd = addPrerunToExistingCmd(cmd, validatePreRun)
	cmd = AddLoadConfigPrerun(cmd, loader)
	return cmd
}

// AddLoadConfigPrerun adds the load config pre run to the command
func AddLoadConfigPrerun(cmd *cobra.Command, loader Loader) *cobra.Command {
	var configArg string
	var profileArg string

	configLoadPreRun := func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if configArg != "" {
			return loader.LoadConfig(ctx, configArg)
		}
		return loader.LoadProfile(ctx, profileArg)
	}

	cmd = addPrerunToExistingCmd(cmd, configLoadPreRun)

	// These flags are command line only
	flags := cmd.PersistentFlags()
	flags.StringVarP(&configArg, "config", "c", "", "full path to configuration file")
	flags.StringVar(&profileArg, "profile", "", "configuration profile name to use")
	return cmd
}
