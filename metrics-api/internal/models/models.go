package models

import (
	"errors"
	"time"
)

// Common errors
var (
	ErrInvalidQuery      = errors.New("invalid query")
	ErrInvalidTimeRange  = errors.New("invalid time range")
	ErrMetricNotFound    = errors.New("metric not found")
	ErrTooManyDataPoints = errors.New("query would return too many data points")
)

// QueryResponse represents the response from an instant query
type QueryResponse struct {
	Query     string      `json:"query"`
	QueryTime time.Time   `json:"query_time"`
	Status    string      `json:"status"`
	Data      []DataPoint `json:"data"`
}

// DataPoint represents a single data point from a query
type DataPoint struct {
	MetricName string            `json:"metric_name"`
	Labels     map[string]string `json:"labels"`
	Value      float64           `json:"value"`
	Timestamp  time.Time         `json:"timestamp"`
}

// RangeQueryResponse represents the response from a range query
type RangeQueryResponse struct {
	Query   string       `json:"query"`
	Start   time.Time    `json:"start"`
	End     time.Time    `json:"end"`
	Step    time.Duration `json:"step"`
	Status  string       `json:"status"`
	Series  []TimeSeries `json:"series"`
}

// TimeSeries represents a time series of data points
type TimeSeries struct {
	MetricName string            `json:"metric_name"`
	Labels     map[string]string `json:"labels"`
	DataPoints []TimeValuePair   `json:"data_points"`
}

// TimeValuePair represents a time-value pair
type TimeValuePair struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// InstantQueryParams represents parameters for an instant query
type InstantQueryParams struct {
	Query string    `json:"query"`
	Time  time.Time `json:"time"`
}

// RangeQueryParams represents parameters for a range query
type RangeQueryParams struct {
	Query string    `json:"query"`
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	Step  int       `json:"step"` // Step in seconds
}

// QueryValidation represents the result of validating a query
type QueryValidation struct {
	Query   string `json:"query"`
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}

// Alert represents a Prometheus alert
type Alert struct {
	Name        string            `json:"name"`
	State       string            `json:"state"`
	Severity    string            `json:"severity"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Summary     string            `json:"summary"`
	ActiveAt    time.Time         `json:"active_at"`
	Value       float64           `json:"value"`
}

// AlertGroup represents a group of alerts
type AlertGroup struct {
	Name   string  `json:"name"`
	Count  int     `json:"count"`
	Alerts []Alert `json:"alerts"`
}

// AlertSummary represents a summary of alerts
type AlertSummary struct {
	FiringCount       int             `json:"firing_count"`
	PendingCount      int             `json:"pending_count"`
	ResolvedCount     int             `json:"resolved_count"`
	TotalCount        int             `json:"total_count"`
	SeverityBreakdown []SeverityCount `json:"severity_breakdown"`
	MostRecentAlert   Alert           `json:"most_recent_alert,omitempty"`
	TimeSinceLastAlert string         `json:"time_since_last_alert,omitempty"`
	LastUpdated       time.Time       `json:"last_updated"`
}

// SeverityCount represents the count of alerts by severity
type SeverityCount struct {
	Severity string `json:"severity"`
	Count    int    `json:"count"`
}

// MetricSummary represents a summary of a metric
type MetricSummary struct {
	Name        string        `json:"name"`
	Labels      []string      `json:"labels"`
	Cardinality int           `json:"cardinality"`
	Stats       MetricStats   `json:"stats"`
	LastUpdated time.Time     `json:"last_updated"`
	Samples     []MetricSample `json:"samples"`
}

// MetricStats represents statistical information about a metric
type MetricStats struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
	Avg float64 `json:"avg"`
}

// MetricSample represents a sample of a metric
type MetricSample struct {
	Labels    map[string]string `json:"labels"`
	Value     float64           `json:"value"`
	Timestamp time.Time         `json:"timestamp"`
}

// TopMetric represents a metric with additional information
type TopMetric struct {
	Name        string  `json:"name"`
	Cardinality int     `json:"cardinality"`
	SampleRate  float64 `json:"sample_rate,omitempty"`
}

// MetricHealth represents the health status of a metric
type MetricHealth struct {
	Name        string    `json:"name"`
	Exists      bool      `json:"exists"`
	IsStale     bool      `json:"is_stale"`
	HasGaps     bool      `json:"has_gaps"`
	LastScraped time.Time `json:"last_scraped"`
	CheckedAt   time.Time `json:"checked_at"`
}

// HealthStatus represents the overall health status of the system
type HealthStatus struct {
	Status    string            `json:"status"`
	Version   string            `json:"version"`
	Uptime    string            `json:"uptime"`
	Timestamp time.Time         `json:"timestamp"`
	Checks    map[string]string `json:"checks"`
	Details   map[string]any    `json:"details,omitempty"`
}