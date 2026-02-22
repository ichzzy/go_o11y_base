package main

import (
	"context"
	"fmt"
	_ "net/http/pprof"
	"os"
	"runtime"

	"github.com/ichzzy/go_o11y_base/internal/app/cx"
	"github.com/ichzzy/go_o11y_base/internal/app/di"
	"github.com/ichzzy/go_o11y_base/internal/config"
	"github.com/ichzzy/go_o11y_base/internal/delivery/http"
	"github.com/ichzzy/go_o11y_base/internal/domain"
	domain_uc "github.com/ichzzy/go_o11y_base/internal/domain/usecase"
	"github.com/ichzzy/go_o11y_base/internal/lib/logs"
	"github.com/ichzzy/go_o11y_base/internal/lib/profile"
	"github.com/ichzzy/go_o11y_base/internal/lib/shutdown"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	config, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("config.Load failed, err: %w", err))
	}

	domain.ServiceName = config.App.ServiceName
	podName := os.Getenv("POD_NAME")
	if podName != "" {
		config.App.PodName = podName
	}

	appLogger, err := logs.NewAppLogger(config)
	if err != nil {
		panic(fmt.Errorf("log.New failed, err: %w", err))
	}
	appLogger = appLogger.WithFields(logs.Fields{
		"app":  config.App.ServiceName,
		"type": "server",
	})

	ctx = cx.SetLogger(ctx, appLogger)

	appLogger.WithFields(logrus.Fields{
		"node_cpu":     runtime.NumCPU(),
		"app_maxprocs": runtime.GOMAXPROCS(0),
		"pod_name":     config.App.PodName,
	}).Warn("app info")

	// Register DI containers
	container := di.NewContainer()
	if err := container.RegisterBase(config, appLogger); err != nil {
		panic(fmt.Errorf("register base failed: %w", err))
	}
	if err := container.RegisterInfra(config); err != nil {
		panic(fmt.Errorf("register infra failed: %w", err))
	}
	if err := container.RegisterDomain(); err != nil {
		panic(fmt.Errorf("register domain failed: %w", err))
	}
	if err := container.RegisterDeliveryHTTP(); err != nil {
		panic(fmt.Errorf("register delivery failed: %w", err))
	}

	// Load RBAC policies into memory
	if err := container.Invoke(func(rbacUC domain_uc.RbacUsecase) error {
		return rbacUC.LoadPolicy(ctx)
	}); err != nil {
		panic(fmt.Errorf("initial load rbac policy failed: %w", err))
	}

	// Start servers
	if err := container.Invoke(func(ginServer *http.GinServer, pprof *profile.Pprof) {
		ginServer.Start(ctx)
		pprof.Start(ctx)
	}); err != nil {
		panic(fmt.Errorf("start servers failed: %w", err))
	}

	// Wait for shutdown signal
	if err := container.Invoke(func(sh *shutdown.Handler) {
		sh.Listen(ctx)
	}); err != nil {
		panic(fmt.Errorf("shutdown handler invoke failed: %w", err))
	}

	cancel()
}
