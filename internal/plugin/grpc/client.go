package grpc

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated"
	"github.com/schumann-it/dehydrated-api-go/internal/model"
	plugininterface2 "github.com/schumann-it/dehydrated-api-go/internal/plugin/interface"

	pb "github.com/schumann-it/dehydrated-api-go/proto/plugin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/structpb"
)

// Client implements the plugin.Plugin interface using gRPC
type Client struct {
	conn     *grpc.ClientConn
	client   pb.PluginClient
	cmd      *exec.Cmd
	sockFile string
	mu       sync.RWMutex
	tmpDir   string
}

// NewClient creates a new gRPC client for the given plugin
func NewClient(pluginPath string, config map[string]any, dehydratedConfig *dehydrated.Config) (*Client, error) {
	// Create a temporary directory for the socket file
	tmpDir, err := os.MkdirTemp("", "plugin-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	// Create socket file path
	sockFile := filepath.Join(tmpDir, "plugin.sock")

	// Start the plugin process
	cmd := exec.Command(pluginPath)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PLUGIN_SOCKET=%s", sockFile))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("failed to start plugin: %w", err)
	}

	// Create a client instance
	client := &Client{
		cmd:      cmd,
		sockFile: sockFile,
		tmpDir:   tmpDir,
	}

	// Wait for the socket file to be created
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(sockFile); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Connect to the plugin
	conn, err := grpc.Dial(
		"unix://"+sockFile,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		client.Close(context.Background())
		return nil, fmt.Errorf("failed to connect to plugin: %w", err)
	}

	client.conn = conn
	client.client = pb.NewPluginClient(conn)

	// Convert config to structpb.Value map
	configValues, err := convertToStructValue(config)
	if err != nil {
		client.Close(context.Background())
		return nil, fmt.Errorf("failed to convert config: %w", err)
	}

	// Convert dehydrated config to proto format
	dehydratedConfigProto := &pb.DehydratedConfig{
		User:               dehydratedConfig.User,
		Group:              dehydratedConfig.Group,
		BaseDir:            dehydratedConfig.BaseDir,
		CertDir:            dehydratedConfig.CertDir,
		DomainsDir:         dehydratedConfig.DomainsDir,
		AccountsDir:        dehydratedConfig.AccountsDir,
		ChallengesDir:      dehydratedConfig.ChallengesDir,
		ChainCache:         dehydratedConfig.ChainCache,
		DomainsFile:        dehydratedConfig.DomainsFile,
		ConfigFile:         dehydratedConfig.ConfigFile,
		HookScript:         dehydratedConfig.HookScript,
		LockFile:           dehydratedConfig.LockFile,
		OpensslConfig:      dehydratedConfig.OpensslConfig,
		Openssl:            dehydratedConfig.Openssl,
		KeySize:            int32(dehydratedConfig.KeySize),
		Ca:                 dehydratedConfig.Ca,
		OldCa:              dehydratedConfig.OldCa,
		AcceptTerms:        dehydratedConfig.AcceptTerms,
		Ipv4:               dehydratedConfig.Ipv4,
		Ipv6:               dehydratedConfig.Ipv6,
		PreferredChain:     dehydratedConfig.PreferredChain,
		Api:                dehydratedConfig.Api,
		KeyAlgo:            dehydratedConfig.KeyAlgo,
		RenewDays:          int32(dehydratedConfig.RenewDays),
		ForceRenew:         dehydratedConfig.ForceRenew,
		ForceValidation:    dehydratedConfig.ForceValidation,
		PrivateKeyRenew:    dehydratedConfig.PrivateKeyRenew,
		PrivateKeyRollover: dehydratedConfig.PrivateKeyRollover,
		ChallengeType:      dehydratedConfig.ChallengeType,
		WellKnownDir:       dehydratedConfig.WellKnownDir,
		AlpnDir:            dehydratedConfig.AlpnDir,
		HookChain:          dehydratedConfig.HookChain,
		OcspMustStaple:     dehydratedConfig.OcspMustStaple,
		OcspFetch:          dehydratedConfig.OcspFetch,
		OcspDays:           int32(dehydratedConfig.OcspDays),
		NoLock:             dehydratedConfig.NoLock,
		KeepGoing:          dehydratedConfig.KeepGoing,
		FullChain:          dehydratedConfig.FullChain,
		Ocsp:               dehydratedConfig.Ocsp,
		AutoCleanup:        dehydratedConfig.AutoCleanup,
		ContactEmail:       dehydratedConfig.ContactEmail,
		CurlOpts:           dehydratedConfig.CurlOpts,
		ConfigD:            dehydratedConfig.ConfigD,
	}

	// Initialize the plugin
	_, err = client.client.Initialize(context.Background(), &pb.InitializeRequest{
		Config:           configValues,
		DehydratedConfig: dehydratedConfigProto,
	})
	if err != nil {
		client.Close(context.Background())
		return nil, fmt.Errorf("failed to initialize plugin: %w", err)
	}

	return client, nil
}

// convertToStructValue converts a map[string]any to map[string]*structpb.Value
func convertToStructValue(config map[string]any) (map[string]*structpb.Value, error) {
	if config == nil {
		return nil, nil
	}

	result := make(map[string]*structpb.Value)
	for k, v := range config {
		value, err := structpb.NewValue(v)
		if err != nil {
			return nil, fmt.Errorf("failed to convert value for key %s: %w", k, err)
		}
		result[k] = value
	}
	return result, nil
}

