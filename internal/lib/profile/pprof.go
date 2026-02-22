package profile

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/ichzzy/go_o11y_base/internal/app/cx"
	"github.com/ichzzy/go_o11y_base/internal/config"
	"github.com/ichzzy/go_o11y_base/internal/lib/logs"
	"github.com/ichzzy/go_o11y_base/internal/lib/shutdown"
)

// curl --location 'http://localhost:6060/debug/pprof/heap' --output heap.out
type Pprof struct {
	config     config.PprofConfig
	httpServer *http.Server
	sh         *shutdown.Handler
}

func NewPprofServer(cfg config.PprofConfig, sh *shutdown.Handler) *Pprof {
	return &Pprof{
		config: cfg,
		sh:     sh,
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Port),
			Handler:      nil,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
	}
}

func (p *Pprof) Start(ctx context.Context) {
	if !p.config.Enabled {
		return
	}

	// Register Shutdown Hook
	p.sh.Register(func(ctx context.Context) error {
		p.Shutdown(ctx)
		return nil
	})

	logger := cx.GetLogger(ctx)
	go func(logger logs.AppLogger) {
		defer func() {
			if r := recover(); r != nil {
				logger.WithFields(logs.Fields{
					"error": r,
					"stack": string(debug.Stack()),
				}).Error("panic recovered from pprof server")
			}
		}()

		logger.Infof("pprof.ListenAndServe start: %s", p.httpServer.Addr)

		if err := p.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Error("pprof.ListenAndServe fail")
		}
	}(logger)
}

func (p *Pprof) Shutdown(ctx context.Context) {
	if p.httpServer == nil {
		return
	}
	if err := p.httpServer.Shutdown(ctx); err != nil {
		cx.GetLogger(ctx).WithError(err).Error("pprof.Shutdown fail")
	}
}
