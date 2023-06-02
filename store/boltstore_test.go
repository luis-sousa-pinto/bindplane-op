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

package store

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"

	"github.com/observiq/bindplane-op/model"
	"github.com/observiq/bindplane-op/model/search"
	"github.com/observiq/bindplane-op/store/storetest"
)

var testBuckets = []string{
	BucketResources,
	BucketAgents,
	BucketMeasurements,
	BucketArchive,
}

func TestBoltStoreClear(t *testing.T) {
	db, err := storetest.InitTestBboltDB(t, testBuckets)
	require.NoError(t, err, "error while initializing test database", err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := NewBoltStore(ctx, db, testOptions, zap.NewNop())
	s.ApplyResources(ctx, []model.Resource{
		macosSourceType,
		macosSource,
		cabinDestinationType,
		cabinDestination1,
		testRawConfiguration1,
	})

	s.Clear()

	sources, err := s.Sources(ctx)
	require.NoError(t, err)
	destinations, err := s.Destinations(ctx)
	require.NoError(t, err)
	configurations, err := s.Configurations(ctx)
	require.NoError(t, err)

	assert.Empty(t, sources)
	assert.Empty(t, destinations)
	assert.Empty(t, configurations)
}

func TestBoltStoreAddAgent(t *testing.T) {
	db, err := storetest.InitTestBboltDB(t, testBuckets)
	require.NoError(t, err, "error while initializing test database", err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := NewBoltStore(ctx, db, testOptions, zap.NewNop())
	a1 := &model.Agent{ID: "1", Name: "Fake Agent 1", Labels: model.Labels{Set: model.MakeLabels().Set}}
	a2 := &model.Agent{ID: "2", Name: "Fake Agent 2", Labels: model.Labels{Set: model.MakeLabels().Set}}

	err = addAgent(s, a1)
	require.NoError(t, err)
	err = addAgent(s, a2)
	require.NoError(t, err)

	var agents []*model.Agent

	db.View(func(tx *bbolt.Tx) error {
		c := tx.Bucket([]byte("Agents")).Cursor()

		prefix := []byte("Agent")
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			agent := &model.Agent{}
			jsoniter.Unmarshal(v, agent)
			agents = append(agents, agent)
		}
		return nil
	})

	assert.Len(t, agents, 2)
	assert.ElementsMatch(t, agents, []interface{}{a1, a2})
}

type mockUnknownResource struct {
	model.StatusType[any]
}

func (x mockUnknownResource) ID() string                                { return "" }
func (x mockUnknownResource) SetID(string)                              {}
func (x mockUnknownResource) EnsureID()                                 {}
func (x mockUnknownResource) Version() model.Version                    { return model.VersionLatest }
func (x mockUnknownResource) SetVersion(model.Version)                  {}
func (x mockUnknownResource) GetSpec() any                              { return nil }
func (x mockUnknownResource) EnsureHash(any)                            {}
func (x mockUnknownResource) Hash() string                              { return "" }
func (x mockUnknownResource) DateModified() *time.Time                  { return nil }
func (x mockUnknownResource) SetDateModified(*time.Time)                {}
func (x mockUnknownResource) EnsureMetadata(_ any)                      {}
func (x mockUnknownResource) GetKind() model.Kind                       { return model.KindUnknown }
func (x mockUnknownResource) Name() string                              { return "" }
func (x mockUnknownResource) Description() string                       { return "" }
func (x mockUnknownResource) Validate() (warnings string, errors error) { return "", nil }
func (x mockUnknownResource) ValidateWithStore(context.Context, model.ResourceStore) (warnings string, errors error) {
	return "", nil
}
func (x mockUnknownResource) UpdateDependencies(_ context.Context, _ model.ResourceStore) error {
	return nil
}
func (x mockUnknownResource) GetLabels() model.Labels  { return model.MakeLabels() }
func (x mockUnknownResource) SetLabels(_ model.Labels) {}
func (x mockUnknownResource) UniqueKey() string        { return x.ID() }

func (x mockUnknownResource) IndexID() string                   { return "" }
func (x mockUnknownResource) IndexFields(_ search.Indexer)      {}
func (x mockUnknownResource) IndexLabels(_ search.Indexer)      {}
func (x mockUnknownResource) PrintableKindSingular() string     { return string(model.KindUnknown) }
func (x mockUnknownResource) PrintableKindPlural() string       { return string(model.KindUnknown) }
func (x mockUnknownResource) PrintableFieldTitles() []string    { return []string{} }
func (x mockUnknownResource) PrintableFieldValue(string) string { return "-" }

var _ model.Resource = (*mockUnknownResource)(nil)

func TestKeyFromResource(t *testing.T) {
	cases := []struct {
		name     string
		resource model.Resource
		expect   string
	}{
		{
			"source",
			model.NewSourceType("test", []model.ParameterDefinition{}, []string{"macos", "linux", "windows"}),
			"SourceType|test",
		},
		{
			"destination",
			model.NewDestinationType("test", []model.ParameterDefinition{}),
			"DestinationType|test",
		},
		{
			"nil",
			nil,
			"",
		},
		{
			"unknown",
			&mockUnknownResource{},
			"",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output := string(KeyFromResource(tc.resource))
			require.Equal(t, tc.expect, output)
		})
	}
}

