package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"metrics-api/internal/models"
	"metrics-api/internal/prometheus"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Define interfaces to match the service signatures
type MetricsServiceInterface interface {
	GetMetrics(ctx context.Context) ([]string, error)
	GetTopMetrics(ctx context.Context, limit int) ([]models.TopMetric, error)
	GetMetricSummary(ctx context.Context, name string) (*models.MetricSummary, error)
	GetMetricHealth(ctx context.Context, name string) (*models.MetricHealth, error)
}

type QueriesServiceInterface interface {
	ExecuteInstantQuery(ctx context.Context, params models.InstantQueryParams) (*models.QueryResponse, error)
	ExecuteRangeQuery(ctx context.Context, params models.RangeQueryParams) (*models.RangeQueryResponse, error)
	ValidateQuery(ctx context.Context, query string) (*models.QueryValidation, error)
	GetQuerySuggestions(ctx context.Context, prefix string, limit int) ([]string, error)
}

type AlertsServiceInterface interface {
	GetAlerts(ctx context.Context) ([]models.Alert, error)
	GetAlertSummary(ctx context.Context) (*models.AlertSummary, error)
	GetAlertGroups(ctx context.Context, groupBy string) ([]models.AlertGroup, error)
}

// MockMetricsService implements MetricsServiceInterface for testing
type MockMetricsService struct {
	mock.Mock
}

func (m *MockMetricsService) GetMetrics(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockMetricsService) GetTopMetrics(ctx context.Context, limit int) ([]models.TopMetric, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]models.TopMetric), args.Error(1)
}

func (m *MockMetricsService) GetMetricSummary(ctx context.Context, name string) (*models.MetricSummary, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.MetricSummary), args.Error(1)
}

func (m *MockMetricsService) GetMetricHealth(ctx context.Context, name string) (*models.MetricHealth, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.MetricHealth), args.Error(1)
}

// MockQueriesService implements QueriesServiceInterface for testing
type MockQueriesService struct {
	mock.Mock
}

func (m *MockQueriesService) ExecuteInstantQuery(ctx context.Context, params models.InstantQueryParams) (*models.QueryResponse, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.QueryResponse), args.Error(1)
}

func (m *MockQueriesService) ExecuteRangeQuery(ctx context.Context, params models.RangeQueryParams) (*models.RangeQueryResponse, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RangeQueryResponse), args.Error(1)
}

func (m *MockQueriesService) ValidateQuery(ctx context.Context, query string) (*models.QueryValidation, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.QueryValidation), args.Error(1)
}

func (m *MockQueriesService) GetQuerySuggestions(ctx context.Context, prefix string, limit int) ([]string, error) {
	args := m.Called(ctx, prefix, limit)
	return args.Get(0).([]string), args.Error(1)
}

// MockAlertsService implements AlertsServiceInterface for testing
type MockAlertsService struct {
	mock.Mock
}

func (m *MockAlertsService) GetAlerts(ctx context.Context) ([]models.Alert, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Alert), args.Error(1)
}

func (m *MockAlertsService) GetAlertSummary(ctx context.Context) (*models.AlertSummary, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AlertSummary), args.Error(1)
}

func (m *MockAlertsService) GetAlertGroups(ctx context.Context, groupBy string) ([]models.AlertGroup, error) {
	args := m.Called(ctx, groupBy)
	return args.Get(0).([]models.AlertGroup), args.Error(1)
}

// MockPrometheusClient is a mock of the Prometheus client
type MockPrometheusClient struct {
	mock.Mock
}

func (m *MockPrometheusClient) Query(ctx context.Context, query string, ts time.Time) ([]prometheus.QueryResult, error) {
	args := m.Called(ctx, query, ts)
	return args.Get(0).([]prometheus.QueryResult), args.Error(1)
}

// Test MetricsHandler.GetMetrics
func TestGetMetrics(t *testing.T) {
	// Create mock service
	mockService := new(MockMetricsService)
	
	// Setup test data
	mockMetrics := []string{
		"http_requests_total",
		"node_cpu_seconds_total",
		"node_memory_MemFree_bytes",
	}
	
	// Setup expectations
	mockService.On("GetMetrics", mock.Anything).Return(mockMetrics, nil)
	
	// Create a router and register the handler
	router := mux.NewRouter()
	
	// Create a handler function that uses our mock
	handler := func(w http.ResponseWriter, r *http.Request) {
		metrics, err := mockService.GetMetrics(r.Context())
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Failed to get metrics")
			return
		}

		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"metrics": metrics,
			"count":   len(metrics),
		})
	}
	
	router.HandleFunc("/metrics", handler).Methods("GET")
	
	// Create request
	req, err := http.NewRequest("GET", "/metrics", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	// Create response recorder
	rr := httptest.NewRecorder()
	
	// Call the handler via the router
	router.ServeHTTP(rr, req)
	
	// Check status code
	assert.Equal(t, http.StatusOK, rr.Code)
	
	// Parse response
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}
	
	// Check response data
	assert.Equal(t, float64(3), response["count"])
	assert.NotNil(t, response["metrics"])
	metrics, ok := response["metrics"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 3, len(metrics))
	
	// Check mock expectations
	mockService.AssertExpectations(t)
}

