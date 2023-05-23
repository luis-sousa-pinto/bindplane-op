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

// Package storetest has helper functions for setting up storage for tests
package storetest

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
)

// InitTestBboltDB creates a new Bbolt DB for testing
func InitTestBboltDB(t *testing.T, buckets []string) (*bbolt.DB, error) {
	t.Helper()
	tmpDir := t.TempDir()
	storageFile := filepath.Join(tmpDir, fmt.Sprintf("test-storage-%s", strings.ReplaceAll(t.Name(), "/", "-")))

	db, err := bbolt.Open(storageFile, 0666, nil)
	require.NoError(t, err, "error while opening test database", err)

	// make sure buckets exists
	return db, db.Update(func(tx *bbolt.Tx) error {
		for _, bucket := range buckets {
			_, err := tx.CreateBucketIfNotExists([]byte(bucket))
			require.NoError(t, err, "error while initializing test database, %w", err)
		}

		return nil
	})
}
