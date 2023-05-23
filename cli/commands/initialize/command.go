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

// Package initialize provides the initialize command, which initializes an installation.
package initialize

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Command returns the BindPlane initialize cobra command
func Command(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "init",
		Aliases: []string{"initialize"},
		Short:   "Initialize an installation",
	}

	cmd.AddCommand(
		ClientCommand(builder),
	)

	cmd.AddCommand(
		ServerCommand(builder),
	)

	return cmd
}

// ServerCommand provides the implementation for "bindplane init server"
func ServerCommand(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "server",
		Aliases: []string{"serve"},
		Short:   "Initializes a new server installation",
		Long:    ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			intializer, err := builder.BuildInitializer(ctx)
			if err != nil {
				return err
			}

			if err := intializer.InitializeServer(ctx); err != nil {
				return err
			}

			fmt.Println("")
			fmt.Println("Initialization complete!")
			fmt.Println("Restart the BindPlane server to reload the configuration.")
			return nil
		},
	}
	return cmd
}

// ClientCommand provides the implementation for "bindplane init client"
func ClientCommand(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "client",
		Aliases: []string{"cli"},
		Short:   "Initializes a new client installation",
		Long:    ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			intializer, err := builder.BuildInitializer(ctx)
			if err != nil {
				return err
			}

			if err := intializer.InitializeClient(cmd.Context()); err != nil {
				return err
			}

			fmt.Println("")
			fmt.Println("Initialization complete!")
			fmt.Println("Run \"bindplane version\" to test the login.")
			return nil
		},
	}

	return cmd
}
