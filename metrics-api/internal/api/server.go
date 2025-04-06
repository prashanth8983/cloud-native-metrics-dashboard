package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"metrics-api/internal/api/routes"
	"metrics-api/internal/cache"
	"metrics-api/internal/config"
	"metrics-api/internal/prometheus"
	"metrics-api/pkg/logger"
)

// Server represents the API server
type Server struct {
	config     *config.Config
	log        *logger.Logger
	promClient *prometheus.Client
	cache      *cache.Cache
	router     *routes.APIRouter
	httpServer *http.Server
}

// NewServer creates a new server
func NewServer(cfg *config.Config) (*Server, error) {
	// Create a logger
	log := logger.New(logger.Config{
		Level:       cfg.Log.Level,
		Format:      cfg.Log.Format,
		FilePath:    cfg.Log.FilePath,
		MaxSize:     cfg.Log.MaxSize,
		MaxBackups:  cfg.Log.MaxBackups,
		MaxAge:      cfg.Log.MaxAge,
		Compression: cfg.Log.Compression,
	})
	
	log.Info().Msg("Initializing server")
	
	// Create a cache
	cacheInstance := cache.New(
		cfg.Cache.TTL,
		cfg.Cache.MaxItems,
		cfg.Cache.CleanupPeriod,
	)
	
	// Create a Prometheus client
	promClient, err := prometheus.New(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create Prometheus client: %w", err)
	}
	
	// Create an API router
	router := routes.NewAPIRouter(cfg, promClient, cacheInstance, log)
	
	// Set up the routes
	router.SetupRoutes()
	
	// Create the HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}
	
	return &Server{
		config:     cfg,
		log:        log,
		promClient: promClient,
		cache:      cacheInstance,
		router:     router,
		httpServer: httpServer,
	}, nil
}

// Start starts the server
func (s *Server) Start() error {
	// Channel to listen for errors coming from the listener
	serverErrors := make(chan error, 1)
	
	// Start the server
	go func() {
		s.log.Info().Int("port", s.config.Server.Port).Msg("starting server")
		serverErrors <- s.httpServer.ListenAndServe()
	}()
	
	// Channel to listen for an interrupt or terminate signal from the OS
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)
	
	// Block until an error or OS signal comes through
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
		
	case <-osSignals:
		s.log.Info().Msg("server shutdown initiated")
		
		// Create a context with a timeout to allow for graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		// Attempt to gracefully shutdown the server
		if err := s.httpServer.Shutdown(ctx); err != nil {
			// Log the error and initiate a forceful shutdown
			s.log.Error().Err(err).Msg("graceful shutdown failed, forcing server to close")
			if err := s.httpServer.Close(); err != nil {
				return fmt.Errorf("forced shutdown failed: %w", err)
			}
		}
	}
	
	return nil
}

// Stop stops the server
func (s *Server) Stop() error {
	s.log.Info().Msg("stopping server")
	
	// Create a context with a timeout to allow for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Attempt to gracefully shutdown the server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		// Log the error and initiate a forceful shutdown
		s.log.Error().Err(err).Msg("graceful shutdown failed, forcing server to close")
		if err := s.httpServer.Close(); err != nil {
			return fmt.Errorf("forced shutdown failed: %w", err)
		}
	}
	
	// Close the cache
	s.cache.Close()
	
	return nil
}