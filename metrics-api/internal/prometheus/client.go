package prometheus

import (
	"context"
	"fmt"
	"metrics-api/internal/cache"
	"metrics-api/pkg/logger"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type PrometheusAPI interface {
	Query(ctx context.Context, query string, ts time.Time) ([]QueryResult, error)
	QueryRange(ctx context.Context, query string, r v1.Range) ([]RangeQueryResult, error)
	GetMetrics(ctx context.Context) ([]string, error)
	GetAlerts(ctx context.Context) ([]Alert, error)
	GetLabelsForMetric(ctx context.Context, metricName string) ([]string, error)
}

// Client represents a Prometheus client wrapper
type Client struct {
	api       v1.API
	timeout   time.Duration
	logger logger.Logger
	cache  *cache.Cache
}

// QueryResult represents the result of a Prometheus query
type QueryResult struct {
	MetricName string
	Labels     map[string]string
	Value      float64
	Timestamp  time.Time
}

// RangeQueryResult represents the result of a Prometheus range query
type RangeQueryResult struct {
	MetricName string
	Labels     map[string]string
	Values     []TimeValuePair
}

// TimeValuePair represents a single time-value pair in a range query result
type TimeValuePair struct {
	Timestamp time.Time
	Value     float64
}

// AlertState represents the state of an alert
type AlertState string

const (
	AlertStateFiring   AlertState = "firing"
	AlertStatePending  AlertState = "pending"
	AlertStateInactive AlertState = "inactive"
)

// Alert represents an alert from Prometheus
type Alert struct {
	Name        string
	State       AlertState
	Labels      map[string]string
	Annotations map[string]string
	ActiveAt    time.Time
	Value       float64
}

// NewClient creates a new Prometheus client
func NewClient(url string) (*Client, error) {
	client, err := api.NewClient(api.Config{
		Address: url,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating Prometheus client: %w", err)
	}

	return &Client{
		api:     v1.NewAPI(client),
		timeout: 30 * time.Second,
	}, nil
}

// WithTimeout sets the client timeout for queries
func (c *Client) WithTimeout(timeout time.Duration) *Client {
	c.timeout = timeout
	return c
}

// Query performs an instant query against Prometheus
func (c *Client) Query(ctx context.Context, query string, ts time.Time) ([]QueryResult, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	value, warnings, err := c.api.Query(ctx, query, ts)
	if err != nil {
		return nil, fmt.Errorf("error querying Prometheus: %w", err)
	}

	if len(warnings) > 0 {
		// Log warnings, but continue processing
		for _, w := range warnings {
			fmt.Printf("Warning: %s\n", w)
		}
	}

	return parseQueryResponse(value)
}

// QueryRange performs a range query against Prometheus
func (c *Client) QueryRange(ctx context.Context, query string, r v1.Range) ([]RangeQueryResult, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	value, warnings, err := c.api.QueryRange(ctx, query, r)
	if err != nil {
		return nil, fmt.Errorf("error querying Prometheus range: %w", err)
	}

	if len(warnings) > 0 {
		// Log warnings, but continue processing
		for _, w := range warnings {
			fmt.Printf("Warning: %s\n", w)
		}
	}

	return parseRangeQueryResponse(value)
}

// GetAlerts gets the current alerts from Prometheus
func (c *Client) GetAlerts(ctx context.Context) ([]Alert, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	alertsResult, err := c.api.Alerts(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting alerts from Prometheus: %w", err)
	}

	alerts := make([]Alert, 0, len(alertsResult.Alerts))
	for _, a := range alertsResult.Alerts {
		alert := Alert{
			Name:        string(a.Labels["alertname"]),
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		}

		// Set state
		switch a.State {
		case "firing":
			alert.State = AlertStateFiring
		case "pending":
			alert.State = AlertStatePending
		default:
			alert.State = AlertStateInactive
		}

		// Copy labels and annotations
		for k, v := range a.Labels {
			alert.Labels[string(k)] = string(v)
		}
		for k, v := range a.Annotations {
			alert.Annotations[string(k)] = string(v)
		}

		// Set time and value if available
		if !a.ActiveAt.IsZero() {
			alert.ActiveAt = a.ActiveAt
		}
		
		value, err := strconv.ParseFloat(a.Value, 64)
		if err == nil && value > 0 {
			alert.Value = value
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// GetMetrics gets a list of metric names from Prometheus
func (c *Client) GetMetrics(ctx context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	metrics, _, err := c.api.LabelValues(ctx, "__name__", []string{}, time.Time{}, time.Time{})
	if err != nil {
		return nil, fmt.Errorf("error getting metrics from Prometheus: %w", err)
	}

	result := make([]string, 0, len(metrics))
	for _, m := range metrics {
		result = append(result, string(m))
	}

	return result, nil
}

// GetLabelsForMetric gets all labels for a specific metric
func (c *Client) GetLabelsForMetric(ctx context.Context, metricName string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	labels, _, err := c.api.LabelNames(ctx, []string{metricName}, time.Time{}, time.Time{})
	if err != nil {
		return nil, fmt.Errorf("error getting labels for metric %s: %w", metricName, err)
	}

	return labels, nil
}

// parseQueryResponse converts a Prometheus query result to our internal format
func parseQueryResponse(value model.Value) ([]QueryResult, error) {
	var results []QueryResult

	switch v := value.(type) {
	case model.Vector:
		for _, sample := range v {
			metricName := string(sample.Metric["__name__"])
			
			// Extract labels
			labels := make(map[string]string)
			for labelName, labelValue := range sample.Metric {
				if labelName != "__name__" {
					labels[string(labelName)] = string(labelValue)
				}
			}
			
			results = append(results, QueryResult{
				MetricName: metricName,
				Labels:     labels,
				Value:      float64(sample.Value),
				Timestamp:  sample.Timestamp.Time(),
			})
		}
	case *model.Scalar:
		results = append(results, QueryResult{
			MetricName: "scalar",
			Labels:     map[string]string{},
			Value:      float64(v.Value),
			Timestamp:  v.Timestamp.Time(),
		})
	case *model.String:
		// String results are unusual but possible
		return nil, fmt.Errorf("string results not supported")
	default:
		return nil, fmt.Errorf("unsupported result format: %T", v)
	}

	return results, nil
}

// parseRangeQueryResponse converts a Prometheus range query result to our internal format
func parseRangeQueryResponse(value model.Value) ([]RangeQueryResult, error) {
	var results []RangeQueryResult

	switch v := value.(type) {
	case model.Matrix:
		for _, stream := range v {
			metricName := string(stream.Metric["__name__"])
			
			// Extract labels
			labels := make(map[string]string)
			for labelName, labelValue := range stream.Metric {
				if labelName != "__name__" {
					labels[string(labelName)] = string(labelValue)
				}
			}
			
			// Extract values
			values := make([]TimeValuePair, 0, len(stream.Values))
			for _, pair := range stream.Values {
				values = append(values, TimeValuePair{
					Timestamp: pair.Timestamp.Time(),
					Value:     float64(pair.Value),
				})
			}
			
			results = append(results, RangeQueryResult{
				MetricName: metricName,
				Labels:     labels,
				Values:     values,
			})
		}
	default:
		return nil, fmt.Errorf("unsupported result format for range query: %T", v)
	}

	return results, nil
}

