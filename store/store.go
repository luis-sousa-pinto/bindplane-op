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

// Package store contains the Store interface and implementations
package store

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"github.com/hashicorp/go-multierror"
	"github.com/observiq/bindplane-op/eventbus"
	"github.com/observiq/bindplane-op/model"
	modelSearch "github.com/observiq/bindplane-op/model/search"
	"github.com/observiq/bindplane-op/store/search"
	"github.com/observiq/bindplane-op/store/stats"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

var tracer = otel.Tracer("store")

// Options are options that are common to all store implementations
type Options struct {
	// SessionsSecret is used to encode sessions
	SessionsSecret string
	// MaxEventsToMerge is the maximum number of update events (inserts, updates, deletes, etc) to merge into a single
	// event.
	MaxEventsToMerge int
	// DisableMeasurementsCleanup indicates that the store should not clean up measurements. This is useful for testing.
	DisableMeasurementsCleanup bool
	// DisableRolloutUpdater indicates that the store should not update rollouts. This is useful for testing.
	DisableRolloutUpdater bool
}

// Store handles interacting with a storage backend,
//
//go:generate mockery --name=Store --filename=mock_store.go --structname=MockStore --with-expecter
//go:generate mockery --name Store --inpackage --filename mock_store.go --structname mockStore --with-expecter
type Store interface {
	Clear()

	Close() error

	Agent(ctx context.Context, id string) (*model.Agent, error)
	Agents(ctx context.Context, options ...QueryOption) ([]*model.Agent, error)
	AgentsCount(context.Context, ...QueryOption) (int, error)

	// UpsertAgent adds a new Agent to the Store or updates an existing one
	UpsertAgent(ctx context.Context, agentID string, updater AgentUpdater) (*model.Agent, error)
	UpsertAgents(ctx context.Context, agentIDs []string, updater AgentUpdater) ([]*model.Agent, error)

	// UpdateAgent updates an existing Agent in the Store. If the agentID does not exist, no error is returned but the
	// agent will be nil. An error is only returned if the update fails.
	UpdateAgent(ctx context.Context, agentID string, updater AgentUpdater) (*model.Agent, error)

	// UpdateAgentStatus will update the status of an existing agent. If the agentID does not exist, this does nothing. An
	// error is only returned if updating the status of the agent fails.
	//
	// When only the agent status needs to be modified, this should be preferred over UpdateAgent. In some store
	// implementations this will be more efficient.
	UpdateAgentStatus(ctx context.Context, agentID string, status model.AgentStatus) error

	// UpdateAgents updates existing Agents in the Store. If an agentID does not exist, that agentID is ignored and no
	// agent corresponding to that ID will be returned. An error is only returned if the update fails.
	UpdateAgents(ctx context.Context, agentIDs []string, updater AgentUpdater) ([]*model.Agent, error)

	DeleteAgents(ctx context.Context, agentIDs []string) ([]*model.Agent, error)

	AgentVersion(ctx context.Context, name string) (*model.AgentVersion, error)
	AgentVersions(ctx context.Context) ([]*model.AgentVersion, error)
	DeleteAgentVersion(ctx context.Context, name string) (*model.AgentVersion, error)

	Configurations(ctx context.Context, options ...QueryOption) ([]*model.Configuration, error)
	// Configuration returns a configuration by name.
	//
	// The name can optionally include a specific version, e.g. name:version.
	//
	// name:latest will return the latest version
	// name:current will return the current version that was applied to agents
	//
	// If the name does not include a version, the current version will be returned.
	// If the name does not exist, nil will be returned with no error.
	Configuration(ctx context.Context, name string) (*model.Configuration, error)
	UpdateConfiguration(ctx context.Context, name string, updater ConfigurationUpdater) (config *model.Configuration, status model.UpdateStatus, err error)
	DeleteConfiguration(ctx context.Context, name string) (*model.Configuration, error)

	Source(ctx context.Context, name string) (*model.Source, error)
	Sources(ctx context.Context) ([]*model.Source, error)
	DeleteSource(ctx context.Context, name string) (*model.Source, error)

	SourceType(ctx context.Context, name string) (*model.SourceType, error)
	SourceTypes(ctx context.Context) ([]*model.SourceType, error)
	DeleteSourceType(ctx context.Context, name string) (*model.SourceType, error)

	Processor(ctx context.Context, name string) (*model.Processor, error)
	Processors(ctx context.Context) ([]*model.Processor, error)
	DeleteProcessor(ctx context.Context, name string) (*model.Processor, error)

	ProcessorType(ctx context.Context, name string) (*model.ProcessorType, error)
	ProcessorTypes(ctx context.Context) ([]*model.ProcessorType, error)
	DeleteProcessorType(ctx context.Context, name string) (*model.ProcessorType, error)

	Destination(ctx context.Context, name string) (*model.Destination, error)
	Destinations(ctx context.Context) ([]*model.Destination, error)
	DeleteDestination(ctx context.Context, name string) (*model.Destination, error)

	DestinationType(ctx context.Context, name string) (*model.DestinationType, error)
	DestinationTypes(ctx context.Context) ([]*model.DestinationType, error)
	DeleteDestinationType(ctx context.Context, name string) (*model.DestinationType, error)
	// ApplyResources inserts or updates the specified resources. The resulting status of each resource is returned. The
	// resource may be modified as a result of the operation. If the caller needs to preserve the original resource,
	// model.Clone can be used to create a copy.
	ApplyResources(ctx context.Context, resources []model.Resource) ([]model.ResourceStatus, error)
	// Batch delete of a slice of resources, returns the successfully deleted resources or an error.
	DeleteResources(ctx context.Context, resources []model.Resource) ([]model.ResourceStatus, error)

	// AgentConfiguration returns the configuration that should be applied to an agent.
	//
	// It returns an error if the agent is nil. Returns the pending configuration if a rollout is in progress. Returns the
	// current configuration if there is one associated with the agent. Returns the configuration corresponding to the
	// configuration= label if the label exists on the agent. Returns the configuration matching the agent's labels if one
	// exists. Returns nil if there are no matches. If the agent does not have an associated configuration or the
	// associated configuration does not exist, no error is returned but the configuration will be nil. If there is an
	// error accessing the backend store, the error will be returned.
	AgentConfiguration(ctx context.Context, agent *model.Agent) (*model.Configuration, error)

	// AgentsIDsMatchingConfiguration returns the list of agent IDs that are using the specified configuration
	AgentsIDsMatchingConfiguration(ctx context.Context, conf *model.Configuration) ([]string, error)

	// ReportConnectedAgents sets the ReportedAt time for the specified agents to the specified time. This update should
	// not fire an update event for the agents on the Updates eventbus.
	ReportConnectedAgents(ctx context.Context, agentIDs []string, time time.Time) error

	// DisconnectUnreportedAgents sets the Status of agents to Disconnected if the agent ReportedAt time is before the
	// specified time.
	DisconnectUnreportedAgents(ctx context.Context, since time.Time) error

	// CleanupDisconnectedAgents removes agents that have disconnected before the specified time
	CleanupDisconnectedAgents(ctx context.Context, since time.Time) error

	// Updates will receive pipelines and configurations that have been updated or deleted, either because the
	// configuration changed or a component in them was updated. Agents inserted/updated from UpsertAgent and agents
	// removed from CleanupDisconnectedAgents are also sent with Updates.
	Updates(ctx context.Context) eventbus.Source[BasicEventUpdates]

	// AgentIndex provides access to the search AgentIndex implementation managed by the Store
	AgentIndex(ctx context.Context) search.Index

	// ConfigurationIndex provides access to the search Index for Configurations
	ConfigurationIndex(ctx context.Context) search.Index

	// UserSessions must implement the gorilla sessions.Store interface
	UserSessions() sessions.Store

	// Measurements stores stats for agents and configurations
	Measurements() stats.Measurements

	// StartRollout will start a rollout for the specified configuration with the specified options. If nil is passed for
	// options, any existing rollout options on the configuration status will be used. If there are no rollout options in
	// the configuration status, default values will be used for the rollout. If there is an existing rollout a different
	// version of this configuration, it will be replaced. Does nothing if the rollout does not have a
	// RolloutStatusPending status. Returns the current Configuration with its Rollout status.
	StartRollout(ctx context.Context, configurationName string, options *model.RolloutOptions) (*model.Configuration, error)

	// ResumeRollout will resume a rollout for the specified configuration.
	// Does nothing if the Rollout status is not RolloutStatusStarted or RolloutStatusStarted.
	// For RolloutStatusError - it will increase the maxErrors of the
	// rollout by the current number of errors + 1.
	// For RolloutStatusStarted - it will pause the rollout.
	PauseRollout(ctx context.Context, configurationName string) (*model.Configuration, error)

	// ResumeRollout will resume a rollout for the specified configuration.
	// Does nothing if the Rollout status is not RolloutStatusStarted or RolloutStatusStarted.
	// For RolloutStatusError - it will increase the maxErrors of the
	// rollout by the current number of errors + 1.
	// For RolloutStatusStarted - it will pause the rollout.
	ResumeRollout(ctx context.Context, configurationName string) (*model.Configuration, error)

	// UpdateRollout updates a rollout in progress. Does nothing if the rollout does not have a RolloutStatusStarted
	// status. Returns the current Configuration with its Rollout status.
	UpdateRollout(ctx context.Context, configurationName string) (*model.Configuration, error)

	// UpdatesRollouts updates all rollouts in progress. It returns each of the Configurations that contains an active
	// rollout. It is possible for this to partially succeed and return some updated rollouts and an error containing
	// errors from updating some other rollouts.
	UpdateRollouts(ctx context.Context) ([]*model.Configuration, error)

	// UpdateAllRollouts updates all active rollouts. The error returned is not intended to be returned to a client but can
	// be logged.
	UpdateAllRollouts(ctx context.Context) error

	ArchiveStore
}

