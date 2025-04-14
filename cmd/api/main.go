// Package main provides the entry point for the dehydrated-api-go application.
// It initializes the server with configuration from a YAML file and handles graceful shutdown.
package main

import (
	"flag"
	"fmt"
	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/schumann-it/dehydrated-api-go/internal/logger"
	"github.com/schumann-it/dehydrated-api-go/internal/server"
)

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
	configPath := flag.String("config", "config.yaml", "Path to the configuration file")
	showVersion := flag.Bool("version", false, "Show version information")
	verbose := flag.Bool("verbose", false, "Verbose output")
	flag.Parse()

	// load server config
	sc := server.NewConfig().Load(*configPath)

	// load dehydrated config
	dc := dehydrated.NewConfig().WithBaseDir(sc.DehydratedBaseDir)
	if sc.DehydratedConfigFile != "" {
		dc = dc.WithConfigFile(sc.DehydratedConfigFile)
	}
	dc.Load()

	// show version info
	if *showVersion {
		fmt.Printf("dehydrated-api-go version %s (commit: %s, built: %s)\n", Version, Commit, BuildTime)
		if *verbose {
			fmt.Println("Server Config:")
			fmt.Printf("%s", sc.String())
			fmt.Println("")
			fmt.Println("Dehydrated Config")
			fmt.Printf("%s", dc.String())
		}
		os.Exit(0)
	}

	// start the server
	s := server.NewServer(sc, dc)

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan

	log := logger.L()
	log.Debug("Received signal",
		zap.String("signal", sig.String()),
	)

	// Shutdown server
	s.Shutdown()
}
