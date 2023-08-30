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

// Package model contains the data model for resources used in BindPlane
package model

import (
	"database/sql/driver"
	"errors"
	"sort"
	"time"

	jsoniter "github.com/json-iterator/go"
	modelSearch "github.com/observiq/bindplane-op/model/search"
	"github.com/observiq/bindplane-op/util/semver"
)

// AgentTypeName is the name of a type of agent
type AgentTypeName string

const (
	// AgentTypeNameObservIQOtelCollector is the name of the observIQ Distro for OpenTelemetry Collector
	AgentTypeNameObservIQOtelCollector AgentTypeName = "observiq-otel-collector"
)

// AgentStatus is the status of the Agent's connection to bindplane platform
type AgentStatus uint8

const (
	// Disconnected is the state of an agent that was formerly Connected to the management platform but is no longer
	// connected. This could mean that the agent has stopped running, the network connection has been interrupted, or that
	// the server terminated the connection.
	Disconnected AgentStatus = 0

	// Connected is the normal state of a healthy agent that is Connected to the management platform.
	Connected AgentStatus = 1

	// Error occurs if there is an error running or Configuring the agent.
	Error AgentStatus = 2

	// ComponentFailed is deprecated.
	ComponentFailed AgentStatus = 4

	// Deleted is set on a deleted Agent before notifying observers of the change.
	Deleted AgentStatus = 5

	// Configuring is set on an Agent when it is sent a new configuration that has not been applied. After successful
	// Configuring, it will transition back to Connected. If there is an error Configuring, it will transition to Error.
	Configuring AgentStatus = 6

	// Upgrading is set on an Agent when it has been sent a new package that is being applied. After Upgrading, it will
	// transition back to Connected or Error unless it already has the Configuring status.
	Upgrading AgentStatus = 7
)

// AgentUpgradeStatus is the status of the AgentUpgrade
type AgentUpgradeStatus uint8

const (
	// UpgradePending is set when the upgrade is initially started
	UpgradePending AgentUpgradeStatus = 0
	// UpgradeStarted is set when the upgrade has been sent to the agent
	UpgradeStarted AgentUpgradeStatus = 1
	// UpgradeFailed is set when the upgrade is complete but there was an error. If the upgrade is successful, the Agent
	// Upgrade field will be set to nil and there is no corresponding status.
	UpgradeFailed AgentUpgradeStatus = 2
)

// AgentUpgrade stores information on an Agent about the upgrade process.
type AgentUpgrade struct {
	// Status indicates the progress of the agent upgrade
	Status AgentUpgradeStatus `json:"status" yaml:"status"`

	// Version is used to indicate that an agent should be or is being upgraded. The agent status will be set to Upgrading
	// when the upgrade begins.
	Version string `json:"version,omitempty" yaml:"version,omitempty"`

	// AllPackagesHash is the hash of the packages sent to the agent to upgrade
	AllPackagesHash []byte `json:"allPackagesHash,omitempty" yaml:"allPackagesHash,omitempty"`

	// Error is set if there were errors upgrading the agent
	Error string `json:"error,omitempty" yaml:"error,omitempty"`
}

// Value is used to translate to a JSONB field for postgres storage
func (s AgentUpgrade) Value() (driver.Value, error) {
	return jsoniter.Marshal(s)
}

// Scan is used to translate from a JSONB field in postgres to AgentUpgrade
func (s *AgentUpgrade) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return jsoniter.Unmarshal(b, &s)
}

// AgentFeatures is a bitmask of features supported by the Agent, usually based on its version.
type AgentFeatures uint32

const (
	// AgentSupportsUpgrade will be set if this Agent can be upgraded remotely.
	AgentSupportsUpgrade AgentFeatures = 1 << iota

	// AgentSupportsSnapshots will be set if this Agent can send snapshots of recent telemetry signals.
	AgentSupportsSnapshots

	// AgentSupportsMeasurements will be set if this Agent supports the throughputmeasurement processor for measuring throughput.
	AgentSupportsMeasurements

	// AgentSupportsLogBasedMetrics will be set if this agent supports the logcount processor and route receiver for creating metrics from logs.
	AgentSupportsLogBasedMetrics
)