// ArchiveStore provides access to archived resources for version history.
//
//go:generate mockery --name=ArchiveStore --filename=mock_archive_store.go --structname=MockArchiveStore
type ArchiveStore interface {
	// ResourceHistory returns all versions of the specified resource.
	ResourceHistory(ctx context.Context, resourceKind model.Kind, resourceName string) ([]*model.AnyResource, error)
}

// ErrDoesNotSupportHistory is used when a store does not implement resource history.
var ErrDoesNotSupportHistory = errors.New("store does not support resource history")

// AgentUpdater is given the current Agent model (possibly empty except for ID) and should update the Agent directly. We
// take this approach so that appropriate locking and/or transactions can be used for the operation as needed by the
// Store implementation.
type AgentUpdater func(current *model.Agent)

// ConfigurationUpdater is given the current Configuration model and should update the Configuration directly.
type ConfigurationUpdater func(current *model.Configuration)

// ----------------------------------------------------------------------
// search index helpers

// StartedRolloutsFromIndex returns a list of all rollouts that are not pending.
func StartedRolloutsFromIndex(_ context.Context, index search.Index) ([]string, error) {
	pendingConfigs, err := index.Suggestions(search.ParseQuery("rollout-pending:"))
	pendingConfigNames := make([]string, 0, len(pendingConfigs))
	for _, c := range pendingConfigs {
		pendingConfigNames = append(pendingConfigNames, c.Label)
	}
	if err != nil {
		return nil, err
	}
	return pendingConfigNames, nil
}

