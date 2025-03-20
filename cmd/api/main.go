package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated/config"
	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated/handler"
	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated/plugin"
	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated/plugin/rpc"
	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated/service"
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
	defer pluginRegistry.Close()

	// Register RPC certs plugin
	certsPlugin, err := rpc.NewClient("bin/certs-plugin")
	if err != nil {
		log.Fatalf("Failed to create certs plugin client: %v", err)
	}
	if err := pluginRegistry.Register(certsPlugin); err != nil {
		certsPlugin.Close()
		log.Fatalf("Failed to register certs plugin: %v", err)
	}

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
