// Package model provides data structures and validation logic for the dehydrated-api-go application.
// It includes domain entry models, request/response structures, and protobuf conversion utilities.
package model

import (
	"encoding/json"
	"sort"

	"github.com/schumann-it/dehydrated-api-go/internal/dehydrated"

	pb "github.com/schumann-it/dehydrated-api-go/plugin/proto"
)

// DomainEntries is a slice of DomainEntry pointers that provides convenient methods for manipulation.
type DomainEntries []*DomainEntry

// Sort sorts the domain entries alphabetically by domain name.
// Entries for the same domain are grouped together, with non-aliased entries first,
// followed by aliased entries for that domain.
// The sorting considers only the domain name and alias, ignoring alternative names and comments.
// This method modifies the slice in-place.
func (e DomainEntries) Sort() {
	sort.Slice(e, func(i, j int) bool {
		// Primary sort: domain name
		if e[i].Domain != e[j].Domain {
			return e[i].Domain < e[j].Domain
		}

		// Secondary sort: within same domain, no alias comes first
		hasAliasI := e[i].Alias != ""
		hasAliasJ := e[j].Alias != ""

		if hasAliasI != hasAliasJ {
			return !hasAliasI // No alias comes first
		}

		// Tertiary sort: if both have aliases, sort by alias name
		return e[i].Alias < e[j].Alias
	})
}

// DomainEntry represents a domain configuration entry in the dehydrated system.
// It contains all the necessary information for managing a domain's SSL certificate.
// @Description Domain configuration entry for SSL certificate management
type DomainEntry struct {
	pb.DomainEntry

	// Metadata contains additional information about the domain entry.
	// @Description Additional metadata about the domain entry
	Metadata *pb.Metadata `json:"metadata,omitempty"`
}

// MarshalJSON implements the json.Marshaler interface to ensure all fields are included
func (e *DomainEntry) MarshalJSON() ([]byte, error) {
	metadata := make(map[string]any)
	if e.Metadata != nil {
		protoMap, err := e.Metadata.ToProto()
		if err != nil {
			return nil, err
		}
		for k, v := range protoMap {
			metadata[k] = v.AsInterface()
		}
	}

	return json.Marshal(map[string]any{
		"domain":            e.GetDomain(),
		"alternative_names": e.GetAlternativeNames(),
		"alias":             e.GetAlias(),
		"enabled":           e.GetEnabled(),
		"comment":           e.GetComment(),
		"metadata":          metadata,
	})
}

func (e *DomainEntry) Equals(entry *DomainEntry) bool {
	if e == nil || entry == nil {
		return false
	}

	if e.Domain != entry.Domain ||
		e.Alias != entry.Alias ||
		e.Enabled != entry.Enabled ||
		e.Comment != entry.Comment {
		return false
	}

	if len(e.AlternativeNames) != len(entry.AlternativeNames) {
		return false
	}

	for i, name := range e.AlternativeNames {
		if name != entry.AlternativeNames[i] {
			return false
		}
	}

	return true
}

func (e *DomainEntry) SetMetadata(m *pb.Metadata) {
	e.Metadata = m
}

func (e *DomainEntry) PathName() string {
	n := e.Domain
	if e.Alias != "" {
		n = e.Alias
	}

	return n
}

// CreateDomainRequest represents a request to create a new domain entry.
// It contains all the necessary fields to create a new domain configuration.
// @Description Request to create a new domain entry
type CreateDomainRequest struct {
	// Domain is the primary domain name (required).
	// @Description Primary domain name (required)
	// @required
	Domain string `json:"domain" validate:"required" example:"example.com"`

	// AlternativeNames is a list of additional domain names.
	// @Description List of additional domain names (e.g., "www.example.com")
	AlternativeNames []string `json:"alternative_names,omitempty" example:"www.example.com,api.example.com"`

	// Alias is an optional alternative identifier.
	// @Description Optional alternative identifier for the domain
	Alias string `json:"alias,omitempty" example:"my-domain"`

	// Enabled indicates whether the domain should be active.
	// @Description Whether the domain is enabled for certificate issuance
	Enabled bool `json:"enabled" example:"true"`

	// Comment is an optional description.
	// @Description Optional description or comment for the domain
	Comment string `json:"comment,omitempty" example:"Production domain for web application"`
}