// FindAgents returns a list of agent IDs that match the specified key/value pair. The key must be a valid search field
func FindAgents(ctx context.Context, idx search.Index, key string, value string) ([]string, error) {
	q := key + ":" + value
	query := search.ParseQuery(q)
	return idx.Search(ctx, query)
}

// CurrentRolloutsForConfiguration returns a list of all rollouts that are currently in progress for the specified configuration.
func CurrentRolloutsForConfiguration(idx search.Index, configurationName string) ([]string, error) {
	pending, err := FindSuggestions(idx, model.FieldConfigurationPending, configurationName+":")
	if err != nil {
		return nil, err
	}
	future, err := FindSuggestions(idx, model.FieldConfigurationFuture, configurationName+":")
	if err != nil {
		return nil, err
	}
	result := []string{}
	result = append(result, pending...)
	result = append(result, future...)
	return result, nil
}

// FindSuggestions returns a list of all values for the specified key that start with the specified prefix.
func FindSuggestions(idx search.Index, key string, prefix string) ([]string, error) {
	q := key + ":" + prefix
	query := &search.Query{
		Original: q,
		Tokens: []*search.QueryToken{
			{
				Original: q,
				Name:     key,
				Value:    prefix,
			},
		},
	}
	suggestions, err := idx.Suggestions(query)
	if err != nil {
		return nil, err
	}

	var completions []string
	for _, s := range suggestions {
		completions = append(completions, strings.TrimSpace(strings.TrimPrefix(s.Query, key+":")))
	}
	return completions, nil
}