func TestBoltStoreConfigurations(t *testing.T) {
	db, err := storetest.InitTestBboltDB(t, testBuckets)
	require.NoError(t, err, "error while initializing test database", err)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := NewBoltStore(ctx, db, testOptions, zap.NewNop())

	runConfigurationsTests(t, s)
}

func TestBoltstoreConfiguration(t *testing.T) {
	db, err := storetest.InitTestBboltDB(t, testBuckets)
	require.NoError(t, err, "error while initializing test database", err)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := NewBoltStore(ctx, db, testOptions, zap.NewNop())

	runConfigurationTests(t, s)
}
func TestAgents(t *testing.T) {
	db, err := storetest.InitTestBboltDB(t, testBuckets)
	require.NoError(t, err, "error while initializing test database", err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := NewBoltStore(ctx, db, testOptions, zap.NewNop())
	a1 := &model.Agent{ID: "1", Name: "Fake Agent 1", Labels: model.Labels{Set: model.MakeLabels().Set}}
	a2 := &model.Agent{ID: "2", Name: "Fake Agent 2", Labels: model.Labels{Set: model.MakeLabels().Set}}

	addAgent(s, a1)
	addAgent(s, a2)

	agents, err := s.Agents(context.TODO())
	assert.NoError(t, err)
	assert.Len(t, agents, 2)
	assert.ElementsMatch(t, agents, []interface{}{a1, a2})
}

func TestAgent(t *testing.T) {
	db, err := storetest.InitTestBboltDB(t, testBuckets)
	require.NoError(t, err, "error while initializing test database", err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := NewBoltStore(ctx, db, testOptions, zap.NewNop())
	a1 := &model.Agent{ID: "1", Name: "Fake Agent 1", Labels: model.Labels{Set: model.MakeLabels().Set}}
	a2 := &model.Agent{ID: "2", Name: "Fake Agent 2", Labels: model.Labels{Set: model.MakeLabels().Set}}

	addAgent(s, a1)
	addAgent(s, a2)

	agent, err := s.Agent(ctx, a1.ID)
	assert.NoError(t, err)
	assert.Equal(t, a1, agent)
}

var updaterCalled bool

func testUpdater(agent *model.Agent) {
	updaterCalled = true
	agent.Name = "updated"
}

func TestUpsertAgent(t *testing.T) {
	db, err := storetest.InitTestBboltDB(t, testBuckets)
	require.NoError(t, err, "error while initializing test database", err)

	// Seed with one
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := NewBoltStore(ctx, db, testOptions, zap.NewNop())
	a1 := &model.Agent{ID: "1", Name: "Fake Agent 1", Labels: model.Labels{Set: model.MakeLabels().Set}}
	addAgent(s, a1)

	t.Run("creates a new agent if not found", func(t *testing.T) {
		newAgentID := "3"
		s.UpsertAgent(ctx, newAgentID, testUpdater)

		got, err := s.Agent(ctx, newAgentID)
		require.NoError(t, err)

		assert.NotNil(t, got)
		assert.Equal(t, got.ID, newAgentID)
	})
	t.Run("calls updater and updates an agent if exists", func(t *testing.T) {
		updaterCalled = false
		s.UpsertAgent(context.TODO(), a1.ID, testUpdater)

		assert.True(t, updaterCalled)

		got, err := s.Agent(ctx, a1.ID)
		require.NoError(t, err)

		assert.Equal(t, got.Name, "updated")
	})
}
func TestBoltStoreNotifyUpdates(t *testing.T) {
	db, err := storetest.InitTestBboltDB(t, testBuckets)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := NewBoltStore(ctx, db, testOptions, zap.NewNop())
	defer store.Close()
	done := make(chan bool, 1)

	runNotifyUpdatesTests(t, store, done)
}
func TestAgentConfigurationTests(t *testing.T) {
	db, err := storetest.InitTestBboltDB(t, testBuckets)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := NewBoltStore(ctx, db, testOptions, zap.NewNop())
	defer store.Close()

	runAgentConfigurationTests(ctx, t, store, func(s Store) {})
}

func TestBoltStoreDeleteChannel(t *testing.T) {
	db, err := storetest.InitTestBboltDB(t, testBuckets)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := NewBoltStore(ctx, db, testOptions, zap.NewNop())
	defer store.Close()
	done := make(chan bool, 1)

	runDeleteChannelTests(t, store, done)
}

func TestBoltStoreAgentSubscriptionChannel(t *testing.T) {
	db, err := storetest.InitTestBboltDB(t, testBuckets)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := NewBoltStore(ctx, db, testOptions, zap.NewNop())
	defer store.Close()
	runAgentSubscriptionsTest(t, store)
}

func TestBoltStoreAgentUpdatesChannel(t *testing.T) {
	db, err := storetest.InitTestBboltDB(t, testBuckets)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := NewBoltStore(ctx, db, testOptions, zap.NewNop())
	defer store.Close()

	runUpdateAgentsTests(t, store)
}

func TestBoltstoreApplyResourceReturn(t *testing.T) {
	db, err := storetest.InitTestBboltDB(t, testBuckets)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := NewBoltStore(ctx, db, testOptions, zap.NewNop())
	defer store.Close()
	runApplyResourceReturnTests(t, store)
}

func TestBoltstoreDeleteResourcesReturn(t *testing.T) {
	db, err := storetest.InitTestBboltDB(t, testBuckets)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := NewBoltStore(ctx, db, testOptions, zap.NewNop())
	defer store.Close()
	runDeleteResourcesReturnTests(t, store)
}

func TestBoltstoreValidateApplyResourcesTests(t *testing.T) {
	db, err := storetest.InitTestBboltDB(t, testBuckets)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := NewBoltStore(ctx, db, testOptions, zap.NewNop())
	defer store.Close()
	runValidateApplyResourcesTests(t, store)
}

func TestInitDB(t *testing.T) {
	cases := []struct {
		name      string
		setupFunc func() (string, error)
		errStr    string
	}{
		{
			"valid_path",
			func() (string, error) {
				return ioutil.TempDir("./", "tmp_store_test")
			},
			"",
		},
		{
			"invalid_path",
			func() (string, error) {
				return "not/valid/path", nil
			},
			"error while opening bbolt storage file: not/valid/path/bindplane.db, open not/valid/path/bindplane.db: no such file or directory",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup the test directory
			dir, err := tc.setupFunc()
			require.NoError(t, err, "failed to initialize test case")
			defer os.RemoveAll(dir)

			// Begin test
			path := filepath.Join(dir, "bindplane.db")
			db, err := InitBoltstoreDB(path)
			if tc.errStr != "" {
				require.Error(t, err)
				require.Equal(t, tc.errStr, err.Error())
				return
			}
			require.NoError(t, err, "did not expect an error while creating database at path %s", path)
			require.NotNil(t, db)
			require.Equal(t, path, db.Path())
			require.Equal(t, fmt.Sprintf("DB<\"%s\">", path), db.String())
			require.False(t, db.IsReadOnly(), "expected the boltstore to be read write")
			require.NoError(t, db.Close())

			// 1. root bucket
			// 2. agents
			// 3. resources
			// 4. archive
			// 5. measurements
			// 6. - otelcol_processor_throughputmeasurement_log_data_size
			// 7. - otelcol_processor_throughputmeasurement_metric_data_size
			// 8. - otelcol_processor_throughputmeasurement_trace_data_size
			bucketCount := 8
			require.Equal(t, bucketCount*2, db.Stats().TxStats.CursorCount)

			// InitDB creates buckets: Resources, Tasks, Agents, Measurements, and sub-buckets in measurements for each metric
			_ = db.Update(func(tx *bbolt.Tx) error {
				for _, bucket := range []string{BucketResources, BucketAgents, BucketMeasurements, BucketArchive} {
					// Deleting the bucket
					err := tx.DeleteBucket([]byte(bucket))
					require.NoError(t, err, "expected bucket %s to exist", bucket)
				}
				return nil
			})
		})
	}
}

func TestNewBoltStore(t *testing.T) {
	cases := []struct {
		name string
	}{
		{
			"valid",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup bbolt
			dir, err := ioutil.TempDir("./", "tmp_store_test")
			require.NoError(t, err, "failed to initialize test directory")
			defer os.RemoveAll(dir)
			db, err := InitBoltstoreDB(filepath.Join(dir, "bindplane.db"))
			require.NoError(t, err, "failed to initialize test bbolt, got error")
			require.NotNil(t, db, "failed to initialize test bbolt, is nil")

			// Test
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			output := NewBoltStore(ctx, db, testOptions, zap.NewNop())
			require.NotNil(t, output)
			require.IsType(t, &boltstore{}, output)
			require.Equal(t, db, output.(*boltstore).DB)
			require.Equal(t, 0, output.Updates(ctx).Subscribers())
		})
	}
}
func TestBucketNames(t *testing.T) {
	require.Equal(t, "Archive", BucketArchive)
	require.Equal(t, "Resources", BucketResources)
	require.Equal(t, "Agents", BucketAgents)
}

