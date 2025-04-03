package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/schumann-it/dehydrated-api-go/internal"
	"github.com/schumann-it/dehydrated-api-go/internal/handler"
	"github.com/schumann-it/dehydrated-api-go/internal/logger"
	"github.com/schumann-it/dehydrated-api-go/internal/service"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// Server represents a running server instance
type Server struct {
	app      *fiber.App
	shutdown chan struct{}
	wg       sync.WaitGroup
	port     int
}

// runServer is a function that can be used for testing without redefining flags
func runServer(configPath string) *Server {
	log := logger.L()
	log.Debug("Using configuration file",
		zap.String("config", configPath),
	)

	// Load configuration
	cfg := internal.NewConfig().Load(configPath)
	log.Debug("Loading configuration",
		zap.String("config_file", configPath),
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

	// Create server instance
	server := &Server{
		app:      app,
		shutdown: make(chan struct{}),
		port:     cfg.Port,
	}

	// Start server in a goroutine
	server.wg.Add(1)
	go func() {
		defer server.wg.Done()
		host := "0.0.0.0" // Listen on all interfaces
		log.Info("Starting server",
			zap.String("host", host),
			zap.Int("port", cfg.Port),
			zap.String("config", configPath),
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

	// Handle shutdown in a separate goroutine
	go func() {
		// Wait for shutdown signal
		<-server.shutdown

		// Graceful shutdown
		log.Info("Starting graceful shutdown")

		if err := app.Shutdown(); err != nil {
			log.Error("Error during shutdown",
				zap.Error(err),
			)
		} else {
			log.Info("Server shutdown completed successfully")
		}
	}()

	return server
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() {
	close(s.shutdown)
	s.wg.Wait()
}

// GetPort returns the port the server is listening on
func (s *Server) GetPort() int {
	return s.port
}

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yaml", "Path to the configuration file")
	flag.Parse()

	server := runServer(*configPath)

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan

	log := logger.L()
	log.Debug("Received signal",
		zap.String("signal", sig.String()),
	)

	// Shutdown server
	server.Shutdown()
}
