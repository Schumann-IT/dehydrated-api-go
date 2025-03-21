package main

import (
	"github.com/schumann-it/dehydrated-api-go/internal/config"
	"github.com/schumann-it/dehydrated-api-go/internal/handler"
	"github.com/schumann-it/dehydrated-api-go/internal/plugin"
	"github.com/schumann-it/dehydrated-api-go/internal/service"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Load configuration
	cfg := config.NewConfig()
	if os.Getenv("DEHYDRATED_BASE_DIR") != "" {
		cfg.WithBaseDir(os.Getenv("DEHYDRATED_BASE_DIR"))
	}
	cfg.Load()

	// Create fiber app
	app := fiber.New()

	// Initialize plugin registry
	pluginRegistry := plugin.NewRegistry(cfg)

	// Create domain service
	domainService, err := service.NewDomainService(service.DomainServiceConfig{
		DomainsFile:    cfg.DomainsFile,
		EnableWatcher:  true,
		PluginRegistry: pluginRegistry,
	})
	if err != nil {
		log.Fatalf("Failed to create domain service: %v", err)
	}
	defer domainService.Close()

	// Create domain handler
	domainHandler := handler.NewDomainHandler(domainService)
	domainHandler.RegisterRoutes(app)

	// Start server
	log.Fatal(app.Listen(":3000"))
}
