// Package server provides the HTTP server implementation for the dehydrated-api-go application.
// It handles server lifecycle management, configuration loading, and graceful shutdown.
package server

import (
	"fmt"
	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated"
	"sync"

	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/schumann-it/dehydrated-api-go/internal/handler"
	"github.com/schumann-it/dehydrated-api-go/internal/logger"
	"github.com/schumann-it/dehydrated-api-go/internal/service"
)

// Server represents a running server instance that manages the HTTP server lifecycle.
// It handles server startup, shutdown, and maintains the application state.
type Server struct {
	app      *fiber.App     // The Fiber web framework instance
	shutdown chan struct{}  // Channel for signaling shutdown
	wg       sync.WaitGroup // WaitGroup for managing goroutines
	port     int            // Port number the server listens on
}

// NewServer creates a new server instance with the specified configs.
func NewServer(cfg *Config, dcfg *dehydrated.Config) *Server {
	log := logger.L()

	// Initialize logger with config
	if err := logger.Init(cfg.Logging); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		return nil
	}
	defer logger.Sync()

	log = logger.L()

	log.Info("Logger initialized",
		zap.String("level", cfg.Logging.Level),
		zap.String("encoding", cfg.Logging.Encoding),
		zap.String("output", cfg.Logging.OutputPath),
	)

	// Create domain service
	log.Debug("Creating domain service",
		zap.String("dehydrated_dir", cfg.DehydratedBaseDir),
		zap.String("dehydrated_config_file", cfg.DehydratedConfigFile),
		zap.Bool("watcher_enabled", cfg.EnableWatcher),
	)

	domainService := service.NewDomainService(dcfg.DomainsFile).WithPlugins(cfg.Plugins, dcfg)
	if cfg.EnableWatcher {
		domainService.WithFileWatcher()
	}
	err := domainService.Reload()

	if err != nil {
		log.Fatal("Failed to load domains",
			zap.Error(err),
		)
		return nil
	}

	log.Info("Domain service created successfully",
		zap.Int("enabled_plugins", len(cfg.Plugins)),
	)

	// Create fiber app
	app := fiber.New()

	app.Use(fiberzap.New(fiberzap.Config{
		Logger: log,
	}))

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

		domainService.Close()

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

// Shutdown gracefully shuts down the server and its associated resources.
// It signals all goroutines to stop and waits for them to complete.
// This method blocks until all resources are cleaned up.
func (s *Server) Shutdown() {
	close(s.shutdown)
	s.wg.Wait()
}

// GetPort returns the port number that the server is listening on.
// This is useful for testing and monitoring purposes.
func (s *Server) GetPort() int {
	return s.port
}
