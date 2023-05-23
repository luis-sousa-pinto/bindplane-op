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

// Package rollout contains the rollout commands for the BindPlane CLI.
package rollout

import (
	"errors"

	"github.com/spf13/cobra"
)

// allFlag when --all is set it starts all rollouts
var allFlag bool

// Command returns the BindPlane rollout cobra command.
func Command(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollout",
		Short: "Manage one or more rollout",
	}
	cmd.AddCommand(
		UpdateCommand(builder),
		StartCommand(builder),
		PauseCommand(builder),
		ResumeCommand(builder),
		StatusCommand(builder),
	)

	return cmd
}

// UpdateCommand the update command runs one cycle of a rollout
func UpdateCommand(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update [configuration]",
		Aliases: []string{"up"},
		Short:   "Updates a rollout for a configuration",
		Long:    "Update runs one cycle of a rollout",
		RunE: func(cmd *cobra.Command, args []string) error {
			rollouter, err := builder.BuildRollouter(cmd.Context())
			if err != nil {
				return err
			}

			switch {
			case len(args) > 1:
				return errors.New("must specify no more than one rollout")
			case len(args) == 1:
				rolloutName := args[0]
				if err := rollouter.UpdateRollout(cmd.Context(), rolloutName); err != nil {
					return err
				}
			default:
				return rollouter.UpdateRollouts(cmd.Context())
			}

			return nil
		},
	}

	return cmd
}

// StartCommand the start command starts a rollout
func StartCommand(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start <configuration>",
		Short: "Starts a rollout for a configuration",
		Long:  "A rollout is a process that applies a configuration to agents in batches",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && !allFlag {
				_ = cmd.Help()
				return nil
			}

			rollouter, err := builder.BuildRollouter(cmd.Context())
			if err != nil {
				return err
			}

			switch {
			case allFlag:
				return rollouter.StartAllRollouts(cmd.Context())
			default:
				rolloutName := args[0]
				return rollouter.StartRollout(cmd.Context(), rolloutName)
			}
		},
	}

	cmd.Flags().BoolVar(&allFlag, "all", false, "start a rollout for every configuration")

	return cmd
}

// PauseCommand the pause command pauses the rollout
func PauseCommand(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pause <configuration>",
		Short: "Pauses the rollout",
		Long:  "A rollout pause pauses the rollout.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				_ = cmd.Help()
				return nil
			}

			rollouter, err := builder.BuildRollouter(cmd.Context())
			if err != nil {
				return err
			}

			rolloutName := args[0]
			return rollouter.PauseRollout(cmd.Context(), rolloutName)
		},
	}

	return cmd
}

// ResumeCommand the resume command resumes a paused rollout
func ResumeCommand(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume <configuration>",
		Short: "Resumes the rollout",
		Long:  "A rollout resume resumes the rollout.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				_ = cmd.Help()
				return nil
			}

			rollouter, err := builder.BuildRollouter(cmd.Context())
			if err != nil {
				return err
			}

			rolloutName := args[0]
			return rollouter.ResumeRollout(cmd.Context(), rolloutName)
		},
	}

	return cmd
}

// StatusCommand the status command retrieves the status of a rollout
func StatusCommand(builder Builder) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <configuration>",
		Short: "Status of the rollout",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				_ = cmd.Help()
				return nil
			}

			rollouter, err := builder.BuildRollouter(cmd.Context())
			if err != nil {
				return err
			}

			rolloutName := args[0]
			return rollouter.RolloutStatus(cmd.Context(), rolloutName)
		},
	}

	return cmd
}
