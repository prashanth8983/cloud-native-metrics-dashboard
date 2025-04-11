package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"metrics-api/internal/models"
	"metrics-api/internal/service"
	"metrics-api/pkg/logger"

	"github.com/gorilla/mux"
)

// QueryHandler handles query-related HTTP requests
type QueryHandler struct {
	service service.QueriesService
	logger  logger.Logger
}

// NewQueryHandler creates a new query handler
func NewQueryHandler(service *service.QueriesService, logger logger.Logger) *QueryHandler {
	return &QueryHandler{
		service: *service,
		logger:  logger,
	}
}

// RegisterRoutes registers the handler routes
func (h *QueryHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/query/range", h.QueryRange).Methods(http.MethodPost)
}

type RangeQueryRequest struct {
	Query string `json:"query"`
	Start string `json:"start"`
	End   string `json:"end"`
	Step  string `json:"step"`
}

// QueryRange handles range queries
func (h *QueryHandler) QueryRange(w http.ResponseWriter, r *http.Request) {
	// Add content type check
	if ct := r.Header.Get("Content-Type"); ct != "application/json" {
		h.logger.Error("invalid content type", "content-type", ct)
		RespondWithError(w, http.StatusBadRequest, "Content-Type must be application/json")
		return
	}

	var req struct {
		Query string      `json:"query"`
		Start string      `json:"start"`
		End   string      `json:"end"`
		Step  interface{} `json:"step"` // Accept both string and number
	}
	
	// Read the body for logging in case of error
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("failed to read request body", "error", err)
		RespondWithError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	
	// Restore the body for later use
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	// Try to decode and log the raw body if it fails
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request body",
			"error", err,
			"body", string(body),
		)
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate required fields
	if req.Query == "" || req.Start == "" || req.End == "" {
		h.logger.Error("missing required fields",
			"query", req.Query,
			"start", req.Start,
			"end", req.End,
			"step", req.Step,
		)
		RespondWithError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	// Parse start time
	start, err := time.Parse(time.RFC3339, req.Start)
	if err != nil {
		h.logger.Error("invalid start time format", "error", err, "start", req.Start)
		RespondWithError(w, http.StatusBadRequest, "Invalid start time format")
		return
	}

	// Parse end time
	end, err := time.Parse(time.RFC3339, req.End)
	if err != nil {
		h.logger.Error("invalid end time format", "error", err, "end", req.End)
		RespondWithError(w, http.StatusBadRequest, "Invalid end time format")
		return
	}

	// Handle step parameter
	var stepStr string
	switch v := req.Step.(type) {
	case string:
		stepStr = v
		if !strings.HasSuffix(stepStr, "s") {
			stepStr += "s"
		}
	case float64:
		stepStr = fmt.Sprintf("%.0fs", v)
	case int:
		stepStr = fmt.Sprintf("%ds", v)
	default:
		h.logger.Error("invalid step format", "step", req.Step)
		RespondWithError(w, http.StatusBadRequest, "Invalid step format")
		return
	}

	// Remove 's' suffix for parsing
	step := strings.TrimSuffix(stepStr, "s")
	stepInt, err := strconv.Atoi(step)
	if err != nil {
		h.logger.Error("invalid step format", "error", err, "step", stepStr)
		RespondWithError(w, http.StatusBadRequest, "Invalid step format")
		return
	}

	// Create query params
	params := models.RangeQueryParams{
		Query: req.Query,
		Start: start,
		End:   end,
		Step:  fmt.Sprintf("%ds", stepInt),
	}

	// Execute the query
	response, err := h.service.ExecuteRangeQuery(r.Context(), params)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrInvalidQuery):
			RespondWithError(w, http.StatusBadRequest, "Invalid query")
		case errors.Is(err, models.ErrInvalidTimeRange):
			RespondWithError(w, http.StatusBadRequest, "Invalid time range")
		case errors.Is(err, models.ErrTooManyDataPoints):
			RespondWithError(w, http.StatusBadRequest, "Query would return too many data points")
		default:
			h.logger.Error("failed to execute range query", "error", err)
			RespondWithError(w, http.StatusInternalServerError, "Failed to execute query")
		}
		return
	}

	RespondWithJSON(w, http.StatusOK, response)
}

// Helper function to read request body
func readBody(r *http.Request) string {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return ""
	}
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore the body
	return string(bodyBytes)
}

