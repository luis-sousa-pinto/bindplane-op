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

// Package resources embeds the files in `resources` for seeding on startup
package resources

import "embed"

// Files contains the files embedded in resources/destination-types/*, resources/source-types/*, resources/processor-types/*, and resources/agent-versions/*
//
//go:embed destination-types/* source-types/* processor-types/* agent-versions/*
var Files embed.FS

// SeedFolders is the list of folders that we seed on startup
var SeedFolders = []string{
	"destination-types",
	"source-types",
	"processor-types",
	"agent-versions",
}
