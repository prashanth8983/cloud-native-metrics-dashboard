package models

import (
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
)

func TestToAPIResponse(t *testing.T) {
	// Test success response
	data := map[string]string{"key": "value"}
	response := ToAPIResponse("success", data)
	
	assert.Equal(t, "success", response.Status)
	assert.Equal(t, data, response.Data)
	assert.Empty(t, response.Error)
	assert.Zero(t, response.Code)
	
	// Test with nil data
	response = ToAPIResponse("success", nil)
	assert.Equal(t, "success", response.Status)
	assert.Nil(t, response.Data)
}

func TestErrorResponse(t *testing.T) {
	// Test error response
	response := ErrorResponse(404, "Not found")
	
	assert.Equal(t, "error", response.Status)
	assert.Equal(t, "Not found", response.Error)
	assert.Equal(t, 404, response.Code)
	assert.Nil(t, response.Data)
}

func TestConvertVectorResult(t *testing.T) {
	// Create a test vector
	timestamp := model.Time(1607443034458)
	vector := model.Vector{
		&model.Sample{
			Metric: model.Metric{
				"__name__": "up",
				"instance": "localhost:9090",
				"job":      "prometheus",
			},
			Value:     model.SampleValue(1),
			Timestamp: timestamp,
		},
		&model.Sample{
			Metric: model.Metric{
				"__name__": "up",
				"instance": "localhost:9100",
				"job":      "node",
			},
			Value:     model.SampleValue(0),
			Timestamp: timestamp,
		},
	}
	
	// Convert the vector
	result := ConvertVectorResult(vector)
	
	// Verify the result
	assert.Len(t, result, 2)
	
	// Verify the first sample
	assert.Equal(t, "up", result[0].Metric["__name__"])
	assert.Equal(t, "localhost:9090", result[0].Metric["instance"])
	assert.Equal(t, "prometheus", result[0].Metric["job"])
	assert.Equal(t, float64(1), result[0].Value)
	assert.Equal(t, timestamp.Time(), result[0].Time)
	
	// Verify the second sample
	assert.Equal(t, "up", result[1].Metric["__name__"])
	assert.Equal(t, "localhost:9100", result[1].Metric["instance"])
	assert.Equal(t, "node", result[1].Metric["job"])
	assert.Equal(t, float64(0), result[1].Value)
	assert.Equal(t, timestamp.Time(), result[1].Time)
}

func TestConvertMatrixResult(t *testing.T) {
	// Create a test matrix
	timestamp1 := model.Time(1607443000458)
	timestamp2 := model.Time(1607443060458)
	matrix := model.Matrix{
		&model.SampleStream{
			Metric: model.Metric{
				"__name__": "up",
				"instance": "localhost:9090",
				"job":      "prometheus",
			},
			Values: []model.SamplePair{
				{
					Timestamp: timestamp1,
					Value:     model.SampleValue(1),
				},
				{
					Timestamp: timestamp2,
					Value:     model.SampleValue(1),
				},
			},
		},
		&model.SampleStream{
			Metric: model.Metric{
				"__name__": "up",
				"instance": "localhost:9100",
				"job":      "node",
			},
			Values: []model.SamplePair{
				{
					Timestamp: timestamp1,
					Value:     model.SampleValue(0),
				},
				{
					Timestamp: timestamp2,
					Value:     model.SampleValue(1),
				},
			},
		},
	}
	
	// Convert the matrix
	result := ConvertMatrixResult(matrix)
	
	// Verify the result
	assert.Len(t, result, 2)
	
	// Verify the first stream
	assert.Equal(t, "up", result[0].Metric["__name__"])
	assert.Equal(t, "localhost:9090", result[0].Metric["instance"])
	assert.Equal(t, "prometheus", result[0].Metric["job"])
	assert.Len(t, result[0].Values, 2)
	assert.Equal(t, timestamp1.Time(), result[0].Values[0].Time)
	assert.Equal(t, float64(1), result[0].Values[0].Value)
	assert.Equal(t, timestamp2.Time(), result[0].Values[1].Time)
	assert.Equal(t, float64(1), result[0].Values[1].Value)
	
	// Verify the second stream
	assert.Equal(t, "up", result[1].Metric["__name__"])
	assert.Equal(t, "localhost:9100", result[1].Metric["instance"])
	assert.Equal(t, "node", result[1].Metric["job"])
	assert.Len(t, result[1].Values, 2)
	assert.Equal(t, timestamp1.Time(), result[1].Values[0].Time)
	assert.Equal(t, float64(0), result[1].Values[0].Value)
	assert.Equal(t, timestamp2.Time(), result[1].Values[1].Time)
	assert.Equal(t, float64(1), result[1].Values[1].Value)
}