// ErrStoreResourceMissing is used internally in store functions to indicate a delete
// could not be performed because no such resource exists.  It
// should not ever be returned by a store function as an error
var ErrStoreResourceMissing = errors.New("resource not found")

// ErrResourceInUse is used in delete functions to indicate the delete
// could not be performed because the Resource is a dependency of another.
// i.e. the Source that is being deleted is being referenced in a Configuration.
var ErrResourceInUse = errors.New("resource in use")

// ----------------------------------------------------------------------

// QueryOptions represents the set of options available for a store query
type QueryOptions struct {
	Selector model.Selector
	Query    *search.Query
	Offset   int
	Limit    int
	Sort     string
}

// MakeQueryOptions creates a queryOptions struct from a slice of QueryOption
func MakeQueryOptions(options []QueryOption) QueryOptions {
	opts := QueryOptions{
		Selector: model.EverythingSelector(),
	}
	for _, opt := range options {
		opt(&opts)
	}
	return opts
}

// QueryOption is an option used in Store queries
type QueryOption func(*QueryOptions)

// WithSelector adds a selector to the query options
func WithSelector(selector model.Selector) QueryOption {
	return func(opts *QueryOptions) {
		opts.Selector = selector
	}
}

// WithQuery adds a search query string to the query options
func WithQuery(query *search.Query) QueryOption {
	return func(opts *QueryOptions) {
		opts.Query = query
	}
}

// WithOffset sets the offset for the results to return. For paging, if the pages have 10 items per page and this is the
// 3rd page, set the offset to 20.
func WithOffset(offset int) QueryOption {
	return func(opts *QueryOptions) {
		opts.Offset = offset
	}
}

// WithLimit sets the maximum number of results to return. For paging, if the pages have 10 items per page, set the
// limit to 10.
func WithLimit(limit int) QueryOption {
	return func(opts *QueryOptions) {
		opts.Limit = limit
	}
}

// WithSort sets the sort order for the request. The sort value is the name of the field, sorted ascending. To sort
// descending, prefix the field with a minus sign (-). Some Stores only allow sorting by certain fields. Sort values not
// supported will be ignored.
func WithSort(field string) QueryOption {
	return func(opts *QueryOptions) {
		opts.Sort = field
	}
}

// ----------------------------------------------------------------------
// seeding resources

