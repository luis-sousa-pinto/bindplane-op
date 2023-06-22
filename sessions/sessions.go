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

// Package sessions provides session management for the UI
package sessions

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/observiq/bindplane-op/authenticator"
	exposedserver "github.com/observiq/bindplane-op/server"
	"go.uber.org/zap"
)

// Login is the handler for the login page
func Login(ctx *gin.Context, bindplane exposedserver.BindPlane) {
	// Ignore this error because we're gauranteed to get a session here, if
	// the cookie was invalid for any reason we'll overwrite it.
	session, err := bindplane.Store().UserSessions().Get(ctx.Request, authenticator.CookieName)
	if err != nil {
		bindplane.Logger().Error("unexpected error when retrieving session at login", zap.Error(err))
	}

	loginInfo, err := bindplane.Authenticator().Login(ctx, session)
	if err != nil {
		bindplane.Logger().Error(fmt.Sprintf("failed to authenticate user: %s", loginInfo.Username), zap.Error(err))
		return
	}

	bindplane.Logger().Info("logging in user.", zap.String("user", loginInfo.Username))

	// Save and write the session
	if err := session.Save(ctx.Request, ctx.Writer); err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New("failed to save session"))
		bindplane.Logger().Error("failed to save session after login", zap.Error(err))
	}
}

// Logout is the handler for the logout button
func Logout(ctx *gin.Context, bindplane exposedserver.BindPlane) {
	session, err := bindplane.Store().UserSessions().Get(ctx.Request, authenticator.CookieName)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New("failed to retrieve session"))
		bindplane.Logger().Debug("failed to retrieve session during verify", zap.Error(err))
		return
	}

	// Revoke users authentication
	session.Values[authenticator.LoginKey] = ""
	session.Values[authenticator.PasswordKey] = ""
	// Delete the cookie
	session.Options.MaxAge = -1

	bindplane.Logger().Info("logging out user.", zap.Any("user", session.Values["user"]))
	// Save and write the session
	if err := session.Save(ctx.Request, ctx.Writer); err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New("failed to save session"))
		bindplane.Logger().Error("failed to save session after logout", zap.Error(err))
	}
}

// Verify is the handler for the verify route that checks if the user is authenticated
func Verify(c *gin.Context, bindplane exposedserver.BindPlane) {
	session, err := bindplane.Store().UserSessions().Get(c.Request, authenticator.CookieName)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, errors.New("failed to retrieve session"))
		bindplane.Logger().Debug("failed to retrieve session during verify", zap.Error(err))
		return
	}

	err = bindplane.Authenticator().Verify(c, session)
	if err != nil {
		bindplane.Logger().Debug("session invalid during verify", zap.Error(err))
	}
}

// AddRoutes adds the login, logout, and verify route used for session authentication.
func AddRoutes(router gin.IRouter, bindplane exposedserver.BindPlane) {
	router.POST("/login", func(ctx *gin.Context) { Login(ctx, bindplane) })
	router.PUT("/logout", func(ctx *gin.Context) { Logout(ctx, bindplane) })
	router.GET("/verify", func(ctx *gin.Context) { Verify(ctx, bindplane) })
}
