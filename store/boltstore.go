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
	"errors"
	"fmt"
	"time"

	"github.com/gorilla/sessions"
	"github.com/hashicorp/go-multierror"
	jsoniter "github.com/json-iterator/go"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"

	"github.com/observiq/bindplane-op/model"
	"github.com/observiq/bindplane-op/store/search"
	"github.com/observiq/bindplane-op/store/stats"
)

// bucket names
const (
	BucketResources    = "Resources"
	BucketAgents       = "Agents"
	BucketMeasurements = "Measurements"
	BucketArchive      = "Archive"
)

type boltstore struct {
	agentIndex         search.Index
	configurationIndex search.Index

	*BoltstoreCore
}

var _ Store = (*boltstore)(nil)
var _ ArchiveStore = (*boltstore)(nil)

// NewBoltStore returns a new store boltstore struct that implements the store.Store interface.
func NewBoltStore(ctx context.Context, db *bbolt.DB, options Options, logger *zap.Logger) Store {
	store := &boltstore{
		agentIndex:         search.NewInMemoryIndex("agent"),
		configurationIndex: search.NewInMemoryIndex("configuration"),
		BoltstoreCore: &BoltstoreCore{
			DB:             db,
			Logger:         logger,
			RolloutBatcher: NewNopRolloutBatcher(),
			SessionStorage: NewBPCookieStore(options.SessionsSecret),
		},
	}

	// There is a cyclic dependency here that's not great where the rollout batcher needs the store and the updates need the rollout batcher.
	if !options.DisableRolloutUpdater {
		// Assign a real batcher if we are not disabling rollout updater
		store.RolloutBatcher = NewDefaultBatcher(ctx, logger, DefaultRolloutBatchFlushInterval, store)
	}
	store.StoreUpdates = NewUpdates(ctx, options, logger, store.RolloutBatcher, BuildBasicEventBroadcast())

	// it might seem unintuitive, but it's important to point the boltstoreCommon interface to the store
	store.BoltstoreCommon = store

	// boltstore is not used for clusters, disconnect all agents
	store.disconnectAllAgents(ctx)

	// start the timer that runs cleanup on measurements
	if !options.DisableMeasurementsCleanup {
		// start the timer that runs cleanup on measurements
		store.StartMeasurements(ctx)
	}
	SeedSearchIndexes(ctx, store, logger)

	return store
}

// InitBoltstoreDB takes in the full path to a storage file and returns an opened bbolt database.
// It will return an error if the file cannot be opened.
func InitBoltstoreDB(storageFilePath string) (*bbolt.DB, error) {
	var db, err = bbolt.Open(storageFilePath, 0640, nil)
	if err != nil {
		return nil, fmt.Errorf("error while opening bbolt storage file: %s, %w", storageFilePath, err)
	}

	buckets := []string{
		BucketResources,
		BucketAgents,
		BucketMeasurements,
		BucketArchive,
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		for _, bucket := range buckets {
			_, _ = tx.CreateBucketIfNotExists([]byte(bucket))
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create bbolt storage bucket: %w", err)
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BucketMeasurements))
		for _, metric := range stats.SupportedMetricNames {
			_, _ = b.CreateBucketIfNotExists([]byte(metric))
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create bbolt metrics bucket: %w", err)
	}

	return db, nil
}

func (s *boltstore) Close() error {
	var errs error
	s.StoreUpdates.Shutdown(context.Background())
	if err := s.RolloutBatcher.Shutdown(context.Background()); err != nil {
		errs = errors.Join(errs, fmt.Errorf("failed to shutdown rollout batcher: %w", err))
	}
	if err := s.DB.Close(); err != nil {
		errs = errors.Join(errs, fmt.Errorf("failed to shutdown DB: %w", err))
	}

	return errs
}

