package service

import (
	"context"
	"fmt"
	"time"

	"metrics-api/internal/models"
	"metrics-api/internal/prometheus"
	"metrics-api/pkg/logger"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

// QueriesService handles Prometheus query operations
type QueriesService struct {
	client     *prometheus.Client
	logger     logger.Logger
	maxPoints  int
}

// NewQueriesService creates a new queries service
func NewQueriesService(client *prometheus.Client, logger logger.Logger) *QueriesService {
	return &QueriesService{
		client:    client,
		logger:    logger,
		maxPoints: 11000, // Default max points limit
	}
}

// WithMaxPoints sets the maximum number of data points allowed in a query
func (s *QueriesService) WithMaxPoints(maxPoints int) *QueriesService {
	s.maxPoints = maxPoints
	return s
}

// ExecuteInstantQuery executes an instant query against Prometheus
func (s *QueriesService) ExecuteInstantQuery(ctx context.Context, queryParams models.InstantQueryParams) (*models.QueryResponse, error) {
	// Validate query
	if queryParams.Query == "" {
		return nil, models.ErrInvalidQuery
	}

	// Set default time to now if not provided
	queryTime := time.Now()
	if !queryParams.Time.IsZero() {
		queryTime = queryParams.Time
	}

	// Log query for debugging and audit
	s.logger.Infof("Executing instant query: %s at %s", queryParams.Query, queryTime)

	// Execute query
	results, err := s.client.Query(ctx, queryParams.Query, queryTime)
	if err != nil {
		s.logger.Errorf("Failed to execute query %s: %v", queryParams.Query, err)
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	// Convert to response model
	response := &models.QueryResponse{
		Query:     queryParams.Query,
		QueryTime: queryTime,
		Status:    "success",
		Data:      make([]models.DataPoint, 0, len(results)),
	}

	for _, result := range results {
		response.Data = append(response.Data, models.DataPoint{
			MetricName: result.MetricName,
			Labels:     result.Labels,
			Value:      result.Value,
			Timestamp:  result.Timestamp,
		})
	}

	return response, nil
}

// ExecuteRangeQuery executes a range query against Prometheus
func (s *QueriesService) ExecuteRangeQuery(ctx context.Context, params models.RangeQueryParams) (*models.RangeQueryResponse, error) {
	if params.Query == "" {
		return nil, models.ErrInvalidQuery
	}

	// Parse and validate time range
	end := time.Now()
	if !params.End.IsZero() {
		end = params.End
	}

	start := end.Add(-1 * time.Hour)
	if !params.Start.IsZero() {
		start = params.Start
	}

	if start.After(end) {
		return nil, models.ErrInvalidTimeRange
	}

	// Parse step duration
	step, err := time.ParseDuration(params.Step)
	if err != nil {
		return nil, fmt.Errorf("invalid step duration: %w", err)
	}

	// Calculate number of points
	duration := end.Sub(start)
	points := int(duration / step)
	if points > s.maxPoints {
		return nil, models.ErrTooManyDataPoints
	}

	// Create Prometheus range
	r := v1.Range{
		Start: start,
		End:   end,
		Step:  step,
	}

	// Execute query
	s.logger.Infof("Executing range query: %s from %s to %s with step %s", 
		params.Query, start.Format(time.RFC3339), end.Format(time.RFC3339), step)

	results, err := s.client.QueryRange(ctx, params.Query, r)
	if err != nil {
		s.logger.Errorf("Failed to execute range query: %v", err)
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	// Build response
	response := &models.RangeQueryResponse{
		Query:  params.Query,
		Start:  start,
		End:    end,
		Step:   step,
		Status: "success",
		Series: make([]models.TimeSeries, 0, len(results)),
	}

	for _, result := range results {
		series := models.TimeSeries{
			MetricName: result.MetricName,
			Labels:     result.Labels,
			DataPoints: make([]models.TimeValuePair, 0, len(result.Values)),
		}

		for _, pair := range result.Values {
			series.DataPoints = append(series.DataPoints, models.TimeValuePair{
                Timestamp: pair.Timestamp,
                Value:     pair.Value,
            })
		}

		response.Series = append(response.Series, series)
	}

	return response, nil
}

// ValidateQuery checks if a query is valid
func (s *QueriesService) ValidateQuery(ctx context.Context, query string) (*models.QueryValidation, error) {
	if query == "" {
		return &models.QueryValidation{
			Query:   query,
			Valid:   false,
			Message: "Query cannot be empty",
		}, nil
	}

	// Try to execute with minimal time range
	now := time.Now()
	_, err := s.client.Query(ctx, query, now)

	validation := &models.QueryValidation{
		Query: query,
		Valid: err == nil,
	}

	if err != nil {
		validation.Message = fmt.Sprintf("Query validation failed: %v", err)
	} else {
		validation.Message = "Query is valid"
	}

	return validation, nil
}

// GetQuerySuggestions attempts to provide helpful query suggestions
func (s *QueriesService) GetQuerySuggestions(ctx context.Context, prefix string, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	// Get all metrics as a starting point for suggestions
	metrics, err := s.client.GetMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics for suggestions: %w", err)
	}

	// Basic suggestions based on common patterns
	suggestions := []string{
		"rate(http_requests_total[5m])",
		"sum by (instance) (node_cpu_seconds_total)",
		"histogram_quantile(0.95, sum by(le) (rate(http_request_duration_seconds_bucket[5m])))",
		"up",
		"node_memory_MemFree_bytes / node_memory_MemTotal_bytes * 100",
	}

	// Add metrics that match the prefix
	for _, metric := range metrics {
		if len(suggestions) >= limit {
			break
		}
		
		// Basic prefix match
		if len(prefix) == 0 || (len(prefix) > 0 && startsWith(metric, prefix)) {
			suggestions = append(suggestions, metric)
		}
	}

	// Ensure we don't exceed the limit
	if len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}

	return suggestions, nil
}

// Helper function to check if a string starts with a prefix
func startsWith(s, prefix string) bool {
	if len(prefix) > len(s) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		if s[i] != prefix[i] {
			return false
		}
	}
	return true
}
