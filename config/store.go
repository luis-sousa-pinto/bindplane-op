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
	"fmt"
	"path/filepath"

	"github.com/observiq/bindplane-op/common"
)

const (
	// DefaultMaxEvents is the default maximum number of events to merge into a single event.
	DefaultMaxEvents = 100

	// StoreTypeMap is the type of store that uses an in-memory store.
	StoreTypeMap = "map"

	// StoreTypeBBolt is the type of store that uses bbolt.
	StoreTypeBBolt = "bbolt"
)

// DefaultBBoltPath is the default path to the bbolt file.
var DefaultBBoltPath = filepath.Join(common.GetHome(), "storage")

// Store is the configuration for a store.
type Store struct {
	// Type is the type of store.
	Type string `mapstructure:"type,omitempty" yaml:"type,omitempty"`

	// MaxEvents is the maximum number of events to merge into a single event.
	MaxEvents int `mapstructure:"maxEvents,omitempty" yaml:"maxEvents,omitempty"`

	// BBolt is the configuration for a bbolt store.
	BBolt BBolt `mapstructure:"bbolt,omitempty" yaml:"bbolt,omitempty"`
}

// Validate validates the store configuration.
func (s *Store) Validate() error {
	switch s.Type {
	case StoreTypeMap:
	case StoreTypeBBolt:
		if err := s.BBolt.Validate(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid store type: %s", s.Type)
	}

	if s.MaxEvents < 1 {
		return fmt.Errorf("maxEvents must be greater than 0")
	}

	return nil
}

// BBolt is the configuration for a bbolt store.
type BBolt struct {
	// Path is the path to the bbolt file.
	Path string `mapstructure:"path,omitempty" yaml:"path,omitempty"`
}

// Validate validates the bbolt configuration.
func (b *BBolt) Validate() error {
	if b.Path == "" {
		return fmt.Errorf("bbolt path must be set for bbolt store")
	}
	return nil
}
