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

package initialize

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/AlecAivazis/survey/v2/terminal"
	expect "github.com/Netflix/go-expect"
	pseudotty "github.com/creack/pty"
	"github.com/hinshun/vt10x"
	"github.com/observiq/bindplane-op/cli/commands/profile"
	"github.com/observiq/bindplane-op/config"
	modelversion "github.com/observiq/bindplane-op/model/version"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestInitializeClient(t *testing.T) {
	testCases := []struct {
		name            string
		initializerFunc func(cfg *config.Config, profileDir, path string) Initializer
		handleConsole   func(console *expect.Console)
		validateFunc    func(t *testing.T, cfg *config.Config, profileDir, path string)
		expectedErr     error
	}{
		{
			name: "valid survey",
			initializerFunc: func(cfg *config.Config, profileDir, path string) Initializer {
				profiler := profile.NewProfiler(profileDir)
				return NewInitializer(cfg, path, profiler)
			},
			handleConsole: func(console *expect.Console) {
				console.ExpectString("Server URL")
				console.SendLine("https://example.com")
				console.ExpectString("Username")
				console.SendLine("username")
				console.ExpectString("Password")
				console.SendLine("password")
				console.ExpectString("********")
			},
			validateFunc: func(t *testing.T, cfg *config.Config, profileDir, path string) {
				diskCfg := &config.Config{}
				data, err := os.ReadFile(path)
				require.NoError(t, err)

				err = yaml.Unmarshal(data, diskCfg)
				require.NoError(t, err)

				expectedCfg := &config.Config{
					APIVersion:  modelversion.V1,
					ProfileName: cfg.ProfileName,
					Network: config.Network{
						RemoteURL: "https://example.com",
					},
					Auth: config.Auth{
						Username: "username",
						Password: "password",
					},
				}
				require.Equal(t, expectedCfg, diskCfg)

				validateCurrentProfileSet(t, profileDir, cfg.ProfileName)
			},
		},
		{
			name: "invalid config permissions",
			initializerFunc: func(cfg *config.Config, profileDir, path string) Initializer {
				profiler := profile.NewProfiler(profileDir)
				os.WriteFile(path, []byte{}, 0400)
				return NewInitializer(cfg, path, profiler)
			},
			handleConsole: func(console *expect.Console) {
				console.ExpectString("Server URL")
				console.SendLine("https://example.com")
				console.ExpectString("Username")
				console.SendLine("username")
				console.ExpectString("Password")
				console.SendLine("password")
				console.ExpectString("********")
			},
			validateFunc: func(t *testing.T, cfg *config.Config, profileDir, path string) {},
			expectedErr:  errors.New("failed to write responses to config"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.NewConfig()
			cfg.ProfileName = tc.name

			tmpDir := t.TempDir()
			cfgPath := filepath.Join(tmpDir, tc.name+".yaml")

			console := newConsole(t)
			defer console.Close()
			go tc.handleConsole(console)

			i := tc.initializerFunc(cfg, tmpDir, cfgPath)
			i.(*defaultInitializer).stdio = terminal.Stdio{In: console.Tty(), Out: console.Tty(), Err: console.Tty()}
			err := i.InitializeClient(context.Background())

			switch tc.expectedErr {
			case nil:
				require.NoError(t, err)
			default:
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr.Error())
			}

			tc.validateFunc(t, cfg, tmpDir, cfgPath)
		})
	}
}