// Apply resources iterates through a slice of resources, then adds them to storage,
// and calls notify updates on the updated resources.
func (s *boltstore) ApplyResources(ctx context.Context, resources []model.Resource) ([]model.ResourceStatus, error) {
	updates, ctx, shouldNotify := UpdatesForContext(ctx)

	// resourceStatuses to return for the applied resources
	resourceStatuses := make([]model.ResourceStatus, 0)
	var errs error
	for _, resource := range resources {
		// Set the resource's initial ID, which wil be overwritten if
		// the resource already exists (using the existing resource ID)
		resource.EnsureID()
		warn, err := resource.ValidateWithStore(ctx, s)
		if err != nil {
			resourceStatuses = append(resourceStatuses, *model.NewResourceStatusWithReason(resource, model.StatusInvalid, err.Error()))
			continue
		}
		err = s.DB.Update(func(tx *bbolt.Tx) error {
			// update the resource in the database
			status, err := UpsertResource(ctx, s, tx, resource)
			if err != nil {
				resourceStatuses = append(resourceStatuses, *model.NewResourceStatusWithReason(resource, model.StatusError, err.Error()))
				return err
			}
			resourceStatuses = append(resourceStatuses, *model.NewResourceStatusWithReason(resource, status, warn))
			switch status {
			case model.StatusCreated:
				updates.IncludeResource(resource, EventTypeInsert)
			case model.StatusConfigured:
				updates.IncludeResource(resource, EventTypeUpdate)
			}
			// some resources need special treatment
			switch r := resource.(type) {
			case *model.Configuration:
				// update the index
				err = s.configurationIndex.Upsert(ctx, r)
				if err != nil {
					s.Logger.Error("failed to update the search index", zap.String("configuration", r.Name()))
				}
			}
			return nil
		})
		if err != nil {
			errs = errors.Join(errs, err)
		}
	}

	if shouldNotify {
		// notify the updates and Create TransitiveUpdates resource updates
		s.Notify(ctx, updates)
		// update dependencies and accumulate any new statuses and errors
		resourceStatuses, errs = UpdateDependentResources(ctx, s, updates.TransitiveUpdates(), resourceStatuses, errs)
	}

	return resourceStatuses, errs
}

// ----------------------------------------------------------------------
// SystemStore

// UpdateAllRollouts updates all active rollouts.
func (s *boltstore) UpdateAllRollouts(ctx context.Context) error {
	_, err := s.BoltstoreCore.UpdateRollouts(ctx)
	return err
}

// addTransitiveUpdates adds all the transitive updates based on the resource type.
func (s *boltstore) addTransitiveUpdates(ctx context.Context, updates BasicEventUpdates) error {
	if updates.CouldAffectProcessors() {
		processors, err := s.Processors(ctx)
		if err != nil {
			return fmt.Errorf("failed to get processors: %w", err)
		}

		updates.AddAffectedProcessors(processors)
	}

	if updates.CouldAffectSources() {
		sources, err := s.Sources(ctx)
		if err != nil {
			return fmt.Errorf("failed to get sources: %w", err)
		}

		updates.AddAffectedSources(sources)
	}

	if updates.CouldAffectDestinations() {
		destinations, err := s.Destinations(ctx)
		if err != nil {
			return fmt.Errorf("failed to get destinations: %w", err)
		}

		updates.AddAffectedDestinations(destinations)
	}

	if updates.CouldAffectConfigurations() {
		configurations, err := s.Configurations(ctx)
		if err != nil {
			return fmt.Errorf("failed to get configurations: %w", err)
		}

		updates.AddAffectedConfigurations(configurations)
	}

	return nil
}

// ----------------------------------------------------------------------

func (s *boltstore) Notify(ctx context.Context, updates BasicEventUpdates) {
	ctx, span := tracer.Start(ctx, "store/notify")
	defer span.End()

	err := s.addTransitiveUpdates(ctx, updates)
	if err != nil {
		// TODO: if we can't notify about all updates, what do we do?
		s.Logger.Error("unable to add transitive updates", zap.Any("updates", updates), zap.Error(err))
	}
	if !updates.Empty() {
		s.StoreUpdates.Send(ctx, updates)
	}
}

func (s *boltstore) CreateEventUpdate() BasicEventUpdates {
	return NewEventUpdates()
}

// ----------------------------------------------------------------------

// Clear clears the db store of resources, agents, and tasks.  Mostly used for testing.
func (s *boltstore) Clear() {
	// Disregarding error from update because these actions errors are known and prevented
	_ = s.DB.Update(func(tx *bbolt.Tx) error {
		// Delete all the buckets.
		// Disregarding errors because it will only error if the bucket doesn't exist
		// or isn't a bucket key - which we're confident its not.
		_ = tx.DeleteBucket([]byte(BucketResources))
		_ = tx.DeleteBucket([]byte(BucketAgents))
		_ = tx.DeleteBucket([]byte(BucketMeasurements))
		_ = tx.DeleteBucket([]byte(BucketArchive))

		// create them again
		// Disregarding errors because bucket names are valid.
		_, _ = tx.CreateBucketIfNotExists([]byte(BucketResources))
		_, _ = tx.CreateBucketIfNotExists([]byte(BucketAgents))
		b, _ := tx.CreateBucketIfNotExists([]byte(BucketMeasurements))
		_, _ = tx.CreateBucketIfNotExists([]byte(BucketArchive))

		for _, metric := range stats.SupportedMetricNames {
			_, _ = b.CreateBucketIfNotExists([]byte(metric))
		}
		return nil
	})
}

