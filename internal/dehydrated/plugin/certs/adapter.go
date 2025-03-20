package certs

import (
	"context"
	"time"

	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated/config"
	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated/model"
	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated/plugin/rpc"
)

// RPCAdapter adapts the certs plugin to the RPC interface
type RPCAdapter struct {
	rpc.UnimplementedPluginServer
	plugin *Plugin
}

// NewRPCAdapter creates a new RPC adapter for the certs plugin
func NewRPCAdapter(plugin *Plugin) *RPCAdapter {
	return &RPCAdapter{plugin: plugin}
}

// Initialize initializes the plugin with configuration
func (a *RPCAdapter) Initialize(ctx context.Context, req *rpc.InitializeRequest) (*rpc.InitializeResponse, error) {
	cfg := &config.Config{
		BaseDir: req.Config.BaseDir,
		CertDir: req.Config.CertDir,
	}

	if err := a.plugin.Initialize(cfg); err != nil {
		return &rpc.InitializeResponse{Error: err.Error()}, nil
	}

	return &rpc.InitializeResponse{}, nil
}

// EnrichDomainEntry enriches a domain entry with additional metadata
func (a *RPCAdapter) EnrichDomainEntry(ctx context.Context, req *rpc.EnrichDomainEntryRequest) (*rpc.EnrichDomainEntryResponse, error) {
	entry := &model.DomainEntry{
		Domain:           req.Entry.Domain,
		AlternativeNames: req.Entry.AlternativeNames,
		Enabled:          req.Entry.Enabled,
		Metadata:         make(map[string]interface{}),
	}

	// Convert metadata
	for k, v := range req.Entry.Metadata {
		switch val := v.Value.(type) {
		case *rpc.MetadataValue_CertInfo:
			notBefore, _ := time.Parse(time.RFC3339, val.CertInfo.NotBefore)
			notAfter, _ := time.Parse(time.RFC3339, val.CertInfo.NotAfter)
			entry.Metadata[k] = &model.CertInfo{
				IsValid:   val.CertInfo.IsValid,
				Issuer:    val.CertInfo.Issuer,
				Subject:   val.CertInfo.Subject,
				NotBefore: notBefore,
				NotAfter:  notAfter,
				Error:     val.CertInfo.Error,
			}
		}
	}

	if err := a.plugin.EnrichDomainEntry(entry); err != nil {
		return &rpc.EnrichDomainEntryResponse{Error: err.Error()}, nil
	}

	// Convert metadata back
	rpcEntry := &rpc.DomainEntry{
		Domain:           entry.Domain,
		AlternativeNames: entry.AlternativeNames,
		Enabled:          entry.Enabled,
		Metadata:         make(map[string]*rpc.MetadataValue),
	}

	for k, v := range entry.Metadata {
		if certInfo, ok := v.(*model.CertInfo); ok {
			rpcEntry.Metadata[k] = &rpc.MetadataValue{
				Value: &rpc.MetadataValue_CertInfo{
					CertInfo: &rpc.CertInfo{
						IsValid:   certInfo.IsValid,
						Issuer:    certInfo.Issuer,
						Subject:   certInfo.Subject,
						NotBefore: certInfo.NotBefore.Format(time.RFC3339),
						NotAfter:  certInfo.NotAfter.Format(time.RFC3339),
						Error:     certInfo.Error,
					},
				},
			}
		}
	}

	return &rpc.EnrichDomainEntryResponse{Entry: rpcEntry}, nil
}

// Close cleans up any resources used by the plugin
func (a *RPCAdapter) Close(ctx context.Context, req *rpc.CloseRequest) (*rpc.CloseResponse, error) {
	if err := a.plugin.Close(); err != nil {
		return &rpc.CloseResponse{Error: err.Error()}, nil
	}

	return &rpc.CloseResponse{}, nil
}
