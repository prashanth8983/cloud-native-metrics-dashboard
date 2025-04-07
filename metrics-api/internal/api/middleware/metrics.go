package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsMiddleware collects metrics for HTTP requests
type MetricsMiddleware struct {
	requestCounter     *prometheus.CounterVec
	requestDuration    *prometheus.HistogramVec
	responseSize       *prometheus.HistogramVec
	currentConnections *prometheus.GaugeVec
	registry           *prometheus.Registry
}

// NewMetricsMiddleware creates a new metrics middleware with default metrics
func NewMetricsMiddleware() *MetricsMiddleware {
	registry := prometheus.NewRegistry()
	
	// Create metrics
	requestCounter := promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests by method, path, and status code",
		},
		[]string{"method", "path", "status"},
	)
	
	requestDuration := promauto.With(registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds by method and path",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path"},
	)
	
	responseSize := promauto.With(registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes by method and path",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000},
		},
		[]string{"method", "path"},
	)
	
	currentConnections := promauto.With(registry).NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_connections_current",
			Help: "Current number of HTTP connections by method and path",
		},
		[]string{"method", "path"},
	)
	
	// Register metrics with the default registry for exposition
	prometheus.DefaultRegisterer.MustRegister(requestCounter)
	prometheus.DefaultRegisterer.MustRegister(requestDuration)
	prometheus.DefaultRegisterer.MustRegister(responseSize)
	prometheus.DefaultRegisterer.MustRegister(currentConnections)
	
	return &MetricsMiddleware{
		requestCounter:     requestCounter,
		requestDuration:    requestDuration,
		responseSize:       responseSize,
		currentConnections: currentConnections,
		registry:           registry,
	}
}

// Middleware returns the HTTP middleware for collecting metrics
func (m *MetricsMiddleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract path template from route
			path := getNormalizedPath(r)
			
			// Track current connections
			m.currentConnections.WithLabelValues(r.Method, path).Inc()
			defer m.currentConnections.WithLabelValues(r.Method, path).Dec()
			
			// Start timer
			start := time.Now()
			
			// Create a wrapped response writer to capture status code and response size
			wrw := NewWrapResponseWriter(w)
			
			// Process request
			next.ServeHTTP(wrw, r)
			
			// Calculate duration
			duration := time.Since(start).Seconds()
			
			// Record metrics
			m.requestCounter.WithLabelValues(
				r.Method,
				path,
				strconv.Itoa(wrw.Status()),
			).Inc()
			
			m.requestDuration.WithLabelValues(
				r.Method,
				path,
			).Observe(duration)
			
			m.responseSize.WithLabelValues(
				r.Method,
				path,
			).Observe(float64(wrw.BytesWritten()))
		})
	}
}

// getNormalizedPath extracts the route pattern from the current route
// or returns the actual path if no route match is found
func getNormalizedPath(r *http.Request) string {
	// Try to get the route from the gorilla/mux context
	if route := mux.CurrentRoute(r); route != nil {
		if path, err := route.GetPathTemplate(); err == nil {
			return path
		}
	}
	
	// Fallback to the actual path
	return r.URL.Path
}

// Registry returns the Prometheus registry used by the middleware
func (m *MetricsMiddleware) Registry() *prometheus.Registry {
	return m.registry
}

// InstrumentHandlerDuration is a helper function to measure the duration of individual handlers
func InstrumentHandlerDuration(handlerName string) func(http.Handler) http.Handler {
	durationMetric := promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "handler_duration_seconds",
			Help:    "Duration of handlers in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
		},
		[]string{"handler"},
	)
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			duration := time.Since(start).Seconds()
			durationMetric.WithLabelValues(handlerName).Observe(duration)
		})
	}
}

// CustomMetricMiddleware allows adding custom metrics specific to your application
type CustomMetricMiddleware struct {
	*MetricsMiddleware
	errorCounter *prometheus.CounterVec
	queryCount   *prometheus.CounterVec
}

// NewCustomMetricMiddleware creates a metrics middleware with additional custom metrics
func NewCustomMetricMiddleware() *CustomMetricMiddleware {
	base := NewMetricsMiddleware()
	
	// Add custom metrics
	errorCounter := promauto.With(base.registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_error_total",
			Help: "Total number of API errors by type",
		},
		[]string{"type"},
	)
	
	queryCount := promauto.With(base.registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "prometheus_query_total",
			Help: "Total number of Prometheus queries by type",
		},
		[]string{"type"},
	)
	
	// Register with default registry
	prometheus.DefaultRegisterer.MustRegister(errorCounter)
	prometheus.DefaultRegisterer.MustRegister(queryCount)
	
	return &CustomMetricMiddleware{
		MetricsMiddleware: base,
		errorCounter:      errorCounter,
		queryCount:        queryCount,
	}
}

// IncrementErrorCount increments the error counter for a specific error type
func (m *CustomMetricMiddleware) IncrementErrorCount(errorType string) {
	m.errorCounter.WithLabelValues(errorType).Inc()
}

// IncrementQueryCount increments the query counter for a specific query type
func (m *CustomMetricMiddleware) IncrementQueryCount(queryType string) {
	m.queryCount.WithLabelValues(queryType).Inc()
}

// RecordAPILatency records the latency of an API operation
func RecordAPILatency(name string, duration time.Duration) {
	// Define a metric for API operations
	static := promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_operation_duration_seconds",
			Help:    "Duration of API operations in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5},
		},
		[]string{"operation"},
	)
	
	static.WithLabelValues(name).Observe(duration.Seconds())
}

// CreatePrometheusHandler creates a handler for exposing Prometheus metrics
func CreatePrometheusHandler() http.Handler {
	return promhttp.Handler()
}