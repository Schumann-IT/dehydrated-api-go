package builtin

import (
	"context"
	"fmt"
	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated"
	"github.com/schumann-it/dehydrated-api-go/internal/model"
	pb "github.com/schumann-it/dehydrated-api-go/proto/plugin"
	"google.golang.org/protobuf/types/known/structpb"
)

// wrapper wraps a built-in plugin to implement the Plugin interface
type wrapper struct {
	server pb.PluginServer
}

func (w *wrapper) Initialize(ctx context.Context, config map[string]any) error {
	// Convert Config to map[string]*structpb.Value
	configMap := make(map[string]*structpb.Value)
	for k, v := range config {
		value, err := structpb.NewValue(v)
		if err != nil {
			return fmt.Errorf("failed to convert Config value for key %s: %w", k, err)
		}
		configMap[k] = value
	}

	req := &pb.InitializeRequest{
		Config: configMap,
	}
	_, err := w.server.Initialize(ctx, req)
	return err
}

func (w *wrapper) GetMetadata(ctx context.Context, entry *model.DomainEntry, dehydratedConfig *dehydrated.Config) (map[string]any, error) {
	req := &pb.GetMetadataRequest{
		DomainEntry:      &entry.DomainEntry,
		DehydratedConfig: &dehydratedConfig.DehydratedConfig,
	}

	resp, err := w.server.GetMetadata(ctx, req)
	if err != nil {
		return nil, err
	}

	return model.FromProto(resp).Metadata, nil
}

func (w *wrapper) Close(ctx context.Context) error {
	req := &pb.CloseRequest{}
	_, err := w.server.Close(ctx, req)
	return err
}
