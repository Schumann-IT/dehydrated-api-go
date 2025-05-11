// Package service provides core business logic for the dehydrated-api-go application.
// It includes domain management, file operations, and plugin integration services.
package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/schumann-it/dehydrated-api-go/internal/plugin/registry"
	"github.com/schumann-it/dehydrated-api-go/internal/util"

	pb "github.com/schumann-it/dehydrated-api-go/plugin/proto"
	"go.uber.org/zap"

	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated"
	"github.com/schumann-it/dehydrated-api-go/internal/model"
)

// DomainService handles domain-related business logic and operations.
// It manages domain entries, integrates with plugins, and provides thread-safe access to domain data.
type DomainService struct {
	DehydratedConfig *dehydrated.Config   // Path to the domains.txt file
	watcher          *FileWatcher         // File watcher for monitoring changes
	cache            []*model.DomainEntry // In-memory cache of domain entries
	mutex            sync.RWMutex         // Mutex for thread-safe access to the cache
	logger           *zap.Logger
	registry         *registry.Registry
}

// NewDomainService creates a new DomainService instance with the provided configuration.
// It initializes the dehydrated client, sets up the plugin registry, and optionally
// enables file watching for automatic updates.
func NewDomainService(cfg *dehydrated.Config, registry *registry.Registry) *DomainService {
	// Ensure the domains file exists
	if _, err := os.Stat(cfg.DomainsFile); err != nil {
		// Create the directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(cfg.DomainsFile), 0755); err != nil {
			panic(err)
		}
		// Create an empty domains file
		if err := os.WriteFile(cfg.DomainsFile, []byte{}, 0644); err != nil {
			panic(err)
		}
	}

	s := &DomainService{
		logger:           zap.NewNop(),
		registry:         registry,
		DehydratedConfig: cfg,
	}

	return s
}

func (s *DomainService) WithLogger(l *zap.Logger) *DomainService {
	s.logger = l
	return s
}

func (s *DomainService) WithFileWatcher() *DomainService {
	s.logger.Info("Enabling file watcher")

	watcher, err := NewFileWatcher(s.DehydratedConfig.DomainsFile, s.Reload)
	if err != nil {
		s.logger.Error("Failed to set up file watcher", zap.Error(err))
		panic(err)
	}
	watcher.WithLogger(s.logger)
	s.watcher = watcher
	s.watcher.Watch()

	s.logger.Info("File watcher is now enabled")

	return s
}

// Reload reloads the domain entries from the file into the cache.
// This method is called during initialization and when file changes are detected.
func (s *DomainService) Reload() error {
	s.logger.Info("Reloading domains file")

	entries, err := ReadDomainsFile(s.DehydratedConfig.DomainsFile)
	if err != nil {
		s.logger.Error("Failed to read domains file", zap.Error(err))
		return err
	}

	// Convert entries to pointers
	pointerEntries := make([]*model.DomainEntry, len(entries))
	for i := range entries {
		pointerEntries[i] = &entries[i]
	}

	s.mutex.Lock()
	s.cache = pointerEntries
	s.mutex.Unlock()

	s.logger.Info("Entries reloaded", zap.Int("count", len(pointerEntries)))

	return nil
}

// Close cleans up resources used by the DomainService.
// It stops the file watcher and closes all plugin connections.
func (s *DomainService) Close() error {
	s.logger.Info("Closing domain service")

	if s.watcher != nil {
		if err := s.watcher.Close(); err != nil {
			s.logger.Error("Failed to  close watcher", zap.Error(err))
		}
	}

	s.logger.Sync()

	return nil
}

