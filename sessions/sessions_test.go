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

package sessions

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	gorillaSessions "github.com/gorilla/sessions"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/observiq/bindplane-op/authenticator"
	"github.com/observiq/bindplane-op/config"
	"github.com/observiq/bindplane-op/internal/server"
	exposedserver "github.com/observiq/bindplane-op/server"
	"github.com/observiq/bindplane-op/store"
	storeMocks "github.com/observiq/bindplane-op/store/mocks"
	statsmocks "github.com/observiq/bindplane-op/store/stats/mocks"
)

func TestAddRoutes(t *testing.T) {
	// Setup
	router := gin.Default()
	svr := httptest.NewServer(router)
	defer svr.Close()

	logger := zap.NewNop()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := store.NewMapStore(ctx, store.Options{
		SessionsSecret:   "super-secret-key",
		MaxEventsToMerge: 1,
	}, logger)

	mockBatcher := statsmocks.NewMockMeasurementBatcher(t)
	bindplane := server.NewBindPlane(&config.Config{}, zap.NewNop(), s, nil, mockBatcher)

	t.Run("adds /login /logout and /verify", func(t *testing.T) {
		AddRoutes(router, bindplane)

		routes := router.Routes()

		var hasLogin bool
		var hasLogout bool
		var hasVerify bool

		for _, r := range routes {
			switch r.Path {
			case "/login":
				hasLogin = true
			case "/logout":
				hasLogout = true
			case "/verify":
				hasVerify = true
			}
		}

		require.True(t, hasLogin)
		require.True(t, hasLogout)
		require.True(t, hasVerify)
	})
}

func TestHandleLogin(t *testing.T) {
	// Setup
	cfg := &config.Config{}
	cfg.Auth.Password = "secret"
	cfg.Auth.Username = "user"

	router := gin.Default()
	svr := httptest.NewServer(router)
	defer svr.Close()

	logger := zap.NewNop()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := store.NewMapStore(ctx, store.Options{
		SessionsSecret:   "super-secret-key",
		MaxEventsToMerge: 1,
	}, logger)
	mockBatcher := statsmocks.NewMockMeasurementBatcher(t)
	bindplane := server.NewBindPlane(cfg, zap.NewNop(), s, nil, mockBatcher)
	AddRoutes(router, bindplane)

	t.Run(fmt.Sprintf("sets the %s cookie with correct credentials", authenticator.CookieName), func(t *testing.T) {
		client := resty.New()
		client.SetBaseURL(svr.URL)

		resp, err := client.R().SetFormData(
			map[string]string{
				"u": "user",
				"p": "secret",
			},
		).Post("/login")

		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())

		var foundCookie bool
		for _, cookie := range resp.Cookies() {
			if strings.Contains(cookie.Name, authenticator.CookieName) {
				foundCookie = true
			}
		}

		require.True(t, foundCookie)
	})

	t.Run("will not set a cookie for bad credentials", func(t *testing.T) {
		client := resty.New()
		client.SetBaseURL(svr.URL)

		resp, err := client.R().SetFormData(
			map[string]string{
				"u": "user",
				"p": "bad-secret",
			},
		).Post("/login")

		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode())
		require.Empty(t, resp.Cookies())
	})
}

func TestLogin(t *testing.T) {
	// Setup
	cfg := &config.Config{}
	testUser := "user"
	testPass := "secret"
	cfg.Auth.Password = testPass
	cfg.Auth.Username = testUser

	logger := zap.NewNop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := store.NewMapStore(ctx, store.Options{
		SessionsSecret:   "super-secret-key",
		MaxEventsToMerge: 1,
	}, logger)

	mockBatcher := statsmocks.NewMockMeasurementBatcher(t)
	bindplane := server.NewBindPlane(cfg, zap.NewNop(), s, nil, mockBatcher)

	t.Run("will not set authenticated to true for invalid creds", func(t *testing.T) {
		// Create a Post Form Request with username and password
		req := httptest.NewRequest("POST", "/login", nil)
		req.PostForm = url.Values{
			"u": []string{"user"},
			"p": []string{"bad-secret"},
		}

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req

		// Log in with context
		Login(ctx, bindplane)

		// Make sure authenticated is set
		session, err := bindplane.Store().UserSessions().Get(ctx.Request, authenticator.CookieName)
		require.NoError(t, err)
		require.Nil(t, session.Values[authenticator.AuthenticatedKey], "expect the authenticated key to not be set")
	})

	t.Run("sets authenticated to true on the cookie with valid creds", func(t *testing.T) {
		// Create a Post Form Request with username and password
		req := httptest.NewRequest("POST", "/login", nil)
		req.PostForm = url.Values{
			"u": []string{"user"},
			"p": []string{"secret"},
		}

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req

		// Log in with context
		Login(ctx, bindplane)

		// Make sure username and password  is set
		session, err := bindplane.Store().UserSessions().Get(ctx.Request, authenticator.CookieName)
		require.NoError(t, err)
		require.Equal(t, "user", session.Values[authenticator.LoginKey])
		require.Equal(t, "secret", session.Values[authenticator.PasswordKey])
	})

	t.Run("will continue with login when the session store Get returns an error", func(t *testing.T) {
		mockStore := storeMocks.NewMockStore(t)
		mockStore.On("UserSessions").Return(&mockCookieStore{})

		mockBatcher := statsmocks.NewMockMeasurementBatcher(t)
		mockServer := server.NewBindPlane(cfg, zap.NewNop(), mockStore, nil, mockBatcher)

		req := httptest.NewRequest("POST", "/login", nil)
		req.PostForm = url.Values{
			"u": []string{"user"},
			"p": []string{"secret"},
		}

		w := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req

		// Login
		Login(ctx, mockServer)
		require.Equal(t, w.Result().StatusCode, http.StatusOK)
	})
}

