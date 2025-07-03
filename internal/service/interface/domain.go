package serviceinterface

import "github.com/schumann-it/dehydrated-api-go/internal/model"

// DomainService defines the interface for domain operations.
// It provides methods for managing domain entries in the dehydrated configuration.
type DomainService interface {
	// ListDomains returns all configured domain entries.
	ListDomains() ([]*model.DomainEntry, error)

	// GetDomain retrieves a specific domain entry by its domain name.
	// If multiple entries exist with the same domain, returns the first match.
	GetDomain(domain string) (*model.DomainEntry, error)

	// GetDomainByAlias retrieves a specific domain entry by its domain name and optional alias.
	// This is useful when multiple entries exist with the same domain but different aliases.
	// If alias is empty, behaves the same as GetDomain.
	GetDomainByAlias(domain string, alias string) (*model.DomainEntry, error)

	// CreateDomain creates a new domain entry with the given configuration.
	CreateDomain(req *model.CreateDomainRequest) (*model.DomainEntry, error)

	// UpdateDomain updates an existing domain entry with the given configuration.
	UpdateDomain(domain string, req model.UpdateDomainRequest) (*model.DomainEntry, error)

	// UpdateDomainByAlias updates an existing domain entry by its domain name and optional alias.
	// This is useful when multiple entries exist with the same domain but different aliases.
	// If alias is empty, behaves the same as UpdateDomain.
	UpdateDomainByAlias(domain string, alias string, req model.UpdateDomainRequest) (*model.DomainEntry, error)

	// DeleteDomain removes a domain entry by its domain name.
	DeleteDomain(domain string) error

	// DeleteDomainByAlias removes a domain entry by its domain name and optional alias.
	// This is useful when multiple entries exist with the same domain but different aliases.
	// If alias is empty, behaves the same as DeleteDomain.
	DeleteDomainByAlias(domain string, alias string) error

	// Close performs any necessary cleanup when the service is no longer needed.
	Close() error
}
