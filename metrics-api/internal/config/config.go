package config

import (
	"flag"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all the application configuration
type Config struct {
	// Server settings
	Server struct {
		Port         int           `json:"port"`
		Host         string        `json:"host"`
		ReadTimeout  time.Duration `json:"read_timeout"`
		WriteTimeout time.Duration `json:"write_timeout"`
		IdleTimeout  time.Duration `json:"idle_timeout"`
	} `json:"server"`

	// Prometheus settings
	Prometheus struct {
		URL            string        `json:"url"`
		Timeout        time.Duration `json:"timeout"`
		KeepAlive      time.Duration `json:"keep_alive"`
		MaxConnections int           `json:"max_connections"`
	} `json:"prometheus"`

	// CORS settings
	CORS struct {
		Enabled        bool     `json:"enabled"`
		AllowedOrigins []string `json:"allowed_origins"`
		AllowedMethods []string `json:"allowed_methods"`
		AllowedHeaders []string `json:"allowed_headers"`
		MaxAge         int      `json:"max_age"`
	} `json:"cors"`

	// Cache settings
	Cache struct {
		Enabled       bool          `json:"enabled"`
		TTL           time.Duration `json:"ttl"`
		CleanupPeriod time.Duration `json:"cleanup_period"`
		MaxItems      int           `json:"max_items"`
	} `json:"cache"`

	// Logging settings
	Log struct {
		Level       string `json:"level"`
		Format      string `json:"format"`
		FilePath    string `json:"file_path"`
		MaxSize     int    `json:"max_size"`     // in MB
		MaxBackups  int    `json:"max_backups"`  // number of files
		MaxAge      int    `json:"max_age"`      // in days
		Compression bool   `json:"compression"`  // compress rotated files
	} `json:"log"`

	// Metrics settings
	Metrics struct {
		Enabled bool   `json:"enabled"`
		Path    string `json:"path"`
	} `json:"metrics"`
}

// New creates a new configuration with defaults, environment variables, and flags
func New() *Config {
	cfg := defaultConfig()
	
	// Apply environment variables to override defaults
	applyEnvironment(cfg)
	
	// Apply command line flags to override environment and defaults
	applyFlags(cfg)
	
	return cfg
}

// defaultConfig creates a configuration with default values
func defaultConfig() *Config {
	cfg := &Config{}
	
	// Server defaults
	cfg.Server.Port = 8000
	cfg.Server.Host = ""
	cfg.Server.ReadTimeout = 10 * time.Second
	cfg.Server.WriteTimeout = 20 * time.Second
	cfg.Server.IdleTimeout = 120 * time.Second
	
	// Prometheus defaults
	cfg.Prometheus.URL = "http://prometheus:9090"
	cfg.Prometheus.Timeout = 10 * time.Second
	cfg.Prometheus.KeepAlive = 30 * time.Second
	cfg.Prometheus.MaxConnections = 100
	
	// CORS defaults
	cfg.CORS.Enabled = true
	cfg.CORS.AllowedOrigins = []string{"*"}
	cfg.CORS.AllowedMethods = []string{"GET", "POST", "OPTIONS"}
	cfg.CORS.AllowedHeaders = []string{"Content-Type", "Authorization"}
	cfg.CORS.MaxAge = 86400
	
	// Cache defaults
	cfg.Cache.Enabled = true
	cfg.Cache.TTL = 60 * time.Second
	cfg.Cache.CleanupPeriod = 5 * time.Minute
	cfg.Cache.MaxItems = 1000
	
	// Logging defaults
	cfg.Log.Level = "info"
	cfg.Log.Format = "text"
	cfg.Log.FilePath = ""
	cfg.Log.MaxSize = 100
	cfg.Log.MaxBackups = 5
	cfg.Log.MaxAge = 30
	cfg.Log.Compression = true
	
	// Metrics defaults
	cfg.Metrics.Enabled = true
	cfg.Metrics.Path = "/metrics"
	
	return cfg
}

// applyEnvironment applies environment variables to the configuration
func applyEnvironment(cfg *Config) {
	// Server settings
	if val, exists := os.LookupEnv("API_SERVER_PORT"); exists {
		if port, err := strconv.Atoi(val); err == nil && port > 0 {
			cfg.Server.Port = port
		}
	}
	if val, exists := os.LookupEnv("API_SERVER_HOST"); exists {
		cfg.Server.Host = val
	}
	if val, exists := os.LookupEnv("API_READ_TIMEOUT"); exists {
		if d, err := time.ParseDuration(val); err == nil {
			cfg.Server.ReadTimeout = d
		}
	}
	if val, exists := os.LookupEnv("API_WRITE_TIMEOUT"); exists {
		if d, err := time.ParseDuration(val); err == nil {
			cfg.Server.WriteTimeout = d
		}
	}
	if val, exists := os.LookupEnv("API_IDLE_TIMEOUT"); exists {
		if d, err := time.ParseDuration(val); err == nil {
			cfg.Server.IdleTimeout = d
		}
	}

	// Prometheus settings
	if val, exists := os.LookupEnv("PROMETHEUS_URL"); exists {
		cfg.Prometheus.URL = val
	}
	if val, exists := os.LookupEnv("PROMETHEUS_TIMEOUT"); exists {
		if d, err := time.ParseDuration(val); err == nil {
			cfg.Prometheus.Timeout = d
		} else if seconds, err := strconv.Atoi(val); err == nil {
			cfg.Prometheus.Timeout = time.Duration(seconds) * time.Second
		}
	}
	if val, exists := os.LookupEnv("PROMETHEUS_KEEP_ALIVE"); exists {
		if d, err := time.ParseDuration(val); err == nil {
			cfg.Prometheus.KeepAlive = d
		}
	}
	if val, exists := os.LookupEnv("PROMETHEUS_MAX_CONNECTIONS"); exists {
		if n, err := strconv.Atoi(val); err == nil {
			cfg.Prometheus.MaxConnections = n
		}
	}

	// CORS settings
	if val, exists := os.LookupEnv("CORS_ENABLED"); exists {
		cfg.CORS.Enabled = strings.ToLower(val) == "true"
	}
	if val, exists := os.LookupEnv("CORS_ALLOWED_ORIGINS"); exists {
		cfg.CORS.AllowedOrigins = splitAndTrim(val, ",")
	}
	if val, exists := os.LookupEnv("CORS_ALLOWED_METHODS"); exists {
		cfg.CORS.AllowedMethods = splitAndTrim(val, ",")
	}
	if val, exists := os.LookupEnv("CORS_ALLOWED_HEADERS"); exists {
		cfg.CORS.AllowedHeaders = splitAndTrim(val, ",")
	}
	if val, exists := os.LookupEnv("CORS_MAX_AGE"); exists {
		if n, err := strconv.Atoi(val); err == nil {
			cfg.CORS.MaxAge = n
		}
	}

	// Cache settings
	if val, exists := os.LookupEnv("CACHE_ENABLED"); exists {
		cfg.Cache.Enabled = strings.ToLower(val) == "true"
	}
	if val, exists := os.LookupEnv("CACHE_TTL"); exists {
		if d, err := time.ParseDuration(val); err == nil {
			cfg.Cache.TTL = d
		} else if seconds, err := strconv.Atoi(val); err == nil {
			cfg.Cache.TTL = time.Duration(seconds) * time.Second
		}
	}
	if val, exists := os.LookupEnv("CACHE_CLEANUP_PERIOD"); exists {
		if d, err := time.ParseDuration(val); err == nil {
			cfg.Cache.CleanupPeriod = d
		}
	}
	if val, exists := os.LookupEnv("CACHE_MAX_ITEMS"); exists {
		if n, err := strconv.Atoi(val); err == nil {
			cfg.Cache.MaxItems = n
		}
	}

	// Logging settings
	if val, exists := os.LookupEnv("LOG_LEVEL"); exists {
		cfg.Log.Level = val
	}
	if val, exists := os.LookupEnv("LOG_FORMAT"); exists {
		cfg.Log.Format = val
	}
	if val, exists := os.LookupEnv("LOG_FILE_PATH"); exists {
		cfg.Log.FilePath = val
	}
	if val, exists := os.LookupEnv("LOG_MAX_SIZE"); exists {
		if n, err := strconv.Atoi(val); err == nil {
			cfg.Log.MaxSize = n
		}
	}
	if val, exists := os.LookupEnv("LOG_MAX_BACKUPS"); exists {
		if n, err := strconv.Atoi(val); err == nil {
			cfg.Log.MaxBackups = n
		}
	}
	if val, exists := os.LookupEnv("LOG_MAX_AGE"); exists {
		if n, err := strconv.Atoi(val); err == nil {
			cfg.Log.MaxAge = n
		}
	}
	if val, exists := os.LookupEnv("LOG_COMPRESSION"); exists {
		cfg.Log.Compression = strings.ToLower(val) == "true"
	}

	// Metrics settings
	if val, exists := os.LookupEnv("METRICS_ENABLED"); exists {
		cfg.Metrics.Enabled = strings.ToLower(val) == "true"
	}
	if val, exists := os.LookupEnv("METRICS_PATH"); exists {
		cfg.Metrics.Path = val
	}
}

// applyFlags applies command line flags to the configuration
func applyFlags(cfg *Config) {
	// Server flags
	port := flag.Int("port", cfg.Server.Port, "Server port number")
	host := flag.String("host", cfg.Server.Host, "Server host address")
	readTimeout := flag.Duration("read-timeout", cfg.Server.ReadTimeout, "HTTP read timeout")
	writeTimeout := flag.Duration("write-timeout", cfg.Server.WriteTimeout, "HTTP write timeout")
	idleTimeout := flag.Duration("idle-timeout", cfg.Server.IdleTimeout, "HTTP idle timeout")

	// Prometheus flags
	prometheusURL := flag.String("prometheus-url", cfg.Prometheus.URL, "Prometheus server URL")
	prometheusTimeout := flag.Duration("prometheus-timeout", cfg.Prometheus.Timeout, "Prometheus query timeout")
	prometheusKeepAlive := flag.Duration("prometheus-keep-alive", cfg.Prometheus.KeepAlive, "Prometheus connection keepalive")
	prometheusMaxConnections := flag.Int("prometheus-max-conn", cfg.Prometheus.MaxConnections, "Prometheus max connections")

	// CORS flags
	corsEnabled := flag.Bool("cors-enabled", cfg.CORS.Enabled, "Enable CORS")
	corsOrigins := flag.String("cors-origins", strings.Join(cfg.CORS.AllowedOrigins, ","), "CORS allowed origins, comma separated")
	corsMethods := flag.String("cors-methods", strings.Join(cfg.CORS.AllowedMethods, ","), "CORS allowed methods, comma separated")
	corsHeaders := flag.String("cors-headers", strings.Join(cfg.CORS.AllowedHeaders, ","), "CORS allowed headers, comma separated")
	corsMaxAge := flag.Int("cors-max-age", cfg.CORS.MaxAge, "CORS max age in seconds")

	// Cache flags
	cacheEnabled := flag.Bool("cache-enabled", cfg.Cache.Enabled, "Enable query caching")
	cacheTTL := flag.Duration("cache-ttl", cfg.Cache.TTL, "Cache TTL")
	cacheCleanup := flag.Duration("cache-cleanup", cfg.Cache.CleanupPeriod, "Cache cleanup period")
	cacheMaxItems := flag.Int("cache-max-items", cfg.Cache.MaxItems, "Cache maximum items")

	// Logging flags
	logLevel := flag.String("log-level", cfg.Log.Level, "Log level (debug, info, warn, error)")
	logFormat := flag.String("log-format", cfg.Log.Format, "Log format (text, json)")
	logFilePath := flag.String("log-file", cfg.Log.FilePath, "Log file path (empty for stdout)")
	logMaxSize := flag.Int("log-max-size", cfg.Log.MaxSize, "Log maximum file size in MB")
	logMaxBackups := flag.Int("log-max-backups", cfg.Log.MaxBackups, "Log maximum backup files")
	logMaxAge := flag.Int("log-max-age", cfg.Log.MaxAge, "Log maximum age in days")
	logCompression := flag.Bool("log-compression", cfg.Log.Compression, "Log file compression")

	// Metrics flags
	metricsEnabled := flag.Bool("metrics-enabled", cfg.Metrics.Enabled, "Enable metrics endpoint")
	metricsPath := flag.String("metrics-path", cfg.Metrics.Path, "Metrics endpoint path")

	// Parse all flags
	flag.Parse()

	// Apply parsed flags
	cfg.Server.Port = *port
	cfg.Server.Host = *host
	cfg.Server.ReadTimeout = *readTimeout
	cfg.Server.WriteTimeout = *writeTimeout
	cfg.Server.IdleTimeout = *idleTimeout

	cfg.Prometheus.URL = *prometheusURL
	cfg.Prometheus.Timeout = *prometheusTimeout
	cfg.Prometheus.KeepAlive = *prometheusKeepAlive
	cfg.Prometheus.MaxConnections = *prometheusMaxConnections

	cfg.CORS.Enabled = *corsEnabled
	cfg.CORS.AllowedOrigins = splitAndTrim(*corsOrigins, ",")
	cfg.CORS.AllowedMethods = splitAndTrim(*corsMethods, ",")
	cfg.CORS.AllowedHeaders = splitAndTrim(*corsHeaders, ",")
	cfg.CORS.MaxAge = *corsMaxAge

	cfg.Cache.Enabled = *cacheEnabled
	cfg.Cache.TTL = *cacheTTL
	cfg.Cache.CleanupPeriod = *cacheCleanup
	cfg.Cache.MaxItems = *cacheMaxItems

	cfg.Log.Level = *logLevel
	cfg.Log.Format = *logFormat
	cfg.Log.FilePath = *logFilePath
	cfg.Log.MaxSize = *logMaxSize
	cfg.Log.MaxBackups = *logMaxBackups
	cfg.Log.MaxAge = *logMaxAge
	cfg.Log.Compression = *logCompression

	cfg.Metrics.Enabled = *metricsEnabled
	cfg.Metrics.Path = *metricsPath
}

// ValidateConfig validates the configuration
func (cfg *Config) Validate() []string {
	var errors []string

	// Validate Server configuration
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		errors = append(errors, "server port must be between 1 and 65535")
	}

	// Validate Prometheus configuration
	if cfg.Prometheus.URL == "" {
		errors = append(errors, "prometheus URL cannot be empty")
	}
	if cfg.Prometheus.Timeout <= 0 {
		errors = append(errors, "prometheus timeout must be positive")
	}
	if cfg.Prometheus.MaxConnections <= 0 {
		errors = append(errors, "prometheus max connections must be positive")
	}

	// Validate CORS configuration
	if cfg.CORS.Enabled && len(cfg.CORS.AllowedMethods) == 0 {
		errors = append(errors, "CORS allowed methods cannot be empty when CORS is enabled")
	}

	// Validate Cache configuration
	if cfg.Cache.Enabled && cfg.Cache.TTL <= 0 {
		errors = append(errors, "cache TTL must be positive when cache is enabled")
	}
	if cfg.Cache.Enabled && cfg.Cache.MaxItems <= 0 {
		errors = append(errors, "cache max items must be positive when cache is enabled")
	}

	// Validate Log configuration
	switch cfg.Log.Level {
	case "debug", "info", "warn", "error":
		// Valid levels
	default:
		errors = append(errors, "log level must be one of: debug, info, warn, error")
	}

	switch cfg.Log.Format {
	case "text", "json":
		// Valid formats
	default:
		errors = append(errors, "log format must be one of: text, json")
	}

	// Validate Metrics configuration
	if cfg.Metrics.Enabled && cfg.Metrics.Path == "" {
		errors = append(errors, "metrics path cannot be empty when metrics are enabled")
	}

	return errors
}

// Helper function to split a string by a delimiter and trim spaces
func splitAndTrim(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	
	return result
}