// GetSeedResources returns all of the resources contained in SeedFolders.
func GetSeedResources(ctx context.Context, files embed.FS, folders []string) ([]model.Resource, error) {
	ctx, span := tracer.Start(ctx, "GetSeedResources")
	defer span.End()

	allEmbedded := make([]model.Resource, 0, 100)

	for _, dir := range folders {
		filesystem := files
		err := fs.WalkDir(filesystem, dir, func(path string, d fs.DirEntry, err error) error {
			if d.IsDir() {
				return nil
			}

			file, err := filesystem.Open(path)
			if err != nil {
				return err
			}

			r, err := model.ResourcesFromReader(file)
			if err != nil {
				return err
			}

			parsed, err := model.ParseResources(r)
			if err != nil {
				return err
			}

			allEmbedded = append(allEmbedded, parsed...)
			return nil
		})

		if err != nil {
			return nil, err
		}
	}
	return allEmbedded, nil
}

// separateDeprecatedResources will separate the deprecated resources from the non-deprecated resources.
func separateDeprecatedResources(resources []model.Resource) (deprecated []model.Resource, notDeprecated []model.Resource) {
	for _, r := range resources {
		if r.IsDeprecated() {
			deprecated = append(deprecated, r)
		} else {
			notDeprecated = append(notDeprecated, r)
		}
	}
	return deprecated, notDeprecated
}

// resourceExists will return true if the resource already exists in the store and any error encountered.
func resourceExists(ctx context.Context, store Store, resource model.Resource) (bool, error) {
	switch resource.GetKind() {
	case model.KindSourceType:
		existing, err := store.SourceType(ctx, resource.Name())
		return existing != nil, err

	case model.KindProcessorType:
		existing, err := store.ProcessorType(ctx, resource.Name())
		return existing != nil, err

	case model.KindDestinationType:
		existing, err := store.DestinationType(ctx, resource.Name())
		return existing != nil, err

	default:
		return false, fmt.Errorf("unsupported resource type: %s", resource.GetKind())
	}
}

// seedDeprecatedResource will only seed the resource if the resource already exists in the store. If the resource does
// not exist, it will return a resource status with model.StatusDeprecated, nil. If the resource exists, it will be
// updated and the status will be returned.
func seedDeprecatedResource(ctx context.Context, store Store, resource model.Resource) (*model.ResourceStatus, error) {
	// check if the resource already exists
	alreadyExists, err := resourceExists(ctx, store, resource)
	if err != nil {
		return model.NewResourceStatusWithError(resource, err), nil
	}

	// only apply deprecated resources if they already exist
	if alreadyExists {
		updates, err := store.ApplyResources(ctx, []model.Resource{resource})
		if len(updates) > 0 {
			return &updates[0], err
		}
		return nil, err
	}

	return model.NewResourceStatus(resource, model.StatusDeprecated), nil
}

// Seed adds bundled resources to the store
func Seed(ctx context.Context, store Store, logger *zap.Logger, files embed.FS, folders []string) error {
	ctx, span := tracer.Start(ctx, "store/Seed")
	defer span.End()

	resourceTypes, err := GetSeedResources(ctx, files, folders)
	if err != nil {
		span.RecordError(err)
		return err
	}

	deprecated, notDeprecated := separateDeprecatedResources(resourceTypes)

	// seed non-deprecated resources first
	updates, err := store.ApplyResources(ctx, notDeprecated)
	if err != nil {
		span.RecordError(err)
		return err
	}

	// only update deprecated resources if they already exist in the store
	for _, r := range deprecated {
		status, err := seedDeprecatedResource(ctx, store, r)
		if err != nil {
			span.RecordError(err)
			return err
		}
		// status will be nil if the deprecated resource is ignored
		if status != nil {
			updates = append(updates, *status)
		}
	}

	messages := make([]string, len(updates))
	for i, update := range updates {
		messages[i] = fmt.Sprintf("%s %s", update.Resource.Name(), update.Status)
	}

	logger.Info("Seeded ResourceTypes", zap.Any("resourceTypes", messages))
	return nil
}

