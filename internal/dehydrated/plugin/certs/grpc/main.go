package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated/plugin/certs/grpc/server"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 0, "Port to listen on (0 for random)")
	flag.Parse()

	srv := server.NewServer()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		srv.Close(nil, nil)
		os.Exit(0)
	}()

	if err := srv.Serve(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to serve: %v\n", err)
		os.Exit(1)
	}
}
