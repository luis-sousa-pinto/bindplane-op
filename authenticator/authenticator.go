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

// Package authenticator provides a way to authenticate users using different methods.
package authenticator

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

const (
	// CookieName is the name of the cookie used for session authentication.
	CookieName = "BP_OP_AUTH"
	// LoginKey is a key that is used to identify the login id.
	LoginKey = "loginid"
	// PasswordKey is a key that is used to identify the password.
	PasswordKey = "p"
	// UsernameKey is a key that is used to identify the username.
	UsernameKey = "u"
	// AuthenticatedKey is a key that is used to identify if the user has been authenticated in the middleware.
	AuthenticatedKey = "authenticated"
)

// ErrBadCreds for invalid authentication credentials
var ErrBadCreds = errors.New("incorrect username or password")

// ErrBadBindCreds for invalid authentication credentials
var ErrBadBindCreds = errors.New("incorrect username or password for BindUser")

// ErrInvalidSession for when username or password are expected and not present
var ErrInvalidSession = errors.New("failed to retrieve session")

// ErrMissingCreds is an error for when a google authentication payload doesn't have credentials
var ErrMissingCreds = errors.New("missing field 'credential'")

// Authenticator represents a way to authenticate a given username and password
//
//go:generate mockery --name=Authenticator --filename=mock_authenticator.go --structname=MockAuthenticator
type Authenticator interface {

	// Login takes in a ctx that has the username and password attached to it in some way. currently in either the Postform or the request Header.
	// This function should store some form of the user on the session to be used in the verify function.
	// The string returned is the unique identifier of the user that was just logged in
	// 			Ex: "email" in google, "DN (distinguished name)" in LDAP, or just the authenticator.LoginSession for system auth.
	// Login is expected to atleast return a username on the LoginInfo for logging purposes.
	Login(ctx *gin.Context, session *sessions.Session) (*LoginInfo, error)

	// Verify pulls the information Login put on the session and checks that those credentials are still valid
	// this should mainly be used in verifying established UI connections
	Verify(ctx *gin.Context, session *sessions.Session) error

	// Middleware authenticates a user that is directly using the REST endpoints. Nothing needs to be saved on the session for this function.
	// once there is verification that the user is authenticated, authenticatedKey should be set to "true" on the context
	// It is expected that loginID is set on the context to be used by the check user middleware. if the display name is different than the loginID, set usernameKey.
	Middleware() gin.HandlerFunc
}

// LoginInfo is a representation of user that has logged in.
type LoginInfo struct {
	LoginID  string
	Username string
}

// LoginIDPasswordFromSession returns the login id and password from the session.
func LoginIDPasswordFromSession(session *sessions.Session) (login, password string, inSession bool) {
	interLogin := session.Values[LoginKey]
	interPassword := session.Values[PasswordKey]

	login, ok := interLogin.(string)
	if !ok {
		return "", "", false
	}

	password, ok = interPassword.(string)
	if !ok {
		return "", "", false
	}
	return login, password, true
}

// LoginIDFromSession returns the login id from the session.
func LoginIDFromSession(session *sessions.Session) (login string, inSession bool) {
	interLogin := session.Values[LoginKey]

	login, ok := interLogin.(string)
	if !ok {
		return "", false
	}

	return login, true
}

// AbortWithError aborts the gin context with the given error.
func AbortWithError(c *gin.Context, err error) error {
	switch err {
	case ErrBadCreds:
		c.AbortWithError(http.StatusUnauthorized, err)
	case ErrBadBindCreds:
		c.AbortWithError(http.StatusUnauthorized, err)
	case ErrInvalidSession:
		c.AbortWithError(http.StatusUnauthorized, err)
	case ErrMissingCreds:
		c.AbortWithError(http.StatusBadRequest, err)
	default:
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	return err
}

// HandleSessionError handles the gin context and session after a non nil error is
// returned while getting session.  It will clear the session and abort with error.
// Must call return during the gin handler function directly after.
func HandleSessionError(c *gin.Context, session sessions.Session, err error) {
	// Clear the cookie, this can happen when sessions-secrets change
	// and we see a cookie with the previous secret is read.
	session.Options.MaxAge = -1

	saveErr := session.Save(c.Request, c.Writer)
	if saveErr != nil {
		c.AbortWithError(http.StatusInternalServerError, saveErr)
		return
	}

	c.AbortWithError(http.StatusUnauthorized, err)
	return
}
