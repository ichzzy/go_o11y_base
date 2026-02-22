package mysql

import (
	"context"
	"time"

	"github.com/ichzzy/go_o11y_base/internal/app/cx"
	"github.com/ichzzy/go_o11y_base/internal/lib/logs"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 實作 GORM logger.Interface，將 SQL 日誌輸出到 logrus
type GormLogger struct {
	fallbackLogger       logs.AppLogger
	slowThreshold        time.Duration
	ignoreRecordNotFound bool
	logLevel             logger.LogLevel
}

func NewGormLogger(fallbackLogger logs.AppLogger, slowThreshold time.Duration) logger.Interface {
	return &GormLogger{
		fallbackLogger:       fallbackLogger,
		slowThreshold:        slowThreshold,
		ignoreRecordNotFound: true,
		logLevel:             logger.Warn, // 只記錄警告級別以上 (Slow SQL, Error)
	}
}

func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.logLevel = level
	return &newLogger
}

// getLogger 優先從 context 拿帶有 trace_id 的 logger，否則使用 fallback
func (l *GormLogger) getLogger(ctx context.Context) logs.AppLogger {
	if ctx != nil {
		if ctxLogger := cx.GetLogger(ctx); ctxLogger != nil {
			return ctxLogger
		}
	}
	return l.fallbackLogger
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Info {
		l.getLogger(ctx).Infof(msg, data...)
	}
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Warn {
		l.getLogger(ctx).Warnf(msg, data...)
	}
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Error {
		l.getLogger(ctx).Errorf(msg, data...)
	}
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.logLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	fields := logs.Fields{
		"elapsed_ms": elapsed.Milliseconds(),
		"rows":       rows,
		"sql":        sql,
	}

	switch {
	case err != nil && l.logLevel >= logger.Error && (!l.ignoreRecordNotFound || err != gorm.ErrRecordNotFound):
		l.getLogger(ctx).WithFields(fields).WithError(err).Error("gorm query error")
	case elapsed > l.slowThreshold && l.slowThreshold != 0 && l.logLevel >= logger.Warn:
		l.getLogger(ctx).WithFields(fields).Warnf("slow sql (>= %v)", l.slowThreshold)
	case l.logLevel >= logger.Info:
		l.getLogger(ctx).WithFields(fields).Info("gorm query")
	}
}