func TestBoltstoreDependentResources(t *testing.T) {
	db, err := storetest.InitTestBboltDB(t, testBuckets)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := NewBoltStore(ctx, db, testOptions, zap.NewNop())
	defer store.Close()
	runDependentResourcesTests(t, store)
}

func TestBoltstoreIndividualDelete(t *testing.T) {
	db, err := storetest.InitTestBboltDB(t, testBuckets)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := NewBoltStore(ctx, db, testOptions, zap.NewNop())
	defer store.Close()
	runIndividualDeleteTests(t, store)
}

func TestBoltstorePaging(t *testing.T) {
	db, err := storetest.InitTestBboltDB(t, testBuckets)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := NewBoltStore(ctx, db, testOptions, zap.NewNop())
	defer store.Close()
	runPagingTests(t, store)
}

func TestBoltStoreDeleteAgents(t *testing.T) {
	db, err := storetest.InitTestBboltDB(t, testBuckets)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := NewBoltStore(ctx, db, testOptions, zap.NewNop())
	defer store.Close()
	runDeleteAgentsTests(t, store)
}

func TestBoltstoreArchiveKey(t *testing.T) {
	tests := []struct {
		version model.Version
		expect  string
	}{
		{
			version: 1,
			expect:  "Configuration|c|000001",
		},
		{
			version: 99999999999999,
			expect:  "Configuration|c|99999999999999",
		},
		{
			version: 123,
			expect:  "Configuration|c|000123",
		},
		{
			version: 0,
			expect:  "Configuration|c|000000",
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("version=%d", test.version), func(t *testing.T) {
			require.Equal(t, test.expect, string(archiveKey(model.KindConfiguration, "c", test.version)))
		})
	}
}

