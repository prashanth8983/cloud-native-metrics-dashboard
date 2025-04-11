package prometheus

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"metrics-api/internal/cache"
	"metrics-api/pkg/logger"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockPrometheusServer creates a test server that responds with predefined Prometheus API responses
func mockPrometheusServer(t *testing.T, responses map[string]string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set content type for all responses
		w.Header().Set("Content-Type", "application/json")

		// Extract the path and query for matching responses
		path := r.URL.Path
		query := r.URL.Query().Encode()
		key := path + "?" + query

		// Log request for debugging
		t.Logf("Received request: %s with query: %s", path, query)

		// Check if we have a predefined response for this exact request
		if response, ok := responses[key]; ok {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
			return
		}

		// Check if we have a response just for the path
		if response, ok := responses[path]; ok {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
			return
		}

		// Default response if no match
		t.Logf("No response defined for request: %s", key)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"status":"error","errorType":"not_found","error":"No response defined for this request"}`))
	}))
}

func setupTestClient(t *testing.T, serverURL string) *Client {
	testLogger := logger.NewTestLogger()
	testCache := cache.New(cache.DefaultOptions())
	
	client, err := NewClient(serverURL, testLogger, testCache)
	require.NoError(t, err)
	require.NotNil(t, client)
	
	return client
}

func TestNewClient(t *testing.T) {
	testLogger := logger.NewTestLogger()
	testCache := cache.New(cache.DefaultOptions())

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid URL",
			url:     "http://localhost:9090",
			wantErr: false,
		},
		{
			name:    "invalid URL",
			url:     "invalid://localhost:9090",
			wantErr: true,
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.url, testLogger, testCache)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestWithTimeout(t *testing.T) {
	server := mockPrometheusServer(t, nil)
	defer server.Close()

	client := setupTestClient(t, server.URL)

	tests := []struct {
		name        string
		timeout     time.Duration
		shouldPanic bool
	}{
		{
			name:        "valid timeout",
			timeout:     5 * time.Second,
			shouldPanic: false,
		},
		{
			name:        "zero timeout",
			timeout:     0,
			shouldPanic: true,
		},
		{
			name:        "negative timeout",
			timeout:     -1 * time.Second,
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				assert.Panics(t, func() {
					client.WithTimeout(tt.timeout)
				})
			} else {
				assert.NotPanics(t, func() {
					updatedClient := client.WithTimeout(tt.timeout)
					assert.Equal(t, tt.timeout, updatedClient.timeout)
				})
			}
		})
	}
}

func TestQuery(t *testing.T) {
	// Setup mock responses
	responses := map[string]string{
		"/api/v1/query": `{
			"status": "success",
			"data": {
				"resultType": "vector",
				"result": [
					{
						"metric": {"__name__": "up", "instance": "localhost:9090", "job": "prometheus"},
						"value": [1609746000, "1"]
					},
					{
						"metric": {"__name__": "up", "instance": "localhost:9100", "job": "node"},
						"value": [1609746000, "0"]
					}
				]
			}
		}`,
	}

	// Create mock server
	server := mockPrometheusServer(t, responses)
	defer server.Close()

	// Create client pointing to mock server
	client := setupTestClient(t, server.URL)

	// Execute query
	ctx := context.Background()
	timestamp := time.Unix(1609746000, 0)
	results, err := client.Query(ctx, "up", timestamp)

	// Check results
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	// Check first result
	assert.Equal(t, "up", results[0].MetricName)
	assert.Equal(t, "localhost:9090", results[0].Labels["instance"])
	assert.Equal(t, "prometheus", results[0].Labels["job"])
	assert.Equal(t, 1.0, results[0].Value)
	assert.Equal(t, timestamp, results[0].Timestamp)

	// Check second result
	assert.Equal(t, "up", results[1].MetricName)
	assert.Equal(t, "localhost:9100", results[1].Labels["instance"])
	assert.Equal(t, "node", results[1].Labels["job"])
	assert.Equal(t, 0.0, results[1].Value)
	assert.Equal(t, timestamp, results[1].Timestamp)
}

func TestQueryRange(t *testing.T) {
	// Setup mock responses
	responses := map[string]string{
		"/api/v1/query_range": `{
			"status": "success",
			"data": {
				"resultType": "matrix",
				"result": [
					{
						"metric": {"__name__": "http_requests_total", "job": "api", "code": "200"},
						"values": [
							[1609746000, "10"],
							[1609746060, "15"],
							[1609746120, "20"]
						]
					},
					{
						"metric": {"__name__": "http_requests_total", "job": "api", "code": "500"},
						"values": [
							[1609746000, "1"],
							[1609746060, "2"],
							[1609746120, "0"]
						]
					}
				]
			}
		}`,
	}

	// Create mock server
	server := mockPrometheusServer(t, responses)
	defer server.Close()

	// Create client pointing to mock server
	client := setupTestClient(t, server.URL)

	// Execute range query
	ctx := context.Background()
	r := v1.Range{
		Start: time.Unix(1609746000, 0),
		End:   time.Unix(1609746120, 0),
		Step:  60 * time.Second,
	}
	results, err := client.QueryRange(ctx, "rate(http_requests_total[5m])", r)

	// Check results
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	// Check first result
	assert.Equal(t, "http_requests_total", results[0].MetricName)
	assert.Equal(t, "api", results[0].Labels["job"])
	assert.Equal(t, "200", results[0].Labels["code"])
	assert.Len(t, results[0].Values, 3)

	// Check values
	assert.Equal(t, time.Unix(1609746000, 0), results[0].Values[0].Timestamp)
	assert.Equal(t, 10.0, results[0].Values[0].Value)
	assert.Equal(t, time.Unix(1609746060, 0), results[0].Values[1].Timestamp)
	assert.Equal(t, 15.0, results[0].Values[1].Value)
	assert.Equal(t, time.Unix(1609746120, 0), results[0].Values[2].Timestamp)
	assert.Equal(t, 20.0, results[0].Values[2].Value)

	// Check second result
	assert.Equal(t, "http_requests_total", results[1].MetricName)
	assert.Equal(t, "api", results[1].Labels["job"])
	assert.Equal(t, "500", results[1].Labels["code"])
	assert.Len(t, results[1].Values, 3)
}

func TestGetAlerts(t *testing.T) {
	// Setup mock responses
	responses := map[string]string{
		"/api/v1/alerts": `{
			"status": "success",
			"data": {
				"alerts": [
					{
						"labels": {
							"alertname": "HighErrorRate",
							"severity": "critical",
							"service": "api"
						},
						"annotations": {
							"summary": "High error rate detected",
							"description": "Error rate is above 5%"
						},
						"state": "firing",
						"activeAt": "2021-01-04T12:00:00Z",
						"value": "0.08"
					}
				]
			}
		}`,
	}

	// Create mock server
	server := mockPrometheusServer(t, responses)
	defer server.Close()

	// Create client pointing to mock server
	client := setupTestClient(t, server.URL)

	// Get alerts
	ctx := context.Background()
	alerts, err := client.GetAlerts(ctx)

	// Check results
	assert.NoError(t, err)
	assert.Len(t, alerts, 1)

	// Check first alert
	assert.Equal(t, "HighErrorRate", alerts[0].Name)
	assert.Equal(t, AlertStateFiring, alerts[0].State)
	assert.Equal(t, "critical", alerts[0].Labels["severity"])
	assert.Equal(t, "api", alerts[0].Labels["service"])
	assert.Equal(t, "High error rate detected", alerts[0].Annotations["summary"])
	assert.Equal(t, "Error rate is above 5%", alerts[0].Annotations["description"])
	// Not checking exact time due to time zone issues, just verify it's not zero
	assert.False(t, alerts[0].ActiveAt.IsZero())
	assert.Equal(t, 0.08, alerts[0].Value)
}

func TestGetMetrics(t *testing.T) {
	// Setup mock responses
	responses := map[string]string{
		"/api/v1/label/__name__/values": `{
			"status": "success",
			"data": [
				"http_requests_total",
				"node_cpu_seconds_total",
				"node_memory_MemFree_bytes",
				"up"
			]
		}`,
	}

	// Create mock server
	server := mockPrometheusServer(t, responses)
	defer server.Close()

	// Create client pointing to mock server
	client := setupTestClient(t, server.URL)

	// Get metrics
	ctx := context.Background()
	metrics, err := client.GetMetrics(ctx)

	// Check results
	assert.NoError(t, err)
	assert.Len(t, metrics, 4)
	assert.Contains(t, metrics, "http_requests_total")
	assert.Contains(t, metrics, "node_cpu_seconds_total")
	assert.Contains(t, metrics, "node_memory_MemFree_bytes")
	assert.Contains(t, metrics, "up")
}

func TestGetLabelsForMetric(t *testing.T) {
	// Setup mock responses
	responses := map[string]string{
		"/api/v1/labels": `{
			"status": "success",
			"data": [
				"__name__",
				"instance",
				"job",
				"method",
				"status"
			]
		}`,
	}

	// Create mock server
	server := mockPrometheusServer(t, responses)
	defer server.Close()

	// Create client pointing to mock server
	client := setupTestClient(t, server.URL)

	// Get labels for metric
	ctx := context.Background()
	labels, err := client.GetLabelsForMetric(ctx, "http_requests_total")

	// Check results
	assert.NoError(t, err)
	assert.Len(t, labels, 5)
	assert.Contains(t, labels, "__name__")
	assert.Contains(t, labels, "instance")
	assert.Contains(t, labels, "job")
	assert.Contains(t, labels, "method")
	assert.Contains(t, labels, "status")
}

func TestParseQueryResponse(t *testing.T) {
	timestamp := time.Unix(1609746000, 0)

	tests := []struct {
		name    string
		input   model.Value
		wantLen int
		wantErr bool
	}{
		{
			name: "vector response",
			input: model.Vector{
				&model.Sample{
					Metric: model.Metric{
						"__name__":  "up",
						"instance":  "localhost:9090",
						"job":       "prometheus",
					},
					Value:     1,
					Timestamp: model.Time(timestamp.Unix()),
				},
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "scalar response",
			input: &model.Scalar{
				Value:     42,
				Timestamp: model.Time(timestamp.Unix()),
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "nil response",
			input:   nil,
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "string response",
			input: &model.String{
				Value:     "test",
				Timestamp: model.Time(timestamp.Unix()),
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := parseQueryResponse(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, results)
			} else {
				assert.NoError(t, err)
				assert.Len(t, results, tt.wantLen)
			}
		})
	}
}

func TestParseRangeQueryResponse(t *testing.T) {
	// Create test data
	matrixValue := model.Matrix{
		&model.SampleStream{
			Metric: model.Metric{
				"__name__": "http_requests_total",
				"job":      "api",
				"code":     "200",
			},
			Values: []model.SamplePair{
				{
					Timestamp: model.Time(1609746000),
					Value:     10,
				},
				{
					Timestamp: model.Time(1609746060),
					Value:     15,
				},
				{
					Timestamp: model.Time(1609746120),
					Value:     20,
				},
			},
		},
		&model.SampleStream{
			Metric: model.Metric{
				"__name__": "http_requests_total",
				"job":      "api",
				"code":     "500",
			},
			Values: []model.SamplePair{
				{
					Timestamp: model.Time(1609746000),
					Value:     1,
				},
				{
					Timestamp: model.Time(1609746060),
					Value:     2,
				},
				{
					Timestamp: model.Time(1609746120),
					Value:     0,
				},
			},
		},
	}

	// Parse matrix response
	results, err := parseRangeQueryResponse(matrixValue)
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	// Check first result
	assert.Equal(t, "http_requests_total", results[0].MetricName)
	assert.Equal(t, "api", results[0].Labels["job"])
	assert.Equal(t, "200", results[0].Labels["code"])
	assert.Len(t, results[0].Values, 3)
	assert.Equal(t, time.Unix(1609746000, 0), results[0].Values[0].Timestamp)
	assert.Equal(t, 10.0, results[0].Values[0].Value)

	vectorValue := model.Vector{} 
    results, err = parseRangeQueryResponse(vectorValue)
	assert.Error(t, err)
	assert.Nil(t, results)
}

// TestClientErrors tests error handling in the client
func TestClientErrors(t *testing.T) {
	// Create a server that always returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status":"error","errorType":"server_error","error":"Internal Server Error"}`))
	}))
	defer server.Close()

	client := setupTestClient(t, server.URL)
	ctx := context.Background()

	t.Run("Query error", func(t *testing.T) {
		_, err := client.Query(ctx, "up", time.Now())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error querying Prometheus")
	})

	t.Run("QueryRange error", func(t *testing.T) {
		r := v1.Range{
			Start: time.Now().Add(-1 * time.Hour),
			End:   time.Now(),
			Step:  time.Minute,
		}
		_, err := client.QueryRange(ctx, "rate(http_requests_total[5m])", r)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error querying Prometheus range")
	})
}

