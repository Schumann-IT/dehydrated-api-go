// Package main provides the entry point for the dehydrated-api-go application.
// It initializes the server with configuration from a YAML file and handles graceful shutdown.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated"

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

// ANSI escape codes for text formatting
const (
	bold  = "\033[1m"
	reset = "\033[0m"
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

	// show version info and exit if --version exists
	outputVersion(*showVersion, true)

	// load server config
	sc := server.NewConfig().Load(*configPath)

	// load dehydrated config
	dc := dehydrated.NewConfig().WithBaseDir(sc.DehydratedBaseDir)
	if sc.DehydratedConfigFile != "" {
		dc = dc.WithConfigFile(sc.DehydratedConfigFile)
	}
	dc.Load()

	// show info and exit if --info exists
	outputConfigs(*showInfo, sc, dc)

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

func outputVersion(doPrint, exit bool) {
	if doPrint {
		fmt.Printf("dehydrated-api-go version %s (commit: %s, built: %s)\n", Version, Commit, BuildTime)
		if exit {
			os.Exit(0)
		}
	}
}

func outputConfigs(doPrint bool, sc *server.Config, dc *dehydrated.Config) {
	if doPrint {
		outputVersion(true, false)
		fmt.Printf("%sServer Config:%s", bold, reset)
		fmt.Println()
		fmt.Printf("%s", sc.String())
		fmt.Println()
		fmt.Printf("%sDehydrated Config:%s", bold, reset)
		fmt.Println()
		fmt.Printf("%s", dc.String())
		os.Exit(0)
	}
}
