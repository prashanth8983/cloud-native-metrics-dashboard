package models

import (
	"time"

	"github.com/prometheus/common/model"
)

// APIResponse is the standard response format for all API endpoints
type APIResponse struct {
	Status  string      `json:"status"`            // "success" or "error"
	Data    interface{} `json:"data,omitempty"`    // Response data (omitted if error)
	Error   string      `json:"error,omitempty"`   // Error message (omitted if success)
	Code    int         `json:"code,omitempty"`    // HTTP status code (omitted if success)
	TraceID string      `json:"traceId,omitempty"` // Request trace ID for debugging
}

// QueryRequest represents a Prometheus query request
type QueryRequest struct {
	Query string     `json:"query" validate:"required"`                      // Prometheus query string
	Time  *time.Time `json:"time,omitempty" validate:"omitempty"`            // Optional query time
	Stats bool       `json:"stats,omitempty" validate:"omitempty"`           // Include query stats
	Debug bool       `json:"debug,omitempty" validate:"omitempty"`           // Include debug info
	Limit int        `json:"limit,omitempty" validate:"omitempty,min=1"`     // Limit number of results
	Format string    `json:"format,omitempty" validate:"omitempty,oneof='' json csv table"`
}

// RangeQueryRequest represents a Prometheus range query request
type RangeQueryRequest struct {
	Query string     `json:"query" validate:"required"`                      // Prometheus query string
	Start *time.Time `json:"start" validate:"required"`                      // Start time
	End   *time.Time `json:"end" validate:"required,gtfield=Start"`          // End time
	Step  string     `json:"step" validate:"required"`                       // Step interval (duration)
	Stats bool       `json:"stats,omitempty" validate:"omitempty"`           // Include query stats
	Debug bool       `json:"debug,omitempty" validate:"omitempty"`           // Include debug info
	Limit int        `json:"limit,omitempty" validate:"omitempty,min=1"`     // Limit number of results
	Format string    `json:"format,omitempty" validate:"omitempty,oneof='' json csv table"`
}

// AlertsRequest represents a request for alerts
type AlertsRequest struct {
	Filter      string   `json:"filter,omitempty" validate:"omitempty"`               // Filter alerts by label
	Silenced    *bool    `json:"silenced,omitempty" validate:"omitempty"`             // Include silenced alerts
	Inhibited   *bool    `json:"inhibited,omitempty" validate:"omitempty"`            // Include inhibited alerts
	Active      *bool    `json:"active,omitempty" validate:"omitempty"`               // Include active alerts
	Severity    []string `json:"severity,omitempty" validate:"omitempty,dive,oneof=critical warning info"` // Filter by severity
	Format      string   `json:"format,omitempty" validate:"omitempty,oneof='' json csv table"`
}

// MetricsSummaryRequest represents a request for a metrics summary
type MetricsSummaryRequest struct {
	Metrics []string `json:"metrics,omitempty" validate:"omitempty,dive,required"` // Specific metrics to include
	Format  string   `json:"format,omitempty" validate:"omitempty,oneof='' json csv table"`
}

// MetricMetadata represents metadata for a metric
type MetricMetadata struct {
	Type  string `json:"type"`  // Metric type (counter, gauge, etc.)
	Help  string `json:"help"`  // Help text
	Unit  string `json:"unit"`  // Unit of measurement
}

// Target represents a scrape target
type Target struct {
	Endpoint     string            `json:"endpoint"`               // Target endpoint
	Labels       map[string]string `json:"labels"`                 // Target labels
	State        string            `json:"state"`                  // Current state (up/down)
	LastScrape   time.Time         `json:"lastScrape"`             // Last scrape time
	ScrapeDuration float64         `json:"scrapeDuration"`         // Duration of last scrape
	Error        string            `json:"error,omitempty"`        // Last error (if any)
	Health       string            `json:"health"`                 // Health status
}

// MetricValue represents a single metric value
type MetricValue struct {
	Metric map[string]string `json:"metric"`                       // Metric labels
	Value  float64           `json:"value"`                        // Current value
	Time   time.Time         `json:"time"`                         // Timestamp
}

// MetricSeries represents a time series of metric values
type MetricSeries struct {
	Metric map[string]string `json:"metric"`                       // Metric labels
	Values []TimeValuePair   `json:"values"`                       // Time series values
}

// TimeValuePair represents a single time-value pair in a time series
type TimeValuePair struct {
	Time  time.Time `json:"time"`                                 // Timestamp
	Value float64   `json:"value"`                                // Value
}

// QueryResult represents a processed result from Prometheus
type QueryResult struct {
	ResultType string        `json:"resultType"`                   // Type of result (vector, matrix, etc.)
	Vectors    []MetricValue `json:"vectors,omitempty"`            // Vector results
	Matrix     []MetricSeries `json:"matrix,omitempty"`            // Matrix results
	Scalar     *TimeValuePair `json:"scalar,omitempty"`            // Scalar result
	String     *string        `json:"string,omitempty"`            // String result
	Warnings   []string       `json:"warnings,omitempty"`          // Any warnings
}