func (s *boltstore) DeleteAgents(ctx context.Context, agentIDs []string) ([]*model.Agent, error) {
	updates := s.CreateEventUpdate()
	deleted := make([]*model.Agent, 0, len(agentIDs))

	err := s.DB.Update(func(tx *bbolt.Tx) error {
		bucket, err := s.AgentsBucket(ctx, tx)
		if err != nil {
			return err
		}
		c := bucket.Cursor()
		for _, id := range agentIDs {
			agentKey := AgentKey(id)
			k, v := c.Seek(agentKey)

			if k != nil && bytes.Equal(k, agentKey) {

				// Save the agent to return and set its status to deleted.
				agent := &model.Agent{}
				err := jsoniter.Unmarshal(v, agent)
				if err != nil {
					return err
				}

				agent.Status = model.Deleted
				deleted = append(deleted, agent)

				// delete it
				err = c.Delete()
				if err != nil {
					return err
				}

				// include it in updates
				updates.IncludeAgent(agent, EventTypeRemove)
			}
		}

		return nil
	})

	if err != nil {
		return deleted, err
	}

	// remove deleted agents from the index
	for _, agent := range deleted {
		if err := s.agentIndex.Remove(ctx, agent); err != nil {
			s.Logger.Error("failed to remove from the search index", zap.String("agentID", agent.ID))
		}
	}

	// notify updates
	s.Notify(ctx, updates)

	return deleted, nil
}

func (s *boltstore) UserSessions() sessions.Store {
	return s.SessionStorage
}

// Measurements stores stats for agents and configurations
func (s *boltstore) Measurements() stats.Measurements {
	return s
}

// ----------------------------------------------------------------------

func (s *boltstore) disconnectAllAgents(ctx context.Context) {
	if agents, err := s.Agents(ctx); err != nil {
		s.Logger.Error("error while disconnecting all agents on startup", zap.Error(err))
	} else {
		s.Logger.Info("disconnecting all agents on startup", zap.Int("count", len(agents)))
		for _, agent := range agents {
			_, err := s.UpdateAgent(ctx, agent.ID, func(a *model.Agent) {
				a.Disconnect()
			})
			if err != nil {
				s.Logger.Error("error while disconnecting agent on startup", zap.Error(err))
			}
		}
	}
}

/* ---------------------------- helper functions ---------------------------- */

// ResourcesPrefix returns the prefix for a resource kind in the store.
func ResourcesPrefix(kind model.Kind) []byte {
	return []byte(fmt.Sprintf("%s|", kind))
}

// ResourceKey returns the key for a resource in the store.
func ResourceKey(kind model.Kind, name string) []byte {
	return []byte(fmt.Sprintf("%s|%s", kind, name))
}

// AgentKey returns the key for an agent in the store.
func AgentKey(id string) []byte {
	return []byte(fmt.Sprintf("%s|%s", "Agent", id))
}

// AgentPrefix returns the prefix for agents in the store.
func AgentPrefix() []byte {
	return []byte("Agent|")
}

// KeyFromResource returns the key for a resource in the store.
func KeyFromResource(r model.Resource) []byte {
	if r == nil || r.GetKind() == model.KindUnknown {
		return make([]byte, 0)
	}
	return ResourceKey(r.GetKind(), r.UniqueKey())
}

/* --------------------------- transaction helpers -------------------------- */
/* ------- These helper functions happen inside of a bbolt transaction ------ */

// UpsertResource upserts a resource into the store.  If the resource already exists, it will be updated.
func UpsertResource(ctx context.Context, s BoltstoreCommon, tx *bbolt.Tx, r model.Resource) (model.UpdateStatus, error) {
	key := KeyFromResource(r)
	bucket, err := s.ResourcesBucket(ctx, tx, r.GetKind())
	if err != nil {
		// error, status unchanged
		return model.StatusUnchanged, fmt.Errorf("upsert resource: %w", err)
	}

	r.EnsureMetadata(r.GetSpec())

	existing := bucket.Get(key)
	hasExisting := len(existing) > 0
	var cur *model.AnyResource

	if hasExisting {
		// preserve some existing fields of the resource
		if err := jsoniter.Unmarshal(existing, &cur); err == nil {
			// special case for existing resources without Versions
			if cur.Version() > 0 {
				r.SetVersion(cur.Version())
			}

			// preserve the metadata for comparison, update later if needed
			r.SetID(cur.ID())
			r.SetDateModified(cur.DateModified())
			if err := r.SetStatus(cur.GetStatus()); err != nil {
				s.ZapLogger().Warn("upsert resource failed to preserve status", zap.Error(err))
			}
		}
	} else {
		// don't allow status to be set on new resources. it is managed by the system, not the user.
		_ = r.SetStatus(nil)
		r.SetVersion(model.Version(1))
	}

	status, err := storeResource(ctx, s, bucket, tx, existing, cur, r)
	if err != nil {
		// error, status unchanged
		return model.StatusUnchanged, fmt.Errorf("upsert resource, storeResource: %w", err)
	}

	return status, nil
}

