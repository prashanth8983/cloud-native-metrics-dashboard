package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"metrics-api/internal/models"
	"metrics-api/internal/service"
	"metrics-api/pkg/logger"

	"github.com/gorilla/mux"
)

// QueriesHandler handles query-related HTTP requests
type QueriesHandler struct {
	service service.QueriesService
	logger  logger.Logger
}

// NewQueriesHandler creates a new queries handler
func NewQueriesHandler(service *service.QueriesService, logger logger.Logger) *QueriesHandler {
	return &QueriesHandler{
		service: *service,
		logger:  logger,
	}
}

// RegisterRoutes registers the handler routes
func (h *QueriesHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/query", h.InstantQuery).Methods("POST")
	r.HandleFunc("/query/range", h.RangeQuery).Methods("POST")
	r.HandleFunc("/query/validate", h.ValidateQuery).Methods("POST")
	r.HandleFunc("/query/suggestions", h.GetQuerySuggestions).Methods("GET")
}

// InstantQuery executes an instant query
func (h *QueriesHandler) InstantQuery(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var params models.InstantQueryParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if params.Query == "" {
		RespondWithError(w, http.StatusBadRequest, "Query cannot be empty")
		return
	}

	response, err := h.service.ExecuteInstantQuery(ctx, params)
	if err != nil {
		if errors.Is(err, models.ErrInvalidQuery) {
			RespondWithError(w, http.StatusBadRequest, "Invalid query")
			return
		}
		h.logger.Errorf("Failed to execute instant query: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to execute query")
		return
	}

	RespondWithJSON(w, http.StatusOK, response)
}

// RangeQuery executes a range query
func (h *QueriesHandler) RangeQuery(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var params models.RangeQueryParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if params.Query == "" {
		RespondWithError(w, http.StatusBadRequest, "Query cannot be empty")
		return
	}

	response, err := h.service.ExecuteRangeQuery(ctx, params)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrInvalidQuery):
			RespondWithError(w, http.StatusBadRequest, "Invalid query")
			return
		case errors.Is(err, models.ErrInvalidTimeRange):
			RespondWithError(w, http.StatusBadRequest, "Invalid time range")
			return
		case errors.Is(err, models.ErrTooManyDataPoints):
			RespondWithError(w, http.StatusBadRequest, "Query would return too many data points")
			return
		default:
			h.logger.Errorf("Failed to execute range query: %v", err)
			RespondWithError(w, http.StatusInternalServerError, "Failed to execute range query")
			return
		}
	}

	RespondWithJSON(w, http.StatusOK, response)
}

// ValidateQuery validates a query without executing it
func (h *QueriesHandler) ValidateQuery(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var payload struct {
		Query string `json:"query"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	validation, err := h.service.ValidateQuery(ctx, payload.Query)
	if err != nil {
		h.logger.Errorf("Failed to validate query: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to validate query")
		return
	}

	RespondWithJSON(w, http.StatusOK, validation)
}

// GetQuerySuggestions returns query suggestions based on a prefix
func (h *QueriesHandler) GetQuerySuggestions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	prefix := r.URL.Query().Get("prefix")
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

	suggestions, err := h.service.GetQuerySuggestions(ctx, prefix, limit)
	if err != nil {
		h.logger.Errorf("Failed to get query suggestions: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to get query suggestions")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"suggestions": suggestions,
		"count":       len(suggestions),
	})
}

// parseTime parses a time string from a query parameter
func parseTime(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Now(), nil
	}

	// Try parsing as unix timestamp
	if timestamp, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
		return time.Unix(timestamp, 0), nil
	}

	// Try parsing as RFC3339
	return time.Parse(time.RFC3339, timeStr)
}