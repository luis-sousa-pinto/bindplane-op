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

package serve

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	cors "github.com/itsjamie/gin-cors"
	"github.com/observiq/bindplane-op/agent"
	"github.com/observiq/bindplane-op/config"
	bpserver "github.com/observiq/bindplane-op/internal/server"
	"github.com/observiq/bindplane-op/metrics"
	"github.com/observiq/bindplane-op/resources"
	exposedserver "github.com/observiq/bindplane-op/server"
	"github.com/observiq/bindplane-op/stopqueue"
	"github.com/observiq/bindplane-op/store"
	"github.com/observiq/bindplane-op/store/stats"
	"github.com/observiq/bindplane-op/tracer"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

const (
	defaultReadTimeout     = 20 * time.Second
	defaultWriteTimeout    = 20 * time.Second
	defaultIdleTimeout     = 60 * time.Second
	defaultShutdownTimeout = 20 * time.Second
)

// Server is an interface for serving BindPlane.
//
//go:generate mockery --inpackage --with-expecter --name Server --filename mock_server.go --structname MockServer
type Server interface {
	// Seed will seed the BindPlane server.
	Seed(ctx context.Context) error

	// Serve will run a BindPlane server until an error occurs or the context is canceled.
	Serve(ctx context.Context) error
}

// Builder is an interface for building a Server.
//
//go:generate mockery --inpackage --with-expecter --name Builder --filename mock_builder.go --structname MockBuilder
type Builder interface {
	// Build returns a new Server.
	BuildServer(ctx context.Context) (Server, error)

	// SupportsServer returns true if the OS supports the serve command.
	// i.e. its not windows.
	SupportsServer() bool
}

var mp metric.Meter = otel.Meter("server")

// NewServer returns a new Server.
func NewServer(cfg *config.Config, s store.Store, tracer tracer.Tracer,
	logger *zap.Logger, mp metrics.Provider, routeBuilder exposedserver.RouteBuilder,
) Server {
	return &defaultServer{
		cfg:          cfg,
		store:        s,
		tracer:       tracer,
		logger:       logger,
		mp:           mp,
		routeBuilder: routeBuilder,
		stopQueue:    stopqueue.NewStopQueue(),
	}
}

// defaultServer is the default implementation of the Server interface.
type defaultServer struct {
	cfg          *config.Config
	store        store.Store
	tracer       tracer.Tracer
	logger       *zap.Logger
	mp           metrics.Provider
	routeBuilder exposedserver.RouteBuilder
	stopQueue    *stopqueue.Queue
}

// Seed will seed the BindPlane server.
func (s *defaultServer) Seed(ctx context.Context) error {
	if err := store.Seed(ctx, s.store, s.logger, resources.Files, resources.SeedFolders); err != nil {
		s.logger.Error("Failed to seed resources on startup", zap.Error(err))
	}

	configurations, err := s.store.Configurations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get configurations: %w", err)
	}

	for _, c := range configurations {
		if err := s.store.ConfigurationIndex(ctx).Upsert(c); err != nil {
			return fmt.Errorf("failed to seed configuration index: %s, %w", c.ID(), err)
		}
	}

	agents, err := s.store.Agents(ctx)
	if err != nil {
		return fmt.Errorf("failed to get agents: %w", err)
	}

	for _, a := range agents {
		if err := s.store.AgentIndex(ctx).Upsert(a); err != nil {
			return fmt.Errorf("failed to seed agent index: %s, %w", a.ID, err)
		}
	}

	return nil
}

// Serve will run a BindPlane server until an error occurs or the context is canceled.
func (s *defaultServer) Serve(ctx context.Context) error {
	agentVersions := s.createAgentVersions(ctx)

	batcher := s.createMeasurementBatcher(ctx)

	bindplane := bpserver.NewBindPlane(s.cfg, s.logger, s.store, agentVersions, batcher)

	s.startManager(ctx, bindplane)

	s.setGinMode()
	router, err := s.createRouter(bindplane)
	if err != nil {
		return fmt.Errorf("failed to create router: %w", err)
	}

	httpServer, err := s.createHTTPServer(router)
	if err != nil {
		return fmt.Errorf("failed to create http server: %w", err)
	}

	s.startScheduler(ctx)

	s.startTracer(ctx)

	s.startMetrics(ctx)

	serverErr := make(chan error, 1)
	go func() {
		switch httpServer.TLSConfig {
		case nil:
			serverErr <- httpServer.ListenAndServe()
		default:
			serverErr <- httpServer.ListenAndServeTLS("", "")
		}
	}()

	signalCtx, signalCancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer signalCancel()

	select {
	case <-signalCtx.Done():
		timedCtx, timedCancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
		defer timedCancel()

		s.logger.Info("Context canceled, stopping server")

		var shutDownErrs error
		if err := httpServer.Shutdown(timedCtx); err != nil {
			s.logger.Error("Server error while shutting down", zap.Error(err))
			shutDownErrs = errors.Join(shutDownErrs, err)
		}

		// stop all dependent objects after server can no long accept calls
		if err := s.stopQueue.StopAll(timedCtx); err != nil {
			shutDownErrs = errors.Join(shutDownErrs, err)
		}

		return shutDownErrs
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("Listen error", zap.Error(err))
			fmt.Println("listen error:", err.Error())
			os.Exit(200)
		}

		return nil
	}
}