// UpdateResource updates a resource in the store. If the resource does not exist, model.StatusUnchanged is returned
// with the error ErrResourceMissing. If the resource does exist, it is updated using the updater function and the
// updated resource is returned.
func UpdateResource[R model.Resource](ctx context.Context, s BoltstoreCommon, tx *bbolt.Tx, kind model.Kind, name string, updater func(R) error) (r R, status model.UpdateStatus, err error) {
	key := ResourceKey(kind, name)
	bucket, err := s.ResourcesBucket(ctx, tx, kind)
	if err != nil {
		// error, status unchanged
		return r, model.StatusUnchanged, fmt.Errorf("upsert resource: %w", err)
	}

	existing := bucket.Get(key)
	hasExisting := len(existing) > 0
	if !hasExisting {
		return r, model.StatusUnchanged, ErrStoreResourceMissing
	}

	// find the existing resource and unmarshal it
	if err := jsoniter.Unmarshal(existing, &r); err != nil {
		return r, model.StatusUnchanged, fmt.Errorf("upsert resource, unmarshal error: %w", err)
	}
	var cur *model.AnyResource
	if err := jsoniter.Unmarshal(existing, &cur); err != nil {
		return r, model.StatusUnchanged, fmt.Errorf("upsert resource, unmarshal error: %w", err)
	}

	// update the resource
	if err := updater(r); err != nil {
		// error, status unchanged
		return r, model.StatusUnchanged, fmt.Errorf("upsert resource, updater returned error: %w", err)
	}

	status, err = storeResource(ctx, s, bucket, tx, existing, cur, r)
	if err != nil {
		// error, status unchanged
		return r, model.StatusUnchanged, fmt.Errorf("upsert resource, storeResource: %w", err)
	}

	return r, status, nil
}

// updateOrUpsertAgentTx is a transaction helper that updates the given agent,
// puts it into the agent bucket and includes it in the passed updates.
// it does *not* update the search index or notify any subscribers of updates.
func (s *BoltstoreCore) updateOrUpsertAgentTx(ctx context.Context, requireExists bool, bucket *bbolt.Bucket, agentID string, updater AgentUpdater, updates BasicEventUpdates) (*model.Agent, error) {
	key := AgentKey(agentID)

	agentEventType := EventTypeInsert
	agent := &model.Agent{ID: agentID}

	// load the existing agent or create it
	dataBefore := bucket.Get(key)
	if dataBefore != nil {
		// existing agent, unmarshal
		if err := jsoniter.Unmarshal(dataBefore, agent); err != nil {
			return agent, err
		}
		agentEventType = EventTypeUpdate
	} else if requireExists {
		return nil, nil
	}

	// compare labels before/after and notify if they change
	labelsBefore := agent.Labels.String()
	pendingBefore := agent.ConfigurationStatus.Pending

	// update the agent
	updater(agent)

	labelsAfter := agent.Labels.String()
	pendingAfter := agent.ConfigurationStatus.Pending

	// if the labels changes is this is just an update, use EventTypeLabel
	if labelsAfter != labelsBefore && agentEventType == EventTypeUpdate {
		agentEventType = EventTypeLabel
		_, err := s.FindAgentConfiguration(ctx, agent)
		if err != nil {
			return agent, err
		}
	}

	// if Pending changes to a new configuration, use EventTypeRollout
	if pendingAfter != "" && pendingBefore != pendingAfter {
		agentEventType = EventTypeRollout
	}

	// marshal it back to to json
	dataAfter, err := jsoniter.Marshal(agent)
	if err != nil {
		return agent, err
	}

	// only write and include in updates if there are actual changes
	if !bytes.Equal(dataBefore, dataAfter) {
		err = bucket.Put(key, dataAfter)
		if err != nil {
			return agent, err
		}
		updates.IncludeAgent(agent, agentEventType)
	}

	return agent, nil
}

// ----------------------------------------------------------------------
// boltCommon

var _ BoltstoreCommon = (*boltstore)(nil)

