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

package profile

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/observiq/bindplane-op/config"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Profiler is an interface for managing BindPlane profiles
type Profiler interface {
	// GetProfileRaw returns the raw representation of a profile.
	GetProfileRaw(ctx context.Context, name string) (string, error)

	// GetCurrentProfileName returns the name of the currently used profile.
	GetCurrentProfileName(ctx context.Context) (string, error)

	// SetCurrentProfileName sets the name of the currently used profile.
	SetCurrentProfileName(ctx context.Context, name string) error

	// GetProfileNames returns a list of profile names.
	GetProfileNames(ctx context.Context) ([]string, error)

	// ProfileExists returns true if the profile exists.
	ProfileExists(ctx context.Context, name string) bool

	// CreateProfile creates a new profile.
	CreateProfile(ctx context.Context, name string) error

	// DeleteProfile deletes a profile.
	DeleteProfile(ctx context.Context, name string) error

	// UpdateProfile updates a profile.
	UpdateProfile(ctx context.Context, name string, values map[string]string) error
}

// Builder is an interface for building a Profiler.
type Builder interface {
	// Build returns a new Profiler.
	BuildProfiler(ctx context.Context) (Profiler, error)
}

// NewProfiler returns a new Profiler.
func NewProfiler(folder string) *DefaultProfiler {
	return &DefaultProfiler{
		folder: folder,
	}
}

// DefaultProfiler is the default implementation of Profiler.
type DefaultProfiler struct {
	folder string
}

// GetProfileRaw returns the raw representation of a profile.
func (p *DefaultProfiler) GetProfileRaw(_ context.Context, name string) (string, error) {
	bytes, err := p.ReadProfileBytes(name)
	if err != nil {
		return "", fmt.Errorf("failed to read profile bytes: %w", err)
	}

	return string(bytes), nil
}

// GetCurrentProfileName returns the name of the current profile.
func (p *DefaultProfiler) GetCurrentProfileName(_ context.Context) (string, error) {
	filename := p.getCurrentProfileFile()
	bytes, err := ioutil.ReadFile(path.Clean(filename))
	if err != nil {
		return "", fmt.Errorf("failed to read current: %w", err)
	}

	var current currentProfile
	if err = yaml.Unmarshal(bytes, &current); err != nil {
		return "", fmt.Errorf("failed to parse current: %w", err)
	}

	if current.Name == "" {
		return "", errors.New("missing current name")
	}

	return current.Name, nil
}

// SetCurrentProfileName sets the name of the currently used profile.
func (p *DefaultProfiler) SetCurrentProfileName(ctx context.Context, name string) error {
	if !p.ProfileExists(ctx, name) {
		return errors.New("profile does not exist")
	}

	filename := p.getCurrentProfileFile()
	contents := fmt.Sprintf("name: %s", name)

	if err := ioutil.WriteFile(filename, []byte(contents), 0600); err != nil {
		return fmt.Errorf("failed to write current: %w", err)
	}

	return nil
}

// GetProfileNames returns a list of profile names.
func (p *DefaultProfiler) GetProfileNames(_ context.Context) ([]string, error) {
	files, err := ioutil.ReadDir(p.folder)
	if err != nil {
		return nil, fmt.Errorf("failed to read profiles folder: %w", err)
	}

	names := []string{}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".yaml") {
			name := strings.TrimSuffix(file.Name(), ".yaml")
			names = append(names, name)
		}
	}

	return names, nil
}

// ProfileExists returns true if the profile with the given name exists.
func (p *DefaultProfiler) ProfileExists(_ context.Context, name string) bool {
	filename := p.getProfilePath(name)
	_, err := os.Stat(filename)
	return err == nil
}

// CreateProfile creates a new profile.
func (p *DefaultProfiler) CreateProfile(ctx context.Context, name string) error {
	if err := p.EnsureProfilesFolder(); err != nil {
		return fmt.Errorf("failed to ensure profiles folder: %w", err)
	}

	if p.ProfileExists(ctx, name) {
		return errors.New("profile already exists")
	}

	cfg := config.NewConfig()
	cfg.ProfileName = name
	bytes, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	filename := p.getProfilePath(name)
	if err := ioutil.WriteFile(filename, bytes, 0600); err != nil {
		return fmt.Errorf("failed to write profile: %w", err)
	}

	return nil
}

// DeleteProfile deletes a profile.
func (p *DefaultProfiler) DeleteProfile(_ context.Context, name string) error {
	filename := p.getProfilePath(name)
	return os.Remove(filename)
}

// UpdateProfile updates a profile using the supplied flags.
func (p *DefaultProfiler) UpdateProfile(_ context.Context, name string, values map[string]string) error {
	cfg, err := p.readProfile(name)
	if err != nil {
		return fmt.Errorf("failed to read profile: %w", err)
	}

	v := viper.New()
	for flag, value := range values {
		for _, override := range config.DefaultOverrides() {
			if override.Flag == flag {
				v.Set(override.Field, value)
			}
		}
	}

	if err := v.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to marshal profile values: %w", err)
	}

	cfg.ProfileName = name
	return p.writeProfile(cfg)
}

// getProfilePath returns the path to the profile with the given name.
func (p *DefaultProfiler) getProfilePath(name string) string {
	return filepath.Join(p.folder, fmt.Sprintf("%s.yaml", name))
}

// getCurrentProfileFile returns the path to the current profile file.
func (p *DefaultProfiler) getCurrentProfileFile() string {
	return filepath.Join(p.folder, "current")
}

// readProfile reads the profile with the given name.
func (p *DefaultProfiler) readProfile(name string) (*config.Config, error) {
	bytes, err := p.ReadProfileBytes(name)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile bytes: %w", err)
	}

	var profile config.Config
	if err = yaml.Unmarshal(bytes, &profile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal profile: %w", err)
	}

	return &profile, nil
}

// ReadProfileBytes reads the raw bytes of a profile.
func (p *DefaultProfiler) ReadProfileBytes(name string) ([]byte, error) {
	filename := p.getProfilePath(name)
	return ioutil.ReadFile(path.Clean(filename))
}

// WriteProfile writes the profile.
func (p *DefaultProfiler) writeProfile(profile *config.Config) error {
	bytes, err := yaml.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	err = p.WriteProfileBytes(profile.ProfileName, bytes)
	if err != nil {
		return fmt.Errorf("failed to write profile bytes: %w", err)
	}

	return nil
}

// WriteProfileBytes writes the profile with raw bytes.
func (p *DefaultProfiler) WriteProfileBytes(name string, bytes []byte) error {
	filename := p.getProfilePath(name)
	return ioutil.WriteFile(filename, bytes, 0600)
}

// EnsureProfilesFolder ensures the profiles folder exists.
func (p *DefaultProfiler) EnsureProfilesFolder() error {
	info, err := os.Stat(p.folder)

	switch {
	case errors.Is(err, os.ErrNotExist):
		return os.MkdirAll(p.folder, 0750)
	case err != nil:
		return fmt.Errorf("failed to access directory: %w", err)
	case !info.IsDir():
		return errors.New("not a directory")
	}

	return nil
}

// currentProfile is the structure of the current profile file.
type currentProfile struct {
	Name string
}
