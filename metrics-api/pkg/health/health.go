package health

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// Status represents the health status of the service
type Status string

const (
	// StatusUp indicates the service is healthy
	StatusUp Status = "up"
	// StatusDegraded indicates the service is working but with issues
	StatusDegraded Status = "degraded"
	// StatusDown indicates the service is not functioning
	StatusDown Status = "down"
)

// Check represents a health check function
type Check func(ctx context.Context) (Status, map[string]interface{}, error)

// CheckResult represents the result of a health check
type CheckResult struct {
	Status    Status
	Details   map[string]interface{}
	Error     error
	Timestamp time.Time
	Duration  time.Duration
}

// Checker manages health checks for the application
type Checker struct {
	checks       map[string]Check
	mu           sync.RWMutex
	startTime    time.Time
	lastResults  map[string]CheckResult
	resultsMu    sync.RWMutex
	checkTimeout time.Duration
}

// NewChecker creates a new health checker
func NewChecker(checkTimeout time.Duration) *Checker {
	return &Checker{
		checks:       make(map[string]Check),
		startTime:    time.Now(),
		lastResults:  make(map[string]CheckResult),
		checkTimeout: checkTimeout,
	}
}

// AddCheck adds a named health check
func (c *Checker) AddCheck(name string, check Check) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.checks[name] = check
}

// RemoveCheck removes a named health check
func (c *Checker) RemoveCheck(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.checks, name)
}

// RunCheck runs a specific health check
func (c *Checker) RunCheck(ctx context.Context, name string) (CheckResult, bool) {
	c.mu.RLock()
	check, exists := c.checks[name]
	c.mu.RUnlock()

	if !exists {
		return CheckResult{}, false
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, c.checkTimeout)
	defer cancel()

	startTime := time.Now()
	status, details, err := check(ctx)
	duration := time.Since(startTime)

	result := CheckResult{
		Status:    status,
		Details:   details,
		Error:     err,
		Timestamp: time.Now(),
		Duration:  duration,
	}

	// Store the result
	c.resultsMu.Lock()
	c.lastResults[name] = result
	c.resultsMu.Unlock()

	return result, true
}

// RunChecks runs all health checks and returns overall status
func (c *Checker) RunChecks(ctx context.Context) (Status, map[string]CheckResult) {
	results := make(map[string]CheckResult)
	
	// Create a channel to collect results from goroutines
	resultChan := make(chan struct {
		name   string
		result CheckResult
	})

	// Run each check in its own goroutine
	c.mu.RLock()
	checkCount := len(c.checks)
	for name, check := range c.checks {
		go func(name string, check Check) {
			// Create timeout context for this check
			checkCtx, cancel := context.WithTimeout(ctx, c.checkTimeout)
			defer cancel()

			startTime := time.Now()
			status, details, err := check(checkCtx)
			duration := time.Since(startTime)

			result := CheckResult{
				Status:    status,
				Details:   details,
				Error:     err,
				Timestamp: time.Now(),
				Duration:  duration,
			}

			resultChan <- struct {
				name   string
				result CheckResult
			}{name, result}
		}(name, check)
	}
	c.mu.RUnlock()

	// Collect all results with a timeout
	timeout := time.After(c.checkTimeout + 100*time.Millisecond) // Add a little extra time
	for i := 0; i < checkCount; i++ {
		select {
		case result := <-resultChan:
			results[result.name] = result.result
		case <-timeout:
			// If we time out waiting for checks, consider any missing checks as failures
			c.mu.RLock()
			for name := range c.checks {
				if _, exists := results[name]; !exists {
					results[name] = CheckResult{
						Status:    StatusDown,
						Error:     fmt.Errorf("health check timed out"),
						Timestamp: time.Now(),
					}
				}
			}
			c.mu.RUnlock()
			i = checkCount // Exit the loop
		}
	}

	// Store results
	c.resultsMu.Lock()
	for name, result := range results {
		c.lastResults[name] = result
	}
	c.resultsMu.Unlock()

	// Determine overall status
	return c.determineOverallStatus(results), results
}

// GetLastResults gets the last known results for all checks
func (c *Checker) GetLastResults() map[string]CheckResult {
	c.resultsMu.RLock()
	defer c.resultsMu.RUnlock()

	// Create a copy to avoid race conditions
	results := make(map[string]CheckResult, len(c.lastResults))
	for name, result := range c.lastResults {
		results[name] = result
	}

	return results
}

// GetUptime returns the service uptime
func (c *Checker) GetUptime() time.Duration {
	return time.Since(c.startTime)
}

// determineOverallStatus determines the overall status based on individual check results
func (c *Checker) determineOverallStatus(results map[string]CheckResult) Status {
	if len(results) == 0 {
		return StatusDown
	}

	hasDown := false
	hasDegraded := false

	for _, result := range results {
		if result.Status == StatusDown {
			hasDown = true
		} else if result.Status == StatusDegraded {
			hasDegraded = true
		}
	}

	if hasDown {
		return StatusDown
	} else if hasDegraded {
		return StatusDegraded
	}
	return StatusUp
}

