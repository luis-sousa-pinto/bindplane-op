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

package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/observiq/bindplane-op/authenticator"
	exposedserver "github.com/observiq/bindplane-op/server"
)

// CheckSession checks to see if the attached cookie session is authenticated
// and if so sets authenticated to true on the context.  If not authenticated it
// goes to the next handler.
func CheckSession(server exposedserver.BindPlane) gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := server.Store().UserSessions().Get(c.Request, authenticator.CookieName)
		if err != nil {
			authenticator.HandleSessionError(c, *session, err)
			return
		}

		// Check the username and password values in the session storage - if either are unset, go to next handler
		if session.Values[authenticator.LoginKey] == nil {
			c.Next()
			return
		}

		c.Set(authenticator.AuthenticatedKey, true)
		// Extend the cookies life by 60 minutes since the user is active and making requests.
		session.Options.MaxAge = 60 * 60
		err = session.Save(c.Request, c.Writer)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}
}
