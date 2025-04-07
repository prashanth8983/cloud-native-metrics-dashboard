package service

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"metrics-api/internal/models"
	"metrics-api/internal/prometheus"
	"metrics-api/pkg/logger"
)

// MetricsService handles metrics-related operations
type MetricsService struct {
	client  *prometheus.Client
	logger  logger.Logger
	cache   map[string]cachedMetricSummary
	cacheMu sync.RWMutex
	cacheTTL time.Duration
}

type cachedMetricSummary struct {
	data      models.MetricSummary
	timestamp time.Time
}

// NewMetricsService creates a new metrics service
func NewMetricsService(client *prometheus.Client, logger logger.Logger) *MetricsService {
	return &MetricsService{
		client:   client,
		logger:   logger,
		cache:    make(map[string]cachedMetricSummary),
		cacheTTL: 5 * time.Minute, // Default cache TTL
	}
}

// WithCacheTTL sets the cache TTL
func (s *MetricsService) WithCacheTTL(ttl time.Duration) *MetricsService {
	s.cacheTTL = ttl
	return s
}

// GetMetrics retrieves the list of available metrics
func (s *MetricsService) GetMetrics(ctx context.Context) ([]string, error) {
	metrics, err := s.client.GetMetrics(ctx)
	if err != nil {
		s.logger.Errorf("Failed to get metrics: %v", err)
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}
	
	// Sort metrics for consistent output
	sort.Strings(metrics)
	
	return metrics, nil
}

// GetMetricSummary provides a summary of a specific metric
func (s *MetricsService) GetMetricSummary(ctx context.Context, metricName string) (*models.MetricSummary, error) {
	// Check cache first
	s.cacheMu.RLock()
	cached, exists := s.cache[metricName]
	s.cacheMu.RUnlock()
	
	if exists && time.Since(cached.timestamp) < s.cacheTTL {
		s.logger.Debugf("Cache hit for metric summary: %s", metricName)
		return &cached.data, nil
	}
	
	s.logger.Debugf("Cache miss for metric summary: %s", metricName)
	
	// Fetch labels
	labels, err := s.client.GetLabelsForMetric(ctx, metricName)
	if err != nil {
		s.logger.Errorf("Failed to get labels for metric %s: %v", metricName, err)
		return nil, fmt.Errorf("failed to get labels for metric %s: %w", metricName, err)
	}
	
	// Query current value (if available)
	now := time.Now()
	query := metricName
	results, err := s.client.Query(ctx, query, now)
	if err != nil {
		s.logger.Errorf("Failed to query metric %s: %v", metricName, err)
		return nil, fmt.Errorf("failed to query metric %s: %w", metricName, err)
	}
	
	// Get cardinality
	cardinalityQuery := fmt.Sprintf("count(%s)", metricName)
	cardinalityResults, err := s.client.Query(ctx, cardinalityQuery, now)
	if err != nil {
		s.logger.Errorf("Failed to get cardinality for metric %s: %v", metricName, err)
		return nil, fmt.Errorf("failed to get cardinality for metric %s: %w", metricName, err)
	}
	
	var cardinality int
	if len(cardinalityResults) > 0 {
		cardinality = int(cardinalityResults[0].Value)
	}
	
	// Get min, max, avg for the last hour
	_ = now.Add(-1 * time.Hour)
	
	statsQueries := map[string]string{
		"min": fmt.Sprintf("min_over_time(%s[1h])", metricName),
		"max": fmt.Sprintf("max_over_time(%s[1h])", metricName),
		"avg": fmt.Sprintf("avg_over_time(%s[1h])", metricName),
	}
	
	stats := models.MetricStats{}
	
	for statName, statQuery := range statsQueries {
		statResults, err := s.client.Query(ctx, statQuery, now)
		if err != nil {
			s.logger.Warnf("Failed to get %s for metric %s: %v", statName, metricName, err)
			continue
		}
		
		if len(statResults) > 0 {
			switch statName {
			case "min":
				stats.Min = statResults[0].Value
			case "max":
				stats.Max = statResults[0].Value
			case "avg":
				stats.Avg = statResults[0].Value
			}
		}
	}
	
	// Build samples
	samples := make([]models.MetricSample, 0, len(results))
	for _, result := range results {
		samples = append(samples, models.MetricSample{
			Labels:    result.Labels,
			Value:     result.Value,
			Timestamp: result.Timestamp,
		})
	}
	
	// Limited to first 10 samples for summary
	if len(samples) > 10 {
		samples = samples[:10]
	}
	
	summary := &models.MetricSummary{
		Name:        metricName,
		Labels:      labels,
		Cardinality: cardinality,
		Stats:       stats,
		LastUpdated: now,
		Samples:     samples,
	}
	
	// Update cache
	s.cacheMu.Lock()
	s.cache[metricName] = cachedMetricSummary{
		data:      *summary,
		timestamp: time.Now(),
	}
	s.cacheMu.Unlock()
	
	return summary, nil
}

