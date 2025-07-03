// Package main provides the entry point for the dehydrated-api-go application.
// It initializes the server with configuration from a YAML file and handles graceful shutdown.
//
//nolint:lll // mostly docs
package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/schumann-it/dehydrated-api-go/internal/server"
	"go.uber.org/zap"
)

// @title Dehydrated API
// @version 1.0
// @description This API provides a REST interface to manage domains for https://github.com/dehydrated-io/dehydrated

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:3000
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token. Authentication is optional and depends on server configuration. When authentication is disabled, this header is not required.

// @security BearerAuth
// @description Authentication is optional and depends on server configuration. When enabled, all API endpoints require a valid JWT token in the Authorization header. When disabled, no authentication is required.

var (
	// Version is set during build time
	Version = "dev"
	// Commit is set during build time
	Commit = "unknown"
	// BuildTime is set during build time
	BuildTime = "unknown"
)

// main is the entry point for the dehydrated-api-go application.
// It parses command line flags, initializes the server with the specified configuration,
// and handles graceful shutdown when receiving interrupt signals.
func main() {
	// Parse command line flags
	showVersion := flag.Bool("version", false, "Show version information")
	configPath := flag.String("config", "config.yaml", "Path to the configuration file")
	showInfo := flag.Bool("info", false, "Show parsed config")
	flag.Parse()

	// load server config
	s := server.NewServer().
		WithVersionInfo(Version, Commit, BuildTime).
		WithConfig(*configPath).
		WithLogger().
		WithDomainService()

	s.PrintInfo(*showVersion, *showInfo)

	// start the server
	s.Start()
	defer s.Shutdown()

	// Wait for the interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan

	s.Logger.Debug("Received signal",
		zap.String("signal", sig.String()),
	)
}