func TestBoltstoreArchive(t *testing.T) {
	db, err := storetest.InitTestBboltDB(t, testBuckets)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := NewBoltStore(ctx, db, testOptions, zap.NewNop())
	defer store.Close()
	runTestArchive(ctx, t, store)
}

func TestBoltstoreUpsertAgents(t *testing.T) {
	db, err := storetest.InitTestBboltDB(t, testBuckets)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := NewBoltStore(ctx, db, testOptions, zap.NewNop())
	defer store.Close()
	runTestUpsertAgents(t, store)
}

func TestBoltstoreMeasurements(t *testing.T) {
	db, err := storetest.InitTestBboltDB(t, testBuckets)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := NewBoltStore(ctx, db, testOptions, zap.NewNop())
	defer store.Close()
	runTestMeasurements(t, store)
}

func TestCleanupDisconnectedAgents(t *testing.T) {
	db, err := storetest.InitTestBboltDB(t, testBuckets)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := NewBoltStore(ctx, db, testOptions, zap.NewNop())
	defer store.Close()
	runTestCleanupDisconnectedAgents(t, store)
}
func TestCountAgents(t *testing.T) {
	db, err := storetest.InitTestBboltDB(t, testBuckets)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := NewBoltStore(ctx, db, testOptions, zap.NewNop())
	defer store.Close()

	runTestCountAgents(ctx, t, store)
}

