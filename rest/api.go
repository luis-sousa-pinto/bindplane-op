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

// Package rest provides a HTTP API for interacting with BindPlane
package rest

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"

	"github.com/observiq/bindplane-op/model"
	exposedserver "github.com/observiq/bindplane-op/server"
	"github.com/observiq/bindplane-op/store"
	"github.com/observiq/bindplane-op/store/search"
	"github.com/observiq/bindplane-op/version"
)

var tracer = otel.Tracer("rest")

// AddRestRoutes adds all API routes to the gin HTTP router
func AddRestRoutes(router gin.IRouter, bindplane exposedserver.BindPlane) {
	router.GET("/agents", func(c *gin.Context) { Agents(c, bindplane) })
	router.GET("/agents/:id", func(c *gin.Context) { GetAgent(c, bindplane) })
	router.DELETE("/agents", func(c *gin.Context) { DeleteAgents(c, bindplane) })
	router.PATCH("/agents/labels", func(c *gin.Context) { LabelAgents(c, bindplane) })
	router.GET("/agents/:id/labels", func(c *gin.Context) { GetAgentLabels(c, bindplane) })
	router.PATCH("/agents/:id/labels", func(c *gin.Context) { PatchAgentLabels(c, bindplane) })
	router.PUT("/agents/:id/restart", func(c *gin.Context) { RestartAgent(c, bindplane) })
	router.POST("/agents/:id/version", func(c *gin.Context) { UpgradeAgent(c, bindplane) })
	router.PATCH("/agents/version", func(c *gin.Context) { UpgradeAgents(c, bindplane) })
	router.GET("/agents/:id/configuration", func(c *gin.Context) { GetAgentConfiguration(c, bindplane) })

	router.GET("/agent-versions", func(c *gin.Context) { AgentVersions(c, bindplane) })
	router.GET("/agent-versions/:name", func(c *gin.Context) { AgentVersion(c, bindplane) })
	router.DELETE("/agent-versions/:name", func(c *gin.Context) { DeleteAgentVersion(c, bindplane) })
	router.GET("/agent-versions/:name/install-command", func(c *gin.Context) { getInstallCommand(c, bindplane) })
	router.POST("/agent-versions/:name/sync", func(c *gin.Context) { SyncAgentVersion(c, bindplane) })

	router.GET("/configurations", func(c *gin.Context) { Configurations(c, bindplane) })
	router.GET("/configurations/:name", func(c *gin.Context) { Configuration(c, bindplane) })
	router.DELETE("/configurations/:name", func(c *gin.Context) { DeleteConfiguration(c, bindplane) })
	router.POST("/configurations/:name/copy", func(c *gin.Context) { CopyConfig(c, bindplane) })

	router.GET("/sources", func(c *gin.Context) { Sources(c, bindplane) })
	router.GET("/sources/:name", func(c *gin.Context) { Source(c, bindplane) })
	router.DELETE("/sources/:name", func(c *gin.Context) { DeleteSource(c, bindplane) })

	router.GET("/source-types", func(c *gin.Context) { SourceTypes(c, bindplane) })
	router.GET("/source-types/:name", func(c *gin.Context) { SourceType(c, bindplane) })
	router.DELETE("/source-types/:name", func(c *gin.Context) { DeleteSourceType(c, bindplane) })

	router.GET("/processors", func(c *gin.Context) { Processors(c, bindplane) })
	router.GET("/processors/:name", func(c *gin.Context) { Processor(c, bindplane) })
	router.DELETE("/processors/:name", func(c *gin.Context) { DeleteProcessor(c, bindplane) })

	router.GET("/processor-types", func(c *gin.Context) { ProcessorTypes(c, bindplane) })
	router.GET("/processor-types/:name", func(c *gin.Context) { ProcessorType(c, bindplane) })
	router.DELETE("/processor-types/:name", func(c *gin.Context) { DeleteProcessorType(c, bindplane) })

	router.GET("/destinations", func(c *gin.Context) { Destinations(c, bindplane) })
	router.GET("/destinations/:name", func(c *gin.Context) { Destination(c, bindplane) })
	router.DELETE("/destinations/:name", func(c *gin.Context) { DeleteDestination(c, bindplane) })

	router.GET("/destination-types", func(c *gin.Context) { DestinationTypes(c, bindplane) })
	router.GET("/destination-types/:name", func(c *gin.Context) { DestinationType(c, bindplane) })
	router.DELETE("/destination-types/:name", func(c *gin.Context) { DeleteDestinationType(c, bindplane) })

	router.POST("/apply", func(c *gin.Context) { ApplyResources(c, bindplane) })
	router.POST("/delete", func(c *gin.Context) { DeleteResources(c, bindplane) })

	router.GET("/version", func(c *gin.Context) { BindplaneVersion(c) })

	router.GET("/rollouts", func(c *gin.Context) { Rollouts(c, bindplane) })
	router.POST("/rollouts", func(c *gin.Context) { RolloutsUpdate(c, bindplane) })
	router.GET("/rollouts/:name", func(c *gin.Context) { Rollout(c, bindplane) })
	router.GET("/rollouts/:name/status", func(c *gin.Context) { RolloutStatus(c, bindplane) })
	router.POST("/rollouts/:name/start", func(c *gin.Context) { RolloutStart(c, bindplane) })
	router.PUT("/rollouts/:name/pause", func(c *gin.Context) { RolloutPause(c, bindplane) })
	router.PUT("/rollouts/:name/resume", func(c *gin.Context) { RolloutResume(c, bindplane) })
	router.POST("/rollouts/:name/update", func(c *gin.Context) { RolloutUpdate(c, bindplane) })

	router.GET("/:kind/:name/history", func(c *gin.Context) { History(c, bindplane) })
}

