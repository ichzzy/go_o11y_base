package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ichzzy/go_o11y_base/internal/delivery/http/response"
	"github.com/ichzzy/go_o11y_base/internal/domain"
	"github.com/redis/go-redis/v9"
	"go.uber.org/dig"
	"gorm.io/gorm"
)

type HealthHandlerParam struct {
	dig.In

	RedisCluster *redis.ClusterClient `name:"redis-cluster"`
	Mysql        *gorm.DB             `name:"mysql"`
}

type HealthHandler struct {
	redisCluster *redis.ClusterClient
	mysql        *gorm.DB
}

func NewHealthHandler(param HealthHandlerParam) *HealthHandler {
	return &HealthHandler{
		redisCluster: param.RedisCluster,
		mysql:        param.Mysql,
	}
}

func (handler *HealthHandler) Ready(c *gin.Context) {
	// Ping MySQL
	db, err := handler.mysql.DB()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, response.Error(domain.ErrInternal.ReWrap(err)))
		return
	}
	if err := db.Ping(); err != nil {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, response.Error(domain.ErrInternal.ReWrap(err)))
		return
	}

	// Ping Redis
	if err := handler.redisCluster.Ping(c.Request.Context()).Err(); err != nil {
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, response.Error(domain.ErrInternal.ReWrap(err)))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"mysql": "ok",
		"redis": "ok",
	}))
}

func (handler *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, response.Success(gin.H{
		"status": "up",
	}))
}