func TestInitializeServer(t *testing.T) {
	testCases := []struct {
		name            string
		initializerFunc func(cfg *config.Config, profileDir, path string) Initializer
		handleConsole   func(console *expect.Console)
		validateFunc    func(t *testing.T, cfg *config.Config, profileDir, path string)
		expectedErr     error
	}{
		{
			name: "valid survey",
			initializerFunc: func(cfg *config.Config, profileDir, path string) Initializer {
				profiler := profile.NewProfiler(profileDir)
				return NewInitializer(cfg, path, profiler)
			},
			handleConsole: func(console *expect.Console) {
				console.ExpectString("Server Host")
				console.SendLine("0.0.0.0")
				console.ExpectString("Server Port")
				console.SendLine("8080")
				console.ExpectString("Remote URL")
				console.SendLine("")
				console.ExpectString("Secret Key")
				console.SendLine("c3e93bfb-caa9-45db-a58d-809561723f28")
				console.ExpectString("Sessions Secret")
				console.SendLine("c3e93bfb-caa9-45db-a58d-809561723f27")
				console.ExpectString("Username")
				console.SendLine("username")
				console.ExpectString("Password (must not be empty)")
				console.SendLine("password")
				console.ExpectString("********")
			},
			validateFunc: func(t *testing.T, cfg *config.Config, profileDir, path string) {
				diskCfg := &config.Config{}
				data, err := os.ReadFile(path)
				require.NoError(t, err)

				err = yaml.Unmarshal(data, diskCfg)
				require.NoError(t, err)

				hostname, _ := os.Hostname()
				expectedCfg := &config.Config{
					APIVersion:  modelversion.V1,
					ProfileName: cfg.ProfileName,
					Network: config.Network{
						Host:      "0.0.0.0",
						Port:      "8080",
						RemoteURL: fmt.Sprintf("http://%s:8080", hostname),
					},
					Auth: config.Auth{
						Username:      "username",
						Password:      "password",
						SecretKey:     "c3e93bfb-caa9-45db-a58d-809561723f28",
						SessionSecret: "c3e93bfb-caa9-45db-a58d-809561723f27",
					},
				}
				require.Equal(t, expectedCfg, diskCfg)

				validateCurrentProfileSet(t, profileDir, cfg.ProfileName)
			},
		},
		{
			name: "preserves current config values for non init questions",
			initializerFunc: func(cfg *config.Config, profileDir, path string) Initializer {
				profiler := profile.NewProfiler(profileDir)
				cfg.Store = config.Store{
					Type:      config.StoreTypeBBolt,
					MaxEvents: 10,
					BBolt: config.BBolt{
						Path: "bbolt/path",
					},
				}
				return NewInitializer(cfg, path, profiler)
			},
			handleConsole: func(console *expect.Console) {
				console.ExpectString("Server Host")
				console.SendLine("0.0.0.0")
				console.ExpectString("Server Port")
				console.SendLine("8080")
				console.ExpectString("Remote URL")
				console.SendLine("")
				console.ExpectString("Secret Key")
				console.SendLine("c3e93bfb-caa9-45db-a58d-809561723f28")
				console.ExpectString("Sessions Secret")
				console.SendLine("c3e93bfb-caa9-45db-a58d-809561723f27")
				console.ExpectString("Username")
				console.SendLine("username")
				console.ExpectString("Password (must not be empty)")
				console.SendLine("password")
				console.ExpectString("********")
			},
			validateFunc: func(t *testing.T, cfg *config.Config, profileDir, path string) {
				diskCfg := &config.Config{}
				data, err := os.ReadFile(path)
				require.NoError(t, err)

				err = yaml.Unmarshal(data, diskCfg)
				require.NoError(t, err)

				hostname, _ := os.Hostname()
				expectedCfg := &config.Config{
					APIVersion:  modelversion.V1,
					ProfileName: cfg.ProfileName,
					Network: config.Network{
						Host:      "0.0.0.0",
						Port:      "8080",
						RemoteURL: fmt.Sprintf("http://%s:8080", hostname),
					},
					Auth: config.Auth{
						Username:      "username",
						Password:      "password",
						SecretKey:     "c3e93bfb-caa9-45db-a58d-809561723f28",
						SessionSecret: "c3e93bfb-caa9-45db-a58d-809561723f27",
					},
					Store: config.Store{
						Type:      config.StoreTypeBBolt,
						MaxEvents: 10,
						BBolt: config.BBolt{
							Path: "bbolt/path",
						},
					},
				}
				require.Equal(t, expectedCfg, diskCfg)

				validateCurrentProfileSet(t, profileDir, cfg.ProfileName)
			},
		},
		{
			name: "Remote URL uses hostname for 127.0.0.1",
			initializerFunc: func(cfg *config.Config, profileDir, path string) Initializer {
				profiler := profile.NewProfiler(profileDir)
				return NewInitializer(cfg, path, profiler)
			},
			handleConsole: func(console *expect.Console) {
				console.ExpectString("Server Host")
				console.SendLine("127.0.0.1")
				console.ExpectString("Server Port")
				console.SendLine("8080")
				console.ExpectString("Remote URL")
				console.SendLine("")
				console.ExpectString("Secret Key")
				console.SendLine("c3e93bfb-caa9-45db-a58d-809561723f28")
				console.ExpectString("Sessions Secret")
				console.SendLine("c3e93bfb-caa9-45db-a58d-809561723f27")
				console.ExpectString("Username")
				console.SendLine("username")
				console.ExpectString("Password (must not be empty)")
				console.SendLine("password")
				console.ExpectString("********")
			},
			validateFunc: func(t *testing.T, cfg *config.Config, profileDir, path string) {
				diskCfg := &config.Config{}
				data, err := os.ReadFile(path)
				require.NoError(t, err)

				err = yaml.Unmarshal(data, diskCfg)
				require.NoError(t, err)

				hostname, _ := os.Hostname()
				expectedCfg := &config.Config{
					APIVersion:  modelversion.V1,
					ProfileName: cfg.ProfileName,
					Network: config.Network{
						Host:      "127.0.0.1",
						Port:      "8080",
						RemoteURL: fmt.Sprintf("http://%s:8080", hostname),
					},
					Auth: config.Auth{
						Username:      "username",
						Password:      "password",
						SecretKey:     "c3e93bfb-caa9-45db-a58d-809561723f28",
						SessionSecret: "c3e93bfb-caa9-45db-a58d-809561723f27",
					},
				}
				require.Equal(t, expectedCfg, diskCfg)

				validateCurrentProfileSet(t, profileDir, cfg.ProfileName)
			},
		},
		{
			name: "use existing password",
			initializerFunc: func(cfg *config.Config, profileDir, path string) Initializer {
				profiler := profile.NewProfiler(profileDir)
				cfg.Auth.Password = "my-password"
				return NewInitializer(cfg, path, profiler)
			},
			handleConsole: func(console *expect.Console) {
				console.ExpectString("Server Host")
				console.SendLine("0.0.0.0")
				console.ExpectString("Server Port")
				console.SendLine("8080")
				console.ExpectString("Remote URL")
				console.SendLine("")
				console.ExpectString("Secret Key")
				console.SendLine("c3e93bfb-caa9-45db-a58d-809561723f28")
				console.ExpectString("Sessions Secret")
				console.SendLine("c3e93bfb-caa9-45db-a58d-809561723f27")
				console.ExpectString("Username")
				console.SendLine("username")
				console.ExpectString("Password (blank will preserve the current password)")
				console.SendLine("")
				console.ExpectString("***********")
			},
			validateFunc: func(t *testing.T, cfg *config.Config, profileDir, path string) {
				diskCfg := &config.Config{}
				data, err := os.ReadFile(path)
				require.NoError(t, err)

				err = yaml.Unmarshal(data, diskCfg)
				require.NoError(t, err)

				hostname, _ := os.Hostname()
				expectedCfg := &config.Config{
					APIVersion:  modelversion.V1,
					ProfileName: cfg.ProfileName,
					Network: config.Network{
						Host:      "0.0.0.0",
						Port:      "8080",
						RemoteURL: fmt.Sprintf("http://%s:8080", hostname),
					},
					Auth: config.Auth{
						Username:      "username",
						Password:      "my-password",
						SecretKey:     "c3e93bfb-caa9-45db-a58d-809561723f28",
						SessionSecret: "c3e93bfb-caa9-45db-a58d-809561723f27",
					},
				}
				require.Equal(t, expectedCfg, diskCfg)

				validateCurrentProfileSet(t, profileDir, cfg.ProfileName)
			},
		},
		{
			name: "use existing sessionSecret and secretKey",
			initializerFunc: func(cfg *config.Config, profileDir, path string) Initializer {
				profiler := profile.NewProfiler(profileDir)
				cfg.Auth.SecretKey = "c3e93bfb-caa9-45db-a58d-809561723f28"
				cfg.Auth.SessionSecret = "32794ad7-b4b6-441e-8438-8248692ec48f"
				return NewInitializer(cfg, path, profiler)
			},
			handleConsole: func(console *expect.Console) {
				console.ExpectString("Server Host")
				console.SendLine("0.0.0.0")
				console.ExpectString("Server Port")
				console.SendLine("8080")
				console.ExpectString("Remote URL")
				console.SendLine("")
				console.ExpectString("Secret Key")
				console.SendLine("")
				console.ExpectString("Sessions Secret")
				console.SendLine("")
				console.ExpectString("Username")
				console.SendLine("username")
				console.ExpectString("Password (must not be empty)")
				console.SendLine("password")
				console.ExpectString("********")
			},
			validateFunc: func(t *testing.T, cfg *config.Config, profileDir, path string) {
				diskCfg := &config.Config{}
				data, err := os.ReadFile(path)
				require.NoError(t, err)

				err = yaml.Unmarshal(data, diskCfg)
				require.NoError(t, err)

				hostname, _ := os.Hostname()
				expectedCfg := &config.Config{
					APIVersion:  modelversion.V1,
					ProfileName: cfg.ProfileName,
					Network: config.Network{
						Host:      "0.0.0.0",
						Port:      "8080",
						RemoteURL: fmt.Sprintf("http://%s:8080", hostname),
					},
					Auth: config.Auth{
						Username:      "username",
						Password:      "password",
						SecretKey:     "c3e93bfb-caa9-45db-a58d-809561723f28",
						SessionSecret: "32794ad7-b4b6-441e-8438-8248692ec48f",
					},
				}
				require.Equal(t, expectedCfg, diskCfg)

				validateCurrentProfileSet(t, profileDir, cfg.ProfileName)
			},
		},
		{
			name: "Set RemoteURL",
			initializerFunc: func(cfg *config.Config, profileDir, path string) Initializer {
				profiler := profile.NewProfiler(profileDir)
				return NewInitializer(cfg, path, profiler)
			},
			handleConsole: func(console *expect.Console) {
				console.ExpectString("Server Host")
				console.SendLine("0.0.0.0")
				console.ExpectString("Server Port")
				console.SendLine("8080")
				console.ExpectString("Remote URL")
				console.SendLine("http://remote.url:8080")
				console.ExpectString("Secret Key")
				console.SendLine("c3e93bfb-caa9-45db-a58d-809561723f28")
				console.ExpectString("Sessions Secret")
				console.SendLine("c3e93bfb-caa9-45db-a58d-809561723f27")
				console.ExpectString("Username")
				console.SendLine("username")
				console.ExpectString("Password (must not be empty)")
				console.SendLine("password")
				console.ExpectString("********")
			},
			validateFunc: func(t *testing.T, cfg *config.Config, profileDir, path string) {
				diskCfg := &config.Config{}
				data, err := os.ReadFile(path)
				require.NoError(t, err)

				err = yaml.Unmarshal(data, diskCfg)
				require.NoError(t, err)

				expectedCfg := &config.Config{
					APIVersion:  modelversion.V1,
					ProfileName: cfg.ProfileName,
					Network: config.Network{
						Host:      "0.0.0.0",
						Port:      "8080",
						RemoteURL: "http://remote.url:8080",
					},
					Auth: config.Auth{
						Username:      "username",
						Password:      "password",
						SecretKey:     "c3e93bfb-caa9-45db-a58d-809561723f28",
						SessionSecret: "c3e93bfb-caa9-45db-a58d-809561723f27",
					},
				}
				require.Equal(t, expectedCfg, diskCfg)

				validateCurrentProfileSet(t, profileDir, cfg.ProfileName)
			},
		},
		{
			name: "default sessionSecret and secretKey causes random values",
			initializerFunc: func(cfg *config.Config, profileDir, path string) Initializer {
				profiler := profile.NewProfiler(profileDir)
				cfg.Auth.SecretKey = config.DefaultSecretKey
				cfg.Auth.SessionSecret = config.DefaultSessionSecret
				return NewInitializer(cfg, path, profiler)
			},
			handleConsole: func(console *expect.Console) {
				console.ExpectString("Server Host")
				console.SendLine("0.0.0.0")
				console.ExpectString("Server Port")
				console.SendLine("8080")
				console.ExpectString("Remote URL")
				console.SendLine("")
				console.ExpectString("Secret Key")
				console.SendLine("")
				console.ExpectString("Sessions Secret")
				console.SendLine("")
				console.ExpectString("Username")
				console.SendLine("username")
				console.ExpectString("Password (must not be empty)")
				console.SendLine("password")
				console.ExpectString("********")
			},
			validateFunc: func(t *testing.T, cfg *config.Config, profileDir, path string) {
				diskCfg := &config.Config{}
				data, err := os.ReadFile(path)
				require.NoError(t, err)

				err = yaml.Unmarshal(data, diskCfg)
				require.NoError(t, err)

				require.NotEqual(t, config.DefaultSecretKey, diskCfg.Auth.SecretKey)
				require.NotEmpty(t, cfg.Auth.SecretKey)

				require.NotEqual(t, config.DefaultSessionSecret, diskCfg.Auth.SessionSecret)
				require.NotEmpty(t, cfg.Auth.SessionSecret)

				validateCurrentProfileSet(t, profileDir, cfg.ProfileName)
			},
		},
		{
			name: "invalid file permissions",
			initializerFunc: func(cfg *config.Config, profileDir, path string) Initializer {
				profiler := profile.NewProfiler(profileDir)
				os.WriteFile(path, []byte{}, 0400)
				return NewInitializer(cfg, path, profiler)
			},
			handleConsole: func(console *expect.Console) {
				console.ExpectString("Server Host")
				console.SendLine("0.0.0.0")
				console.ExpectString("Server Port")
				console.SendLine("8080")
				console.ExpectString("Remote URL")
				console.SendLine("")
				console.ExpectString("Secret Key")
				console.SendLine("c3e93bfb-caa9-45db-a58d-809561723f28")
				console.ExpectString("Sessions Secret")
				console.SendLine("c3e93bfb-caa9-45db-a58d-809561723f27")
				console.ExpectString("Username")
				console.SendLine("username")
				console.ExpectString("Password (must not be empty)")
				console.SendLine("password")
				console.ExpectString("********")
			},
			validateFunc: func(t *testing.T, cfg *config.Config, profileDir, path string) {},
			expectedErr:  errors.New("failed to write responses to config"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.NewConfig()
			cfg.ProfileName = tc.name

			tmpDir := t.TempDir()
			cfgPath := filepath.Join(tmpDir, tc.name+".yaml")

			console := newConsole(t)
			defer console.Close()
			go tc.handleConsole(console)

			i := tc.initializerFunc(cfg, tmpDir, cfgPath)
			i.(*defaultInitializer).stdio = terminal.Stdio{In: console.Tty(), Out: console.Tty(), Err: console.Tty()}
			err := i.InitializeServer(context.Background())

			switch tc.expectedErr {
			case nil:
				require.NoError(t, err)
			default:
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr.Error())
			}

			tc.validateFunc(t, cfg, tmpDir, cfgPath)
		})
	}
}

// newConsole creates a new console for testing.
func newConsole(t *testing.T) *expect.Console {
	pty, tty, err := pseudotty.Open()
	require.NoError(t, err)

	term := vt10x.New(vt10x.WithWriter(tty))
	console, err := expect.NewConsole(expect.WithStdin(pty), expect.WithStdout(term), expect.WithCloser(pty, tty))
	require.NoError(t, err)

	return console
}

// validateCurrentProfileSet loads the current profile file and checks the name
func validateCurrentProfileSet(t *testing.T, profileDir, expectedName string) {
	t.Helper()
	currentPath := filepath.Join(profileDir, "current")
	data, err := os.ReadFile(currentPath)
	require.NoError(t, err)
	require.Contains(t, string(data), expectedName)
}
