package main

import (
	"log"
	"os"

	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated/plugin/certs"
	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated/plugin/rpc"
)

func main() {
	// Configure logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	if os.Getenv("DEHYDRATED_DEBUG") == "1" {
		log.Printf("Debug mode enabled")
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Llongfile)
	}

	// Create certs plugin
	plugin := certs.New()
	log.Printf("Created certs plugin")

	// Create RPC adapter
	adapter := certs.NewRPCAdapter(plugin)
	log.Printf("Created RPC adapter")

	// Start plugin server
	log.Printf("Starting plugin server...")
	if err := rpc.Serve(adapter); err != nil {
		log.Printf("Failed to serve plugin: %v", err)
		os.Exit(1)
	}
}
