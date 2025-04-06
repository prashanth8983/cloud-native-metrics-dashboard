// internal/prometheus/client.go
package prometheus

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"

	"metrics-api/internal/config"
	"metrics-api/pkg/logger"
)

// Client wraps the Prometheus API client
type Client struct {
	api    v1.API
	config *config.Config
	log    *logger.Logger
}

// QueryResult represents the result of a Prometheus query
type QueryResult struct {
	ResultType string      `json:"resultType"`
	Result     model.Value `json:"result"`
	Warnings   []string    `json:"warnings,omitempty"`
}

// RangeQueryOptions holds the parameters for a range query
type RangeQueryOptions struct {
	Start time.Time
	End   time.Time
	Step  time.Duration
}

// New creates a new Prometheus client
func New(cfg *config.Config, log *logger.Logger) (*Client, error) {
	if log == nil {
		log = logger.Default()
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        cfg.Prometheus.MaxConnections,
			MaxIdleConnsPerHost: cfg.Prometheus.MaxConnections,
			IdleConnTimeout:     cfg.Prometheus.KeepAlive,
		},
		Timeout: cfg.Prometheus.Timeout,
	}

	apiClient, err := api.NewClient(api.Config{
		Address: cfg.Prometheus.URL,
		Client:  httpClient,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Prometheus client: %w", err)
	}

	return &Client{
		api:    v1.NewAPI(apiClient),
		config: cfg,
		log:    log.WithField("component", "prometheus_client"),
	}, nil
}

// Query performs an instant query at the specified time
func (c *Client) Query(ctx context.Context, query string, ts time.Time) (*QueryResult, error) {
	c.log.Debug().Str("query", query).Msg("executing instant query")

	ctx, cancel := context.WithTimeoutCause(ctx, c.config.Prometheus.Timeout, fmt.Errorf("query timeout exceeded"))
	defer cancel()

	result, warnings, err := c.api.Query(ctx, query, ts)
	if err != nil {
		c.log.Error().Err(err).Str("query", query).Msg("failed to execute query")
		return nil, fmt.Errorf("query %q failed: %w", query, err)
	}

	if len(warnings) > 0 {
		c.log.Warn().Strs("warnings", warnings).Str("query", query).Msg("query returned warnings")
	}

	return &QueryResult{
		ResultType: result.Type().String(),
		Result:     result,
		Warnings:   warnings,
	}, nil
}

// QueryRange performs a range query over the specified time range
func (c *Client) QueryRange(ctx context.Context, query string, opts RangeQueryOptions) (*QueryResult, error) {
	c.log.Debug().
		Str("query", query).
		Time("start", opts.Start).
		Time("end", opts.End).
		Dur("step", opts.Step).
		Msg("executing range query")

	ctx, cancel := context.WithTimeoutCause(ctx, c.config.Prometheus.Timeout, fmt.Errorf("range query timeout exceeded"))
	defer cancel()

	r := v1.Range{
		Start: opts.Start,
		End:   opts.End,
		Step:  opts.Step,
	}

	result, warnings, err := c.api.QueryRange(ctx, query, r)
	if err != nil {
		c.log.Error().
			Err(err).
			Str("query", query).
			Time("start", opts.Start).
			Time("end", opts.End).
			Dur("step", opts.Step).
			Msg("failed to execute range query")
		return nil, fmt.Errorf("range query %q failed: %w", query, err)
	}

	if len(warnings) > 0 {
		c.log.Warn().
			Strs("warnings", warnings).
			Str("query", query).
			Msg("range query returned warnings")
	}

	return &QueryResult{
		ResultType: result.Type().String(),
		Result:     result,
		Warnings:   warnings,
	}, nil
}

// Alerts gets the list of active alerts
func (c *Client) Alerts(ctx context.Context) (v1.AlertsResult, error) {
	c.log.Debug().Msg("fetching alerts")

	ctx, cancel := context.WithTimeoutCause(ctx, c.config.Prometheus.Timeout, fmt.Errorf("alerts fetch timeout exceeded"))
	defer cancel()

	alerts, err := c.api.Alerts(ctx)
	if err != nil {
		c.log.Error().Err(err).Msg("failed to fetch alerts")
		return v1.AlertsResult{}, fmt.Errorf("failed to fetch alerts: %w", err)
	}

	return alerts, nil
}

// Rules gets the list of alerting and recording rules
func (c *Client) Rules(ctx context.Context) (v1.RulesResult, error) {
	c.log.Debug().Msg("fetching rules")

	ctx, cancel := context.WithTimeoutCause(ctx, c.config.Prometheus.Timeout, fmt.Errorf("rules fetch timeout exceeded"))
	defer cancel()

	rules, err := c.api.Rules(ctx)
	if err != nil {
		c.log.Error().Err(err).Msg("failed to fetch rules")
		return v1.RulesResult{}, fmt.Errorf("failed to fetch rules: %w", err)
	}

	return rules, nil
}

// Targets gets the current targets
func (c *Client) Targets(ctx context.Context) (v1.TargetsResult, error) {
	c.log.Debug().Msg("fetching targets")

	ctx, cancel := context.WithTimeoutCause(ctx, c.config.Prometheus.Timeout, fmt.Errorf("targets fetch timeout exceeded"))
	defer cancel()

	targets, err := c.api.Targets(ctx)
	if err != nil {
		c.log.Error().Err(err).Msg("failed to fetch targets")
		return v1.TargetsResult{}, fmt.Errorf("failed to fetch targets: %w", err)
	}

	return targets, nil
}

