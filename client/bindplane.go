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

// Package client provides a go client for interacting with the BindPlane OP server. Most of the functions depend on the
// BindPlane REST API internally.
package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"

	"github.com/observiq/bindplane-op/config"
	"github.com/observiq/bindplane-op/model"
	"github.com/observiq/bindplane-op/rest"
	"github.com/observiq/bindplane-op/version"
)

// AgentInstallOptions contains configuration options used for installing an agent.
type AgentInstallOptions struct {
	// Platform is the platform the agent will run on, e.g. "linux"
	Platform string

	// Version is the agent release version to install. Available release versions of the observiq-otel-collector are
	// available at [observiq-otel-collector Releases]
	//
	// [observiq-otel-collector Releases]: https://github.com/observIQ/observiq-otel-collector/releases
	Version string

	// Labels is a string representation of the agents labels, e.g. "platform=dev,os=windows,app=nginx"
	Labels string

	// SecretKey is the secret key used to authenticate agents with BindPlane OP
	SecretKey string

	// RemoteURL is the URL that the agent will use to connect to BindPlane OP
	RemoteURL string
}

// ----------------------------------------------------------------------

// QueryOptions represents the set of options available for a store query
type QueryOptions struct {
	Selector string
	Query    string
	Offset   int
	Limit    int
	Sort     string
}

