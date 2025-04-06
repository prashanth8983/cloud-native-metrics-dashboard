package main

import (
	"log"
	"os"

	"metrics-api/internal/api"
	"metrics-api/internal/config"
)

func main() {
	// Load the configuration
	cfg := config.New()
	
	// Validate the configuration
	validationErrors := cfg.Validate()
	if len(validationErrors) > 0 {
		for _, err := range validationErrors {
			log.Printf("Configuration error: %s", err)
		}
		os.Exit(1)
	}
	
	// Create a new server
	server, err := api.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	
	// Start the server
	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}