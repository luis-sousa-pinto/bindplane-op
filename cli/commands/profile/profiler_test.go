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
	"os"
	"path/filepath"
	"testing"

	"github.com/observiq/bindplane-op/config"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestGetProfileRaw(t *testing.T) {
	testCases := []struct {
		name             string
		profileName      string
		setupFunc        func(profileFolder string) error
		expectedContents string
		expectedError    error
	}{
		{
			name:          "missing profile",
			profileName:   "missing",
			setupFunc:     func(profileFolder string) error { return nil },
			expectedError: errors.New("failed to read profile bytes"),
		},
		{
			name:        "valid profile",
			profileName: "valid",
			setupFunc: func(profileFolder string) error {
				contents := []byte("apiVersion: 0.0.0")
				filename := filepath.Join(profileFolder, "valid.yaml")
				return os.WriteFile(filename, contents, 0644)
			},
			expectedContents: "apiVersion: 0.0.0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			folder := newTestFolder(t, tc.name)
			err := tc.setupFunc(folder)
			require.NoError(t, err)

			p := NewProfiler(folder)
			contents, err := p.GetProfileRaw(context.Background(), tc.profileName)

			switch tc.expectedError {
			case nil:
				require.NoError(t, err)
				require.Equal(t, tc.expectedContents, contents)
			default:
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedError.Error())
			}
		})
	}
}

func TestGetCurrentProfileName(t *testing.T) {
	testCases := []struct {
		name          string
		setupFunc     func(profileFolder string) error
		expectedName  string
		expectedError error
	}{
		{
			name:          "missing profile",
			setupFunc:     func(profileFolder string) error { return nil },
			expectedError: errors.New("failed to read current"),
		},
		{
			name: "invalid yaml",
			setupFunc: func(profileFolder string) error {
				contents := []byte("invalid yaml")
				profilePath := filepath.Join(profileFolder, "current")
				return os.WriteFile(profilePath, contents, 0644)
			},
			expectedError: errors.New("failed to parse current"),
		},
		{
			name: "missing name",
			setupFunc: func(profileFolder string) error {
				contents := []byte("key: value")
				profilePath := filepath.Join(profileFolder, "current")
				return os.WriteFile(profilePath, contents, 0644)
			},
			expectedError: errors.New("missing current name"),
		},
		{
			name: "valid profile",
			setupFunc: func(profileFolder string) error {
				contents := []byte("name: valid")
				profilePath := filepath.Join(profileFolder, "current")
				return os.WriteFile(profilePath, contents, 0644)
			},
			expectedName: "valid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			folder := newTestFolder(t, tc.name)
			err := tc.setupFunc(folder)
			require.NoError(t, err)

			p := NewProfiler(folder)
			contents, err := p.GetCurrentProfileName(context.Background())

			switch tc.expectedError {
			case nil:
				require.NoError(t, err)
				require.Equal(t, tc.expectedName, contents)
			default:
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedError.Error())
			}
		})
	}
}

func TestSetCurrentProfileName(t *testing.T) {
	testCases := []struct {
		name          string
		setupFunc     func(profileFolder string) error
		profileName   string
		expectedError error
	}{
		{
			name:          "missing profile",
			setupFunc:     func(profileFolder string) error { return nil },
			profileName:   "missing",
			expectedError: errors.New("profile does not exist"),
		},
		{
			name: "readonly current",
			setupFunc: func(profileFolder string) error {
				contents := []byte("apiVersion: 0.0.0")
				filename := filepath.Join(profileFolder, "profile.yaml")
				if err := os.WriteFile(filename, contents, 0644); err != nil {
					return err
				}

				currentPath := filepath.Join(profileFolder, "current")
				return os.WriteFile(currentPath, contents, 0444)
			},
			profileName:   "profile",
			expectedError: errors.New("failed to write current"),
		},
		{
			name: "valid profile",
			setupFunc: func(profileFolder string) error {
				contents := []byte("apiVersion: 0.0.0")
				filename := filepath.Join(profileFolder, "valid.yaml")
				return os.WriteFile(filename, contents, 0644)
			},
			profileName: "valid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			folder := newTestFolder(t, tc.name)
			err := tc.setupFunc(folder)
			require.NoError(t, err)

			p := NewProfiler(folder)
			err = p.SetCurrentProfileName(context.Background(), tc.profileName)

			switch tc.expectedError {
			case nil:
				require.NoError(t, err)
			default:
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedError.Error())
			}
		})
	}
}

