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

	"github.com/observiq/bindplane-op/cli"
	"github.com/observiq/bindplane-op/cli/commands/root"
	"github.com/observiq/bindplane-op/routes"
	"github.com/spf13/cobra"
)

func main() {
	routeBuilder := &routes.CombinedRouteBuilder{}
	factory := cli.NewFactory(routeBuilder)
	rootCmd, err := root.Command()
	if err != nil {
		fmt.Printf("Failed to create root command: %s", err.Error())
		os.Exit(1)
	}

	// Server contains all commands
	rootCmd.AddCommand(
		cli.SharedCommands(factory)...,
	)

	cobra.CheckErr(rootCmd.Execute())
}
