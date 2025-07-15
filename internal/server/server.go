// Package server provides the HTTP server implementation for the dehydrated-api-go application.
// It handles server lifecycle management, configuration loading, and graceful shutdown.
package server

import (
	"fmt"
	"net"
	"os"
	"sync"

	pluginregistry "github.com/schumann-it/dehydrated-api-go/internal/plugin/registry"

	"github.com/gofiber/fiber/v2/middleware/cors"

	"github.com/gofiber/swagger"

	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/fiber/v2"
	_ "github.com/schumann-it/dehydrated-api-go/docs"
	"github.com/schumann-it/dehydrated-api-go/internal/auth"
	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated"
	"github.com/schumann-it/dehydrated-api-go/internal/handler"
	"github.com/schumann-it/dehydrated-api-go/internal/logger"
	"github.com/schumann-it/dehydrated-api-go/internal/service"
	"go.uber.org/zap"
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
	mu       sync.RWMutex   // RWMutex for protecting server state
	running  bool           // Flag to track if server is running
	port     int            // Port number the server listens on
	started  chan struct{}  // Channel to signal server has started

	Config        *Config
	Logger        *zap.Logger
	domainService *service.DomainService
}

// NewServer creates a new server instance.
func NewServer() *Server {
	return &Server{
		app:      fiber.New(),
		shutdown: make(chan struct{}),
		started:  make(chan struct{}),
		Logger:   zap.NewNop(),
	}
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
	cfg := dehydrated.NewConfig().
		WithBaseDir(s.Config.DehydratedBaseDir).
		WithConfigFile(s.Config.DehydratedConfigFile).
		Load()

	// Create domain service
	s.Logger.Debug("Creating domain service",
		zap.String("dehydrated_dir", s.Config.DehydratedBaseDir),
		zap.String("dehydrated_config_file", s.Config.DehydratedConfigFile),
		zap.Bool("watcher_enabled", s.Config.EnableWatcher),
	)

	r := pluginregistry.New(cfg.BaseDir, s.Config.Plugins, s.Logger)
	domainService := service.NewDomainService(cfg, r)

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

	s.Logger.Info("Domain service created successfully")

	s.domainService = domainService

	return s
}

// Start starts the server and begins listening for requests.
func (s *Server) Start() {
	if !s.setRunning() {
		return
	}

	s.setupMiddleware()
	s.setupRoutes()
	s.startServerGoroutine()
	s.startShutdownGoroutine()

	// Wait for server to start
	<-s.started
}

// setRunning sets the server as running and returns false if already running
func (s *Server) setRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return false
	}

	s.running = true
	s.port = s.Config.Port
	return true
}

// setupMiddleware configures CORS and other middleware
func (s *Server) setupMiddleware() {
	s.app.Use(cors.New())
}

// setupRoutes configures all routes including health, swagger, and API routes
func (s *Server) setupRoutes() {
	// Add health handler
	handler.NewHealthHandler().RegisterRoutes(s.app)

	// Add Swagger documentation
	s.app.Get("/docs/*", swagger.HandlerDefault)

	// add API group
	g := s.app.Group("/api/v1")
	s.setupAuthMiddleware(g)
	s.setupDomainRoutes(g)
}

// setupAuthMiddleware configures authentication middleware for the API group
func (s *Server) setupAuthMiddleware(g fiber.Router) {
	if s.Config.Auth != nil {
		s.Logger.Info("Adding authentication middleware",
			zap.String("tenant_id", s.Config.Auth.TenantID),
			zap.String("client_id", s.Config.Auth.ClientID),
		)

		// Add authentication middleware to the api group
		g.Use(auth.Middleware(s.Config.Auth, s.Logger))
	} else {
		s.Logger.Warn("No authentication middleware configured!!")
	}
}

// setupDomainRoutes configures domain-related routes
func (s *Server) setupDomainRoutes(g fiber.Router) {
	if s.domainService != nil {
		handler.NewDomainHandler(s.domainService).RegisterRoutes(g)
		handler.NewConfigHandler(s.domainService.DehydratedConfig).RegisterRoutes(s.app)
	}
}

// startServerGoroutine starts the server in a separate goroutine
func (s *Server) startServerGoroutine() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.runServer()
	}()
}

// runServer handles the actual server startup and listening
func (s *Server) runServer() {
	host := "0.0.0.0" // Listen on all interfaces

	s.mu.RLock()
	port := s.port
	s.mu.RUnlock()

	// Signal that we're about to start
	close(s.started)

	s.Logger.Info("Starting server",
		zap.String("host", host),
		zap.Int("port", port),
		zap.Bool("watcher_enabled", s.Config.EnableWatcher),
	)

	err := s.listenOnPort(host, port)
	if err != nil {
		s.handleServerError(err, host, port)
	}
}

// listenOnPort handles listening on the specified port
func (s *Server) listenOnPort(host string, port int) error {
	if port == 0 {
		return s.listenOnDynamicPort(host)
	}

	// Use the specified port
	return s.app.Listen(fmt.Sprintf("%s:%d", host, port))
}

// listenOnDynamicPort handles listening on a dynamically assigned port
func (s *Server) listenOnDynamicPort(host string) error {
	// Use a custom listener to get the assigned port
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, 0))
	if err != nil {
		s.Logger.Error("Failed to create listener",
			zap.Error(err),
			zap.String("host", host),
		)
		return err
	}

	// Get the actual assigned port
	addr := listener.Addr().(*net.TCPAddr)
	assignedPort := addr.Port

	// Update the server's port field
	s.mu.Lock()
	s.port = assignedPort
	s.mu.Unlock()

	s.Logger.Info("Server assigned to port",
		zap.Int("assigned_port", assignedPort),
	)

	// Close the listener and let Fiber create its own
	listener.Close()

	// Start Fiber with the assigned port
	return s.app.Listen(fmt.Sprintf("%s:%d", host, assignedPort))
}

// handleServerError handles server startup errors
func (s *Server) handleServerError(err error, host string, port int) {
	// Only log if it's not a normal shutdown
	if err.Error() != "server closed" {
		s.Logger.Error("Server error",
			zap.Error(err),
			zap.String("host", host),
			zap.Int("port", port),
		)
	}
}

// startShutdownGoroutine starts the shutdown handler in a separate goroutine
func (s *Server) startShutdownGoroutine() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.handleShutdown()
	}()
}

// handleShutdown handles graceful shutdown of the server
func (s *Server) handleShutdown() {
	// Wait for shutdown signal
	<-s.shutdown

	s.mu.Lock()
	s.running = false
	s.mu.Unlock()

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
}

// Shutdown gracefully shuts down the server and its associated resources.
func (s *Server) Shutdown() {
	s.mu.RLock()
	if !s.running {
		s.mu.RUnlock()
		return
	}
	s.mu.RUnlock()

	close(s.shutdown)
	s.wg.Wait()
}

// GetPort returns the port number that the server is listening on.
func (s *Server) GetPort() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.port
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
	fmt.Printf("%s\n", s.domainService.DehydratedConfig.String())
}
