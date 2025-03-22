package main

import (
	"context"
	"fmt"
	"net"
	"os"

	pb "github.com/schumann-it/dehydrated-api-go/proto/plugin"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

type mockServer struct {
	pb.UnimplementedPluginServer
}

func (s *mockServer) Initialize(ctx context.Context, req *pb.InitializeRequest) (*pb.InitializeResponse, error) {
	return &pb.InitializeResponse{}, nil
}

func (s *mockServer) GetMetadata(ctx context.Context, req *pb.GetMetadataRequest) (*pb.GetMetadataResponse, error) {
	if req.Domain == "error.com" {
		return nil, fmt.Errorf("metadata error")
	}

	metadata := make(map[string]*structpb.Value)

	// String value
	strValue, _ := structpb.NewValue("value")
	metadata["test"] = strValue

	return &pb.GetMetadataResponse{
		Metadata: metadata,
	}, nil
}

func (s *mockServer) Close(ctx context.Context, req *pb.CloseRequest) (*pb.CloseResponse, error) {
	return &pb.CloseResponse{}, nil
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
	pb.RegisterPluginServer(s, &mockServer{})

	// Serve
	if err := s.Serve(lis); err != nil {
		fmt.Fprintf(os.Stderr, "failed to serve: %v\n", err)
		os.Exit(1)
	}
}
