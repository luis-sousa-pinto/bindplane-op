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
	"context"
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
		Use:   "get [name]",
		Short: "Get details of a profile",
		Long:  "Get details of a profile. If no name is specified, the current profile is returned.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			writer := cmd.OutOrStdout()
			profiler, err := builder.BuildProfiler(ctx)
			if err != nil {
				return err
			}

			name, err := profileNameArgOrCurrent(ctx, args, profiler)
			if err != nil {
				return err
			}

			contents, err := profiler.GetProfileRaw(ctx, name)
			if err != nil {
				return err
			}

			fmt.Fprintln(writer, contents)
			return nil
		},
	}

	cmd.Flags().BoolVar(&currentFlag, "current", false, "get the settings for the current profile")
	_ = cmd.Flags().MarkHidden("current")
	return cmd
}

// SetCommand returns the BindPlane profile set cobra command
func SetCommand(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set [name]",
		Short: "Set details of a profile",
		Long:  "Set details of a profile. If no name is specified, the current profile is modified.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			profiler, err := builder.BuildProfiler(ctx)
			if err != nil {
				return err
			}

			name, err := profileNameArgOrCurrent(ctx, args, profiler)
			if err != nil {
				// rather than fail, just create a new profile called "default"
				name = "default"
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

	cmd.Flags().BoolVar(&currentFlag, "current", false, "set the settings for the current profile")
	_ = cmd.Flags().MarkHidden("current")
	return cmd
}

// DeleteCommand returns the BindPlane profile delete cobra command
func DeleteCommand(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a profile",
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

			fmt.Fprintf(writer, "deleted profile '%s'\n", name)
			return nil
		},
	}
	return cmd
}

// CreateCommand returns the BindPlane profile create cobra command
func CreateCommand(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new profile",
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
		Short: "List all profiles",
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
				fmt.Fprintf(writer, "%s\n", "No profiles found.")
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
		Short: "Set the current profile to the specified profile",
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
		Short: "Print the name of the current profile",
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

func profileNameArgOrCurrent(ctx context.Context, args []string, profiler Profiler) (string, error) {
	if currentFlag || len(args) == 0 {
		return profiler.GetCurrentProfileName(ctx)
	}
	return args[0], nil
}