// Agents returns a list of agents
// @Summary List Agents
// @Produce json
// @Router /agents [get]
// @Success 200 {object} model.AgentsResponse
// @Failure 500 {object} ErrorResponse
func Agents(c *gin.Context, bindplane exposedserver.BindPlane) {
	ctx, span := tracer.Start(c.Request.Context(), "rest/agents")
	defer span.End()

	options := []store.QueryOption{}

	selectorString := c.DefaultQuery("selector", "")
	selector, err := model.SelectorFromString(selectorString)
	if err != nil {
		HandleErrorResponse(c, http.StatusBadRequest, err)
		return
	}
	options = append(options, store.WithSelector(selector))

	query := c.DefaultQuery("query", "")
	if query != "" {
		q := search.ParseQuery(query)
		q.ReplaceVersionLatest(ctx, bindplane.Versions())
		options = append(options, store.WithQuery(q))
	}

	offset := c.DefaultQuery("offset", "0")
	offsetValue, err := strconv.Atoi(offset)
	if err != nil {
		HandleErrorResponse(c, http.StatusBadRequest, fmt.Errorf("offset must be a number: %v", err))
		return
	}
	options = append(options, store.WithOffset(offsetValue))

	limit := c.DefaultQuery("limit", "0")
	limitValue, err := strconv.Atoi(limit)
	if err != nil {
		HandleErrorResponse(c, http.StatusBadRequest, fmt.Errorf("limit must be a number: %v", err))
		return
	}
	options = append(options, store.WithLimit(limitValue))

	sort := c.DefaultQuery("sort", "")
	if sort != "" {
		options = append(options, store.WithSort(sort))
	}

	agents, err := bindplane.Store().Agents(ctx, options...)
	if err != nil {
		HandleErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, model.AgentsResponse{
		Agents: agents,
	})
}

// DeleteAgents deletes agents by id
// @Summary Delete agents by ids
// @Produce json
// @Router /agents [delete]
// @Param 	id	body	[]string	true "list of agent ids to delete"
// @Success 200 {object} model.DeleteAgentsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func DeleteAgents(c *gin.Context, bindplane exposedserver.BindPlane) {
	ctx, span := tracer.Start(c.Request.Context(), "rest/agents")
	defer span.End()

	p := &model.DeleteAgentsPayload{}

	if err := c.BindJSON(p); err != nil {
		HandleErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	deleted, err := bindplane.Store().DeleteAgents(ctx, p.IDs)
	if err != nil {
		HandleErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, &model.DeleteAgentsResponse{
		Agents: deleted,
	})
}

// GetAgent returns an agent by id
// @Summary Get agent by id
// @Produce json
// @Router /agents/{id} [get]
// @Param 	id	path	string	true "the id of the agent"
// @Success 200 {object} model.AgentResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func GetAgent(c *gin.Context, bindplane exposedserver.BindPlane) {
	id := c.Param("id")

	agent, err := bindplane.Store().Agent(c, id)

	switch {
	case err != nil:
		HandleErrorResponse(c, http.StatusInternalServerError, err)
	case agent == nil:
		HandleErrorResponse(c, http.StatusNotFound, ErrResourceNotFound)
	default:
		c.JSON(http.StatusOK, model.AgentResponse{
			Agent: agent,
		})
	}
}

// GetAgentLabels returns an agent's labels by id
// @Summary Get agent labels by agent id
// @Produce json
// @Router /agents/{id}/labels [get]
// @Param 	id	path	string	true "the id of the agent"
// @Success 200 {object} model.AgentLabelsResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func GetAgentLabels(c *gin.Context, bindplane exposedserver.BindPlane) {
	id := c.Param("id")

	agent, err := bindplane.Store().Agent(c, id)

	switch {
	case err != nil:
		HandleErrorResponse(c, http.StatusInternalServerError, err)
	case agent == nil:
		HandleErrorResponse(c, http.StatusNotFound, ErrResourceNotFound)
	default:
		c.JSON(http.StatusOK, model.AgentLabelsResponse{
			Labels: &agent.Labels,
		})
	}
}

// GetAgentConfiguration returns an agent's configuration by agent id
// @Summary Get configuration for a given agent
// @Produce json
// @Router /agents/{id}/configuration [get]
// @Param 	id	path	string	true "the id of the agent"
// @Success 200 {object} model.ConfigurationResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func GetAgentConfiguration(c *gin.Context, bindplane exposedserver.BindPlane) {
	id := c.Param("id")

	agent, err := bindplane.Store().Agent(c, id)
	switch {
	case err != nil:
		HandleErrorResponse(c, http.StatusInternalServerError, err)
		return
	case agent == nil:
		HandleErrorResponse(c, http.StatusNotFound, ErrResourceNotFound)
		return
	}

	config, err := bindplane.Store().AgentConfiguration(c, agent)
	if err != nil {
		HandleErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, &model.ConfigurationResponse{Configuration: config})
}

