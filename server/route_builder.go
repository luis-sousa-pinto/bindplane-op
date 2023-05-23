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

package server

import (
	"github.com/gin-gonic/gin"
)

// RouteBuilder builds routes for a given router
//
//go:generate mockery --name=RouteBuilder --filename=mock_route_builder.go --structname=MockRouteBuilder
type RouteBuilder interface {
	// AddRoutes adds routes to the passed in router
	AddRoutes(gin.IRouter, BindPlane) error
}
