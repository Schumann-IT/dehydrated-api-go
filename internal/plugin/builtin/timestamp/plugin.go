package timestamp

import (
	"context"
	"fmt"
	"time"

	pb "github.com/schumann-it/dehydrated-api-go/proto/plugin"
	"google.golang.org/protobuf/types/known/structpb"
)

// Plugin implements the timestamp metadata plugin
type Plugin struct {
	pb.UnimplementedPluginServer
	timeFormat       string
	dehydratedConfig *pb.DehydratedConfig
}

// New creates a new timestamp plugin instance
func New() *Plugin {
	return &Plugin{
		timeFormat: time.RFC3339,
	}
}

// Initialize initializes the plugin with configuration
func (p *Plugin) Initialize(ctx context.Context, req *pb.InitializeRequest) (*pb.InitializeResponse, error) {
	// Get time format from config, default to RFC3339
	if timeFormat, ok := req.Config["time_format"]; ok {
		if str, ok := timeFormat.GetKind().(*structpb.Value_StringValue); ok {
			p.timeFormat = str.StringValue
		}
	}
	p.dehydratedConfig = req.DehydratedConfig
	return &pb.InitializeResponse{}, nil
}

// GetMetadata returns metadata for the given domain
func (p *Plugin) GetMetadata(ctx context.Context, req *pb.GetMetadataRequest) (*pb.GetMetadataResponse, error) {
	metadata := make(map[string]*structpb.Value)

	// Create timestamp value
	timestamp, err := structpb.NewValue(time.Now().Format(p.timeFormat))
	if err != nil {
		return nil, fmt.Errorf("failed to create timestamp value: %v", err)
	}
	metadata["timestamp"] = timestamp

	// Create domain value
	domain, err := structpb.NewValue(req.Domain)
	if err != nil {
		return nil, fmt.Errorf("failed to create domain value: %v", err)
	}
	metadata["domain"] = domain

	return &pb.GetMetadataResponse{
		Metadata: metadata,
	}, nil
}

// Close cleans up any resources
func (p *Plugin) Close(ctx context.Context, req *pb.CloseRequest) (*pb.CloseResponse, error) {
	return &pb.CloseResponse{}, nil
}
