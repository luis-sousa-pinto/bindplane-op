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

// PreRun is a function that runs before the command
type PreRun func(cmd *cobra.Command, args []string) error

// PreRunAdder is a function that adds a pre run to a command
type PreRunAdder func(cmd *cobra.Command, loader Loader) PreRun

// CombinePreruns combines preruns into a single prerun
func CombinePreruns(oldPreRun PreRun, newPreRuns []PreRun) PreRun {
	return func(cmd *cobra.Command, args []string) error {
		// run all preruns
		for _, prerun := range newPreRuns {
			if err := prerun(cmd, args); err != nil {
				return err
			}
		}

		if oldPreRun != nil {
			return oldPreRun(cmd, args)
		}

		return nil
	}
}

// AddPrerunsToExistingCmd adds preruns to an existing command in the order they are given
func AddPrerunsToExistingCmd(cmd *cobra.Command, loader Loader, prerunAdders ...PreRunAdder) *cobra.Command {
	var oldPreRun PreRun
	if cmd.PersistentPreRunE != nil {
		oldPreRun = cmd.PersistentPreRunE
	}
	preruns := []PreRun{}
	for _, adder := range prerunAdders {
		preruns = append(preruns, adder(cmd, loader))
	}

	cmd.PersistentPreRunE = CombinePreruns(oldPreRun, preruns)
	return cmd
}

// AddValidationPrerun adds the validation pre run to the command
func AddValidationPrerun(_ *cobra.Command, loader Loader) PreRun {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		return loader.ValidateConfig(ctx)
	}
}

// AddLoadConfigPrerun adds the load config pre run to the command
func AddLoadConfigPrerun(cmd *cobra.Command, loader Loader) PreRun {
	var configArg string
	var profileArg string
	// These flags are command line only
	flags := cmd.PersistentFlags()
	flags.StringVarP(&configArg, "config", "c", "", "full path to configuration file")
	flags.StringVar(&profileArg, "profile", "", "configuration profile name to use")

	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if configArg != "" {
			return loader.LoadConfig(ctx, configArg)
		}
		return loader.LoadProfile(ctx, profileArg)
	}
}
