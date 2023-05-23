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

// Package legacy is used to convert legacy configs to new configs
package legacy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/observiq/bindplane-op/config"
	modelversion "github.com/observiq/bindplane-op/model/version"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// BackupLegacyCfg backs up the legacy config if it needs to be
func BackupLegacyCfg(path string) error {
	cleanPath := filepath.Clean(path)
	backupPath := fmt.Sprintf("%s.backup", cleanPath)
	_, err := os.Stat(backupPath)
	if err == nil {
		// file exists already so exit
		return nil
	}

	fileContents, err := os.ReadFile(cleanPath)
	if err != nil {
		return err
	}

	return os.WriteFile(backupPath, fileContents, 0600)
}

// ConvertFile converts a legacy config file to a new config file.
func ConvertFile(path string) error {
	cleanPath := filepath.Clean(path)
	bytes, err := os.ReadFile(cleanPath)
	if err != nil {
		return err
	}

	var newCfg config.Config
	if err := yaml.Unmarshal(bytes, &newCfg); err != nil {
		return err
	}

	// If APIVersion is set, then the config is already converted.
	if newCfg.APIVersion != "" {
		return nil
	}

	// backup the legacy config
	if err := BackupLegacyCfg(cleanPath); err != nil {
		return fmt.Errorf("failed while backing up config: %w", err)
	}

	legacyCfg, err := readLegacyCfg(cleanPath)
	if err != nil {
		return err
	}

	newCfg = Convert(*legacyCfg)
	newCfg.ProfileName = getProfileName(path)

	bytes, err = yaml.Marshal(&newCfg)
	if err != nil {
		return err
	}

	return os.WriteFile(cleanPath, bytes, 0600)
}

// Convert converts a legacy config to a new config
func Convert(legacyCfg Config) config.Config {
	return config.Config{
		APIVersion:    modelversion.V1,
		Env:           string(legacyCfg.Server.Env),
		Output:        legacyCfg.Command.Output,
		Offline:       legacyCfg.Offline,
		Auth:          convertAuth(legacyCfg),
		Network:       convertNetwork(legacyCfg),
		AgentVersions: convertAgentVersions(legacyCfg),
		Store:         convertStore(legacyCfg),
		Tracing:       convertTracing(legacyCfg),
		Logging:       convertLogging(legacyCfg),
	}
}

// convertAuth converts a legacy config to a new auth config
func convertAuth(legacyCfg Config) config.Auth {
	return config.Auth{
		Username:      legacyCfg.Server.Username,
		Password:      legacyCfg.Server.Password,
		SecretKey:     legacyCfg.Server.SecretKey,
		SessionSecret: legacyCfg.SessionsSecret,
	}
}

// convertNetwork converts a legacy config to a new network config
func convertNetwork(legacyCfg Config) config.Network {
	tlsConfig := config.TLS{
		Certificate:          legacyCfg.Server.TLSConfig.Certificate,
		PrivateKey:           legacyCfg.Server.TLSConfig.PrivateKey,
		CertificateAuthority: legacyCfg.Server.TLSConfig.CertificateAuthority,
		InsecureSkipVerify:   legacyCfg.Server.TLSConfig.InsecureSkipVerify,
	}

	return config.Network{
		Host:      legacyCfg.Server.Host,
		Port:      legacyCfg.Server.Port,
		RemoteURL: legacyCfg.Server.ServerURL,
		TLS:       tlsConfig,
	}
}

// convertAgentVersions converts a legacy config to a new agent versions config
func convertAgentVersions(legacyCfg Config) config.AgentVersions {
	return config.AgentVersions{
		SyncInterval: legacyCfg.Server.SyncAgentVersionsInterval,
	}
}

// convertStore converts a legacy config to a new store config
func convertStore(legacyCfg Config) config.Store {
	return config.Store{
		Type: legacyCfg.Server.StoreType,
		BBolt: config.BBolt{
			Path: legacyCfg.Server.StorageFilePath,
		},
	}
}

// convertLogging converts a legacy config to a new logging config
func convertLogging(legacyCfg Config) config.Logging {
	return config.Logging{
		FilePath: legacyCfg.Server.LogFilePath,
		Output:   string(legacyCfg.Server.LogOutput),
	}
}

// convertTracing converts a legacy config to a new tracing config
func convertTracing(legacyCfg Config) config.Tracing {
	return config.Tracing{
		Type: legacyCfg.Server.TraceType,
		OTLP: config.OTLPTracing{
			Endpoint: legacyCfg.Server.OpenTelemetryTracing.Endpoint,
		},
		GoogleCloud: config.GoogleCloudTracing{
			ProjectID:       legacyCfg.Server.GoogleCloudTracing.ProjectID,
			CredentialsFile: legacyCfg.Server.GoogleCloudTracing.CredentialsFile,
		},
	}
}

// readLegacyCfg reads the legacy config in using the same strategy as before
func readLegacyCfg(cfgPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(cfgPath)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error while reading in legacy config: %w", err)
	}

	var legacyCfg Config
	err := v.Unmarshal(&legacyCfg, func(dc *mapstructure.DecoderConfig) {
		dc.Squash = true
	})

	if err != nil {
		return nil, fmt.Errorf("error while parsing legacy config: %w", err)
	}

	serverURL := legacyCfg.Server.BindPlaneURL()

	err = v.Unmarshal(&legacyCfg)
	if err != nil {
		return nil, fmt.Errorf("error while parsing legacy config: %w", err)
	}

	legacyCfg.Server.ServerURL = serverURL
	legacyCfg.Client.ServerURL = serverURL
	return &legacyCfg, nil
}

// getProfileName determines the profile name based on the configPath
func getProfileName(configPath string) string {
	profileBase := filepath.Base(configPath)
	return strings.TrimSuffix(profileBase, filepath.Ext(profileBase))
}
