package service

import (
	"github.com/schumann-it/dehydrated-api-go/internal"
	"github.com/schumann-it/dehydrated-api-go/internal/model"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDomainService(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()
	domainsFile := filepath.Join(tmpDir, "domains.txt")

	// Build test plugin
	pluginPath := buildTestPlugin(t)
	defer os.Remove(pluginPath)

	// Create plugin config
	pluginConfig := map[string]internal.PluginConfig{
		"test": {
			Enabled: true,
			Path:    pluginPath,
			Config:  map[string]any{"key": "value"},
		},
	}

	// Test with watcher enabled
	t.Run("WithWatcher", func(t *testing.T) {
		service, err := NewDomainService(DomainServiceConfig{
			DehydratedBaseDir: tmpDir,
			EnableWatcher:     true,
			PluginConfig:      pluginConfig,
		})
		if err != nil {
			t.Fatalf("Failed to create domain service: %v", err)
		}
		defer service.Close()

		testDomainServiceOperations(t, service)
	})

	// Test without watcher
	t.Run("WithoutWatcher", func(t *testing.T) {
		service, err := NewDomainService(DomainServiceConfig{
			DehydratedBaseDir: tmpDir,
			EnableWatcher:     false,
			PluginConfig:      pluginConfig,
		})
		if err != nil {
			t.Fatalf("Failed to create domain service: %v", err)
		}
		defer service.Close()

		testDomainServiceOperations(t, service)

		// Test manual file modification without watcher
		// Create a new domain entry in the file directly
		entries := []model.DomainEntry{
			{
				Domain:  "manual.example.com",
				Enabled: true,
			},
		}
		if err := WriteDomainsFile(domainsFile, entries); err != nil {
			t.Fatalf("Failed to write domains file: %v", err)
		}

		// Without watcher, the service should not detect the change
		domains, err := service.ListDomains()
		if err != nil {
			t.Fatalf("Failed to list domains: %v", err)
		}
		if len(domains) != 0 {
			t.Errorf("Expected 0 domains (cache not updated), got %d", len(domains))
		}

		// Manual reload should update the cache
		if err := service.reloadCache(); err != nil {
			t.Fatalf("Failed to reload cache: %v", err)
		}

		domains, err = service.ListDomains()
		if err != nil {
			t.Fatalf("Failed to list domains: %v", err)
		}
		if len(domains) != 1 {
			t.Errorf("Expected 1 domain after manual reload, got %d", len(domains))
		}
	})
}

func testDomainServiceOperations(t *testing.T, service *DomainService) {
	// Test CreateDomain
	t.Run("CreateDomain", func(t *testing.T) {
		entry, err := service.CreateDomain(model.CreateDomainRequest{
			Domain:           "example.com",
			AlternativeNames: []string{"www.example.com"},
			Enabled:          true,
		})
		if err != nil {
			t.Fatalf("Failed to create domain: %v", err)
		}
		if entry.Domain != "example.com" {
			t.Errorf("Expected domain example.com, got %s", entry.Domain)
		}
	})

	// Test CreateInvalidDomain
	t.Run("CreateInvalidDomain", func(t *testing.T) {
		_, err := service.CreateDomain(model.CreateDomainRequest{
			Domain: "invalid..com",
		})
		if err == nil {
			t.Error("Expected error for invalid domain")
		}
	})

	// Test CreateDuplicateDomain
	t.Run("CreateDuplicateDomain", func(t *testing.T) {
		_, err := service.CreateDomain(model.CreateDomainRequest{
			Domain:  "example.com",
			Enabled: true,
		})
		if err == nil {
			t.Error("Expected error when creating duplicate domain")
		}
	})

	// Test GetDomain
	t.Run("GetDomain", func(t *testing.T) {
		entry, err := service.GetDomain("example.com")
		if err != nil {
			t.Fatalf("Failed to get domain: %v", err)
		}
		if entry.Domain != "example.com" {
			t.Errorf("Expected domain example.com, got %s", entry.Domain)
		}
	})

	// Test GetNonExistentDomain
	t.Run("GetNonExistentDomain", func(t *testing.T) {
		_, err := service.GetDomain("nonexistent.com")
		if err == nil {
			t.Error("Expected error for non-existent domain")
		}
	})

	// Test UpdateDomain
	t.Run("UpdateDomain", func(t *testing.T) {
		entry, err := service.UpdateDomain("example.com", model.UpdateDomainRequest{
			AlternativeNames: []string{"www.example.com", "api.example.com"},
			Enabled:          true,
		})
		if err != nil {
			t.Fatalf("Failed to update domain: %v", err)
		}
		if len(entry.AlternativeNames) != 2 {
			t.Errorf("Expected 2 alternative names, got %d", len(entry.AlternativeNames))
		}
	})

	// Test DeleteDomain
	t.Run("DeleteDomain", func(t *testing.T) {
		if err := service.DeleteDomain("example.com"); err != nil {
			t.Fatalf("Failed to delete domain: %v", err)
		}
		domains, err := service.ListDomains()
		if err != nil {
			t.Fatalf("Failed to list domains: %v", err)
		}
		if len(domains) != 0 {
			t.Errorf("Expected 0 domains, got %d", len(domains))
		}
	})

	// Test DeleteNonExistentDomain
	t.Run("DeleteNonExistentDomain", func(t *testing.T) {
		if err := service.DeleteDomain("nonexistent.com"); err == nil {
			t.Error("Expected error when deleting non-existent domain")
		}
	})

	// Test UpdateNonExistentDomain
	t.Run("UpdateNonExistentDomain", func(t *testing.T) {
		_, err := service.UpdateDomain("nonexistent.com", model.UpdateDomainRequest{
			Enabled: true,
		})
		if err == nil {
			t.Error("Expected error when updating non-existent domain")
		}
	})
}

func TestNewDomainService(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()
	domainsFile := filepath.Join(tmpDir, "domains.txt")

	// Test with valid config
	t.Run("ValidConfig", func(t *testing.T) {
		service, err := NewDomainService(DomainServiceConfig{
			DehydratedBaseDir: tmpDir,
			EnableWatcher:     true,
		})
		if err != nil {
			t.Fatalf("Failed to create domain service: %v", err)
		}
		defer service.Close()

		if service.domainsFile != domainsFile {
			t.Errorf("Expected domains file %s, got %s", domainsFile, service.domainsFile)
		}
		if service.watcher == nil {
			t.Error("Expected watcher to be initialized")
		}
	})

	// Test with invalid path
	t.Run("InvalidPath", func(t *testing.T) {
		invalidPath := filepath.Join(tmpDir, "nonexistent", "domains.txt")
		service, err := NewDomainService(DomainServiceConfig{
			DehydratedBaseDir: filepath.Join(tmpDir, "nonexistent"),
			EnableWatcher:     true,
		})
		if err != nil {
			t.Fatalf("Failed to create domain service: %v", err)
		}
		defer service.Close()

		// Verify that the directory and file were created
		if _, err := os.Stat(invalidPath); os.IsNotExist(err) {
			t.Error("Expected domains file to be created")
		}
	})

	// Test without watcher
	t.Run("WithoutWatcher", func(t *testing.T) {
		service, err := NewDomainService(DomainServiceConfig{
			DehydratedBaseDir: tmpDir,
			EnableWatcher:     false,
		})
		if err != nil {
			t.Fatalf("Failed to create domain service: %v", err)
		}
		defer service.Close()

		if service.watcher != nil {
			t.Error("Expected watcher to be nil")
		}
	})
}

// buildTestPlugin builds the test plugin
func buildTestPlugin(t *testing.T) string {
	// Get the current directory
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Build the test plugin
	pluginPath := filepath.Join(dir, "..", "..", "plugin", "grpc", "testdata", "test-plugin", "test-plugin")
	cmd := exec.Command("go", "build", "-o", pluginPath, "main.go")
	cmd.Dir = filepath.Join(dir, "..", "..", "plugin", "grpc", "testdata", "test-plugin")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build test plugin: %v", err)
	}

	return pluginPath
}
