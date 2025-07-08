package server

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/schumann-it/dehydrated-api-go/internal/plugin/cache"

	"github.com/schumann-it/dehydrated-api-go/internal/model"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestMain handles global state for all tests in this package.
// It saves and restores the original command line arguments and flags.
func TestMain(m *testing.M) {
	// Save original state
	originalArgs := os.Args
	originalFlags := flag.CommandLine

	// Run tests
	_ = m.Run()

	// Restore original state
	os.Args = originalArgs
	flag.CommandLine = originalFlags
}

// TestConfigLoading verifies that the configuration file is properly loaded
// and parsed with the correct values for port, dehydrated directory,
// watcher status, and plugin configuration.
func TestConfigLoading(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := fmt.Sprintf(`
port: 8080
dehydratedBaseDir: %s
enableWatcher: true
`, tmpDir)

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Test config loading
	cfg := NewConfig().Load(configPath)
	require.Equal(t, 8080, cfg.Port)
	require.Equal(t, tmpDir, cfg.DehydratedBaseDir)
	require.True(t, cfg.EnableWatcher)
}

// TestMainAccIntegration performs an integration test of the main application.
// It creates a temporary configuration, starts the server, and verifies it runs correctly.
// This test is skipped unless the ACC_TEST environment variable is set.
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
`, tmpDir)

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Start server
	s := NewServer().
		WithConfig(configPath).
		WithDomainService()
	s.Start()
	defer s.Shutdown()

	// Give the server time to start
	time.Sleep(100 * time.Millisecond)

	cache.Clean()
}

// TestMainAccWithInvalidPort tests the server's behavior when configured with an invalid port.
// It verifies that the server handles the error gracefully without panicking.
// This test is skipped unless the ACC_TEST environment variable is set.
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
	s := NewServer().
		WithConfig(configPath).
		WithDomainService()
	s.Start()
	defer s.Shutdown()

	// Give the server time to start and log the error
	time.Sleep(100 * time.Millisecond)

	cache.Clean()
}

// TestServerInitialization tests the server initialization with various configurations.
func TestServerInitialization(t *testing.T) {
	t.Run("WithVersionInfo", func(t *testing.T) {
		s := NewServer().WithVersionInfo("1.0.0", "abc123", "2024-01-01")
		require.Equal(t, "1.0.0", s.Version)
		require.Equal(t, "abc123", s.Commit)
		require.Equal(t, "2024-01-01", s.BuildTime)
	})

	t.Run("WithLogger", func(t *testing.T) {
		// Create a temporary config file with logging configuration
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")
		configContent := `
port: 8080
logging:
  level: debug
  format: json
`
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		s := NewServer().WithConfig(configPath).WithLogger()
		require.NotNil(t, s.Logger)
		require.NotEqual(t, zap.NewNop(), s.Logger)
	})

	t.Run("WithInvalidConfig", func(t *testing.T) {
		s := NewServer().WithConfig("non-existent-config.yaml")
		require.NotNil(t, s.Config)
		// Should use default values when config file doesn't exist
		require.Equal(t, 3000, s.Config.Port)
	})
}

// TestServerPrintFunctions tests the server's print functions.
func TestServerPrintFunctions(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := `
port: 8080
dehydratedBaseDir: /tmp/dehydrated
enableWatcher: true
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	s := NewServer().
		WithVersionInfo("1.0.0", "abc123", "2024-01-01").
		WithConfig(configPath)

	// Test PrintVersion
	t.Run("PrintVersion", func(t *testing.T) {
		// Capture stdout
		old := os.Stdout
		r, w, err := os.Pipe()
		require.NoError(t, err)
		os.Stdout = w

		s.PrintVersion()

		w.Close()
		os.Stdout = old

		var output string
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			output += scanner.Text() + "\n"
		}
		require.NoError(t, scanner.Err())

		require.Contains(t, output, "dehydrated-api-go version 1.0.0")
		require.Contains(t, output, "commit: abc123")
		require.Contains(t, output, "built: 2024-01-01")
	})

	// Test PrintServerConfig
	t.Run("PrintServerConfig", func(t *testing.T) {
		old := os.Stdout
		r, w, err := os.Pipe()
		require.NoError(t, err)
		os.Stdout = w

		s.PrintServerConfig()

		w.Close()
		os.Stdout = old

		var output string
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			output += scanner.Text() + "\n"
		}
		require.NoError(t, scanner.Err())

		require.Contains(t, output, "Resolved Server Config")
		require.Contains(t, output, "port: 8080")
		require.Contains(t, output, "dehydratedBaseDir: /tmp/dehydrated")
	})

	// Test PrintDehydratedConfig
	t.Run("PrintDehydratedConfig", func(t *testing.T) {
		// Initialize dehydrated config first
		s.WithDomainService()

		old := os.Stdout
		r, w, err := os.Pipe()
		require.NoError(t, err)
		os.Stdout = w

		s.PrintDehydratedConfig()

		w.Close()
		os.Stdout = old

		var output string
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			output += scanner.Text() + "\n"
		}
		require.NoError(t, scanner.Err())

		require.Contains(t, output, "Resolved Dehydrated Config")
	})

	cache.Clean()
}

