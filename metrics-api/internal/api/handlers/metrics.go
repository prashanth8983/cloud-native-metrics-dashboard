package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"metrics-api/internal/cache"
	"metrics-api/internal/models"
	"metrics-api/internal/prometheus"
	"metrics-api/pkg/logger"
)

// MetricsSummaryHandler handles requests for metrics summaries
type MetricsSummaryHandler struct {
	promClient *prometheus.Client
	cache      *cache.Cache
	log        *logger.Logger
}

// NewMetricsSummaryHandler creates a new metrics summary handler
func NewMetricsSummaryHandler(promClient *prometheus.Client, cache *cache.Cache, log *logger.Logger) *MetricsSummaryHandler {
	return &MetricsSummaryHandler{
		promClient: promClient,
		cache:      cache,
		log:        log.WithField("component", "metrics_summary_handler"),
	}
}

// Handle handles a metrics summary request
func (h *MetricsSummaryHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// Parse the request
	var request models.MetricsSummaryRequest
	
	// Check if it's a GET or POST request
	if r.Method == http.MethodGet {
		// Parse metrics parameter
		metricsStr := r.URL.Query().Get("metrics")
		if metricsStr != "" {
			request.Metrics = strings.Split(metricsStr, ",")
		}
	} else {
		// Parse JSON body
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&request); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
	}
	
	// Create a cache key
	cacheKey := "metrics_summary"
	if len(request.Metrics) > 0 {
		cacheKey += "_" + strings.Join(request.Metrics, "_")
	}
	
	// Try to get from cache first
	var summary models.MetricsSummary
	
	if cachedResult, found := h.cache.Get(cacheKey); found {
		h.log.Debug().Msg("cache hit for metrics summary")
		summary = cachedResult.(models.MetricsSummary)
	} else {
		// Get metrics summary
		h.log.Debug().Msg("fetching metrics summary")
		
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		
		// Create a new summary with current timestamp
		summary = models.MetricsSummary{
			Timestamp: time.Now(),
		}
		
		// Fetch CPU usage
		cpuQuery := "avg(rate(container_cpu_usage_seconds_total{namespace!=\"kube-system\"}[5m])) * 100"
		cpuResult, err := h.promClient.Query(ctx, cpuQuery, time.Now())
		if err == nil {
			if value, err := prometheus.ExtractVectorValue(cpuResult); err == nil {
				summary.CPUUsage = value
			}
		}
		
		// Fetch memory usage (in GB)
		memoryQuery := "sum(container_memory_usage_bytes{namespace!=\"kube-system\"}) / (1024 * 1024 * 1024)"
		memoryResult, err := h.promClient.Query(ctx, memoryQuery, time.Now())
		if err == nil {
			if value, err := prometheus.ExtractVectorValue(memoryResult); err == nil {
				summary.MemoryUsage = value
			}
		}
		
		// Fetch pod count
		podCountQuery := "count(kube_pod_info)"
		podCountResult, err := h.promClient.Query(ctx, podCountQuery, time.Now())
		if err == nil {
			if value, err := prometheus.ExtractVectorValue(podCountResult); err == nil {
				summary.PodCount = int(value)
			}
		}
		
		// Fetch active pods
		activePodQuery := "count(kube_pod_status_phase{phase=\"Running\"})"
		activePodResult, err := h.promClient.Query(ctx, activePodQuery, time.Now())
		if err == nil {
			if value, err := prometheus.ExtractVectorValue(activePodResult); err == nil {
				summary.ActivePods = int(value)
			}
		}
		
		// Fetch error rate
		errorRateQuery := "sum(rate(app_errors_total[5m])) / sum(rate(app_request_total[5m])) * 100"
		errorRateResult, err := h.promClient.Query(ctx, errorRateQuery, time.Now())
		if err == nil {
			if value, err := prometheus.ExtractVectorValue(errorRateResult); err == nil {
				summary.ErrorRate = value
			}
		}
		
		// Fetch response time (in ms)
		responseTimeQuery := "histogram_quantile(0.95, sum(rate(app_request_duration_seconds_bucket[5m])) by (le)) * 1000"
		responseTimeResult, err := h.promClient.Query(ctx, responseTimeQuery, time.Now())
		if err == nil {
			if value, err := prometheus.ExtractVectorValue(responseTimeResult); err == nil {
				summary.ResponseTime = value
			}
		}
		
		// Cache the result
		h.cache.Set(cacheKey, summary)
	}
	
	// Create the response
	response := models.ToAPIResponse("success", summary)
	
	// Set content type
	w.Header().Set("Content-Type", "application/json")
	
	// Write the response
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(response); err != nil {
		h.log.Error().Err(err).Msg("failed to encode response")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}