func TestGetProfileNames(t *testing.T) {
	testCases := []struct {
		name          string
		setupFunc     func(profileFolder string) error
		expectedNames []string
		expectedError error
	}{
		{
			name: "missing directory",
			setupFunc: func(profileFolder string) error {
				return os.RemoveAll(profileFolder)
			},
			expectedError: errors.New("failed to read profiles folder"),
		},
		{
			name:          "empty directory",
			setupFunc:     func(profileFolder string) error { return nil },
			expectedNames: []string{},
		},
		{
			name: "valid profiles",
			setupFunc: func(profileFolder string) error {
				contents := []byte("apiVersion: 0.0.0")
				filename1 := filepath.Join(profileFolder, "profile-1.yaml")
				if err := os.WriteFile(filename1, contents, 0644); err != nil {
					return err
				}

				filename2 := filepath.Join(profileFolder, "profile-2.yaml")
				return os.WriteFile(filename2, contents, 0644)
			},
			expectedNames: []string{"profile-1", "profile-2"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			folder := newTestFolder(t, tc.name)
			err := tc.setupFunc(folder)
			require.NoError(t, err)

			p := NewProfiler(folder)
			names, err := p.GetProfileNames(context.Background())

			switch tc.expectedError {
			case nil:
				require.NoError(t, err)
				require.Equal(t, tc.expectedNames, names)
			default:
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedError.Error())
			}
		})
	}
}

func TestCreateProfile(t *testing.T) {
	testCases := []struct {
		name          string
		profileName   string
		setupFunc     func(profileFolder string) error
		validateFunc  func(t *testing.T, profileFolder string)
		expectedError error
	}{
		{
			name:        "valid create",
			profileName: "test",
			setupFunc: func(profileFolder string) error {
				return nil
			},
			validateFunc: func(t *testing.T, profileFolder string) {
				bytes, err := os.ReadFile(filepath.Join(profileFolder, "test.yaml"))
				require.NoError(t, err)

				cfg := config.NewConfig()
				require.NoError(t, yaml.Unmarshal(bytes, cfg))
				require.Equal(t, "test", cfg.ProfileName)
			},
		},
		{
			name: "existing profile",
			setupFunc: func(profileFolder string) error {
				contents := []byte("apiVersion: 0.0.0")
				filename := filepath.Join(profileFolder, "test.yaml")
				return os.WriteFile(filename, contents, 0644)
			},
			validateFunc:  func(t *testing.T, profileFolder string) {},
			profileName:   "test",
			expectedError: errors.New("profile already exists"),
		},
		{
			name: "missing profiles folder",
			setupFunc: func(profileFolder string) error {
				return os.RemoveAll(profileFolder)
			},
			profileName: "test",
			validateFunc: func(t *testing.T, profileFolder string) {
				_, err := os.ReadFile(filepath.Join(profileFolder, "test.yaml"))
				require.NoError(t, err)
			},
		},
		{
			name: "profiles folder is a file",
			setupFunc: func(profileFolder string) error {
				if err := os.RemoveAll(profileFolder); err != nil {
					return err
				}

				return os.WriteFile(profileFolder, []byte("test"), 0644)
			},
			validateFunc:  func(t *testing.T, profileFolder string) {},
			profileName:   "test",
			expectedError: errors.New("not a directory"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			folder := newTestFolder(t, tc.name)
			err := tc.setupFunc(folder)
			require.NoError(t, err)

			p := NewProfiler(folder)
			err = p.CreateProfile(context.Background(), tc.profileName)

			switch tc.expectedError {
			case nil:
				require.NoError(t, err)
			default:
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedError.Error())
			}

			tc.validateFunc(t, folder)
		})
	}
}

func TestDeleteProfile(t *testing.T) {
	testCases := []struct {
		name          string
		setupFunc     func(profileFolder string) error
		profileName   string
		expectedError error
	}{
		{
			name:          "missing profile",
			setupFunc:     func(profileFolder string) error { return nil },
			profileName:   "missing",
			expectedError: errors.New("no such file or directory"),
		},
		{
			name: "valid profile",
			setupFunc: func(profileFolder string) error {
				contents := []byte("apiVersion: 0.0.0")
				filename := filepath.Join(profileFolder, "valid.yaml")
				return os.WriteFile(filename, contents, 0644)
			},
			profileName: "valid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			folder := newTestFolder(t, tc.name)
			err := tc.setupFunc(folder)
			require.NoError(t, err)

			p := NewProfiler(folder)
			err = p.DeleteProfile(context.Background(), tc.profileName)

			switch tc.expectedError {
			case nil:
				require.NoError(t, err)
			default:
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedError.Error())
			}
		})
	}
}

