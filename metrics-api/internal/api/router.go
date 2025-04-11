package api

import (
	"net/http"
	"time"

	"metrics-api/internal/api/handlers"
	"metrics-api/internal/api/middleware"
	"metrics-api/internal/config"
	"metrics-api/internal/service"
	"metrics-api/pkg/logger"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// RouterOption represents a function that configures a router
type RouterOption func(*RouterConfig)

// RouterConfig contains all dependencies needed for the router
type RouterConfig struct {
	Logger         logger.Logger
	MetricsService *service.MetricsService
	QueriesService *service.QueriesService
	AlertsService  *service.AlertsService
	Config         *config.Config
	Version        string
}

// WithLogger sets the logger for the router
func WithLogger(logger logger.Logger) RouterOption {
	return func(c *RouterConfig) {
		c.Logger = logger
	}
}

// WithMetricsService sets the metrics service for the router
func WithMetricsService(service *service.MetricsService) RouterOption {
	return func(c *RouterConfig) {
		c.MetricsService = service
	}
}

// WithQueriesService sets the queries service for the router
func WithQueriesService(service *service.QueriesService) RouterOption {
	return func(c *RouterConfig) {
		c.QueriesService = service
	}
}

// WithAlertsService sets the alerts service for the router
func WithAlertsService(service *service.AlertsService) RouterOption {
	return func(c *RouterConfig) {
		c.AlertsService = service
	}
}

// WithConfig sets the config for the router
func WithConfig(config *config.Config) RouterOption {
	return func(c *RouterConfig) {
		c.Config = config
	}
}

// WithVersion sets the version for the router
func WithVersion(version string) RouterOption {
	return func(c *RouterConfig) {
		c.Version = version
	}
}

// NewRouter creates a new router with all necessary routes and middleware
func NewRouter(options ...RouterOption) *mux.Router {
	// Create default config
	cfg := &RouterConfig{
		Logger:  logger.NewNopLogger(),
		Version: "dev",
	}
	
	// Apply options
	for _, opt := range options {
		opt(cfg)
	}
	
	// Create router
	router := mux.NewRouter()
	
	// Apply CORS middleware at the root level
	router.Use(middleware.CORSMiddleware)
	
	// Set up API routes
	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	
	// Add other middleware
	apiRouter.Use(middleware.RequestID)
	apiRouter.Use(middleware.LogHTTPErrorMiddleware(cfg.Logger))
	apiRouter.Use(middleware.RequestDurationMiddleware(cfg.Logger, 5*time.Second))
	apiRouter.Use(middleware.LoggingMiddleware(cfg.Logger))
	apiRouter.Use(middleware.RecoveryMiddleware(cfg.Logger))
	
	// Create handlers
	if cfg.MetricsService != nil {
		metricsHandler := handlers.NewMetricsHandler(cfg.MetricsService, cfg.Logger)
		metricsHandler.RegisterRoutes(apiRouter)
	}
	
	if cfg.QueriesService != nil {
		queriesHandler := handlers.NewQueriesHandler(cfg.QueriesService, cfg.Logger)
		queriesHandler.RegisterRoutes(apiRouter)
	}
	
	if cfg.AlertsService != nil {
		alertsHandler := handlers.NewAlertsHandler(cfg.AlertsService, cfg.Logger)
		alertsHandler.RegisterRoutes(apiRouter)
	}
	
	// Always register health handler
	healthHandler := handlers.NewHealthHandler(nil, cfg.Logger, cfg.Version)
	healthHandler.RegisterRoutes(apiRouter)
	
	// Add Prometheus metrics endpoint at /metrics (outside of /api/v1)
	router.Handle("/metrics", promhttp.Handler())
	
	// Add OPTIONS handler for CORS preflight requests
	router.PathPrefix("/").Methods("OPTIONS").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	
	// Add catch-all 404 handler
	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.RespondWithError(w, http.StatusNotFound, "Endpoint not found")
	})
	
	return router
}
