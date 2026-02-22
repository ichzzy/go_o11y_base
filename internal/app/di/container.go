package di

import (
	"context"
	"fmt"
	"time"

	"github.com/ichzzy/go_o11y_base/internal/config"
	"github.com/ichzzy/go_o11y_base/internal/delivery/http"
	"github.com/ichzzy/go_o11y_base/internal/delivery/http/handler"
	"github.com/ichzzy/go_o11y_base/internal/lib/logs"
	"github.com/ichzzy/go_o11y_base/internal/lib/metrics"
	"github.com/ichzzy/go_o11y_base/internal/lib/mysql"
	"github.com/ichzzy/go_o11y_base/internal/lib/profile"
	redisInit "github.com/ichzzy/go_o11y_base/internal/lib/redis"
	"github.com/ichzzy/go_o11y_base/internal/lib/shutdown"
	"github.com/ichzzy/go_o11y_base/internal/lib/telemetry"
	"github.com/ichzzy/go_o11y_base/internal/repository"
	"github.com/ichzzy/go_o11y_base/internal/usecase"
	"github.com/redis/go-redis/v9"
	"go.uber.org/dig"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Container struct {
	container *dig.Container
}

func NewContainer() *Container {
	return &Container{
		container: dig.New(),
	}
}

func (c *Container) Invoke(fn interface{}) error {
	return c.container.Invoke(fn)
}

func (c *Container) Provide(fn interface{}, opts ...dig.ProvideOption) error {
	return c.container.Provide(fn, opts...)
}

func (c *Container) RegisterBase(conf *config.Config, logger logs.AppLogger) error {
	// Base Config
	if err := c.container.Provide(func() *config.Config {
		return conf
	}, dig.Name("config")); err != nil {
		return fmt.Errorf("provide config failed: %w", err)
	}

	// App Logger, Request 鏈路中務必使用 cx.GetLogger(ctx) 獲取 logger, 盡量避免使用 DI 或原生 Logrus 來獲取
	if err := c.container.Provide(func() logs.AppLogger {
		return logger
	}); err != nil {
		return fmt.Errorf("provide logger failed: %w", err)
	}

	// Shutdown Handler
	if err := c.container.Provide(func(l logs.AppLogger) *shutdown.Handler {
		return shutdown.New(l, 15*time.Second)
	}); err != nil {
		return fmt.Errorf("provide shutdown handler failed: %w", err)
	}

	return nil
}

func (c *Container) RegisterInfra(conf *config.Config) error {
	// pprof
	if err := c.container.Provide(func(param struct {
		dig.In
		Config          *config.Config `name:"config"`
		ShutdownHandler *shutdown.Handler
	}) *profile.Pprof {
		return profile.NewPprofServer(param.Config.Observability.Pprof, param.ShutdownHandler)
	}); err != nil {
		return fmt.Errorf("provide pprof server failed: %w", err)
	}

	// OpenTelemetry
	if err := c.container.Invoke(func(param struct {
		dig.In
		Config          *config.Config `name:"config"`
		ShutdownHandler *shutdown.Handler
	}) error {
		otelInstance, err := telemetry.InitOTEL(context.Background(), &telemetry.OtelConfig{
			Enabled:     param.Config.Observability.OTEL.Enabled,
			Endpoint:    param.Config.Observability.OTEL.Endpoint,
			ServiceName: param.Config.App.ServiceName,
			Env:         param.Config.App.Env,
			SampleRatio: param.Config.Observability.OTEL.SampleRatio,
		})
		if err != nil {
			return err
		}
		param.ShutdownHandler.Register(func(ctx context.Context) error {
			return otelInstance.Shutdown(ctx)
		})
		return nil
	}); err != nil {
		return fmt.Errorf("invoke init otel failed: %w", err)
	}

	// MySQL
	if err := c.container.Provide(func(param struct {
		dig.In
		Config          *config.Config `name:"config"`
		AppLogger       logs.AppLogger
		ShutdownHandler *shutdown.Handler
	}) (*gorm.DB, error) {
		var gormLogger logger.Interface
		if param.Config.Connections.Mysql.Gorm.LogEnabled {
			gormLogger = mysql.NewGormLogger(param.AppLogger, param.Config.Connections.Mysql.Gorm.LogSlowThreshold)
		} else {
			gormLogger = logger.Default.LogMode(logger.Silent)
		}

		m, err := mysql.NewMySQL(mysql.MySqlConfig{
			EnableOtel:  param.Config.Observability.OTEL.Enabled,
			GormLogger:  gormLogger,
			MySQLConfig: param.Config.Connections.Mysql,
		})
		if err != nil {
			return nil, err
		}
		param.ShutdownHandler.Register(func(ctx context.Context) error {
			return m.Shutdown(ctx)
		})
		return m.Orm, nil
	}, dig.Name("mysql")); err != nil {
		return fmt.Errorf("provide mysql failed: %w", err)
	}

	// Redis
	if err := c.container.Provide(func(param struct {
		dig.In
		Config          *config.Config `name:"config"`
		ShutdownHandler *shutdown.Handler
	}) (*redis.ClusterClient, error) {
		m, err := redisInit.NewRedisCluster(context.Background(), param.Config)
		if err != nil {
			return nil, err
		}
		param.ShutdownHandler.Register(func(ctx context.Context) error {
			return m.Close()
		})
		return m, nil
	}, dig.Name("redis-cluster")); err != nil {
		return fmt.Errorf("provide redis-cluster failed: %w", err)
	}

	// Prometheus
	if err := c.container.Provide(metrics.NewPrometheusMetric); err != nil {
		return fmt.Errorf("provide prometheus metrics failed: %w", err)
	}

	// TODO: Kafka
	return nil
}

// Repository & Usecase
func (c *Container) RegisterDomain() error {
	if err := c.container.Provide(repository.NewRedisClientRepository); err != nil {
		return err
	}

	if err := c.container.Provide(repository.NewRbacRepository); err != nil {
		return err
	}
	if err := c.container.Provide(usecase.NewRbacUsecase); err != nil {
		return err
	}

	if err := c.container.Provide(usecase.NewAuthUsecase); err != nil {
		return err
	}

	if err := c.container.Provide(repository.NewUserRepository); err != nil {
		return err
	}

	return nil
}

// 按 main.go 決定要加載內容的方法
func (c *Container) RegisterDeliveryHTTP() error {
	// Middleware & Server
	if err := c.container.Provide(http.NewMiddleware); err != nil {
		return err
	}
	if err := c.container.Provide(http.NewGinServer); err != nil {
		return err
	}

	// Handlers
	if err := c.container.Provide(handler.NewHealthHandler); err != nil {
		return err
	}
	if err := c.container.Provide(handler.NewMetricHandler); err != nil {
		return err
	}
	if err := c.container.Provide(handler.NewAuthHandler); err != nil {
		return err
	}
	if err := c.container.Provide(handler.NewRbacHandler); err != nil {
		return err
	}

	return nil
}
