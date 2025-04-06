// internal/config/config_test.go
package config

import (
	"os"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()

	// Test default values
	if cfg.Server.Port != 8000 {
		t.Errorf("Expected default server port to be 8000, got %d", cfg.Server.Port)
	}

	if cfg.Prometheus.URL != "http://prometheus:9090" {
		t.Errorf("Expected default prometheus URL to be http://prometheus:9090, got %s", cfg.Prometheus.URL)
	}

	if cfg.CORS.Enabled != true {
		t.Errorf("Expected default CORS enabled to be true, got %v", cfg.CORS.Enabled)
	}

	if len(cfg.CORS.AllowedOrigins) != 1 || cfg.CORS.AllowedOrigins[0] != "*" {
		t.Errorf("Expected default CORS allowed origins to be [*], got %v", cfg.CORS.AllowedOrigins)
	}

	if cfg.Cache.TTL != 60*time.Second {
		t.Errorf("Expected default cache TTL to be 60 seconds, got %v", cfg.Cache.TTL)
	}

	if cfg.Log.Level != "info" {
		t.Errorf("Expected default log level to be info, got %s", cfg.Log.Level)
	}
}

func TestEnvironmentOverrides(t *testing.T) {
	// Save current environment to restore later
	envBackup := make(map[string]string)
	for _, key := range []string{
		"API_SERVER_PORT", "PROMETHEUS_URL", "CORS_ENABLED",
		"CACHE_TTL", "LOG_LEVEL",
	} {
		if val, exists := os.LookupEnv(key); exists {
			envBackup[key] = val
		}
	}

	// Cleanup function to restore environment
	defer func() {
		for key := range envBackup {
			os.Unsetenv(key)
		}
		for key, val := range envBackup {
			os.Setenv(key, val)
		}
	}()

	// Set test environment variables
	os.Setenv("API_SERVER_PORT", "9000")
	os.Setenv("PROMETHEUS_URL", "http://test-prometheus:9090")
	os.Setenv("CORS_ENABLED", "false")
	os.Setenv("CACHE_TTL", "120s")
	os.Setenv("LOG_LEVEL", "debug")

	// Create config with environment variables
	cfg := defaultConfig()
	applyEnvironment(cfg)

	// Verify environment overrides
	if cfg.Server.Port != 9000 {
		t.Errorf("Expected server port to be overridden to 9000, got %d", cfg.Server.Port)
	}

	if cfg.Prometheus.URL != "http://test-prometheus:9090" {
		t.Errorf("Expected prometheus URL to be overridden to http://test-prometheus:9090, got %s", cfg.Prometheus.URL)
	}

	if cfg.CORS.Enabled != false {
		t.Errorf("Expected CORS enabled to be overridden to false, got %v", cfg.CORS.Enabled)
	}

	if cfg.Cache.TTL != 120*time.Second {
		t.Errorf("Expected cache TTL to be overridden to 120 seconds, got %v", cfg.Cache.TTL)
	}

	if cfg.Log.Level != "debug" {
		t.Errorf("Expected log level to be overridden to debug, got %s", cfg.Log.Level)
	}
}

func TestSplitAndTrim(t *testing.T) {
	tests := []struct {
		input    string
		sep      string
		expected []string
	}{
		{"a, b,c ", ",", []string{"a", "b", "c"}},
		{"", ",", []string{}},
		{"  ", ",", []string{}},
		{"a", ",", []string{"a"}},
		{"a,,b", ",", []string{"a", "b"}},
		{"a|b|c", "|", []string{"a", "b", "c"}},
	}

	for _, test := range tests {
		result := splitAndTrim(test.input, test.sep)
		
		if len(result) != len(test.expected) {
			t.Errorf("For input '%s': expected length %d, got %d", test.input, len(test.expected), len(result))
			continue
		}
		
		for i, v := range result {
			if v != test.expected[i] {
				t.Errorf("For input '%s': expected result[%d] to be '%s', got '%s'", test.input, i, test.expected[i], v)
			}
		}
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name           string
		modifyConfig   func(*Config)
		expectedErrors int
	}{
		{
			name:           "Valid config",
			modifyConfig:   func(cfg *Config) {},
			expectedErrors: 0,
		},
		{
			name: "Invalid server port",
			modifyConfig: func(cfg *Config) {
				cfg.Server.Port = -1
			},
			expectedErrors: 1,
		},
		{
			name: "Empty prometheus URL",
			modifyConfig: func(cfg *Config) {
				cfg.Prometheus.URL = ""
			},
			expectedErrors: 1,
		},
		{
			name: "Invalid prometheus timeout",
			modifyConfig: func(cfg *Config) {
				cfg.Prometheus.Timeout = 0
			},
			expectedErrors: 1,
		},
		{
			name: "CORS enabled but no methods",
			modifyConfig: func(cfg *Config) {
				cfg.CORS.Enabled = true
				cfg.CORS.AllowedMethods = []string{}
			},
			expectedErrors: 1,
		},
		{
			name: "Cache enabled but invalid TTL",
			modifyConfig: func(cfg *Config) {
				cfg.Cache.Enabled = true
				cfg.Cache.TTL = 0
			},
			expectedErrors: 1,
		},
		{
			name: "Invalid log level",
			modifyConfig: func(cfg *Config) {
				cfg.Log.Level = "invalid"
			},
			expectedErrors: 1,
		},
		{
			name: "Invalid log format",
			modifyConfig: func(cfg *Config) {
				cfg.Log.Format = "invalid"
			},
			expectedErrors: 1,
		},
		{
			name: "Metrics enabled but empty path",
			modifyConfig: func(cfg *Config) {
				cfg.Metrics.Enabled = true
				cfg.Metrics.Path = ""
			},
			expectedErrors: 1,
		},
		{
			name: "Multiple errors",
			modifyConfig: func(cfg *Config) {
				cfg.Server.Port = -1
				cfg.Prometheus.URL = ""
				cfg.Log.Level = "invalid"
			},
			expectedErrors: 3,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := defaultConfig()
			test.modifyConfig(cfg)
			
			errors := cfg.Validate()
			
			if len(errors) != test.expectedErrors {
				t.Errorf("Expected %d validation errors, got %d: %v", 
					test.expectedErrors, len(errors), errors)
			}
		})
	}
}