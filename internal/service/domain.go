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
func NewDomainService(cfg *dehydrated.Config, r *registry.Registry) *DomainService {
	// Ensure the domains file exists
	if _, err := os.Stat(cfg.DomainsFile); err != nil {
		// Create the directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(cfg.DomainsFile), 0755); err != nil {
			panic(err)
		}
		// Create an empty domains file
		//nolint:gosec // This is a safe operation, we just want to ensure the file exists
		if err := os.WriteFile(cfg.DomainsFile, []byte{}, 0644); err != nil {
			panic(err)
		}
	}

	s := &DomainService{
		logger:           zap.NewNop(),
		registry:         r,
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
	copy(pointerEntries, entries)

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

// findDomainEntry finds a domain entry in the cache by domain and optional alias.
// If alias is empty, it looks for entries without an alias.
func (s *DomainService) findDomainEntry(domain, alias string) (*model.DomainEntry, int) {
	for i, entry := range s.cache {
		if entry.Domain == domain && entry.Alias == alias {
			return entry, i
		}
	}
	return nil, -1
}

// writeCacheToFile writes the current cache to the domains file.
// It converts pointer entries to values for file writing.
func (s *DomainService) writeCacheToFile() error {
	entries := make(model.DomainEntries, 0, len(s.cache))
	for _, entry := range s.cache {
		entries = append(entries, &model.DomainEntry{
			DomainEntry: pb.DomainEntry{
				Domain:           entry.Domain,
				AlternativeNames: entry.AlternativeNames,
				Alias:            entry.Alias,
				Enabled:          entry.Enabled,
				Comment:          entry.Comment,
			},
		})
	}

	s.logger.Info("Dumping domains to disk", zap.Int("count", len(s.cache)))
	return WriteDomainsFile(s.DehydratedConfig.DomainsFile, entries)
}

// writeEntriesToFile writes a specific set of domain entries to the domains file.
// It converts pointer entries to values for file writing.
func (s *DomainService) writeEntriesToFile(entries []*model.DomainEntry) error {
	// Convert pointers to values for file writing
	valueEntries := make(model.DomainEntries, 0, len(entries))
	for _, entry := range entries {
		valueEntries = append(valueEntries, &model.DomainEntry{
			DomainEntry: pb.DomainEntry{
				Domain:           entry.Domain,
				AlternativeNames: entry.AlternativeNames,
				Alias:            entry.Alias,
				Enabled:          entry.Enabled,
				Comment:          entry.Comment,
			},
		})
	}

	s.logger.Info("Dumping domains to disk", zap.Int("count", len(entries)))
	return WriteDomainsFile(s.DehydratedConfig.DomainsFile, valueEntries)
}

// updateEntry creates a new domain entry with updated fields from the request.
// It preserves existing values for fields that are not provided in the request.
func updateEntry(entry *model.DomainEntry, req model.UpdateDomainRequest) *model.DomainEntry {
	alt := entry.AlternativeNames
	if req.AlternativeNames != nil {
		alt = util.StringSlice(req.AlternativeNames)
	}

	enabled := entry.Enabled
	if req.Enabled != nil {
		enabled = util.Bool(req.Enabled)
	}

	comment := entry.Comment
	if req.Comment != nil {
		comment = util.String(req.Comment)
	}

	return &model.DomainEntry{
		DomainEntry: pb.DomainEntry{
			Domain:           entry.Domain,
			AlternativeNames: alt,
			Alias:            entry.Alias,
			Enabled:          enabled,
			Comment:          comment,
		},
	}
}

// entriesWithout retrieves all domain entries from the cache except for the specified domain and alias.
// It also returns whether the domain was found and removed.
func (s *DomainService) entriesWithout(domain string, alias *string) ([]*model.DomainEntry, bool) {
	found := false
	newEntries := make([]*model.DomainEntry, 0, len(s.cache))
	for _, entry := range s.cache {
		if alias != nil && *alias != "" {
			if entry.Domain == domain && entry.Alias == *alias {
				found = true
				continue
			}
		} else {
			if entry.Domain == domain && entry.Alias == "" {
				found = true
				continue
			}
		}
		newEntries = append(newEntries, entry)
	}
	return newEntries, found
}

// CreateDomain adds a new domain entry to the domains file.
// It validates the entry, checks for duplicates, and updates both the cache and file.
func (s *DomainService) CreateDomain(req *model.CreateDomainRequest) (*model.DomainEntry, error) {
	s.logger.Info("Creating domain", zap.Any("domain", req.Domain), zap.Any("req", req))

	if s.watcher != nil {
		s.watcher.Disable()
	}

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

	existing, _ := s.findDomainEntry(req.Domain, req.Alias)
	if existing != nil {
		s.logger.Error("Domain already exists", zap.Any("entry", entry))
		return nil, errors.New("domain exists")
	}

	// Add the new entry
	s.cache = append(s.cache, entry)

	// Write back to file
	if err := s.writeCacheToFile(); err != nil {
		// Revert cache on error
		s.cache = s.cache[:len(s.cache)-1]
		s.logger.Error("Failed to write domains file", zap.Error(err))
		return nil, err
	}

	if s.watcher != nil {
		s.watcher.Enable()
	}

	return entry, nil
}

// enrichMetadata enriches the domain entry with metadata from all enabled plugins.
// It calls each plugin's GetMetadata method and merges the results into the entry.
func (s *DomainService) enrichMetadata(entry *model.DomainEntry) {
	if entry.Metadata == nil {
		entry.Metadata = pb.NewMetadata()
	}

	for name, plugin := range s.registry.Plugins() {
		resp, err := plugin.GetMetadata(context.Background(), &pb.GetMetadataRequest{
			DomainEntry:      &entry.DomainEntry,
			DehydratedConfig: s.DehydratedConfig.DomainSpecificConfig(entry.PathName()).ToProto(),
		})

		if err != nil {
			s.logger.Error("plugin request failed", zap.String("plugin", name), zap.String("domain", entry.Domain), zap.Error(err))
			entry.Metadata.SetMap(name, map[string]string{"error": err.Error()})
			continue
		}

		if resp.Error != "" {
			s.logger.Error("plugin request failed", zap.String("plugin", name),
				zap.String("domain", entry.Domain), zap.Error(errors.New(resp.Error)))
			entry.Metadata.SetMap(name, map[string]string{"error": resp.Error})
			continue
		}

		if resp.Metadata != nil {
			entry.Metadata.FromProto(name, resp.Metadata)
		}
	}
}

// GetDomain retrieves a domain entry by its domain name.
// It returns a copy of the entry with metadata enriched from plugins.
func (s *DomainService) GetDomain(domain, alias string) (*model.DomainEntry, error) {
	s.logger.Info("Load domain", zap.String("domain", domain), zap.Any("alias", alias))

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	entry, _ := s.findDomainEntry(domain, alias)
	if entry == nil {
		s.logger.Error("Domain not found", zap.String("domain", domain), zap.Any("alias", alias))
		return nil, errors.New("domain not found")
	}

	entryCopy := entry
	s.enrichMetadata(entryCopy)
	return entryCopy, nil
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
		s.enrichMetadata(entries[i])
	}

	s.logger.Info("Loaded domains", zap.Int("count", len(entries)))

	return entries, nil
}

