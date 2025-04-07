package prometheus

import (
	"context"
	"fmt"
	"time"

	"metrics-api/internal/cache"
	"metrics-api/pkg/logger" // Adjust path as needed

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

// QueryOption allows for functional parameter pattern
type QueryOption func(*QueryOptions)

// QueryOptions holds all configurable options for queries
type QueryOptions struct {
	Timeout      time.Duration
	CacheTTL     time.Duration
	UseCache     bool
	Labels       map[string]string
	SkipSanitize bool
}

// defaultQueryOptions provides sensible defaults
var defaultQueryOptions = QueryOptions{
	Timeout:  30 * time.Second,
	CacheTTL: 60 * time.Second,
	UseCache: true,
}

// WithTimeout sets a custom timeout for the query
func WithTimeout(timeout time.Duration) QueryOption {
	return func(o *QueryOptions) {
		o.Timeout = timeout
	}
}

// WithCacheTTL sets a custom cache TTL for the query results
func WithCacheTTL(ttl time.Duration) QueryOption {
	return func(o *QueryOptions) {
		o.CacheTTL = ttl
	}
}

// WithoutCache disables caching for this query
func WithoutCache() QueryOption {
	return func(o *QueryOptions) {
		o.UseCache = false
	}
}

// WithLabels adds additional label filters to the query
func WithLabels(labels map[string]string) QueryOption {
	return func(o *QueryOptions) {
		o.Labels = labels
	}
}

// WithoutSanitize skips query sanitization
func WithoutSanitize() QueryOption {
	return func(o *QueryOptions) {
		o.SkipSanitize = true
	}
}

// NewQueryClient creates a new Prometheus query client
func NewQueryClient(url string, logger logger.Logger, cache *cache.Cache) (*Client, error) {
	client, err := api.NewClient(api.Config{
		Address: url,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating Prometheus client: %w", err)
	}

	return &Client{
		api:    v1.NewAPI(client),
		logger: logger,
		cache:  cache,
	}, nil
}

// ExecuteInstantQuery performs an instant query against Prometheus
func (c *Client) ExecuteInstantQuery(ctx context.Context, query string, ts time.Time, opts ...QueryOption) ([]QueryResult, error) {
	options := defaultQueryOptions
	for _, opt := range opts {
		opt(&options)
	}

	if !options.SkipSanitize {
		var err error
		query, err = sanitizeQuery(query)
		if err != nil {
			return nil, fmt.Errorf("invalid query: %w", err)
		}
	}

	// Apply additional label filters if provided
	if len(options.Labels) > 0 {
		query = applyLabelFilters(query, options.Labels)
	}

	cacheKey := fmt.Sprintf("instant:%s:%d", query, ts.Unix())
	
	// Check cache first if enabled
	if options.UseCache {
		if cached, found := c.cache.Get(cacheKey); found {
			c.logger.Debug("cache hit for query", "query", query)
			return cached.([]QueryResult), nil
		}
	}

	// Set up timeout context
	queryCtx, cancel := context.WithTimeout(ctx, options.Timeout)
	defer cancel()

	// Execute query
	c.logger.Debug("executing instant query", "query", query)
	result, warnings, err := c.api.Query(queryCtx, query, ts)
	if err != nil {
		c.logger.Error("instant query failed", "query", query, "error", err)
		return nil, fmt.Errorf("prometheus query failed: %w", err)
	}

	// Log warnings if any
	for _, w := range warnings {
		c.logger.Warn("prometheus query warning", "query", query, "warning", w)
	}

	// Parse result
	queryResult, err := parseQueryResponse(result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query result: %w", err)
	}

	// Cache result if enabled
	if options.UseCache {
		c.cache.Set(cacheKey, queryResult)
	}

	return queryResult, nil
}

// ExecuteRangeQuery performs a range query against Prometheus
func (c *Client) ExecuteRangeQuery(
	ctx context.Context,
	query string,
	r v1.Range,
	opts ...QueryOption,
) ([]RangeQueryResult, error) {
	options := defaultQueryOptions
	for _, opt := range opts {
		opt(&options)
	}

	if !options.SkipSanitize {
		var err error
		query, err = sanitizeQuery(query)
		if err != nil {
			return nil, fmt.Errorf("invalid query: %w", err)
		}
	}

	// Apply additional label filters if provided
	if len(options.Labels) > 0 {
		query = applyLabelFilters(query, options.Labels)
	}

	cacheKey := fmt.Sprintf("range:%s:%d:%d:%d", query, r.Start.Unix(), r.End.Unix(), int(r.Step.Seconds()))
	
	// Check cache first if enabled
	if options.UseCache {
		if cached, found := c.cache.Get(cacheKey); found {
			c.logger.Debug("cache hit for range query", "query", query)
			return cached.([]RangeQueryResult), nil
		}
	}

	// Set up timeout context
	queryCtx, cancel := context.WithTimeout(ctx, options.Timeout)
	defer cancel()

	// Execute query
	c.logger.Debug("executing range query",
		"query", query,
		"start", r.Start.Format(time.RFC3339),
		"end", r.End.Format(time.RFC3339),
		"step", r.Step.String())

	result, warnings, err := c.api.QueryRange(queryCtx, query, r)
	if err != nil {
		c.logger.Error("range query failed",
			"query", query,
			"start", r.Start.Format(time.RFC3339),
			"end", r.End.Format(time.RFC3339),
			"step", r.Step.String(),
			"error", err)
		return nil, fmt.Errorf("prometheus range query failed: %w", err)
	}

	// Log warnings if any
	for _, w := range warnings {
		c.logger.Warn("prometheus range query warning", "query", query, "warning", w)
	}

	// Parse result
	queryResult, err := parseRangeQueryResponse(result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse range query result: %w", err)
	}

	// Cache result if enabled
	if options.UseCache {
		c.cache.Set(cacheKey, queryResult)
	}

	return queryResult, nil
}

// GetMetricSeries gets time series data for a specific metric
func (c *Client) GetMetricSeries(ctx context.Context, metric string, duration time.Duration, opts ...QueryOption) ([]RangeQueryResult, error) {
	options := defaultQueryOptions
	for _, opt := range opts {
		opt(&options)
	}

	// Set end time to now and start time based on the duration
	end := time.Now()
	start := end.Add(-duration)

	// Default step to 15 seconds or 1/100th of the duration, whichever is larger
	step := duration / 100
	if step < 15*time.Second {
		step = 15 * time.Second
	}

	// Create range
	r := v1.Range{
		Start: start,
		End:   end,
		Step:  step,
	}

	// Create a simple query to fetch the metric
	query := metric
	if !options.SkipSanitize {
		var err error
		query, err = sanitizeQuery(query)
		if err != nil {
			return nil, fmt.Errorf("invalid metric name: %w", err)
		}
	}

	// Apply additional label filters if provided
	if len(options.Labels) > 0 {
		query = applyLabelFilters(query, options.Labels)
	}

	return c.ExecuteRangeQuery(ctx, query, r, opts...)
}

// sanitizeQuery performs basic query sanitization
func sanitizeQuery(query string) (string, error) {
	if query == "" {
		return "", fmt.Errorf("empty query")
	}
	// Add more comprehensive query validation as needed
	return query, nil
}

// applyLabelFilters adds label filters to a query
func applyLabelFilters(query string, labels map[string]string) string {
	if len(labels) == 0 {
		return query
	}

	labelStr := "{"
	first := true
	for k, v := range labels {
		if !first {
			labelStr += ","
		}
		labelStr += fmt.Sprintf("%s=\"%s\"", k, v)
		first = false
	}
	labelStr += "}"
	
	return query + labelStr
}

// containsLabelFilters checks if a query already contains label filters
func containsLabelFilters(query string) bool {
	return query[len(query)-1] == '}' || query[len(query)-1] == ')'
}

// addToExistingLabelFilters adds labels to a query that already has filters
func addToExistingLabelFilters(query string, labels map[string]string) string {
	filterStr := ""
	for k, v := range labels {
		filterStr += fmt.Sprintf(" and %s=\"%s\"", k, v)
	}
	return query + filterStr
}