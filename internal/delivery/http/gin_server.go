package http

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/ichzzy/go_o11y_base/internal/app/cx"
	"github.com/ichzzy/go_o11y_base/internal/lib/logs"
	"github.com/ichzzy/go_o11y_base/internal/lib/shutdown"
)

type GinServer struct {
	httpServer      *http.Server
	shutdownHandler *shutdown.Handler
	middleware      *Middleware
	handlers        handlers
	mode            string
}

func NewGinServer(param GinServerParam) (*GinServer, error) {
	gin.SetMode(gin.ReleaseMode)
	if param.Config.App.Env == "local" {
		gin.SetMode(gin.DebugMode)
	}

	// Configure CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true

	// Configure Gin engine
	engine := gin.New()
	engine.Use(cors.New(corsConfig))

	// Configure routes
	server := &GinServer{
		mode:            param.Config.App.Env,
		middleware:      param.Middleware,
		shutdownHandler: param.ShutdownHandler,
	}
	server.setHandlers(param)

	if err := server.indexRoute(engine); err != nil {
		return nil, fmt.Errorf("register http routes failed: %w", err)
	}

	server.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", param.Config.Server.Http.Port),
		Handler:      engine,
		ReadTimeout:  param.Config.Server.Http.Timeout.ReadTimeout,
		WriteTimeout: param.Config.Server.Http.Timeout.WriteTimeout,
		IdleTimeout:  param.Config.Server.Http.Timeout.IdleTimeout,
	}

	return server, nil
}

func (s *GinServer) Start(ctx context.Context) {
	logger := cx.GetLogger(ctx)
	go func(logger logs.AppLogger) {
		defer func() {
			if r := recover(); r != nil {
				logger.WithFields(logs.Fields{
					"error": r,
					"stack": string(debug.Stack()),
				}).Error("panic recovered from gin server")
			}
		}()

		logger.Infof("gin.ListenAndServe start: %s", s.httpServer.Addr)

		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Error("gin.ListenAndServe fail")
		}
	}(logger)

	// Register Shutdown Hook
	s.shutdownHandler.Register(func(ctx context.Context) error {
		return s.Shutdown(ctx)
	})
}

func (s *GinServer) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("gin.Shutdown fail: %w", err)
	}
	return nil
}
