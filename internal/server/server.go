package server

import (
	"fmt"
	"sync"

	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/schumann-it/dehydrated-api-go/internal/handler"
	"github.com/schumann-it/dehydrated-api-go/internal/logger"
	"github.com/schumann-it/dehydrated-api-go/internal/service"
	"go.uber.org/zap"
)

// Server represents a running server instance
type Server struct {
	app      *fiber.App
	shutdown chan struct{}
	wg       sync.WaitGroup
	port     int
}

// NewServer creates a new server instance
func NewServer(configPath string) *Server {
	log := logger.L()
	log.Debug("Using configuration file",
		zap.String("config", configPath),
	)

	// Load configuration
	cfg := NewConfig().Load(configPath)
	log.Debug("Loading configuration",
		zap.String("config_file", configPath),
		zap.String("dehydrated_dir", cfg.DehydratedBaseDir),
	)

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

	domainService, err := service.NewDomainService(service.DomainServiceConfig{
		DehydratedBaseDir:    cfg.DehydratedBaseDir,
		DehydratedConfigFile: cfg.DehydratedConfigFile,
		EnableWatcher:        cfg.EnableWatcher,
		PluginConfig:         cfg.Plugins,
	})
	if err != nil {
		log.Fatal("Failed to create domain service",
			zap.Error(err),
			zap.String("dehydrated_dir", cfg.DehydratedBaseDir),
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

	// Create config handler
	configHandler := handler.NewConfigHandler(domainService)
	configHandler.RegisterRoutes(app)

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

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() {
	close(s.shutdown)
	s.wg.Wait()
}

// GetPort returns the port the server is listening on
func (s *Server) GetPort() int {
	return s.port
}