// LabelValues gets the label values for a given label name
func (c *Client) LabelValues(ctx context.Context, label string, matches []string) (model.LabelValues, error) {
	c.log.Debug().Str("label", label).Strs("matches", matches).Msg("fetching label values")

	ctx, cancel := context.WithTimeoutCause(ctx, c.config.Prometheus.Timeout, fmt.Errorf("label values fetch timeout exceeded"))
	defer cancel()

	values, warnings, err := c.api.LabelValues(ctx, label, matches, time.Now())
	if err != nil {
		c.log.Error().Err(err).Str("label", label).Msg("failed to fetch label values")
		return nil, fmt.Errorf("failed to fetch label values for %q: %w", label, err)
	}

	if len(warnings) > 0 {
		c.log.Warn().Strs("warnings", warnings).Str("label", label).Msg("label values returned warnings")
	}

	return values, nil
}

// Series finds series matching the given selectors
func (c *Client) Series(ctx context.Context, matches []string, start, end time.Time) ([]model.LabelSet, error) {
	c.log.Debug().
		Strs("matches", matches).
		Time("start", start).
		Time("end", end).
		Msg("fetching series")

	ctx, cancel := context.WithTimeoutCause(ctx, c.config.Prometheus.Timeout, fmt.Errorf("series fetch timeout exceeded"))
	defer cancel()

	series, warnings, err := c.api.Series(ctx, matches, start, end)
	if err != nil {
		c.log.Error().
			Err(err).
			Strs("matches", matches).
			Msg("failed to fetch series")
		return nil, fmt.Errorf("failed to fetch series for matches %v: %w", matches, err)
	}

	if len(warnings) > 0 {
		c.log.Warn().Strs("warnings", warnings).Strs("matches", matches).Msg("series returned warnings")
	}

	return series, nil
}

// MetricMetadata gets the metadata for metrics
func (c *Client) MetricMetadata(ctx context.Context, metric string) (v1.MetricMetadata, error) {
	c.log.Debug().Str("metric", metric).Msg("fetching metric metadata")

	ctx, cancel := context.WithTimeoutCause(ctx, c.config.Prometheus.Timeout, fmt.Errorf("metric metadata fetch timeout exceeded"))
	defer cancel()

	metadata, err := c.api.Metadata(ctx, metric, "")
	if err != nil {
		c.log.Error().Err(err).Str("metric", metric).Msg("failed to fetch metric metadata")
		return nil, fmt.Errorf("failed to fetch metadata for metric %q: %w", metric, err)
	}

	return metadata, nil
}

// BuildInfo gets the build information of the Prometheus server
func (c *Client) BuildInfo(ctx context.Context) (v1.BuildInfo, error) {
	c.log.Debug().Msg("fetching build info")

	ctx, cancel := context.WithTimeoutCause(ctx, c.config.Prometheus.Timeout, fmt.Errorf("build info fetch timeout exceeded"))
	defer cancel()

	buildInfo, err := c.api.Buildinfo(ctx)
	if err != nil {
		c.log.Error().Err(err).Msg("failed to fetch build info")
		return v1.BuildInfo{}, fmt.Errorf("failed to fetch build info: %w", err)
	}

	return buildInfo, nil
}

// ExtractVectorValue extracts the value from a vector result
func ExtractVectorValue(result *QueryResult) (float64, error) {
	if result == nil || result.Result == nil {
		return 0, fmt.Errorf("nil query result or result value")
	}

	if result.ResultType != "vector" {
		return 0, fmt.Errorf("expected vector result, got %s", result.ResultType)
	}

	vector, ok := result.Result.(model.Vector)
	if !ok {
		return 0, fmt.Errorf("failed to cast result to vector")
	}

	if len(vector) == 0 {
		return 0, fmt.Errorf("empty vector result")
	}

	return float64(vector[0].Value), nil
}

// FormatDuration formats a duration in a human-readable format
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.2fms", float64(d.Milliseconds()))
	} else if d < time.Minute {
		return fmt.Sprintf("%.2fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.2fm", d.Minutes())
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%.2fh", d.Hours())
	}
	return fmt.Sprintf("%.2fd", d.Hours()/24)
}

// FormatBytes formats bytes in a human-readable format
func FormatBytes(bytes float64) string {
	const unit = 1024.0
	if bytes < unit {
		return fmt.Sprintf("%.2f B", bytes)
	}
	div, exp := float64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB", bytes/div, "KMGTPE"[exp])
}

// IsHealthy checks if the Prometheus server is healthy
func (c *Client) IsHealthy(ctx context.Context) (bool, error) {
	ctx, cancel := context.WithTimeoutCause(ctx, 5*time.Second, fmt.Errorf("health check timeout exceeded"))
	defer cancel()

	_, err := c.api.Buildinfo(ctx)
	if err != nil {
		return false, fmt.Errorf("health check failed: %w", err)
	}

	return true, nil
}

// RoundTimeForStep rounds a time to a step boundary to ensure consistent results
func RoundTimeForStep(t time.Time, step time.Duration) time.Time {
	stepNano := step.Nanoseconds()
	timestamp := t.UnixNano()
	remainder := timestamp % stepNano
	if remainder == 0 {
		return t
	}
	return time.Unix(0, timestamp-remainder).UTC()
}