// GetTopMetrics gets the top N metrics by cardinality or activity
func (s *MetricsService) GetTopMetrics(ctx context.Context, limit int) ([]models.TopMetric, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}
	
	// Get all metrics first
	allMetrics, err := s.GetMetrics(ctx)
	if err != nil {
		return nil, err
	}
	
	// Limit to first 100 metrics to avoid too many queries
	if len(allMetrics) > 100 {
		allMetrics = allMetrics[:100]
	}
	
	// Query cardinality for each metric
	now := time.Now()
	topMetrics := make([]models.TopMetric, 0, len(allMetrics))
	
	for _, metricName := range allMetrics {
		cardinalityQuery := fmt.Sprintf("count(%s)", metricName)
		results, err := s.client.Query(ctx, cardinalityQuery, now)
		if err != nil {
			s.logger.Warnf("Failed to get cardinality for %s: %v", metricName, err)
			continue
		}
		
		var cardinality int
		if len(results) > 0 {
			cardinality = int(results[0].Value)
		}
		
		// Also get sample rate (changes per minute) if possible
		rateQuery := fmt.Sprintf("rate(%s[5m])", metricName)
		rateResults, err := s.client.Query(ctx, rateQuery, now)
		if err != nil {
			// Rate queries might fail for some metrics, just log and continue
			s.logger.Debugf("Rate query failed for %s: %v", metricName, err)
		}
		
		var sampleRate float64
		if len(rateResults) > 0 {
			sampleRate = rateResults[0].Value
		}
		
		topMetrics = append(topMetrics, models.TopMetric{
			Name:        metricName,
			Cardinality: cardinality,
			SampleRate:  sampleRate,
		})
	}
	
	// Sort by cardinality (highest first)
	sort.Slice(topMetrics, func(i, j int) bool {
		return topMetrics[i].Cardinality > topMetrics[j].Cardinality
	})
	
	// Limit results
	if len(topMetrics) > limit {
		topMetrics = topMetrics[:limit]
	}
	
	return topMetrics, nil
}

// GetMetricHealth provides health information about a specific metric
func (s *MetricsService) GetMetricHealth(ctx context.Context, metricName string) (*models.MetricHealth, error) {
	now := time.Now()
	
	// Check if metric exists
	query := fmt.Sprintf("count(%s)", metricName)
	results, err := s.client.Query(ctx, query, now)
	if err != nil {
		s.logger.Errorf("Failed to query metric existence for %s: %v", metricName, err)
		return nil, fmt.Errorf("failed to query metric existence: %w", err)
	}
	
	// Check if the metric is being scraped
	exists := len(results) > 0 && results[0].Value > 0
	
	// Get last scrape time
	scrapeTimeQuery := "scrape_time_seconds{instance=~\".+\", job=~\".+\"} > 0"
	scrapeResults, err := s.client.Query(ctx, scrapeTimeQuery, now)
	if err != nil {
		s.logger.Warnf("Failed to query scrape time: %v", err)
		// Continue anyway as this is not critical
	}
	
	var lastScraped time.Time
	if len(scrapeResults) > 0 {
		lastScraped = scrapeResults[0].Timestamp
	}
	
	// Check staleness
	staleThreshold := now.Add(-5 * time.Minute)
	isStale := lastScraped.Before(staleThreshold)
	
	// Check for gaps in data (if there are no samples in the last 5 minutes)
	gapQuery := fmt.Sprintf("count_over_time(%s[5m]) > 0", metricName)
	gapResults, err := s.client.Query(ctx, gapQuery, now)
	if err != nil {
		s.logger.Warnf("Failed to query for gaps: %v", err)
	}
	
	hasGaps := len(gapResults) == 0 || gapResults[0].Value == 0
	
	health := &models.MetricHealth{
		Name:        metricName,
		Exists:      exists,
		IsStale:     isStale,
		HasGaps:     hasGaps,
		LastScraped: lastScraped,
		CheckedAt:   now,
	}
	
	return health, nil
}