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

// Package semver provides a model and functions for working with semantic versions.
package semver

import (
	"fmt"
	"regexp"
	"strconv"
)

// Version represents a very simple semantic version that consists of three
// parts: major, minor, and patch and represents them separated by periods.
type Version struct {
	Major, Minor, Patch int
}

func (v *Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// IsOlder returns true if v is older than o
func (v *Version) IsOlder(o *Version) bool {
	return v.Compare(o) < 0
}

// IsNewer returns true if v is newer than o
func (v *Version) IsNewer(o *Version) bool {
	return v.Compare(o) > 0
}

// Equals returns true if v is equal to o
func (v *Version) Equals(o *Version) bool {
	return v.Compare(o) == 0
}

// Compare returns a value less than 0 if v1 < v2, 0 if v1 == v2, and greater
// than 0 if v1 > v2
func (v *Version) Compare(o *Version) int {
	diff := v.Major - o.Major
	if diff != 0 {
		return diff
	}
	diff = v.Minor - o.Minor
	if diff != 0 {
		return diff
	}
	return v.Patch - o.Patch
}

var versionRegexp = regexp.MustCompile("([0-9]+)(.([0-9]+))?(.([0-9]+))?")

// New returns a new Version with the specified major, minor, and patch
// components
func New(major, minor, patch int) *Version {
	return &Version{
		Major: major,
		Minor: minor,
		Patch: patch,
	}
}

// Parse parses a version from a string. It matches a very simple regexp
// in the string and is very lenient. If no version text found, the version will
// be 0.0.0.
func Parse(version string) *Version {
	major, minor, patch := 0, 0, 0

	matches := versionRegexp.FindStringSubmatch(version)
	switch l := len(matches); {
	case l > 1:
		major, _ = strconv.Atoi(matches[1])
		fallthrough
	case l > 3:
		minor, _ = strconv.Atoi(matches[3])
		fallthrough
	case l > 5:
		patch, _ = strconv.Atoi(matches[5])
	}

	return &Version{major, minor, patch}
}
