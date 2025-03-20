package service

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated/model"
)

// DomainService handles domain-related business logic
type DomainService struct {
	domainsFile string
}

// NewDomainService creates a new DomainService instance
func NewDomainService(domainsFile string) (*DomainService, error) {
	// Ensure the domains file exists
	if _, err := os.Stat(domainsFile); os.IsNotExist(err) {
		// Create the directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(domainsFile), 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}
		// Create an empty domains file
		if err := os.WriteFile(domainsFile, []byte{}, 0644); err != nil {
			return nil, fmt.Errorf("failed to create domains file: %w", err)
		}
	}

	return &DomainService{
		domainsFile: domainsFile,
	}, nil
}

// CreateDomain adds a new domain entry
func (s *DomainService) CreateDomain(req model.CreateDomainRequest) (*model.DomainEntry, error) {
	entry := model.DomainEntry{
		Domain:           req.Domain,
		AlternativeNames: req.AlternativeNames,
		Alias:            req.Alias,
		Enabled:          req.Enabled,
		Comment:          req.Comment,
	}

	// Validate the domain entry
	if !model.IsValidDomainEntry(entry) {
		return nil, fmt.Errorf("invalid domain entry: %v", entry)
	}

	// Read existing entries
	entries, err := ReadDomainsFile(s.domainsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read domains file: %w", err)
	}

	// Check if domain already exists
	for _, existing := range entries {
		if existing.Domain == entry.Domain {
			return nil, fmt.Errorf("domain already exists: %s", entry.Domain)
		}
	}

	// Add the new entry
	entries = append(entries, entry)

	// Write back to file
	if err := WriteDomainsFile(s.domainsFile, entries); err != nil {
		return nil, fmt.Errorf("failed to write domains file: %w", err)
	}

	return &entry, nil
}

// GetDomain retrieves a domain entry
func (s *DomainService) GetDomain(domain string) (*model.DomainEntry, error) {
	entries, err := ReadDomainsFile(s.domainsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read domains file: %w", err)
	}

	for _, entry := range entries {
		if entry.Domain == domain {
			return &entry, nil
		}
	}

	return nil, fmt.Errorf("domain not found: %s", domain)
}

// ListDomains returns all domain entries
func (s *DomainService) ListDomains() ([]model.DomainEntry, error) {
	entries, err := ReadDomainsFile(s.domainsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read domains file: %w", err)
	}
	return entries, nil
}

// UpdateDomain updates an existing domain entry
func (s *DomainService) UpdateDomain(domain string, req model.UpdateDomainRequest) (*model.DomainEntry, error) {
	entries, err := ReadDomainsFile(s.domainsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read domains file: %w", err)
	}

	// Find and update the domain
	found := false
	var updatedEntry model.DomainEntry
	for i, existing := range entries {
		if existing.Domain == domain {
			updatedEntry = model.DomainEntry{
				Domain:           domain,
				AlternativeNames: req.AlternativeNames,
				Alias:            req.Alias,
				Enabled:          req.Enabled,
				Comment:          req.Comment,
			}

			// Validate the updated entry
			if !model.IsValidDomainEntry(updatedEntry) {
				return nil, fmt.Errorf("invalid domain entry: %v", updatedEntry)
			}

			entries[i] = updatedEntry
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("domain not found: %s", domain)
	}

	// Write back to file
	if err := WriteDomainsFile(s.domainsFile, entries); err != nil {
		return nil, fmt.Errorf("failed to write domains file: %w", err)
	}

	return &updatedEntry, nil
}

// DeleteDomain removes a domain entry
func (s *DomainService) DeleteDomain(domain string) error {
	entries, err := ReadDomainsFile(s.domainsFile)
	if err != nil {
		return fmt.Errorf("failed to read domains file: %w", err)
	}

	found := false
	newEntries := make([]model.DomainEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.Domain == domain {
			found = true
			continue
		}
		newEntries = append(newEntries, entry)
	}

	if !found {
		return fmt.Errorf("domain not found: %s", domain)
	}

	if err := WriteDomainsFile(s.domainsFile, newEntries); err != nil {
		return fmt.Errorf("failed to write domains file: %w", err)
	}

	return nil
}
