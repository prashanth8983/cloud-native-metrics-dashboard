package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"metrics-api/internal/models"
	"metrics-api/internal/service"
	"metrics-api/pkg/logger"

	"github.com/gorilla/mux"
)

// MetricsHandler handles metrics-related HTTP requests
type MetricsHandler struct {
	service *service.MetricsService // Changed to pointer
	logger  logger.Logger
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(service *service.MetricsService, logger logger.Logger) *MetricsHandler {
	return &MetricsHandler{
		service: service, // Store pointer directly instead of dereferencing
		logger:  logger,
	}
}

// RegisterRoutes registers the handler routes
func (h *MetricsHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/metrics", h.GetMetrics).Methods("GET")
	r.HandleFunc("/metrics/top", h.GetTopMetrics).Methods("GET")
	r.HandleFunc("/metrics/{name}", h.GetMetricSummary).Methods("GET")
	r.HandleFunc("/metrics/{name}/health", h.GetMetricHealth).Methods("GET")
}

// GetMetrics returns a list of available metrics
func (h *MetricsHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metrics, err := h.service.GetMetrics(ctx)
	if err != nil {
		h.logger.Errorf("Failed to get metrics: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to get metrics")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"metrics": metrics,
		"count":   len(metrics),
	})
}

// GetTopMetrics returns the top metrics by cardinality
func (h *MetricsHandler) GetTopMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse limit from query string
	limitStr := r.URL.Query().Get("limit")
	limit := 10 // Default

	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit <= 0 {
			RespondWithError(w, http.StatusBadRequest, "Invalid limit parameter")
			return
		}
		limit = parsedLimit
	}

	topMetrics, err := h.service.GetTopMetrics(ctx, limit)
	if err != nil {
		h.logger.Errorf("Failed to get top metrics: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to get top metrics")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"metrics": topMetrics,
		"count":   len(topMetrics),
	})
}

// GetMetricSummary returns a summary of a specific metric
func (h *MetricsHandler) GetMetricSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	metricName := vars["name"]

	if metricName == "" {
		RespondWithError(w, http.StatusBadRequest, "Metric name is required")
		return
	}

	summary, err := h.service.GetMetricSummary(ctx, metricName)
	if err != nil {
		if errors.Is(err, models.ErrMetricNotFound) {
			RespondWithError(w, http.StatusNotFound, "Metric not found")
			return
		}
		h.logger.Errorf("Failed to get metric summary for %s: %v", metricName, err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to get metric summary")
		return
	}

	RespondWithJSON(w, http.StatusOK, summary)
}

// GetMetricHealth returns health information about a specific metric
func (h *MetricsHandler) GetMetricHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	metricName := vars["name"]

	if metricName == "" {
		RespondWithError(w, http.StatusBadRequest, "Metric name is required")
		return
	}

	health, err := h.service.GetMetricHealth(ctx, metricName)
	if err != nil {
		h.logger.Errorf("Failed to get metric health for %s: %v", metricName, err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to get metric health")
		return
	}

	RespondWithJSON(w, http.StatusOK, health)
}
