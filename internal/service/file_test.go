package service

import (
	"github.com/schumann-it/dehydrated-api-go/internal/model"
	"os"
	"path/filepath"
	"testing"
)

func TestFileOperations(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()
	domainsFile := filepath.Join(tmpDir, "domains.txt")

	// Test data
	testEntries := []model.DomainEntry{
		{
			Domain:           "example.com",
			AlternativeNames: []string{"www.example.com"},
			Enabled:          true,
			Comment:          "Test comment",
		},
		{
			Domain:           "example.org",
			AlternativeNames: []string{"www.example.org"},
			Enabled:          false,
			Comment:          "Disabled domain",
		},
		{
			Domain:           "example.net",
			AlternativeNames: []string{"www.example.net"},
			Alias:            "certalias",
			Enabled:          true,
			Comment:          "With alias",
		},
	}

	// Test writing domains file
	t.Run("WriteDomainsFile", func(t *testing.T) {
		err := WriteDomainsFile(domainsFile, testEntries)
		if err != nil {
			t.Fatalf("Failed to write domains file: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(domainsFile); os.IsNotExist(err) {
			t.Error("Expected domains file to be created")
		}
	})

	// Test reading domains file
	t.Run("ReadDomainsFile", func(t *testing.T) {
		entries, err := ReadDomainsFile(domainsFile)
		if err != nil {
			t.Fatalf("Failed to read domains file: %v", err)
		}

		// Verify number of entries
		if len(entries) != len(testEntries) {
			t.Errorf("Expected %d entries, got %d", len(testEntries), len(entries))
		}

		// Verify each entry
		for i, entry := range entries {
			if i >= len(testEntries) {
				t.Errorf("Extra entry found: %v", entry)
				continue
			}

			expected := testEntries[i]
			if entry.Domain != expected.Domain {
				t.Errorf("Entry %d: Expected domain %s, got %s", i, expected.Domain, entry.Domain)
			}

			if len(entry.AlternativeNames) != len(expected.AlternativeNames) {
				t.Errorf("Entry %d: Expected %d alternative names, got %d", i, len(expected.AlternativeNames), len(entry.AlternativeNames))
				continue
			}

			for j, altName := range entry.AlternativeNames {
				if altName != expected.AlternativeNames[j] {
					t.Errorf("Entry %d: Expected alternative name %s, got %s", i, expected.AlternativeNames[j], altName)
				}
			}

			if entry.Alias != expected.Alias {
				t.Errorf("Entry %d: Expected alias %s, got %s", i, expected.Alias, entry.Alias)
			}

			if entry.Enabled != expected.Enabled {
				t.Errorf("Entry %d: Expected enabled %t, got %t", i, expected.Enabled, entry.Enabled)
			}

			if entry.Comment != expected.Comment {
				t.Errorf("Entry %d: Expected comment %s, got %s", i, expected.Comment, entry.Comment)
			}
		}
	})

	// Test reading non-existent file
	t.Run("ReadNonExistentFile", func(t *testing.T) {
		nonExistentFile := filepath.Join(tmpDir, "nonexistent.txt")
		entries, err := ReadDomainsFile(nonExistentFile)
		if err != nil {
			t.Errorf("Failed to read non-existent file: %v", err)
		}
		if len(entries) != 0 {
			t.Errorf("Expected 0 entries, got %d", len(entries))
		}
	})

	// Test writing to read-only directory
	t.Run("WriteToReadOnlyDirectory", func(t *testing.T) {
		readOnlyDir := filepath.Join(tmpDir, "readonly")
		err := os.Mkdir(readOnlyDir, 0444)
		if err != nil {
			t.Fatalf("Failed to create read-only directory: %v", err)
		}

		readOnlyFile := filepath.Join(readOnlyDir, "domains.txt")
		err = WriteDomainsFile(readOnlyFile, testEntries)
		if err == nil {
			t.Error("Expected error when writing to read-only directory, got nil")
		}
	})

	// Test reading file with invalid entries
	t.Run("ReadFileWithInvalidEntries", func(t *testing.T) {
		invalidFile := filepath.Join(tmpDir, "invalid.txt")
		err := os.WriteFile(invalidFile, []byte("invalid@domain.com\n"), 0644)
		if err != nil {
			t.Fatalf("Failed to create invalid file: %v", err)
		}

		entries, err := ReadDomainsFile(invalidFile)
		if err != nil {
			t.Errorf("Failed to read invalid file: %v", err)
		}
		if len(entries) != 0 {
			t.Errorf("Expected 0 entries, got %d", len(entries))
		}
	})
}

func TestComplexDomainsFile(t *testing.T) {
	// Create a temporary file for testing
	tmpFile := filepath.Join(t.TempDir(), "complex_domains.txt")

	// Complex test entries matching the original test
	complexEntries := []model.DomainEntry{
		{Domain: "example.org", AlternativeNames: []string{"www.example.org"}, Enabled: true},
		{Domain: "example.com", AlternativeNames: []string{"www.example.com", "wiki.example.com"}, Enabled: true},
		{Domain: "example.net", AlternativeNames: []string{"www.example.net"}, Alias: "certalias", Enabled: true},
		{Domain: "*.service.example.com", Alias: "service_example_com", Enabled: true},
		{Domain: "*.service.example.org", AlternativeNames: []string{"service.example.org"}, Alias: "star_service_example_org", Enabled: true},
		{Domain: "*.service.example.org", AlternativeNames: []string{"service.example.org"}, Alias: "star_service_example_org_rsa", Enabled: true},
		{Domain: "*.service.example.org", AlternativeNames: []string{"service.example.org"}, Alias: "star_service_example_org_ecdsa", Enabled: true},
		{Domain: "service.example.net", AlternativeNames: []string{"*.service.example.net"}, Enabled: true},
		{Domain: "service.example.net", AlternativeNames: []string{"*.service.example.net"}, Enabled: false},
		{Domain: "service.example.net", AlternativeNames: []string{"*.service.example.net"}, Enabled: false},
	}

	// Write the complex entries
	err := WriteDomainsFile(tmpFile, complexEntries)
	if err != nil {
		t.Fatalf("Failed to write complex domains file: %v", err)
	}

	// Read back the entries
	entries, err := ReadDomainsFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read complex domains file: %v", err)
	}

	// Compare the number of entries
	if len(entries) != len(complexEntries) {
		t.Errorf("Expected %d entries, got %d", len(complexEntries), len(entries))
	}

	// Compare each entry in detail
	for i, entry := range entries {
		if i >= len(complexEntries) {
			t.Errorf("Extra entry found: %v", entry)
			continue
		}

		expected := complexEntries[i]
		if entry.Domain != expected.Domain {
			t.Errorf("Entry %d: Expected domain %s, got %s", i, expected.Domain, entry.Domain)
		}

		if len(entry.AlternativeNames) != len(expected.AlternativeNames) {
			t.Errorf("Entry %d: Expected %d alternative names, got %d", i, len(expected.AlternativeNames), len(entry.AlternativeNames))
			continue
		}

		for j, altName := range entry.AlternativeNames {
			if altName != expected.AlternativeNames[j] {
				t.Errorf("Entry %d: Expected alternative name %s, got %s", i, expected.AlternativeNames[j], altName)
			}
		}

		if entry.Alias != expected.Alias {
			t.Errorf("Entry %d: Expected alias %s, got %s", i, expected.Alias, entry.Alias)
		}

		if entry.Enabled != expected.Enabled {
			t.Errorf("Entry %d: Expected enabled %t, got %t", i, expected.Enabled, entry.Enabled)
		}
	}
}
