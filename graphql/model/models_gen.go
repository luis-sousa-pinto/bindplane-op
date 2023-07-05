// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"

	"github.com/observiq/bindplane-op/model"
	"github.com/observiq/bindplane-op/model/graph"
	"github.com/observiq/bindplane-op/otlp/record"
	"github.com/observiq/bindplane-op/store/search"
)

type AgentChange struct {
	Agent      *model.Agent    `json:"agent"`
	ChangeType AgentChangeType `json:"changeType"`
}

type AgentConfiguration struct {
	Collector *string                `json:"Collector"`
	Logging   *string                `json:"Logging"`
	Manager   map[string]interface{} `json:"Manager"`
}

type Agents struct {
	Query         *string              `json:"query"`
	Agents        []*model.Agent       `json:"agents"`
	Suggestions   []*search.Suggestion `json:"suggestions"`
	LatestVersion string               `json:"latestVersion"`
}

type ClearAgentUpgradeErrorInput struct {
	AgentID string `json:"agentId"`
}

type ConfigurationChange struct {
	Configuration *model.Configuration `json:"configuration"`
	EventType     EventType            `json:"eventType"`
}

type Configurations struct {
	Query          *string                `json:"query"`
	Configurations []*model.Configuration `json:"configurations"`
	Suggestions    []*search.Suggestion   `json:"suggestions"`
}

type DestinationWithType struct {
	Destination     *model.Destination     `json:"destination"`
	DestinationType *model.DestinationType `json:"destinationType"`
}

type EditConfigurationDescriptionInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type GraphMetric struct {
	Name         string  `json:"name"`
	NodeID       string  `json:"nodeID"`
	PipelineType string  `json:"pipelineType"`
	Value        float64 `json:"value"`
	Unit         string  `json:"unit"`
	AgentID      *string `json:"agentID"`
}

type GraphMetrics struct {
	Metrics        []*GraphMetric `json:"metrics"`
	MaxMetricValue float64        `json:"maxMetricValue"`
	MaxLogValue    float64        `json:"maxLogValue"`
	MaxTraceValue  float64        `json:"maxTraceValue"`
}

type OverviewPage struct {
	Graph *graph.Graph `json:"graph"`
}

type RemoveAgentConfigurationInput struct {
	AgentID string `json:"agentId"`
}

type Snapshot struct {
	Logs    []*record.Log    `json:"logs"`
	Metrics []*record.Metric `json:"metrics"`
	Traces  []*record.Trace  `json:"traces"`
}

type UpdateProcessorsInput struct {
	Configuration string                         `json:"configuration"`
	ResourceType  ResourceTypeKind               `json:"resourceType"`
	ResourceIndex int                            `json:"resourceIndex"`
	Processors    []*model.ResourceConfiguration `json:"processors"`
}

type AgentChangeType string

const (
	AgentChangeTypeInsert AgentChangeType = "INSERT"
	AgentChangeTypeUpdate AgentChangeType = "UPDATE"
	AgentChangeTypeRemove AgentChangeType = "REMOVE"
)

var AllAgentChangeType = []AgentChangeType{
	AgentChangeTypeInsert,
	AgentChangeTypeUpdate,
	AgentChangeTypeRemove,
}

func (e AgentChangeType) IsValid() bool {
	switch e {
	case AgentChangeTypeInsert, AgentChangeTypeUpdate, AgentChangeTypeRemove:
		return true
	}
	return false
}

func (e AgentChangeType) String() string {
	return string(e)
}

func (e *AgentChangeType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = AgentChangeType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid AgentChangeType", str)
	}
	return nil
}

func (e AgentChangeType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type EventType string

const (
	EventTypeInsert EventType = "INSERT"
	EventTypeUpdate EventType = "UPDATE"
	EventTypeRemove EventType = "REMOVE"
)

var AllEventType = []EventType{
	EventTypeInsert,
	EventTypeUpdate,
	EventTypeRemove,
}

func (e EventType) IsValid() bool {
	switch e {
	case EventTypeInsert, EventTypeUpdate, EventTypeRemove:
		return true
	}
	return false
}

func (e EventType) String() string {
	return string(e)
}

func (e *EventType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = EventType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid EventType", str)
	}
	return nil
}

