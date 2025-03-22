package main

import (
	"context"
	"log"
	"net"
	"os"

	pb "github.com/schumann-it/dehydrated-api-go/proto/plugin"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedPluginServer
}

func (s *server) Initialize(ctx context.Context, req *pb.InitializeRequest) (*pb.InitializeResponse, error) {
	log.Printf("Initialize called with config: %v", req.Config)
	return &pb.InitializeResponse{}, nil
}

func (s *server) GetMetadata(ctx context.Context, req *pb.GetMetadataRequest) (*pb.GetMetadataResponse, error) {
	log.Printf("GetMetadata called for domain: %s", req.Domain)
	return &pb.GetMetadataResponse{
		Metadata: map[string]string{
			"test": "value",
			"type": "registry-test-plugin",
		},
	}, nil
}

func (s *server) Close(ctx context.Context, req *pb.CloseRequest) (*pb.CloseResponse, error) {
	log.Printf("Close called")
	return &pb.CloseResponse{}, nil
}

func main() {
	sockPath := os.Getenv("PLUGIN_SOCKET")
	if sockPath == "" {
		sockPath = "/tmp/plugin.sock"
	}

	// Remove existing socket file if it exists
	os.Remove(sockPath)

	lis, err := net.Listen("unix", sockPath)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterPluginServer(s, &server{})

	log.Printf("Starting gRPC server on %s", sockPath)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
