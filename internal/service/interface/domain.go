package serviceinterface

import "github.com/schumann-it/dehydrated-api-go/internal/model"

// DomainService defines the interface for domain operations.
// It provides methods for managing domain entries in the dehydrated configuration.
type DomainService interface {
	// ListDomains returns paginated domain entries with pagination metadata.
	// page and perPage are 1-based. If page is 0 or negative, it defaults to 1.
	// If perPage is 0 or negative, it defaults to DefaultPerPage (100).
	// If perPage exceeds MaxPerPage (1000), it is capped to MaxPerPage.
	// sortOrder can be "asc" or "desc" to sort by domain field (optional - defaults to alphabetical order).
	// search is an optional search term to filter domains by domain field using contains().
	ListDomains(page, perPage int, sortOrder, search string) ([]*model.DomainEntry, *model.PaginationInfo, error)

	// GetDomain retrieves a specific domain entry by its domain name.
	// If multiple entries exist with the same domain, returns the first match.
	GetDomain(domain, alias string) (*model.DomainEntry, error)

	// CreateDomain creates a new domain entry with the given configuration.
	CreateDomain(req *model.CreateDomainRequest) (*model.DomainEntry, error)

	// UpdateDomain updates an existing domain entry with the given configuration.
	UpdateDomain(domain string, req model.UpdateDomainRequest) (*model.DomainEntry, error)

	// DeleteDomain removes a domain entry by its domain name.
	DeleteDomain(domain string, req model.DeleteDomainRequest) error

	// Close performs any necessary cleanup when the service is no longer needed.
	Close() error
}
