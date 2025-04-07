package handlers

import (
	"metrics-api/internal/service"
	"metrics-api/pkg/logger"
	"net/http"

	"github.com/gorilla/mux"
)

// AlertsHandler handles alert-related HTTP requests
type AlertsHandler struct {
	service service.AlertsService
	logger  logger.Logger
}

// NewAlertsHandler creates a new alerts handler
func NewAlertsHandler(service *service.AlertsService, logger logger.Logger) *AlertsHandler {
	return &AlertsHandler{
		service: *service,
		logger:  logger,
	}
}

// RegisterRoutes registers the handler routes
func (h *AlertsHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/alerts", h.GetAlerts).Methods("GET")
	r.HandleFunc("/alerts/summary", h.GetAlertSummary).Methods("GET")
	r.HandleFunc("/alerts/groups", h.GetAlertGroups).Methods("GET")
}

// GetAlerts returns all current alerts
func (h *AlertsHandler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	alerts, err := h.service.GetAlerts(ctx)
	if err != nil {
		h.logger.Errorf("Failed to get alerts: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to get alerts")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

// GetAlertSummary returns a summary of current alert status
func (h *AlertsHandler) GetAlertSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	summary, err := h.service.GetAlertSummary(ctx)
	if err != nil {
		h.logger.Errorf("Failed to get alert summary: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to get alert summary")
		return
	}

	RespondWithJSON(w, http.StatusOK, summary)
}

// GetAlertGroups returns alerts grouped by a specified label
func (h *AlertsHandler) GetAlertGroups(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	groupBy := r.URL.Query().Get("by")
	if groupBy == "" {
		groupBy = "severity" // Default grouping
	}

	groups, err := h.service.GetAlertGroups(ctx, groupBy)
	if err != nil {
		h.logger.Errorf("Failed to get alert groups: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to get alert groups")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"groups": groups,
		"count":  len(groups),
		"by":     groupBy,
	})
}