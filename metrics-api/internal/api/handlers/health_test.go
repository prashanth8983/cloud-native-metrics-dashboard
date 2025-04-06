package handlers

import (
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
	"metrics-api/pkg/logger"
)

// MockPrometheusClientForHealth is a mock for the Prometheus client
type MockPrometheusClientForHealth struct {
	mock.Mock
}

func (m *MockPrometheusClientForHealth) IsHealthy(ctx interface{}) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}

func (m *MockPrometheusClientForHealth) BuildInfo(ctx interface{}) (interface{}, error) {
	args := m.Called(ctx)
	return args.Get(0), args.Error(1)
}

func TestHealthHandler_Handle(t *testing.T) {
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
		setupMock      func(mock *MockPrometheusClientForHealth)
		expectedStatus int
		validateBody   func(t *testing.T, body []byte)
	}{
		{
			name: "healthy prometheus",
			setupMock: func(mockClient *MockPrometheusClientForHealth) {
				mockClient.On("IsHealthy", mock.Anything).Return(true, nil)
				mockClient.On("BuildInfo", mock.Anything).Return(map[string]string{
					"version": "2.30.0",
				}, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var response models.APIResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				
				assert.Equal(t, "success", response.Status)
				
				// Extract health status from the response
				health, ok := response.Data.(map[string]interface{})
				require.True(t, ok)
				
				assert.Equal(t, "up", health["status"])
				assert.Equal(t, "2.30.0", health["version"])
				assert.Contains(t, health, "uptime")
			},
		},
		{
			name: "unhealthy prometheus",
			setupMock: func(mockClient *MockPrometheusClientForHealth) {
				mockClient.On("IsHealthy", mock.Anything).Return(false, nil)
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var response models.APIResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				
				assert.Equal(t, "success", response.Status)
				
				// Extract health status from the response
				health, ok := response.Data.(map[string]interface{})
				require.True(t, ok)
				
				assert.Equal(t, "down", health["status"])
				assert.NotContains(t, health, "version") // Version should not be included when down
			},
		},
		{
			name: "prometheus error",
			setupMock: func(mockClient *MockPrometheusClientForHealth) {
				mockClient.On("IsHealthy", mock.Anything).Return(false, errors.New("connection error"))
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var response models.APIResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				
				assert.Equal(t, "success", response.Status)
				
				// Extract health status from the response
				health, ok := response.Data.(map[string]interface{})
				require.True(t, ok)
				
				assert.Equal(t, "down", health["status"])
			},
		},
		{
			name: "healthy but build info error",
			setupMock: func(mockClient *MockPrometheusClientForHealth) {
				mockClient.On("IsHealthy", mock.Anything).Return(true, nil)
				mockClient.On("BuildInfo", mock.Anything).Return(nil, errors.New("failed to get build info"))
			},
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var response models.APIResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				
				assert.Equal(t, "success", response.Status)
				
				// Extract health status from the response
				health, ok := response.Data.(map[string]interface{})
				require.True(t, ok)
				
				assert.Equal(t, "up", health["status"])
				assert.NotContains(t, health, "version") // Version should not be included when build info fails
				assert.Contains(t, health, "uptime")    // Uptime should still be included
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock client
			mockClient := new(MockPrometheusClientForHealth)
			
			// Setup mock expectations
			tc.setupMock(mockClient)
			
			// Create the handler
			handler := &HealthHandler{
				promClient: mockClient,
				cache:      cacheInstance,
				log:        log,
				startTime:  time.Now().Add(-1 * time.Hour), // Started 1 hour ago
			}
			
			// Create a request
			req, err := http.NewRequest(http.MethodGet, "/health", nil)
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

func TestHealthHandler_CacheHit(t *testing.T) {
	// Create a logger
	log := logger.New(logger.Config{
		Level:  logger.DebugLevel,
		Format: logger.JSONFormat,
	})

	// Create a cache
	cacheInstance := cache.New(1*time.Minute, 100, 5*time.Minute)
	
	// Create a mock client
	mockClient := new(MockPrometheusClientForHealth)
	
	// Setup mock expectations - should be called only once due to caching
	mockClient.On("IsHealthy", mock.Anything).Return(true, nil).Once()
	mockClient.On("BuildInfo", mock.Anything).Return(map[string]string{
		"version": "2.30.0",
	}, nil).Once()
	
	// Create the handler
	handler := &HealthHandler{
		promClient: mockClient,
		cache:      cacheInstance,
		log:        log,
		startTime:  time.Now().Add(-1 * time.Hour), // Started 1 hour ago
	}
	
	// Make two requests - the second should be a cache hit
	for i := 0; i < 2; i++ {
		// Create a request
		req, err := http.NewRequest(http.MethodGet, "/health", nil)
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
		
		// Extract health status from the response
		health, ok := response.Data.(map[string]interface{})
		require.True(t, ok)
		
		assert.Equal(t, "up", health["status"])
		assert.Equal(t, "2.30.0", health["version"])
	}
	
	// Verify all expectations were met (called exactly once)
	mockClient.AssertExpectations(t)
}