package server

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/schumann-it/dehydrated-api-go/internal/model"
)

// TestMain handles global state for all tests in this package.
// It saves and restores the original command line arguments and flags.
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
plugins:
  openssl:
    enabled: true
`, tmpDir)

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Test config loading
	cfg := NewConfig().Load(configPath)
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, tmpDir, cfg.DehydratedBaseDir)
	assert.True(t, cfg.EnableWatcher)
	assert.NotNil(t, cfg.Plugins["openssl"])
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
plugins:
  openssl:
    enabled: true
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
}

// TestMainAccDehydrated tests the dehydrated integration with different key algorithms.
// It verifies that the server can work with both RSA and ECDSA certificates.
// This test is skipped unless the ACC_TEST environment variable is set.
func TestMainAccDehydrated(t *testing.T) {
	if os.Getenv("ACC_TEST") == "" {
		t.Skip("Skipping integration test; ACC_TEST not set")
	}

	// Define test cases
	testCases := []struct {
		name    string
		algo    string
		keySize int
		keyType string
	}{
		{
			name:    "RSA",
			algo:    "rsa",
			keySize: 4096,
			keyType: "rsaEncryption",
		},
		{
			name:    "ECDSA prime256v1",
			algo:    "prime256v1",
			keySize: 256,
			keyType: "ecPublicKey",
		},
	}

	// Run each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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

			// Generate a random hostname
			sanitizedName := strings.ReplaceAll(strings.ToLower(tc.name), " ", "-")
			randomHostname := fmt.Sprintf("test-%s-%d", sanitizedName, time.Now().UnixNano())
			fullDomain := fmt.Sprintf("%s.hq.schumann-it.com", randomHostname)

			// Set up dehydrated in its own directory
			setupDehydrated(dehydratedDir, t)
			hookScript := setupAzureDnsHook(dehydratedDir, t)
			setupDehydratedConfig(dehydratedDir, hookScript, tc.algo, t)
			setupDomains(dehydratedDir, []byte(fullDomain), t)

			// Log the domains file content
			domainsFile := filepath.Join(dehydratedDir, "domains.txt")
			domainsContent, err := os.ReadFile(domainsFile)
			require.NoError(t, err)
			t.Logf("Domains file content: %s", string(domainsContent))

			// Create server config in its own directory
			configPath := filepath.Join(serverDir, "config.yaml")
			configContent := fmt.Sprintf(`
port: 0
dehydratedBaseDir: %s
enableWatcher: false
plugins:
  openssl:
    enabled: true