// AgentFeaturesDefault is the default bitmask of features supported by Agents
const AgentFeaturesDefault = AgentSupportsUpgrade

// Agent is a single observIQ OTel Collector instance
type Agent struct {
	ID              string `json:"id" yaml:"id" db:"id"`
	Name            string `json:"name" yaml:"name" db:"name"`
	Type            string `json:"type" yaml:"type" db:"type"`
	Architecture    string `json:"arch" yaml:"arch" db:"architecture"`
	HostName        string `json:"hostname" yaml:"hostname" db:"host_name"`
	Labels          Labels `json:"labels,omitempty" yaml:"labels" db:"labels"`
	Version         string `json:"version" yaml:"version" db:"version"`
	Home            string `json:"home" yaml:"home" db:"home_directory"`
	Platform        string `json:"platform" yaml:"platform" db:"platform"`
	OperatingSystem string `json:"operatingSystem" yaml:"operatingSystem" db:"operating_system"`
	MacAddress      string `json:"macAddress" yaml:"macAddress" db:"mac_address"`
	RemoteAddress   string `json:"remoteAddress,omitempty" yaml:"remoteAddress,omitempty" db:"remote_address"`

	// SecretKey is provided by the agent to authenticate
	SecretKey string `json:"-" yaml:"-"`

	// TLS from the agent's manager.yaml
	TLS *ManagerTLS `json:"tls,omitempty" yaml:"tls,omitempty"`

	// Upgrade stores information about an agent upgrade
	Upgrade *AgentUpgrade `json:"upgrade,omitempty" yaml:"upgrade,omitempty" db:"upgrade,omitempty"`

	// reported by Status messages
	Status       AgentStatus `json:"status" db:"status"`
	ErrorMessage string      `json:"errorMessage,omitempty" yaml:"errorMessage,omitempty" db:"error_message"`

	// tracked by BindPlane
	Configuration  any        `json:"configuration,omitempty" yaml:"configuration,omitempty" db:"configuration"`
	ConnectedAt    *time.Time `json:"connectedAt,omitempty" yaml:"connectedAt,omitempty" db:"connected_at"`
	DisconnectedAt *time.Time `json:"disconnectedAt,omitempty" yaml:"disconnectedAt,omitempty" db:"disconnected_at"`
	ReportedAt     *time.Time `json:"reported_at,omitempty" yaml:"reported_at,omitempty" db:"reported_at"`

	Protocol            string                `json:"protocol,omitempty" yaml:"protocol,omitempty" db:"management_protocol"`
	State               any                   `json:"state,omitempty" yaml:"state,omitempty" db:"management_state"`
	ConfigurationStatus ConfigurationVersions `json:"configurationStatus,omitempty" yaml:"configurationStatus,omitempty" db:",inline"`
}

// ConfigurationVersions tracks the current, pending, and future configurations for this agent.
// format: <configuration name>:<version>
type ConfigurationVersions struct {
	// Current is the configuration currently applied to the agent.
	Current string `json:"current,omitempty" yaml:"current,omitempty" db:"current_configuration"`

	// Pending is the configuration that is assigned to the agent but may not be applied. Once this configuration is
	// confirmed, Current will be set to this value and this will be cleared.
	Pending string `json:"pending,omitempty" yaml:"pending,omitempty" db:"pending_configuration"`

	// Future is the configuration that will be assigned to this agent when the rollout assigns the new configuration to
	// the agent. Once the rollout assigns the configuration, Pending will be set to this value and this will be cleared.
	Future string `json:"future,omitempty" yaml:"future,omitempty" db:"future_configuration"`
}

// Clear clears the configuration versions, setting them all to ""
func (cv *ConfigurationVersions) Clear() {
	cv.clear()
}

// Set is a convenience method to set current, pending, and future configuration versions
func (cv *ConfigurationVersions) Set(current, pending, future string) {
	cv.Current = current
	cv.Pending = pending
	cv.Future = future
}

