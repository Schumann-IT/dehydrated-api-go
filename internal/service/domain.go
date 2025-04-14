// Package service provides core business logic for the dehydrated-api-go application.
// It includes domain management, file operations, and plugin integration services.
package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated"
	"github.com/schumann-it/dehydrated-api-go/internal/model"
	"github.com/schumann-it/dehydrated-api-go/internal/plugin"
	"github.com/schumann-it/dehydrated-api-go/internal/plugin/registry"
)

// DomainService handles domain-related business logic and operations.
// It manages domain entries, integrates with plugins, and provides thread-safe access to domain data.
type DomainService struct {
	domainsFile string              // Path to the domains.txt file
	watcher     *FileWatcher        // File watcher for monitoring changes
	cache       []model.DomainEntry // In-memory cache of domain entries
	mutex       sync.RWMutex        // Mutex for thread-safe access to the cache
	Registry    *registry.Registry  // Plugin registry for metadata enrichment
}

// NewDomainService creates a new DomainService instance with the provided configuration.
// It initializes the dehydrated client, sets up the plugin registry, and optionally
// enables file watching for automatic updates.
func NewDomainService(domainsFile string) *DomainService {
	// Ensure the domains file exists
	if _, err := os.Stat(domainsFile); err != nil {
		// Create the directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(domainsFile), 0755); err != nil {
			panic(err)
		}
		// Create an empty domains file
		if err := os.WriteFile(domainsFile, []byte{}, 0644); err != nil {
			panic(err)
		}
	}

	s := &DomainService{
		domainsFile: domainsFile,
	}

	return s
}

func (s *DomainService) WithPlugins(plugins map[string]plugin.PluginConfig, cfg *dehydrated.Config) *DomainService {
	reg, err := registry.NewRegistry(plugins, cfg)
	if err != nil {
		panic(err)
	}
	s.Registry = reg
	return s
}

func (s *DomainService) WithFileWatcher() *DomainService {
	watcher, err := NewFileWatcher(s.domainsFile, s.Reload)
	if err != nil {
		panic(err)
	}
	s.watcher = watcher

	return s
}

// Reload reloads the domain entries from the file into the cache.
// This method is called during initialization and when file changes are detected.
func (s *DomainService) Reload() error {
	entries, err := ReadDomainsFile(s.domainsFile)
	if err != nil {
		return fmt.Errorf("failed to read domains file: %w", err)
	}

	s.mutex.Lock()
	s.cache = entries
	s.mutex.Unlock()

	return nil
}

// Close cleans up resources used by the DomainService.
// It stops the file watcher and closes all plugin connections.
func (s *DomainService) Close() error {
	var errs []error

	if s.watcher != nil {
		if err := s.watcher.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close watcher: %w", err))
		}
	}

	if s.Registry != nil {
		if err := s.Registry.Close(context.Background()); err != nil {
			errs = append(errs, fmt.Errorf("failed to close plugin Registry: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing domain service: %v", errs)
	}
	return nil
}

// CreateDomain adds a new domain entry to the domains file.
// It validates the entry, checks for duplicates, and updates both the cache and file.
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

// enrichMetadata enriches the domain entry with metadata from all enabled plugins.
// It calls each plugin's GetMetadata method and merges the results into the entry.
func (s *DomainService) enrichMetadata(entry *model.DomainEntry) error {
	if s.Registry == nil {
		return nil
	}

	ctx := context.Background()
	for name, p := range s.Registry.GetPlugins() {
		metadata, err := p.GetMetadata(ctx, *entry)
		if err != nil {
			return fmt.Errorf("failed to get metadata from plugin %s: %w", name, err)
		}
		if entry.Metadata == nil {
			entry.Metadata = make(map[string]any)
		}
		entry.Metadata[name] = metadata
	}
	return nil
}

// GetDomain retrieves a domain entry by its domain name.
// It returns a copy of the entry with metadata enriched from plugins.
func (s *DomainService) GetDomain(domain string) (*model.DomainEntry, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, entry := range s.cache {
		if entry.Domain == domain {
			entryCopy := entry
			if err := s.enrichMetadata(&entryCopy); err != nil {
				return nil, fmt.Errorf("failed to enrich metadata: %w", err)
			}
			return &entryCopy, nil
		}
	}

	return nil, fmt.Errorf("domain not found: %s", domain)
}

// ListDomains returns all domain entries with their metadata enriched from plugins.
// It returns a copy of the cached entries to prevent modification of the cache.
func (s *DomainService) ListDomains() ([]model.DomainEntry, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Return a copy of the cache with enriched metadata
	entries := make([]model.DomainEntry, len(s.cache))
	for i, entry := range s.cache {
		entries[i] = entry
		if err := s.enrichMetadata(&entries[i]); err != nil {
			return nil, fmt.Errorf("failed to enrich metadata: %w", err)
		}
	}

	return entries, nil
}

// UpdateDomain updates an existing domain entry with new information.
// It validates the updated entry and writes the changes to both cache and file.
func (s *DomainService) UpdateDomain(domain string, req model.UpdateDomainRequest) (*model.DomainEntry, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Find and update the domain
	found := false
	var updatedEntry model.DomainEntry
	for i, existing := range s.cache {
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

			s.cache[i] = updatedEntry
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("domain not found: %s", domain)
	}

	// Write back to file
	if err := WriteDomainsFile(s.domainsFile, s.cache); err != nil {
		return nil, fmt.Errorf("failed to write domains file: %w", err)
	}

	return &updatedEntry, nil
}

// DeleteDomain removes a domain entry from both the cache and the domains file.
// It returns an error if the domain is not found.
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