// BindPlane is a REST client for BindPlane OP.
//
//go:generate mockery --name=BindPlane --filename=mock_bindplane.go --structname=MockBindPlane
type BindPlane interface {
	// Agents returns a list of Agents.
	Agents(ctx context.Context, options QueryOptions) ([]*model.Agent, error)
	// Agent returns a single Agent.
	Agent(ctx context.Context, id string) (*model.Agent, error)
	// DeleteAgents deletes multiple agents by ID.
	DeleteAgents(ctx context.Context, agentIDs []string) ([]*model.Agent, error)

	// AgentVersions returns a list of AgentVersion resources.
	AgentVersions(ctx context.Context) ([]*model.AgentVersion, error)
	// AgentVersion returns a single AgentVersion resources by name.
	AgentVersion(ctx context.Context, name string) (*model.AgentVersion, error)
	// DeleteAgentVersion deletes an AgentVersion resource by name.
	DeleteAgentVersion(ctx context.Context, name string) error

	// SyncAgentVersions builds agent-version from the release data in GitHub.
	// If version is empty, it syncs the last 10 releases.
	SyncAgentVersions(ctx context.Context, version string) ([]*model.AnyResourceStatus, error)

	// Configurations returns a list of Configuration resources.
	Configurations(ctx context.Context) ([]*model.Configuration, error)
	// Configuration returns a single Configuration resource from GET /v1/configurations/:name
	Configuration(ctx context.Context, name string) (*model.Configuration, error)
	// Delete configuration deletes a single configuration reseource.
	DeleteConfiguration(ctx context.Context, name string) error
	// RawConfiguration returns the raw OpenTelemetry configuration for the configuration with
	// the specified name. This can either be the raw value of a raw configuration or the
	// rendered value of a configuration with sources and destinations.
	RawConfiguration(ctx context.Context, name string) (string, error)
	// CopyConfig creates a deep copy of an existing resource under a new name.
	CopyConfig(ctx context.Context, name, copyName string) error

	// Sources returns a list of all Source resources.
	Sources(ctx context.Context) ([]*model.Source, error)
	// Source returns a single Source resource by name.
	Source(ctx context.Context, name string) (*model.Source, error)
	// DeleteSource deletes a single Source resource by name.
	DeleteSource(ctx context.Context, name string) error

	// SourceTypes returns a list of all SourceType resources.
	SourceTypes(ctx context.Context) ([]*model.SourceType, error)
	// SourceType returns a single SourceType resource by name.
	SourceType(ctx context.Context, name string) (*model.SourceType, error)
	// DeleteSourceType deletes a single SourceType resource by name.
	DeleteSourceType(ctx context.Context, name string) error

	// Processors returns a list of all Processor resources.
	Processors(ctx context.Context) ([]*model.Processor, error)
	// Processor returns a single Processor resource by name.
	Processor(ctx context.Context, name string) (*model.Processor, error)
	// DeleteProcessor deletes a single Processor resource by name.
	DeleteProcessor(ctx context.Context, name string) error

	// ProcessorTypes returns a list of all ProcessorType resources.
	ProcessorTypes(ctx context.Context) ([]*model.ProcessorType, error)
	// ProcessorType returns a single ProcessorType resource by name.
	ProcessorType(ctx context.Context, name string) (*model.ProcessorType, error)
	// DeleteProcessorType deletes a single ProcessorType resource by name.
	DeleteProcessorType(ctx context.Context, name string) error

	// Destinations returns a list of all Destination resources.
	Destinations(ctx context.Context) ([]*model.Destination, error)
	// Destination returns a single Destination resource by name.
	Destination(ctx context.Context, name string) (*model.Destination, error)
	// DeleteDestination deletes a single Destination resource by name.
	DeleteDestination(ctx context.Context, name string) error

	// DestinationTypes returns a list of all DestinationType resources.
	DestinationTypes(ctx context.Context) ([]*model.DestinationType, error)
	// DestinationType returns a single DestinationType by name.
	DestinationType(ctx context.Context, name string) (*model.DestinationType, error)
	// DeleteDestinationType deletes a single Destination resource by name.
	DeleteDestinationType(ctx context.Context, name string) error

	// Apply upserts multiple resources of any kind.
	Apply(ctx context.Context, r []*model.AnyResource) ([]*model.AnyResourceStatus, error)
	// Delete deletes multiple resources, minimum required fields to delete are Kind and Metadata.Name.
	Delete(ctx context.Context, r []*model.AnyResource) ([]*model.AnyResourceStatus, error)

	// Version returns the version of the BindPlane-OP server.
	Version(ctx context.Context) (version.Version, error)

	// AgentInstallCommand returns the installation command for the given AgentInstallationOptions.
	AgentInstallCommand(ctx context.Context, options AgentInstallOptions) (string, error)
	// AgentUpgrade upgrades the agent with given ID to the specified version.
	AgentUpgrade(ctx context.Context, id string, version string) error

	// AgentLabels gets the labels for an agent
	AgentLabels(ctx context.Context, id string) (*model.Labels, error)
	// ApplyAgentLabels applies the specified labels to an agent, merging the specified labels with the existing labels
	// and returning the labels of the agent
	ApplyAgentLabels(ctx context.Context, id string, labels *model.Labels, override bool) (*model.Labels, error)

	// Rollouts

	// RolloutStatus returns the status of a rollout
	RolloutStatus(ctx context.Context, name string) (*model.Configuration, error)

	// StartRollout starts a rollout that is pending
	StartRollout(ctx context.Context, name string, options *model.RolloutOptions) (*model.Configuration, error)

	// PauseRollout pauses a rollout that is started
	PauseRollout(ctx context.Context, name string) (*model.Configuration, error)

	// ResumeRollout resumes a rollout that is paused
	ResumeRollout(ctx context.Context, name string) (*model.Configuration, error)

	// UpdateRollout updates a rollout
	UpdateRollout(ctx context.Context, name string) (*model.Configuration, error)

	// UpdateRollouts updates all active rollouts
	UpdateRollouts(ctx context.Context) ([]*model.Configuration, error)

	// ResourceHistory retrieves the history of the rollout
	ResourceHistory(ctx context.Context, kind model.Kind, name string) ([]*model.AnyResource, error)
}

// BindplaneClient is the implementation of the Bindplane interface
type BindplaneClient struct {
	Client *resty.Client
	*zap.Logger
}