// ManagerTLS are the TLS settings for the agent when connecting to BPOP
type ManagerTLS struct {
	InsecureSkipVerify bool    `json:"insecure_skip_verify,omitempty" yaml:"insecure_skip_verify,omitempty"`
	CAFile             *string `json:"ca_file,omitempty" yaml:"ca_file,omitempty"`
	CertFile           *string `json:"cert_file,omitempty" yaml:"cert_file,omitempty"`
	KeyFile            *string `json:"key_file,omitempty" yaml:"key_file,omitempty"`
}

var _ modelSearch.Indexed = (*Agent)(nil)
var _ HasUniqueKey = (*Agent)(nil)
var _ Labeled = (*Agent)(nil)

// UniqueKey returns the agent ID to uniquely identify an Agent
func (a *Agent) UniqueKey() string {
	return a.ID
}

// StatusDisplayText returns the string representation of the agent's status.
func (a *Agent) StatusDisplayText() string {
	switch a.Status {
	case Disconnected:
		return "Disconnected"
	case Connected:
		return "Connected"
	case Error:
		return "Error"
	case ComponentFailed:
		return "Component Failed"
	case Deleted:
		return "Deleted"
	case Configuring:
		return "Configuring"
	case Upgrading:
		return "Upgrading"
	default:
		return "Unknown"
	}
}

// GetLabels implements the Labeled interface for Agents
func (a *Agent) GetLabels() Labels {
	return a.Labels
}

// SetLabels implements the Labeled interface for Agents
func (a *Agent) SetLabels(l Labels) {
	a.Labels.Set = l.Set
}

// ConnectedDurationDisplayText returns the duration since the agent connected.
func (a *Agent) ConnectedDurationDisplayText() string {
	if a.Status == Disconnected {
		return "-"
	}
	return durationDisplay(a.ConnectedAt)
}

// ReportedDurationDisplayText returns the duration since the agent last reported.
func (a *Agent) ReportedDurationDisplayText() string {
	if a.Status == Disconnected {
		return "-"
	}
	return durationDisplay(a.ReportedAt)
}

// DisconnectedDurationDisplayText returns the duration since the agent disconnected.
func (a *Agent) DisconnectedDurationDisplayText() string {
	return durationDisplay(a.DisconnectedAt)
}

// MatchesSelector returns true if the given selector matches the agent's labels.
func (a *Agent) MatchesSelector(selector Selector) bool {
	return selector.Matches(a.Labels)
}

// DisconnectedSince returns true if the agent has been disconnected since a given time.
func (a *Agent) DisconnectedSince(since time.Time) bool {
	return a.DisconnectedAt != nil && a.DisconnectedAt.Before(since)
}

// ReportedSince returns true if the agent has reported since a given time.
func (a *Agent) ReportedSince(since time.Time) bool {
	return a.ReportedAt != nil && !a.ReportedAt.Before(since)
}

// Connect updates the ConnectedAt, ReportedAt, and DisconnectedAt fields of the agent and should be called when the
// agent connects.
func (a *Agent) Connect(newAgentVersion string) {
	// always update ReportedAt
	now := time.Now()
	a.ReportedAt = &now
	// only update ConnectedAt if this is a new version or never connected
	if a.Version != newAgentVersion || a.ConnectedAt == nil {
		a.ConnectedAt = &now
	}
	a.DisconnectedAt = nil
}

// Disconnect updates the DisconnectedAt and Status fields of the agent and should be called when the agent disconnects.
// If the agent is already disconnected, this does nothing.
func (a *Agent) Disconnect() {
	if a.Status != Disconnected {
		now := time.Now()
		a.DisconnectedAt = &now
		a.Status = Disconnected

		// clear out the ConnectedAt so that it gets reset when the agent reconnects
		a.ConnectedAt = nil
	}
}

func durationDisplay(t *time.Time) string {
	if t == nil || t.IsZero() {
		return "-"
	}
	return time.Since(*t).Round(time.Second).String()
}

// ----------------------------------------------------------------------
// configure

