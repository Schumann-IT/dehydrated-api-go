package service

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated/model"
)

// DomainServiceConfig holds configuration options for the DomainService
type DomainServiceConfig struct {
	DomainsFile   string
	EnableWatcher bool
}

// DomainService handles domain-related business logic
type DomainService struct {
	domainsFile string
	watcher     *FileWatcher
	cache       []model.DomainEntry
	mutex       sync.RWMutex
}

// NewDomainService creates a new DomainService instance
func NewDomainService(config DomainServiceConfig) (*DomainService, error) {
	// Ensure the domains file exists
	if _, err := os.Stat(config.DomainsFile); os.IsNotExist(err) {
		// Create the directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(config.DomainsFile), 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}
		// Create an empty domains file
		if err := os.WriteFile(config.DomainsFile, []byte{}, 0644); err != nil {
			return nil, fmt.Errorf("failed to create domains file: %w", err)
		}
	}

	s := &DomainService{
		domainsFile: config.DomainsFile,
	}

	// Initialize the cache
	if err := s.reloadCache(); err != nil {
		return nil, fmt.Errorf("failed to load initial cache: %w", err)
	}

	// Set up file watcher if enabled
	if config.EnableWatcher {
		watcher, err := NewFileWatcher(config.DomainsFile, s.reloadCache)
		if err != nil {
			return nil, fmt.Errorf("failed to set up file watcher: %w", err)
		}
		s.watcher = watcher
	}

	return s, nil
}

// reloadCache reloads the domain entries from the file into the cache
func (s *DomainService) reloadCache() error {
	entries, err := ReadDomainsFile(s.domainsFile)
	if err != nil {
		return fmt.Errorf("failed to read domains file: %w", err)
	}

	s.mutex.Lock()
	s.cache = entries
	s.mutex.Unlock()

	return nil
}

// Close cleans up resources
func (s *DomainService) Close() error {
	if s.watcher != nil {
		return s.watcher.Close()
	}
	return nil
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

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if domain already exists
	for _, existing := range s.cache {
		if existing.Domain == entry.Domain {
			return nil, fmt.Errorf("domain already exists: %s", entry.Domain)
		}
	}

	// Add the new entry
	s.cache = append(s.cache, entry)

	// Write back to file
	if err := WriteDomainsFile(s.domainsFile, s.cache); err != nil {
		// Revert cache on error
		s.cache = s.cache[:len(s.cache)-1]
		return nil, fmt.Errorf("failed to write domains file: %w", err)
	}

	return &entry, nil
}

// GetDomain retrieves a domain entry
func (s *DomainService) GetDomain(domain string) (*model.DomainEntry, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, entry := range s.cache {
		if entry.Domain == domain {
			entryCopy := entry
			return &entryCopy, nil
		}
	}

	return nil, fmt.Errorf("domain not found: %s", domain)
}

// ListDomains returns all domain entries
func (s *DomainService) ListDomains() ([]model.DomainEntry, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Return a copy of the cache
	entries := make([]model.DomainEntry, len(s.cache))
	copy(entries, s.cache)
	return entries, nil
}

// UpdateDomain updates an existing domain entry
func (s *DomainService) UpdateDomain(domain string, req model.UpdateDomainRequest) (*model.DomainEntry, error) {
	s.mutex.RLock()
	// Make a copy of the current entries
	currentEntries := make([]model.DomainEntry, len(s.cache))
	copy(currentEntries, s.cache)
	s.mutex.RUnlock()

	// Find and update the domain
	found := false
	var updatedEntry model.DomainEntry
	for i, existing := range currentEntries {
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

			currentEntries[i] = updatedEntry
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("domain not found: %s", domain)
	}

	// Write back to file
	if err := WriteDomainsFile(s.domainsFile, currentEntries); err != nil {
		return nil, fmt.Errorf("failed to write domains file: %w", err)
	}

	return &updatedEntry, nil
}

// DeleteDomain removes a domain entry
func (s *DomainService) DeleteDomain(domain string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	found := false
	newEntries := make([]model.DomainEntry, 0, len(s.cache))
	for _, entry := range s.cache {
		if entry.Domain == domain {
			found = true
			continue
		}
		newEntries = append(newEntries, entry)
	}

	if !found {
		return fmt.Errorf("domain not found: %s", domain)
	}

	// Write back to file
	if err := WriteDomainsFile(s.domainsFile, newEntries); err != nil {
		return fmt.Errorf("failed to write domains file: %w", err)
	}

	// Update cache only after successful write
	s.cache = newEntries
	return nil
}