// NewBindPlane takes a client configuration, logger and returns a new BindPlane.
func NewBindPlane(config *config.Config, logger *zap.Logger) (BindPlane, error) {
	client := resty.New()
	// Don't log warning if using HTTP
	client.SetDisableWarn(true)
	client.SetTimeout(time.Second * 20)
	client.SetBasicAuth(config.Auth.Username, config.Auth.Password)
	client.SetBaseURL(fmt.Sprintf("%s/v1", config.Network.ServerURL()))

	tlsConfig, err := config.Network.Convert()
	if err != nil {
		return nil, fmt.Errorf("failed to configure TLS client: %w", err)
	}
	client.SetTLSClientConfig(tlsConfig)

	return &BindplaneClient{
		Client: client,
		Logger: logger.Named("bindplane-client"),
	}, nil
}

// Agents retries agents based on the query
func (c *BindplaneClient) Agents(_ context.Context, options QueryOptions) ([]*model.Agent, error) {
	c.Debug("Agents called")

	ar := &model.AgentsResponse{}
	resp, err := c.Client.R().
		SetResult(ar).
		SetQueryParam("selector", options.Selector).
		SetQueryParam("query", options.Query).
		SetQueryParam("offset", fmt.Sprintf("%d", options.Offset)).
		SetQueryParam("limit", fmt.Sprintf("%d", options.Limit)).
		SetQueryParam("sort", options.Sort).
		Get("/agents")
	if err != nil {
		LogRequestError(c.Logger, err, "/agents")
		return nil, err
	}

	return ar.Agents, c.StatusError(resp, err, "unable to get agents")
}

// Agent returns the agent with the id
func (c *BindplaneClient) Agent(_ context.Context, id string) (*model.Agent, error) {
	c.Debug("Agent called")

	ar := &model.AgentResponse{}
	agentsEndpoint := fmt.Sprintf("/agents/%s", id)
	resp, err := c.Client.R().SetResult(ar).Get(agentsEndpoint)
	if err != nil {
		LogRequestError(c.Logger, err, agentsEndpoint)
		return nil, err
	}

	return ar.Agent, c.StatusError(resp, err, "unable to get agents")
}

// DeleteAgents deletes agents with the ids
func (c *BindplaneClient) DeleteAgents(_ context.Context, ids []string) ([]*model.Agent, error) {
	c.Debug("DeleteAgents called")

	body := &model.DeleteAgentsPayload{
		IDs: ids,
	}
	result := &model.DeleteAgentsResponse{}
	resp, err := c.Client.R().SetBody(body).SetResult(result).Delete("/agents")
	return result.Agents, c.StatusError(resp, err, "unable to delete agents")
}

// AgentVersions retries all gent versions
func (c *BindplaneClient) AgentVersions(ctx context.Context) ([]*model.AgentVersion, error) {
	result := model.AgentVersionsResponse{}
	err := c.Resources(ctx, "/agent-versions", &result)
	return result.AgentVersions, err
}

// AgentVersion retrieves the agent version with name
func (c *BindplaneClient) AgentVersion(ctx context.Context, name string) (*model.AgentVersion, error) {
	result := model.AgentVersionResponse{}
	err := c.Resource(ctx, "/agent-versions", name, &result)
	return result.AgentVersion, err
}

// DeleteAgentVersion deletes the agent version with name
func (c *BindplaneClient) DeleteAgentVersion(ctx context.Context, name string) error {
	return c.DeleteResource(ctx, "/agent-versions", name)
}

// SyncAgentVersions syncs the specific agent version
func (c *BindplaneClient) SyncAgentVersions(_ context.Context, version string) ([]*model.AnyResourceStatus, error) {
	ar := &model.ApplyResponseClientSide{}
	resp, err := c.Client.R().
		SetHeader("Content-Type", "application/json").
		SetResult(ar).
		Post(fmt.Sprintf("/agent-versions/%s/sync", version))
	if err != nil {
		LogRequestError(c.Logger, err, "/agent-versions/:name/sync")
		return nil, err
	}
	return ar.Updates, c.StatusError(resp, err, "unable to sync agent-versions")
}

