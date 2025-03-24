package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/schumann-it/dehydrated-api-go/internal"
	"github.com/schumann-it/dehydrated-api-go/internal/handler"
	"github.com/schumann-it/dehydrated-api-go/internal/logger"
	"github.com/schumann-it/dehydrated-api-go/internal/service"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yaml", "Path to the configuration file")
	flag.Parse()

	// Load configuration
	cfg := internal.NewConfig().Load(*configPath)

	// Initialize logger
	if err := logger.Init(cfg.Logging); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	log := logger.L()

	// Create domain service
	domainService, err := service.NewDomainService(service.DomainServiceConfig{
		DehydratedBaseDir: cfg.DehydratedBaseDir,
		EnableWatcher:     cfg.EnableWatcher,
		PluginConfig:      cfg.Plugins,
	})
	if err != nil {
		log.Fatal("Failed to create domain service", zap.Error(err))
	}
	defer domainService.Close()

	// Create fiber app
	app := fiber.New()

	// Create domain handler
	domainHandler := handler.NewDomainHandler(domainService)
	domainHandler.RegisterRoutes(app)

	// Start server in a goroutine
	go func() {
		log.Info("Starting server",
			zap.Int("port", cfg.Port),
			zap.String("config", *configPath),
		)
		if err := app.Listen(fmt.Sprintf(":%d", cfg.Port)); err != nil {
			log.Error("Server error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	log.Info("Shutting down server...")
	if err := app.Shutdown(); err != nil {
		log.Error("Error during shutdown", zap.Error(err))
	}
}
