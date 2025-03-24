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

	log := logger.L()
	log.Debug("Parsing command line flags",
		zap.String("config", *configPath),
	)

	// Load configuration
	cfg := internal.NewConfig().Load(*configPath)
	log.Debug("Loading configuration",
		zap.String("config_file", *configPath),
		zap.String("dehydrated_dir", cfg.DehydratedBaseDir),
	)

	// Initialize logger with config
	if err := logger.Init(cfg.Logging); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	log.Info("Logger initialized",
		zap.String("level", cfg.Logging.Level),
		zap.String("encoding", cfg.Logging.Encoding),
		zap.String("output", cfg.Logging.OutputPath),
	)

	// Create domain service
	log.Debug("Creating domain service",
		zap.String("dehydrated_dir", cfg.DehydratedBaseDir),
		zap.Bool("watcher_enabled", cfg.EnableWatcher),
	)

	domainService, err := service.NewDomainService(service.DomainServiceConfig{
		DehydratedBaseDir: cfg.DehydratedBaseDir,
		EnableWatcher:     cfg.EnableWatcher,
		PluginConfig:      cfg.Plugins,
	})
	if err != nil {
		log.Fatal("Failed to create domain service",
			zap.Error(err),
			zap.String("dehydrated_dir", cfg.DehydratedBaseDir),
		)
	}
	defer domainService.Close()

	log.Info("Domain service created successfully",
		zap.Int("enabled_plugins", len(cfg.Plugins)),
	)

	// Create fiber app
	app := fiber.New()

	// Create domain handler
	domainHandler := handler.NewDomainHandler(domainService)
	domainHandler.RegisterRoutes(app)

	// Start server in a goroutine
	go func() {
		host := "0.0.0.0" // Listen on all interfaces
		log.Info("Starting server",
			zap.String("host", host),
			zap.Int("port", cfg.Port),
			zap.String("config", *configPath),
			zap.Bool("watcher_enabled", cfg.EnableWatcher),
			zap.Int("enabled_plugins", len(cfg.Plugins)),
		)
		if err := app.Listen(fmt.Sprintf("%s:%d", host, cfg.Port)); err != nil {
			log.Error("Server error",
				zap.Error(err),
				zap.String("host", host),
				zap.Int("port", cfg.Port),
			)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan

	log.Debug("Received signal",
		zap.String("signal", sig.String()),
	)

	// Graceful shutdown
	log.Info("Starting graceful shutdown",
		zap.String("signal", sig.String()),
	)

	if err := app.Shutdown(); err != nil {
		log.Error("Error during shutdown",
			zap.Error(err),
			zap.String("signal", sig.String()),
		)
	} else {
		log.Info("Server shutdown completed successfully",
			zap.String("signal", sig.String()),
		)
	}
}