func (s *boltstore) AgentsBucket(_ context.Context, tx *bbolt.Tx) (*bbolt.Bucket, error) {
	return tx.Bucket([]byte(BucketAgents)), nil
}
func (s *boltstore) MeasurementsBucket(_ context.Context, tx *bbolt.Tx, metric string) (*bbolt.Bucket, error) {
	b := tx.Bucket([]byte(BucketMeasurements))
	if b != nil {
		return b.Bucket([]byte(metric)), nil
	}
	return nil, nil
}
func (s *boltstore) ResourcesBucket(_ context.Context, tx *bbolt.Tx, _ model.Kind) (*bbolt.Bucket, error) {
	return tx.Bucket([]byte(BucketResources)), nil
}
func (s *boltstore) ArchiveBucket(_ context.Context, tx *bbolt.Tx) (*bbolt.Bucket, error) {
	return tx.Bucket([]byte(BucketArchive)), nil
}
func (s *boltstore) ResourceKey(r model.Resource) []byte {
	return KeyFromResource(r)
}
func (s *boltstore) AgentsIndex(_ context.Context) search.Index {
	return s.agentIndex
}
func (s *boltstore) ConfigurationsIndex(_ context.Context) search.Index {
	return s.configurationIndex
}
func (s *boltstore) ZapLogger() *zap.Logger {
	return s.Logger
}

// ----------------------------------------------------------------------
// generic resource accessors

// FindResource finds a resource by kind and unique key. If the resource is versioned, the latest version is returned.
func FindResource[R model.Resource](ctx context.Context, s BoltstoreCommon, tx *bbolt.Tx, kind model.Kind, uniqueKey string) (resource R, key []byte, bucket *bbolt.Bucket, exists bool, err error) {
	uniqueKey, version := model.SplitVersion(uniqueKey)

	// start with the latest version from the resources bucket
	key = ResourceKey(kind, uniqueKey)
	bucket, err = s.ResourcesBucket(ctx, tx, kind)
	if err != nil {
		return
	}
	data := bucket.Get(key)
	if data == nil {
		return
	}
	exists = true
	err = jsoniter.Unmarshal(data, &resource)
	if err != nil {
		return
	}

	MaskSensitiveParameters(ctx, resource)

	// if the resource isn't versioned, we're done
	if !model.HasVersionKind(kind) {
		return
	}

	// for now assume this is the latest version
	resource.SetLatest(true)
	if kind == model.KindConfiguration {
		// configurations are special and track pending and current versions in the latest version. since this is the latest
		// version, we can use the versions tracked here to determine pending and current for all versions.
		if status, ok := model.ParseResourceStatus[model.ConfigurationStatus](resource); ok {
			defer func() {
				resource.SetPending(resource.Version() == status.PendingVersion)
				resource.SetCurrent(resource.Version() == status.CurrentVersion)
			}()
		}
	}

	// if the Version is latest, we're done
	if version == model.VersionLatest {
		return
	}

	// if Version is current or pending, return the latest if this is not a configuration or check the configuration
	// status for the current or pending Version number
	if version == model.VersionCurrent || version == model.VersionPending {
		// configurations are a special case where current and latest may be different based on the progress of a rollout
		if kind != model.KindConfiguration {
			return
		}

		// if the Version is current, check the configuration status for the current Version
		if configuration, ok := any(resource).(*model.Configuration); ok {
			switch version {
			case model.VersionCurrent:
				version = configuration.Status.CurrentVersion
			case model.VersionPending:
				version = configuration.Status.PendingVersion
			}
			if version < 1 {
				// no current version is established, so no rollout is complete for current or started for pending
				return
			}
		}
	}

	// at this point this is a versioned resource and there is a specific Version >= 1 but it might coincidentally be
	// the Version that was retrieved
	if version == resource.Version() {
		return
	}

	// assume the archive version doesn't exist
	exists = false

	// check the archive bucket for a specific Version
	var archiveResource R
	key = archiveKey(kind, uniqueKey, version)
	bucket, err = s.ArchiveBucket(ctx, tx)
	if err != nil {
		return
	}
	data = bucket.Get(key)
	if data == nil {
		return
	}
	exists = true
	err = jsoniter.Unmarshal(data, &archiveResource)
	if err != nil {
		return
	}
	resource = archiveResource
	resource.SetLatest(false)
	MaskSensitiveParameters(ctx, resource)

	return
}

// Resource returns a resource of the given kind and unique key.
func Resource[R model.Resource](ctx context.Context, s BoltstoreCommon, kind model.Kind, uniqueKey string) (resource R, exists bool, err error) {
	err = s.Database().View(func(tx *bbolt.Tx) error {
		resource, _, _, exists, err = FindResource[R](ctx, s, tx, kind, uniqueKey)
		return err
	})
	return resource, exists, err
}