// createAgentVersions will create the agent versions for the server.
func (s *defaultServer) createAgentVersions(ctx context.Context) agent.Versions {
	var versionClient agent.VersionClient
	if !s.cfg.Offline {
		versionClient = agent.NewGitHubVersionClient()
	} else {
		versionClient = agent.NewNoopClient()
	}

	settings := agent.VersionsSettings{
		Logger:                    s.logger.Named("versions"),
		SyncAgentVersionsInterval: s.cfg.AgentVersions.SyncInterval,
	}

	return agent.NewVersions(ctx, versionClient, s.store, settings)
}

// createMeasurementBatcher creates a measurement batcher and adds it to the stop queue
func (s *defaultServer) createMeasurementBatcher(ctx context.Context) stats.MeasurementBatcher {
	batcher := stats.NewDefaultBatcher(ctx, s.logger, s.store.Measurements())

	s.stopQueue.Add(func(stopCtx context.Context) error {
		return batcher.Shutdown(stopCtx)
	})

	return batcher
}

// setGinMode will set the gin mode based on the environment.
func (s *defaultServer) setGinMode() {
	switch s.cfg.Env {
	case config.EnvDevelopment:
		gin.SetMode(gin.DebugMode)
	case config.EnvTest:
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}
}

// createRouter will create the gin router.
func (s *defaultServer) createRouter(bindplane exposedserver.BindPlane) (*gin.Engine, error) {
	router := gin.New()
	if s.cfg.Env == config.EnvDevelopment {
		router.Use(gin.Logger())
	}
	router.Use(ginzap.Ginzap(s.logger, "", false))
	router.Use(ginzap.RecoveryWithZap(s.logger, true))

	corsMiddleware := cors.Middleware(cors.Config{
		Origins:        "*",
		Methods:        "GET, PUT, POST, DELETE",
		RequestHeaders: "Origin, Authorization, Content-Type",
		Credentials:    true,
	})
	router.Use(corsMiddleware)

	router.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	if err := s.routeBuilder.AddRoutes(router, bindplane); err != nil {
		return nil, fmt.Errorf("failed to add routes: %w", err)
	}
	return router, nil
}

// createHTTPServer will create the http server.
func (s *defaultServer) createHTTPServer(router *gin.Engine) (*http.Server, error) {
	httpServer := &http.Server{
		Addr:              s.cfg.Network.BindAddress(),
		Handler:           router,
		ReadTimeout:       defaultReadTimeout,
		ReadHeaderTimeout: defaultReadTimeout,
		WriteTimeout:      defaultWriteTimeout,
		IdleTimeout:       defaultIdleTimeout,
	}

	if s.cfg.Network.TLSEnabled() {
		tlsConfig, err := s.cfg.Network.TLS.Convert()
		if err != nil {
			return nil, fmt.Errorf("failed to configure tls: %w", err)
		}
		httpServer.TLSConfig = tlsConfig
	}

	return httpServer, nil
}

// startTracer starts the tracer and adds it's cleanup onto the stopQueue
func (s *defaultServer) startTracer(ctx context.Context) {
	if err := s.tracer.Start(ctx); err != nil {
		s.logger.Warn("Failed to start tracer", zap.Error(err))
	}

	s.stopQueue.Add(
		func(stopCtx context.Context) error {
			return s.tracer.Shutdown(stopCtx)
		},
	)
}

// startMetrics starts the tracer and adds it's cleanup onto the stopQueue
func (s *defaultServer) startMetrics(ctx context.Context) {
	if err := s.mp.Start(ctx); err != nil {
		s.logger.Warn("Failed to start metrics", zap.Error(err))
	}

	s.stopQueue.Add(
		func(stopCtx context.Context) error {
			return s.mp.Shutdown(stopCtx)
		},
	)
}

// startManager starts the bindplane manager
func (s *defaultServer) startManager(ctx context.Context, bindplane exposedserver.BindPlane) {
	bindplane.Manager().Start(ctx)

	s.stopQueue.Add(
		func(stopCtx context.Context) error {
			return bindplane.Manager().Shutdown(ctx)
		},
	)
}

func (s *defaultServer) startScheduler(ctx context.Context) {
	scheduler := exposedserver.NewScheduler(s.store, s.logger, s.cfg.RolloutsInterval)
	scheduler.Start(ctx)

	s.stopQueue.Add(
		func(stopCtx context.Context) error {
			return scheduler.Stop(stopCtx)
		},
	)
}
