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

package authenticator

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

// BasicAuthenticator is an authenticator that uses the server profile username and password
type BasicAuthenticator struct {
	serverUsername string
	serverPassword string
}

// NewBasicAuthenticator creates an authenticator for internal server profile
func NewBasicAuthenticator(username, password string) Authenticator {
	return &BasicAuthenticator{
		serverUsername: username,
		serverPassword: password,
	}
}

// Verify attempts to login in the user from the session
func (a *BasicAuthenticator) Verify(c *gin.Context, session *sessions.Session) error {
	username, password, inSession := LoginIDPasswordFromSession(session)
	if !inSession || !(username == a.serverUsername && a.serverPassword == password) {
		return AbortWithError(c, ErrBadCreds)
	}
	return nil
}

// Login attempts to login in a user from the postform and returns the proper loginID
func (a *BasicAuthenticator) Login(c *gin.Context, session *sessions.Session) (*LoginInfo, error) {
	username := c.PostForm(UsernameKey)
	password := c.PostForm(PasswordKey)

	loginInfo := &LoginInfo{Username: username}

	if !(username == a.serverUsername && password == a.serverPassword) {
		return loginInfo, AbortWithError(c, ErrBadCreds)
	}

	loginInfo.LoginID = username
	session.Values[LoginKey] = username
	session.Values[PasswordKey] = password
	return loginInfo, nil
}

// Middleware returns Authentication middleware for Basic server profile verification
func (a *BasicAuthenticator) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		username, password, ok := c.Request.BasicAuth()
		if !ok {
			// Go to next middleware in chain, the final middleware will require authentication is set to true.
			c.Next()
			return
		}

		if username != a.serverUsername || password != a.serverPassword {
			_ = AbortWithError(c, ErrBadCreds)
			return
		}

		c.Set(LoginKey, username)
		c.Set(AuthenticatedKey, true)
	}
}
