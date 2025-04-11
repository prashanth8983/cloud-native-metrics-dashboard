package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsMiddleware holds the Prometheus registry and metrics
type MetricsMiddleware struct {
	registry        *prometheus.Registry
	requestCounter  *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

// NewMetricsMiddleware creates a new metrics middleware
func NewMetricsMiddleware() *MetricsMiddleware {
	registry := prometheus.NewRegistry()
	
	requestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
	
	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2.5, 5},
		},
		[]string{"method", "path"},
	)
	
	registry.MustRegister(requestCounter, requestDuration)
	
	return &MetricsMiddleware{
		registry:        registry,
		requestCounter:  requestCounter,
		requestDuration: requestDuration,
	}
}

// MetricsHandler returns a handler for the /metrics endpoint
func (m *MetricsMiddleware) MetricsHandler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

// Middleware wraps an http.Handler with metrics collection
func (m *MetricsMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrw := NewWrapResponseWriter(w)
		
		next.ServeHTTP(wrw, r)
		
		// Record metrics after the request is processed
		duration := time.Since(start).Seconds()
		m.requestCounter.WithLabelValues(r.Method, r.URL.Path, fmt.Sprint(wrw.Status())).Inc()
		m.requestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
	})
}
