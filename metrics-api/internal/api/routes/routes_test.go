// internal/api/routes/routes_test.go
package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"metrics-api/internal/cache"
	"metrics-api/internal/config"
	"metrics-api/internal/prometheus"
	"metrics-api/pkg/logger"
)

func createTestRouter() (*APIRouter, *httptest.Server) {
	// Create a mock Prometheus server
	promServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		
		// Return mock responses based on the URL path
		switch r.URL.Path {
		case "/api/v1/query":
			w.Write([]byte(`{
				"status": "success",
				"data": {
					"resultType": "vector",
					"result": [
						{
							"metric": {"__name__": "up", "instance": "localhost:9090", "job": "prometheus"},
							"value": [1607443034.458, "1"]
						}
					]
				}
			}`))
			
		case "/api/v1/query_range":
			w.Write([]byte(`{
				"status": "success",
				"data": {
					"resultType": "matrix",
					"result": [
						{
							"metric": {"__name__": "up", "instance": "localhost:9090", "job": "prometheus"},
							"values": [
								[1607443000.458, "1"],
								[1607443060.458, "1"],
								[1607443120.458, "1"]
							]
						}
					]
				}
			}`))
			
		case "/api/v1/alerts":
			w.Write([]byte(`{
				"status": "success",
				"data": {
					"alerts": []
				}
			}`))
			
		case "/api/v1/rules":
			w.Write([]byte(`{
				"status": "success",
				"data": {
					"groups": []
				}
			}`))
			
		case "/api/v1/targets":
			w.Write([]byte(`{
				"status": "success",
				"data": {
					"activeTargets": [],
					"droppedTargets": []
				}
			}`))
			
		case "/api/v1/buildinfo":
			w.Write([]byte(`{
				"status": "success",
				"data": {
					"version": "2.30.0",
					"revision": "abc123",
					"branch": "master",
					"buildUser": "user",
					"buildDate": "2021-01-01T00:00:00Z",
					"goVersion": "go1.16"
				}
			}`))
			
		default:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"status":"error","error":"not found"}`))
		}
	}))
	
	// Create the config
	cfg := &config.Config{}
	cfg.Prometheus.URL = promServer.URL
	cfg.Prometheus.Timeout = 5 * time.Second
	cfg.Server.Port = 8000
	cfg.CORS.Enabled = true
	cfg.CORS.AllowedOrigins = []string{"*"}
	cfg.CORS.AllowedMethods = []string{"GET", "POST", "OPTIONS"}
	cfg.CORS.AllowedHeaders = []string{"Content-Type", "Authorization"}
	cfg.Metrics.Enabled = true
	cfg.Metrics.Path = "/metrics"
	
	// Create a logger
	log := logger.New(logger.Config{
		Level:  logger.DebugLevel,
		Format: logger.JSONFormat,
	})
	
	// Create a cache
	cacheInstance := cache.New(1*time.Minute, 100, 5*time.Minute)
	
	// Create a Prometheus client
	promClient, err := prometheus.New(cfg, log)
	if err != nil {
		panic(err)
	}
	
	// Create the API router
	router := NewAPIRouter(cfg, promClient, cacheInstance, log)
	
	// Set up the routes
	router.SetupRoutes()
	
	return router, promServer
}

func TestAPIRouter_HealthEndpoint(t *testing.T) {
	router, server := createTestRouter()
	defer server.Close()
	
	// Create a test server using the router
	ts := httptest.NewServer(router)
	defer ts.Close()
	
	// Make a request to the health endpoint
	resp, err := http.Get(ts.URL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()
	
	// Check the response
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
}

func TestAPIRouter_MetricsEndpoint(t *testing.T) {
	router, server := createTestRouter()
	defer server.Close()
	
	// Create a test server using the router
	ts := httptest.NewServer(router)
	defer ts.Close()
	
	// Make a request to the metrics endpoint
	resp, err := http.Get(ts.URL + "/metrics")
	require.NoError(t, err)
	defer resp.Body.Close()
	
	// Check the response
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAPIRouter_QueryEndpoint(t *testing.T) {
	router, server := createTestRouter()
	defer server.Close()
	
	// Create a test server using the router
	ts := httptest.NewServer(router)
	defer ts.Close()
	
	// Make a request to the query endpoint
	resp, err := http.Get(ts.URL + "/api/query?query=up")
	require.NoError(t, err)
	defer resp.Body.Close()
	
	// Check the response
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
}

func TestAPIRouter_QueryRangeEndpoint(t *testing.T) {
	router, server := createTestRouter()
	defer server.Close()
	
	// Create a test server using the router
	ts := httptest.NewServer(router)
	defer ts.Close()
	
	// Get the current time
	end := time.Now()
	start := end.Add(-1 * time.Hour)
	
	// Make a request to the query range endpoint
	resp, err := http.Get(ts.URL + "/api/query_range?query=up&start=" + start.Format(time.RFC3339) + "&end=" + end.Format(time.RFC3339) + "&step=1m")
	require.NoError(t, err)
	defer resp.Body.Close()
	
	// Check the response
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
}

func TestAPIRouter_AlertsEndpoint(t *testing.T) {
	router, server := createTestRouter()
	defer server.Close()
	
	// Create a test server using the router
	ts := httptest.NewServer(router)
	defer ts.Close()
	
	// Make a request to the alerts endpoint
	resp, err := http.Get(ts.URL + "/api/alerts")
	require.NoError(t, err)
	defer resp.Body.Close()
	
	// Check the response
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
}

func TestAPIRouter_MetricsSummaryEndpoint(t *testing.T) {
	router, server := createTestRouter()
	defer server.Close()
	
	// Create a test server using the router
	ts := httptest.NewServer(router)
	defer ts.Close()
	
	// Make a request to the metrics summary endpoint
	resp, err := http.Get(ts.URL + "/api/metrics/summary")
	require.NoError(t, err)
	defer resp.Body.Close()
	
	// Check the response
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
}

func TestAPIRouter_NotFoundEndpoint(t *testing.T) {
	router, server := createTestRouter()
	defer server.Close()
	
	// Create a test server using the router
	ts := httptest.NewServer(router)
	defer ts.Close()
	
	// Make a request to a non-existent endpoint
	resp, err := http.Get(ts.URL + "/api/not-found")
	require.NoError(t, err)
	defer resp.Body.Close()
	
	// Check the response
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestAPIRouter_CORS(t *testing.T) {
	router, server := createTestRouter()
	defer server.Close()
	
	// Create a test server using the router
	ts := httptest.NewServer(router)
	defer ts.Close()
	
	// Create a request
	req, err := http.NewRequest("OPTIONS", ts.URL+"/api/query", nil)
	require.NoError(t, err)
	
	// Set the origin header
	req.Header.Set("Origin", "http://example.com")
	
	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	
	// Check the response
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
	assert.Contains(t, resp.Header.Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(t, resp.Header.Get("Access-Control-Allow-Methods"), "POST")
}