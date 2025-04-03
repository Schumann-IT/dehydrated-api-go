package model

import (
	pb "github.com/schumann-it/dehydrated-api-go/proto/plugin"
	"google.golang.org/protobuf/types/known/structpb"
)

// DomainEntry represents a domain configuration entry
type DomainEntry struct {
	// Domain name
	Domain string `json:"domain" protobuf:"bytes,1,opt,name=domain,proto3"`

	// Alternative names
	AlternativeNames []string `json:"alternative_names,omitempty" protobuf:"bytes,2,rep,name=alternative_names,json=alternativeNames,proto3"`

	// Alias
	Alias string `json:"alias,omitempty" protobuf:"bytes,3,opt,name=alias,proto3"`

	// Whether the domain is enabled
	Enabled bool `json:"enabled" protobuf:"varint,4,opt,name=enabled,proto3"`

	// Comment
	Comment string `json:"comment,omitempty" protobuf:"bytes,5,opt,name=comment,proto3"`

	// Metadata contains additional information about the domain entry.
	// During runtime, this is stored as map[string]any for flexibility.
	// During serialization (protobuf), this is converted to map[string]*structpb.Value.
	// The conversion is handled by ToProto() and FromProto() methods.
	Metadata map[string]any `json:"metadata,omitempty" protobuf:"bytes,6,rep,name=metadata,proto3"`
}

// ToProto converts the DomainEntry to a protobuf GetMetadataRequest
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

// FromProto creates a DomainEntry from a protobuf GetMetadataResponse
func FromProto(resp *pb.GetMetadataResponse) *DomainEntry {
	metadata := make(map[string]any)
	for k, v := range resp.Metadata {
		metadata[k] = v.AsInterface()
	}

	return &DomainEntry{
		Metadata: metadata,
	}
}

// CreateDomainRequest represents a request to create a new domain entry
type CreateDomainRequest struct {
	Domain           string            `json:"domain" validate:"required"`
	AlternativeNames []string          `json:"alternative_names,omitempty"`
	Alias            string            `json:"alias,omitempty"`
	Enabled          bool              `json:"enabled"`
	Comment          string            `json:"comment,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}

// UpdateDomainRequest represents a request to update an existing domain entry
type UpdateDomainRequest struct {
	AlternativeNames []string          `json:"alternative_names,omitempty"`
	Alias            string            `json:"alias,omitempty"`
	Enabled          bool              `json:"enabled"`
	Comment          string            `json:"comment,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}

// DomainResponse represents a response containing a single domain entry
type DomainResponse struct {
	Success bool        `json:"success"`
	Data    DomainEntry `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// DomainsResponse represents a response containing multiple domain entries
type DomainsResponse struct {
	Success bool          `json:"success"`
	Data    []DomainEntry `json:"data,omitempty"`
	Error   string        `json:"error,omitempty"`
}
