// Package model provides data structures and validation logic for the dehydrated-api-go application.
// It includes domain entry models, request/response structures, and protobuf conversion utilities.
package model

import (
	"encoding/json"

	pb "github.com/schumann-it/dehydrated-api-go/plugin/proto"
)

// DomainEntry represents a domain configuration entry in the dehydrated system.
// It contains all the necessary information for managing a domain's SSL certificate.
type DomainEntry struct {
	pb.DomainEntry

	// Metadata contains additional information about the domain entry.
	Metadata *pb.Metadata `json:"metadata,omitempty"`
}

// MarshalJSON implements the json.Marshaler interface to ensure all fields are included
func (e *DomainEntry) MarshalJSON() ([]byte, error) {
	metadata := make(map[string]interface{})
	if e.Metadata != nil {
		protoMap, err := e.Metadata.ToProto()
		if err != nil {
			return nil, err
		}
		for k, v := range protoMap {
			metadata[k] = v.AsInterface()
		}
	}

	return json.Marshal(map[string]interface{}{
		"domain":            e.GetDomain(),
		"alternative_names": e.GetAlternativeNames(),
		"alias":             e.GetAlias(),
		"enabled":           e.GetEnabled(),
		"comment":           e.GetComment(),
		"metadata":          metadata,
	})
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
type CreateDomainRequest struct {
	// Domain is the primary domain name (required).
	Domain string `json:"domain" validate:"required"`

	// AlternativeNames is a list of additional domain names.
	AlternativeNames []string `json:"alternative_names,omitempty"`

	// Alias is an optional alternative identifier.
	Alias string `json:"alias,omitempty"`

	// Enabled indicates whether the domain should be active.
	Enabled bool `json:"enabled"`

	// Comment is an optional description.
	Comment string `json:"comment,omitempty"`
}

// UpdateDomainRequest represents a request to update an existing domain entry.
// It contains the fields that can be modified for an existing domain.
type UpdateDomainRequest struct {
	// AlternativeNames is a list of additional domain names.
	AlternativeNames *[]string `json:"alternative_names,omitempty"`

	// Alias is an optional alternative identifier.
	Alias *string `json:"alias,omitempty"`

	// Enabled indicates whether the domain should be active.
	Enabled *bool `json:"enabled,omitempty"`

	// Comment is an optional description.
	Comment *string `json:"comment,omitempty"`
}

// DomainResponse represents a response containing a single domain entry.
// It includes a success flag, the domain data, and an optional error message.
type DomainResponse struct {
	// Success indicates whether the operation was successful.
	Success bool `json:"success"`

	// Data contains the domain entry if the operation was successful.
	Data *DomainEntry `json:"data,omitempty"`

	// Error contains an error message if the operation failed.
	Error string `json:"error,omitempty"`
}

// DomainsResponse represents a response containing multiple domain entries.
// It includes a success flag, a list of domain data, and an optional error message.
type DomainsResponse struct {
	// Success indicates whether the operation was successful.
	Success bool `json:"success"`

	// Data contains the list of domain entries if the operation was successful.
	Data []*DomainEntry `json:"data,omitempty"`

	// Error contains an error message if the operation failed.
	Error string `json:"error,omitempty"`
}