// CreateDomain adds a new domain entry to the domains file.
// It validates the entry, checks for duplicates, and updates both the cache and file.
func (s *DomainService) CreateDomain(req model.CreateDomainRequest) (*model.DomainEntry, error) {
	s.logger.Info("Creating domain", zap.Any("domain", req))

	entry := &model.DomainEntry{
		DomainEntry: pb.DomainEntry{
			Domain:           req.Domain,
			AlternativeNames: req.AlternativeNames,
			Alias:            req.Alias,
			Enabled:          req.Enabled,
			Comment:          req.Comment,
		},
	}

	// Validate the domain entry
	if !model.IsValidDomainEntry(entry) {
		s.logger.Error("Invalid domain entry", zap.Any("entry", entry))
		return nil, errors.New("invalid domain entry")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if domain already exists
	for _, existing := range s.cache {
		if existing.Domain == entry.Domain {
			s.logger.Error("Domain already exists", zap.Any("entry", entry))
			return nil, errors.New("domain exists")
		}
	}

	// Add the new entry
	s.cache = append(s.cache, entry)

	s.logger.Info("Dumping domains to disk", zap.Int("count", len(s.cache)))

	// Convert pointers to values for file writing
	entries := make([]model.DomainEntry, len(s.cache))
	for i, entry := range s.cache {
		entries[i] = *entry
	}

	// Write back to file
	if err := WriteDomainsFile(s.DehydratedConfig.DomainsFile, entries); err != nil {
		// Revert cache on error
		s.cache = s.cache[:len(s.cache)-1]
		s.logger.Error("Failed to write domains file", zap.Error(err))
		return nil, err
	}

	return entry, nil
}

// enrichMetadata enriches the domain entry with metadata from all enabled plugins.
// It calls each plugin's GetMetadata method and merges the results into the entry.
func (s *DomainService) enrichMetadata(entry *model.DomainEntry) error {
	if entry.Metadata == nil {
		entry.Metadata = pb.NewMetadata()
	}

	for name, plugin := range s.registry.Plugins() {
		resp, err := plugin.GetMetadata(context.Background(), &pb.GetMetadataRequest{
			DomainEntry:      &entry.DomainEntry,
			DehydratedConfig: s.DehydratedConfig.DomainSpecificConfig(entry.PathName()).ToProto(),
		})
		if err != nil {
			return err
		}

		if resp.Metadata != nil {
			entry.Metadata.FromProto(name, resp.Metadata)
		}
	}

	return nil
}

// GetDomain retrieves a domain entry by its domain name.
// It returns a copy of the entry with metadata enriched from plugins.
func (s *DomainService) GetDomain(domain string) (*model.DomainEntry, error) {
	s.logger.Info("Load domain", zap.String("domain", domain))

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, entry := range s.cache {
		if entry.Domain == domain {
			entryCopy := entry
			if err := s.enrichMetadata(entryCopy); err != nil {
				s.logger.Error("failed to enrich metadata", zap.Error(err))
				return nil, err
			}
			return entryCopy, nil
		}
	}

	s.logger.Error("Domain not found", zap.String("domain", domain))

	return nil, errors.New("domain not found")
}

// ListDomains returns all domain entries with their metadata enriched from plugins.
// It returns a copy of the cached entries to prevent modification of the cache.
func (s *DomainService) ListDomains() ([]*model.DomainEntry, error) {
	s.logger.Info("Load domains")

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Return a copy of the cache with enriched metadata
	entries := make([]*model.DomainEntry, len(s.cache))
	for i, entry := range s.cache {
		entries[i] = entry
		if err := s.enrichMetadata(entries[i]); err != nil {
			s.logger.Error("failed to enrich metadata", zap.String("domain", entries[i].Domain), zap.Error(err))
			return nil, err
		}
	}

	s.logger.Info("Loaded domains", zap.Int("count", len(entries)))

	return entries, nil
}

// UpdateDomain updates an existing domain entry with new information.
// It validates the updated entry and writes the changes to both cache and file.
func (s *DomainService) UpdateDomain(domain string, req model.UpdateDomainRequest) (*model.DomainEntry, error) {
	s.logger.Info("Update domain", zap.String("domain", domain), zap.Any("req", req))

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Find and update the domain
	found := false
	var updatedEntry *model.DomainEntry
	for i, existing := range s.cache {
		if existing.Domain == domain {
			alt := existing.AlternativeNames
			if req.AlternativeNames != nil {
				alt = util.StringSlice(req.AlternativeNames)
			}
			alias := existing.Alias
			if req.Alias != nil {
				alias = util.String(req.Alias)
			}
			enabled := existing.Enabled
			if req.Enabled != nil {
				enabled = util.Bool(req.Enabled)
			}
			comment := existing.Comment
			if req.Comment != nil {
				comment = util.String(req.Comment)
			}
			updatedEntry = &model.DomainEntry{
				DomainEntry: pb.DomainEntry{
					Domain:           domain,
					AlternativeNames: alt,
					Alias:            alias,
					Enabled:          enabled,
					Comment:          comment,
				},
			}

			// Validate the updated entry
			if !model.IsValidDomainEntry(updatedEntry) {
				s.logger.Error("Invalid domain entry", zap.Any("entry", updatedEntry))
				return nil, errors.New("invalid domain entry")
			}

			s.cache[i] = updatedEntry
			found = true
			break
		}
	}

	if !found {
		s.logger.Error("Domain not found", zap.String("domain", domain))
		return nil, errors.New("domain not found")
	}

	// Convert pointers to values for file writing
	entries := make([]model.DomainEntry, len(s.cache))
	for i, entry := range s.cache {
		entries[i] = *entry
	}

	// Write back to file
	s.logger.Info("Dumping domains to disk", zap.Int("count", len(s.cache)))
	if err := WriteDomainsFile(s.DehydratedConfig.DomainsFile, entries); err != nil {
		s.logger.Error("Failed to write domains file", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Updated domain", zap.String("domain", domain))

	return updatedEntry, nil
}

// DeleteDomain removes a domain entry from both the cache and the domains file.
// It returns an error if the domain is not found.
func (s *DomainService) DeleteDomain(domain string) error {
	s.logger.Info("Delete domain", zap.String("domain", domain))

	s.mutex.Lock()
	defer s.mutex.Unlock()

	found := false
	newEntries := make([]*model.DomainEntry, 0, len(s.cache))
	for _, entry := range s.cache {
		if entry.Domain == domain {
			found = true
			continue
		}
		newEntries = append(newEntries, entry)
	}

	if !found {
		s.logger.Error("Domain not found", zap.String("domain", domain))
		return errors.New("domain not found")
	}

	// Convert pointers to values for file writing
	entries := make([]model.DomainEntry, len(newEntries))
	for i, entry := range newEntries {
		entries[i] = *entry
	}

	// Write back to file
	s.logger.Info("Dumping domains to disk", zap.Int("count", len(entries)))
	if err := WriteDomainsFile(s.DehydratedConfig.DomainsFile, entries); err != nil {
		s.logger.Error("Failed to write domains file", zap.Error(err))
		return err
	}

	// Update cache only after successful write
	s.cache = newEntries

	s.logger.Info("Deleted domain", zap.String("domain", domain))

	return nil
}