// ----------------------------------------------------------------------
// seeding search indexes

// UpdateRolloutMetrics finds the agent counts for the configuration, updates the rollout status on the configuration,
// and returns a slice of the agent IDs that are waiting for a rollout and the number of those agents that should be
// moved into the pending state.
func UpdateRolloutMetrics(ctx context.Context, agentIndex search.Index, config *model.Configuration) ([]string, int, error) {
	nameAndVersion := config.NameAndVersion()
	agentsComplete, err := FindAgents(ctx, agentIndex, model.FieldRolloutComplete, nameAndVersion)
	if err != nil {
		return nil, 0, err
	}
	agentsError, err := FindAgents(ctx, agentIndex, model.FieldRolloutError, nameAndVersion)
	if err != nil {
		return nil, 0, err
	}
	agentsPending, err := FindAgents(ctx, agentIndex, model.FieldRolloutPending, nameAndVersion)
	if err != nil {
		return nil, 0, err
	}
	agentsWaiting, err := FindAgents(ctx, agentIndex, model.FieldRolloutWaiting, nameAndVersion)
	if err != nil {
		return nil, 0, err
	}

	newAgentsPending := config.Status.Rollout.UpdateStatus(model.RolloutProgress{
		Completed: len(agentsComplete),
		Errors:    len(agentsError),
		Pending:   len(agentsPending),
		Waiting:   len(agentsWaiting),
	})
	return agentsWaiting, newAgentsPending, nil
}

// SeedSearchIndexes seeds the search indexes with the current data in the store
func SeedSearchIndexes(ctx context.Context, store Store, logger *zap.Logger) {
	ctx, span := tracer.Start(ctx, "store/seedSearchIndexes")
	defer span.End()

	// seed search indexes
	err := seedConfigurationsIndex(ctx, store)
	if err != nil {
		logger.Error("unable to seed configurations into the search index, search results will be empty", zap.Error(err))
	}
	err = seedAgentsIndex(ctx, store)
	if err != nil {
		logger.Error("unable to seed agents into the search index, search results will be empty", zap.Error(err))
	}
}

func seedConfigurationsIndex(ctx context.Context, s Store) error {
	configurations, err := s.Configurations(ctx)
	if err != nil {
		return err
	}
	return seedIndex(configurations, s.ConfigurationIndex(ctx))
}

func seedAgentsIndex(ctx context.Context, s Store) error {
	agents, err := s.Agents(ctx)
	if err != nil {
		return err
	}
	return seedIndex(agents, s.AgentIndex(ctx))
}

func seedIndex[T modelSearch.Indexed](indexed []T, index search.Index) error {
	var errs error
	for _, i := range indexed {
		err := index.Upsert(i)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	return errs
}

// Dependency is a single item in DependentResources.
type Dependency struct {
	Name string
	Kind model.Kind
}

// DependentResources is the return type of store.dependentResources
// and used to construct DependencyError.
// It has help methods empty(), message(), and add().
type DependentResources []Dependency

// Empty returns true if the list of dependencies is empty.
func (r *DependentResources) Empty() bool {
	return len(*r) == 0
}

// Message returns a string representation of the list of dependencies.
func (r *DependentResources) Message() string {
	msg := "Dependent resources:\n"
	for _, item := range *r {
		msg += fmt.Sprintf("%s %s\n", item.Kind, item.Name)
	}
	return msg
}

// Add adds a new dependency to the list.
func (r *DependentResources) Add(d Dependency) {
	*r = append(*r, d)
}

// DependencyError is returned when trying to delete a resource
// that is being referenced by other resources.
type DependencyError struct {
	Dependencies DependentResources
}

func (de *DependencyError) Error() string {
	return de.Dependencies.Message()
}

// NewDependencyError creates a new DependencyError with the given dependencies.
func NewDependencyError(d DependentResources) error {
	return &DependencyError{
		Dependencies: d,
	}
}

// ----------------------------------------------------------------------

// FindDependentResources finds the dependent resources using the ConfigurationIndex provided by the Store.
func FindDependentResources(ctx context.Context, configurationIndex search.Index, name string, kind model.Kind) (DependentResources, error) {
	var dependencies DependentResources

	// ignore version when searching for dependencies
	name, _ = model.SplitVersion(name)
	switch kind {
	case model.KindSource:
		ids, err := search.Field(ctx, configurationIndex, "source", name)
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			dependencies.Add(Dependency{Name: id, Kind: model.KindConfiguration})
		}

	case model.KindDestination:
		ids, err := search.Field(ctx, configurationIndex, "destination", name)
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			dependencies.Add(Dependency{Name: id, Kind: model.KindConfiguration})
		}
	}

	return dependencies, nil
}

