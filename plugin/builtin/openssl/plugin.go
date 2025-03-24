package openssl

import (
	"context"
	pb "github.com/schumann-it/dehydrated-api-go/proto/plugin"
	"google.golang.org/protobuf/types/known/structpb"
)

// Plugin implements the openssl metadata plugin
type Plugin struct {
	pb.UnimplementedPluginServer
	dehydratedConfig *pb.DehydratedConfig
}

// New creates a new openssl plugin instance
func New() *Plugin {
	return &Plugin{}
}

// Initialize initializes the plugin with configuration
func (p *Plugin) Initialize(ctx context.Context, req *pb.InitializeRequest) (*pb.InitializeResponse, error) {
	p.dehydratedConfig = req.DehydratedConfig
	return &pb.InitializeResponse{}, nil
}

// GetMetadata returns metadata for the given domain
func (p *Plugin) GetMetadata(ctx context.Context, req *pb.GetMetadataRequest) (*pb.GetMetadataResponse, error) {
	metadata := make(map[string]*structpb.Value)

	return &pb.GetMetadataResponse{
		Metadata: metadata,
	}, nil
}

// Close cleans up any resources
func (p *Plugin) Close(ctx context.Context, req *pb.CloseRequest) (*pb.CloseResponse, error) {
	return &pb.CloseResponse{}, nil
}
