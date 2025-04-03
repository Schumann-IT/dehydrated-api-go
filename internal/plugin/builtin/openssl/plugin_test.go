package openssl

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	pb "github.com/schumann-it/dehydrated-api-go/proto/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyzeRSACertificate(t *testing.T) {
	// Create a new plugin instance
	plugin := New()

	// Initialize the plugin with test configuration
	dehydratedConfig := &pb.DehydratedConfig{
		CertDir: "testdata/certs",
	}
	_, err := plugin.Initialize(context.Background(), &pb.InitializeRequest{
		DehydratedConfig: dehydratedConfig,
	})
	require.NoError(t, err)

	// Test GetMetadata with the example domain
	req := &pb.GetMetadataRequest{
		Domain: "hotspot.hq.schumann-it.com",
	}
	resp, err := plugin.GetMetadata(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Get the certificate info from metadata
	certInfo, ok := resp.Metadata[filepath.Join("testdata", "certs", "hotspot.hq.schumann-it.com", "cert.pem")].AsInterface().(map[string]any)
	require.True(t, ok)

	// Verify the parsed certificate information
	assert.Equal(t, "CN=hotspot.hq.schumann-it.com", certInfo["subject"])
	assert.Equal(t, "CN=R10,O=Let's Encrypt,C=US", certInfo["issuer"])
	assert.Equal(t, "rsaEncryption", certInfo["key_type"])
	assert.Equal(t, float64(2048), certInfo["key_size"])

	// Parse and verify dates
	notBefore, err := time.Parse(time.RFC3339, certInfo["not_before"].(string))
	require.NoError(t, err)
	assert.Equal(t, time.Date(2025, 1, 22, 22, 10, 49, 0, time.UTC), notBefore)

	notAfter, err := time.Parse(time.RFC3339, certInfo["not_after"].(string))
	require.NoError(t, err)
	assert.Equal(t, time.Date(2025, 4, 22, 22, 10, 48, 0, time.UTC), notAfter)
}

func TestAnalyzeECCertificate(t *testing.T) {
	// Create a new plugin instance
	plugin := New()

	// Initialize the plugin with test configuration
	dehydratedConfig := &pb.DehydratedConfig{
		CertDir: "testdata/certs",
	}
	_, err := plugin.Initialize(context.Background(), &pb.InitializeRequest{
		DehydratedConfig: dehydratedConfig,
	})
	require.NoError(t, err)

	// Test GetMetadata with the example domain
	req := &pb.GetMetadataRequest{
		Domain: "synology.hq.schumann-it.com",
	}
	resp, err := plugin.GetMetadata(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Get the certificate info from metadata
	certInfo, ok := resp.Metadata[filepath.Join("testdata", "certs", "synology.hq.schumann-it.com", "cert.pem")].AsInterface().(map[string]any)
	require.True(t, ok)

	// Verify the parsed certificate information
	assert.Equal(t, "CN=synology.hq.schumann-it.com", certInfo["subject"])
	assert.Equal(t, "CN=R3,O=Let's Encrypt,C=US", certInfo["issuer"])
	assert.Equal(t, "ecPublicKey", certInfo["key_type"])
	assert.Equal(t, float64(256), certInfo["key_size"])

	// Not Before: Dec 23 22:10:29 2022 GMT
	//            Not After : Mar 23 22:10:28 2023 GMT
	// Parse and verify dates
	notBefore, err := time.Parse(time.RFC3339, certInfo["not_before"].(string))
	require.NoError(t, err)
	assert.Equal(t, time.Date(2022, 12, 23, 22, 10, 29, 0, time.UTC), notBefore)

	notAfter, err := time.Parse(time.RFC3339, certInfo["not_after"].(string))
	require.NoError(t, err)
	assert.Equal(t, time.Date(2023, 3, 23, 22, 10, 28, 0, time.UTC), notAfter)
}
