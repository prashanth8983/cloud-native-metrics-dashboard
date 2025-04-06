// internal/api/handlers/metrics_summary_test.go
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"metrics-api/internal/cache"
	"metrics-api/internal/models"
	"metrics-api/internal/prometheus"
	"metrics-api/pkg/logger"
)

// MockPrometheusClientForMetrics is a mock for the Prometheus client
type MockPrometheusClientForMetrics struct {
	mock.Mock
}

func (m *MockPrometheusClientForMetrics) Query(ctx context.Context, query string, time time.Time) (*prometheus.QueryResult, error) {
	args := m.Called(ctx, query, time)
	return args.Get(0).(*prometheus.QueryResult), args.Error(1)
}

func TestMetricsSummaryHandler_Handle(t *testing.T) {
	// Create a logger
	log := logger.New(logger.Config{
		Level:  logger.DebugLevel,
		Format: logger.JSONFormat,
	})

	// Create a cache
	cacheInstance := cache.New(1*time.Minute, 100, 5*time.Minute)

	// Create test cases
	testCases := []struct {
		name           string
		setupMock      func(mock *MockPrometheusClientForMetrics)
		expectedStatus int
		validateBody   func(t *testing.T, body []byte)
	}{
		{
			name: "successful metrics summary",
			setupMock: func(mockClient *MockPrometheusClientForMetrics) {
				// Mock CPU usage query
				mockClient.On("Query", 
					mock.Anything, 
					"avg(rate(container_cpu_usage_seconds_total{namespace!=\"kube-system\"}[5m])) * 100", 
					mock.Anything,
				).Return(createMockQueryResult("vector", 65.5), nil)
				
				// Mock memory usage query
				mockClient.On("Query", 
					mock.Anything, 
					"sum(container_memory_usage_bytes{namespace!=\"kube-system\"}) / (1024 * 1024 * 1024)", 
					mock.Anything,
				).Return(createMockQueryResult("vector", 12.3), nil)
				
				// Mock pod count query
				mockClient.On("Query", 
					mock.Anything, 
					"count(kube_pod_info)", 
					mock.Anything,
				).Return(createMockQueryResult("vector", 42), nil)
				
				// Mock active pods query
				mockClient.On("Query", 
					mock.Anything, 
					"count(kube_pod_status_phase{phase=\"Running\"})", 
					mock.Anything,
				).Return(createMockQueryResult("vector", 38), nil)
				
				// Mock error rate query
				mockClient.On("Query", 
					mock.Anything, 
					"sum(rate(app_errors_total[5m])) / sum(rate(app_request_total[5m])) * 100", 
					mock.Anything,
				).Return(createMockQueryResult("vector", 0.52), nil)
				
				// Mock response time query
				mockClient.On("Query", 
					mock.Anything, 
					"histogram_quantile(0.95, sum(rate(app_request_duration_seconds_bucket[5m])) by (le)) * 1000", 
					mock.Anything,
				).Return(createMockQueryResult("vector", 187.5), nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var response models.APIResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				
				assert.Equal(t, "success", response.Status)
				
				// Extract metrics summary from the response
				summary, ok := response.Data.(map[string]interface{})
				require.True(t, ok)
				
				// Validate metrics values
				assert.Equal(t, 65.5, summary["cpuUsage"])
				assert.Equal(t, 12.3, summary["memoryUsage"])
				assert.Equal(t, float64(42), summary["podCount"])
				assert.Equal(t, float64(38), summary["activePods"])
				assert.Equal(t, 0.52, summary["errorRate"])
				assert.Equal(t, 187.5, summary["responseTime"])
				assert.Contains(t, summary, "timestamp")
			},
		},
		{
			name: "partial metrics success",
			setupMock: func(mockClient *MockPrometheusClientForMetrics) {
				// Mock CPU usage query - success
				mockClient.On("Query", 
					mock.Anything, 
					"avg(rate(container_cpu_usage_seconds_total{namespace!=\"kube-system\"}[5m])) * 100", 
					mock.Anything,
				).Return(createMockQueryResult("vector", 65.5), nil)
				
				// Mock memory usage query - success
				mockClient.On("Query", 
					mock.Anything, 
					"sum(container_memory_usage_bytes{namespace!=\"kube-system\"}) / (1024 * 1024 * 1024)", 
					mock.Anything,
				).Return(createMockQueryResult("vector", 12.3), nil)
				
				// Mock pod count query - error
				mockClient.On("Query", 
					mock.Anything, 
					"count(kube_pod_info)", 
					mock.Anything,
				).Return(&prometheus.QueryResult{}, errors.New("query failed"))
				
				// Mock active pods query - error
				mockClient.On("Query", 
					mock.Anything, 
					"count(kube_pod_status_phase{phase=\"Running\"})", 
					mock.Anything,
				).Return(&prometheus.QueryResult{}, errors.New("query failed"))
				
				// Mock error rate query - success
				mockClient.On("Query", 
					mock.Anything, 
					"sum(rate(app_errors_total[5m])) / sum(rate(app_request_total[5m])) * 100", 
					mock.Anything,
				).Return(createMockQueryResult("vector", 0.52), nil)
				
				// Mock response time query - success
				mockClient.On("Query", 
					mock.Anything, 
					"histogram_quantile(0.95, sum(rate(app_request_duration_seconds_bucket[5m])) by (le)) * 1000", 
					mock.Anything,
				).Return(createMockQueryResult("vector", 187.5), nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var response models.APIResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				
				assert.Equal(t, "success", response.Status)
				
				// Extract metrics summary from the response
				summary, ok := response.Data.(map[string]interface{})
				require.True(t, ok)
				
				// Validate available metrics
				assert.Equal(t, 65.5, summary["cpuUsage"])
				assert.Equal(t, 12.3, summary["memoryUsage"])
				assert.Equal(t, 0.52, summary["errorRate"])
				assert.Equal(t, 187.5, summary["responseTime"])
				
				// Pod metrics should be zero due to query errors
				assert.Equal(t, float64(0), summary["podCount"])
				assert.Equal(t, float64(0), summary["activePods"])
				
				assert.Contains(t, summary, "timestamp")
			},
		},
		{
			name: "empty vector results",
			setupMock: func(mockClient *MockPrometheusClientForMetrics) {
				// Create empty result (no vector samples)
				emptyResult := &prometheus.QueryResult{
					ResultType: "vector",
					Result:     models.Vector{},
				}
				
				// Mock all queries to return empty results
				mockClient.On("Query", 
					mock.Anything, 
					"avg(rate(container_cpu_usage_seconds_total{namespace!=\"kube-system\"}[5m])) * 100", 
					mock.Anything,
				).Return(emptyResult, nil)
				
				mockClient.On("Query", 
					mock.Anything, 
					"sum(container_memory_usage_bytes{namespace!=\"kube-system\"}) / (1024 * 1024 * 1024)", 
					mock.Anything,
				).Return(emptyResult, nil)
				
				mockClient.On("Query", 
					mock.Anything, 
					"count(kube_pod_info)", 
					mock.Anything,
				).Return(emptyResult, nil)
				
				mockClient.On("Query", 
					mock.Anything, 
					"count(kube_pod_status_phase{phase=\"Running\"})", 
					mock.Anything,
				).Return(emptyResult, nil)
				
				mockClient.On("Query", 
					mock.Anything, 
					"sum(rate(app_errors_total[5m])) / sum(rate(app_request_total[5m])) * 100", 
					mock.Anything,
				).Return(emptyResult, nil)
				
				mockClient.On("Query", 
					mock.Anything, 
					"histogram_quantile(0.95, sum(rate(app_request_duration_seconds_bucket[5m])) by (le)) * 1000", 
					mock.Anything,
				).Return(emptyResult, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var response models.APIResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				
				assert.Equal(t, "success", response.Status)
				
				// Extract metrics summary from the response
				summary, ok := response.Data.(map[string]interface{})
				require.True(t, ok)
				
				// All metrics should be zero
				assert.Equal(t, float64(0), summary["cpuUsage"])
				assert.Equal(t, float64(0), summary["memoryUsage"])
				assert.Equal(t, float64(0), summary["podCount"])
				assert.Equal(t, float64(0), summary["activePods"])
				assert.Equal(t, float64(0), summary["errorRate"])
				assert.Equal(t, float64(0), summary["responseTime"])
				
				assert.Contains(t, summary, "timestamp")
			},
		},
		{
			name: "invalid vector results",
			setupMock: func(mockClient *MockPrometheusClientForMetrics) {
				// Create a result with wrong type (should still handle gracefully)
				invalidResult := &prometheus.QueryResult{
					ResultType: "matrix", // Different than expected vector
					Result:     models.Matrix{},
				}
				
				// Mock all queries to return invalid results
				mockClient.On("Query", 
					mock.Anything, 
					"avg(rate(container_cpu_usage_seconds_total{namespace!=\"kube-system\"}[5m])) * 100", 
					mock.Anything,
				).Return(invalidResult, nil)
				
				mockClient.On("Query", 
					mock.Anything, 
					"sum(container_memory_usage_bytes{namespace!=\"kube-system\"}) / (1024 * 1024 * 1024)", 
					mock.Anything,
				).Return(invalidResult, nil)
				
				mockClient.On("Query", 
					mock.Anything, 
					"count(kube_pod_info)", 
					mock.Anything,
				).Return(invalidResult, nil)
				
				mockClient.On("Query", 
					mock.Anything, 
					"count(kube_pod_status_phase{phase=\"Running\"})", 
					mock.Anything,
				).Return(invalidResult, nil)
				
				mockClient.On("Query", 
					mock.Anything, 
					"sum(rate(app_errors_total[5m])) / sum(rate(app_request_total[5m])) * 100", 
					mock.Anything,
				).Return(invalidResult, nil)
				
				mockClient.On("Query", 
					mock.Anything, 
					"histogram_quantile(0.95, sum(rate(app_request_duration_seconds_bucket[5m])) by (le)) * 1000", 
					mock.Anything,
				).Return(invalidResult, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var response models.APIResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				
				assert.Equal(t, "success", response.Status)
				
				// Extract metrics summary from the response
				summary, ok := response.Data.(map[string]interface{})
				require.True(t, ok)
				
				// All metrics should be zero
				assert.Equal(t, float64(0), summary["cpuUsage"])
				assert.Equal(t, float64(0), summary["memoryUsage"])
				assert.Equal(t, float64(0), summary["podCount"])
				assert.Equal(t, float64(0), summary["activePods"])
				assert.Equal(t, float64(0), summary["errorRate"])
				assert.Equal(t, float64(0), summary["responseTime"])
				
				assert.Contains(t, summary, "timestamp")
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock client
			mockClient := new(MockPrometheusClientForMetrics)
			
			// Setup mock expectations
			tc.setupMock(mockClient)
			
			// Create the handler
			handler := &MetricsSummaryHandler{
				promClient: mockClient,
				cache:      cacheInstance,
				log:        log,
			}
			
			// Create a request
			req, err := http.NewRequest(http.MethodGet, "/api/metrics/summary", nil)
			require.NoError(t, err)
			
			// Create a response recorder
			rr := httptest.NewRecorder()
			
			// Serve the request
			handler.Handle(rr, req)
			
			// Check the status code
			assert.Equal(t, tc.expectedStatus, rr.Code)
			
			// Validate the response body
			tc.validateBody(t, rr.Body.Bytes())
			
			// Verify all expectations were met
			mockClient.AssertExpectations(t)
		})
	}
}

func TestMetricsSummaryHandler_CacheHit(t *testing.T) {
	// Create a logger
	log := logger.New(logger.Config{
		Level:  logger.DebugLevel,
		Format: logger.JSONFormat,
	})

	// Create a cache
	cacheInstance := cache.New(1*time.Minute, 100, 5*time.Minute)
	
	// Create a mock client
	mockClient := new(MockPrometheusClientForMetrics)
	
	// Setup mock expectations - should be called only once due to caching
	mockClient.On("Query", 
		mock.Anything, 
		"avg(rate(container_cpu_usage_seconds_total{namespace!=\"kube-system\"}[5m])) * 100", 
		mock.Anything,
	).Return(createMockQueryResult("vector", 65.5), nil).Once()
	
	mockClient.On("Query", 
		mock.Anything, 
		"sum(container_memory_usage_bytes{namespace!=\"kube-system\"}) / (1024 * 1024 * 1024)", 
		mock.Anything,
	).Return(createMockQueryResult("vector", 12.3), nil).Once()
	
	mockClient.On("Query", 
		mock.Anything, 
		"count(kube_pod_info)", 
		mock.Anything,
	).Return(createMockQueryResult("vector", 42), nil).Once()
	
	mockClient.On("Query", 
		mock.Anything, 
		"count(kube_pod_status_phase{phase=\"Running\"})", 
		mock.Anything,
	).Return(createMockQueryResult("vector", 38), nil).Once()
	
	mockClient.On("Query", 
		mock.Anything, 
		"sum(rate(app_errors_total[5m])) / sum(rate(app_request_total[5m])) * 100", 
		mock.Anything,
	).Return(createMockQueryResult("vector", 0.52), nil).Once()
	
	mockClient.On("Query", 
		mock.Anything, 
		"histogram_quantile(0.95, sum(rate(app_request_duration_seconds_bucket[5m])) by (le)) * 1000", 
		mock.Anything,
	).Return(createMockQueryResult("vector", 187.5), nil).Once()
	
	// Create the handler
	handler := &MetricsSummaryHandler{
		promClient: mockClient,
		cache:      cacheInstance,
		log:        log,
	}
	
	// Make two requests - the second should be a cache hit
	for i := 0; i < 2; i++ {
		// Create a request
		req, err := http.NewRequest(http.MethodGet, "/api/metrics/summary", nil)
		require.NoError(t, err)
		
		// Create a response recorder
		rr := httptest.NewRecorder()
		
		// Serve the request
		handler.Handle(rr, req)
		
		// Check the status code
		assert.Equal(t, http.StatusOK, rr.Code)
		
		// Both responses should have the same content
		var response models.APIResponse
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, "success", response.Status)
		
		// Extract metrics summary from the response
		summary, ok := response.Data.(map[string]interface{})
		require.True(t, ok)
		
		// Validate metrics values
		assert.Equal(t, 65.5, summary["cpuUsage"])
		assert.Equal(t, 12.3, summary["memoryUsage"])
		assert.Equal(t, float64(42), summary["podCount"])
		assert.Equal(t, float64(38), summary["activePods"])
		assert.Equal(t, 0.52, summary["errorRate"])
		assert.Equal(t, 187.5, summary["responseTime"])
	}
	
	// Verify all expectations were met (each called exactly once)
	mockClient.AssertExpectations(t)
}

// Helper function to create a mock query result
func createMockQueryResult(resultType string, value float64) *prometheus.QueryResult {
	// Create a sample with the given value
	sample := &models.Sample{
		Metric: models.Metric{
			"__name__": "test_metric",
		},
		Value:     models.SampleValue(value),
		Timestamp: models.Time(time.Now().Unix()),
	}
	
	// Create a vector with the sample
	vector := models.Vector{sample}
	
	// Return a query result with the vector
	return &prometheus.QueryResult{
		ResultType: resultType,
		Result:     vector,
	}
}