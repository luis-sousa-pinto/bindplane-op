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

package model

import (
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/observiq/bindplane-op/common"
	"github.com/observiq/bindplane-op/model/validation"
	"github.com/observiq/bindplane-op/model/version"
	"github.com/observiq/bindplane-op/util/semver"
)

type agentVersionKind struct{}

func (k *agentVersionKind) NewEmptyResource() *AgentVersion { return &AgentVersion{} }

// AgentVersion is the resource for a version of an agent and includes links to install scripts and downloads links for
// the agent release.
type AgentVersion struct {
	// ResourceMeta TODO(doc)
	ResourceMeta `yaml:",inline" mapstructure:",squash"`
	// Spec TODO(doc)
	Spec                 AgentVersionSpec `json:"spec" yaml:"spec" mapstructure:"spec"`
	StatusType[NoStatus] `yaml:",inline" mapstructure:",squash"`
}

// AgentVersionSpec is the spec for an AgentVersion
type AgentVersionSpec struct {
	Type            string `yaml:"type" json:"type" mapstructure:"type"`
	Version         string `yaml:"version" json:"version" mapstructure:"version"`
	ReleaseNotesURL string `yaml:"releaseNotesURL" json:"releaseNotesURL" mapstructure:"releaseNotesURL"`
	Draft           bool   `yaml:"draft" json:"draft" mapstructure:"draft"`
	Prerelease      bool   `yaml:"prerelease" json:"prerelease" mapstructure:"prerelease"`
	// Map of platform to installer script URL
	Installer map[string]AgentInstaller `yaml:"installer" json:"installer" mapstructure:"installer"`
	// Map of platform to download URL and hash
	Download map[string]AgentDownload `yaml:"download" json:"download" mapstructure:"download"`

	// ReleaseDate is an RFC3339 encoded date in a string
	ReleaseDate string `yaml:"releaseDate" json:"releaseDate" mapstructure:"releaseDate"`
}

// AgentInstaller contains the url of the install script
type AgentInstaller struct {
	URL string `yaml:"url" json:"url" mapstructure:"url"`
}

// AgentDownload contains the url to download the agent release and a hash to verify the contents of the download.
type AgentDownload struct {
	URL  string `yaml:"url" json:"url" mapstructure:"url"`
	Hash string `yaml:"hash" json:"hash" mapstructure:"hash"`
}

// ----------------------------------------------------------------------

// NewAgentVersion constructs a new AgentVersion with the specific spec. The name will be created based on the type and
// version.
func NewAgentVersion(spec AgentVersionSpec) *AgentVersion {
	return &AgentVersion{
		ResourceMeta: ResourceMeta{
			APIVersion: version.V1,
			Kind:       KindAgentVersion,
			Metadata: Metadata{
				Name: fmt.Sprintf("%s-%s", spec.Type, spec.Version),
			},
		},
		Spec: spec,
	}
}

// ----------------------------------------------------------------------

// GetKind returns "AgentVersion"
func (v *AgentVersion) GetKind() Kind {
	return KindAgentVersion
}

// GetSpec returns the spec for this resource.
func (v *AgentVersion) GetSpec() any {
	return v.Spec
}

// AgentType returns the type of agent for this AgentVersion
func (v *AgentVersion) AgentType() string {
	return v.Spec.Type
}

// AgentVersion returns the version of the AgentVersion
func (v *AgentVersion) AgentVersion() string {
	return v.Spec.Version
}

// Public returns true if the version is not a draft or prerelease
func (v *AgentVersion) Public() bool {
	return !v.Spec.Draft && !v.Spec.Prerelease
}

// Installer returns the agent installer for the specified platform in the form os or os/arch.
func (v *AgentVersion) Installer(platform string) *AgentInstaller {
	if value, ok := v.Spec.Installer[platform]; ok {
		return &value
	}
	os := strings.Split(platform, "/")[0]
	if value, ok := v.Spec.Installer[os]; ok {
		return &value
	}
	return nil
}

// Download returns the agent download for the specified platform in the form os or os/arch.
func (v *AgentVersion) Download(platform string) *AgentDownload {
	if value, ok := v.Spec.Download[platform]; ok {
		return &value
	}
	os := strings.Split(platform, "/")[0]
	if value, ok := v.Spec.Download[os]; ok {
		return &value
	}
	return nil
}

// SemanticVersion returns a parsed semantic version that can be used to compare to other versions
func (v *AgentVersion) SemanticVersion() *semver.Version {
	return semver.Parse(v.AgentVersion())
}