// SetCurrentConfiguration sets the Current configuration in the ConfigurationStatus on the agent based on the specified
// configuration. If Pending or Future are set to this value, they will be cleared.
//
// SetCurrentConfiguration will be called when the configuration is confirmed to be currently running on the Agent.
func (a *Agent) SetCurrentConfiguration(configuration *Configuration) {
	if configuration == nil {
		a.ConfigurationStatus.clear()
		return
	}
	nameAndVersion := configuration.NameAndVersion()
	a.ConfigurationStatus.Current = nameAndVersion

	if a.ConfigurationStatus.Pending == nameAndVersion {
		// configuration name and version not pending, now current
		a.ConfigurationStatus.Pending = ""
	}
	if a.ConfigurationStatus.Future == nameAndVersion {
		// configuration name and version not future, now current
		a.ConfigurationStatus.Future = ""
	}
}

// SetPendingConfiguration sets the Pending configuration in the ConfigurationStatus on the agent based on the specified
// configuration. If Future is already set to this value, it will be cleared. If Current is already set to this value,
// Pending will also be cleared. If configuration is nil, Current, Pending and Future will all be cleared.
//
// SetPendingConfiguration will be called by the rollout manager when the Agent should receive a new configuration.
func (a *Agent) SetPendingConfiguration(configuration *Configuration) {
	if configuration == nil {
		a.ConfigurationStatus.clear()
		return
	}
	nameAndVersion := configuration.NameAndVersion()
	if a.ConfigurationStatus.Current == nameAndVersion {
		// configuration name and version already current
		a.ConfigurationStatus.Pending = ""
		a.ConfigurationStatus.Future = ""
		return
	}
	if a.ConfigurationStatus.Future == nameAndVersion {
		// configuration name and version not future, now pending
		a.ConfigurationStatus.Future = ""
	}
	a.ConfigurationStatus.Pending = nameAndVersion
	if a.Status == Error {
		a.Status = Configuring
	}
}

// SetFutureConfiguration sets the Future configuration in the ConfigurationStatus on the agent based on the specified
// configuration. Typically this will set the Future configuration and clear Pending, but if Pending is already set to
// this Configuration, Future will be cleared. If Current is already set to this Configuration, Pending and Future will
// both be cleared. If configuration is nil, Current, Pending and Future will all be cleared.
//
// SetFutureConfiguration will be called when the configuration of an Agent should change. The rollout manager will
// handle scheduling updates and move it to Pending.
func (a *Agent) SetFutureConfiguration(configuration *Configuration) {
	if configuration == nil {
		a.ConfigurationStatus.clear()
		return
	}
	nameAndVersion := configuration.NameAndVersion()
	if a.ConfigurationStatus.Current == nameAndVersion {
		// configuration name and version already current
		a.ConfigurationStatus.Pending = ""
		a.ConfigurationStatus.Future = ""
		return
	}
	if a.ConfigurationStatus.Pending == nameAndVersion {
		// configuration name and version already pending
		a.ConfigurationStatus.Future = ""
		return
	}
	// clear pending to avoid unnecessary configuration churn
	a.ConfigurationStatus.Pending = ""
	a.ConfigurationStatus.Future = nameAndVersion
}

// clear sets all of the configuration versions to empty strings.
func (cv *ConfigurationVersions) clear() {
	cv.Current = ""
	cv.Pending = ""
	cv.Future = ""
}

// ClearRollout clears the Pending and Future configuration versions.
func (cv *ConfigurationVersions) ClearRollout() {
	cv.Pending = ""
	cv.Future = ""
}

// ----------------------------------------------------------------------
// features

var v1_6_0 = semver.Parse("1.6.0")
var v1_8_0 = semver.Parse("1.8.0")
var v1_9_2 = semver.Parse("1.9.2")
var v1_14_0 = semver.Parse("1.14.0")

