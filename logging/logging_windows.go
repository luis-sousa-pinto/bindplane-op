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

package logging

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	"github.com/observiq/bindplane-op/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var registerSinkOnce = &sync.Once{}

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
		if err := registerWindowsSink(); err != nil {
			return nil, err
		}

		return newFileLogger(level, determineLogPath(cfg.FilePath))
	default:
		return nil, fmt.Errorf("unknown log output: %s", output)
	}
}

func registerWindowsSink() error {
	var err error
	registerSinkOnce.Do(func() {
		err = zap.RegisterSink("winfile", newWinFileSink)
	})
	if err != nil {
		return fmt.Errorf("failed to register windows file sink: %w", err)
	}
	return nil
}

func newWinFileSink(u *url.URL) (zap.Sink, error) {
	// Ensure permissions restrict access to the running user only
	return os.OpenFile(u.Path[1:], os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
}

func pathToURI(path string) string {
	return "winfile:///" + filepath.ToSlash(path)
}
