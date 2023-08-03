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

// Package main is the entrypoint for the bindplane command line interface.
// It determines the home directory, and adds all commands to the root command.
package main

import (
	"fmt"
	"os"

	jsoniter "github.com/json-iterator/go"
	"github.com/observiq/bindplane-op/cli"
	"github.com/observiq/bindplane-op/cli/commands/apply"
	"github.com/observiq/bindplane-op/cli/commands/delete"
	"github.com/observiq/bindplane-op/cli/commands/get"
	"github.com/observiq/bindplane-op/cli/commands/initialize"
	"github.com/observiq/bindplane-op/cli/commands/install"
	"github.com/observiq/bindplane-op/cli/commands/label"
	"github.com/observiq/bindplane-op/cli/commands/profile"
	"github.com/observiq/bindplane-op/cli/commands/rollout"
	"github.com/observiq/bindplane-op/cli/commands/root"
	"github.com/observiq/bindplane-op/cli/commands/serve"
	"github.com/observiq/bindplane-op/cli/commands/sync"
	"github.com/observiq/bindplane-op/cli/commands/update"
	"github.com/observiq/bindplane-op/cli/commands/version"
	"github.com/observiq/bindplane-op/routes"
	"github.com/spf13/cobra"
)

func init() {
	// jsoniter should use a format compatible with the standard library. This is important because there are places where
	// we expect consistent rendering of the JSON output to generate hashes or to compare resources before and after
	// changes and this is only possible with consistent marshalling.
	jsoniter.ConfigDefault = jsoniter.ConfigCompatibleWithStandardLibrary
}

func main() {
	routeBuilder := &routes.CombinedRouteBuilder{}
	factory := cli.NewFactory(routeBuilder)
	rootCmd, err := root.Command()
	if err != nil {
		fmt.Printf("Failed to create root command: %s", err.Error())
		os.Exit(1)
	}

	rootCmd.AddCommand(
		cli.AddPrerunsToExistingCmd(apply.Command(factory), factory, cli.AddLoadConfigPrerun, cli.AddValidationPrerun),
		cli.AddPrerunsToExistingCmd(get.Command(factory), factory, cli.AddLoadConfigPrerun, cli.AddValidationPrerun),
		cli.AddPrerunsToExistingCmd(label.Command(factory), factory, cli.AddLoadConfigPrerun, cli.AddValidationPrerun),
		cli.AddPrerunsToExistingCmd(delete.Command(factory), factory, cli.AddLoadConfigPrerun, cli.AddValidationPrerun),
		cli.AddPrerunsToExistingCmd(profile.Command(factory), factory, cli.AddLoadConfigPrerun),
		cli.AddPrerunsToExistingCmd(version.Command(factory), factory, cli.AddLoadConfigPrerun),
		cli.AddPrerunsToExistingCmd(initialize.Command(factory), factory, cli.AddLoadConfigPrerun),
		cli.AddPrerunsToExistingCmd(install.Command(factory), factory, cli.AddLoadConfigPrerun, cli.AddValidationPrerun),
		cli.AddPrerunsToExistingCmd(sync.Command(factory), factory, cli.AddLoadConfigPrerun, cli.AddValidationPrerun),
		cli.AddPrerunsToExistingCmd(update.Command(factory), factory, cli.AddLoadConfigPrerun, cli.AddValidationPrerun),
		cli.AddPrerunsToExistingCmd(rollout.Command(factory), factory, cli.AddLoadConfigPrerun, cli.AddValidationPrerun),
		cli.AddPrerunsToExistingCmd(serve.Command(factory), factory, cli.AddLoadConfigPrerun, cli.AddValidationPrerun))

	cobra.CheckErr(rootCmd.Execute())
}
