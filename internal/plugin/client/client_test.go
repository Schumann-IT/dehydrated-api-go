package client

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/schumann-it/dehydrated-api-go/internal/plugin/config"

	pb "github.com/schumann-it/dehydrated-api-go/plugin/proto"

	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	// Build the example plugin
	pluginPath := filepath.Join("..", "..", "..", "examples", "plugins", "simple", "simple")
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		t.Skip("Example plugin not built, skipping test")
	}

	// Create a test context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a test configuration
	cfg := &config.PluginConfig{
		Config: map[string]any{
			"name": "example",
		},
	}

	cfgValues, err := cfg.ToProto()
	require.NoError(t, err)

	// Create a new client
	client, err := NewClient(ctx, "example", pluginPath, cfgValues)
	require.NoError(t, err)
	defer client.Close()

	// Create a test domain entry
	domain := &pb.DomainEntry{
		Domain:           "example.com",
		AlternativeNames: []string{"www.example.com"},
		Alias:            "example",
		Enabled:          true,
		Comment:          "Test domain",
	}

	// Create a test dehydrated config
	dehydratedConfig := &pb.DehydratedConfig{
		User:               "test",
		Group:              "test",
		BaseDir:            "/tmp/test",
		CertDir:            "/tmp/test/certs",
		DomainsDir:         "/tmp/test/domains",
		AccountsDir:        "/tmp/test/accounts",
		ChallengesDir:      "/tmp/test/challenges",
		ChainCache:         "/tmp/test/chain",
		DomainsFile:        "/tmp/test/domains.txt",
		ConfigFile:         "/tmp/test/config",
		HookScript:         "/tmp/test/hook.sh",
		LockFile:           "/tmp/test/lock",
		OpensslConfig:      "/tmp/test/openssl.cnf",
		Openssl:            "/usr/bin/openssl",
		KeySize:            2048,
		Ca:                 "https://acme-v02.api.letsencrypt.org/directory",
		OldCa:              "https://acme-v01.api.letsencrypt.org/directory",
		AcceptTerms:        true,
		Ipv4:               true,
		Ipv6:               false,
		PreferredChain:     "ISRG Root X1",
		Api:                "v2",
		KeyAlgo:            "rsa",
		RenewDays:          30,
		ForceRenew:         false,
		ForceValidation:    false,
		PrivateKeyRenew:    false,
		PrivateKeyRollover: false,
		ChallengeType:      "http-01",
		WellKnownDir:       "/tmp/test/.well-known",
		AlpnDir:            "/tmp/test/alpn",
		HookChain:          false,
		OcspMustStaple:     true,
		OcspFetch:          true,
		OcspDays:           7,
		NoLock:             false,
		KeepGoing:          false,
		FullChain:          true,
		Ocsp:               true,
		AutoCleanup:        true,
		ContactEmail:       "test@example.com",
		CurlOpts:           "",
		ConfigD:            "/tmp/test/config.d",
	}

	// Get plugin metadata
	req := &pb.GetMetadataRequest{
		DomainEntry:      domain,
		DehydratedConfig: dehydratedConfig,
	}
	resp, err := client.plugin.GetMetadata(ctx, req)
	require.NoError(t, err)
	require.Equal(t, "example", resp.Metadata["name"].GetStringValue())
	require.Equal(t, "example_value", resp.Metadata["example_key"].GetStringValue())
	//nolint:testifylint // This is a test, so we can use the example number directly
	require.Equal(t, float64(42), resp.Metadata["example_number"].GetNumberValue())
	require.True(t, resp.Metadata["example_bool"].GetBoolValue())
}
