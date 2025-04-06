package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"

	"metrics-api/internal/cache"
	"metrics-api/internal/models"
	"metrics-api/internal/prometheus"
	"metrics-api/pkg/logger"
)

// AlertsHandler handles requests for alerts
type AlertsHandler struct {
	promClient *prometheus.Client
	cache      *cache.Cache
	log        *logger.Logger
}

// NewAlertsHandler creates a new alerts handler
func NewAlertsHandler(promClient *prometheus.Client, cache *cache.Cache, log *logger.Logger) *AlertsHandler {
	return &AlertsHandler{
		promClient: promClient,
		cache:      cache,
		log:        log.WithField("component", "alerts_handler"),
	}
}

// Handle handles a request for alerts
func (h *AlertsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// Parse the request
	var request models.AlertsRequest
	
	// Check if it's a GET or POST request
	if r.Method == http.MethodGet {
		// Parse query parameters
		filter := r.URL.Query().Get("filter")
		request.Filter = filter
		
		// Parse silenced parameter
		silencedStr := r.URL.Query().Get("silenced")
		if silencedStr != "" {
			silenced := silencedStr == "true"
			request.Silenced = &silenced
		}
		
		// Parse inhibited parameter
		inhibitedStr := r.URL.Query().Get("inhibited")
		if inhibitedStr != "" {
			inhibited := inhibitedStr == "true"
			request.Inhibited = &inhibited
		}
		
		// Parse active parameter
		activeStr := r.URL.Query().Get("active")
		if activeStr != "" {
			active := activeStr == "true"
			request.Active = &active
		}
		
		// Parse severity parameter
		severityStr := r.URL.Query().Get("severity")
		if severityStr != "" {
			request.Severity = strings.Split(severityStr, ",")
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
	cacheKey := "alerts"
	if request.Filter != "" {
		cacheKey += "_filter_" + request.Filter
	}
	if request.Silenced != nil {
		if *request.Silenced {
			cacheKey += "_silenced"
		} else {
			cacheKey += "_not_silenced"
		}
	}
	if request.Inhibited != nil {
		if *request.Inhibited {
			cacheKey += "_inhibited"
		} else {
			cacheKey += "_not_inhibited"
		}
	}
	if request.Active != nil {
		if *request.Active {
			cacheKey += "_active"
		} else {
			cacheKey += "_not_active"
		}
	}
	if len(request.Severity) > 0 {
		cacheKey += "_severity_" + strings.Join(request.Severity, "_")
	}
	
	// Try to get from cache first
	var alerts []models.Alert
	
	if cachedResult, found := h.cache.Get(cacheKey); found {
		h.log.Debug().Msg("cache hit for alerts")
		alerts = cachedResult.([]models.Alert)
	} else {
		// Get alerts from Prometheus
		h.log.Debug().Msg("fetching alerts from Prometheus")
		
		alertsResult, err := h.promClient.Alerts(r.Context())
		if err != nil {
			h.log.Error().Err(err).Msg("failed to fetch alerts")
			http.Error(w, "Failed to fetch alerts: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		// Convert alerts to our model
		alerts = convertAlerts(alertsResult)
		
		// Filter alerts
		alerts = filterAlerts(alerts, request)
		
		// Cache the result (short TTL for alerts since they change frequently)
		h.cache.SetWithTTL(cacheKey, alerts, 30*time.Second)
	}
	
	// Create the response
	response := models.ToAPIResponse("success", alerts)
	
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

// convertAlerts converts Prometheus alerts to our model
func convertAlerts(alertsResult v1.AlertsResult) []models.Alert {
	alerts := make([]models.Alert, 0, len(alertsResult.Alerts))
	
	for _, alert := range alertsResult.Alerts {
		// Convert labels
		labels := make(map[string]string, len(alert.Labels))
		for name, value := range alert.Labels {
			labels[string(name)] = string(value)
		}
		
		// Convert annotations
		annotations := make(map[string]string, len(alert.Annotations))
		for name, value := range alert.Annotations {
			annotations[string(name)] = string(value)
		}
		
		// Extract value if available
		value := 0.0
		if val, ok := alert.Labels["value"]; ok {
			if v, err := stringToFloat(string(val)); err == nil {
				value = v
			}
		}
		
		alerts = append(alerts, models.Alert{
			Labels:      labels,
			Annotations: annotations,
			State:       alert.State,
			ActiveAt:    alert.ActiveAt,
			Value:       value,
			Silenced:    len(alert.SilencedBy) > 0,
			Inhibited:   len(alert.InhibitedBy) > 0,
		})
	}
	
	return alerts
}

// filterAlerts filters alerts based on the request
func filterAlerts(alerts []models.Alert, request models.AlertsRequest) []models.Alert {
	filtered := make([]models.Alert, 0, len(alerts))
	
	for _, alert := range alerts {
		// Apply filter
		if request.Filter != "" {
			match := false
			
			// Check labels
			for _, value := range alert.Labels {
				if strings.Contains(strings.ToLower(value), strings.ToLower(request.Filter)) {
					match = true
					break
				}
			}
			
			// Check annotations
			if !match {
				for _, value := range alert.Annotations {
					if strings.Contains(strings.ToLower(value), strings.ToLower(request.Filter)) {
						match = true
						break
					}
				}
			}
			
			if !match {
				continue
			}
		}
		
		// Filter by silenced
		if request.Silenced != nil && alert.Silenced != *request.Silenced {
			continue
		}
		
		// Filter by inhibited
		if request.Inhibited != nil && alert.Inhibited != *request.Inhibited {
			continue
		}
		
		// Filter by active
		if request.Active != nil && (alert.State == "pending" || alert.State == "firing") != *request.Active {
			continue
		}
		
		// Filter by severity
		if len(request.Severity) > 0 {
			severity, ok := alert.Labels["severity"]
			if !ok {
				continue
			}
			
			found := false
			for _, s := range request.Severity {
				if strings.EqualFold(severity, s) {
					found = true
					break
				}
			}
			
			if !found {
				continue
			}
		}
		
		filtered = append(filtered, alert)
	}
	
	return filtered
}

// stringToFloat converts a string to a float64
func stringToFloat(s string) (float64, error) {
	var f float64
	err := json.Unmarshal([]byte(s), &f)
	return f, err
}