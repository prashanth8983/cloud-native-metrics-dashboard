package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDefaultConfig tests that the default configuration is loaded correctly when no env vars are set
func TestDefaultConfig(t *testing.T) {
	// Clear environment variables that might affect the test
	clearEnvironmentVars()

	// Load configuration
	config, err := Load()
	require.NoError(t, err, "Load() should not return an error with default values")
	require.NotNil(t, config, "Load() should return a non-nil config")

	// Check server defaults
	assert.Equal(t, 8080, config.Server.Port, "Default server port should be 8080")
	assert.Equal(t, 5, config.Server.ReadTimeoutSeconds, "Default read timeout should be 5 seconds")
	assert.Equal(t, 10, config.Server.WriteTimeoutSeconds, "Default write timeout should be 10 seconds")
	assert.Equal(t, 120, config.Server.IdleTimeoutSeconds, "Default idle timeout should be 120 seconds")

	// Check Prometheus defaults
	assert.Equal(t, "http://prometheus:9090", config.Prometheus.URL, "Default Prometheus URL should be http://prometheus:9090")
	assert.Equal(t, 30, config.Prometheus.TimeoutSeconds, "Default Prometheus timeout should be 30 seconds")
	assert.Equal(t, 11000, config.Prometheus.MaxQueryPoints, "Default max query points should be 11000")

	// Check logging defaults
	assert.Equal(t, "info", config.Logging.Level, "Default log level should be info")
	assert.Equal(t, "json", config.Logging.Format, "Default log format should be json")

	// Check cache defaults
	assert.Equal(t, true, config.Cache.Enabled, "Cache should be enabled by default")
	assert.Equal(t, 60, config.Cache.TTLSeconds, "Default cache TTL should be 60 seconds")
	assert.Equal(t, 1000, config.Cache.MaxSizeItems, "Default cache max size should be 1000 items")
}

// TestEnvironmentOverrides tests that environment variables correctly override defaults
func TestEnvironmentOverrides(t *testing.T) {
	// Clear environment variables first
	clearEnvironmentVars()

	// Set environment variables for testing
	os.Setenv("SERVER_PORT", "9000")
	os.Setenv("SERVER_READ_TIMEOUT", "10")
	os.Setenv("SERVER_WRITE_TIMEOUT", "20")
	os.Setenv("SERVER_IDLE_TIMEOUT", "180")
	os.Setenv("PROMETHEUS_URL", "http://prom.example.com:9090")
	os.Setenv("PROMETHEUS_TIMEOUT", "45")
	os.Setenv("PROMETHEUS_MAX_QUERY_POINTS", "5000")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_FORMAT", "console")
	os.Setenv("CACHE_ENABLED", "false")
	os.Setenv("CACHE_TTL", "120")
	os.Setenv("CACHE_MAX_SIZE", "2000")

	// Cleanup environment after test
	defer clearEnvironmentVars()

	// Load configuration
	config, err := Load()
	require.NoError(t, err, "Load() should not return an error with valid environment variables")
	require.NotNil(t, config, "Load() should return a non-nil config")

	// Check server settings
	assert.Equal(t, 9000, config.Server.Port, "Server port should be overridden by environment")
	assert.Equal(t, 10, config.Server.ReadTimeoutSeconds, "Read timeout should be overridden by environment")
	assert.Equal(t, 20, config.Server.WriteTimeoutSeconds, "Write timeout should be overridden by environment")
	assert.Equal(t, 180, config.Server.IdleTimeoutSeconds, "Idle timeout should be overridden by environment")

	// Check Prometheus settings
	assert.Equal(t, "http://prom.example.com:9090", config.Prometheus.URL, "Prometheus URL should be overridden by environment")
	assert.Equal(t, 45, config.Prometheus.TimeoutSeconds, "Prometheus timeout should be overridden by environment")
	assert.Equal(t, 5000, config.Prometheus.MaxQueryPoints, "Max query points should be overridden by environment")

	// Check logging settings
	assert.Equal(t, "debug", config.Logging.Level, "Log level should be overridden by environment")
	assert.Equal(t, "console", config.Logging.Format, "Log format should be overridden by environment")

	// Check cache settings
	assert.Equal(t, false, config.Cache.Enabled, "Cache enabled should be overridden by environment")
	assert.Equal(t, 120, config.Cache.TTLSeconds, "Cache TTL should be overridden by environment")
	assert.Equal(t, 2000, config.Cache.MaxSizeItems, "Cache max size should be overridden by environment")
}

// TestInvalidConfig tests validation of the configuration
func TestInvalidConfig(t *testing.T) {
	// Test invalid server port
	clearEnvironmentVars()
	os.Setenv("SERVER_PORT", "-1")
	config, err := Load()
	assert.Error(t, err, "Load() should return an error with invalid server port")
	assert.Nil(t, config, "Config should be nil when validation fails")

	// Test empty Prometheus URL
	clearEnvironmentVars()
	os.Setenv("PROMETHEUS_URL", "")
	config, err = Load()
	assert.Error(t, err, "Load() should return an error with empty Prometheus URL")
	assert.Nil(t, config, "Config should be nil when validation fails")

	// Test invalid Prometheus timeout
	clearEnvironmentVars()
	os.Setenv("PROMETHEUS_TIMEOUT", "0")
	config, err = Load()
	assert.Error(t, err, "Load() should return an error with invalid Prometheus timeout")
	assert.Nil(t, config, "Config should be nil when validation fails")
}