// Alert represents an alert from Prometheus
type Alert struct {
	Labels      map[string]string `json:"labels"`                  // Alert labels
	Annotations map[string]string `json:"annotations"`             // Alert annotations
	State       string            `json:"state"`                   // Alert state
	ActiveAt    time.Time         `json:"activeAt"`                // When the alert became active
	Value       float64           `json:"value"`                   // Current value
	Silenced    bool              `json:"silenced"`                // Whether the alert is silenced
	Inhibited   bool              `json:"inhibited"`               // Whether the alert is inhibited
	SilenceURL  string            `json:"silenceUrl,omitempty"`    // URL to silence the alert
}

// RuleGroup represents a group of alerting or recording rules
type RuleGroup struct {
	Name     string `json:"name"`                                  // Group name
	File     string `json:"file"`                                  // File the group is defined in
	Rules    []Rule `json:"rules"`                                 // Rules in the group
	Interval int    `json:"interval"`                              // Evaluation interval in seconds
}

// Rule represents an alerting or recording rule
type Rule struct {
	Name        string            `json:"name"`                    // Rule name
	Query       string            `json:"query"`                   // PromQL query
	Duration    int               `json:"duration,omitempty"`      // For alerting rules, the alert duration
	Labels      map[string]string `json:"labels,omitempty"`        // Rule labels
	Annotations map[string]string `json:"annotations,omitempty"`   // Rule annotations
	Alerts      []Alert           `json:"alerts,omitempty"`        // For alerting rules, current alerts
	Health      string            `json:"health"`                  // Rule health
	LastError   string            `json:"lastError,omitempty"`     // Last error evaluating the rule
	Type        string            `json:"type"`                    // "alerting" or "recording"
}

// HealthStatus represents the health status of the Prometheus server
type HealthStatus struct {
	Status      string    `json:"status"`                          // "up" or "down"
	Version     string    `json:"version,omitempty"`               // Prometheus version
	Uptime      string    `json:"uptime,omitempty"`                // Server uptime
	LastReload  time.Time `json:"lastReload,omitempty"`            // Last configuration reload time
	StorageSize int64     `json:"storageSize,omitempty"`           // Size of TSDB storage
	NumSeries   int64     `json:"numSeries,omitempty"`             // Number of time series
	NumSamples  int64     `json:"numSamples,omitempty"`            // Number of samples
	Flags       []string  `json:"flags,omitempty"`                 // Command-line flags
}

// MetricsSummary represents a summary of key metrics
type MetricsSummary struct {
	CPUUsage     float64 `json:"cpuUsage"`                         // CPU usage (percentage)
	MemoryUsage  float64 `json:"memoryUsage"`                      // Memory usage (GB)
	PodCount     int     `json:"podCount"`                         // Number of pods
	ActivePods   int     `json:"activePods"`                       // Number of active pods
	ErrorRate    float64 `json:"errorRate"`                        // Error rate (percentage)
	ResponseTime float64 `json:"responseTime"`                     // Response time (ms)
	Timestamp    time.Time `json:"timestamp"`                      // Timestamp
}

// ToAPIResponse creates a standard API response
func ToAPIResponse(status string, data interface{}) APIResponse {
	return APIResponse{
		Status: status,
		Data:   data,
	}
}

// ErrorResponse creates an error API response
func ErrorResponse(code int, message string) APIResponse {
	return APIResponse{
		Status: "error",
		Error:  message,
		Code:   code,
	}
}

// ConvertVectorResult converts a Prometheus vector result to the API model
func ConvertVectorResult(result model.Vector) []MetricValue {
	values := make([]MetricValue, 0, len(result))
	
	for _, sample := range result {
		metricLabels := make(map[string]string, len(sample.Metric))
		for name, value := range sample.Metric {
			metricLabels[string(name)] = string(value)
		}
		
		values = append(values, MetricValue{
			Metric: metricLabels,
			Value:  float64(sample.Value),
			Time:   sample.Timestamp.Time(),
		})
	}
	
	return values
}

// ConvertMatrixResult converts a Prometheus matrix result to the API model
func ConvertMatrixResult(result model.Matrix) []MetricSeries {
	series := make([]MetricSeries, 0, len(result))
	
	for _, s := range result {
		metricLabels := make(map[string]string, len(s.Metric))
		for name, value := range s.Metric {
			metricLabels[string(name)] = string(value)
		}
		
		values := make([]TimeValuePair, 0, len(s.Values))
		for _, point := range s.Values {
			values = append(values, TimeValuePair{
				Time:  point.Timestamp.Time(),
				Value: float64(point.Value),
			})
		}
		
		series = append(series, MetricSeries{
			Metric: metricLabels,
			Values: values,
		})
	}
	
	return series
}