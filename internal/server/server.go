// Package server provides the HTTP server implementation for the dehydrated-api-go application.
// It handles server lifecycle management, configuration loading, and graceful shutdown.
package server

import (
	"fmt"
	"os"
	"sync"

	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated"
	"github.com/schumann-it/dehydrated-api-go/internal/handler"
	"github.com/schumann-it/dehydrated-api-go/internal/logger"
	"go.uber.org/zap"

	"github.com/schumann-it/dehydrated-api-go/internal/service"
)

// ANSI escape codes for text formatting
const (
	bold  = "\033[1m"
	reset = "\033[0m"
)

// Server represents a running server instance that manages the HTTP server lifecycle.
// It handles server startup, shutdown, and maintains the application state.
type Server struct {
	Version   string
	Commit    string
	BuildTime string

	app      *fiber.App     // The Fiber web framework instance
	shutdown chan struct{}  // Channel for signaling shutdown
	wg       sync.WaitGroup // WaitGroup for managing goroutines
	port     int            // Port number the server listens on

	Config           *Config
	Logger           *zap.Logger
	dehydratedConfig *dehydrated.Config
	domainService    *service.DomainService
}

// NewServer creates a new server instance.
func NewServer() *Server {
	server := &Server{
		app:      fiber.New(),
		shutdown: make(chan struct{}),
		Logger:   zap.NewNop(),
	}

	return server
}

func (s *Server) WithVersionInfo(v, c, b string) *Server {
	s.Version = v
	s.Commit = c
	s.BuildTime = b

	return s
}

func (s *Server) WithConfig(path string) *Server {
	s.Config = NewConfig().Load(path)

	return s
}

func (s *Server) WithLogger() *Server {
	if s.Config != nil {
		// Initialize logger with config
		l, _ := logger.NewLogger(s.Config.Logging)
		s.Logger = l
	}

	s.app.Use(fiberzap.New(fiberzap.Config{
		Logger: s.Logger,
	}))

	return s
}

func (s *Server) WithDomainService() *Server {
	cfg := dehydrated.NewConfig().WithBaseDir(s.Config.DehydratedBaseDir)
	if s.Config.DehydratedConfigFile != "" {
		cfg = cfg.WithConfigFile(s.Config.DehydratedConfigFile)
	}
	cfg.Load()

	s.dehydratedConfig = cfg

	// Create domain service
	s.Logger.Debug("Creating domain service",
		zap.String("dehydrated_dir", s.Config.DehydratedBaseDir),
		zap.String("dehydrated_config_file", s.Config.DehydratedConfigFile),
		zap.Bool("watcher_enabled", s.Config.EnableWatcher),
	)

	domainService := service.NewDomainService(s.dehydratedConfig.DomainsFile).
		WithPlugins(s.Config.Plugins, s.dehydratedConfig)

	if s.Logger != nil {
		domainService.WithLogger(s.Logger)
	}

	if s.Config.EnableWatcher {
		domainService.WithFileWatcher()
	}

	err := domainService.Reload()

	if err != nil {
		s.Logger.Fatal("Failed to load domains",
			zap.Error(err),
		)
		return s
	}

	s.Logger.Info("Domain service created successfully",
		zap.Int("enabled_plugins", len(s.Config.Plugins)),
	)

	s.domainService = domainService

	// Create domain handler
	h := handler.NewDomainHandler(s.domainService)
	h.RegisterRoutes(s.app)

	return s
}

func (s *Server) Start() {
	// Create health handler
	h := handler.NewHealthHandler()
	h.RegisterRoutes(s.app)

	// Start server in a goroutine
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		host := "0.0.0.0" // Listen on all interfaces
		s.Logger.Info("Starting server",
			zap.String("host", host),
			zap.Int("port", s.Config.Port),
			zap.Bool("watcher_enabled", s.Config.EnableWatcher),
			zap.Int("enabled_plugins", len(s.Config.Plugins)),
		)

		if err := s.app.Listen(fmt.Sprintf("%s:%d", host, s.Config.Port)); err != nil {
			s.Logger.Error("Server error",
				zap.Error(err),
				zap.String("host", host),
				zap.Int("port", s.Config.Port),
			)
		}
	}()

	// Handle shutdown in a separate goroutine
	go func() {
		// Wait for shutdown signal
		<-s.shutdown

		// Graceful shutdown
		s.Logger.Info("Starting graceful shutdown")

		if s.domainService != nil {
			s.domainService.Close()
		}

		if err := s.app.Shutdown(); err != nil {
			s.Logger.Error("Error during shutdown",
				zap.Error(err),
			)
		} else {
			s.Logger.Info("Server shutdown completed successfully")
		}

		s.Logger.Sync()
	}()
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
	return s.Config.Port
}

func (s *Server) PrintInfo(v, i bool) {
	if v {
		s.PrintVersion()
	}

	if i {
		s.PrintServerConfig()
		s.PrintDehydratedConfig()
	}

	if v || i {
		os.Exit(0)
	}
}

func (s *Server) PrintVersion() {
	fmt.Printf("dehydrated-api-go version %s (commit: %s, built: %s)\n", s.Version, s.Commit, s.BuildTime)
}

func (s *Server) PrintServerConfig() {
	fmt.Printf("%sResolved Server Config:%s\n", bold, reset)
	fmt.Printf("%s\n", s.Config.String())
}

func (s *Server) PrintDehydratedConfig() {
	fmt.Printf("%sResolved Dehydrated Config:%s\n", bold, reset)
	fmt.Printf("%s\n", s.dehydratedConfig.String())
}
