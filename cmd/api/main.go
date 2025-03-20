package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
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

	// Create domain service
	domainService, err := service.NewDomainService(cfg.DomainsFile)
	if err != nil {
		log.Fatalf("Failed to create domain service: %v", err)
	}

	// Create domain handler
	domainHandler := handler.NewDomainHandler(domainService)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Dehydrated API",
	})

	// Add middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	// Register routes
	domainHandler.RegisterRoutes(app)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Starting server on port %s", port)
	log.Printf("Using domains file: %s", cfg.DomainsFile)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
