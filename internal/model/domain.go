// Package model provides data structures and validation logic for the dehydrated-api-go application.
// It includes domain entry models, request/response structures, and protobuf conversion utilities.
package model

import (
	pb "github.com/schumann-it/dehydrated-api-go/proto/plugin"
)

// DomainEntry represents a domain configuration entry in the dehydrated system.
// It contains all the necessary information for managing a domain's SSL certificate.
type DomainEntry struct {
	pb.DomainEntry

	// Metadata contains additional information about the domain entry.
	// During runtime, this is stored as map[string]any for flexibility.
	// During serialization (protobuf), this is converted to map[string]*structpb.Value.
	// The conversion is handled by ToProto() and FromProto() methods.
	Metadata map[string]any `json:"metadata,omitempty" protobuf:"bytes,6,rep,name=metadata,proto3"`
}

func (e *DomainEntry) PathName() string {
	n := e.Domain
	if e.Alias != "" {
		n = e.Alias
	}

	return n
}

// FromProto creates a DomainEntry from a protobuf GetMetadataResponse.
// It converts the protobuf metadata values back to their original types.
// Returns a new DomainEntry with all fields populated from the response.
func FromProto(resp *pb.GetMetadataResponse) *DomainEntry {
	metadata := make(map[string]any)
	for k, v := range resp.Metadata {
		metadata[k] = v.AsInterface()
	}

	return &DomainEntry{
		Metadata: metadata,
	}
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

	// Metadata contains additional domain-specific information.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// UpdateDomainRequest represents a request to update an existing domain entry.
// It contains the fields that can be modified for an existing domain.
type UpdateDomainRequest struct {
	// AlternativeNames is a list of additional domain names.
	AlternativeNames []string `json:"alternative_names,omitempty"`

	// Alias is an optional alternative identifier.
	Alias string `json:"alias,omitempty"`

	// Enabled indicates whether the domain should be active.
	Enabled bool `json:"enabled"`

	// Comment is an optional description.
	Comment string `json:"comment,omitempty"`

	// Metadata contains additional domain-specific information.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// DomainResponse represents a response containing a single domain entry.
// It includes a success flag, the domain data, and an optional error message.
type DomainResponse struct {
	// Success indicates whether the operation was successful.
	Success bool `json:"success"`

	// Data contains the domain entry if the operation was successful.
	Data DomainEntry `json:"data,omitempty"`

	// Error contains an error message if the operation failed.
	Error string `json:"error,omitempty"`
}

// DomainsResponse represents a response containing multiple domain entries.
// It includes a success flag, a list of domain data, and an optional error message.
type DomainsResponse struct {
	// Success indicates whether the operation was successful.
	Success bool `json:"success"`

	// Data contains the list of domain entries if the operation was successful.
	Data []DomainEntry `json:"data,omitempty"`

	// Error contains an error message if the operation failed.
	Error string `json:"error,omitempty"`
}
