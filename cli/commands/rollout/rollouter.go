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

package rollout

import (
	"context"
	"errors"
	"fmt"

	"github.com/observiq/bindplane-op/cli/printer"
	"github.com/observiq/bindplane-op/client"
	"github.com/observiq/bindplane-op/model"
)

// Rollouter is an interface for managing rollouts
type Rollouter interface {
	// PauseRollout pauses the rollout with rolloutName
	PauseRollout(ctx context.Context, rolloutName string) error

	// ResumeRollout resumes the rollout with rolloutName
	ResumeRollout(ctx context.Context, rolloutName string) error

	// RolloutStatus prints the status of the rollout with rolloutName
	RolloutStatus(ctx context.Context, rolloutName string) error

	// StartRollout starts the rollout with rolloutName
	StartRollout(ctx context.Context, rolloutName string) error

	// StartAllRollouts starts a rollout on all configurations
	StartAllRollouts(ctx context.Context) error

	// UpdateRollout updates a single rollout with rolloutName
	UpdateRollout(ctx context.Context, rolloutName string) error

	// UpdateRollouts updates all rollouts
	UpdateRollouts(ctx context.Context) error
}

// Builder is an interface fo building a Rollouter
type Builder interface {
	BuildRollouter(ctx context.Context) (Rollouter, error)
}

// NewRollouter return a new Rollouter
func NewRollouter(client client.BindPlane, printer printer.Printer) Rollouter {
	return &defaultRollouter{
		client:  client,
		printer: printer,
	}
}

// defaultRollouter is the default implementation of Rollouter
type defaultRollouter struct {
	client  client.BindPlane
	printer printer.Printer
}

// UpdateRollout updates a single rollout with rolloutName
func (d *defaultRollouter) UpdateRollout(ctx context.Context, rolloutName string) error {
	cfg, err := d.client.UpdateRollout(ctx, rolloutName)
	if err != nil {
		return fmt.Errorf("failed to update rollout %s: %w", rolloutName, err)
	}

	d.printer.PrintResource(cfg.Rollout())
	return nil
}

// UpdateRollouts updates all rollouts
func (d *defaultRollouter) UpdateRollouts(ctx context.Context) error {
	cfgs, err := d.client.UpdateRollouts(ctx)
	if err != nil {
		return fmt.Errorf("failed to update rollouts: %w", err)
	}

	rollouts := make([]model.Printable, len(cfgs))
	for i, cfg := range cfgs {
		rollouts[i] = cfg.Rollout()
	}

	d.printer.PrintResources(rollouts)
	return nil
}

// StartRollout starts the rollout with rolloutName
func (d *defaultRollouter) StartRollout(ctx context.Context, rolloutName string) error {
	cfg, err := d.client.StartRollout(ctx, rolloutName, nil)
	if err != nil {
		return fmt.Errorf("failed to start rollout %s: %w", rolloutName, err)
	}

	d.printer.PrintResource(cfg.Rollout())
	return nil
}

// StartAllRollouts starts a rollout on all configurations
func (d *defaultRollouter) StartAllRollouts(ctx context.Context) error {
	cfgs, err := d.client.Configurations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get configurations: %w", err)
	}

	rolloutResults := make([]model.Printable, 0, len(cfgs))

	// Apply rollouts
	var errs error
	for _, cfg := range cfgs {
		rolloutCfg, err := d.client.StartRollout(ctx, cfg.Name(), nil)
		if err != nil {
			rolloutErr := fmt.Errorf("failed to start rollout %s: %w", cfg.Name(), err)
			errs = errors.Join(errs, rolloutErr)
			continue
		}
		rolloutResults = append(rolloutResults, rolloutCfg.Rollout())
	}
	d.printer.PrintResources(rolloutResults)
	return errs
}

// PauseRollout pauses the rollout with rolloutName
func (d *defaultRollouter) PauseRollout(ctx context.Context, rolloutName string) error {
	cfg, err := d.client.PauseRollout(ctx, rolloutName)
	if err != nil {
		return fmt.Errorf("failed to pause rollout %s: %w", rolloutName, err)
	}

	d.printer.PrintResource(cfg.Rollout())
	return nil
}

// ResumeRollout resumes the rollout with rolloutName
func (d *defaultRollouter) ResumeRollout(ctx context.Context, rolloutName string) error {
	cfg, err := d.client.ResumeRollout(ctx, rolloutName)
	if err != nil {
		return fmt.Errorf("failed to pause rollout %s: %w", rolloutName, err)
	}

	d.printer.PrintResource(cfg.Rollout())
	return nil
}

// RolloutStatus prints the status of the rollout with rolloutName
func (d *defaultRollouter) RolloutStatus(ctx context.Context, rolloutName string) error {
	cfg, err := d.client.Configuration(ctx, rolloutName)
	if err != nil {
		return fmt.Errorf("failed to retrieve rollout %s: %w", rolloutName, err)
	}

	d.printer.PrintResource(cfg.Rollout())
	return nil
}
