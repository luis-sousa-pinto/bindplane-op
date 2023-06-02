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

package logging

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/observiq/bindplane-op/config"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestNewLogger(t *testing.T) {
	t.Run("Invalid log output type", func(t *testing.T) {
		cfg := config.Logging{
			Output: "invalid",
		}

		logger, err := NewLogger(cfg, zapcore.InfoLevel)
		require.ErrorContains(t, err, "unknown log output")
		require.Nil(t, logger)
	})

	t.Run("Stdout logger", func(t *testing.T) {
		cfg := config.Logging{
			Output: config.LoggingOutputStdout,
		}

		// Save stdout and replace with a pipe
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		logger, err := NewLogger(cfg, zapcore.InfoLevel)
		require.NoError(t, err)

		// Log something and close the pipe
		expectedString := "test"
		logger.Info(expectedString)

		w.Close()
		os.Stdout = old

		// Read the pipe and check the output
		var buf bytes.Buffer
		io.Copy(&buf, r)
		require.Contains(t, buf.String(), expectedString)
	})

	t.Run("File logger, configured path", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg := config.Logging{
			Output:   config.LoggingOutputFile,
			FilePath: filepath.Join(tmpDir, "/test.log"),
		}

		logger, err := NewLogger(cfg, zapcore.InfoLevel)
		require.NoError(t, err)

		// Need to log before checking the file exists
		expectedString := "test"
		logger.Info(expectedString)

		// Check if the file exists
		require.FileExists(t, cfg.FilePath)

		data, err := ioutil.ReadFile(cfg.FilePath)
		require.NoError(t, err)

		require.Contains(t, string(data), expectedString)
	})

	t.Run("File logger, default path", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := os.Setenv("BINDPLANE_CONFIG_HOME", tmpDir)
		require.NoError(t, err)
		defer os.Unsetenv("BINDPLANE_CONFIG_HOME")

		cfg := config.Logging{
			Output: config.LoggingOutputFile,
		}

		logger, err := NewLogger(cfg, zapcore.InfoLevel)
		require.NoError(t, err)

		// Need to log before checking the file exists
		expectedString := "test"
		logger.Info(expectedString)

		expectedLogPath := filepath.Join(tmpDir, bindPlaneLogName)

		// Check if the file exists
		require.FileExists(t, expectedLogPath)

		data, err := ioutil.ReadFile(expectedLogPath)
		require.NoError(t, err)

		require.Contains(t, string(data), expectedString)
	})

	t.Run("default config", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := os.Setenv("BINDPLANE_CONFIG_HOME", tmpDir)
		require.NoError(t, err)
		defer os.Unsetenv("BINDPLANE_CONFIG_HOME")

		cfg := config.Logging{}

		logger, err := NewLogger(cfg, zapcore.InfoLevel)
		require.NoError(t, err)

		// Need to log before checking the file exists
		expectedString := "test"
		logger.Info(expectedString)

		expectedLogPath := filepath.Join(tmpDir, bindPlaneLogName)

		// Check if the file exists
		require.FileExists(t, expectedLogPath)

		data, err := ioutil.ReadFile(expectedLogPath)
		require.NoError(t, err)

		require.Contains(t, string(data), expectedString)
	})
}

func TestNewDefaultLoggerAt(t *testing.T) {
	cases := []struct {
		name      string
		level     zapcore.Level
		path      string
		expectErr bool
	}{
		{
			"info",
			zapcore.InfoLevel,
			"/tmp/zap.log",
			false,
		},
		{
			"invalid-path-causes-error",
			zapcore.WarnLevel,
			"/tmp/valid/zap.log",
			true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := newFileLogger(tc.level, tc.path)

			if tc.expectErr {
				require.Error(t, err, "expected an error")
				return
			}

			require.NotNil(t, output)
		})
	}
}

func TestPathToURI(t *testing.T) {
	cases := []struct {
		name   string
		path   string
		goos   string
		expect string
	}{
		{
			"empty",
			"",
			"",
			"",
		},
		{
			"empty-linux",
			"",
			"linux",
			"",
		},
		{
			"empty-darwin",
			"",
			"darwin",
			"",
		},
		{
			"empty-windows",
			"",
			"windows",
			"winfile:///",
		},
		{
			"linux",
			"/var/log/bindplane/bindplane.log",
			"linux",
			"/var/log/bindplane/bindplane.log",
		},
		{
			"empty-windows",
			`D:\observiq\app.log`,
			"windows",
			"winfile:///D:\\observiq\\app.log",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if runtime.GOOS != tc.goos {
				t.Skip("Test not valid for GOOS", tc.goos)
			}
			output := pathToURI(tc.path)
			require.Equal(t, tc.expect, output)
		})
	}
}
