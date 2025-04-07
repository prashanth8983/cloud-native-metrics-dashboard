package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Server     ServerConfig
	Prometheus PrometheusConfig
	Logging    LoggingConfig
	Cache      CacheConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port                int
	ReadTimeoutSeconds  int
	WriteTimeoutSeconds int
	IdleTimeoutSeconds  int
}

// PrometheusConfig holds Prometheus client configuration
type PrometheusConfig struct {
	URL           string
	TimeoutSeconds int
	MaxQueryPoints int
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	Enabled     bool
	TTLSeconds  int
	MaxSizeItems int
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()
	
	config := &Config{
		Server: ServerConfig{
			Port:                getEnvAsInt("SERVER_PORT", 8080),
			ReadTimeoutSeconds:  getEnvAsInt("SERVER_READ_TIMEOUT", 5),
			WriteTimeoutSeconds: getEnvAsInt("SERVER_WRITE_TIMEOUT", 10),
			IdleTimeoutSeconds:  getEnvAsInt("SERVER_IDLE_TIMEOUT", 120),
		},
		Prometheus: PrometheusConfig{
			URL:            getEnv("PROMETHEUS_URL", "http://prometheus:9090"),
			TimeoutSeconds: getEnvAsInt("PROMETHEUS_TIMEOUT", 30),
			MaxQueryPoints: getEnvAsInt("PROMETHEUS_MAX_QUERY_POINTS", 11000),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		Cache: CacheConfig{
			Enabled:     getEnvAsBool("CACHE_ENABLED", true),
			TTLSeconds:  getEnvAsInt("CACHE_TTL", 60),
			MaxSizeItems: getEnvAsInt("CACHE_MAX_SIZE", 1000),
		},
	}
	
	return config, validateConfig(config)
}

// validateConfig validates the configuration
func validateConfig(cfg *Config) error {
	if cfg.Server.Port <= 0 {
		return fmt.Errorf("server port must be positive")
	}
	
	if cfg.Prometheus.URL == "" {
		return fmt.Errorf("prometheus URL cannot be empty")
	}
	
	if cfg.Prometheus.TimeoutSeconds <= 0 {
		return fmt.Errorf("prometheus timeout must be positive")
	}
	
	return nil
}

// getEnv gets an environment variable or returns a default
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as an integer or returns a default
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// getEnvAsBool gets an environment variable as a boolean or returns a default
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// GetPrometheusTimeout returns the Prometheus timeout as a duration
func (c *PrometheusConfig) GetPrometheusTimeout() time.Duration {
	return time.Duration(c.TimeoutSeconds) * time.Second
}

// GetCacheTTL returns the cache TTL as a duration
func (c *CacheConfig) GetCacheTTL() time.Duration {
	return time.Duration(c.TTLSeconds) * time.Second
}