// Features returns a bitmask of the features supported by this Agent
func (a *Agent) Features() AgentFeatures {
	agentVersion := semver.Parse(a.Version)

	// arrange version checks newest first
	switch {
	// 1.14.0 introduced the logcount processor and route receiver
	case !agentVersion.IsOlder(v1_14_0):
		return AgentSupportsUpgrade | AgentSupportsSnapshots | AgentSupportsMeasurements | AgentSupportsLogBasedMetrics
	// 1.9.2 introduced the throughputmeasurement processor
	case !agentVersion.IsOlder(v1_9_2):
		return AgentSupportsUpgrade | AgentSupportsSnapshots | AgentSupportsMeasurements

	// 1.8.0 introduced snapshots
	case !agentVersion.IsOlder(v1_8_0):
		return AgentSupportsUpgrade | AgentSupportsSnapshots

	// 1.6.0 introduced upgrade
	case !agentVersion.IsOlder(v1_6_0):
		return AgentSupportsUpgrade

	}
	return 0
}

// HasFeatures returns true if this Agent supports the specified feature or features.
func (a *Agent) HasFeatures(feature AgentFeatures) bool {
	return a.Features().Has(feature)
}

// Has returns true if the AgentFeatures bitmask has the specified features enabled. This is simply implemented as
// agentFeatures&feature != 0 but I prefer to avoid using bitmasks directly when possible.
func (agentFeatures AgentFeatures) Has(feature AgentFeatures) bool {
	return agentFeatures&feature != 0
}

// ----------------------------------------------------------------------
// upgrading

// SupportsUpgrade returns true if this agent supports upgrade
func (a *Agent) SupportsUpgrade() bool {
	// Ideally this would be based on the opamp flag AgentCapabilities_AcceptsPackages but agent capabilities aren't
	// currently available on the Agent model. That should change but for now the version will be checked.
	return a.HasFeatures(AgentSupportsUpgrade)
}

// UpgradeTo begins an upgrade by setting the status to Upgrading and setting the Upgrade field to the specified
// version.
func (a *Agent) UpgradeTo(version string) {
	if !a.SupportsUpgrade() {
		return
	}
	a.Upgrade = &AgentUpgrade{
		Version: version,
		Status:  UpgradePending,
	}
	a.Status = Upgrading
}

// UpgradeStarted is set when the upgrade instructions have actually been sent to the Agent.
func (a *Agent) UpgradeStarted(version string, allPackagesHash []byte) {
	a.Upgrade = &AgentUpgrade{
		Version:         version,
		Status:          UpgradeStarted,
		AllPackagesHash: allPackagesHash,
	}
	a.Status = Upgrading
}

// UpgradeComplete completes an upgrade by setting the status back to either Connected or Error (depending on
// ErrorMessage) and either removing the AgentUpgrade field or setting the Error on it if the specified errorMessage is
// not empty.
func (a *Agent) UpgradeComplete(version, errorMessage string) {
	if errorMessage != "" {
		// set the errorMessage on the AgentUpgrade
		if a.Upgrade == nil {
			a.Upgrade = &AgentUpgrade{}
		}
		a.Upgrade.Status = UpgradeFailed
		a.Upgrade.Error = errorMessage
		if version != "" {
			a.Upgrade.Version = version
		}
	} else {
		// clear the upgrade and reset the Status. if the Status is Configuring, allow the configuring process to continue.
		a.Upgrade = nil
	}
	if a.Status != Configuring {
		// preserve Error state if a configuration error exists after update.
		if a.ErrorMessage != "" {
			a.Status = Error
		} else {
			a.Status = Connected
		}
	}
}

// ----------------------------------------------------------------------
// sorting

type byName []*Agent

func (s byName) Len() int {
	return len(s)
}

func (s byName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byName) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

// SortAgentsByName sorts the specified slice of Agents by name.
func SortAgentsByName(agents []*Agent) {
	sort.Sort(byName(agents))
}

// ----------------------------------------------------------------------
// indexing

