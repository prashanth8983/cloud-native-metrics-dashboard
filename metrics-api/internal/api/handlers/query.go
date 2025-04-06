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

// QueryHandler handles instant queries to Prometheus
type QueryHandler struct {
	promClient *prometheus.Client
	cache      *cache.Cache
	log        *logger.Logger
}

// NewQueryHandler creates a new query handler
func NewQueryHandler(promClient *prometheus.Client, cache *cache.Cache, log *logger.Logger) *QueryHandler {
	return &QueryHandler{
		promClient: promClient,
		cache:      cache,
		log:        log.WithField("component", "query_handler"),
	}
}

// Handle handles an instant query request
func (h *QueryHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// Parse the request
	var request models.QueryRequest
	
	// Check if it's a GET or POST request
	if r.Method == http.MethodGet {
		// Parse query parameters
		query := r.URL.Query().Get("query")
		if query == "" {
			http.Error(w, "Query parameter is required", http.StatusBadRequest)
			return
		}
		
		request.Query = query
		
		// Parse time parameter
		timeStr := r.URL.Query().Get("time")
		if timeStr != "" {
			parsedTime, err := time.Parse(time.RFC3339, timeStr)
			if err != nil {
				http.Error(w, "Invalid time format (use RFC3339)", http.StatusBadRequest)
				return
			}
			request.Time = &parsedTime
		}
		
		// Parse debug parameter
		if r.URL.Query().Get("debug") == "true" {
			request.Debug = true
		}
		
		// Parse stats parameter
		if r.URL.Query().Get("stats") == "true" {
			request.Stats = true
		}
	} else {
		// Parse JSON body
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&request); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
	}
	
	// Set default query time if not provided
	queryTime := time.Now()
	if request.Time != nil {
		queryTime = *request.Time
	}
	
	// Create a cache key for this query
	cacheKey := "query_" + request.Query + "_" + queryTime.Format(time.RFC3339)
	
	// Try to get from cache first
	var result *prometheus.QueryResult
	var err error
	
	if cachedResult, found := h.cache.Get(cacheKey); found {
		h.log.Debug().Str("query", request.Query).Msg("cache hit for query")
		result = cachedResult.(*prometheus.QueryResult)
	} else {
		// Execute the query
		h.log.Debug().Str("query", request.Query).Time("time", queryTime).Msg("executing query")
		result, err = h.promClient.Query(r.Context(), request.Query, queryTime)
		if err != nil {
			h.log.Error().Err(err).Str("query", request.Query).Msg("failed to execute query")
			http.Error(w, "Failed to execute query: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		// Cache the result
		h.cache.Set(cacheKey, result)
	}
	
	// Process the result
	queryResult := &models.QueryResult{
		ResultType: result.ResultType,
		Warnings:   result.Warnings,
	}
	
	// Convert the result to the appropriate type
	switch result.ResultType {
	case "vector":
		vector, ok := result.Result.(models.Vector)
		if !ok {
			h.log.Error().Str("query", request.Query).Msg("failed to convert result to vector")
			http.Error(w, "Failed to process query result", http.StatusInternalServerError)
			return
		}
		queryResult.Vectors = models.ConvertVectorResult(vector)
	
	case "matrix":
		matrix, ok := result.Result.(models.Matrix)
		if !ok {
			h.log.Error().Str("query", request.Query).Msg("failed to convert result to matrix")
			http.Error(w, "Failed to process query result", http.StatusInternalServerError)
			return
		}
		queryResult.Matrix = models.ConvertMatrixResult(matrix)
	
	case "scalar":
		scalar, ok := result.Result.(models.Scalar)
		if !ok {
			h.log.Error().Str("query", request.Query).Msg("failed to convert result to scalar")
			http.Error(w, "Failed to process query result", http.StatusInternalServerError)
			return
		}
		queryResult.Scalar = &models.TimeValuePair{
			Time:  scalar.Timestamp.Time(),
			Value: float64(scalar.Value),
		}
	
	case "string":
		str, ok := result.Result.(models.String)
		if !ok {
			h.log.Error().Str("query", request.Query).Msg("failed to convert result to string")
			http.Error(w, "Failed to process query result", http.StatusInternalServerError)
			return
		}
		value := string(str.Value)
		queryResult.String = &value
	}
	
	// Create the response
	response := models.ToAPIResponse("success", queryResult)
	
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

// RangeQueryHandler handles range queries to Prometheus
type RangeQueryHandler struct {
	promClient *prometheus.Client
	cache      *cache.Cache
	log        *logger.Logger
}

// NewRangeQueryHandler creates a new range query handler
func NewRangeQueryHandler(promClient *prometheus.Client, cache *cache.Cache, log *logger.Logger) *RangeQueryHandler {
	return &RangeQueryHandler{
		promClient: promClient,
		cache:      cache,
		log:        log.WithField("component", "range_query_handler"),
	}
}

// Handle handles a range query request
func (h *RangeQueryHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// Parse the request
	var request models.RangeQueryRequest
	
	// Check if it's a GET or POST request
	if r.Method == http.MethodGet {
		// Parse query parameters
		query := r.URL.Query().Get("query")
		if query == "" {
			http.Error(w, "Query parameter is required", http.StatusBadRequest)
			return
		}
		request.Query = query
		
		// Parse start parameter
		startStr := r.URL.Query().Get("start")
		if startStr == "" {
			http.Error(w, "Start parameter is required", http.StatusBadRequest)
			return
		}
		start, err := time.Parse(time.RFC3339, startStr)
		if err != nil {
			http.Error(w, "Invalid start time format (use RFC3339)", http.StatusBadRequest)
			return
		}
		request.Start = &start
		
		// Parse end parameter
		endStr := r.URL.Query().Get("end")
		if endStr == "" {
			http.Error(w, "End parameter is required", http.StatusBadRequest)
			return
		}
		end, err := time.Parse(time.RFC3339, endStr)
		if err != nil {
			http.Error(w, "Invalid end time format (use RFC3339)", http.StatusBadRequest)
			return
		}
		request.End = &end
		
		// Parse step parameter
		step := r.URL.Query().Get("step")
		if step == "" {
			http.Error(w, "Step parameter is required", http.StatusBadRequest)
			return
		}
		request.Step = step
		
		// Parse debug parameter
		if r.URL.Query().Get("debug") == "true" {
			request.Debug = true
		}
		
		// Parse stats parameter
		if r.URL.Query().Get("stats") == "true" {
			request.Stats = true
		}
	} else {
		// Parse JSON body
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&request); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
	}
	
	// Validate request
	if request.Start == nil {
		http.Error(w, "Start time is required", http.StatusBadRequest)
		return
	}
	
	if request.End == nil {
		http.Error(w, "End time is required", http.StatusBadRequest)
		return
	}
	
	if request.Start.After(*request.End) {
		http.Error(w, "Start time must be before end time", http.StatusBadRequest)
		return
	}
	
	if request.Step == "" {
		http.Error(w, "Step is required", http.StatusBadRequest)
		return
	}
	
	// Parse step duration
	step, err := time.ParseDuration(request.Step)
	if err != nil {
		http.Error(w, "Invalid step format (use duration format like 15s, 1m, 1h)", http.StatusBadRequest)
		return
	}
	
	// Create a cache key for this query
	cacheKey := "range_query_" + request.Query + "_" + request.Start.Format(time.RFC3339) + "_" + request.End.Format(time.RFC3339) + "_" + request.Step
	
	// Try to get from cache first
	var result *prometheus.QueryResult
	
	if cachedResult, found := h.cache.Get(cacheKey); found {
		h.log.Debug().Str("query", request.Query).Msg("cache hit for range query")
		result = cachedResult.(*prometheus.QueryResult)
	} else {
		// Create options for the range query
		opts := prometheus.RangeQueryOptions{
			Start: *request.Start,
			End:   *request.End,
			Step:  step,
		}
		
		// Execute the query
		h.log.Debug().
			Str("query", request.Query).
			Time("start", opts.Start).
			Time("end", opts.End).
			Dur("step", opts.Step).
			Msg("executing range query")
		
		result, err = h.promClient.QueryRange(r.Context(), request.Query, opts)
		if err != nil {
			h.log.Error().Err(err).Str("query", request.Query).Msg("failed to execute range query")
			http.Error(w, "Failed to execute range query: "+err.Error(), http.StatusInternalServerError)
			return
		}
		
		// Cache the result
		h.cache.Set(cacheKey, result)
	}
	
	// Process the result
	queryResult := &models.QueryResult{
		ResultType: result.ResultType,
		Warnings:   result.Warnings,
	}
	
	// For range queries, the result should always be a matrix
	if result.ResultType != "matrix" {
		h.log.Error().Str("query", request.Query).Str("resultType", result.ResultType).Msg("unexpected result type for range query")
		http.Error(w, "Unexpected result type for range query", http.StatusInternalServerError)
		return
	}
	
	matrix, ok := result.Result.(models.Matrix)
	if !ok {
		h.log.Error().Str("query", request.Query).Msg("failed to convert result to matrix")
		http.Error(w, "Failed to process query result", http.StatusInternalServerError)
		return
	}
	
	queryResult.Matrix = models.ConvertMatrixResult(matrix)
	
	// Create the response
	response := models.ToAPIResponse("success", queryResult)
	
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