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

//go:build !windows

package logging

import (
	"fmt"
	"path/filepath"

	"github.com/observiq/bindplane-op/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger returns a new Logger for the specified config and level
func NewLogger(cfg config.Logging, level zapcore.Level) (*zap.Logger, error) {
	// Setting default value
	output := cfg.Output
	if cfg.Output == "" {
		output = config.LoggingOutputFile
	}

	switch output {
	case config.LoggingOutputStdout:
		return newStdoutLogger(level)
	case config.LoggingOutputFile:
		return newFileLogger(level, determineLogPath(cfg.FilePath))
	default:
		return nil, fmt.Errorf("unknown log output: %s", output)
	}
}

func pathToURI(path string) string {
	return filepath.ToSlash(path)
}