// UpdateDomainRequest represents a request to update an existing domain entry.
// It contains the fields that can be modified for an existing domain.
// @Description Request to update an existing domain entry
type UpdateDomainRequest struct {
	// AlternativeNames is a list of additional domain names.
	// @Description List of additional domain names (e.g., "www.example.com")
	AlternativeNames *[]string `json:"alternative_names,omitempty" example:"www.example.com,api.example.com"`

	// Alias is an optional alternative identifier.
	// @Description Optional alternative identifier for the domain
	Alias *string `json:"alias,omitempty" example:"my-domain"`

	// Enabled indicates whether the domain should be active.
	// @Description Whether the domain is enabled for certificate issuance
	Enabled *bool `json:"enabled,omitempty" example:"true"`

	// Comment is an optional description.
	// @Description Optional description or comment for the domain
	Comment *string `json:"comment,omitempty" example:"Production domain for web application"`
}

// DeleteDomainRequest represents a request to delete an existing domain entry.
// An optional alias can be provided to uniquely identify the domain entry.
// @Description Request to delete an existing domain entry
type DeleteDomainRequest struct {
	// Alias is an optional alternative identifier.
	// @Description Optional alternative identifier for the domain
	Alias *string `json:"alias,omitempty" example:"my-domain"`
}

// DomainResponse represents a response containing a single domain entry.
// It includes a success flag, the domain data, and an optional error message.
// @Description Response containing a single domain entry
type DomainResponse struct {
	// Success indicates whether the operation was successful.
	// @Description Whether the operation was successful
	Success bool `json:"success" example:"true"`

	// Data contains the domain entry if the operation was successful.
	// @Description Domain entry data if the operation was successful
	Data *DomainEntry `json:"data,omitempty"`

	// Error contains an error message if the operation failed.
	// @Description Error message if the operation failed
	Error string `json:"error,omitempty" example:"Domain not found"`
}

// DomainsResponse represents a response containing multiple domain entries.
// It includes a success flag, a list of domain data, and an optional error message.
// @Description Response containing multiple domain entries
type DomainsResponse struct {
	// Success indicates whether the operation was successful.
	// @Description Whether the operation was successful
	Success bool `json:"success" example:"true"`

	// Data contains the list of domain entries if the operation was successful.
	// @Description List of domain entries if the operation was successful
	Data DomainEntries `json:"data,omitempty"`

	// Error contains an error message if the operation failed.
	// @Description Error message if the operation failed
	Error string `json:"error,omitempty" example:"Failed to load domains"`
}

type ConfigResponse struct {
	Success bool `json:"success" example:"true"`

	Data *dehydrated.Config `json:"data,omitempty"`

	Error string `json:"error,omitempty" example:"Failed to load config"`
}

// Pagination constants
const (
	DefaultPerPage = 100
	MaxPerPage     = 1000
	MinPerPage     = 1
	MinPage        = 1
)

// PaginationInfo contains pagination metadata for responses
// @Description Pagination metadata for responses
type PaginationInfo struct {
	// CurrentPage is the current page number (1-based)
	// @Description Current page number (1-based)
	CurrentPage int `json:"current_page" example:"2"`

	// PerPage is the number of items per page
	// @Description Number of items per page
	PerPage int `json:"per_page" example:"100"`

	// Total is the total number of items
	// @Description Total number of items
	Total int `json:"total" example:"150"`

	// TotalPages is the total number of pages
	// @Description Total number of pages
	TotalPages int `json:"total_pages" example:"2"`

	// HasNext indicates if there is a next page
	// @Description Whether there is a next page
	HasNext bool `json:"has_next" example:"true"`

	// HasPrev indicates if there is a previous page
	// @Description Whether there is a previous page
	HasPrev bool `json:"has_prev" example:"true"`

	// NextURL is the URL for the next page
	// @Description URL for the next page
	NextURL string `json:"next_url,omitempty" example:"/api/v1/domains?page=3&per_page=100"`

	// PrevURL is the URL for the previous page
	// @Description URL for the previous page
	PrevURL string `json:"prev_url,omitempty" example:"/api/v1/domains?page=1&per_page=100"`
}

// PaginatedDomainsResponse represents a paginated response containing multiple domain entries
// @Description Paginated response containing multiple domain entries
type PaginatedDomainsResponse struct {
	// Success indicates whether the operation was successful
	// @Description Whether the operation was successful
	Success bool `json:"success" example:"true"`

	// Data contains the list of domain entries if the operation was successful
	// @Description List of domain entries if the operation was successful
	Data DomainEntries `json:"data,omitempty"`

	// Pagination contains pagination metadata
	// @Description Pagination metadata
	Pagination *PaginationInfo `json:"pagination,omitempty"`

	// Error contains an error message if the operation failed
	// @Description Error message if the operation failed
	Error string `json:"error,omitempty" example:"Failed to load domains"`
}
