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

func (w *wrapper) Initialize(ctx context.Context, config map[string]any, dehydratedConfig *dehydrated.Config) error {
	// Convert Config to map[string]*structpb.Value
	configMap := make(map[string]*structpb.Value)
	for k, v := range config {
		value, err := structpb.NewValue(v)
		if err != nil {
			return fmt.Errorf("failed to convert Config value for key %s: %w", k, err)
		}
		configMap[k] = value
	}

	// Convert dehydrated Config
	dehydratedConfigProto := &pb.DehydratedConfig{
		BaseDir:       dehydratedConfig.BaseDir,
		CertDir:       dehydratedConfig.CertDir,
		DomainsDir:    dehydratedConfig.DomainsDir,
		ChallengeType: dehydratedConfig.ChallengeType,
		Ca:            dehydratedConfig.Ca,
	}

	req := &pb.InitializeRequest{
		Config:           configMap,
		DehydratedConfig: dehydratedConfigProto,
	}
	_, err := w.server.Initialize(ctx, req)
	return err
}

func (w *wrapper) GetMetadata(ctx context.Context, entry model.DomainEntry) (map[string]any, error) {
	req := entry.ToProto()
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
