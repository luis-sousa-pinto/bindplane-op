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

// Package model provides functions to convert models to GraphQL models
package model

import (
	"fmt"
	"strings"

	"github.com/observiq/bindplane-op/model"
	"github.com/observiq/bindplane-op/otlp/record"
	"github.com/observiq/bindplane-op/store"
	"github.com/observiq/bindplane-op/store/stats"
)

// ToAgentChangeArray converts store.AgentChanges to a []*AgentChange for use with graphql
func ToAgentChangeArray(changes store.Events[*model.Agent]) []*AgentChange {
	result := []*AgentChange{}
	for _, change := range changes {
		result = append(result, ToAgentChange(change))
	}
	return result
}

// ToAgentChange converts a store.AgentChange to use for use with graphql
func ToAgentChange(change store.Event[*model.Agent]) *AgentChange {
	agentChangeType := AgentChangeTypeInsert
	switch change.Type {
	case store.EventTypeInsert:
		agentChangeType = AgentChangeTypeInsert
	case store.EventTypeRemove:
		agentChangeType = AgentChangeTypeRemove
	default:
		agentChangeType = AgentChangeTypeUpdate
	}
	return &AgentChange{
		Agent:      change.Item,
		ChangeType: agentChangeType,
	}
}

// ToConfigurationChanges converts store.Events for Configuration to an array of ConfigurationChange for use with graphql
func ToConfigurationChanges(events store.Events[*model.Configuration]) []*ConfigurationChange {
	result := []*ConfigurationChange{}
	for _, event := range events {
		result = append(result, ToConfigurationChange(event))
	}
	return result
}

// ToConfigurationChange converts a store.Event for Configuration to a ConfigurationChange for use with graphql
func ToConfigurationChange(event store.Event[*model.Configuration]) *ConfigurationChange {
	return &ConfigurationChange{
		Configuration: event.Item,
		EventType:     ToEventType(event.Type),
	}
}

// ToEventType converts a store.EventType to a graphql EventType
func ToEventType(eventType store.EventType) EventType {
	switch eventType {
	case store.EventTypeInsert:
		return EventTypeInsert
	case store.EventTypeRemove:
		return EventTypeRemove
	}
	return EventTypeUpdate
}

// ToGraphMetric converts a Metric to a GraphMetric
func ToGraphMetric(m *record.Metric) (*GraphMetric, error) {
	// make sure this is a float64 value
	value, ok := stats.Value(m)
	if !ok {
		return nil, fmt.Errorf("bad value for metric %s", m.Name)
	}
	return &GraphMetric{
		Name:  strings.TrimPrefix(m.Name, "otelcol_processor_throughputmeasurement_"),
		Value: value,
		Unit:  m.Unit,
	}, nil
}

// ClearCurrentAgentUpgradeError clears the current agent upgrade error
func ClearCurrentAgentUpgradeError(cur *model.Agent) {
	if cur.Upgrade == nil {
		return
	}
	cur.Upgrade.Error = ""
}
