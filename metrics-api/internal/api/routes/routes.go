// internal/api/routes/routes.go
package routes

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"metrics-api/internal/api/handlers"
	"metrics-api/internal/api/middleware"
	"metrics-api/internal/cache"
	"metrics-api/internal/config"
	"metrics-api/internal/prometheus"
	"metrics-api/pkg/logger"
)

// APIRouter is the main router for the API
type APIRouter struct {
	router     chi.Router
	promClient *prometheus.Client
	cache      *cache.Cache
	log        *logger.Logger
	config     *config.Config
}

// NewAPIRouter creates a new API router
func NewAPIRouter(cfg *config.Config, promClient *prometheus.Client, cache *cache.Cache, log *logger.Logger) *APIRouter {
	router := chi.NewRouter()
	
	return &APIRouter{
		router:     router,
		promClient: promClient,
		cache:      cache,
		log:        log.WithField("component", "api_router"),
		config:     cfg,
	}
}

// SetupRoutes sets up all the routes for the API
func (r *APIRouter) SetupRoutes() {
	// Create handlers
	queryHandler := handlers.NewQueryHandler(r.promClient, r.cache, r.log)
	rangeQueryHandler := handlers.NewRangeQueryHandler(r.promClient, r.cache, r.log)
	alertsHandler := handlers.NewAlertsHandler(r.promClient, r.cache, r.log)
	healthHandler := handlers.NewHealthHandler(r.promClient, r.cache, r.log)
	metricsSummaryHandler := handlers.NewMetricsSummaryHandler(r.promClient, r.cache, r.log)
	
	// Set up global middleware
	r.router.Use(middleware.RequestLoggerMiddleware(r.log))
	r.router.Use(middleware.RecoveryMiddleware(r.log))
	r.router.Use(middleware.CORSMiddleware(r.config, r.log))
	
	// Add metrics middleware if enabled
	if r.config.Metrics.Enabled {
		r.router.Use(middleware.MetricsMiddleware())
	}
	
	// Configure optional authentication
	authConfig := middleware.AuthConfig{
		Enabled:         false, // Set to true to enable authentication
		Type:            middleware.AuthTypeAPIKey,
		APIKeys:         map[string]middleware.User{},
		SkipAuthForPath: []string{"/health", "/metrics"},
	}
	
	// Add auth middleware if enabled
	if authConfig.Enabled {
		r.router.Use(middleware.AuthMiddleware(authConfig, r.log))
	}
	
	// Health check endpoint
	r.router.Get("/health", healthHandler.Handle)
	
	// Prometheus metrics endpoint if enabled
	if r.config.Metrics.Enabled {
		r.router.Handle(r.config.Metrics.Path, promhttp.Handler())
	}
	
	// API routes
	r.router.Route("/api", func(r chi.Router) {
		// Query endpoints
		r.Get("/query", queryHandler.Handle)
		r.Post("/query", queryHandler.Handle)
		
		r.Get("/query_range", rangeQueryHandler.Handle)
		r.Post("/query_range", rangeQueryHandler.Handle)
		
		// Alerts endpoints
		r.Get("/alerts", alertsHandler.Handle)
		r.Post("/alerts", alertsHandler.Handle)
		
		// Metrics summary endpoint
		r.Get("/metrics/summary", metricsSummaryHandler.Handle)
		r.Post("/metrics/summary", metricsSummaryHandler.Handle)
		
		// Additional API endpoints can be added here
		r.Route("/v1", func(r chi.Router) {
			// Version-specific endpoints
			r.Get("/query", queryHandler.Handle)
			r.Post("/query", queryHandler.Handle)
			
			r.Get("/query_range", rangeQueryHandler.Handle)
			r.Post("/query_range", rangeQueryHandler.Handle)
			
			r.Get("/alerts", alertsHandler.Handle)
			r.Post("/alerts", alertsHandler.Handle)
			
			r.Get("/metrics/summary", metricsSummaryHandler.Handle)
			r.Post("/metrics/summary", metricsSummaryHandler.Handle)
		})
	})
	
	// Documentation routes
	r.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs", http.StatusMovedPermanently)
	})
	
	// Serve static documentation
	// r.router.Handle("/docs/*", http.StripPrefix("/docs/", http.FileServer(http.Dir("./docs"))))
}

// Router returns the router for use by the server
func (r *APIRouter) Router() http.Handler {
	return r.router
}

// ServeHTTP implements the http.Handler interface
func (r *APIRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}

// AddTimeoutMiddleware adds a timeout middleware to the router
func (r *APIRouter) AddTimeoutMiddleware(timeout time.Duration) {
	r.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			
			req = req.WithContext(ctx)
			next.ServeHTTP(w, req)
		})
	})
}

// AddRateLimiter adds a rate limiter middleware to the router
func (r *APIRouter) AddRateLimiter(requestsPerSecond float64, burst int) {
	// Create a new limiter
	limiter := rate.NewLimiter(rate.Limit(requestsPerSecond), burst)
	
	// Add the rate limiter middleware
	r.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if !limiter.Allow() {
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, req)
		})
	})
}