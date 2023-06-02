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

// Package profile provides commands for managing BindPlane profile configurations.
package profile

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// When --current is set, just return the currently specified profile
var currentFlag bool

// Command returns the BindPlane profile cobra command.
func Command(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Profile commands.",
		Long:  "Profile commands for managing BindPlane application configuration",
	}

	cmd.AddCommand(
		GetCommand(builder),
		SetCommand(builder),
		CreateCommand(builder),
		DeleteCommand(builder),
		ListCommand(builder),
		UseCommand(builder),
		CurrentCommand(builder),
	)

	return cmd
}

// GetCommand returns the BindPlane profile get cobra command
func GetCommand(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <name>",
		Short: "Get details on a saved profile.",
		RunE: func(cmd *cobra.Command, args []string) error {
			var name string
			ctx := cmd.Context()
			writer := cmd.OutOrStdout()
			profiler, err := builder.BuildProfiler(ctx)
			if err != nil {
				return err
			}

			// return the current context if --current-context is passed
			if currentFlag || len(args) == 0 {
				currentName, err := profiler.GetCurrentProfileName(ctx)
				if err != nil {
					return err
				}
				name = currentName
			} else {
				name = args[0]
			}

			contents, err := profiler.GetProfileRaw(ctx, name)
			if err != nil {
				return err
			}

			fmt.Fprintln(writer, contents)
			return nil
		},
	}

	cmd.Flags().BoolVar(&currentFlag, "current", false, "show the settings for the current profile")
	return cmd
}

// SetCommand returns the BindPlane profile set cobra command
func SetCommand(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <name>",
		Short: "set a parameter on a saved profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing required argument <name>")
			}

			name := args[0]
			ctx := cmd.Context()
			profiler, err := builder.BuildProfiler(ctx)
			if err != nil {
				return err
			}

			if !profiler.ProfileExists(ctx, name) {
				if err := profiler.CreateProfile(ctx, name); err != nil {
					return fmt.Errorf("failed to create profile: %w", err)
				}
			}

			values := map[string]string{}
			cmd.InheritedFlags().VisitAll(func(f *pflag.Flag) {
				if f.Changed {
					values[f.Name] = f.Value.String()
				}
			})

			cmd.Flags().VisitAll(func(f *pflag.Flag) {
				if f.Changed {
					values[f.Name] = f.Value.String()
				}
			})

			return profiler.UpdateProfile(ctx, name, values)
		},
	}

	return cmd
}

// DeleteCommand returns the BindPlane profile delete cobra command
func DeleteCommand(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "delete a saved profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("missing required argument <name>")
			}

			name := args[0]
			ctx := cmd.Context()
			writer := cmd.OutOrStdout()
			profiler, err := builder.BuildProfiler(ctx)
			if err != nil {
				return err
			}

			if err := profiler.DeleteProfile(ctx, name); err != nil {
				return err
			}

			fmt.Fprintf(writer, "deleted saved profile '%s'\n", name)
			return nil
		},
	}
	return cmd
}

// CreateCommand returns the BindPlane profile create cobra command
func CreateCommand(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "create a new profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("missing required argument <name>")
			}

			name := args[0]
			ctx := cmd.Context()
			writer := cmd.OutOrStdout()
			profiler, err := builder.BuildProfiler(ctx)
			if err != nil {
				return err
			}

			if err := profiler.CreateProfile(ctx, name); err != nil {
				return err
			}

			fmt.Fprintf(writer, "created profile '%s'\n", name)
			return nil
		},
	}
	return cmd
}

// ListCommand returns the BindPlane profile list cobra command
func ListCommand(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list the available saved profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			writer := cmd.OutOrStdout()
			profiler, err := builder.BuildProfiler(ctx)
			if err != nil {
				return err
			}

			names, err := profiler.GetProfileNames(ctx)
			if err != nil {
				return err
			}

			if len(names) == 0 {
				fmt.Fprintf(writer, "%s\n", "No saved profiles found.")
			}

			for _, name := range names {
				fmt.Fprintf(writer, "%s\n", name)
			}
			return nil
		},
	}
	return cmd
}

// UseCommand returns the BindPlane profile use cobra command
func UseCommand(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use <name>",
		Short: "specify the default saved context to use",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("missing required argument <name>")
			}

			name := args[0]
			ctx := cmd.Context()
			profiler, err := builder.BuildProfiler(ctx)
			if err != nil {
				return err
			}

			return profiler.SetCurrentProfileName(ctx, name)
		},
	}
	return cmd
}

// CurrentCommand returns the BindPlane profile current cobra command
func CurrentCommand(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "current",
		Short: "returns the name of the currently used profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			writer := cmd.OutOrStdout()
			profiler, err := builder.BuildProfiler(ctx)
			if err != nil {
				return err
			}

			currentName, err := profiler.GetCurrentProfileName(ctx)
			if err != nil {
				return err
			}

			fmt.Fprintf(writer, "%s\n", currentName)
			return nil
		},
	}

	return cmd
}