// Resources returns all resources of the given kind.
func Resources[R model.Resource](ctx context.Context, s BoltstoreCommon, kind model.Kind) ([]R, error) {
	return resourcesWithFilter[R](ctx, s, kind, nil)
}

func resourcesWithFilter[R model.Resource](ctx context.Context, s BoltstoreCommon, kind model.Kind, include func(R) bool) ([]R, error) {
	var resources []R

	setCurrentAndPending := func(resource R) {}
	if kind == model.KindConfiguration {
		setCurrentAndPending = func(resource R) {
			if status, ok := model.ParseResourceStatus[model.ConfigurationStatus](resource); ok {
				resource.SetCurrent(resource.Version() == status.CurrentVersion)
				resource.SetPending(resource.Version() == status.PendingVersion)
			}
		}
	}

	err := s.Database().View(func(tx *bbolt.Tx) error {
		prefix := ResourcesPrefix(kind)
		bucket, err := s.ResourcesBucket(ctx, tx, kind)
		if err != nil {
			return err
		}
		cursor := bucket.Cursor()

		for k, v := cursor.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = cursor.Next() {
			var resource R
			if err := jsoniter.Unmarshal(v, &resource); err != nil {
				// TODO(andy): if it can't be unmarshaled, it should probably be removed from the store. ignore it for now.
				s.ZapLogger().Error("failed to unmarshal resource", zap.String("key", string(k)), zap.String("kind", string(kind)), zap.Error(err))
				continue
			}

			// these are all the latest versions
			resource.SetLatest(true)
			setCurrentAndPending(resource)
			MaskSensitiveParameters(ctx, resource)

			if include == nil || include(resource) {
				resources = append(resources, resource)
			}
		}

		return nil
	})

	return resources, err
}

// ResourcesByUniqueKeys returns the resources of the specified kind with the specified uniqueKeys. If requesting some resources
// results in an error, the errors will be accumulated and return with the list of resources successfully retrieved.
func ResourcesByUniqueKeys[R model.Resource](ctx context.Context, s BoltstoreCommon, kind model.Kind, uniqueKeys []string, opts QueryOptions) ([]R, error) {
	var errs error
	var results []R

	for _, uniqueKey := range uniqueKeys {
		if result, exists, err := Resource[R](ctx, s, kind, uniqueKey); err != nil {
			errs = multierror.Append(errs, err)
		} else {
			if exists && opts.Selector.Matches(result.GetLabels()) {
				results = append(results, result)
			}
		}
	}
	return results, errs
}

// DeleteResourceAndNotify removes the resource with the given kind and uniqueKey. Returns ResourceMissingError if the resource wasn't found.
func DeleteResourceAndNotify[R model.Resource](ctx context.Context, s BoltstoreCommon, kind model.Kind, uniqueKey string, emptyResource R) (resource R, exists bool, err error) {
	deleted, exists, err := DeleteResource(ctx, s, kind, uniqueKey, emptyResource)

	if err == nil && exists {
		updates := s.CreateEventUpdate()
		updates.IncludeResource(deleted, EventTypeRemove)
		s.Notify(ctx, updates)
	}

	return deleted, exists, err
}

