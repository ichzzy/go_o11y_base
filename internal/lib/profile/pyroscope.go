package profile

import (
	"context"
	"os"
	"runtime"

	"github.com/grafana/pyroscope-go"
	"github.com/ichzzy/go_o11y_base/internal/app/cx"
	"github.com/ichzzy/go_o11y_base/internal/config"
)

type Pyroscope struct {
	config   config.PyroscopeConfig
	appConf  config.AppConfig
	profiler *pyroscope.Profiler
}

func NewPyroscope(cfg config.PyroscopeConfig, appConf config.AppConfig) *Pyroscope {
	return &Pyroscope{
		config:  cfg,
		appConf: appConf,
	}
}

func (p *Pyroscope) Start(ctx context.Context) {
	if !p.config.Enabled {
		return
	}

	logger := cx.GetLogger(ctx)
	logger.Infof("pyroscope.Start: %s", p.config.Endpoint)

	profiler, err := pyroscope.Start(pyroscope.Config{
		ApplicationName: p.appConf.ServiceName,
		ServerAddress:   p.config.Endpoint,
		Logger:          nil,
		Tags:            map[string]string{"hostname": os.Getenv("HOSTNAME"), "env": p.appConf.Env},

		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,           // CPU 使用率
			pyroscope.ProfileAllocObjects,  // 累計記憶體分配對象數
			pyroscope.ProfileAllocSpace,    // 累計記憶體分配空間
			pyroscope.ProfileInuseObjects,  // 當前記憶體佔用對象數
			pyroscope.ProfileInuseSpace,    // 當前記憶體佔用空間
			pyroscope.ProfileGoroutines,    // Goroutine 堆疊追蹤
			pyroscope.ProfileMutexCount,    // Mutex 爭搶次數
			pyroscope.ProfileMutexDuration, // Mutex 爭搶耗時
			pyroscope.ProfileBlockCount,    // 阻塞事件次數
			pyroscope.ProfileBlockDuration, // 阻塞事件耗時
		},
	})

	if err != nil {
		logger.WithError(err).Error("pyroscope.Start failed")
		return
	}

	p.profiler = profiler

	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(5)
}

func (p *Pyroscope) Shutdown() error {
	if p.profiler != nil {
		return p.profiler.Stop()
	}
	return nil
}