const (
	// FieldConfigurationCurrent is the name of the index field used to store the current configuration name:version
	FieldConfigurationCurrent = "configuration-current"

	// FieldConfigurationPending is the name of the index field used to store the pending configuration name:version
	FieldConfigurationPending = "configuration-pending"

	// FieldConfigurationFuture is the name of the index field used to store the future configuration name:version
	FieldConfigurationFuture = "configuration-future"

	// Rollout fields are used for tracking rollout progress for an Agent for a given configuration. For a given Agent,
	// there could be multiple rollout states associated with it. It could have completed version 1, have an error rolling
	// out version 2, and be waiting for version 3.

	// FieldRolloutComplete is the name of the index field used to store the complete rollout configuration name:version
	FieldRolloutComplete = "rollout-complete"

	// FieldRolloutError is the name of the index field used to store the error rollout configuration name:version
	FieldRolloutError = "rollout-error"

	// FieldRolloutPending is the name of the index field used to store the pending rollout configuration name:version
	FieldRolloutPending = "rollout-pending"

	// FieldRolloutWaiting is the name of the index field used to store the waiting rollout configuration name:version
	FieldRolloutWaiting = "rollout-waiting"
)

// IndexID returns an ID used to identify the resource that is indexed
func (a *Agent) IndexID() string {
	return a.ID
}

// IndexFields returns a map of field name to field value to be stored in the index
func (a *Agent) IndexFields(index modelSearch.Indexer) {
	index("id", a.ID)
	index("arch", a.Architecture)
	index("hostname", a.HostName)
	index("platform", a.Platform)
	index("version", a.Version)
	index("name", a.Name)
	index("home", a.Home)
	index("os", a.OperatingSystem)
	index("macAddress", a.MacAddress)
	index("type", a.Type)
	index("status", a.StatusDisplayText())

	// index the configuration name and current, pending, and future versions
	index(FieldConfigurationCurrent, a.ConfigurationStatus.Current)
	index(FieldConfigurationPending, a.ConfigurationStatus.Pending)
	index(FieldConfigurationFuture, a.ConfigurationStatus.Future)

	switch a.Status {
	case Deleted, Disconnected:
		// do nothing
	case Error:
		index(FieldRolloutError, a.ConfigurationStatus.Pending)
		index(FieldRolloutWaiting, a.ConfigurationStatus.Future)
	default:
		// removing this from the index in the case of an error allows the
		// user to find the agents when moving them back to their original config
		configuration, _ := SplitVersion(a.ConfigurationStatus.Current)
		index("configuration", configuration)
		index(FieldRolloutPending, a.ConfigurationStatus.Pending)
		index(FieldRolloutWaiting, a.ConfigurationStatus.Future)
	}
	index(FieldRolloutComplete, a.ConfigurationStatus.Current)
}

// IndexLabels returns a map of label name to label value to be stored in the index
func (a *Agent) IndexLabels(index modelSearch.Indexer) {
	for n, v := range a.Labels.Set {
		index(n, v)
	}
}

// ----------------------------------------------------------------------
// Printable

// PrintableKindSingular returns the singular form of the Kind, e.g. "Agent"
func (a *Agent) PrintableKindSingular() string {
	return "Agent"
}

// PrintableKindPlural returns the singular form of the Kind, e.g. "Agents"
func (a *Agent) PrintableKindPlural() string {
	return "Agents"
}

// PrintableFieldTitles returns the list of field titles, used for printing a table of resources
func (a *Agent) PrintableFieldTitles() []string {
	return []string{"ID", "Name", "Version", "Status", "Configuration", "Connected", "Reported", "Disconnected", "Labels"}
}

// PrintableFieldValue returns the field value for a title, used for printing a table of resources
func (a *Agent) PrintableFieldValue(title string) string {
	switch title {
	case "ID":
		return a.ID
	case "Name":
		return a.Name
	case "Version":
		return a.Version
	case "Status":
		return a.StatusDisplayText()
	case "Connected":
		return a.ConnectedDurationDisplayText()
	case "Reported":
		return a.ReportedDurationDisplayText()
	case "Disconnected":
		return a.DisconnectedDurationDisplayText()
	case "Labels":
		return a.Labels.Custom().String()
	case "Configuration":
		if a.ConfigurationStatus.Current != "" {
			return a.ConfigurationStatus.Current
		}
		return "-"
	case "Current":
		return a.ConfigurationStatus.Current
	case "Pending":
		return a.ConfigurationStatus.Pending
	case "Future":
		return a.ConfigurationStatus.Future
	}
	return ""
}