// Configurations retrieves all configurations
func (c *BindplaneClient) Configurations(_ context.Context) ([]*model.Configuration, error) {
	c.Debug("Configurations called")

	pr := &model.ConfigurationsResponse{}
	resp, err := c.Client.R().SetResult(pr).Get("/configurations")
	return pr.Configurations, c.StatusError(resp, err, "unable to get configurations")
}

// Configuration retrieves configuration with name
func (c *BindplaneClient) Configuration(ctx context.Context, name string) (*model.Configuration, error) {
	result := model.ConfigurationResponse{}
	err := c.Resource(ctx, "/configurations", name, &result)
	return result.Configuration, err
}

// DeleteConfiguration deletes the configuration with name
func (c *BindplaneClient) DeleteConfiguration(ctx context.Context, name string) error {
	return c.DeleteResource(ctx, "/configurations", name)
}

// RawConfiguration retrieves the raw config with name
func (c *BindplaneClient) RawConfiguration(ctx context.Context, name string) (string, error) {
	result := model.ConfigurationResponse{}
	err := c.Resource(ctx, "/configurations", name, &result)
	return result.Raw, err
}

// Sources retrieves all sources
func (c *BindplaneClient) Sources(ctx context.Context) ([]*model.Source, error) {
	result := model.SourcesResponse{}
	err := c.Resources(ctx, "/sources", &result)
	return result.Sources, err
}

// Source retrieves the source with the given name
func (c *BindplaneClient) Source(ctx context.Context, name string) (*model.Source, error) {
	result := model.SourceResponse{}
	err := c.Resource(ctx, "/sources", name, &result)
	return result.Source, err
}

// DeleteSource deletes the source with the given name
func (c *BindplaneClient) DeleteSource(ctx context.Context, name string) error {
	return c.DeleteResource(ctx, "/sources", name)
}

// SourceTypes retrieves all source types
func (c *BindplaneClient) SourceTypes(ctx context.Context) ([]*model.SourceType, error) {
	result := model.SourceTypesResponse{}
	err := c.Resources(ctx, "/source-types", &result)
	return result.SourceTypes, err
}

// SourceType retrieves source type with given name
func (c *BindplaneClient) SourceType(ctx context.Context, name string) (*model.SourceType, error) {
	result := model.SourceTypeResponse{}
	err := c.Resource(ctx, "/source-types", name, &result)
	return result.SourceType, err
}

// DeleteSourceType deletes source type with given name
func (c *BindplaneClient) DeleteSourceType(ctx context.Context, name string) error {
	return c.DeleteResource(ctx, "/source-types", name)
}

// Processors retrieves all processors
func (c *BindplaneClient) Processors(ctx context.Context) ([]*model.Processor, error) {
	result := model.ProcessorsResponse{}
	err := c.Resources(ctx, "/processors", &result)
	return result.Processors, err
}

// Processor retrieves processor with given name
func (c *BindplaneClient) Processor(ctx context.Context, name string) (*model.Processor, error) {
	result := model.ProcessorResponse{}
	err := c.Resource(ctx, "/processors", name, &result)
	return result.Processor, err
}

// DeleteProcessor deletes the processor with the given name
func (c *BindplaneClient) DeleteProcessor(ctx context.Context, name string) error {
	return c.DeleteResource(ctx, "/processors", name)
}

// ProcessorTypes retrieves all processor types
func (c *BindplaneClient) ProcessorTypes(ctx context.Context) ([]*model.ProcessorType, error) {
	result := model.ProcessorTypesResponse{}
	err := c.Resources(ctx, "/processor-types", &result)
	return result.ProcessorTypes, err
}