// LabelAgents applies labels to agents by id
// @Summary Bulk apply labels to agents
// @Produce json
// @Router /agents/labels [patch]
// @Param ids 	body	[]string	true "agent IDs"
// @Param labels 	body	map[string]string	true "labels to apply"
// @Param labels body boolean false "overwrite labels"
// @Success 200 {object} model.BulkAgentLabelsResponse
func LabelAgents(c *gin.Context, bindplane exposedserver.BindPlane) {
	ctx, span := tracer.Start(c.Request.Context(), "rest/labelAgents")
	defer span.End()

	p := &model.BulkAgentLabelsPayload{}

	if err := c.BindJSON(p); err != nil {
		span.SetStatus(codes.Error, err.Error())
		HandleErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	if p.Labels == nil {
		HandleErrorResponse(c, http.StatusBadRequest, fmt.Errorf("body is missing the required labels field"))
		return
	}

	if p.IDs == nil {
		HandleErrorResponse(c, http.StatusBadRequest, fmt.Errorf(("body is missing the required ids field")))
		return
	}

	newLabels, err := model.LabelsFromMap(p.Labels)
	if err != nil {
		HandleErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	// Accumulate API errors outside of upsert, and then upsert agents with valid label operations
	// Check to see if 1) agent exists and 2) there are no label conflicts if overwrite=false.
	upsertIDs := make([]string, 0, len(p.IDs))
	apiErrors := make([]string, 0)
	for _, id := range p.IDs {
		curAgent, err := bindplane.Store().Agent(c, id)

		switch {
		case err != nil:
			HandleErrorResponse(c, http.StatusInternalServerError, err)
			apiErrors = append(apiErrors, fmt.Sprintf("failed to apply labels for agent with id %s, %s", id, err.Error()))
			continue
		case curAgent == nil:
			apiErrors = append(apiErrors, fmt.Sprintf("failed to apply labels for agent with id %s, agent not found", id))
			continue
		case !p.Overwrite && curAgent.Labels.Conflicts(newLabels):
			apiErrors = append(apiErrors, fmt.Sprintf("failed to apply labels for agent with id %s, labels conflict, include overwrite: true in body to overwrite", id))
			continue
		}
		// Agent is cleared to patch - add it to upsertIDs
		upsertIDs = append(upsertIDs, id)
	}

	updater := func(current *model.Agent) {
		current.Labels = model.LabelsFromMerge(current.Labels, newLabels)
	}

	bindplane.Logger().Info("bulkApplyAgentLabels", zap.String("payloadLabels", newLabels.String()), zap.Any("ids", p.IDs), zap.Error(err))

	_, err = bindplane.Store().UpsertAgents(ctx, p.IDs, updater)

	if err != nil {
		HandleErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, &model.BulkAgentLabelsResponse{
		Errors: apiErrors,
	})
}

// PatchAgentLabels patches an agent's labels by agent id
// @Summary Patch agent labels by agent id
// @Produce json
// @Router /agents/{id}/labels [patch]
// @Param 	id	path	string	true "the id of the agent"
// @Param overwrite query string false "if true, overwrite any existing labels with the same names"
// @Param labels 	body	model.AgentLabelsPayload	true "Labels to be merged with existing labels, empty values will delete existing labels"
// @Success 200 {object} model.AgentLabelsResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func PatchAgentLabels(c *gin.Context, bindplane exposedserver.BindPlane) {
	ctx, span := tracer.Start(c.Request.Context(), "rest/patchAgentLabels")
	defer span.End()

	id := c.Param("id")
	span.SetAttributes(attribute.String("bindplane.agent.id", id))

	overwrite := c.DefaultQuery("overwrite", "false") == "true"
	p := &model.AgentLabelsPayload{}
	if err := c.BindJSON(p); err != nil {
		HandleErrorResponse(c, http.StatusBadRequest, err)
		return
	}
	if p.Labels == nil {
		HandleErrorResponse(c, http.StatusBadRequest, fmt.Errorf("body is missing the required labels field"))
		return
	}

	newLabels, err := model.LabelsFromMap(p.Labels)
	if err != nil {
		HandleErrorResponse(c, http.StatusBadRequest, err)
	}

	curAgent, err := bindplane.Store().Agent(c, id)
	switch {
	case err != nil:
		span.SetStatus(codes.Error, err.Error())
		HandleErrorResponse(c, http.StatusInternalServerError, err)
		return
	case curAgent == nil:
		span.SetStatus(codes.Error, ErrResourceNotFound.Error())
		HandleErrorResponse(c, http.StatusNotFound, ErrResourceNotFound)
		return
	case !overwrite && curAgent.Labels.Conflicts(newLabels):
		err := fmt.Errorf("new labels conflict with existing labels, add ?overwrite=true to replace labels")
		span.SetStatus(codes.Error, err.Error())
		c.Error(err)
		c.JSON(http.StatusConflict, model.AgentLabelsResponse{
			Errors: []string{err.Error()},
			Labels: &curAgent.Labels,
		})
		return
	}

	newAgent, err := bindplane.Store().UpsertAgent(ctx, id, func(agent *model.Agent) {
		agent.Labels = model.LabelsFromMerge(agent.Labels, newLabels)
	})

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		HandleErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	bindplane.Logger().Info("patchAgentLabels",
		zap.String("payloadLabels", newLabels.String()),
		zap.String("newLabels", newAgent.Labels.String()),
	)
	c.JSON(http.StatusOK, model.AgentLabelsResponse{
		Labels: &newAgent.Labels,
	})
}

// RestartAgent restarts an agent by id
// @Summary TODO restart agent
// @Produce json
// @Router /agents/{id}/restart [put]
// @Param 	id	path	string	true "the id of the agent"
func RestartAgent(c *gin.Context, bindplane exposedserver.BindPlane) {
	id := c.Param("id")

	// TODO(andy): Do a restart
	bindplane.Logger().Info("TODO Restart agent", zap.String("id", id))

	c.Status(http.StatusAccepted)
}

// UpgradeAgents upgrades agents to latest version by id
// @Summary Update multiple agents
// @Router /agents/version [patch]
// @Param body body model.PatchAgentVersionsRequest true "request body containing ids and version"
func UpgradeAgents(c *gin.Context, bindplane exposedserver.BindPlane) {
	ctx, span := tracer.Start(c.Request.Context(), "rest/upgradeAgents")
	defer span.End()

	req := &struct {
		IDs     []string
		Version string
	}{}

	if err := c.BindJSON(req); err != nil {
		HandleErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	var version string
	if req.Version == "" {
		version = bindplane.Versions().LatestVersionString(ctx)
	} else {
		version = req.Version
	}

	for _, id := range req.IDs {
		// just ignore agents that don't exist or don't support upgrade
		agent, err := bindplane.Store().Agent(c, id)
		if err != nil || agent == nil || !agent.SupportsUpgrade() {
			continue
		}

		_, err = bindplane.Store().UpsertAgent(ctx, id, func(current *model.Agent) {
			current.UpgradeTo(version)
		})
		if err != nil {
			HandleErrorResponse(c, http.StatusInternalServerError, err)
			return
		}
	}

	c.Status(http.StatusNoContent)
}

// UpgradeAgent upgrades an agent to latest version by id
// @Summary Upgrade agent
// @Produce json
// @Router /agents/{id}/version [post]
// @Param 	name	path	string	true "the id of the agent"
// @Param body body model.PostAgentVersionRequest true "request body containing version"
// @Failure 409 {object} ErrorResponse "If the agent does not support upgrade"
// @Failure 500 {object} ErrorResponse
func UpgradeAgent(c *gin.Context, bindplane exposedserver.BindPlane) {
	ctx, span := tracer.Start(c.Request.Context(), "rest/upgradeAgent")
	defer span.End()

	id := c.Param("id")
	var req model.PostAgentVersionRequest

	if err := c.BindJSON(&req); err != nil {
		HandleErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	agent, err := bindplane.Store().Agent(c, id)
	switch {
	case err != nil:
		HandleErrorResponse(c, http.StatusInternalServerError, err)
		return

	case agent == nil:
		HandleErrorResponse(c, http.StatusNotFound, ErrResourceNotFound)
		return

	case !agent.SupportsUpgrade():
		HandleErrorResponse(c, http.StatusConflict, fmt.Errorf("agent %s with version %s does not support upgrade", agent.ID, agent.Version))
		return
	}

	// start an upgrade process
	_, err = bindplane.Store().UpsertAgent(ctx, id, func(current *model.Agent) {
		current.UpgradeTo(req.Version)
	})
	if err != nil {
		HandleErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ----------------------------------------------------------------------

// AgentVersions returns a list of agent versions
// @Summary List agent versions
// @Produce json
// @Router /agent-versions [get]
// @Success 200 {object} model.AgentVersionsResponse
// @Failure 500 {object} ErrorResponse
func AgentVersions(c *gin.Context, bindplane exposedserver.BindPlane) {
	agentVersions, err := bindplane.Store().AgentVersions(c)
	if OkResponse(c, err) {
		c.JSON(http.StatusOK, model.AgentVersionsResponse{
			AgentVersions: agentVersions,
		})
	}
}

// AgentVersion returns an agent version by name
// @Summary Get agent version by name
// @Produce json
// @Router /agent-versions/{name} [get]
// @Param 	name	path	string	true "the name of the agent version"
// @Success 200 {object} model.AgentVersionResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func AgentVersion(c *gin.Context, bindplane exposedserver.BindPlane) {
	name := c.Param("name")
	agentVersion, err := bindplane.Store().AgentVersion(c, name)
	if OkResource(c, agentVersion == nil, err) {
		c.JSON(http.StatusOK, model.AgentVersionResponse{
			AgentVersion: agentVersion,
		})
	}
}

// DeleteAgentVersion deletes an agent version by name
// @Summary Delete agent version by name
// @Produce json
// @Router /agent-versions/{name} [delete]
// @Param 	name	path	string	true "the name of the agent version to delete"
// @Success 204	"Successful Delete, no content"
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func DeleteAgentVersion(c *gin.Context, bindplane exposedserver.BindPlane) {
	name := c.Param("name")
	agentVersion, err := bindplane.Store().DeleteAgentVersion(c, name)
	if OkResource(c, agentVersion == nil, err) {
		c.Status(http.StatusNoContent)
	}
}

// ----------------------------------------------------------------------

// Configurations returns a list of configurations
// @Summary List Configurations
// @Produce json
// @Router /configurations [get]
// @Success 200 {object} model.ConfigurationsResponse
// @Failure 500 {object} ErrorResponse
func Configurations(c *gin.Context, bindplane exposedserver.BindPlane) {
	configs, err := bindplane.Store().Configurations(c)
	if err != nil {
		HandleErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, model.ConfigurationsResponse{
		Configurations: configs,
	})
}

// Configuration returns a configuration by name
// @Summary Get Configuration by name
// @Produce json
// @Router /configurations/{name} [get]
// @Param 	name	path	string	true "the name of the Configuration"
// @Success 200 {object} model.ConfigurationResponse
// @Failure 500 {object} ErrorResponse
func Configuration(c *gin.Context, bindplane exposedserver.BindPlane) {
	ctx, span := tracer.Start(c.Request.Context(), "rest/configuration")
	defer span.End()

	name := c.Param("name")

	config, err := bindplane.Store().Configuration(ctx, name)
	if !OkResource(c, config == nil, err) {
		return
	}

	raw, err := config.Render(ctx, nil, bindplane.BindPlaneURL(), bindplane.BindPlaneInsecureSkipVerify(), bindplane.Store(), model.GetOssOtelHeaders())
	if !OkResponse(c, err) {
		return
	}

	c.JSON(http.StatusOK, model.ConfigurationResponse{
		Configuration: config,
		Raw:           raw,
	})
}

// DeleteConfiguration deletes a configuration by name
// @Summary Delete configuration by name
// @Produce json
// @Router /configurations/{name} [delete]
// @Param 	name	path	string	true "the name of the configuration to delete"
// @Success 204	"Successful Delete, no content"
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func DeleteConfiguration(c *gin.Context, bindplane exposedserver.BindPlane) {
	name := c.Param("name")
	configuration, err := bindplane.Store().DeleteConfiguration(c, name)
	if OkResource(c, configuration == nil, err) {
		c.Status(http.StatusNoContent)
	}
}

// CopyConfig duplicates an existing configuration
// @Summary Duplicate an existing configuration
// @Produce json
// @Router /configurations/{name}/copy [post]
// @Param 	name	path	string	true "the name of the configuration to duplicate"
// @Param name	body	string	true "the desired name of the duplicate configuration"
// @Success 201	"Successful Copy, created"
// @Failure 404 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func CopyConfig(c *gin.Context, bindplane exposedserver.BindPlane) {
	name := c.Param("name")

	// The config to make a duplicate of
	config, err := bindplane.Store().Configuration(c, name)
	if err != nil {
		HandleErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	if config == nil {
		HandleErrorResponse(c, http.StatusNotFound, fmt.Errorf("no configuration with name %s found", name))
		return
	}

	var req model.PostCopyConfigRequest
	if err := c.BindJSON(&req); err != nil {
		HandleErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	duplicateName := req.Name
	duplicateConfig, err := bindplane.Store().Configuration(c, duplicateName)
	if err != nil {
		HandleErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	if duplicateConfig != nil {
		HandleErrorResponse(c, http.StatusConflict, errors.New("a configuration with that name already exists"))
		return
	}

	duplicateConfig = config.Duplicate(duplicateName)

	updates, err := bindplane.Store().ApplyResources(c, []model.Resource{duplicateConfig})
	if err != nil {
		HandleErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	// Verify we got status created
	update := &model.ResourceStatus{}
	for _, u := range updates {
		if u.Resource.Name() == duplicateName {
			*update = u
			break
		}
	}

	if update.Status == model.StatusCreated {
		c.JSON(http.StatusCreated, model.PostCopyConfigResponse{
			Name: update.Resource.Name(),
		})
		return
	}

	err = fmt.Errorf("failed to apply copied configuration, got status %s", update.Status)

	if update.Reason != "" {
		err = errors.Join(err, errors.New(update.Reason))
	}
	HandleErrorResponse(c, http.StatusBadRequest, err)
}

// ----------------------------------------------------------------------

// Sources returns a list of sources
// @Summary List Sources
// @Produce json
// @Router /sources [get]
// @Success 200 {object} model.SourcesResponse
// @Failure 500 {object} ErrorResponse
func Sources(c *gin.Context, bindplane exposedserver.BindPlane) {
	sources, err := bindplane.Store().Sources(c)
	if OkResponse(c, err) {
		c.JSON(http.StatusOK, model.SourcesResponse{
			Sources: sources,
		})
	}
}

// Source returns a source by name
// @Summary Get Source by name
// @Produce json
// @Router /sources/{name} [get]
// @Param 	name	path	string	true "the name of the Source"
// @Success 200 {object} model.SourceResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func Source(c *gin.Context, bindplane exposedserver.BindPlane) {
	name := c.Param("name")
	source, err := bindplane.Store().Source(c, name)
	if OkResource(c, source == nil, err) {
		c.JSON(http.StatusOK, model.SourceResponse{
			Source: source,
		})
	}
}

// DeleteSource deletes a source by name
// @Summary Delete source by name
// @Produce json
// @Router /sources/{name} [delete]
// @Param 	name	path	string	true "the name of the source to delete"
// @Success 204	"Successful Delete, no content"
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func DeleteSource(c *gin.Context, bindplane exposedserver.BindPlane) {
	name := c.Param("name")
	source, err := bindplane.Store().DeleteSource(c, name)

	if OkResource(c, source == nil, err) {
		c.Status(http.StatusNoContent)
	}
}

// ----------------------------------------------------------------------

// SourceTypes returns a list of source types
// @Summary List source types
// @Produce json
// @Router /source-types [get]
// @Success 200 {object} model.SourceTypesResponse
// @Failure 500 {object} ErrorResponse
func SourceTypes(c *gin.Context, bindplane exposedserver.BindPlane) {
	sourceTypes, err := bindplane.Store().SourceTypes(c)
	if OkResponse(c, err) {
		c.JSON(http.StatusOK, model.SourceTypesResponse{
			SourceTypes: sourceTypes,
		})
	}
}

// SourceType returns a source type by name
// @Summary Get source type by name
// @Produce json
// @Router /source-types/{name} [get]
// @Param 	name	path	string	true "the name of the source type"
// @Success 200 {object} model.SourceTypeResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func SourceType(c *gin.Context, bindplane exposedserver.BindPlane) {
	name := c.Param("name")
	sourceType, err := bindplane.Store().SourceType(c, name)
	if OkResource(c, sourceType == nil, err) {
		c.JSON(http.StatusOK, model.SourceTypeResponse{
			SourceType: sourceType,
		})
	}
}

// DeleteSourceType deletes a source type by name
// @Summary Delete source type by name
// @Produce json
// @Router /source-types/{name} [delete]
// @Param 	name	path	string	true "the name of the source type to delete"
// @Success 204	"Successful Delete, no content"
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func DeleteSourceType(c *gin.Context, bindplane exposedserver.BindPlane) {
	name := c.Param("name")
	sourceType, err := bindplane.Store().DeleteSourceType(c, name)
	if OkResource(c, sourceType == nil, err) {
		c.Status(http.StatusNoContent)
	}
}

// ----------------------------------------------------------------------

// Processors returns a list of processors
// @Summary List Processors
// @Produce json
// @Router /processors [get]
// @Success 200 {object} model.ProcessorsResponse
// @Failure 500 {object} ErrorResponse
func Processors(c *gin.Context, bindplane exposedserver.BindPlane) {
	processors, err := bindplane.Store().Processors(c)
	if OkResponse(c, err) {
		c.JSON(http.StatusOK, model.ProcessorsResponse{
			Processors: processors,
		})
	}
}

// Processor returns a processor by name
// @Summary Get Processor by name
// @Produce json
// @Router /processors/{name} [get]
// @Param 	name	path	string	true "the name of the Processor"
// @Success 200 {object} model.ProcessorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func Processor(c *gin.Context, bindplane exposedserver.BindPlane) {
	name := c.Param("name")
	processor, err := bindplane.Store().Processor(c, name)
	if OkResource(c, processor == nil, err) {
		c.JSON(http.StatusOK, model.ProcessorResponse{
			Processor: processor,
		})
	}
}

// DeleteProcessor deletes a processor by name
// @Summary Delete processor by name
// @Produce json
// @Router /processors/{name} [delete]
// @Param 	name	path	string	true "the name of the processor to delete"
// @Success 204	"Successful Delete, no content"
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func DeleteProcessor(c *gin.Context, bindplane exposedserver.BindPlane) {
	name := c.Param("name")
	processor, err := bindplane.Store().DeleteProcessor(c, name)
	if OkResource(c, processor == nil, err) {
		c.Status(http.StatusNoContent)
	}
}

// ----------------------------------------------------------------------

// ProcessorTypes returns a list of processor types
// @Summary List processor types
// @Produce json
// @Router /processor-types [get]
// @Success 200 {object} model.ProcessorTypesResponse
// @Failure 500 {object} ErrorResponse
func ProcessorTypes(c *gin.Context, bindplane exposedserver.BindPlane) {
	processorTypes, err := bindplane.Store().ProcessorTypes(c)
	if OkResponse(c, err) {
		c.JSON(http.StatusOK, model.ProcessorTypesResponse{
			ProcessorTypes: processorTypes,
		})
	}
}

// ProcessorType returns a processor type by name
// @Summary Get processor type by name
// @Produce json
// @Router /processor-types/{name} [get]
// @Param 	name	path	string	true "the name of the processor type"
// @Success 200 {object} model.ProcessorTypeResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func ProcessorType(c *gin.Context, bindplane exposedserver.BindPlane) {
	name := c.Param("name")
	processorType, err := bindplane.Store().ProcessorType(c, name)
	if OkResource(c, processorType == nil, err) {
		c.JSON(http.StatusOK, model.ProcessorTypeResponse{
			ProcessorType: processorType,
		})
	}
}

// DeleteProcessorType deletes a processor type by name
// @Summary Delete processor type by name
// @Produce json
// @Router /processor-types/{name} [delete]
// @Param 	name	path	string	true "the name of the processor type to delete"
// @Success 204	"Successful Delete, no content"
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func DeleteProcessorType(c *gin.Context, bindplane exposedserver.BindPlane) {
	name := c.Param("name")
	processorType, err := bindplane.Store().DeleteProcessorType(c, name)
	if OkResource(c, processorType == nil, err) {
		c.Status(http.StatusNoContent)
	}
}

// ----------------------------------------------------------------------

// Destinations returns a list of destinations
// @Summary List Destinations
// @Produce json
// @Router /destinations [get]
// @Success 200 {object} model.DestinationsResponse
// @Failure 500 {object} ErrorResponse
func Destinations(c *gin.Context, bindplane exposedserver.BindPlane) {
	destinations, err := bindplane.Store().Destinations(c)
	if OkResponse(c, err) {
		c.JSON(http.StatusOK, model.DestinationsResponse{
			Destinations: destinations,
		})
	}
}

// Destination returns a destination by name
// @Summary Get Destination by name
// @Produce json
// @Router /destinations/{name} [get]
// @Param 	name	path	string	true "the name of the Destination"
// @Success 200 {object} model.DestinationResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func Destination(c *gin.Context, bindplane exposedserver.BindPlane) {
	name := c.Param("name")
	destination, err := bindplane.Store().Destination(c, name)
	if OkResource(c, destination == nil, err) {
		c.JSON(http.StatusOK, model.DestinationResponse{
			Destination: destination,
		})
	}
}

// DeleteDestination deletes a destination by name
// @Summary Delete destination by name
// @Produce json
// @Router /destinations/{name} [delete]
// @Param 	name	path	string	true "the name of the destination to delete"
// @Success 204	"Successful Delete, no content"
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func DeleteDestination(c *gin.Context, bindplane exposedserver.BindPlane) {
	name := c.Param("name")
	destination, err := bindplane.Store().DeleteDestination(c, name)
	if OkResource(c, destination == nil, err) {
		c.Status(http.StatusNoContent)
	}
}

// ----------------------------------------------------------------------

// DestinationTypes returns a list of destination types
// @Summary List destination types
// @Produce json
// @Router /destination-types [get]
// @Success 200 {object} model.DestinationTypesResponse
// @Failure 500 {object} ErrorResponse
func DestinationTypes(c *gin.Context, bindplane exposedserver.BindPlane) {
	destinationTypes, err := bindplane.Store().DestinationTypes(c)
	if OkResponse(c, err) {
		c.JSON(http.StatusOK, model.DestinationTypesResponse{
			DestinationTypes: destinationTypes,
		})
	}
}

// DestinationType returns a destination type by name
// @Summary Get destination type by name
// @Produce json
// @Router /destination-types/{name} [get]
// @Param 	name	path	string	true "the name of the destination type"
// @Success 200 {object} model.DestinationTypeResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func DestinationType(c *gin.Context, bindplane exposedserver.BindPlane) {
	name := c.Param("name")
	destinationType, err := bindplane.Store().DestinationType(c, name)
	if OkResource(c, destinationType == nil, err) {
		c.JSON(http.StatusOK, model.DestinationTypeResponse{
			DestinationType: destinationType,
		})
	}
}

// DeleteDestinationType deletes a destination type by name
// @Summary Delete destination type by name
// @Produce json
// @Router /destination-types/{name} [delete]
// @Param 	name	path	string	true "the name of the destination type to delete"
// @Success 204	"Successful Delete, no content"
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func DeleteDestinationType(c *gin.Context, bindplane exposedserver.BindPlane) {
	name := c.Param("name")
	destinationType, err := bindplane.Store().DeleteDestinationType(c, name)
	if OkResource(c, destinationType == nil, err) {
		c.Status(http.StatusNoContent)
	}
}

// ----------------------------------------------------------------------

// ApplyResources creates, edits, and configures multiple resources
// @Summary Create, edit, and configure multiple resources.
// @Description The /apply route will try to parse resources
// @Description and upsert them into the store.  Additionally
// @Description it will send reconfigure tasks to affected agents.
// @Produce json
// @Router /apply [post]
// @Param resources 	body	[]model.AnyResource	true "Resources"
// @Success 200 {object} model.ApplyResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func ApplyResources(c *gin.Context, bindplane exposedserver.BindPlane) {
	p := &model.ApplyPayload{}
	if err := c.BindJSON(p); err != nil {
		HandleErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	// parse the resources
	resources := []model.Resource{}
	for _, res := range p.Resources {
		parsed, err := model.ParseResourceStrict(res)
		// TODO (dsvanlani): Go through all resources and gather errors.
		if err != nil {
			HandleErrorResponse(c, http.StatusBadRequest, err)
			return
		}

		resources = append(resources, parsed)
	}

	// When testing rendering the config, we want to first look at the new resources to apply.
	// We do this, because the config may depend on resources that are currently being applied (e.g. destinations),
	// which are not yet stored.
	memoryFirstStore := newMemoryFirstResourceStore(resources, bindplane.Store())
	// Extra validation for configs; We want to ensure that the configuration CAN be rendered before saving it.
	for _, res := range resources {
		if conf, ok := res.(*model.Configuration); ok {
			_, err := conf.Render(c, nil, bindplane.BindPlaneURL(), bindplane.BindPlaneInsecureSkipVerify(), memoryFirstStore, model.GetOssOtelHeaders())
			if err != nil {
				HandleErrorResponse(c, http.StatusBadRequest, fmt.Errorf("failed to render config (resourceID: %s): %w", res.ID(), err))
				return
			}
		}
	}

	bindplane.Logger().Info("/apply", zap.Int("count", len(resources)))

	resourceStatuses, err := bindplane.Store().ApplyResources(c, resources)
	if err != nil {
		HandleErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusAccepted, &model.ApplyResponse{
		Updates: resourceStatuses,
	})
}

// DeleteResources deletes multiple resources
// @Summary Delete multiple resources
// @Description /delete endpoint will try to parse resources
// @Description and delete them from the store.  Additionally
// @Description it will send reconfigure tasks to affected agents.
// @Produce json
// @Router /delete [post]
// @Param resources 	body	[]model.AnyResource	true "Resources"
// @Success 200 {object} model.DeleteResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func DeleteResources(c *gin.Context, bindplane exposedserver.BindPlane) {
	p := &model.DeletePayload{}
	if err := c.BindJSON(p); err != nil {
		HandleErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	// parse the resources
	resources := []model.Resource{}
	for _, res := range p.Resources {
		// Non-strict parse; We only care about the ID here, so we'll allow resources with extra keys.
		// This is similar to how k8s handles deletion of resources with extra keys.
		parsed, err := model.ParseResource(res)
		if err != nil {
			HandleErrorResponse(c, http.StatusBadRequest, err)
			return
		}
		resources = append(resources, parsed)
	}

	bindplane.Logger().Info("/delete", zap.Int("count", len(resources)))

	resourceStatuses, err := bindplane.Store().DeleteResources(c, resources)
	if err != nil {
		HandleErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusAccepted, &model.DeleteResponse{
		Updates: resourceStatuses,
	})
}

// BindplaneVersion returns the current bindplane version of the server.
// @Summary Server version
// @Description Returns the current bindplane version of the server.
// @Produce json
// @Router /version [get]
// @Success 200 {string} version.Version
func BindplaneVersion(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, version.NewVersion())
}

