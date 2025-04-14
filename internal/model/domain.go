// Package model provides data structures and validation logic for the dehydrated-api-go application.
// It includes domain entry models, request/response structures, and protobuf conversion utilities.
package model

import (
	pb "github.com/schumann-it/dehydrated-api-go/proto/plugin"
	"google.golang.org/protobuf/types/known/structpb"
)

// DomainEntry represents a domain configuration entry in the dehydrated system.
// It contains all the necessary information for managing a domain's SSL certificate.
type DomainEntry struct {
	// Domain is the primary domain name for the certificate.
	Domain string `json:"domain" protobuf:"bytes,1,opt,name=domain,proto3"`

	// AlternativeNames is a list of additional domain names to be included in the certificate.
	AlternativeNames []string `json:"alternative_names,omitempty" protobuf:"bytes,2,rep,name=alternative_names,json=alternativeNames,proto3"`

	// Alias is an optional alternative identifier for the domain.
	Alias string `json:"alias,omitempty" protobuf:"bytes,3,opt,name=alias,proto3"`

	// Enabled indicates whether the domain is currently active for certificate management.
	Enabled bool `json:"enabled" protobuf:"varint,4,opt,name=enabled,proto3"`

	// Comment is an optional description or note about the domain entry.
	Comment string `json:"comment,omitempty" protobuf:"bytes,5,opt,name=comment,proto3"`

	// Metadata contains additional information about the domain entry.
	// During runtime, this is stored as map[string]any for flexibility.
	// During serialization (protobuf), this is converted to map[string]*structpb.Value.
	// The conversion is handled by ToProto() and FromProto() methods.
	Metadata map[string]any `json:"metadata,omitempty" protobuf:"bytes,6,rep,name=metadata,proto3"`
}

// ToProto converts the DomainEntry to a protobuf GetMetadataRequest.
// It handles the conversion of metadata values to protobuf struct values.
// Returns a new GetMetadataRequest with all fields populated from the DomainEntry.
func (e *DomainEntry) ToProto() *pb.GetMetadataRequest {
	metadata := make(map[string]*structpb.Value)
	for k, v := range e.Metadata {
		value, err := structpb.NewValue(v)
		if err == nil {
			metadata[k] = value
		}
	}

	return &pb.GetMetadataRequest{
		Domain:           e.Domain,
		AlternativeNames: e.AlternativeNames,
		Alias:            e.Alias,
		Enabled:          e.Enabled,
		Comment:          e.Comment,
		Metadata:         metadata,
	}
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
