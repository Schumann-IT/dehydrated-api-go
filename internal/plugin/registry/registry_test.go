package registry

import (
	"context"
	"github.com/schumann-it/dehydrated-api-go/internal/model"
	"github.com/schumann-it/dehydrated-api-go/internal/plugin/config"
	"os"
	"path/filepath"
	"testing"
	"time"

	pb "github.com/schumann-it/dehydrated-api-go/plugin/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry(t *testing.T) {
	// Build the example plugin
	pluginPath := filepath.Join("..", "..", "..", "examples", "plugins", "simple", "simple")
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		t.Skip("Example plugin not built, skipping test")
	}

	// Create a test context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a test configuration
	cfg := map[string]config.PluginConfig{
		"simple": config.PluginConfig{
			Enabled: true,
			Path:    pluginPath,
			Config: map[string]interface{}{
				"name": "example",
			},
		},
	}
	r := NewRegistry(cfg)
	defer r.Close()

	var m []model.Metadata
	for _, p := range r.Plugins() {
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
		resp, err := p.GetMetadata(ctx, req)
		require.NoError(t, err)

		m = append(m, model.MetadataFromProto(resp))
	}

	assert.Equal(t, "example", m[0]["name"])
}
