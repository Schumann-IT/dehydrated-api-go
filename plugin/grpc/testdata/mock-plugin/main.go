package main

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/schumann-it/dehydrated-api-go/proto/plugin"
	"google.golang.org/grpc"
)

type mockServer struct {
	plugin.UnimplementedPluginServer
}

func (s *mockServer) Initialize(ctx context.Context, req *plugin.InitializeRequest) (*plugin.InitializeResponse, error) {
	return &plugin.InitializeResponse{}, nil
}

func (s *mockServer) GetMetadata(ctx context.Context, req *plugin.GetMetadataRequest) (*plugin.GetMetadataResponse, error) {
	return &plugin.GetMetadataResponse{
		Metadata: map[string]string{
			"test": "value",
		},
	}, nil
}

func (s *mockServer) Close(ctx context.Context, req *plugin.CloseRequest) (*plugin.CloseResponse, error) {
	return &plugin.CloseResponse{}, nil
}

func main() {
	sockFile := os.Getenv("PLUGIN_SOCKET")
	if sockFile == "" {
		fmt.Fprintln(os.Stderr, "PLUGIN_SOCKET environment variable not set")
		os.Exit(1)
	}

	// Create a Unix domain socket listener
	lis, err := net.Listen("unix", sockFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to listen: %v\n", err)
		os.Exit(1)
	}

	// Create a gRPC server
	s := grpc.NewServer()
	plugin.RegisterPluginServer(s, &mockServer{})

	// Serve
	if err := s.Serve(lis); err != nil {
		fmt.Fprintf(os.Stderr, "failed to serve: %v\n", err)
		os.Exit(1)
	}
}