// TestServerLifecycle tests the server's lifecycle management.
func TestServerLifecycle(t *testing.T) {
	t.Run("StartAndShutdown", func(t *testing.T) {
		// Create a temporary config file
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")
		configContent := `
port: 0
dehydratedBaseDir: /tmp/dehydrated
enableWatcher: false
`
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		s := NewServer().
			WithConfig(configPath).
			WithLogger()

		// Start server
		s.Start()

		// Give the server time to start
		time.Sleep(100 * time.Millisecond)

		// Verify server is running
		require.NotZero(t, s.GetPort())

		s.Shutdown()
		time.Sleep(100 * time.Millisecond)

		// Verify server is stopped
		//nolint:bodyclose // the resp is empty here
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/api/v1/domains", s.GetPort()))
		require.Error(t, err)
		require.Nil(t, resp)
	})

	t.Run("StartWithInvalidPort", func(t *testing.T) {
		// Create a temporary config file with invalid port
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")
		configContent := `
port: 0
dehydratedBaseDir: /tmp/dehydrated
enableWatcher: false
`
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		s := NewServer().
			WithConfig(configPath).
			WithLogger()

		// Start server - should log error but not panic
		s.Start()
		defer s.Shutdown()

		// Give the server time to start and log the error
		time.Sleep(100 * time.Millisecond)
	})
}

// TestDomainServiceIntegration tests the server's integration with the domain service.
func TestDomainServiceIntegration(t *testing.T) {
	t.Run("WithDomainService", func(t *testing.T) {
		// Create a temporary config file
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")
		configContent := `
port: 3000
dehydratedBaseDir: /tmp/dehydrated
enableWatcher: true
`
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		s := NewServer().
			WithConfig(configPath).
			WithLogger().
			WithDomainService()

		// Start server
		s.Start()
		defer s.Shutdown()

		// Give the server time to start
		time.Sleep(100 * time.Millisecond)

		// Verify domain service is initialized
		require.NotNil(t, s.domainService)

		// Test domain operations
		client := &http.Client{}
		baseURL := fmt.Sprintf("http://localhost:%d/api/v1", s.GetPort())

		// Delete any existing domains first
		resp, err := http.Get(baseURL + "/domains")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var existingDomains model.DomainsResponse
		err = json.NewDecoder(resp.Body).Decode(&existingDomains)
		require.NoError(t, err)
		resp.Body.Close()

		for _, domain := range existingDomains.Data {
			req, err2 := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/domains/%s", baseURL, domain.Domain), http.NoBody)
			require.NoError(t, err2)
			resp, err2 = client.Do(req)
			require.NoError(t, err2)
			require.Equal(t, http.StatusNoContent, resp.StatusCode)
			resp.Body.Close()
		}

		// Create a domain
		createReq, err := http.NewRequest("POST", baseURL+"/domains", strings.NewReader(`{"domain": "test.example.com"}`))
		require.NoError(t, err)
		createReq.Header.Set("Content-Type", "application/json")
		resp, err = client.Do(createReq)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		resp.Body.Close()

		// Get domains
		resp, err = http.Get(baseURL + "/domains")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var domainsResp model.DomainsResponse
		err = json.NewDecoder(resp.Body).Decode(&domainsResp)
		resp.Body.Close()
		require.NoError(t, err)
		require.True(t, domainsResp.Success)
		require.Len(t, domainsResp.Data, 1)
		require.Equal(t, "test.example.com", domainsResp.Data[0].Domain)
	})

	t.Run("WithInvalidDomainService", func(t *testing.T) {
		// Create a temporary config file with invalid plugin configuration
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")
		configContent := `
port: 0
dehydratedBaseDir: /tmp/dehydrated
enableWatcher: true
plugins:
  invalid:
    enabled: true
`
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		s := NewServer().
			WithConfig(configPath).
			WithLogger()

		// Start server - should log error but not panic
		s.Start()
		defer s.Shutdown()

		// Give the server time to start and log the error
		time.Sleep(100 * time.Millisecond)
	})

	cache.Clean()
}