// HashBytes returns the Hash of the download decoded as a byte array or nil if the hash is unspecified or invalid. This
// does not return an error because it is expected that errors will be detected in validation and an error in the hash
// can be treated as if there is no hash.
func (d *AgentDownload) HashBytes() []byte {
	hashBytes, _ := hex.DecodeString(d.Hash)
	return hashBytes
}

// ----------------------------------------------------------------------
// validation

// Validate ensures that each of the fields of an AgentVersion is valid. The name must equal "Type-Version"
func (v *AgentVersion) Validate() (warnings string, errors error) {
	errs := validation.NewErrors()
	v.validate(errs)
	return errs.Warnings(), errs.Result()
}

func (v *AgentVersion) validate(errs validation.Errors) {
	v.ResourceMeta.validate(errs)
	v.Spec.validate(v.Name(), errs)
}

func (vs *AgentVersionSpec) validate(name string, errs validation.Errors) {
	if vs.validateHasTypeAndVersion(errs) {
		// name must be equal to "Type-Version"
		if name != fmt.Sprintf("%s-%s", vs.Type, vs.Version) {
			errs.Add(fmt.Errorf("agent-version must have a name equal to [.spec.type]-[.spec.version] and name is %s but should be %s-%s", name, vs.Type, vs.Version))
		}
	}
	for platform, install := range vs.Installer {
		install.validate(platform, errs)
	}
	for platform, download := range vs.Download {
		download.validate(platform, errs)
	}
}

func (vs *AgentVersionSpec) validateHasTypeAndVersion(errs validation.Errors) bool {
	if vs.Type == "" {
		errs.Add(fmt.Errorf("agent-version must specify a type of agent as .spec.type"))
	}
	if vs.Version == "" {
		errs.Add(fmt.Errorf("agent-version must specify a version as .spec.version"))
	}
	return vs.Type != "" && vs.Version != ""
}

func (i *AgentInstaller) validate(platform string, errs validation.Errors) {
	name := fmt.Sprintf("%s install", platform)
	validateURL(name, i.URL, errs)
}

func (d *AgentDownload) validate(platform string, errs validation.Errors) {
	name := fmt.Sprintf("%s download", platform)
	validateURL(name, d.URL, errs)
	validateHash(name, d.Hash, errs)
}

func validateURL(name string, urlString string, errs validation.Errors) {
	if err := common.ValidateURL(urlString, []string{}); err != nil {
		errs.Add(fmt.Errorf("%s is invalid: %w", name, err))
	}
}

// validateHash validates that a hash is a hex encoded 256-bit hash. name will be used to identify the hash in errors
// messages.
func validateHash(name string, hash string, errs validation.Errors) {
	hashBytes, err := hex.DecodeString(hash)
	if err != nil {
		errs.Add(fmt.Errorf("%s hash must be a hex encoded 256-bit hash: %w", name, err))
		return
	}
	// 32 bytes * 8 == 256 bits
	if len(hashBytes) != 32 {
		errs.Add(fmt.Errorf("%s hash must be a hex encoded 256-bit hash: %d", name, len(hashBytes)))
	}
}

// ----------------------------------------------------------------------

// PrintableFieldTitles returns the list of field titles, used for printing a table of resources
func (v *AgentVersion) PrintableFieldTitles() []string {
	return []string{"Name", "Type", "Version", "Public", "Date", "URL"}
}

// PrintableFieldValue returns the field value for a title, used for printing a table of resources
func (v *AgentVersion) PrintableFieldValue(title string) string {
	switch title {
	case "Name":
		return v.Name()
	case "Type":
		return v.AgentType()
	case "Version":
		return v.AgentVersion()
	case "Public":
		return fmt.Sprintf("%t", v.Public())
	case "Date":
		if v.Spec.ReleaseDate != "" {
			if t, err := time.Parse(time.RFC3339, v.Spec.ReleaseDate); err == nil {
				return t.Format("01-02-2006")
			}
		}
		return "unknown"
	case "URL":
		return v.Spec.ReleaseNotesURL
	default:
		return v.ResourceMeta.PrintableFieldValue(title)
	}
}

// ----------------------------------------------------------------------
// sorting

type byAgentVersionSemver []*AgentVersion

func (s byAgentVersionSemver) Len() int {
	return len(s)
}
func (s byAgentVersionSemver) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byAgentVersionSemver) Less(i, j int) bool {
	return s[i].SemanticVersion().IsNewer(s[j].SemanticVersion())
}

// SortAgentVersionsLatestFirst sorts agent versions by their semantic versions, newest first
func SortAgentVersionsLatestFirst(agentVersions []*AgentVersion) {
	sort.Sort(byAgentVersionSemver(agentVersions))
}
