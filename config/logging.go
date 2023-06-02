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

package config

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/observiq/bindplane-op/common"
)

const (
	// LoggingOutputFile will write logs to the file specified by LogFilePath
	LoggingOutputFile = "file"

	// LoggingOutputStdout will write logs to stdout
	LoggingOutputStdout = "stdout"
)

// DefaultLoggingFilePath is the default path for the bindplane log file
var DefaultLoggingFilePath = filepath.Join(common.GetHome(), "bindplane.log")

// Logging contains configuration for logging.
type Logging struct {
	// FilePath is the path of the bindplane log file, defaulting to $HOME/.bindplane/bindplane.log
	FilePath string `mapstructure:"filePath" yaml:"filePath,omitempty"`

	// Output indicates where logs should be written, defaulting to "file"
	Output string `mapstructure:"output" yaml:"output,omitempty"`
}

// Validate validates the logging configuration.
func (l *Logging) Validate() error {
	switch l.Output {
	case LoggingOutputFile:
		if l.FilePath == "" {
			return errors.New("file path must be set when output is file")
		}
	case LoggingOutputStdout:
	default:
		return fmt.Errorf("invalid logging output: %s", l.Output)
	}
	return nil
}
