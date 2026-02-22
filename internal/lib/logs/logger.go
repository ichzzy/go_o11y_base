package logs

import (
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/ichzzy/go_o11y_base/internal/config"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

const (
	TimestampFormat string = "2006-01-02T15:04:05.000Z07:00"
)

type FieldKey = string

const (
	FieldKeyCaller FieldKey = "caller"
)

type Fields = logrus.Fields

type AppLogger interface {
	WithError(err error) AppLogger
	WithField(key string, value any) AppLogger
	WithFields(fields Fields) AppLogger
	Debug(args ...any)
	Debugf(format string, args ...any)
	Info(args ...any)
	Infof(format string, args ...any)
	Warn(args ...any)
	Warnf(format string, args ...any)
	Error(args ...any)
	Errorf(format string, args ...any)
	Fatal(args ...any)
	Fatalf(format string, args ...any)
	Panic(args ...any)
	Panicf(format string, args ...any)
}

type appLogger struct {
	entry *logrus.Entry
}

func NewAppLogger(config *config.Config) (AppLogger, error) {
	lvl, err := logrus.ParseLevel(config.App.Log.Level)
	if err != nil {
		return nil, fmt.Errorf("logrus.ParseLevel failed: %w", err)
	}

	var formatter logrus.Formatter = &logrus.JSONFormatter{
		TimestampFormat: TimestampFormat,
	}
	if config.App.Log.Pretty {
		formatter = &logrus.TextFormatter{
			FullTimestamp:   true,
			ForceColors:     true,
			TimestampFormat: TimestampFormat,
		}
	}

	// 因這裏的 filebeat fs 是撈本地 log, 所以要輸出 log file
	writer, err := rotatelogs.New(
		config.App.Log.Path+".%Y-%m-%d-%H-%M",
		rotatelogs.WithLinkName(config.App.Log.Path),
		rotatelogs.WithMaxAge(config.App.Log.MaxAge),
		rotatelogs.WithRotationTime(config.App.Log.RotationTime),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize rotatelogs: %w", err)
	}
	output := io.MultiWriter(os.Stderr, writer)

	// 配置全域 Logrus, 防呆機制
	logrus.SetLevel(lvl)
	logrus.SetFormatter(formatter)
	logrus.SetOutput(output)
	logrus.SetReportCaller(true) // 開啟

	// 配置 App 專用 Logger
	logger := logrus.New()
	logger.SetLevel(lvl)
	logger.SetFormatter(formatter)
	logger.SetOutput(output)
	logger.SetReportCaller(false) // 關閉，改由 reportCaller() 處理

	return &appLogger{
		entry: logrus.NewEntry(logger),
	}, nil
}

// 調用前確認 Logrus 已由 NewAppLogger 配置完畢
func NewAppLoggerWithoutConfig() AppLogger {
	return &appLogger{
		entry: logrus.NewEntry(logrus.StandardLogger()),
	}
}

func (al *appLogger) WithField(key string, value any) AppLogger {
	return &appLogger{entry: al.entry.WithField(key, value)}
}
func (al *appLogger) WithFields(fields Fields) AppLogger {
	return &appLogger{entry: al.entry.WithFields(fields)}
}
func (al *appLogger) WithError(err error) AppLogger {
	return &appLogger{entry: al.entry.WithError(err)}
}

func (al *appLogger) Debug(args ...any) { al.reportCaller().entry.Debug(args...) }
func (al *appLogger) Debugf(format string, args ...any) {
	al.reportCaller().entry.Debugf(format, args...)
}
func (al *appLogger) Info(args ...any) { al.reportCaller().entry.Info(args...) }
func (al *appLogger) Infof(format string, args ...any) {
	al.reportCaller().entry.Infof(format, args...)
}
func (al *appLogger) Warn(args ...any) { al.reportCaller().entry.Warn(args...) }
func (al *appLogger) Warnf(format string, args ...any) {
	al.reportCaller().entry.Warnf(format, args...)
}
func (al *appLogger) Error(args ...any) { al.reportCaller().entry.Error(args...) }
func (al *appLogger) Errorf(format string, args ...any) {
	al.reportCaller().entry.Errorf(format, args...)
}
func (al *appLogger) Fatal(args ...any) { al.reportCaller().entry.Fatal(args...) }
func (al *appLogger) Fatalf(format string, args ...any) {
	al.reportCaller().entry.Fatalf(format, args...)
}
func (al *appLogger) Panic(args ...any) { al.reportCaller().entry.Panic(args...) }
func (al *appLogger) Panicf(format string, args ...any) {
	al.reportCaller().entry.Panicf(format, args...)
}
func (al *appLogger) reportCaller() *appLogger {
	_, file, line, _ := runtime.Caller(2)
	return &appLogger{
		entry: al.entry.WithField(FieldKeyCaller, fmt.Sprintf("%s:%d", file, line)),
	}
}