// DeleteResource removes the resource with the given kind and uniqueKey. Returns ResourceMissingError if the resource
// wasn't found. Returns DependencyError if the resource is referenced by another.
//
// emptyResource will be populated with the deleted resource. For convenience, if the delete is successful, the
// populated resource will also be returned. If there was an error, nil will be returned for the resource.
func DeleteResource[R model.Resource](ctx context.Context, s BoltstoreCommon, kind model.Kind, uniqueKey string, emptyResource R) (resource R, exists bool, err error) {
	var dependencies DependentResources

	err = s.Database().Update(func(tx *bbolt.Tx) error {
		key := ResourceKey(kind, uniqueKey)
		bucket, err := s.ResourcesBucket(ctx, tx, kind)
		if err != nil {
			return err
		}
		c := bucket.Cursor()
		k, v := c.Seek(key)

		if bytes.Equal(k, key) {
			// populate the emptyResource with the data before deleting
			err := jsoniter.Unmarshal(v, emptyResource)
			if err != nil {
				return err
			}

			exists = true
			MaskSensitiveParameters(ctx, emptyResource)

			// Check if the resource is referenced by another
			dependencies, err = FindDependentResources(ctx, s.ConfigurationsIndex(ctx), emptyResource.Name(), emptyResource.GetKind())
			if !dependencies.Empty() {
				return ErrResourceInUse
			}

			// Delete the key from the store
			err = c.Delete()
			if err != nil {
				return err
			}

			// also delete this archive of this resource
			prefix := archivePrefix(kind, uniqueKey)
			archiveBucket, err := s.ArchiveBucket(ctx, tx)
			if err != nil {
				return err
			}
			archiveCursor := archiveBucket.Cursor()
			for k, _ := archiveCursor.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = archiveCursor.Next() {
				err = archiveCursor.Delete()
				if err != nil {
					return err
				}
			}

			return nil
		}

		return ErrStoreResourceMissing
	})

	switch {
	case errors.Is(err, ErrStoreResourceMissing):
		return resource, exists, nil
	case errors.Is(err, ErrResourceInUse):
		return resource, exists, NewDependencyError(dependencies)
	case err != nil:
		return resource, exists, err
	}

	if emptyResource.GetKind() == model.KindConfiguration {
		if err := s.ConfigurationsIndex(ctx).Remove(ctx, emptyResource); err != nil {
			s.ZapLogger().Error("failed to remove configuration from the search index", zap.String("name", emptyResource.Name()))
		}
	}

	return emptyResource, exists, nil
}

// editResource edits a existing resource version stored in either the resources or archive bucket. It will not create a
// new version of a resource and is meant to only be used by the store when modifying the status of an existing
// resource. The updater function is called with the resource and should modify it in place. If the resource does not
// exist, ErrResourceMissing is returned. If tx is nil, a new transaction is created and committed.
func editResource[R model.Resource](ctx context.Context, s BoltstoreCommon, tx *bbolt.Tx, kind model.Kind, uniqueKey string, updater func(resource R) error) (resource R, wasModified bool, err error) {
	updateUsingTx := func(tx *bbolt.Tx) error {
		// find the existing resource
		var (
			key    []byte
			bucket *bbolt.Bucket
			exists bool
		)
		// don't mask sensitive parameters when updating a resource
		ctx = model.ContextWithoutSensitiveParameterMasking(ctx)
		resource, key, bucket, exists, err = FindResource[R](ctx, s, tx, kind, uniqueKey)
		if err != nil {
			return fmt.Errorf("error finding resource: %w", err)
		}
		if !exists {
			return ErrStoreResourceMissing
		}

		// update the date modified before comparing so we can ignore the impact of updating the date on the diff of the
		// resource.
		oldDateModified := resource.DateModified()
		now := time.Now()
		resource.SetDateModified(&now)

		// marshal before to compare after updater
		dataBefore, err := jsoniter.Marshal(resource)
		if err != nil {
			// existing resource should marshal, but if it doesn't, we will attempt to update it anyway
			dataBefore = []byte{}
		}

		// update the resource
		err = updater(resource)
		if err != nil {
			return fmt.Errorf("error updating resource: %w", err)
		}

		// store the updated resource
		data, err := jsoniter.Marshal(resource)
		if err != nil {
			return fmt.Errorf("error marshaling resource: %w", err)
		}

		// if the resource is the same, we're done
		if bytes.Equal(data, dataBefore) {
			// restore the old modified date
			resource.SetDateModified(oldDateModified)
			return nil
		}

		err = bucket.Put(key, data)
		if err != nil {
			return fmt.Errorf("error storing resource: %w", err)
		}

		wasModified = true
		return nil
	}
	// use the existing transaction if we have one
	if tx != nil {
		err = updateUsingTx(tx)
	} else {
		err = s.Database().Update(updateUsingTx)
	}

	if err != nil {
		return resource, false, err
	}

	// only re-index if modified
	if wasModified {
		if resource.GetKind() == model.KindConfiguration {
			if err := s.ConfigurationsIndex(ctx).Upsert(ctx, resource); err != nil {
				return resource, wasModified, err
			}
		}
	}

	return resource, wasModified, nil
}

