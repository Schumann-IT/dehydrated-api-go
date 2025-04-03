package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/schumann-it/dehydrated-api-go/internal"
	"github.com/schumann-it/dehydrated-api-go/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain handles global state for all tests in this package
func TestMain(m *testing.M) {
	// Save original state
	originalArgs := os.Args
	originalFlags := flag.CommandLine

	// Run tests
	code := m.Run()

	// Restore original state
	os.Args = originalArgs
	flag.CommandLine = originalFlags

	os.Exit(code)
}

func TestConfigLoading(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := fmt.Sprintf(`
port: 8080
dehydratedBaseDir: %s
enableWatcher: true
plugins:
  openssl:
    enabled: true
`, tmpDir)

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Test config loading
	cfg := internal.NewConfig().Load(configPath)
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, tmpDir, cfg.DehydratedBaseDir)
	assert.True(t, cfg.EnableWatcher)
	assert.NotNil(t, cfg.Plugins["openssl"])
}

func TestMainAccIntegration(t *testing.T) {
	if os.Getenv("ACC_TEST") == "" {
		t.Skip("Skipping integration test; ACC_TEST not set")
	}

	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := fmt.Sprintf(`
port: 0
dehydratedBaseDir: %s
enableWatcher: false
plugins:
  openssl:
    enabled: true
`, tmpDir)

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Start server
	server := runServer(configPath)
	defer server.Shutdown()

	// Give the server time to start
	time.Sleep(100 * time.Millisecond)
}

func TestMainAccWithInvalidPort(t *testing.T) {
	if os.Getenv("ACC_TEST") == "" {
		t.Skip("Skipping integration test; ACC_TEST not set")
	}

	// Create a temporary config file with invalid port
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := fmt.Sprintf(`
port: -1
dehydratedBaseDir: %s
enableWatcher: false
`, tmpDir)

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Start server - it should log an error but not panic
	server := runServer(configPath)
	defer server.Shutdown()

	// Give the server time to start and log the error
	time.Sleep(100 * time.Millisecond)
}

func TestMainAccDehydratedWithRSA(t *testing.T) {
	if os.Getenv("ACC_TEST") == "" {
		t.Skip("Skipping integration test; ACC_TEST not set")
	}

	// Create a temporary config file with dynamic port
	tmpDir := t.TempDir()

	// Create separate directories for dehydrated and server
	dehydratedDir := filepath.Join(tmpDir, "dehydrated")
	serverDir := filepath.Join(tmpDir, "server")

	// Create the directories
	err := os.MkdirAll(dehydratedDir, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(serverDir, 0755)
	require.NoError(t, err)

	// Set up dehydrated in its own directory
	setupDehydrated(dehydratedDir, t)
	hookScript := setupAzureDnsHook(dehydratedDir, t)
	setupDehydratedConfig(dehydratedDir, hookScript, "rsa", t)
	setupDomains(dehydratedDir, []byte(`
foo01.hq.schumann-it.com
`), t)

	// Create server config in its own directory
	configPath := filepath.Join(serverDir, "config.yaml")
	configContent := fmt.Sprintf(`
port: 0
dehydratedBaseDir: %s
enableWatcher: true
plugins:
  openssl:
    enabled: true
`, dehydratedDir)

	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Start server with dynamic port
	server := runServer(configPath)
	defer server.Shutdown()

	// Give the server time to start
	time.Sleep(100 * time.Millisecond)

	// Get the actual port the server is listening on
	serverPort := server.GetPort()

	// Verify previously setup domains are accessible
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/api/v1/domains", serverPort))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var domainsResp model.DomainsResponse
	err = json.NewDecoder(resp.Body).Decode(&domainsResp)
	require.NoError(t, err)
	assert.True(t, domainsResp.Success)
	assert.Equal(t, "foo01.hq.schumann-it.com", domainsResp.Data[0].Domain)

	// run dehydrated
	dehydratedPath := filepath.Join(dehydratedDir, "dehydrated")
	cmd := exec.Command(dehydratedPath, "--cron", "--accept-terms")
	cmd.Dir = dehydratedDir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to run dehydrated: %s", output)

	// check openssl plugin metadata
	resp, err = http.Get(fmt.Sprintf("http://localhost:%d/api/v1/domains", serverPort))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&domainsResp)
	require.NoError(t, err)
	assert.True(t, domainsResp.Success)
	assert.Equal(t, "foo01.hq.schumann-it.com", domainsResp.Data[0].Domain)
	assert.NotNil(t, "foo01.hq.schumann-it.com", domainsResp.Data[0].Metadata["openssl"])

	// Delete the domain
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://localhost:%d/api/v1/domains/foo01.hq.schumann-it.com", serverPort), nil)
	require.NoError(t, err)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify domain was deleted
	resp, err = http.Get(fmt.Sprintf("http://localhost:%d/api/v1/domains/foo01.hq.schumann-it.com", serverPort))
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	// run dehydrated
	cmd = exec.Command(dehydratedPath, "--cron", "--accept-terms")
	cmd.Dir = dehydratedDir
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "Failed to run dehydrated: %s", output)

	// Give the server time to start and log the error
	time.Sleep(5000 * time.Millisecond)

	// check domains again
	resp, err = http.Get(fmt.Sprintf("http://localhost:%d/api/v1/domains", serverPort))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&domainsResp)
	require.NoError(t, err)
	assert.True(t, domainsResp.Success)
	assert.Equal(t, "foo01.hq.schumann-it.com", domainsResp.Data[0].Domain)
	assert.NotNil(t, "foo01.hq.schumann-it.com", domainsResp.Data[0].Metadata["openssl"])
}