// ProcessorType retrieves processor type with given name
func (c *BindplaneClient) ProcessorType(ctx context.Context, name string) (*model.ProcessorType, error) {
	result := model.ProcessorTypeResponse{}
	err := c.Resource(ctx, "/processor-types", name, &result)
	return result.ProcessorType, err
}

// DeleteProcessorType deletes processor type with given name
func (c *BindplaneClient) DeleteProcessorType(ctx context.Context, name string) error {
	return c.DeleteResource(ctx, "/processor-types", name)
}

// Destinations retrieves all destinations
func (c *BindplaneClient) Destinations(ctx context.Context) ([]*model.Destination, error) {
	result := model.DestinationsResponse{}
	err := c.Resources(ctx, "/destinations", &result)
	return result.Destinations, err
}

// Destination retrieves destination with given name
func (c *BindplaneClient) Destination(ctx context.Context, name string) (*model.Destination, error) {
	result := model.DestinationResponse{}
	err := c.Resource(ctx, "/destinations", name, &result)
	return result.Destination, err
}

// DeleteDestination deletes destination with given name
func (c *BindplaneClient) DeleteDestination(ctx context.Context, name string) error {
	return c.DeleteResource(ctx, "/destinations", name)
}

// DestinationTypes retrieves all destination types
func (c *BindplaneClient) DestinationTypes(ctx context.Context) ([]*model.DestinationType, error) {
	result := model.DestinationTypesResponse{}
	err := c.Resources(ctx, "/destination-types", &result)
	return result.DestinationTypes, err
}

// DestinationType retrieves destination type with given name
func (c *BindplaneClient) DestinationType(ctx context.Context, name string) (*model.DestinationType, error) {
	result := model.DestinationTypeResponse{}
	err := c.Resource(ctx, "/destination-types", name, &result)
	return result.DestinationType, err
}

// DeleteDestinationType deletes destination type with given name
func (c *BindplaneClient) DeleteDestinationType(ctx context.Context, name string) error {
	return c.DeleteResource(ctx, "/destination-types", name)
}

// Apply apply resources
func (c *BindplaneClient) Apply(_ context.Context, resources []*model.AnyResource) ([]*model.AnyResourceStatus, error) {
	c.Debug("Apply called")

	payload := model.ApplyPayload{
		Resources: resources,
	}

	data, err := jsoniter.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("client apply: %w", err)
	}

	ar := &model.ApplyResponseClientSide{}
	resp, err := c.Client.R().SetHeader("Content-Type", "application/json").
		SetBody(data).SetResult(ar).Post("/apply")
	return ar.Updates, c.StatusError(resp, err, "unable to apply resources")
}

// Delete deletes passed in resources
func (c *BindplaneClient) Delete(_ context.Context, resources []*model.AnyResource) ([]*model.AnyResourceStatus, error) {
	c.Debug("Batch Delete called")

	payload := model.DeletePayload{
		Resources: resources,
	}

	data, err := jsoniter.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling data to json: %w", err)
	}

	resp, err := c.Client.R().SetHeader("Content-Type", "application/json").
		SetBody(data).Post("/delete")
	if err != nil {
		LogRequestError(c.Logger, err, "/delete")
		return nil, err
	}

	dr := &model.DeleteResponseClientSide{}

	switch resp.StatusCode() {
	case http.StatusAccepted:
		return dr.Updates, nil
	case http.StatusUnauthorized:
		return nil, c.UnauthorizedError(resp)
	case http.StatusBadRequest:
		if dr.Errors != nil {
			return nil, errors.New(dr.Errors[0])
		}
		return nil, errors.New("bad request")
	case http.StatusInternalServerError:
		return nil, fmt.Errorf("%s", dr.Errors[0])
	}

	err = jsoniter.Unmarshal(resp.Body(), dr)
	if err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("unknown response from bindplane server")
}