func TestStatus(t *testing.T) {
	db, err := storetest.InitTestBboltDB(t, testBuckets)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := NewBoltStore(ctx, db, testOptions, zap.NewNop())
	defer store.Close()

	runTestStatus(ctx, t, store)
}

func TestUpdateRollout(t *testing.T) {
	db, err := storetest.InitTestBboltDB(t, testBuckets)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := NewBoltStore(ctx, db, testOptions, zap.NewNop())
	defer store.Close()

	testUpdateRollout(ctx, t, store)
}

func TestConfigurationVersions(t *testing.T) {
	db, err := storetest.InitTestBboltDB(t, testBuckets)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := NewBoltStore(ctx, db, testOptions, zap.NewNop())
	defer store.Close()

	runTestConfigurationVersions(ctx, t, store)
}

func TestUpdateRollouts(t *testing.T) {
	db, err := storetest.InitTestBboltDB(t, testBuckets)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := NewBoltStore(ctx, db, testOptions, zap.NewNop())
	defer store.Close()

	runTestUpdateRollouts(ctx, t, store)
}

func TestResumeErroredRollout(t *testing.T) {
	db, err := storetest.InitTestBboltDB(t, testBuckets)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := NewBoltStore(ctx, db, testOptions, zap.NewNop())
	defer store.Close()

	testResumeErroredRollout(ctx, t, store)
}

func TestStartRollout(t *testing.T) {
	db, err := storetest.InitTestBboltDB(t, testBuckets)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := NewBoltStore(ctx, db, testOptions, zap.NewNop())
	defer store.Close()

	testStartRollout(ctx, t, store)
}

func TestDependencyUpdates(t *testing.T) {
	db, err := storetest.InitTestBboltDB(t, testBuckets)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := NewBoltStore(ctx, db, testOptions, zap.NewNop())
	defer store.Close()

	runTestDependencyUpdates(ctx, t, store, func(t *testing.T) {
		store.Clear()
	})
}
func TestCurrentRolloutsForConfiguration(t *testing.T) {
	db, err := storetest.InitTestBboltDB(t, testBuckets)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := NewBoltStore(ctx, db, testOptions, zap.NewNop())
	defer store.Close()

	runTestCurrentRolloutsForConfiguration(ctx, t, store)
}

/* ------------------------ SETUP + HELPER FUNCTIONS ------------------------ */

var testOptions = Options{
	SessionsSecret:   "super-secret-key",
	MaxEventsToMerge: 1,
}
