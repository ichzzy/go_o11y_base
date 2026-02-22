package cx

import (
	"context"

	"github.com/ichzzy/go_o11y_base/internal/domain/entity"
	"github.com/ichzzy/go_o11y_base/internal/lib/logs"
)

type ContextKey string

const (
	ContextKeyLogger       ContextKey = "app_logger"
	ContextKeyUserIdentity ContextKey = "user_identity"
)

func SetLogger(ctx context.Context, logger logs.AppLogger) context.Context {
	return context.WithValue(ctx, ContextKeyLogger, logger)
}
func GetLogger(ctx context.Context) logs.AppLogger {
	logger, ok := ctx.Value(ContextKeyLogger).(logs.AppLogger)
	if !ok {
		return logs.NewAppLoggerWithoutConfig()
	}
	return logger
}

func SetUserIdentity(ctx context.Context, identity entity.UserIdentity) context.Context {
	return context.WithValue(ctx, ContextKeyUserIdentity, identity)
}
func GetUserIdentity(ctx context.Context) entity.UserIdentity {
	identity, ok := ctx.Value(ContextKeyUserIdentity).(entity.UserIdentity)
	if !ok {
		identity = entity.UserIdentity{}
	}
	return identity
}