// @Summary Get Install Command
// @Description Get the proper install command for the provided parameters.
// @Produce json
// @Router /agent-versions/{version}/install-command [get]
// @Param version 	path	string	true "2.1.1"
// @Param secret-key query string false "uuid"
// @Param remote-url query string false "http%3A%2F%2Flocalhost%3A3001"
// @Param platform query string false "windows-amd64"
// @Param labels query string false "env=stage,app=bindplane"
// @Success 200 {object} model.InstallCommandResponse
func getInstallCommand(c *gin.Context, bindplane exposedserver.BindPlane) {

	// note: don't use DefaultQuery because caller may specify secret-key=(empty string) but we want to use the default
	// value in that case
	secretKey := c.Query("secret-key")
	if secretKey == "" {
		secretKey = bindplane.SecretKey()
	}

	remoteURL := c.Query("remote-url")
	if remoteURL == "" {
		remoteURL = fmt.Sprintf("%s/v1/opamp", bindplane.WebsocketURL())
	}

	serverURL := bindplane.BindPlaneURL()

	// if version is empty or "latest", find the latest version
	version := c.Param("name")
	if version == "" || version == "latest" {
		v, err := bindplane.Versions().LatestVersion(c)
		if err != nil {
			HandleErrorResponse(c, http.StatusInternalServerError,
				fmt.Errorf("unable to get the latest version of the agent: %w", err),
			)
			c.Status(http.StatusInternalServerError)
			return
		}
		version = v.AgentVersion()
	}

	platform, ok := normalizePlatform(c.Query("platform"))
	if !ok {
		HandleErrorResponse(c, http.StatusBadRequest,
			fmt.Errorf("unknown platform: %s", c.Query("platform")),
		)
		return
	}

	params := installCommandParameters{
		platform:  platform,
		version:   version,
		labels:    c.Query("labels"),
		secretKey: secretKey,
		remoteURL: remoteURL,
		serverURL: serverURL,
	}
	cmd, err := params.installCommand()
	if err != nil {
		HandleErrorResponse(c, http.StatusInternalServerError,
			fmt.Errorf("failed to generate the install command: %w", err),
		)
		c.Status(http.StatusInternalServerError)
		return
	}
	response := model.InstallCommandResponse{
		Command: cmd,
	}
	c.JSON(http.StatusOK, response)
}

