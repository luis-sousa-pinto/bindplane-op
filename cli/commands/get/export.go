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

package get

import (
	"github.com/observiq/bindplane-op/model"
)

// exportConfiguration sanitizes a configuration for export by removing
// fields not suitable for export, such as the ID, hash, and version.
func exportConfiguration(c *model.Configuration) {
	if c == nil {
		return
	}

	// Cleanup metadata
	c.Metadata = sanitizeMetadataForExport(c.Metadata)

	// Cleanup sources
	if len(c.Spec.Sources) > 0 {
		sources := []model.ResourceConfiguration{}
		for _, source := range c.Spec.Sources {
			sources = append(sources, exportResourceConfiguration(source))
		}
		c.Spec.Sources = sources
	}

	// Cleanup destinations
	if len(c.Spec.Destinations) > 0 {
		destinations := []model.ResourceConfiguration{}
		for _, destination := range c.Spec.Destinations {
			destinations = append(destinations, exportResourceConfiguration(destination))
		}
		c.Spec.Destinations = destinations
	}

	// Remove the status
	c.Status = model.ConfigurationStatus{}
}

func exportResourceConfiguration(rc model.ResourceConfiguration) model.ResourceConfiguration {
	rc.Type = model.TrimVersion(rc.Type)
	rc.Name = model.TrimVersion(rc.Name)
	rc.ID = ""

	if len(rc.Processors) > 0 {
		proccesors := []model.ResourceConfiguration{}
		for _, p := range rc.Processors {
			proccesors = append(proccesors, exportResourceConfiguration(p))
		}
		rc.Processors = proccesors
	}

	return rc
}

// sanitizeMetadataForExport sanitizes a resource's for export.
func sanitizeMetadataForExport(m model.Metadata) model.Metadata {
	m.ID = ""
	m.Hash = ""
	m.Version = 0
	m.DateModified = nil
	return m
}
