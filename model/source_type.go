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

import "github.com/observiq/bindplane-op/model/version"

type sourceTypeKind struct{}

func (k *sourceTypeKind) NewEmptyResource() *SourceType { return &SourceType{} }

// SourceType is a ResourceType used to define sources
type SourceType struct {
	ResourceType `yaml:",inline" json:",inline" mapstructure:",squash"`
}

// NewSourceType creates a new source-type with the specified name,
func NewSourceType(name string, parameters []ParameterDefinition, supportedPlatforms []string) *SourceType {
	return NewSourceTypeWithSpec(name, ResourceTypeSpec{
		Parameters:         parameters,
		SupportedPlatforms: supportedPlatforms,
	})
}

// NewSourceTypeWithSpec creates a new source-type with the specified name and spec.
func NewSourceTypeWithSpec(name string, spec ResourceTypeSpec) *SourceType {
	st := &SourceType{
		ResourceType: ResourceType{
			ResourceMeta: ResourceMeta{
				APIVersion: version.V1,
				Kind:       KindSourceType,
				Metadata: Metadata{
					Name: name,
				},
			},
			Spec: spec,
		},
	}
	st.EnsureMetadata(spec)
	return st
}

// GetKind returns "SourceType"
func (s *SourceType) GetKind() Kind {
	return KindSourceType
}

var sourceTypeTLSExemptions = map[string]bool{
	"aerospike":                false,
	"apache_combined":          true,
	"apache_common":            true,
	"apache_http":              false,
	"apache_spark":             true,
	"awscloudwatch":            true,
	"bigip":                    true,
	"cassandra":                true,
	"ciscoasa":                 true,
	"ciscocatalyst":            true,
	"ciscomeraki":              true,
	"cloudflare":               true,
	"cockroachdb":              false,
	"common_event_format":      true,
	"couchbase":                true,
	"couchdb":                  false,
	"csv":                      true,
	"custom":                   true,
	"elasticsearch":            false,
	"file":                     true,
	"fluentforward":            true,
	"hadoop":                   true,
	"hana":                     true,
	"haproxy":                  true,
	"hbase":                    true,
	"host":                     true,
	"iis":                      true,
	"jboss":                    true,
	"journald":                 true,
	"jvm":                      true,
	"k8s_cluster":              true,
	"k8s_container":            true,
	"k8s_events":               true,
	"k8s_kubelet":              true,
	"kafka_cluster":            true,
	"kafka_node":               true,
	"kafka_otlp_source":        true,
	"m365":                     true,
	"macOS":                    true,
	"microsoftactivedirectory": true,
	"mongodb":                  false,
	"mongodbatlas":             true, // only 'listen mode', different parameter keys
	"mysql":                    true,
	"netweaver":                false,
	"nginx":                    false,
	"oracledb":                 true,
	"otlp":                     true, // actually a listening server
	"pgbouncer":                true,
	"postgresql":               false,
	"prometheus":               false,
	"rabbitmq":                 false,
	"redis":                    false,
	"solr":                     true,
	"splunkhec":                true, // actually a listening server
	"splunk_tcp":               true,
	"sqlserver":                true,
	"statd":                    true,
	"syslog":                   true,
	"tcp":                      true, // actually a listening server
	"tomcat":                   true,
	"ubiquiti":                 true,
	"udp":                      true,
	"vmware_esxi":              true, // actually a listening syslog server
	"vmware_vcenter":           true, // mutual_tls and different parameter keys
	"w3c":                      true,
	"wildfly":                  true,
	"windowsdhcp":              true,
	"windowsevents":            true,
	"zookeeper":                true,
}