// storeResource handles storing a resource and archiving the existing resource if it is versioned.
func storeResource(ctx context.Context, s BoltstoreCommon, bucket *bbolt.Bucket, tx *bbolt.Tx, curBytes []byte, curResource *model.AnyResource, newResource model.Resource) (model.UpdateStatus, error) {
	if curResource != nil {
		// preserve sensitive parameter values
		if err := PreserveSensitiveParameters(ctx, newResource, curResource); err != nil {
			return model.StatusUnchanged, fmt.Errorf("store resource: %w", err)
		}

		// check to see if the resource has changed
		compare, err := jsoniter.Marshal(newResource)
		if err != nil {
			// error, status unchanged
			return model.StatusUnchanged, fmt.Errorf("upsert resource: %w", err)
		}
		if bytes.Equal(curBytes, compare) {
			return model.StatusUnchanged, nil
		}

		newConfigurationVersion, newConfiguration, err := IsNewConfigurationVersion(curResource, newResource)
		if err != nil {
			return model.StatusUnchanged, err
		}

		// archive the existing resource if it is versioned
		if model.IsVersionedKind(newResource.GetKind()) || newConfigurationVersion {
			// special case for existing resources without Versions
			if curResource.Version() == 0 {
				curResource.SetVersion(1)
			}

			// marshal before attempting to archive
			existing, err := jsoniter.Marshal(curResource)
			if err != nil {
				// error, status unchanged
				return model.StatusUnchanged, fmt.Errorf("upsert resource: %w", err)
			}

			if err := archiveResource(ctx, s, tx, newResource, existing); err != nil {
				// unable to save the archive, fail
				return model.StatusUnchanged, fmt.Errorf("archive resource: %w", err)
			}

			// increment the Version
			newResource.SetVersion(curResource.Version() + 1)

			// also reset the rollout status if this is a new version
			if newConfigurationVersion {
				newConfiguration.Status.Rollout = model.Rollout{}
				newResource = newConfiguration
			}
		}
	}

	// update the date modified because the resource is new or has changed
	now := time.Now()
	newResource.SetDateModified(&now)
	newResource.SetLatest(true)

	data, err := jsoniter.Marshal(newResource)
	if err != nil {
		// error, status unchanged
		return model.StatusUnchanged, fmt.Errorf("upsert resource: %w", err)
	}

	key := KeyFromResource(newResource)
	if err = bucket.Put(key, data); err != nil {
		// error, status unchanged
		return model.StatusUnchanged, fmt.Errorf("upsert resource: %w", err)
	}

	if len(curBytes) == 0 {
		return model.StatusCreated, nil
	}

	return model.StatusConfigured, nil
}

func archiveResource(ctx context.Context, s BoltstoreCommon, tx *bbolt.Tx, r model.Resource, data []byte) error {
	key := archiveKeyFromResource(r)
	bucket, err := s.ArchiveBucket(ctx, tx)
	if err != nil || bucket == nil {
		return fmt.Errorf("archive resource: %w", err)
	}
	return bucket.Put(key, data)
}

func resourceHistory[R model.Resource](ctx context.Context, s BoltstoreCommon, kind model.Kind, uniqueKey string) ([]R, error) {
	var history []R

	err := s.Database().View(func(tx *bbolt.Tx) error {
		uniqueKey, _ := model.SplitVersion(uniqueKey)

		// start with the latest version from the resources bucket
		key := ResourceKey(kind, uniqueKey)
		bucket, err := s.ResourcesBucket(ctx, tx, kind)
		if err != nil {
			return err
		}
		data := bucket.Get(key)
		if data == nil {
			return nil
		}
		var resource R
		err = jsoniter.Unmarshal(data, &resource)
		if err != nil {
			return err
		}

		// if the resource isn't versioned, we're done
		if !model.HasVersionKind(kind) {
			return nil
		}

		resource.SetLatest(true)

		// assume this is a noop
		setCurrentAndPending := func(resource R) {}
		if kind == model.KindConfiguration {
			if status, ok := model.ParseResourceStatus[model.ConfigurationStatus](resource); ok {
				currentVersion := status.CurrentVersion
				pendingVersion := status.PendingVersion
				setCurrentAndPending = func(resource R) {
					resource.SetCurrent(resource.Version() == currentVersion)
					resource.SetPending(resource.Version() == pendingVersion)
				}
			}
		}
		setCurrentAndPending(resource)
		MaskSensitiveParameters(ctx, resource)

		// check the archive bucket for older Versions
		bucket, err = s.ArchiveBucket(ctx, tx)
		if err != nil {
			return err
		}
		prefix := archivePrefix(kind, uniqueKey)
		cursor := bucket.Cursor()
		for k, v := cursor.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = cursor.Next() {
			var archiveResource R
			err = jsoniter.Unmarshal(v, &archiveResource)
			if err != nil {
				return err
			}
			archiveResource.SetLatest(false)
			setCurrentAndPending(archiveResource)
			MaskSensitiveParameters(ctx, archiveResource)
			history = append(history, archiveResource)
		}

		// add the latest version last
		history = append(history, resource)

		return nil
	})

	// reverse the slice with newest Version first
	for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
		history[i], history[j] = history[j], history[i]
	}

	return history, err
}