// Initialize initializes the plugin with the given configuration
func (c *Client) Initialize(ctx context.Context, config map[string]any, dehydratedConfig *dehydrated.Config) error {
	c.mu.RLock()
	if c.client == nil {
		c.mu.RUnlock()
		return fmt.Errorf("client is nil")
	}
	c.mu.RUnlock()

	// Convert config to structpb.Value map
	configValues, err := convertToStructValue(config)
	if err != nil {
		return fmt.Errorf("failed to convert config: %w", err)
	}

	// Convert dehydrated config to proto format
	dehydratedConfigProto := &pb.DehydratedConfig{
		User:               dehydratedConfig.User,
		Group:              dehydratedConfig.Group,
		BaseDir:            dehydratedConfig.BaseDir,
		CertDir:            dehydratedConfig.CertDir,
		DomainsDir:         dehydratedConfig.DomainsDir,
		AccountsDir:        dehydratedConfig.AccountsDir,
		ChallengesDir:      dehydratedConfig.ChallengesDir,
		ChainCache:         dehydratedConfig.ChainCache,
		DomainsFile:        dehydratedConfig.DomainsFile,
		ConfigFile:         dehydratedConfig.ConfigFile,
		HookScript:         dehydratedConfig.HookScript,
		LockFile:           dehydratedConfig.LockFile,
		OpensslConfig:      dehydratedConfig.OpensslConfig,
		Openssl:            dehydratedConfig.Openssl,
		KeySize:            int32(dehydratedConfig.KeySize),
		Ca:                 dehydratedConfig.Ca,
		OldCa:              dehydratedConfig.OldCa,
		AcceptTerms:        dehydratedConfig.AcceptTerms,
		Ipv4:               dehydratedConfig.Ipv4,
		Ipv6:               dehydratedConfig.Ipv6,
		PreferredChain:     dehydratedConfig.PreferredChain,
		Api:                dehydratedConfig.Api,
		KeyAlgo:            dehydratedConfig.KeyAlgo,
		RenewDays:          int32(dehydratedConfig.RenewDays),
		ForceRenew:         dehydratedConfig.ForceRenew,
		ForceValidation:    dehydratedConfig.ForceValidation,
		PrivateKeyRenew:    dehydratedConfig.PrivateKeyRenew,
		PrivateKeyRollover: dehydratedConfig.PrivateKeyRollover,
		ChallengeType:      dehydratedConfig.ChallengeType,
		WellKnownDir:       dehydratedConfig.WellKnownDir,
		AlpnDir:            dehydratedConfig.AlpnDir,
		HookChain:          dehydratedConfig.HookChain,
		OcspMustStaple:     dehydratedConfig.OcspMustStaple,
		OcspFetch:          dehydratedConfig.OcspFetch,
		OcspDays:           int32(dehydratedConfig.OcspDays),
		NoLock:             dehydratedConfig.NoLock,
		KeepGoing:          dehydratedConfig.KeepGoing,
		FullChain:          dehydratedConfig.FullChain,
		Ocsp:               dehydratedConfig.Ocsp,
		AutoCleanup:        dehydratedConfig.AutoCleanup,
		ContactEmail:       dehydratedConfig.ContactEmail,
		CurlOpts:           dehydratedConfig.CurlOpts,
		ConfigD:            dehydratedConfig.ConfigD,
	}

	req := &pb.InitializeRequest{
		Config:           configValues,
		DehydratedConfig: dehydratedConfigProto,
	}

	_, err = c.client.Initialize(ctx, req)
	if err != nil {
		return fmt.Errorf("%w: %v", plugininterface2.ErrPluginError, err)
	}
	return nil
}

// GetMetadata retrieves metadata for a domain
func (c *Client) GetMetadata(ctx context.Context, entry model.DomainEntry) (map[string]any, error) {
	c.mu.RLock()
	if c.client == nil {
		c.mu.RUnlock()
		return nil, fmt.Errorf("client is nil")
	}
	c.mu.RUnlock()

	req := &pb.GetMetadataRequest{
		Domain:           entry.Domain,
		AlternativeNames: entry.AlternativeNames,
		Alias:            entry.Alias,
		Enabled:          entry.Enabled,
		Comment:          entry.Comment,
		Metadata:         entry.Metadata,
	}

	resp, err := c.client.GetMetadata(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", plugininterface2.ErrPluginError, err)
	}

	// Convert response metadata to map[string]any
	result := make(map[string]any)
	for k, v := range resp.Metadata {
		result[k] = v.AsInterface()
	}

	return result, nil
}

// Close implements plugin.Plugin
func (c *Client) Close(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return fmt.Errorf("connection is nil")
	}

	if c.client == nil {
		return fmt.Errorf("client is nil")
	}

	_, err := c.client.Close(ctx, &pb.CloseRequest{})
	if err != nil {
		return fmt.Errorf("failed to close plugin: %w", err)
	}
	c.conn.Close()

	if c.cmd != nil && c.cmd.Process != nil {
		c.cmd.Process.Kill()
	}

	if c.sockFile != "" {
		os.RemoveAll(filepath.Dir(c.sockFile))
	}

	return nil
}
