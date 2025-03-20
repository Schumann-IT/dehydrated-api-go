package main

import (
	"log"
	"os"

	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated/plugin/certs"
	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated/plugin/rpc"
)

func main() {
	// Create certs plugin
	plugin := certs.New()

	// Create RPC adapter
	adapter := certs.NewRPCAdapter(plugin)

	// Start plugin server
	if err := rpc.Serve(adapter); err != nil {
		log.Printf("Failed to serve plugin: %v", err)
		os.Exit(1)
	}
}
