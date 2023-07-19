// Copyright  observIQ, Inc
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

// Package cli provides the BindPlane command line client.
package cli

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/observiq/bindplane-op/cli/commands/apply"
	"github.com/observiq/bindplane-op/cli/commands/copy"
	"github.com/observiq/bindplane-op/cli/commands/delete"
	"github.com/observiq/bindplane-op/cli/commands/get"
	"github.com/observiq/bindplane-op/cli/commands/initialize"
	"github.com/observiq/bindplane-op/cli/commands/install"
	"github.com/observiq/bindplane-op/cli/commands/label"
	"github.com/observiq/bindplane-op/cli/commands/profile"
	"github.com/observiq/bindplane-op/cli/commands/rollout"
	"github.com/observiq/bindplane-op/cli/commands/serve"
	"github.com/observiq/bindplane-op/cli/commands/sync"
	"github.com/observiq/bindplane-op/cli/commands/update"
	"github.com/observiq/bindplane-op/cli/commands/version"
	"github.com/observiq/bindplane-op/cli/printer"
	"github.com/observiq/bindplane-op/client"
	"github.com/observiq/bindplane-op/common"
	"github.com/observiq/bindplane-op/config"
	"github.com/observiq/bindplane-op/logging"
	"github.com/observiq/bindplane-op/server"
	"github.com/observiq/bindplane-op/store"
	"github.com/observiq/bindplane-op/tracer"
	"github.com/spf13/viper"
)

// Factory is used to build and load resources for the CLI.
type Factory struct {
	cfg          *config.Config
	cfgPath      string
	writer       io.Writer
	logger       *zap.Logger
	routeBuilder server.RouteBuilder
}

// LoadConfig loads the configuration from the given path.
func (f *Factory) LoadConfig(_ context.Context, path string) error {
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading in configuration file: %s, %w", path, err)
	}

	cfg := config.NewConfig()
	if err := viper.Unmarshal(cfg); err != nil {
		return err
	}

	f.cfg = cfg
	f.cfgPath = path
	return nil
}

// ValidateConfig validates the configuration.
func (f *Factory) ValidateConfig(_ context.Context) error {
	if err := f.cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration file: %w", err)
	}
	return nil
}

// LoadProfile loads the BindPlane profile from the given path.
func (f *Factory) LoadProfile(ctx context.Context, name string) error {
	if name == "" {
		return f.LoadCurrentProfile(ctx)
	}

	profilesFolder := common.GetProfilesFolder()
	profileName := fmt.Sprintf("%s.yaml", name)
	path := filepath.Join(profilesFolder, profileName)
	return f.LoadConfig(ctx, path)
}

// LoadCurrentProfile loads the current profile.
func (f *Factory) LoadCurrentProfile(ctx context.Context) error {
	profiler, err := f.BuildProfiler(ctx)
	if err != nil {
		return fmt.Errorf("failed to build profiler: %w", err)
	}

	name, err := profiler.GetCurrentProfileName(ctx)
	if err != nil {
		return f.LoadEmptyConfig(ctx)
	}

	return f.LoadProfile(ctx, name)
}

// LoadEmptyConfig loads an empty configuration set to the default profile
func (f *Factory) LoadEmptyConfig(_ context.Context) error {
	cfg := config.NewConfig()
	err := viper.Unmarshal(cfg)
	if err != nil {
		return err
	}

	// Set the default profile name
	cfg.ProfileName = common.DefaultProfileName

	f.cfg = cfg
	f.cfgPath = filepath.Join(common.GetProfilesFolder(), fmt.Sprintf("%s.yaml", common.DefaultProfileName))
	return nil
}

// BuildProfiler builds a profiler.
func (f *Factory) BuildProfiler(_ context.Context) (profile.Profiler, error) {
	profilesFolder := common.GetProfilesFolder()
	return profile.NewProfiler(profilesFolder), nil
}

// BuildVersioner builds a versioner.
func (f *Factory) BuildVersioner(ctx context.Context) (version.Versioner, error) {
	c, err := f.BuildClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build client: %w", err)
	}

	return version.NewVersioner(c), nil
}

// BuildUpdater builds an updater.
func (f *Factory) BuildUpdater(ctx context.Context) (update.Updater, error) {
	c, err := f.BuildClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build client: %w", err)
	}

	return update.NewUpdater(c), nil
}

// BuildSyncer builds a syncer.
func (f *Factory) BuildSyncer(ctx context.Context) (sync.Syncer, error) {
	c, err := f.BuildClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build client: %w", err)
	}

	return sync.NewSyncer(c), nil
}

// BuildLabeler builds a labeler.
func (f *Factory) BuildLabeler(ctx context.Context) (label.Labeler, error) {
	c, err := f.BuildClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build client: %w", err)
	}

	return label.NewLabeler(c), nil
}

// BuildInstaller builds an installer.
func (f *Factory) BuildInstaller(ctx context.Context) (install.Installer, error) {
	c, err := f.BuildClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build client: %w", err)
	}

	return install.NewInstaller(c), nil
}

