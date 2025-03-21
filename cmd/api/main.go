package main

import (
	"fmt"
	"github.com/schumann-it/dehydrated-api-go/internal"
	"github.com/schumann-it/dehydrated-api-go/internal/handler"
	"github.com/schumann-it/dehydrated-api-go/internal/plugin"
	"github.com/schumann-it/dehydrated-api-go/internal/service"
	"log"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Create fiber app
	app := fiber.New()

	// Load configuration
	configPath, _ := filepath.Abs("config.yaml")
	cfg, err := internal.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize plugin registry
	pluginRegistry := plugin.NewRegistry(cfg)

	// Create domain service
	domainService, err := service.NewDomainService(service.DomainServiceConfig{
		DehydratedBaseDir: cfg.DehydratedBaseDir,
		EnableWatcher:     true,
		PluginRegistry:    pluginRegistry,
	})
	if err != nil {
		log.Fatalf("Failed to create domain service: %v", err)
	}
	defer domainService.Close()

	// Create domain handler
	domainHandler := handler.NewDomainHandler(domainService)
	domainHandler.RegisterRoutes(app)

	// Start server
	log.Fatal(app.Listen(fmt.Sprintf(":%d", cfg.Port)))
}
