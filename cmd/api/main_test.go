package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/schumann-it/dehydrated-api-go/internal"
	"github.com/schumann-it/dehydrated-api-go/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	// Set up test environment
	os.Args = []string{"cmd", "-config", configPath}

	// Run main in a goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Main panicked: %v", r)
			}
		}()
		main()
	}()

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

	// Set up test environment
	os.Args = []string{"cmd", "-config", configPath}

	// Run main and expect it to fail
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with invalid port, but got none")
		}
	}()
	main()
}

func TestMainAccDehydratedWithRSA(t *testing.T) {
	if os.Getenv("ACC_TEST") == "" {
		t.Skip("Skipping integration test; ACC_TEST not set")
	}

	// Create a temporary config file with invalid port
	tmpDir := t.TempDir()

	setupDehydrated(tmpDir, t)
	hookScript := setupAzureDnsHook(tmpDir, t)
	setupDehydratedConfig(tmpDir, hookScript, "rsa", t)
	setupDomains(tmpDir, []byte(`
foobar.hq.schumann-it.com
`), t)
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := fmt.Sprintf(`
port: 0
dehydratedBaseDir: %s
enableWatcher: false
plugins:
  openssl:
    enabled: true
    path: ""
    config: {}
`, tmpDir)

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Set up test environment
	os.Args = []string{"cmd", "-config", configPath}

	// Run main in a goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Main panicked: %v", r)
			}
		}()
		main()
	}()

	// Give the server time to start
	time.Sleep(100 * time.Millisecond)

	// Test domains setup
	t.Run("Test domains setup", func(t *testing.T) {
		// Verify previously setup domains are accessible
		resp, err := http.Get("http://localhost:3000/api/v1/domains")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var domainsResp model.DomainsResponse
		err = json.NewDecoder(resp.Body).Decode(&domainsResp)
		require.NoError(t, err)
		assert.True(t, domainsResp.Success)
		assert.Equal(t, "foobar.hq.schumann-it.com", domainsResp.Data[0].Domain)

		// run dehydrated
		//dehydratedPath := filepath.Join(tmpDir, "dehydrated")
		//cmd := exec.Command(dehydratedPath, "--cron", "--accept-terms")
		//cmd.Dir = tmpDir
		//output, err := cmd.CombinedOutput()
		//require.NoError(t, err, "Failed to run dehydrated: %s", output)

		// check openssl plugin metadata
		// Verify previously setup domains are accessible
		resp, err = http.Get("http://localhost:3000/api/v1/domains")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		err = json.NewDecoder(resp.Body).Decode(&domainsResp)
		require.NoError(t, err)
		assert.True(t, domainsResp.Success)
		assert.Equal(t, "foobar.hq.schumann-it.com", domainsResp.Data[0].Domain)

		// Delete the domain
		req, err := http.NewRequest(http.MethodDelete, "http://localhost:3000/api/v1/domains/foobar.hq.schumann-it.com", nil)
		require.NoError(t, err)
		resp, err = http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		// Verify domain was deleted
		resp, err = http.Get("http://localhost:3000/api/v1/domains/foobar.hq.schumann-it.com")
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
