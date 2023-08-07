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

package agent

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/observiq/bindplane-op/model"
)

var (
	// oldestReleaseDate represents the oldest release of the bindplane agent that supports upgrade. when we
	// sync, we only include agents after this date to avoid adding a bunch of outdated agent versions.
	oldestReleaseDate = time.Date(2022, time.August, 1, 0, 0, 0, 0, time.UTC)
)

const (
	githubAPIUrl    = "https://api.github.com"
	githubRepo      = "bindplane-agent"
	githubRepoOwner = "observIQ"
)

type github struct {
	client *resty.Client
}

var _ VersionClient = (*github)(nil)

// newGithub creates a new github client for retrieving agent versions
func newGithub() *github {
	c := resty.New()
	c.SetTimeout(time.Second * 20)
	c.SetBaseURL(githubAPIUrl)
	return &github{
		client: c,
	}
}

type githubReleaseAsset struct {
	Name        string
	DownloadURL string `json:"browser_download_url"`
}

type githubRelease struct {
	Name            string
	TagName         string     `json:"tag_name"`
	ReleaseNotesURL string     `json:"html_url"`
	ReleaseDate     *time.Time `json:"published_at"`
	Draft           bool
	Prerelease      bool
	Assets          []githubReleaseAsset
}

func releasesURL() string {
	return fmt.Sprintf("/repos/%s/%s/releases", githubRepoOwner, githubRepo)
}
func latestURL() string {
	return fmt.Sprintf("/repos/%s/%s/releases/latest", githubRepoOwner, githubRepo)
}
func versionURL(version string) string {
	return fmt.Sprintf("/repos/%s/%s/releases/tags/%s", githubRepoOwner, githubRepo, version)
}

// LatestVersion returns the latest agent release.
func (g *github) LatestVersion() (*model.AgentVersion, error) {
	return g.Version(VersionLatest)
}

func (g *github) Version(version string) (*model.AgentVersion, error) {
	var url string
	if version == VersionLatest {
		url = latestURL()
	} else {
		url = versionURL(version)
	}

	var release githubRelease
	res, err := g.client.R().SetResult(&release).Get(url)

	if err != nil {
		return nil, fmt.Errorf("get release: %w", err)
	}
	if res.StatusCode() == 404 {
		return nil, ErrVersionNotFound
	}
	if res.StatusCode() != 200 {
		return nil, fmt.Errorf("unable to get version %s: %s", version, res.Status())
	}

	sums, err := g.GetSha256Sums(&release)
	if err != nil {
		return nil, fmt.Errorf("get sha256 sums: %w", err)
	}

	return convertRelease(&release, sums), nil
}

func (g *github) Versions() ([]*model.AgentVersion, error) {
	var releases []githubRelease
	res, err := g.client.R().SetResult(&releases).Get(releasesURL())

	if err != nil {
		return nil, fmt.Errorf("get releases: %w", err)
	}
	if res.StatusCode() != 200 {
		return nil, fmt.Errorf("unable to get versions: %s", res.Status())
	}

	var results []*model.AgentVersion
	for _, release := range releases {
		// see note about with oldestReleaseDate
		if release.ReleaseDate != nil && release.ReleaseDate.Before(oldestReleaseDate) {
			continue
		}

		r := release
		sums, err := g.GetSha256Sums(&r)
		if err != nil {
			return nil, fmt.Errorf("get sha256 sums: %w", err)
		}

		results = append(results, convertRelease(&r, sums))
	}

	return results, nil
}

func (g *github) GetSha256Sums(release *githubRelease) (Sha256sums, error) {
	// download and parse the sha256sums
	sumsName := fmt.Sprintf("observiq-otel-collector-%s-SHA256SUMS", release.TagName)
	sumsURL := releaseAssetURL(sumsName, release.Assets)

	res, err := g.client.R().Get(sumsURL)
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}
	if res.StatusCode() != 200 {
		return nil, fmt.Errorf("bad status code: %d", res.StatusCode())
	}
	return ParseSha256Sums(res.Body()), nil
}

// PlatformArtifactNames contains the names for a set of artifacts for a platform.
type PlatformArtifactNames struct {
	// DownloadPackageFormat is the format string for the name of the downloadable upgrade package.
	// It must be formatted for use with Sprintf(format, version)
	DownloadPackageFormat string
	// InstallerName is the name of the installer for this platform
	InstallerName string
}

// DownloadPackageName formats DownloadPackageFormat with the version, giving the name of the downloadable upgrade package.
func (p PlatformArtifactNames) DownloadPackageName(version string) string {
	return fmt.Sprintf(p.DownloadPackageFormat, version)
}

// PlatformArtifacts is a map of platform to the download package format and installer name.
var PlatformArtifacts = map[string]PlatformArtifactNames{
	"darwin/amd64": {
		DownloadPackageFormat: "observiq-otel-collector-%s-darwin-amd64.tar.gz",
		InstallerName:         "install_macos.sh",
	},
	"darwin/arm64": {
		DownloadPackageFormat: "observiq-otel-collector-%s-darwin-arm64.tar.gz",
		InstallerName:         "install_macos.sh",
	},
	"linux/amd64": {
		DownloadPackageFormat: "observiq-otel-collector-%s-linux-amd64.tar.gz",
		InstallerName:         "install_unix.sh",
	},
	"linux/arm64": {
		DownloadPackageFormat: "observiq-otel-collector-%s-linux-arm64.tar.gz",
		InstallerName:         "install_unix.sh",
	},
	"linux/arm": {
		DownloadPackageFormat: "observiq-otel-collector-%s-linux-arm.tar.gz",
		InstallerName:         "install_unix.sh",
	},
	"windows/amd64": {
		DownloadPackageFormat: "observiq-otel-collector-%s-windows-amd64.zip",
		InstallerName:         "observiq-otel-collector.msi",
	},
	"windows/386": {
		DownloadPackageFormat: "observiq-otel-collector-%s-windows-386.zip",
		InstallerName:         "observiq-otel-collector-x86.msi",
	},
}

func convertRelease(r *githubRelease, hashes Sha256sums) *model.AgentVersion {
	installer := map[string]model.AgentInstaller{}
	download := map[string]model.AgentDownload{}

	for platform, components := range PlatformArtifacts {
		downloadName := components.DownloadPackageName(r.TagName)
		installerName := components.InstallerName

		installer[platform] = model.AgentInstaller{
			URL: releaseAssetURL(installerName, r.Assets),
		}
		download[platform] = model.AgentDownload{
			URL:  releaseAssetURL(downloadName, r.Assets),
			Hash: hashes.Sha256Sum(downloadName),
		}
	}

	var releaseDate string
	if r.ReleaseDate != nil {
		releaseDate = r.ReleaseDate.UTC().Format(time.RFC3339)
	}

	return model.NewAgentVersion(model.AgentVersionSpec{
		Type:            string(model.AgentTypeNameObservIQOtelCollector),
		Version:         r.TagName,
		Prerelease:      r.Prerelease,
		Draft:           r.Draft,
		ReleaseNotesURL: r.ReleaseNotesURL,
		ReleaseDate:     releaseDate,
		Installer:       installer,
		Download:        download,
	})
}

func releaseAssetURL(name string, assets []githubReleaseAsset) string {
	for _, asset := range assets {
		if asset.Name == name {
			return asset.DownloadURL
		}
	}
	return ""
}
