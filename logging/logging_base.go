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

// Package logging contains the logging logic for BindPlane
package logging

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/observiq/bindplane-op/common"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// bindPlaneLogName is the default name of the BindPlane log file
const bindPlaneLogName = "bindplane.log"

// determineLogPath returns the path to the log file, either the configured path or the default path
func determineLogPath(configuredPath string) string {
	if configuredPath != "" {
		return configuredPath
	}
	return filepath.Join(common.GetHome(), bindPlaneLogName)
}

// newFileLogger takes a logging level and log file path and returns a zip.Logger
func newFileLogger(level zapcore.Level, path string) (*zap.Logger, error) {
	writer := &lumberjack.Logger{
		Filename:   pathToURI(path),
		MaxSize:    100, // mb
		MaxBackups: 10,
		MaxAge:     30,
		Compress:   true,
	}
	core := zapcore.NewCore(newEncoder(), zapcore.AddSync(writer), level)
	return zap.New(core), validatePath(path)
}

// newStdoutLogger returns a new Logger with the specified level, writing to stdout
func newStdoutLogger(level zapcore.Level) (*zap.Logger, error) {
	core := zapcore.NewCore(newEncoder(), zapcore.Lock(os.Stdout), level)
	return zap.New(core), nil
}

func newEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.CallerKey = ""
	encoderConfig.StacktraceKey = ""
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.MessageKey = "message"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	return zapcore.NewJSONEncoder(encoderConfig)
}

// validatePath attempts to create a temp file under the log
// directory.
func validatePath(p string) error {
	dir, _ := filepath.Split(p)

	// If directory is set, ensure it exists
	if dir != "" {
		if _, err := os.Stat(dir); err != nil {
			return fmt.Errorf("log directory: %w", err)
		}
	}

	// Create test file in directory
	f, err := os.CreateTemp(dir, "validate")
	if err != nil {
		return fmt.Errorf("log file creation: %w", err)
	}

	// Grab file path and close right away
	validationPath := f.Name()
	if err := f.Close(); err != nil {
		return fmt.Errorf("close log file %s: %w", validationPath, err)
	}

	if err := os.Remove(f.Name()); err != nil {
		return fmt.Errorf("cleanup log file %s: %w", validationPath, err)
	}

	return nil
}