// SyncAgentVersion creates an agent-version from the contents of a github release.
// @Summary Sync Agent Version
// @Description Create an agent-version from the contents of a github release.
// @Produce json
// @Router /agent-versions/{version}/sync [post]
// @Param version 	path	string	true "2.1.1"
// @Success 200 {object} model.ApplyResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func SyncAgentVersion(c *gin.Context, bindplane exposedserver.BindPlane) {
	var resources []model.Resource

	// create an agent-version resource from the contents of the github release

	// if version is empty or "latest", find the latest version
	version := c.Param("name")
	if version == "" {
		agentVersions, err := bindplane.Versions().SyncVersions()
		if err != nil {
			HandleErrorResponse(c, http.StatusInternalServerError, err)
			return
		}
		for _, agentVersion := range agentVersions {
			resources = append(resources, agentVersion)
		}
	} else {
		agentVersion, err := bindplane.Versions().SyncVersion(version)
		if err != nil {
			HandleErrorResponse(c, http.StatusInternalServerError, err)
			return
		}
		resources = append(resources, agentVersion)
	}

	resourceStatuses, err := bindplane.Store().ApplyResources(c, resources)
	if err != nil {
		HandleErrorResponse(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusAccepted, &model.ApplyResponse{
		Updates: resourceStatuses,
	})
}

