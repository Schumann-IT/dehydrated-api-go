package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/schumann-it/dehydrated-api-go/internal"
	"github.com/schumann-it/dehydrated-api-go/internal/handler"
	"github.com/schumann-it/dehydrated-api-go/internal/service"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yaml", "Path to the configuration file")
	flag.Parse()

	cfg := internal.NewConfig().Load(*configPath)

	domainService, err := service.NewDomainService(service.DomainServiceConfig{
		DehydratedBaseDir: cfg.DehydratedBaseDir,
		EnableWatcher:     cfg.EnableWatcher,
		PluginConfig:      cfg.Plugins,
	})
	if err != nil {
		log.Fatalf("Failed to create domain service: %v", err)
	}
	defer domainService.Close()

	// Create fiber app
	app := fiber.New()

	// Create domain handler
	domainHandler := handler.NewDomainHandler(domainService)
	domainHandler.RegisterRoutes(app)

	// Start server
	log.Printf("Starting server on port %d with config from %s", cfg.Port, *configPath)
	log.Fatal(app.Listen(fmt.Sprintf(":%d", cfg.Port)))
}
