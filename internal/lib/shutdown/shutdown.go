package shutdown

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ichzzy/go_o11y_base/internal/lib/logs"
)

type Callback func(ctx context.Context) error

type Handler struct {
	logger    logs.AppLogger
	callbacks []Callback
	timeout   time.Duration
}

func New(logger logs.AppLogger, timeout time.Duration) *Handler {
	return &Handler{
		logger:  logger,
		timeout: timeout,
	}
}

// Register adds a callback to be executed during shutdown.
func (h *Handler) Register(callback Callback) {
	h.callbacks = append(h.callbacks, callback)
}

// Listen block until signal received, then execute registered callbacks.
func (h *Handler) Listen(ctx context.Context) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	s := <-quit
	h.logger.WithField("signal", s.String()).Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	// LIFO
	for i := len(h.callbacks) - 1; i >= 0; i-- {
		if err := h.callbacks[i](shutdownCtx); err != nil {
			h.logger.WithError(err).Error("shutdown callback failed")
		}
	}

	h.logger.Info("server exited properly")
}
