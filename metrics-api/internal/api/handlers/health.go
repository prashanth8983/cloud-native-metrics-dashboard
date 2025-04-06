package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"metrics-api/internal/cache"
	"metrics-api/internal/models"
	"metrics-api/internal/prometheus"
	"metrics-api/pkg/logger"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	promClient *prometheus.Client
	cache      *cache.Cache
	log        *logger.Logger
	startTime  time.Time
}

// NewHealthHandler creates a new health check handler
func NewHealthHandler(promClient *prometheus.Client, cache *cache.Cache, log *logger.Logger) *HealthHandler {
	return &HealthHandler{
		promClient: promClient,
		cache:      cache,
		log:        log.WithField("component", "health_handler"),
		startTime:  time.Now(),
	}
}

// Handle handles a health check request
func (h *HealthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// Create a cache key
	cacheKey := "health"
	
	// Try to get from cache first
	var status models.HealthStatus
	
	if cachedResult, found := h.cache.Get(cacheKey); found {
		h.log.Debug().Msg("cache hit for health check")
		status = cachedResult.(models.HealthStatus)
	} else {
		// Check if Prometheus is healthy
		h.log.Debug().Msg("checking Prometheus health")
		
		healthy, err := h.promClient.IsHealthy(r.Context())
		
		if err != nil {
			h.log.Error().Err(err).Msg("Prometheus health check failed")
			status = models.HealthStatus{
				Status: "down",
			}
		} else if !healthy {
			h.log.Warn().Msg("Prometheus is unhealthy")
			status = models.HealthStatus{
				Status: "down",
			}
		} else {
			// Get build info from Prometheus
			buildInfo, err := h.promClient.BuildInfo(r.Context())
			if err != nil {
				h.log.Error().Err(err).Msg("failed to get Prometheus build info")
				status = models.HealthStatus{
					Status: "up",
				}
			} else {
				// Calculate uptime
				uptime := prometheus.FormatDuration(time.Since(h.startTime))
				
				// Create status
				status = models.HealthStatus{
					Status:  "up",
					Version: buildInfo.Version,
					Uptime:  uptime,
				}
			}
		}
		
		// Cache the result (short TTL)
		h.cache.SetWithTTL(cacheKey, status, 10*time.Second)
	}
	
	// Create the response
	response := models.ToAPIResponse("success", status)
	
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