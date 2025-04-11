package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"metrics-api/internal/api"
	"metrics-api/internal/cache"
	"metrics-api/internal/config"
	"metrics-api/internal/prometheus"
	"metrics-api/internal/service"
	"metrics-api/pkg/logger"

	"golang.org/x/sync/errgroup"
)

func main() {
	// Initialize logger
	log := logger.NewLogger()
	
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	
	// Create context that listens for termination signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Setup signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		sig := <-sigCh
		log.Infof("Received signal: %v", sig)
		cancel()
	}()
	
	// Initialize cache
	cacheOptions := cache.Options{
		DefaultExpiration: time.Duration(cfg.Cache.TTLSeconds) * time.Second,
		CleanupInterval:   time.Duration(cfg.Cache.TTLSeconds/2) * time.Second,
		MaxItems:          cfg.Cache.MaxSizeItems,
	}
	cacheInstance := cache.New(cacheOptions)
	
	// Initialize Prometheus client
	promClient, err := prometheus.NewClient(
		cfg.Prometheus.URL,
		log,
		cacheInstance,
	)
	if err != nil {
		log.Fatalf("Failed to create Prometheus client: %v", err)
	}
	
	// Initialize services
	metricsSvc := service.NewMetricsService(promClient, log)
	queriesSvc := service.NewQueriesService(promClient, log)
	alertsSvc := service.NewAlertsService(promClient, log)
	
	// Create router with all handlers
	router := api.NewRouter(
		api.WithLogger(log),
		api.WithMetricsService(metricsSvc),
		api.WithQueriesService(queriesSvc),
		api.WithAlertsService(alertsSvc),
		api.WithConfig(cfg),
	)
	
	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeoutSeconds) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeoutSeconds) * time.Second,
	}
	
	// Run server in goroutine
	g, gCtx := errgroup.WithContext(ctx)
	
	g.Go(func() error {
		log.Infof("Starting server on port %d", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
		return nil
	})
	
	g.Go(func() error {
		<-gCtx.Done()
		log.Info("Shutting down server...")
		
		// Create shutdown context with timeout
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server shutdown error: %w", err)
		}
		
		log.Info("Server shut down gracefully")
		return nil
	})
	
	// Wait for all goroutines to complete
	if err := g.Wait(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
