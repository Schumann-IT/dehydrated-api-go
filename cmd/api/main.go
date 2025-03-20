package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated/config"
	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated/handler"
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

	// Create domain service
	domainService, err := service.NewDomainService(service.DomainServiceConfig{
		DomainsFile:   cfg.DomainsFile,
		EnableWatcher: true,
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