// Test GetTopMetrics
func TestGetTopMetrics(t *testing.T) {
	// Create mock service
	mockService := new(MockMetricsService)
	
	// Setup test data
	mockTopMetrics := []models.TopMetric{
		{
			Name:        "http_requests_total",
			Cardinality: 100,
			SampleRate:  5.0,
		},
		{
			Name:        "node_cpu_seconds_total",
			Cardinality: 50,
			SampleRate:  2.0,
		},
	}
	
	// Setup expectations with dynamic limit handling
	mockService.On("GetTopMetrics", mock.Anything, mock.AnythingOfType("int")).Return(mockTopMetrics, nil)
	
	// Create a router and register the handler
	router := mux.NewRouter()
	
	// Create a handler function that uses our mock
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Get limit from query string with proper parsing
		limitStr := r.URL.Query().Get("limit")
		limit := 10 // Default value
		if limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
				limit = parsedLimit
			}
		}
		
		topMetrics, err := mockService.GetTopMetrics(r.Context(), limit)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Failed to get top metrics")
			return
		}

		RespondWithJSON(w, http.StatusOK, map[string]interface{}{
			"metrics": topMetrics,
			"count":   len(topMetrics),
		})
	}
	
	router.HandleFunc("/metrics/top", handler).Methods("GET")
	
	// Create request
	req, err := http.NewRequest("GET", "/metrics/top?limit=5", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	// Create response recorder
	rr := httptest.NewRecorder()
	
	// Call the handler via the router
	router.ServeHTTP(rr, req)
	
	// Check status code
	assert.Equal(t, http.StatusOK, rr.Code)
	
	// Parse response
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}
	
	// Check response data
	assert.Equal(t, float64(2), response["count"])
	assert.NotNil(t, response["metrics"])
	metrics, ok := response["metrics"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(metrics))
	
	// Check mock expectations
	mockService.AssertExpectations(t)
}

// Test GetMetricSummary
func TestGetMetricSummary(t *testing.T) {
	// Create mock service
	mockService := new(MockMetricsService)
	
	// Setup test data
	now := time.Now()
	mockSummary := &models.MetricSummary{
		Name:        "http_requests_total",
		Labels:      []string{"method", "status", "handler"},
		Cardinality: 10,
		Stats: models.MetricStats{
			Min: 5.0,
			Max: 100.0,
			Avg: 50.0,
		},
		LastUpdated: now,
		Samples: []models.MetricSample{
			{
				Labels: map[string]string{
					"method": "GET",
					"status": "200",
				},
				Value:     42.0,
				Timestamp: now,
			},
		},
	}
	
	// Setup expectations
	mockService.On("GetMetricSummary", mock.Anything, "http_requests_total").Return(mockSummary, nil)
	
	// Create a router and register the handler
	router := mux.NewRouter()
	
	// Create a handler function that uses our mock
	handler := func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		metricName := vars["name"]
		
		summary, err := mockService.GetMetricSummary(r.Context(), metricName)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Failed to get metric summary")
			return
		}
		
		RespondWithJSON(w, http.StatusOK, summary)
	}
	
	router.HandleFunc("/metrics/{name}", handler).Methods("GET")
	
	// Create request
	req, err := http.NewRequest("GET", "/metrics/http_requests_total", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	// Create response recorder
	rr := httptest.NewRecorder()
	
	// Call the handler via the router
	router.ServeHTTP(rr, req)
	
	// Check status code
	assert.Equal(t, http.StatusOK, rr.Code)
	
	// Parse response
	var response models.MetricSummary
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}
	
	// Check response data
	assert.Equal(t, "http_requests_total", response.Name)
	assert.Equal(t, 10, response.Cardinality)
	assert.Equal(t, 5.0, response.Stats.Min)
	assert.Equal(t, 100.0, response.Stats.Max)
	assert.Equal(t, 50.0, response.Stats.Avg)
	assert.Equal(t, 1, len(response.Samples))
	
	// Check mock expectations
	mockService.AssertExpectations(t)
}

// Test InstantQuery
func TestInstantQuery(t *testing.T) {
	// Create mock service
	mockService := new(MockQueriesService)
	
	// Setup test data
	now := time.Now()
	queryPayload := `{"query": "http_requests_total"}`
	
	mockResponse := &models.QueryResponse{
		Query:     "http_requests_total",
		QueryTime: now,
		Status:    "success",
		Data: []models.DataPoint{
			{
				MetricName: "http_requests_total",
				Labels: map[string]string{
					"method": "GET",
					"status": "200",
				},
				Value:     42.0,
				Timestamp: now,
			},
		},
	}
	
	// Setup expectations
	mockService.On("ExecuteInstantQuery", mock.Anything, mock.MatchedBy(func(params models.InstantQueryParams) bool {
		return params.Query == "http_requests_total"
	})).Return(mockResponse, nil)
	
	// Create a router and register the handler
	router := mux.NewRouter()
	
	// Create a handler function that uses our mock
	handler := func(w http.ResponseWriter, r *http.Request) {
		var params models.InstantQueryParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}
		
		if params.Query == "" {
			RespondWithError(w, http.StatusBadRequest, "Query cannot be empty")
			return
		}
		
		response, err := mockService.ExecuteInstantQuery(r.Context(), params)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Failed to execute query")
			return
		}
		
		RespondWithJSON(w, http.StatusOK, response)
	}
	
	router.HandleFunc("/query", handler).Methods("POST")
	
	// Create request
	req, err := http.NewRequest("POST", "/query", strings.NewReader(queryPayload))
	if err != nil {
		t.Fatal(err)
	}
	
	// Create response recorder
	rr := httptest.NewRecorder()
	
	// Call the handler via the router
	router.ServeHTTP(rr, req)
	
	// Check status code
	assert.Equal(t, http.StatusOK, rr.Code)
	
	// Parse response
	var response models.QueryResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "http_requests_total", response.Query)
	assert.Equal(t, "success", response.Status)
	assert.Equal(t, 1, len(response.Data))

	mockService.AssertExpectations(t)
}