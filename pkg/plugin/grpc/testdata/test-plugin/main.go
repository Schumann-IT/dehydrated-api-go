package main

import (
	"context"
	"log"
	"net"
	"os"

	"github.com/schumann-it/dehydrated-api-go/proto/plugin"
	"google.golang.org/grpc"
)

type server struct {
	plugin.UnimplementedPluginServer
}

func (s *server) Initialize(ctx context.Context, req *plugin.InitializeRequest) (*plugin.InitializeResponse, error) {
	log.Printf("Initialize called with config: %v", req.Config)
	return &plugin.InitializeResponse{}, nil
}

func (s *server) GetMetadata(ctx context.Context, req *plugin.GetMetadataRequest) (*plugin.GetMetadataResponse, error) {
	log.Printf("GetMetadata called for domain: %s with config: %v", req.Domain, req.Config)
	return &plugin.GetMetadataResponse{
		Metadata: map[string]string{
			"test": "value",
		},
	}, nil
}

func (s *server) Close(ctx context.Context, req *plugin.CloseRequest) (*plugin.CloseResponse, error) {
	log.Printf("Close called")
	return &plugin.CloseResponse{}, nil
}

func main() {
	sockFile := os.Getenv("PLUGIN_SOCKET")
	if sockFile == "" {
		sockFile = "/tmp/plugin.sock"
	}

	// Remove existing socket file if it exists
	os.Remove(sockFile)

	lis, err := net.Listen("unix", sockFile)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	plugin.RegisterPluginServer(s, &server{})

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