// UpdateDomain updates an existing domain entry with new information.
// It validates the updated entry and writes the changes to both cache and file.
func (s *DomainService) UpdateDomain(domain string, req model.UpdateDomainRequest) (*model.DomainEntry, error) {
	s.logger.Info("Update domain", zap.String("domain", domain), zap.Any("req", req))

	if s.watcher != nil {
		s.watcher.Disable()
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	alias := ""
	if req.Alias != nil {
		alias = *req.Alias
	}
	entry, index := s.findDomainEntry(domain, alias)
	if entry == nil {
		s.logger.Error("Domain not found", zap.String("domain", domain), zap.Any("req", req))
		return nil, errors.New("domain not found")
	}

	updatedEntry := updateEntry(entry, req)

	// Validate the updated entry
	if !model.IsValidDomainEntry(updatedEntry) {
		s.logger.Error("Invalid domain entry", zap.Any("entry", updatedEntry))
		return nil, errors.New("invalid domain entry")
	}

	if !updatedEntry.Equals(entry) {
		s.cache[index] = updatedEntry

		// Write back to file
		if err := s.writeCacheToFile(); err != nil {
			s.logger.Error("Failed to write domains file", zap.Error(err))
			return nil, err
		}

		s.logger.Info("Updated domain", zap.String("domain", domain), zap.Any("req", req))
	} else {
		s.logger.Info("No changes detected for domain", zap.String("domain", domain), zap.Any("req", req))
	}

	if s.watcher != nil {
		s.watcher.Enable()
	}

	return updatedEntry, nil
}

// DeleteDomain removes a domain entry from both the cache and the domains file.
// It returns an error if the domain is not found.
func (s *DomainService) DeleteDomain(domain string, req model.DeleteDomainRequest) error {
	s.logger.Info("Delete domain", zap.String("domain", domain), zap.Any("req", req))

	if s.watcher != nil {
		s.watcher.Disable()
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	newEntries, found := s.entriesWithout(domain, req.Alias)
	if !found {
		s.logger.Error("Domain without alias not found", zap.String("domain", domain), zap.Any("req", req))
		return errors.New("domain without specified alias not found")
	}

	// Write back to file
	if err := s.writeEntriesToFile(newEntries); err != nil {
		s.logger.Error("Failed to write domains file", zap.Error(err))
		return err
	}

	// Update cache only after successful write
	s.cache = newEntries

	s.logger.Info("Deleted domain", zap.String("domain", domain), zap.Any("req", req))

	if s.watcher != nil {
		s.watcher.Enable()
	}

	return nil
}