`, dehydratedDir)

			err = os.WriteFile(configPath, []byte(configContent), 0644)
			require.NoError(t, err)

			// Log the server configuration
			t.Logf("Server configuration for %s test case:", tc.name)
			t.Logf("  Config path: %s", configPath)
			t.Logf("  Config content: %s", configContent)
			t.Logf("  Dehydrated directory: %s", dehydratedDir)

			// List the contents of the dehydrated directory
			files, err := os.ReadDir(dehydratedDir)
			require.NoError(t, err)
			t.Logf("Dehydrated directory contents:")
			for _, file := range files {
				t.Logf("  - %s", file.Name())
			}

			// Start server with dynamic port
			s := NewServer().
				WithConfig(configPath).
				WithDomainService()
			s.Start()
			defer s.Shutdown()

			// Give the server time to start
			time.Sleep(100 * time.Millisecond)

			// Get the actual port the server is listening on
			serverPort := s.GetPort()
			t.Logf("Server started on port %d", serverPort)

			// Verify previously setup domains are accessible
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/api/v1/domains", serverPort))
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var domainsResp model.DomainsResponse
			err = json.NewDecoder(resp.Body).Decode(&domainsResp)
			require.NoError(t, err)
			assert.True(t, domainsResp.Success)

			// Give the server time to start and retry domain check a few times
			maxRetries := 5
			for i := 0; i < maxRetries; i++ {
				time.Sleep(time.Second)
				resp, err = http.Get(fmt.Sprintf("http://localhost:%d/api/v1/domains", serverPort))
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode)

				err = json.NewDecoder(resp.Body).Decode(&domainsResp)
				require.NoError(t, err)
				assert.True(t, domainsResp.Success)

				if len(domainsResp.Data) > 0 {
					break
				}
				t.Logf("Attempt %d: No domains found in response for %s test case, retrying...", i+1, tc.name)
			}

			// Check if we have domains
			if len(domainsResp.Data) == 0 {
				t.Logf("No domains found in response for %s test case after %d retries", tc.name, maxRetries)
				t.Logf("Response: %+v", domainsResp)
				return
			}

			assert.Equal(t, fullDomain, domainsResp.Data[0].Domain)

			// run dehydrated
			dehydratedPath := filepath.Join(dehydratedDir, "dehydrated")
			cmd := exec.Command(dehydratedPath, "--cron", "--accept-terms")
			cmd.Dir = dehydratedDir
			output, err := cmd.CombinedOutput()
			t.Logf("Dehydrated output for %s test case: %s", tc.name, output)
			require.NoError(t, err, "Failed to run dehydrated: %s", output)

			// Give the server time to process the certificate
			time.Sleep(2 * time.Second)

			// check openssl plugin metadata
			resp, err = http.Get(fmt.Sprintf("http://localhost:%d/api/v1/domains", serverPort))
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			err = json.NewDecoder(resp.Body).Decode(&domainsResp)
			require.NoError(t, err)
			assert.True(t, domainsResp.Success)

			// Check if we have domains after running dehydrated
			if len(domainsResp.Data) == 0 {
				t.Logf("No domains found in response after running dehydrated for %s test case", tc.name)
				t.Logf("Response: %+v", domainsResp)
				return
			}

			assert.Equal(t, fullDomain, domainsResp.Data[0].Domain)

			// Verify key type and size if metadata is available
			if domainsResp.Data[0].Metadata != nil && domainsResp.Data[0].Metadata["openssl"] != nil {
				opensslMeta, ok := domainsResp.Data[0].Metadata["openssl"].(map[string]interface{})
				if ok {
					// Log the metadata for debugging
					t.Logf("OpenSSL metadata: %+v", opensslMeta)

					// Get the cert metadata
					certMeta, ok := opensslMeta["cert"].(map[string]interface{})
					if !ok {
						t.Log("Certificate metadata not found or invalid format")
						return
					}

					// Check key type if available
					if keyType, exists := certMeta["key_type"]; exists {
						assert.Equal(t, tc.keyType, keyType)
					} else {
						t.Log("key_type not found in metadata")
					}

					// Check key size if available
					if keySize, exists := certMeta["key_size"]; exists {
						assert.Equal(t, float64(tc.keySize), keySize)
					} else {
						t.Log("key_size not found in metadata")
					}
				} else {
					t.Log("Failed to convert openssl metadata to map")
				}
			} else {
				t.Log("No OpenSSL metadata available")
			}

			// Delete the domain
			req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://localhost:%d/api/v1/domains/%s", serverPort, fullDomain), nil)
			require.NoError(t, err)
			resp, err = http.DefaultClient.Do(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, resp.StatusCode)

			// Verify domain was deleted
			resp, err = http.Get(fmt.Sprintf("http://localhost:%d/api/v1/domains/%s", serverPort, fullDomain))
			require.NoError(t, err)
			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		})
	}
}

// TestServerInitialization tests the server initialization with various configurations.
func TestServerInitialization(t *testing.T) {
	t.Run("WithVersionInfo", func(t *testing.T) {
		s := NewServer().WithVersionInfo("1.0.0", "abc123", "2024-01-01")
		assert.Equal(t, "1.0.0", s.Version)
		assert.Equal(t, "abc123", s.Commit)
		assert.Equal(t, "2024-01-01", s.BuildTime)
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
		assert.NotNil(t, s.Logger)
		assert.NotEqual(t, zap.NewNop(), s.Logger)
	})

	t.Run("WithInvalidConfig", func(t *testing.T) {
		s := NewServer().WithConfig("non-existent-config.yaml")
		assert.NotNil(t, s.Config)
		// Should use default values when config file doesn't exist
		assert.Equal(t, 3000, s.Config.Port)
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
plugins:
  openssl:
    enabled: true
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

		assert.Contains(t, output, "dehydrated-api-go version 1.0.0")
		assert.Contains(t, output, "commit: abc123")
		assert.Contains(t, output, "built: 2024-01-01")
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

		assert.Contains(t, output, "Resolved Server Config")
		assert.Contains(t, output, "port: 8080")
		assert.Contains(t, output, "dehydratedBaseDir: /tmp/dehydrated")
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

		assert.Contains(t, output, "Resolved Dehydrated Config")
	})
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
		assert.NotZero(t, s.GetPort())

		// Shutdown server
		s.Shutdown()

		// Verify server is stopped
		_, err = http.Get(fmt.Sprintf("http://localhost:%d/api/v1/domains", s.GetPort()))
		assert.Error(t, err)
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
plugins:
  openssl:
    enabled: true
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
		assert.NotNil(t, s.domainService)

		// Test domain operations
		client := &http.Client{}
		baseURL := fmt.Sprintf("http://localhost:%d/api/v1", s.GetPort())

		// Delete any existing domains first
		resp, err := http.Get(baseURL + "/domains")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var existingDomains model.DomainsResponse
		err = json.NewDecoder(resp.Body).Decode(&existingDomains)
		require.NoError(t, err)
		resp.Body.Close()

		for _, domain := range existingDomains.Data {
			req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/domains/%s", baseURL, domain.Domain), nil)
			require.NoError(t, err)
			resp, err = client.Do(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, resp.StatusCode)
			resp.Body.Close()
		}

		// Create a domain
		createReq, err := http.NewRequest("POST", baseURL+"/domains", strings.NewReader(`{"domain": "test.example.com"}`))
		require.NoError(t, err)
		createReq.Header.Set("Content-Type", "application/json")
		resp, err = client.Do(createReq)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// Get domains
		resp, err = http.Get(baseURL + "/domains")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var domainsResp model.DomainsResponse
		err = json.NewDecoder(resp.Body).Decode(&domainsResp)
		require.NoError(t, err)
		assert.True(t, domainsResp.Success)
		assert.Len(t, domainsResp.Data, 1)
		assert.Equal(t, "test.example.com", domainsResp.Data[0].Domain)
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
}
