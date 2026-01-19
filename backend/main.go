package main

import (
	"log"

	"shortlink/internal/config"
	"shortlink/internal/handler"
	"shortlink/internal/repository"
	"shortlink/internal/router"
	"shortlink/internal/service"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize repository (Redis)
	repo, err := repository.NewRedisRepository(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}
	defer repo.Close()

	// Initialize service layer
	linkService := service.NewLinkService(repo)

	// Initialize handler
	linkHandler := handler.NewLinkHandler(linkService, cfg.Server.BaseURL)

	// Setup router
	r := router.Setup(linkHandler, cfg.Server.Mode)

	// Start server
	log.Printf("Server starting on port %s", cfg.Server.Port)
	log.Printf("Base URL: %s", cfg.Server.BaseURL)

	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