// Rollouts returns all configurations with active rollouts.
// @Summary Get all rollouts
// @Produce json
// @Router /rollouts [get]
// @Success 200 {object} model.ConfigurationsResponse
// @Failure 500 {object} ErrorResponse
func Rollouts(c *gin.Context, bindplane exposedserver.BindPlane) {
	ctx, span := tracer.Start(c.Request.Context(), "rest/rollouts")
	defer span.End()

	// TODO(andy): only return configurations with active rollouts
	configurations, err := bindplane.Store().Configurations(ctx)

	if !OkResource(c, configurations == nil, err) {
		return
	}

	c.JSON(http.StatusOK, model.ConfigurationsResponse{
		Configurations: configurations,
	})
}

// Rollout returns the configuration with the provided name.
// @Summary Get rollout configuration by name
// @Produce json
// @Router /rollouts/{name} [get]
// @Param 	name	path	string	true "the name of the configuration"
// @Success 200 {object} model.ConfigurationResponse
// @Failure 500 {object} ErrorResponse
func Rollout(c *gin.Context, bindplane exposedserver.BindPlane) {
	ctx, span := tracer.Start(c.Request.Context(), "rest/rollout")
	defer span.End()

	name := c.Param("name")

	config, err := bindplane.Store().Configuration(ctx, name)
	if !OkResource(c, config == nil, err) {
		return
	}

	c.JSON(http.StatusOK, model.ConfigurationResponse{
		Configuration: config,
	})
}