// ----------------------------------------------------------------------
// generic helpers for sorting and paging

// ApplySortOffsetAndLimit applies the sort, offset, and limit options to the list.
func ApplySortOffsetAndLimit[T any](list []T, opts QueryOptions, fieldAccessor fieldAccessor[T]) []T {
	if opts.Sort != "" {
		sortField := opts.Sort
		ascending := true
		if opts.Sort[0] == '-' {
			sortField = opts.Sort[1:]
			ascending = false
		}
		sort.Sort(byField[T]{
			list:          list,
			field:         sortField,
			ascending:     ascending,
			fieldAccessor: fieldAccessor,
		})
	}
	if opts.Offset != 0 {
		offset := opts.Offset
		if offset > len(list) {
			offset = len(list)
		}
		list = list[offset:]
	}
	if opts.Limit != 0 {
		limit := opts.Limit
		if limit > len(list) {
			limit = len(list)
		}
		list = list[:limit]
	}
	return list
}

// NewBPCookieStore creates a new CookieStore with the specified secret.
func NewBPCookieStore(secret string) *sessions.CookieStore {
	store := sessions.NewCookieStore([]byte(secret))
	store.Options.MaxAge = 60 * 60 // 60 minutes
	store.Options.SameSite = http.SameSiteStrictMode
	return store
}

type fieldAccessor[T any] func(field string, item T) string

type byField[T any] struct {
	list          []T
	field         string
	ascending     bool
	fieldAccessor fieldAccessor[T]
}

var _ sort.Interface = (*byField[any])(nil)

// Len returns the length of the list
func (r byField[T]) Len() int {
	return len(r.list)
}

// Swap swaps to items in the list
func (r byField[T]) Swap(i, j int) {
	r.list[i], r.list[j] = r.list[j], r.list[i]
}

// Less returns true if the i'th item is less than the j'th
func (r byField[T]) Less(i, j int) bool {
	f1 := r.fieldAccessor(r.field, r.list[i])
	f2 := r.fieldAccessor(r.field, r.list[j])
	if r.ascending {
		return f1 < f2
	}
	return f1 > f2
}

// temporary function until all stores are migrated to the new interface
func updateConfigurationUsingApplyResources(ctx context.Context, store Store, name string, updater ConfigurationUpdater) (config *model.Configuration, status model.UpdateStatus, err error) {
	config, err = store.Configuration(ctx, name)
	if err != nil {
		return nil, model.StatusError, err
	}
	if config == nil {
		return nil, model.StatusUnchanged, nil
	}
	updater(config)
	statuses, err := store.ApplyResources(ctx, []model.Resource{config})
	if err != nil || statuses == nil || len(statuses) == 0 {
		return nil, model.StatusError, err
	}
	status = statuses[0].Status
	return config, status, err
}

