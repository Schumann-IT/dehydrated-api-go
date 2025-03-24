package main

import (
	"context"
	"log"
	"net"
	"os"

	pb "github.com/schumann-it/dehydrated-api-go/proto/plugin"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
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

	// Create metadata with different value types
	metadata := make(map[string]*structpb.Value)

	// String value
	strValue, _ := structpb.NewValue("value")
	metadata["test"] = strValue

	// Number value
	numValue, _ := structpb.NewValue(42)
	metadata["count"] = numValue

	// Boolean value
	boolValue, _ := structpb.NewValue(true)
	metadata["enabled"] = boolValue

	// List value
	listValue, _ := structpb.NewValue([]interface{}{"item1", "item2"})
	metadata["items"] = listValue

	// Struct value
	structValue, _ := structpb.NewValue(map[string]interface{}{
		"name": "test",
		"age":  42,
	})
	metadata["user"] = structValue

	return &pb.GetMetadataResponse{
		Metadata: metadata,
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