func TestTimeValuePair(t *testing.T) {
	// Test time value pair
	now := time.Now()
	pair := TimeValuePair{
		Time:  now,
		Value: 123.456,
	}
	
	assert.Equal(t, now, pair.Time)
	assert.Equal(t, 123.456, pair.Value)
}

func TestMetricValue(t *testing.T) {
	// Test metric value
	now := time.Now()
	metric := map[string]string{
		"__name__": "test_metric",
		"instance": "localhost",
	}
	
	metricValue := MetricValue{
		Metric: metric,
		Value:  123.456,
		Time:   now,
	}
	
	assert.Equal(t, metric, metricValue.Metric)
	assert.Equal(t, 123.456, metricValue.Value)
	assert.Equal(t, now, metricValue.Time)
}

func TestMetricSeries(t *testing.T) {
	// Test metric series
	metric := map[string]string{
		"__name__": "test_metric",
		"instance": "localhost",
	}
	
	time1 := time.Now()
	time2 := time1.Add(time.Minute)
	
	values := []TimeValuePair{
		{
			Time:  time1,
			Value: 123.456,
		},
		{
			Time:  time2,
			Value: 234.567,
		},
	}
	
	series := MetricSeries{
		Metric: metric,
		Values: values,
	}
	
	assert.Equal(t, metric, series.Metric)
	assert.Equal(t, values, series.Values)
	assert.Len(t, series.Values, 2)
}

func TestQueryResult(t *testing.T) {
	// Test query result
	vectors := []MetricValue{
		{
			Metric: map[string]string{"__name__": "test_metric"},
			Value:  123.456,
			Time:   time.Now(),
		},
	}
	
	warnings := []string{"Warning 1", "Warning 2"}
	
	queryResult := QueryResult{
		ResultType: "vector",
		Vectors:    vectors,
		Warnings:   warnings,
	}
	
	assert.Equal(t, "vector", queryResult.ResultType)
	assert.Equal(t, vectors, queryResult.Vectors)
	assert.Equal(t, warnings, queryResult.Warnings)
}

func TestAlert(t *testing.T) {
	// Test alert
	now := time.Now()
	labels := map[string]string{
		"alertname": "HighCPUUsage",
		"severity":  "warning",
	}
	
	annotations := map[string]string{
		"summary":     "High CPU usage detected",
		"description": "CPU usage is above 80%",
	}
	
	alert := Alert{
		Labels:      labels,
		Annotations: annotations,
		State:       "firing",
		ActiveAt:    now,
		Value:       85.5,
		Silenced:    false,
		Inhibited:   false,
	}
	
	assert.Equal(t, labels, alert.Labels)
	assert.Equal(t, annotations, alert.Annotations)
	assert.Equal(t, "firing", alert.State)
	assert.Equal(t, now, alert.ActiveAt)
	assert.Equal(t, 85.5, alert.Value)
	assert.False(t, alert.Silenced)
	assert.False(t, alert.Inhibited)
}

func TestRule(t *testing.T) {
	// Test rule
	labels := map[string]string{
		"severity": "warning",
	}
	
	annotations := map[string]string{
		"summary":     "High CPU usage detected",
		"description": "CPU usage is above 80%",
	}
	
	alerts := []Alert{
		{
			Labels:      map[string]string{"instance": "server1"},
			Annotations: annotations,
			State:       "firing",
			ActiveAt:    time.Now(),
			Value:       85.5,
		},
	}
	
	rule := Rule{
		Name:        "HighCPUUsage",
		Query:       "avg(cpu_usage_percent) > 80",
		Duration:    300,
		Labels:      labels,
		Annotations: annotations,
		Alerts:      alerts,
		Health:      "ok",
		Type:        "alerting",
	}
	
	assert.Equal(t, "HighCPUUsage", rule.Name)
	assert.Equal(t, "avg(cpu_usage_percent) > 80", rule.Query)
	assert.Equal(t, 300, rule.Duration)
	assert.Equal(t, labels, rule.Labels)
	assert.Equal(t, annotations, rule.Annotations)
	assert.Equal(t, alerts, rule.Alerts)
	assert.Equal(t, "ok", rule.Health)
	assert.Equal(t, "alerting", rule.Type)
}