// IsNewConfigurationVersion handles the special case for configurations where we only create new versions in certain
// cases. if the rollout has already started (not pending) and the spec has changed, we create a new version.
func IsNewConfigurationVersion(curResource *model.AnyResource, newResource model.Resource) (bool, *model.Configuration, error) {
	// if not a configuration resource, not a new version
	if newResource.GetKind() != model.KindConfiguration {
		return false, nil, nil
	}

	// if there is no existing resource, not a new version (somewhat arbitrary, but avoid panic)
	if curResource == nil {
		return false, nil, nil
	}

	// if new is not a configuration resource, error
	newConfig, err := model.AsKind[*model.Configuration](newResource)
	if err != nil {
		return false, nil, fmt.Errorf("newResource is not a Configuration: %w", err)
	}

	// if old is not a configuration resource, error
	oldConfig, err := model.AsKind[*model.Configuration](curResource)
	if err != nil {
		return false, nil, fmt.Errorf("curResource is not a Configuration: %w", err)
	}

	// if existing rollout hasn't started, not a new version
	if oldConfig.Status.Rollout.Status == model.RolloutStatusPending {
		return false, nil, nil
	}

	// if the hash is the same, not a new version
	if oldConfig.Hash() == newConfig.Hash() {
		return false, nil, nil
	}

	// compare hash to compare spec
	return true, newConfig, nil
}

// processAgentRolloutUpdate calls update rollout on the configName
func processAgentRolloutUpdate(ctx context.Context, s Store, configName string) error {
	name := model.TrimVersion(configName)
	if name == "" {
		return nil
	}
	_, err := s.UpdateRollout(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to update rollout %s: %w", configName, err)
	}

	return nil
}

// ----------------------------------------------------------------------
// updating dependencies

// UpdateDependentResources updates any resources that are dependent on the resources updated. The updated resources
// will be retrieved from the Updates and additional resources will be added to Updates as they are updated.
//
// This function also takes the existing statuses and errors (joined using errors.Join) and returns them with any
// additionally modified resources and errors.
func UpdateDependentResources(ctx context.Context, store Store, resources []model.Resource, statuses []model.ResourceStatus, errs error) ([]model.ResourceStatus, error) {
	// it's important to apply the updates in groups by kind so that transitive dependencies (configuration => source =>
	// processor) are updated in the correct (reverse) order.

	// group by kind
	byKind := map[model.Kind][]model.Resource{}
	for _, resource := range resources {
		kind := resource.GetKind()
		byKind[kind] = append(byKind[kind], resource)
	}

	// apply each group in order with processors first and configurations last
	for _, kind := range []model.Kind{
		model.KindProcessor,
		model.KindSource,
		model.KindDestination,
		model.KindConfiguration,
	} {
		group := byKind[kind]
		if len(group) == 0 {
			continue
		}
		// for each resource that needs to be updated, update its dependencies
		for _, resource := range group {
			err := resource.UpdateDependencies(ctx, store)
			if err != nil {
				errs = errors.Join(errs, err)
			}
		}
		// after updating their dependencies, apply the updates
		updateStatuses, err := store.ApplyResources(ctx, group)
		statuses = append(statuses, updateStatuses...)
		if err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return statuses, errs
}

// ----------------------------------------------------------------------
// sensitive parameters

// MaskSensitiveParameters masks sensitive parameter values based on the ParameterDefinitions in the ResourceType. This
// should be called after reading a value from the store before returning it.
func MaskSensitiveParameters[R model.Resource](ctx context.Context, resource R) {
	if resourceWithSensitiveParameters, ok := any(resource).(model.HasSensitiveParameters); ok {
		resourceWithSensitiveParameters.MaskSensitiveParameters(ctx)
	}
}

// PreserveSensitiveParameters will replace sensitive parameters in the current Resource with values from the existing
// Resource. This should be called before writing an updated value to the store.
func PreserveSensitiveParameters(ctx context.Context, updated model.Resource, existing *model.AnyResource) error {
	if resourceWithSensitiveParameters, ok := any(updated).(model.HasSensitiveParameters); ok {
		return resourceWithSensitiveParameters.PreserveSensitiveParameters(ctx, existing)
	}
	return nil
}
