package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/schumann-it/dehydrated-api-go/internal/handler"
	"github.com/schumann-it/dehydrated-api-go/internal/service"
)

func main() {
	// Get domains file path from environment variable or use default
	domainsFile := os.Getenv("DOMAINS_FILE")
	if domainsFile == "" {
		domainsFile = "domains.txt"
	}

	// Create domain service
	domainService, err := service.NewDomainService(domainsFile)
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

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
