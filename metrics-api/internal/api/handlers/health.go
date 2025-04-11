package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"metrics-api/internal/models"
	"metrics-api/internal/prometheus"
	"metrics-api/pkg/logger"

	"github.com/gorilla/mux"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	promClient *prometheus.Client
	logger     logger.Logger
	startTime  time.Time
	version    string
}

// NewHealthHandler creates a new health check handler
func NewHealthHandler(promClient *prometheus.Client, logger logger.Logger, version string) *HealthHandler {
	return &HealthHandler{
		promClient: promClient,
		logger:     logger,
		startTime:  time.Now(),
		version:    version,
	}
}

// RegisterRoutes registers the handler routes
func (h *HealthHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/health", h.GetHealth).Methods("GET")
	r.HandleFunc("/health/detailed", h.GetDetailedHealth).Methods("GET")
	r.HandleFunc("/ready", h.GetReadiness).Methods("GET")
	r.HandleFunc("/live", h.GetLiveness).Methods("GET")
}

// GetHealth returns the basic health status of the service
func (h *HealthHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	status := "up"
	
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status":  status,
		"version": h.version,
		"time":    time.Now().Format(time.RFC3339),
	})
}

// GetDetailedHealth returns detailed health information including all checks
func (h *HealthHandler) GetDetailedHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Create a timeout context for health checks
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	checks := make(map[string]string)
	details := make(map[string]any)
	overallStatus := "up"
	
	// Check Prometheus connection
	promStatus, promDetails := h.checkPrometheusHealth(timeoutCtx)
	checks["prometheus"] = promStatus
	details["prometheus"] = promDetails
	
	if promStatus != "up" {
		overallStatus = "degraded"
	}
	
	// Calculate uptime
	uptime := time.Since(h.startTime)
	uptimeStr := formatDuration(uptime)
	
	// Create the health status
	healthStatus := models.HealthStatus{
		Status:    overallStatus,
		Version:   h.version,
		Uptime:    uptimeStr,
		Timestamp: time.Now(),
		Checks:    checks,
		Details:   details,
	}
	
	RespondWithJSON(w, http.StatusOK, healthStatus)
}

// GetReadiness checks if the service is ready to receive traffic
func (h *HealthHandler) GetReadiness(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Create a short timeout context for readiness checks
	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	
	// Check if Prometheus is reachable
	promStatus, _ := h.checkPrometheusHealth(timeoutCtx)
	
	if promStatus != "up" {
		h.logger.Warn("Service is not ready: Prometheus is down")
		http.Error(w, "Service is not ready", http.StatusServiceUnavailable)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Service is ready"))
}

// GetLiveness checks if the service is alive
func (h *HealthHandler) GetLiveness(w http.ResponseWriter, r *http.Request) {
	// For liveness, we just need to return 200 OK
	// This function is intentionally simple and fast
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Service is alive"))
}

// checkPrometheusHealth checks if Prometheus is healthy
func (h *HealthHandler) checkPrometheusHealth(ctx context.Context) (string, map[string]interface{}) {
	details := make(map[string]interface{})
	startTime := time.Now()
	
	// Try to execute a simple query
	results, err := h.promClient.Query(ctx, "up", time.Now())
	
	responseTime := time.Since(startTime)
	details["response_time_ms"] = responseTime.Milliseconds()
	
	if err != nil {
		details["error"] = err.Error()
		h.logger.Error("prometheus health check failed", "error", err)
		return "down", details
	}
	
	// Verify we got a response
	if len(results) == 0 {
		details["error"] = "no results returned"
		return "degraded", details
	}
	
	details["error"] = nil
	return "up", details
}

// formatDuration converts a duration to a human-readable string
func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
