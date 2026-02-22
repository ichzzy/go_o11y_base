package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusMetric struct {
	requestDuration *prometheus.HistogramVec // HTTP 請求的響應時間
	requestCount    *prometheus.CounterVec   // HTTP 請求的累加器
	requestInFlight prometheus.Gauge         // 目前處理中的請求數
	panicCount      prometheus.Counter       // 系統崩潰次數
}

func NewPrometheusMetric() *PrometheusMetric {
	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_duration_seconds",
			Help:    "Duration of HTTP requests.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "code"},
	)

	requestCount := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "request_count_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "code"},
	)

	requestInFlight := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "request_in_flight",
			Help: "Current in-flight requests count.",
		},
	)

	panicCount := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "panic_total",
			Help: "Total number of panics occurred.",
		},
	)

	prometheus.MustRegister(requestDuration, requestCount, requestInFlight, panicCount)

	return &PrometheusMetric{
		requestDuration: requestDuration,
		requestCount:    requestCount,
		requestInFlight: requestInFlight,
		panicCount:      panicCount,
	}
}

func (pm *PrometheusMetric) RecordRequestDuration(method, path, code string, duration time.Duration) {
	pm.requestDuration.WithLabelValues(method, path, code).Observe(duration.Seconds())
}

func (pm *PrometheusMetric) RecordRequestCount(method, path, code string) {
	pm.requestCount.WithLabelValues(method, path, code).Inc()
}

func (pm *PrometheusMetric) IncInFlight() {
	pm.requestInFlight.Inc()
}

func (pm *PrometheusMetric) DecInFlight() {
	pm.requestInFlight.Dec()
}

func (pm *PrometheusMetric) RecordPanic() {
	pm.panicCount.Inc()
}
