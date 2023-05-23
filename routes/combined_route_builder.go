// Copyright observIQ, Inc.
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

// Package routes contains route builders
package routes

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/observiq/bindplane-op/docs/swagger"
	"github.com/observiq/bindplane-op/graphql"
	"github.com/observiq/bindplane-op/internal/opamp"
	"github.com/observiq/bindplane-op/middleware"
	"github.com/observiq/bindplane-op/otlp"
	"github.com/observiq/bindplane-op/rest"
	"github.com/observiq/bindplane-op/server"
	"github.com/observiq/bindplane-op/sessions"
	"github.com/observiq/bindplane-op/ui"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

var _ server.RouteBuilder = (*CombinedRouteBuilder)(nil)

// CombinedRouteBuilder the combined router adds routes for all endpoints in the server
type CombinedRouteBuilder struct{}

// AddRoutes adds routes for the following:
// - sessions
// - rest
// - graphql
// - swagger
// - otlp
// - opamp
// - ui
func (c *CombinedRouteBuilder) AddRoutes(router gin.IRouter, bindplane server.BindPlane) error {
	// Order that routes are added matters!!!
	sessions.AddRoutes(router, bindplane)

	v1 := router.Group("/v1")
	v1.Use(otelgin.Middleware("bindplane"))

	authv1 := v1.Group("/", middleware.Chain(bindplane)...)
	rest.AddRestRoutes(authv1, bindplane)
	graphql.AddRoutes(authv1, bindplane)
	swagger.AddRoutes(router)

	// opamp does its own authorization based on the OnConnecting callback
	err := opamp.AddRoutes(v1, bindplane)
	if err != nil {
		return fmt.Errorf("failed to start OpAMP: %w", err)
	}
	otlp.AddRoutes(v1, bindplane)
	ui.AddRoutes(router, bindplane)

	return nil
}