// SystemInfo returns information about the system
func SystemInfo() map[string]interface{} {
	info := make(map[string]interface{})

	// Add Go runtime information
	info["go_version"] = runtime.Version()
	info["go_os"] = runtime.GOOS
	info["go_arch"] = runtime.GOARCH

	// Add memory statistics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	info["memory_allocated"] = memStats.Alloc
	info["memory_total_allocated"] = memStats.TotalAlloc
	info["memory_system"] = memStats.Sys
	info["memory_num_gc"] = memStats.NumGC

	// Add goroutine count
	info["goroutines"] = runtime.NumGoroutine()

	return info
}

// CommonChecks contains health check implementations for common components

// PrometheusCheck creates a health check for Prometheus connectivity
func PrometheusCheck(queryFunc func(ctx context.Context, query string) error) Check {
	return func(ctx context.Context) (Status, map[string]interface{}, error) {
		details := make(map[string]interface{})
		startTime := time.Now()

		// Simple "up" query to check Prometheus
		err := queryFunc(ctx, "up")

		// Calculate response time
		responseTime := time.Since(startTime)
		details["response_time_ms"] = responseTime.Milliseconds()

		if err != nil {
			details["error"] = err.Error()
			return StatusDown, details, err
		}

		return StatusUp, details, nil
	}
}

// DatabaseCheck creates a health check for database connectivity
func DatabaseCheck(pingFunc func(ctx context.Context) error) Check {
	return func(ctx context.Context) (Status, map[string]interface{}, error) {
		details := make(map[string]interface{})
		startTime := time.Now()

		// Ping the database
		err := pingFunc(ctx)

		// Calculate response time
		responseTime := time.Since(startTime)
		details["response_time_ms"] = responseTime.Milliseconds()

		if err != nil {
			details["error"] = err.Error()
			return StatusDown, details, err
		}

		return StatusUp, details, nil
	}
}

// DependencyCheck creates a health check for generic dependency
func DependencyCheck(name string, checkFunc func(ctx context.Context) error) Check {
	return func(ctx context.Context) (Status, map[string]interface{}, error) {
		details := make(map[string]interface{})
		details["dependency"] = name
		startTime := time.Now()

		// Check the dependency
		err := checkFunc(ctx)

		// Calculate response time
		responseTime := time.Since(startTime)
		details["response_time_ms"] = responseTime.Milliseconds()

		if err != nil {
			details["error"] = err.Error()
			return StatusDown, details, err
		}

		return StatusUp, details, nil
	}
}

// MemoryCheck creates a health check for memory usage
func MemoryCheck(threshold float64) Check {
	return func(ctx context.Context) (Status, map[string]interface{}, error) {
		details := make(map[string]interface{})

		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		// Calculate memory usage
		memoryUsed := float64(memStats.Alloc)
		memorySystem := float64(memStats.Sys)
		memoryUsage := memoryUsed / memorySystem

		details["memory_used_bytes"] = memStats.Alloc
		details["memory_system_bytes"] = memStats.Sys
		details["memory_usage_ratio"] = memoryUsage

		if memoryUsage > threshold {
			return StatusDegraded, details, fmt.Errorf("memory usage too high: %.2f%%", memoryUsage*100)
		}

		return StatusUp, details, nil
	}
}

// CPUCheck creates a health check for CPU usage
// Note: This is a simple version, a more accurate check would involve sampling over time
func CPUCheck(numGoroutinesThreshold int) Check {
	return func(ctx context.Context) (Status, map[string]interface{}, error) {
		details := make(map[string]interface{})

		numGoroutines := runtime.NumGoroutine()
		numCPU := runtime.NumCPU()

		details["goroutines"] = numGoroutines
		details["cpu_cores"] = numCPU
		details["goroutines_per_core"] = float64(numGoroutines) / float64(numCPU)

		if numGoroutines > numGoroutinesThreshold {
			return StatusDegraded, details, fmt.Errorf("high number of goroutines: %d", numGoroutines)
		}

		return StatusUp, details, nil
	}
}

// HealthStatus represents a complete health status report
type HealthStatus struct {
	Status       Status                      `json:"status"`
	Version      string                      `json:"version"`
	Uptime       string                      `json:"uptime"`
	Timestamp    time.Time                   `json:"timestamp"`
	Checks       map[string]string           `json:"checks"`
	CheckDetails map[string]CheckResult      `json:"check_details,omitempty"`
	SystemInfo   map[string]interface{}      `json:"system_info,omitempty"`
}

// GenerateHealthStatus creates a consolidated health status report
func GenerateHealthStatus(checker *Checker, version string, includeDetails bool) HealthStatus {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Run all health checks
	status, results := checker.RunChecks(ctx)

	// Format checks for simple status display
	checks := make(map[string]string)
	for name, result := range results {
		checks[name] = string(result.Status)
	}

	// Create health status report
	healthStatus := HealthStatus{
		Status:    status,
		Version:   version,
		Uptime:    formatDuration(checker.GetUptime()),
		Timestamp: time.Now(),
		Checks:    checks,
	}

	// Include details if requested
	if includeDetails {
		healthStatus.CheckDetails = results
		healthStatus.SystemInfo = SystemInfo()
	}

	return healthStatus
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}