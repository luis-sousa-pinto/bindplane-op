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

package config

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Override is a configuration override
type Override struct {
	// Field is the config field to override
	Field string
	// Flag is the flag that will override the field
	Flag string
	// Env is the environment variable that will override the field
	Env string
	// Usage is the usage for the override
	Usage string
	// Default is the default value for the override
	Default any
	// ShortHand is the shorthand for the flag
	ShortHand string

	// Hidden signals if the cli flag should be hidden in the help menu
	Hidden bool
}

// NewOverride creates a new override
func NewOverride(field, usage string, def any) *Override {
	return &Override{
		Field:   field,
		Flag:    createFlagName(field),
		Env:     createEnvName(field),
		Usage:   usage,
		Default: def,
		Hidden:  false,
	}
}

// NewHiddenOverride creates a new override that is hidden
func NewHiddenOverride(field, usage string, def any) *Override {
	o := NewOverride(field, usage, def)
	o.Hidden = true
	return o
}

// NewOverrideWithShortHand creates a new override with a shorthand
func NewOverrideWithShortHand(field, shorthand, usage string, def any) *Override {
	o := NewOverride(field, usage, def)
	o.ShortHand = shorthand
	return o
}

// NewOverrideWithoutPrefix creates a new override without the original field prefix for the flag and env
func NewOverrideWithoutPrefix(field, usage string, def any) *Override {
	o := NewOverride(field, usage, def)

	fields := strings.Split(field, ".")
	fieldWithoutPrefix := strings.Join(fields[1:], ".")

	o.Flag = createFlagName(fieldWithoutPrefix)
	o.Env = createEnvName(fieldWithoutPrefix)

	return o
}

// createFlagName creates a flag name from a field
func createFlagName(field string) string {
	// Separate camel case
	updatedField := convertCamelCase(field)

	// Convert separators to dashes
	updatedField = strings.ReplaceAll(updatedField, ".", "-")

	// Convert to lowercase
	return strings.ToLower(updatedField)
}

// createEnvName creates an environment variable name from a field
func createEnvName(field string) string {
	// Separate camel case
	updatedField := convertCamelCase(field)

	// Convert separators to underscores
	updatedField = strings.ReplaceAll(updatedField, ".", "_")

	// Convert to uppercase
	updatedField = strings.ToUpper(updatedField)

	// Add BINDPLANE prefix
	return fmt.Sprintf("BINDPLANE_%s", updatedField)
}

// Bind binds the override to the viper instance
func (o *Override) Bind(flags *pflag.FlagSet) error {
	flag, err := o.createFlag(flags)
	if err != nil {
		return fmt.Errorf("failed while created flag %s: %w", o.Flag, err)
	} else if flag == nil {
		return errors.New("flag not found")
	}

	if err := viper.BindPFlag(o.Field, flag); err != nil {
		return fmt.Errorf("failed to bind pflag: %w", err)
	}

	if err := viper.BindEnv(o.Field, o.Env); err != nil {
		return fmt.Errorf("failed to bind env: %w", err)
	}

	return nil
}

// createFlag creates a flag for the override
func (o *Override) createFlag(flags *pflag.FlagSet) (*pflag.Flag, error) {
	if exitingFlag := flags.Lookup(o.Flag); exitingFlag != nil {
		return exitingFlag, nil
	}

	switch v := o.Default.(type) {
	case string:
		if o.ShortHand != "" {
			_ = flags.StringP(o.Flag, o.ShortHand, v, o.Usage)
		} else {
			_ = flags.String(o.Flag, v, o.Usage)
		}
	case []string:
		_ = flags.StringSlice(o.Flag, v, o.Usage)
	case bool:
		_ = flags.Bool(o.Flag, v, o.Usage)
	case int:
		_ = flags.Int(o.Flag, v, o.Usage)
	case time.Duration:
		_ = flags.Duration(o.Flag, v, o.Usage)
	default:
		_ = flags.String(o.Flag, "", o.Usage)
	}

	if o.Hidden {
		if err := flags.MarkHidden(o.Flag); err != nil {
			return nil, err
		}
	}

	return flags.Lookup(o.Flag), nil
}

// convertCamelCase converts camel case
func convertCamelCase(s string) string {
	var buf bytes.Buffer
	var last rune
	for _, r := range s {
		if unicode.IsUpper(r) && last != 0 && !unicode.IsUpper(last) {
			buf.WriteRune('.')
		}
		buf.WriteRune(unicode.ToLower(r))
		last = r
	}
	return buf.String()
}

// DefaultOverrides returns the default overrides
func DefaultOverrides() []*Override {
	return []*Override{
		// Standard overrides
		NewOverride("env", "env to use. One of test|development|production", EnvProduction),
		NewOverrideWithShortHand("output", "o", "output format for client commands. One of: json|table|yaml|raw", DefaultOutput),
		NewOverride("offline", "whether the server is in offline mode", false),
		NewOverride("rolloutsInterval", "interval between updates of rollouts", DefaultRolloutsInterval),

		// Logging overrides
		NewOverride("logging.output", "output of the log. One of: file|stdout", LoggingOutputFile),
		NewOverride("logging.filePath", "path to the log file", DefaultLoggingFilePath),

		// Network overrides
		NewOverrideWithoutPrefix("network.host", "domain on which the server will run", DefaultHost),
		NewOverrideWithoutPrefix("network.port", "port on which the server will run", DefaultPort),
		NewOverrideWithoutPrefix("network.remoteURL", "the URL used by the client to reach the server", ""),
		NewOverrideWithoutPrefix("network.tlsCert", "the path to the TLS certificate", ""),
		NewOverrideWithoutPrefix("network.tlsKey", "the path to the TLS private key", ""),
		NewOverrideWithoutPrefix("network.tlsCA", "the path to the TLS CA files", []string{}),
		NewOverrideWithoutPrefix("network.tlsSkipVerify", "whether to skip TLS verification", false),

		// Auth overrides
		NewOverrideWithoutPrefix("auth.username", "username for basic auth", DefaultUsername),
		NewOverrideWithoutPrefix("auth.password", "password for basic auth", DefaultPassword),
		NewOverrideWithoutPrefix("auth.secretKey", "secret key for agent auth", DefaultSecretKey),
		NewOverrideWithoutPrefix("auth.sessionSecret", "secret used to encode sessions", DefaultSessionSecret),

		// Tracing overrides
		NewOverride("tracing.type", "the type of tracing to use. One of: otlp|google", ""),
		NewOverride("tracing.otlp.endpoint", "the endpoint to send tracing data to, if using OTLP", ""),
		NewOverride("tracing.otlp.insecure", "whether to use insecure TLS for tracing", false),

		// Store overrides
		NewOverride("store.type", "the type of store to use. One of: bbolt|mapstore", StoreTypeBBolt),
		NewOverride("store.bbolt.path", "the path to the store file", DefaultBBoltPath),
		NewOverride("store.maxEvents", "the maximum number of events to batch in a store operation", DefaultMaxEvents),

		// Agent version overrides
		NewOverride("agentVersions.syncInterval", "the interval at which to sync agent versions", DefaultSyncInterval),
	}
}