// BuildInitializer builds an initializer.
func (f *Factory) BuildInitializer(ctx context.Context) (initialize.Initializer, error) {
	profiler, err := f.BuildProfiler(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build profiler: %w", err)
	}

	return initialize.NewInitializer(f.cfg, f.cfgPath, profiler), nil
}

// BuildGetter builds a getter.
func (f *Factory) BuildGetter(ctx context.Context) (get.Getter, error) {
	c, err := f.BuildClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build client: %w", err)
	}

	printer, err := f.BuildPrinter(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build printer: %w", err)
	}

	return get.NewGetter(c, printer, f.cfg.Output), nil
}

// BuildRollouter builds a rollouter.
func (f *Factory) BuildRollouter(ctx context.Context) (rollout.Rollouter, error) {
	c, err := f.BuildClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build client: %w", err)
	}

	printer, err := f.BuildPrinter(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build printer: %w", err)
	}

	return rollout.NewRollouter(c, printer), nil
}

// BuildDeleter builds a deleter.
func (f *Factory) BuildDeleter(ctx context.Context) (delete.Deleter, error) {
	c, err := f.BuildClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build client: %w", err)
	}

	return delete.NewDeleter(c), nil
}

// BuildCopier builds a copier.
func (f *Factory) BuildCopier(ctx context.Context) (copy.Copier, error) {
	c, err := f.BuildClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build client: %w", err)
	}

	return copy.NewCopier(c), nil
}

// BuildApplier builds an applier.
func (f *Factory) BuildApplier(ctx context.Context) (apply.Applier, error) {
	c, err := f.BuildClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build client: %w", err)
	}

	return apply.NewApplier(c), nil
}

// BuildServer builds a server.
func (f *Factory) BuildServer(ctx context.Context) (serve.Server, error) {
	logger, err := f.BuildLogger(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	st, err := f.BuildStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load store: %w", err)
	}

	tr, err := f.BuildTracer(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load tracer: %w", err)
	}

	return serve.NewServer(f.cfg, st, tr, logger, f.routeBuilder), nil
}

// SupportsServer returns true if the OS is not windows.
func (f *Factory) SupportsServer() bool {
	return runtime.GOOS != "windows"
}

// BuildLogger builds a logger, if it doesn't already exist.
func (f *Factory) BuildLogger(_ context.Context) (*zap.Logger, error) {
	if f.logger != nil {
		return f.logger, nil
	}

	logLevel := zapcore.DebugLevel
	if f.cfg.Env == config.EnvProduction {
		logLevel = zapcore.InfoLevel
	}

	logger, err := logging.NewLogger(f.cfg.Logging, logLevel)
	if err != nil {
		return nil, err
	}

	f.logger = logger
	return logger, nil
}

// BuildPrinter builds a printer.
func (f *Factory) BuildPrinter(ctx context.Context) (printer.Printer, error) {
	logger, err := f.BuildLogger(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return printer.Build(f.cfg.Output, f.writer, logger), nil
}

// BuildClient builds a client.
func (f *Factory) BuildClient(ctx context.Context) (client.BindPlane, error) {
	logger, err := f.BuildLogger(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return client.NewBindPlane(f.cfg, logger)
}

// BuildTracer builds the tracer for the server
func (f *Factory) BuildTracer(_ context.Context) (tracer.Tracer, error) {
	cfg := &f.cfg.Tracing
	samplingRate := math.Min(math.Max(cfg.SamplingRate, 0), 1)
	resource := tracer.DefaultResource()

	switch cfg.Type {
	case config.TracerTypeOTLP:
		return tracer.NewOTLP(&cfg.OTLP, samplingRate, resource), nil
	case config.TracerTypeGoogleCloud:
		return tracer.NewGoogleCloud(&cfg.GoogleCloud, samplingRate, resource), nil
	case config.TracerTypeNop:
		return tracer.NewNop(), nil
	default:
		return nil, fmt.Errorf("unknown tracer type: %s", cfg.Type)
	}
}

// BuildStore builds the store for the server
func (f *Factory) BuildStore(ctx context.Context) (store.Store, error) {
	logger, err := f.BuildLogger(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	options := store.Options{
		SessionsSecret:   f.cfg.Auth.SessionSecret,
		MaxEventsToMerge: f.cfg.Store.MaxEvents,
	}

	switch f.cfg.Store.Type {
	case config.StoreTypeMap:
		return store.NewMapStore(ctx, options, logger), nil
	case config.StoreTypeBBolt:
		db, err := store.InitBoltstoreDB(f.cfg.Store.BBolt.Path)
		if err != nil {
			return nil, fmt.Errorf("bbolt storage file failed to open: %w", err)
		}
		return store.NewBoltStore(ctx, db, options, logger), nil
	default:
		return nil, fmt.Errorf("unknown store type: %s", f.cfg.Store.Type)
	}
}

// NewFactory creates a new factory.
func NewFactory(routeBuilder server.RouteBuilder) *Factory {
	return &Factory{
		writer:       os.Stdout,
		routeBuilder: routeBuilder,
	}
}
