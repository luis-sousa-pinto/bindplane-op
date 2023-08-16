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

package routes

import (
	"testing"

	"github.com/gin-gonic/gin"
	authmocks "github.com/observiq/bindplane-op/authenticator/mocks"
	"github.com/observiq/bindplane-op/eventbus"
	"github.com/observiq/bindplane-op/server"
	servermocks "github.com/observiq/bindplane-op/server/mocks"
	"github.com/observiq/bindplane-op/store"
	storemocks "github.com/observiq/bindplane-op/store/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCombinedRouteBuilderAddRoutes(t *testing.T) {
	router := gin.New()

	routeBuilder := &CombinedRouteBuilder{}

	mockAuthenticator := authmocks.NewMockAuthenticator(t)
	mockAuthenticator.On("Middleware").Return(gin.HandlerFunc(func(ctx *gin.Context) {}))

	mockStore := storemocks.NewMockStore(t)
	mockStore.On("Updates", mock.Anything).Return(eventbus.NewSource[store.BasicEventUpdates]())

	mockManager := servermocks.NewMockManager(t)
	mockManager.On("EnableProtocol", mock.Anything).Return()
	mockManager.On("Store", mock.Anything).Return(mockStore)
	mockManager.On("AgentMessages", mock.Anything).Maybe().Return(eventbus.NewSource[server.Message]())

	mockBindplane := servermocks.NewMockBindPlane(t)
	mockBindplane.On("Authenticator").Return(mockAuthenticator)
	mockBindplane.On("Store").Return(mockStore)
	mockBindplane.On("Logger").Return(zap.NewNop())
	mockBindplane.On("Manager").Return(mockManager)
	mockBindplane.On("BindPlaneURL").Return("https://localhost:3001")

	err := routeBuilder.AddRoutes(router, mockBindplane)
	require.NoError(t, err)
}