func (e EventType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type ParameterType string

const (
	ParameterTypeString                  ParameterType = "string"
	ParameterTypeStrings                 ParameterType = "strings"
	ParameterTypeInt                     ParameterType = "int"
	ParameterTypeBool                    ParameterType = "bool"
	ParameterTypeEnum                    ParameterType = "enum"
	ParameterTypeEnums                   ParameterType = "enums"
	ParameterTypeMap                     ParameterType = "map"
	ParameterTypeYaml                    ParameterType = "yaml"
	ParameterTypeTimezone                ParameterType = "timezone"
	ParameterTypeMetrics                 ParameterType = "metrics"
	ParameterTypeAwsCloudwatchNamedField ParameterType = "awsCloudwatchNamedField"
)

var AllParameterType = []ParameterType{
	ParameterTypeString,
	ParameterTypeStrings,
	ParameterTypeInt,
	ParameterTypeBool,
	ParameterTypeEnum,
	ParameterTypeEnums,
	ParameterTypeMap,
	ParameterTypeYaml,
	ParameterTypeTimezone,
	ParameterTypeMetrics,
	ParameterTypeAwsCloudwatchNamedField,
}

func (e ParameterType) IsValid() bool {
	switch e {
	case ParameterTypeString, ParameterTypeStrings, ParameterTypeInt, ParameterTypeBool, ParameterTypeEnum, ParameterTypeEnums, ParameterTypeMap, ParameterTypeYaml, ParameterTypeTimezone, ParameterTypeMetrics, ParameterTypeAwsCloudwatchNamedField:
		return true
	}
	return false
}

func (e ParameterType) String() string {
	return string(e)
}

func (e *ParameterType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = ParameterType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid ParameterType", str)
	}
	return nil
}

func (e ParameterType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type RelevantIfOperatorType string

const (
	RelevantIfOperatorTypeEquals      RelevantIfOperatorType = "equals"
	RelevantIfOperatorTypeNotEquals   RelevantIfOperatorType = "notEquals"
	RelevantIfOperatorTypeContainsAny RelevantIfOperatorType = "containsAny"
)

var AllRelevantIfOperatorType = []RelevantIfOperatorType{
	RelevantIfOperatorTypeEquals,
	RelevantIfOperatorTypeNotEquals,
	RelevantIfOperatorTypeContainsAny,
}

func (e RelevantIfOperatorType) IsValid() bool {
	switch e {
	case RelevantIfOperatorTypeEquals, RelevantIfOperatorTypeNotEquals, RelevantIfOperatorTypeContainsAny:
		return true
	}
	return false
}

func (e RelevantIfOperatorType) String() string {
	return string(e)
}

func (e *RelevantIfOperatorType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = RelevantIfOperatorType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid RelevantIfOperatorType", str)
	}
	return nil
}

func (e RelevantIfOperatorType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type ResourceTypeKind string

const (
	ResourceTypeKindSource      ResourceTypeKind = "SOURCE"
	ResourceTypeKindDestination ResourceTypeKind = "DESTINATION"
)

var AllResourceTypeKind = []ResourceTypeKind{
	ResourceTypeKindSource,
	ResourceTypeKindDestination,
}

func (e ResourceTypeKind) IsValid() bool {
	switch e {
	case ResourceTypeKindSource, ResourceTypeKindDestination:
		return true
	}
	return false
}

func (e ResourceTypeKind) String() string {
	return string(e)
}

func (e *ResourceTypeKind) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = ResourceTypeKind(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid ResourceTypeKind", str)
	}
	return nil
}

func (e ResourceTypeKind) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleUser   Role = "user"
	RoleViewer Role = "viewer"
)

var AllRole = []Role{
	RoleAdmin,
	RoleUser,
	RoleViewer,
}

func (e Role) IsValid() bool {
	switch e {
	case RoleAdmin, RoleUser, RoleViewer:
		return true
	}
	return false
}

func (e Role) String() string {
	return string(e)
}

func (e *Role) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = Role(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid Role", str)
	}
	return nil
}

func (e Role) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
