package main

import (
	"fmt"
	"github.com/schumann-it/dehydrated-api-go/dehydrated"
	"github.com/schumann-it/dehydrated-api-go/dehydrated/handler"
	"github.com/schumann-it/dehydrated-api-go/dehydrated/service"
	"log"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Create fiber app
	app := fiber.New()

	// Load configuration
	configPath, _ := filepath.Abs("config.yaml")
	cfg, err := dehydrated.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	// Create domain service
	domainService, err := service.NewDomainService(service.DomainServiceConfig{
		DehydratedBaseDir: cfg.DehydratedBaseDir,
		EnableWatcher:     true,
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