// Version retrieves bindplane server version
func (c *BindplaneClient) Version(_ context.Context) (version.Version, error) {
	c.Debug("Version called")

	v := version.Version{}
	resp, err := c.Client.R().SetResult(&v).Get("/version")
	return v, c.StatusError(resp, err, "unable to get version")
}

// AgentInstallCommand returns the agent install command based on the install options
func (c *BindplaneClient) AgentInstallCommand(_ context.Context, options AgentInstallOptions) (string, error) {
	c.Debug("AgentInstallCommand called")

	var command model.InstallCommandResponse
	endpoint := fmt.Sprintf("/agent-versions/%s/install-command", options.Version)

	resp, err := c.Client.R().
		SetQueryParam("platform", options.Platform).
		SetQueryParam("version", options.Version).
		SetQueryParam("labels", options.Labels).
		SetQueryParam("remote-url", options.RemoteURL).
		SetQueryParam("secret-key", options.SecretKey).
		SetResult(&command).
		Get(endpoint)

	return command.Command, c.StatusError(resp, err, "unable to get install command")
}

// AgentUpgrade sends a request to upgrade agent with id to version
func (c *BindplaneClient) AgentUpgrade(_ context.Context, id string, version string) error {
	endpoint := fmt.Sprintf("/agents/%s/version", id)
	resp, err := c.Client.R().
		SetBody(model.PostAgentVersionRequest{
			Version: version,
		}).
		Post(endpoint)

	if err != nil {
		return err
	}

	// look for errors
	if resp.StatusCode() != http.StatusNoContent {
		errResponse := &model.ErrorResponse{}
		err = jsoniter.Unmarshal(resp.Body(), errResponse)
		if err != nil {
			return fmt.Errorf("unable to parse api response: %w", err)
		}

		if len(errResponse.Errors) > 0 {
			var errs error
			for _, e := range errResponse.Errors {
				errs = errors.Join(errs, errors.New(e))
			}
			return errs
		}
	}

	return nil
}

// AgentLabels retrieves labels for agent with id
func (c *BindplaneClient) AgentLabels(_ context.Context, id string) (*model.Labels, error) {
	var response model.AgentLabelsResponse
	endpoint := fmt.Sprintf("/agents/%s/labels", id)

	resp, err := c.Client.R().
		SetResult(&response).
		Get(endpoint)

	return response.Labels, c.StatusError(resp, err, "unable to get agent labels")
}

// ApplyAgentLabels apply labels to agent with id
func (c *BindplaneClient) ApplyAgentLabels(_ context.Context, id string, labels *model.Labels, overwrite bool) (*model.Labels, error) {
	payload := model.AgentLabelsPayload{
		Labels: labels.AsMap(),
	}

	data, err := jsoniter.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal labels to apply: %w", err)
	}

	endpoint := fmt.Sprintf("/agents/%s/labels", id)
	resp, err := c.Client.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParam("overwrite", strconv.FormatBool(overwrite)).
		SetBody(data).
		Patch(endpoint)

	if resp.StatusCode() != http.StatusConflict {
		err = c.StatusError(resp, err, "unable to apply labels")
		if err != nil {
			return nil, err
		}
	}

	var response model.AgentLabelsResponse
	err = jsoniter.Unmarshal(resp.Body(), &response)
	if err != nil {
		return nil, fmt.Errorf("unable to parse api response: %w", err)
	}

	if response.Errors != nil {
		err = fmt.Errorf(strings.Join(response.Errors, "\n"))
	}

	return response.Labels, err
}