func TestUpdateProfile(t *testing.T) {
	testCases := []struct {
		name            string
		setupFunc       func(profileFolder string) error
		profileName     string
		values          map[string]string
		expectedProfile *config.Config
		expectedError   error
	}{
		{
			name:          "missing profile",
			setupFunc:     func(profileFolder string) error { return nil },
			profileName:   "missing",
			expectedError: errors.New("no such file or directory"),
		},
		{
			name: "empty values",
			setupFunc: func(profileFolder string) error {
				contents := []byte("apiVersion: 0.0.0")
				filename := filepath.Join(profileFolder, "profile.yaml")
				return os.WriteFile(filename, contents, 0644)
			},
			profileName: "profile",
			values:      map[string]string{},
			expectedProfile: &config.Config{
				APIVersion:  "0.0.0",
				ProfileName: "profile",
			},
		},
		{
			name: "failed read",
			setupFunc: func(profileFolder string) error {
				contents := []byte("invalid yaml")
				filename := filepath.Join(profileFolder, "profile.yaml")
				return os.WriteFile(filename, contents, 0644)
			},
			profileName:   "profile",
			values:        map[string]string{},
			expectedError: errors.New("failed to unmarshal profile"),
		},
		{
			name: "readonly profile",
			setupFunc: func(profileFolder string) error {
				contents := []byte("apiVersion: 0.0.0")
				filename := filepath.Join(profileFolder, "profile.yaml")
				return os.WriteFile(filename, contents, 0444)
			},
			profileName:   "profile",
			values:        map[string]string{},
			expectedError: errors.New("failed to write profile bytes"),
		},
		{
			name: "valid values",
			setupFunc: func(profileFolder string) error {
				contents := []byte("apiVersion: 0.0.0")
				filename := filepath.Join(profileFolder, "profile.yaml")
				return os.WriteFile(filename, contents, 0644)
			},
			profileName: "profile",
			values: map[string]string{
				"logging-output":    "stdout",
				"logging-file-path": "/log/test.log",
				"port":              "0",
				"host":              "localhost",
				"env":               "test",
				"username":          "user",
				"password":          "pass",
				"secret-key":        "secret",
				"session-secret":    "11fc92ba-c045-11ed-afa1-0242ac120002",
				"offline":           "true",
				"store-type":        "bbolt",
				"store-bbolt-path":  "/storage",
				"remote-url":        "http://localhost:1234",
				"output":            "json",
			},
			expectedProfile: &config.Config{
				APIVersion:  "0.0.0",
				ProfileName: "profile",
				Env:         "test",
				Output:      "json",
				Offline:     true,
				Auth: config.Auth{
					Username:      "user",
					Password:      "pass",
					SecretKey:     "secret",
					SessionSecret: "11fc92ba-c045-11ed-afa1-0242ac120002",
				},
				Network: config.Network{
					Host:      "localhost",
					Port:      "0",
					RemoteURL: "http://localhost:1234",
				},
				Store: config.Store{
					Type: "bbolt",
					BBolt: config.BBolt{
						Path: "/storage",
					},
				},
				Logging: config.Logging{
					Output:   "stdout",
					FilePath: "/log/test.log",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			folder := newTestFolder(t, tc.name)
			err := tc.setupFunc(folder)
			require.NoError(t, err)

			p := NewProfiler(folder)
			err = p.UpdateProfile(context.Background(), tc.profileName, tc.values)

			switch tc.expectedError {
			case nil:
				require.NoError(t, err)

				rawProfile, err := p.GetProfileRaw(context.Background(), tc.profileName)
				require.NoError(t, err)

				cfg := &config.Config{}
				err = yaml.Unmarshal([]byte(rawProfile), cfg)
				require.NoError(t, err)
				require.Equal(t, tc.expectedProfile, cfg)
			default:
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedError.Error())
			}
		})
	}
}

// newTestFolder returns a new temporary folder for the test.
func newTestFolder(t *testing.T, name string) string {
	folder := filepath.Join(t.TempDir(), name)
	err := os.Mkdir(folder, 0755)
	require.NoError(t, err)
	return folder
}