// RolloutStatus returns the status of the configuration rollout with the provided name.
// @Summary Status of configuration rollout by name
// @Produce json
// @Router /rollouts/{name}/status [get]
// @Param 	name	path	string	true "the name of the configuration"
// @Success 202 {object} model.ConfigurationResponse
// @Failure 500 {object} ErrorResponse
func RolloutStatus(c *gin.Context, bindplane exposedserver.BindPlane) {
	bindplane.Logger().Debug("statusOfRollout")
	name := c.Param("name")
	config, err := bindplane.Store().Configuration(c, name)
	if !OkResource(c, config == nil, err) {
		return
	}
	c.JSON(http.StatusAccepted, model.ConfigurationResponse{
		Configuration: config,
	})
}

// RolloutStart starts a rollout by configuration name.
// @Summary Start rollout by configuration name
// @Produce json
// @Router /rollouts/{name}/start [post]
// @Param 	name	path	string	true "the name of the configuration"
// @Param   options body model.RolloutOptions false "the options for the rollout"
// @Success 202 {object} model.ConfigurationResponse
// @Failure 500 {object} ErrorResponse
func RolloutStart(c *gin.Context, bindplane exposedserver.BindPlane) {
	ctx, span := tracer.Start(c.Request.Context(), "rest/rolloutStart")
	defer span.End()

	bindplane.Logger().Debug("startRollout")
	name := c.Param("name")

	payload := &model.StartRolloutPayload{}

	if err := c.BindJSON(payload); err != nil {
		span.SetStatus(codes.Error, err.Error())
		HandleErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	config, err := bindplane.Store().StartRollout(ctx, name, payload.Options)

	if !OkResource(c, config == nil, err) {
		return
	}
	c.JSON(http.StatusAccepted, model.ConfigurationResponse{
		Configuration: config,
	})
}

// RolloutResume resumes a rollout by configuration name.
// @Summary Resume rollout by configuration name
// @Produce json
// @Router /rollouts/{name}/resume [post]
// @Param 	name	path	string	true "the name of the configuration"
// @Success 202 {object} model.ConfigurationResponse
// @Failure 500 {object} ErrorResponse
func RolloutResume(c *gin.Context, bindplane exposedserver.BindPlane) {
	bindplane.Logger().Debug("resumeRollout")
	name := c.Param("name")

	configuration, err := bindplane.Store().ResumeRollout(c, name)

	if !OkResource(c, configuration == nil, err) {
		return
	}
	c.JSON(http.StatusAccepted, model.ConfigurationResponse{
		Configuration: configuration,
	})
}

// RolloutPause pauses a rollout by configuration name.
// @Summary Pause rollout by configuration name
// @Produce json
// @Router /rollouts/{name}/pause [post]
// @Param 	name	path	string	true "the name of the configuration"
// @Success 202 {object} model.ConfigurationResponse
// @Failure 500 {object} ErrorResponse
func RolloutPause(c *gin.Context, bindplane exposedserver.BindPlane) {
	bindplane.Logger().Debug("pauseRollout")

	name := c.Param("name")

	configuration, err := bindplane.Store().PauseRollout(c, name)

	if !OkResource(c, configuration == nil, err) {
		return
	}
	c.JSON(http.StatusAccepted, model.ConfigurationResponse{
		Configuration: configuration,
	})
}

// RolloutUpdate updates a rollout by configuration name.
// @Summary Update rollout by configuration name
// @Produce json
// @Router /rollouts/{name}/update [post]
// @Param 	name	path	string	true "the name of the configuration"
// @Success 202 {object} model.ConfigurationResponse
// @Failure 500 {object} ErrorResponse
func RolloutUpdate(c *gin.Context, bindplane exposedserver.BindPlane) {
	bindplane.Logger().Debug("updateRollout")
	name := c.Param("name")
	configuration, err := bindplane.Store().Configuration(c, name)
	if !OkResource(c, configuration == nil, err) {
		return
	}
	config, err := bindplane.Store().UpdateRollout(c, name)

	if !OkResponse(c, err) {
		return
	}

	c.JSON(http.StatusAccepted, model.ConfigurationResponse{
		Configuration: config,
	})
}

// RolloutsUpdate updates all active rollouts.
// @Summary Update all active rollouts
// @Produce json
// @Router /rollouts [post]
// @Success 202 {object} model.ConfigurationsResponse
// @Failure 500 {object} ErrorResponse
func RolloutsUpdate(c *gin.Context, bindplane exposedserver.BindPlane) {
	bindplane.Logger().Debug("updateRollout")
	configurations, err := bindplane.Store().UpdateRollouts(c)

	if !OkResponse(c, err) {
		return
	}

	c.JSON(http.StatusAccepted, model.ConfigurationsResponse{
		Configurations: configurations,
	})
}

// ----------------------------------------------------------------------

// History returns the history of a resource.
// @Summary Get the history of a resource
// @Produce json
// @Router /{kind}/{name}/history [get]
// @Param 	kind	path	string	true "the kind of the resource"
// @Param 	name	path	string	true "the name of the resource"
// @Success 200 {object} model.HistoryResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func History(c *gin.Context, bindplane exposedserver.BindPlane) {
	ctx, span := tracer.Start(c.Request.Context(), "rest/history")
	defer span.End()

	kind := model.ParseKind(c.Param("kind"))
	if kind == model.KindUnknown {
		HandleErrorResponse(c, http.StatusBadRequest, fmt.Errorf("invalid kind: %s", c.Param("kind")))
		return
	}
	name := c.Param("name")

	archiveStore, ok := bindplane.Store().(store.ArchiveStore)
	if !ok {
		HandleErrorResponse(c, http.StatusInternalServerError, store.ErrDoesNotSupportHistory)
		return
	}

	history, err := archiveStore.ResourceHistory(ctx, kind, name)
	if !OkResponse(c, err) {
		return
	}

	c.JSON(http.StatusOK, model.HistoryResponse{
		Versions: history,
	})
}

// ----------------------------------------------------------------------

// OkResponse returns true if there should be an OK response based on the error provided. It will set an error response on the
// gin.Context if appropriate.
func OkResponse(c *gin.Context, err error) bool {
	switch {
	case err == nil:
		return true
	case errors.Is(err, ErrResourceNotFound):
		HandleErrorResponse(c, http.StatusNotFound, err)
	case isDependencyError(err):
		HandleErrorResponse(c, http.StatusConflict, err)
	default:
		HandleErrorResponse(c, http.StatusInternalServerError, err)
	}
	return false
}

// OkResource returns true if there should be an OK response based on the resource and error provided. It will set an
// error response on the gin.Context if appropriate.
func OkResource(c *gin.Context, resourceIsNil bool, err error) bool {
	if !OkResponse(c, err) {
		return false
	}
	if resourceIsNil {
		HandleErrorResponse(c, http.StatusNotFound, ErrResourceNotFound)
		return false
	}
	return true
}

func isDependencyError(err error) bool {
	_, ok := err.(*store.DependencyError)
	return ok
}