// TestNonNumericEnvVars tests handling of non-numeric values in numeric environment variables
func TestNonNumericEnvVars(t *testing.T) {
	// Clear environment variables first
	clearEnvironmentVars()

	// Set non-numeric environment variables
	os.Setenv("SERVER_PORT", "not-a-number")
	os.Setenv("CACHE_TTL", "invalid")

	// Load configuration
	config, err := Load()
	require.NoError(t, err, "Load() should not return an error for non-numeric values (should use defaults)")
	require.NotNil(t, config, "Load() should return a non-nil config")

	// Check that default values are used
	assert.Equal(t, 8080, config.Server.Port, "Server port should use default for invalid value")
	assert.Equal(t, 60, config.Cache.TTLSeconds, "Cache TTL should use default for invalid value")
}

// TestGetPrometheusTimeout tests the GetPrometheusTimeout method
func TestGetPrometheusTimeout(t *testing.T) {
	// Create a config with known timeout
	promConfig := PrometheusConfig{
		TimeoutSeconds: 45,
	}

	// Test the method
	timeout := promConfig.GetPrometheusTimeout()
	assert.Equal(t, 45*time.Second, timeout, "GetPrometheusTimeout should return the correct duration")
}

// TestGetCacheTTL tests the GetCacheTTL method
func TestGetCacheTTL(t *testing.T) {
	// Create a config with known TTL
	cacheConfig := CacheConfig{
		TTLSeconds: 120,
	}

	// Test the method
	ttl := cacheConfig.GetCacheTTL()
	assert.Equal(t, 120*time.Second, ttl, "GetCacheTTL should return the correct duration")
}

// TestBooleanEnvVars tests handling of boolean environment variables
func TestBooleanEnvVars(t *testing.T) {
	testCases := []struct {
		value    string
		expected bool
	}{
		{"true", true},
		{"True", true},
		{"TRUE", true},
		{"1", true},
		{"false", false},
		{"False", false},
		{"FALSE", false},
		{"0", false},
		{"invalid", false}, // Invalid values should default to false
	}

	for _, tc := range testCases {
		clearEnvironmentVars()
		os.Setenv("CACHE_ENABLED", tc.value)

		config, err := Load()
		require.NoError(t, err)
		assert.Equal(t, tc.expected, config.Cache.Enabled, "CACHE_ENABLED=%s should result in %v", tc.value, tc.expected)
	}
}

// TestOverrideSubset tests that setting only a subset of environment variables works correctly
func TestOverrideSubset(t *testing.T) {
	// Clear environment variables first
	clearEnvironmentVars()

	// Set only a subset of environment variables
	os.Setenv("SERVER_PORT", "9000")
	os.Setenv("LOG_LEVEL", "debug")

	// Load configuration
	config, err := Load()
	require.NoError(t, err)
	require.NotNil(t, config)

	// Check overridden values
	assert.Equal(t, 9000, config.Server.Port, "Server port should be overridden")
	assert.Equal(t, "debug", config.Logging.Level, "Log level should be overridden")

	// Check that other values are still defaults
	assert.Equal(t, 5, config.Server.ReadTimeoutSeconds, "Read timeout should use default")
	assert.Equal(t, "http://prometheus:9090", config.Prometheus.URL, "Prometheus URL should use default")
	assert.Equal(t, true, config.Cache.Enabled, "Cache enabled should use default")
}

// Helper function to clear all environment variables used by the config
func clearEnvironmentVars() {
	// Server config
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("SERVER_READ_TIMEOUT")
	os.Unsetenv("SERVER_WRITE_TIMEOUT")
	os.Unsetenv("SERVER_IDLE_TIMEOUT")

	// Prometheus config
	os.Unsetenv("PROMETHEUS_URL")
	os.Unsetenv("PROMETHEUS_TIMEOUT")
	os.Unsetenv("PROMETHEUS_MAX_QUERY_POINTS")

	// Logging config
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("LOG_FORMAT")

	// Cache config
	os.Unsetenv("CACHE_ENABLED")
	os.Unsetenv("CACHE_TTL")
	os.Unsetenv("CACHE_MAX_SIZE")
}

// TestDotEnvLoading tests loading configuration from a .env file
func TestDotEnvLoading(t *testing.T) {
	// Skip this test if .env file doesn't exist to avoid failing in CI environments
	if _, err := os.Stat(".env.test"); os.IsNotExist(err) {
		t.Skip(".env.test file not found, skipping test")
	}

	// Create a temporary .env file for testing
	envContent := `
SERVER_PORT=7000
PROMETHEUS_URL=http://prometheus-test:9090
LOG_LEVEL=debug
`
	err := os.WriteFile(".env.test", []byte(envContent), 0644)
	require.NoError(t, err, "Failed to create test .env file")
	defer os.Remove(".env.test") // Clean up after test

	// Set the env file name for godotenv to pick up
	os.Setenv("ENV_FILE", ".env.test")
	defer os.Unsetenv("ENV_FILE")

	// Clear other environment variables
	clearEnvironmentVars()

	// This part depends on your Load() implementation
	// If it looks for a specific env file name, you might need to mock or modify it
	// For this test, we assume Load() can be configured to use .env.test

	// Load configuration
	config, err := Load()
	require.NoError(t, err)
	require.NotNil(t, config)

	// Check that values from .env file are used
	assert.Equal(t, 7000, config.Server.Port, "Server port should be loaded from .env file")
	assert.Equal(t, "http://prometheus-test:9090", config.Prometheus.URL, "Prometheus URL should be loaded from .env file")
	assert.Equal(t, "debug", config.Logging.Level, "Log level should be loaded from .env file")
}