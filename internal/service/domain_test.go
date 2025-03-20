package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/schumann-it/dehydrated-api-go/internal/model"
)

func TestDomainService(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()
	domainsFile := filepath.Join(tmpDir, "domains.txt")

	// Create a new domain service
	ds, err := NewDomainService(domainsFile)
	if err != nil {
		t.Fatalf("Failed to create domain service: %v", err)
	}

	// Test adding a domain
	t.Run("CreateDomain", func(t *testing.T) {
		req := model.CreateDomainRequest{
			Domain:           "example.com",
			AlternativeNames: []string{"www.example.com"},
			Enabled:          true,
		}

		entry, err := ds.CreateDomain(req)
		if err != nil {
			t.Errorf("Failed to create domain: %v", err)
		}
		if entry.Domain != req.Domain {
			t.Errorf("Expected domain %s, got %s", req.Domain, entry.Domain)
		}
	})

	// Test adding an invalid domain
	t.Run("CreateInvalidDomain", func(t *testing.T) {
		req := model.CreateDomainRequest{
			Domain:           "invalid@domain.com",
			AlternativeNames: []string{"www.example.com"},
			Enabled:          true,
		}

		_, err := ds.CreateDomain(req)
		if err == nil {
			t.Error("Expected error when creating invalid domain, got nil")
		}
	})

	// Test adding a duplicate domain
	t.Run("CreateDuplicateDomain", func(t *testing.T) {
		req := model.CreateDomainRequest{
			Domain:           "example.com",
			AlternativeNames: []string{"www.example.com"},
			Enabled:          true,
		}

		_, err := ds.CreateDomain(req)
		if err == nil {
			t.Error("Expected error when creating duplicate domain, got nil")
		}
	})

	// Test getting a domain
	t.Run("GetDomain", func(t *testing.T) {
		entry, err := ds.GetDomain("example.com")
		if err != nil {
			t.Errorf("Failed to get domain: %v", err)
		}
		if entry.Domain != "example.com" {
			t.Errorf("Expected domain example.com, got %s", entry.Domain)
		}
	})

	// Test getting a non-existent domain
	t.Run("GetNonExistentDomain", func(t *testing.T) {
		_, err := ds.GetDomain("nonexistent.com")
		if err == nil {
			t.Error("Expected error when getting non-existent domain, got nil")
		}
	})

	// Test updating a domain
	t.Run("UpdateDomain", func(t *testing.T) {
		req := model.UpdateDomainRequest{
			AlternativeNames: []string{"www.example.com", "mail.example.com"},
			Enabled:          true,
		}

		entry, err := ds.UpdateDomain("example.com", req)
		if err != nil {
			t.Errorf("Failed to update domain: %v", err)
		}
		if len(entry.AlternativeNames) != 2 {
			t.Errorf("Expected 2 alternative names, got %d", len(entry.AlternativeNames))
		}
	})

	// Test deleting a domain
	t.Run("DeleteDomain", func(t *testing.T) {
		err := ds.DeleteDomain("example.com")
		if err != nil {
			t.Errorf("Failed to delete domain: %v", err)
		}

		// Verify the domain was deleted
		entries, err := ds.ListDomains()
		if err != nil {
			t.Fatalf("Failed to list domains: %v", err)
		}
		if len(entries) != 0 {
			t.Errorf("Expected 0 domains, got %d", len(entries))
		}
	})

	// Test deleting a non-existent domain
	t.Run("DeleteNonExistentDomain", func(t *testing.T) {
		err := ds.DeleteDomain("nonexistent.com")
		if err == nil {
			t.Error("Expected error when deleting non-existent domain, got nil")
		}
	})

	// Test updating a non-existent domain
	t.Run("UpdateNonExistentDomain", func(t *testing.T) {
		req := model.UpdateDomainRequest{
			AlternativeNames: []string{"www.example.com"},
			Enabled:          true,
		}

		_, err := ds.UpdateDomain("nonexistent.com", req)
		if err == nil {
			t.Error("Expected error when updating non-existent domain, got nil")
		}
	})
}

func TestNewDomainService(t *testing.T) {
	// Test creating a domain service with a non-existent directory
	tmpDir := t.TempDir()
	domainsFile := filepath.Join(tmpDir, "subdir", "domains.txt")

	ds, err := NewDomainService(domainsFile)
	if err != nil {
		t.Fatalf("Failed to create domain service: %v", err)
	}

	// Verify the file was created
	if _, err := os.Stat(domainsFile); os.IsNotExist(err) {
		t.Error("Expected domains file to be created")
	}

	// Verify we can use the domain service
	entries, err := ds.ListDomains()
	if err != nil {
		t.Errorf("Failed to list domains: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Expected 0 domains, got %d", len(entries))
	}
}
