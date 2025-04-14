package openssl

import (
	"context"
	"testing"
	"time"

	pb "github.com/schumann-it/dehydrated-api-go/proto/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAnalyzeRSACertificate verifies that the plugin correctly analyzes RSA certificates.
// It tests the extraction of certificate metadata including:
// - Subject and issuer DN
// - Validity period (NotBefore and NotAfter)
// - Key type (RSA) and size (4096 bits)
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
		Domain: "rsa.example.com",
	}
	resp, err := plugin.GetMetadata(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Get the certificate info from metadata
	certInfo, ok := resp.Metadata["cert"].AsInterface().(map[string]any)
	require.True(t, ok)

	// Verify the parsed certificate information
	assert.Equal(t, "CN=rsa.example.com", certInfo["subject"])
	assert.Equal(t, "CN=rsa.example.com", certInfo["issuer"])
	assert.Equal(t, "rsaEncryption", certInfo["key_type"])
	assert.Equal(t, float64(4096), certInfo["key_size"])

	// Parse and verify dates
	notBefore, err := time.Parse(time.RFC3339, certInfo["not_before"].(string))
	require.NoError(t, err)
	assert.Equal(t, time.Date(2025, 4, 3, 15, 42, 36, 0, time.UTC), notBefore)

	notAfter, err := time.Parse(time.RFC3339, certInfo["not_after"].(string))
	require.NoError(t, err)
	assert.Equal(t, time.Date(2026, 4, 3, 15, 42, 36, 0, time.UTC), notAfter)
}

// TestAnalyzeECCertificate verifies that the plugin correctly analyzes EC certificates.
// It tests the extraction of certificate metadata including:
// - Subject and issuer DN
// - Validity period (NotBefore and NotAfter)
// - Key type (EC) and size (256 bits)
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
		Domain: "ec.example.com",
	}
	resp, err := plugin.GetMetadata(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Get the certificate info from metadata
	certInfo, ok := resp.Metadata["cert"].AsInterface().(map[string]any)
	require.True(t, ok)

	// Verify the parsed certificate information
	assert.Equal(t, "CN=ec.example.com", certInfo["subject"])
	assert.Equal(t, "CN=ec.example.com", certInfo["issuer"])
	assert.Equal(t, "ecPublicKey", certInfo["key_type"])
	assert.Equal(t, float64(256), certInfo["key_size"])

	// Parse and verify dates
	notBefore, err := time.Parse(time.RFC3339, certInfo["not_before"].(string))
	require.NoError(t, err)
	assert.Equal(t, time.Date(2025, 4, 3, 15, 42, 45, 0, time.UTC), notBefore)

	notAfter, err := time.Parse(time.RFC3339, certInfo["not_after"].(string))
	require.NoError(t, err)
	assert.Equal(t, time.Date(2026, 4, 3, 15, 42, 45, 0, time.UTC), notAfter)
}