// CopyConfig copies config with name and gives the new config copyName
func (c *BindplaneClient) CopyConfig(_ context.Context, name, copyName string) error {
	payload := model.PostCopyConfigRequest{
		Name: copyName,
	}

	endpoint := fmt.Sprintf("/configurations/%s/copy", name)

	resp, err := c.Client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		Post(endpoint)

	if err != nil {
		return err
	}

	switch resp.StatusCode() {
	case http.StatusCreated:
		return nil
	case http.StatusConflict:
		return fmt.Errorf("a configuration with name '%s' already exists", copyName)
	default:
		errs := fmt.Errorf("failed to copy configuration, got status %v", resp.StatusCode())

		// check for errors field in response
		errResponse := &model.ErrorResponse{}
		if err := jsoniter.Unmarshal(resp.Body(), errResponse); err != nil {
			c.Logger.Error("failed to unmarshal error response when copying config", zap.Error(err))
		}

		if len(errResponse.Errors) > 0 {
			for _, e := range errResponse.Errors {
				errs = errors.Join(err, errors.New(e))
			}
		}

		return errs
	}
}

// StartRollout starts a rollout that is pending
func (c *BindplaneClient) StartRollout(ctx context.Context, name string, options *model.RolloutOptions) (*model.Configuration, error) {
	var response model.ConfigurationResponse
	endpoint := fmt.Sprintf("/rollouts/%s/start", name)

	body := model.StartRolloutPayload{}
	if options != nil {
		body.Options = options
	}

	resp, err := c.Client.R().
		SetContext(ctx).
		SetResult(&response).
		SetBody(body).
		Post(endpoint)

	return response.Configuration, c.StatusError(resp, err, "unable to start")
}

// RolloutStatus returns the status of a rollout
func (c *BindplaneClient) RolloutStatus(ctx context.Context, name string) (*model.Configuration, error) {
	var response model.ConfigurationResponse
	endpoint := fmt.Sprintf("/rollouts/%s/status", name)

	resp, err := c.Client.R().
		SetContext(ctx).
		SetResult(&response).
		Get(endpoint)

	return response.Configuration, c.StatusError(resp, err, "unable to get rollout status")
}

// PauseRollout pauses a rollout that is started
func (c *BindplaneClient) PauseRollout(ctx context.Context, name string) (*model.Configuration, error) {
	var response model.ConfigurationResponse
	endpoint := fmt.Sprintf("/rollouts/%s/pause", name)

	resp, err := c.Client.R().
		SetContext(ctx).
		SetResult(&response).
		Put(endpoint)

	return response.Configuration, c.StatusError(resp, err, "unable to pause")
}

// ResumeRollout resumes a rollout that is paused
func (c *BindplaneClient) ResumeRollout(ctx context.Context, name string) (*model.Configuration, error) {
	var response model.ConfigurationResponse
	endpoint := fmt.Sprintf("/rollouts/%s/resume", name)

	resp, err := c.Client.R().
		SetContext(ctx).
		SetResult(&response).
		Put(endpoint)

	return response.Configuration, c.StatusError(resp, err, "unable to resume")
}

// UpdateRollout updates a rollout
func (c *BindplaneClient) UpdateRollout(ctx context.Context, name string) (*model.Configuration, error) {
	var response model.ConfigurationResponse
	endpoint := fmt.Sprintf("/rollouts/%s/update", name)

	resp, err := c.Client.R().
		SetContext(ctx).
		SetResult(&response).
		Post(endpoint)

	return response.Configuration, c.StatusError(resp, err, "unable to update")
}

// UpdateRollouts updates all active rollouts
func (c *BindplaneClient) UpdateRollouts(ctx context.Context) ([]*model.Configuration, error) {
	var response model.ConfigurationsResponse
	endpoint := fmt.Sprintf("/rollouts")

	resp, err := c.Client.R().
		SetContext(ctx).
		SetResult(&response).
		Post(endpoint)

	return response.Configurations, c.StatusError(resp, err, "unable to update")
}

// ResourceHistory retrieves the history of the rollout
func (c *BindplaneClient) ResourceHistory(ctx context.Context, kind model.Kind, name string) ([]*model.AnyResource, error) {
	var response model.HistoryResponse
	endpoint := fmt.Sprintf("/%s/%s/history", kind, name)

	resp, err := c.Client.R().
		SetContext(ctx).
		SetResult(&response).
		Get(endpoint)

	return response.Versions, c.StatusError(resp, err, "unable to get resource history")
}

