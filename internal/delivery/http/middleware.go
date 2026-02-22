package http

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ichzzy/go_o11y_base/internal/app/cx"
	"github.com/ichzzy/go_o11y_base/internal/delivery/http/response"
	"github.com/ichzzy/go_o11y_base/internal/domain"
	"github.com/ichzzy/go_o11y_base/internal/domain/entity"
	"github.com/ichzzy/go_o11y_base/internal/domain/repository"
	domain_uc "github.com/ichzzy/go_o11y_base/internal/domain/usecase"
	"github.com/ichzzy/go_o11y_base/internal/lib/logs"
	"github.com/ichzzy/go_o11y_base/internal/lib/metrics"
	"github.com/rs/xid"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/dig"
)

type Middleware struct {
	appLogger        logs.AppLogger
	prometheusMetric *metrics.PrometheusMetric
	rbacUsecase      domain_uc.RbacUsecase
	authUsecase      domain_uc.AuthUsecase
	redisRepo        repository.RedisClientRepository
}

type MiddlewareParam struct {
	dig.In

	AppLogger        logs.AppLogger
	PrometheusMetric *metrics.PrometheusMetric
	RbacUsecase      domain_uc.RbacUsecase
	AuthUsecase      domain_uc.AuthUsecase
	RedisRepo        repository.RedisClientRepository
}

func NewMiddleware(param MiddlewareParam) *Middleware {
	return &Middleware{
		appLogger:        param.AppLogger,
		prometheusMetric: param.PrometheusMetric,
		rbacUsecase:      param.RbacUsecase,
		authUsecase:      param.AuthUsecase,
		redisRepo:        param.RedisRepo,
	}
}

func (m *Middleware) Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		m.prometheusMetric.IncInFlight()
		defer m.prometheusMetric.DecInFlight()

		c.Next()

		latencyDuration := time.Since(start)
		latencyMs := latencyDuration.Milliseconds()

		// 1. 處理 Metrics
		code := domain.CodeSuccess.String()
		if val, exists := c.Get("biz_code"); exists {
			code = val.(string)
		}

		m.prometheusMetric.RecordRequestCount(c.Request.Method, c.FullPath(), code)
		m.prometheusMetric.RecordRequestDuration(c.Request.Method, c.FullPath(), code, latencyDuration)

		// 2. 處理 Access Log (僅記錄 slow request)
		if latencyMs > 500 {
			cx.GetLogger(c.Request.Context()).WithFields(logs.Fields{
				"ip":       c.ClientIP(),
				"method":   c.Request.Method,
				"path":     c.FullPath(),
				"url":      c.Request.URL.String(),
				"status":   c.Writer.Status(),
				"biz_code": code,
				"latency":  fmt.Sprintf("%d ms", latencyMs),
				"agent":    c.Request.UserAgent(),
			}).Warn("slow request detected")
		}
	}
}

// TODO: 有上下游鏈路後 再擴展成沿著 OTEL->Baggage->Fallback 的傳接方式
func (m *Middleware) LoggerAndTrace() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		spanCtx := trace.SpanFromContext(ctx).SpanContext()
		traceID := ""
		if spanCtx.IsValid() {
			traceID = spanCtx.TraceID().String()
		} else {
			traceID = xid.New().String()
		}

		ctx = cx.SetLogger(ctx, m.appLogger.WithField("trace_id", traceID))
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func (m *Middleware) Recover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if exception := recover(); exception != nil {
				m.prometheusMetric.RecordPanic()

				logger := cx.GetLogger(c.Request.Context())
				logger.WithFields(logs.Fields{
					"error": exception,
					"stack": string(debug.Stack()),
				}).Error("panic recovered")

				// Rethrow error
				_ = c.Error(domain.ErrInternal.ReWrap(errors.New("panic")))
			}
		}()

		c.Next()
	}
}

func (m *Middleware) ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		logger := cx.GetLogger(c.Request.Context())
		if err := c.Errors.Last(); err != nil {
			logger = logger.WithFields(logs.Fields{
				"method":  c.Request.Method,
				"url":     c.Request.RequestURI,
				"request": c.Request.Body,
			})

			if appErr, ok := err.Err.(domain.AppError); ok {
				c.Set("biz_code", appErr.Code().String())

				logger.WithFields(logs.Fields{
					"app_code":    appErr.Code(),
					"stack_trace": appErr.StackTrace(),
					"http_status": appErr.Status(),
				}).WithError(appErr).Error(appErr.Message())

				// 前端目前規範只收 200 OK
				c.AbortWithStatusJSON(http.StatusOK, response.Error(appErr))
			} else {
				c.Set("biz_code", domain.ErrInternal.Code().String())

				logger.WithError(err).Errorf("unrecognized error occurred: %s", err.Error())

				// 前端目前規範只收 200 OK
				c.AbortWithStatusJSON(http.StatusOK, response.Error(domain.ErrInternal))
			}
		}
	}
}

func (m *Middleware) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse JWT
		bearerHeader := c.GetHeader("Authorization")
		if bearerHeader == "" {
			_ = c.Error(domain.ErrUnauthorized.New())
			c.Abort()
			return
		}
		parts := strings.SplitN(bearerHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			_ = c.Error(domain.ErrUnauthorized.New())
			c.Abort()
			return
		}
		claims, err := m.authUsecase.ParseToken(parts[1])
		if err != nil {
			_ = c.Error(domain.ErrUnauthorized.ReWrap(err))
			c.Abort()
			return
		}

		// Validate permission
		ok, err := m.rbacUsecase.Enforce(
			strconv.FormatUint(claims.RoleID, 10),
			c.Request.URL.Path,
			c.Request.Method,
		)
		if err != nil {
			_ = c.Error(domain.ErrInternal.ReWrap(err))
			c.Abort()
			return
		}
		if !ok { // 403
			_ = c.Error(domain.ErrForbidden.New())
			c.Abort()
			return
		}

		userID, _ := strconv.ParseUint(claims.Subject, 10, 64)
		ctx := cx.SetUserIdentity(c.Request.Context(), entity.UserIdentity{
			UserID: userID,
			RoleID: claims.RoleID,
		})
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
