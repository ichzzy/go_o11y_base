package http

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/ichzzy/go_o11y_base/internal/config"
	"github.com/ichzzy/go_o11y_base/internal/delivery/http/handler"
	"github.com/ichzzy/go_o11y_base/internal/domain"
	"github.com/ichzzy/go_o11y_base/internal/lib/shutdown"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/dig"
)

type GinServerParam struct {
	dig.In

	Config          *config.Config `name:"config"`
	Middleware      *Middleware
	ShutdownHandler *shutdown.Handler

	HealthHandler *handler.HealthHandler
	MetricHandler *handler.MetricHandler
	AuthHandler   *handler.AuthHandler
	RbacHandler   *handler.RbacHandler
}

type handlers struct {
	health *handler.HealthHandler
	metric *handler.MetricHandler
	auth   *handler.AuthHandler
	rbac   *handler.RbacHandler
}

func (gs *GinServer) setHandlers(param GinServerParam) {
	gs.handlers = handlers{
		health: param.HealthHandler,
		metric: param.MetricHandler,
		auth:   param.AuthHandler,
		rbac:   param.RbacHandler,
	}
}

func (gs *GinServer) indexRoute(engine *gin.Engine) error {
	mw := gs.middleware
	root := engine.Group("") // ingress path prefix

	// Health Probes
	root.GET("/readyz", gs.handlers.health.Ready)
	root.GET("/livez", gs.handlers.health.Live)
	// Metrics
	root.GET("/metrics", gs.handlers.metric.Metrics)

	api := root.Group("/")
	api.Use(
		gzip.Gzip(gzip.DefaultCompression),
		otelgin.Middleware(domain.ServiceName),
		mw.LoggerAndTrace(),
		mw.Metrics(),
		mw.ErrorHandler(),
		mw.Recover(),
	)

	// 公開接口
	noAuthGateway := api.Group("/")
	{
		// 快速登入
		if gs.mode == "local" || gs.mode == "dev" {
			noAuthGateway.POST("/v1/auth/login-dev", gs.handlers.auth.DevLogin)
		}

		// 刷新 Token
		noAuthGateway.POST("/v1/auth/refresh", gs.handlers.auth.RefreshToken)
		noAuthGateway.POST("/v1/auth/login", gs.handlers.auth.Login)
	}

	// 授權用戶接口
	authGateway := api.Group("/")
	authGateway.Use(mw.Auth())
	{
		// 取得角色清單
		authGateway.GET("/v1/roles", gs.handlers.rbac.ListRoles)
		// 建立角色(包含權限)
		authGateway.POST("/v1/roles", gs.handlers.rbac.CreateRole)
		// 更新角色(包含權限)
		authGateway.PUT("/v1/roles/:id", gs.handlers.rbac.UpdateRole)
		// 取得角色的權限清單
		authGateway.GET("/v1/roles/:roleID/permissions", gs.handlers.rbac.ListRolePermissions)
		// 取得權限清單
		authGateway.GET("/v1/permissions", gs.handlers.rbac.ListPermissions)
		// 取得當前用戶可視菜單
		authGateway.GET("/v1/me/visible-menus", gs.handlers.rbac.ListMyVisibleMenus)

		// 建立用戶
		authGateway.POST("/v1/users", gs.handlers.auth.CreateUser)
	}

	return nil
}