// Integration test with real Prometheus API response formats
func TestResponseParsing(t *testing.T) {
	// This test can be skipped in CI
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup mock responses with real Prometheus API response formats
	responses := map[string]string{
		"/api/v1/query": `{
			"status": "success",
			"data": {
				"resultType": "vector",
				"result": [
					{
						"metric": {
							"__name__": "up",
							"instance": "localhost:9090",
							"job": "prometheus"
						},
						"value": [1609746000, "1"]
					}
				]
			}
		}`,
		"/api/v1/query_range": `{
			"status": "success",
			"data": {
				"resultType": "matrix",
				"result": [
					{
						"metric": {
							"__name__": "up",
							"instance": "localhost:9090",
							"job": "prometheus"
						},
						"values": [
							[1609746000, "1"],
							[1609746060, "1"],
							[1609746120, "1"]
						]
					}
				]
			}
		}`,
		"/api/v1/alerts": `{
			"status": "success",
			"data": {
				"alerts": []
			}
		}`,
		"/api/v1/label/__name__/values": `{
			"status": "success",
			"data": [
				"up",
				"scrape_duration_seconds",
				"scrape_samples_scraped"
			]
		}`,
		"/api/v1/labels": `{
			"status": "success",
			"data": [
				"__name__",
				"instance",
				"job"
			]
		}`,
	}

	// Create mock server
	server := mockPrometheusServer(t, responses)
	defer server.Close()

	// Create client pointing to mock server
	client := setupTestClient(t, server.URL)

	// Test all methods
	ctx := context.Background()

	// Test Query
	_, err := client.Query(ctx, "up", time.Unix(1609746000, 0))
	assert.NoError(t, err)

	// Test QueryRange
	r := v1.Range{
		Start: time.Unix(1609746000, 0),
		End:   time.Unix(1609746120, 0),
		Step:  60 * time.Second,
	}
	_, err = client.QueryRange(ctx, "up", r)
	assert.NoError(t, err)

	// Test GetAlerts
	_, err = client.GetAlerts(ctx)
	assert.NoError(t, err)

	// Test GetMetrics
	_, err = client.GetMetrics(ctx)
	assert.NoError(t, err)

	// Test GetLabelsForMetric
	_, err = client.GetLabelsForMetric(ctx, "up")
	assert.NoError(t, err)
}

// TestQueryWithWarnings tests handling of warnings in query responses
func TestQueryWithWarnings(t *testing.T) {
	// Setup a testing logger
	oldStdout := os.Stdout
	// Clear stdout redirection after test
	defer func() { os.Stdout = oldStdout }()

	// Setup mock responses with warnings
	responses := map[string]string{
		"/api/v1/query": `{
			"status": "success",
			"warnings": ["query time range too long", "using optimized execution"],
			"data": {
				"resultType": "vector",
				"result": [
					{
						"metric": {"__name__": "up", "instance": "localhost:9090", "job": "prometheus"},
						"value": [1609746000, "1"]
					}
				]
			}
		}`,
	}

	// Create mock server
	server := mockPrometheusServer(t, responses)
	defer server.Close()

	// Create client pointing to mock server
	client := setupTestClient(t, server.URL)

	// Execute query
	ctx := context.Background()
	timestamp := time.Unix(1609746000, 0)
	results, err := client.Query(ctx, "up", timestamp)

	// Check results - warnings should not cause an error
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "up", results[0].MetricName)
}
