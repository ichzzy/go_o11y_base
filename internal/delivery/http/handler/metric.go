package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/dig"
)

type MetricHandlerParam struct {
	dig.In
}

type MetricHandler struct {
	handler http.Handler
}

func NewMetricHandler(param MetricHandlerParam) *MetricHandler {
	return &MetricHandler{
		handler: promhttp.HandlerFor(
			prometheus.DefaultGatherer,
			promhttp.HandlerOpts{
				Registry: prometheus.DefaultRegisterer,
			},
		),
	}
}

func (h *MetricHandler) Metrics(c *gin.Context) {
	h.handler.ServeHTTP(c.Writer, c.Request)
}