// ----------------------------------------------------------------------

// Resources gets the Resources from the REST server and stores them in the provided result.
func (c *BindplaneClient) Resources(ctx context.Context, resourcesURL string, result any) error {
	return c.get(ctx, resourcesURL, result)
}

// Resource gets the Resource with the specified name from the REST server and stores it in the provided result.
func (c *BindplaneClient) Resource(ctx context.Context, resourcesURL string, name string, result any) error {
	resourceURL := fmt.Sprintf("%s/%s", resourcesURL, name)
	return c.get(ctx, resourceURL, result)
}

func (c *BindplaneClient) get(ctx context.Context, url string, result any) error {
	resp, err := c.Client.R().
		SetContext(ctx).
		SetResult(result).
		Get(url)

	if err != nil {
		LogRequestError(c.Logger, err, url)
		return err
	}

	return c.StatusError(resp, err, fmt.Sprintf("unable to get %s", url))
}

// DeleteResource deletes the resource at the URL of name
func (c *BindplaneClient) DeleteResource(_ context.Context, resourcesURL string, name string) error {
	deleteEndpoint := fmt.Sprintf("%s/%s", resourcesURL, name)
	resp, err := c.Client.R().Delete(deleteEndpoint)
	if err != nil {
		LogRequestError(c.Logger, err, deleteEndpoint)
		return fmt.Errorf("error making request to remote bindplane server, %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusNoContent:
		return nil
	case http.StatusUnauthorized:
		return c.UnauthorizedError(resp)
	case http.StatusNotFound:
		return fmt.Errorf("%s not found", deleteEndpoint)
	case http.StatusBadRequest:
		dr := &model.DeleteResponse{}
		err = jsoniter.Unmarshal(resp.Body(), dr)
		if err != nil {
			return err
		}

		if dr.Errors != nil {
			return errors.New(dr.Errors[0])
		}

		return errors.New("bad request")
	case http.StatusConflict:
		errorResponse := &rest.ErrorResponse{}
		err = jsoniter.Unmarshal(resp.Body(), errorResponse)
		if err != nil {
			return errors.New("failed to parse response, status 409 Conflict")
		}

		if errorResponse.Errors != nil {
			return errors.New(errorResponse.Errors[0])
		}

		return errors.New("got status 409 Conflict")
	default:
		c.Logger.Error("unexpected status code received while trying to delete resource", zap.Int("statusCode", resp.StatusCode()), zap.String("endpoint", deleteEndpoint))
		return fmt.Errorf("unexpected status code received while trying to delete resource '%s': %s", name, resp.Status())
	}

}

// UnauthorizedError checks if response is Unauthorized error and returns error if it is
func (c *BindplaneClient) UnauthorizedError(resp *resty.Response) error {
	if resp.StatusCode() == http.StatusUnauthorized {
		err := fmt.Errorf(resp.Status())
		LogRequestError(c.Logger, err, resp.Request.URL)
		return err
	}
	return nil
}

// StatusError returns and error if resp is not 2XX or err is not nil
func (c *BindplaneClient) StatusError(resp *resty.Response, err error, message string) error {
	if err != nil {
		LogRequestError(c.Logger, err, resp.Request.URL)
		return err
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		return nil
	case http.StatusCreated:
		return nil
	case http.StatusAccepted:
		return nil
	case http.StatusNoContent:
		return nil

	default:
		err := fmt.Errorf("%s, got %s", message, resp.Status())
		LogRequestError(c.Logger, err, resp.Request.URL)
		return err
	}
}

// LogRequestError logs the error for a request against the endpoint
func LogRequestError(logger *zap.Logger, err error, endpoint string) {
	logger.Error("Error making request", zap.Error(err), zap.String("endpoint", endpoint))
}
