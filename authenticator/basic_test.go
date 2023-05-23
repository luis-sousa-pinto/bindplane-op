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
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/observiq/bindplane-op/store"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestBasicLogin(t *testing.T) {
	tcs := []struct {
		name            string
		auth            Authenticator
		username        string
		password        string
		expectedLoginID *LoginInfo
		expectedSession func(*testing.T, *sessions.Session)
		expectedErr     error
	}{
		{
			name:            "Valid",
			auth:            NewBasicAuthenticator("admin", "adminPassword"),
			username:        "admin",
			password:        "adminPassword",
			expectedLoginID: &LoginInfo{LoginID: "admin", Username: "admin"},
			expectedSession: func(t *testing.T, s *sessions.Session) {
				require.Equal(t, "admin", s.Values[LoginKey])
				require.Equal(t, "adminPassword", s.Values[PasswordKey])
			},
		},
		{
			name:            "Invalid username",
			auth:            NewBasicAuthenticator("admin", "adminPassword"),
			username:        "bad-user",
			password:        "adminPassword",
			expectedLoginID: &LoginInfo{Username: "bad-user"},
			expectedSession: func(t *testing.T, s *sessions.Session) {
				require.Equal(t, nil, s.Values[LoginKey])
				require.Equal(t, nil, s.Values["adminPassword"])
			},
			expectedErr: ErrBadCreds,
		},
		{
			name:            "Invalid Password",
			auth:            NewBasicAuthenticator("admin", "adminPassword"),
			username:        "admin",
			password:        "bad-password",
			expectedLoginID: &LoginInfo{Username: "admin"},
			expectedSession: func(t *testing.T, s *sessions.Session) {
				require.Equal(t, nil, s.Values[LoginKey])
				require.Equal(t, nil, s.Values["adminPassword"])
			},
			expectedErr: ErrBadCreds,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx := testGinContext(t, nil)
			ctx.Request.PostForm.Add(UsernameKey, tc.username)
			ctx.Request.PostForm.Add(PasswordKey, tc.password)

			mpStore := store.NewMapStore(ctx, store.Options{
				SessionsSecret:   "super-secret-key",
				MaxEventsToMerge: 1,
			}, zap.NewNop())

			session, _ := mpStore.UserSessions().Get(ctx.Request, CookieName)
			loginID, err := tc.auth.Login(ctx, session)
			if tc.expectedErr != nil {
				require.Equal(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
			}
			tc.expectedSession(t, session)

			require.Equal(t, tc.expectedLoginID, loginID)
		})
	}
}

func TestBasicVerify(t *testing.T) {
	tcs := []struct {
		name        string
		auth        Authenticator
		username    string
		password    string
		expectedErr error
	}{
		{
			name:     "Valid",
			auth:     NewBasicAuthenticator("admin", "adminPassword"),
			username: "admin",
			password: "adminPassword",
		},
		{
			name:        "Invalid username",
			auth:        NewBasicAuthenticator("admin", "adminPassword"),
			username:    "bad-user",
			password:    "adminPassword",
			expectedErr: ErrBadCreds,
		},
		{
			name:        "Invalid password",
			auth:        NewBasicAuthenticator("admin", "adminPassword"),
			username:    "admin",
			password:    "bad-Password",
			expectedErr: ErrBadCreds,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx := testGinContext(t, nil)
			mpStore := store.NewMapStore(ctx, store.Options{
				SessionsSecret:   "super-secret-key",
				MaxEventsToMerge: 1,
			}, zap.NewNop())

			session, _ := mpStore.UserSessions().Get(ctx.Request, CookieName)
			session.Values[LoginKey] = tc.username
			session.Values[PasswordKey] = tc.password

			err := tc.auth.Verify(ctx, session)
			if tc.expectedErr != nil {
				require.Equal(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestBasicMiddlWare(t *testing.T) {
	tcs := []struct {
		name        string
		auth        Authenticator
		username    string
		password    string
		expectedErr error
	}{
		{
			name:     "Valid",
			auth:     NewBasicAuthenticator("admin", "adminPassword"),
			username: "admin",
			password: "adminPassword",
		},
		{
			name:        "Invalid username",
			auth:        NewBasicAuthenticator("admin", "adminPassword"),
			username:    "bad-user",
			password:    "adminPassword",
			expectedErr: ErrBadCreds,
		},
		{
			name:        "Invalid password",
			auth:        NewBasicAuthenticator("admin", "adminPassword"),
			username:    "admin",
			password:    "bad-Password",
			expectedErr: ErrBadCreds,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx := testGinContext(t, nil)
			ctx.Request.SetBasicAuth(tc.username, tc.password)
			middleware := tc.auth.Middleware()
			middleware(ctx)
			if tc.expectedErr != nil {
				require.Equal(t, tc.expectedErr, ctx.Errors.Last().Err)
			} else {
				require.Nil(t, ctx.Errors.Last())
				require.Equal(t, true, ctx.Value(AuthenticatedKey))
			}

		})
	}
}

func testGinContext(_ *testing.T, r io.Reader) *gin.Context {
	// Make a login request
	w := httptest.NewRecorder()

	var loginRequest *http.Request
	if r != nil {
		loginRequest = httptest.NewRequest("POST", "/login", r)
	} else {
		loginRequest = httptest.NewRequest("POST", "/login", nil)
	}

	loginRequest.PostForm = url.Values{}

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = loginRequest

	return ctx
}