type mockCookieStore struct{}

var _ gorillaSessions.Store = (*mockCookieStore)(nil)

// Get returns a session and an error for testing
func (m *mockCookieStore) Get(_ *http.Request, _ string) (*gorillaSessions.Session, error) {
	return gorillaSessions.NewSession(m, authenticator.CookieName), errors.New("error")

}

// New returns an invalid session for testing
func (m *mockCookieStore) New(_ *http.Request, _ string) (*gorillaSessions.Session, error) {
	return nil, nil
}

// Save returns nil for testing
func (m *mockCookieStore) Save(_ *http.Request, _ http.ResponseWriter, _ *gorillaSessions.Session) error {
	return nil
}

func TestLogout(t *testing.T) {
	// Setup
	cfg := &config.Config{}
	cfg.Auth.Password = "secret"
	cfg.Auth.Username = "user"

	logger := zap.NewNop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := store.NewMapStore(ctx, store.Options{
		SessionsSecret:   "super-secret-key",
		MaxEventsToMerge: 1,
	}, logger)

	mockBatcher := statsmocks.NewMockMeasurementBatcher(t)
	bindplane := server.NewBindPlane(cfg, zap.NewNop(), s, nil, mockBatcher)

	t.Run("will set username and password to empty for a logged in context", func(t *testing.T) {
		cookie := getLoggedInCookie(t, bindplane)

		// Make a logout request with the cookie we got from the login request
		logoutReq := httptest.NewRequest("PUT", "/logout", nil)
		logoutReq.AddCookie(cookie)

		logoutCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
		logoutCtx.Request = logoutReq

		// log the context out
		Logout(logoutCtx, bindplane)

		// verify password and username is not set
		session, _ := bindplane.Store().UserSessions().Get(logoutCtx.Request, authenticator.CookieName)
		require.Equal(t, "", session.Values[authenticator.LoginKey].(string))
		require.Equal(t, "", session.Values[authenticator.PasswordKey].(string))
	})
}

func TestVerify(t *testing.T) {
	// Setup
	cfg := &config.Config{}
	cfg.Auth.Password = "secret"
	cfg.Auth.Username = "user"

	logger := zap.NewNop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := store.NewMapStore(ctx, store.Options{
		SessionsSecret:   "super-secret-key",
		MaxEventsToMerge: 1,
	}, logger)

	mockBatcher := statsmocks.NewMockMeasurementBatcher(t)
	bindplane := server.NewBindPlane(cfg, zap.NewNop(), s, nil, mockBatcher)

	t.Run("aborts with status 401 when username is unset", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/verify", nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req

		Verify(ctx, bindplane)

		session, err := s.UserSessions().Get(req, authenticator.CookieName)
		require.NoError(t, err)
		require.Equal(t, session.Values[authenticator.LoginKey], nil)

		require.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
	})

	t.Run("aborts with status 401 when password is unset", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/verify", nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req

		Verify(ctx, bindplane)

		session, err := s.UserSessions().Get(req, authenticator.CookieName)
		require.NoError(t, err)
		require.Equal(t, nil, session.Values[authenticator.LoginKey])

		require.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
	})
	t.Run("aborts with status 401 when username and password are incorrect", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/verify", nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req

		Verify(ctx, bindplane)
		session, err := s.UserSessions().Get(req, authenticator.CookieName)
		require.NoError(t, err)
		require.Equal(t, session.Values[authenticator.LoginKey], nil)

		require.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
	})

}

func getLoggedInContext(_ *testing.T, bindplane exposedserver.BindPlane, w http.ResponseWriter) *gin.Context {
	// Make a login request
	loginRequest := httptest.NewRequest("POST", "/login", nil)
	loginRequest.PostForm = url.Values{
		"u": []string{"user"},
		"p": []string{"secret"},
	}

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = loginRequest

	Login(ctx, bindplane)
	return ctx
}

func getLoggedInCookie(t *testing.T, bindplane exposedserver.BindPlane) *http.Cookie {
	w := httptest.NewRecorder()
	getLoggedInContext(t, bindplane, w)

	return w.Result().Cookies()[0]
}
