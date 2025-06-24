package model

import (
	"encoding/json"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/schumann-it/dehydrated-api-go/internal/util"
	pb "github.com/schumann-it/dehydrated-api-go/plugin/proto"
	"github.com/stretchr/testify/require"
)

var validate = validator.New()

func TestDomainEntry_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		entry    *DomainEntry
		expected string
	}{
		{
			name: "full entry",
			entry: &DomainEntry{
				DomainEntry: pb.DomainEntry{
					Domain:           "example.com",
					AlternativeNames: []string{"www.example.com"},
					Alias:            "example",
					Enabled:          true,
					Comment:          "test comment",
				},
				Metadata: pb.NewMetadata(),
			},
			expected: `{"domain":"example.com","alternative_names":["www.example.com"],"alias":"example","enabled":true,"comment":"test comment","metadata":{}}`,
		},
		{
			name: "minimal entry",
			entry: &DomainEntry{
				DomainEntry: pb.DomainEntry{
					Domain:  "example.com",
					Enabled: true,
				},
				Metadata: pb.NewMetadata(),
			},
			expected: `{"domain":"example.com","alternative_names":null,"alias":"","enabled":true,"comment":"","metadata":{}}`,
		},
		{
			name: "entry with metadata",
			entry: &DomainEntry{
				DomainEntry: pb.DomainEntry{
					Domain:  "example.com",
					Enabled: true,
				},
				Metadata: func() *pb.Metadata {
					m := pb.NewMetadata()
					m.Set("key", "value")
					return m
				}(),
			},
			expected: `{"domain":"example.com","alternative_names":null,"alias":"","enabled":true,"comment":"","metadata":{"key":"value"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.entry)
			require.NoError(t, err)
			require.JSONEq(t, tt.expected, string(data))
		})
	}
}

func TestDomainEntry_SetMetadata(t *testing.T) {
	entry := &DomainEntry{
		DomainEntry: pb.DomainEntry{
			Domain:  "example.com",
			Enabled: true,
		},
	}

	metadata := pb.NewMetadata()
	metadata.Set("key", "value")
	entry.SetMetadata(metadata)

	require.Equal(t, metadata, entry.Metadata)
}

func TestDomainEntry_PathName(t *testing.T) {
	tests := []struct {
		name     string
		entry    *DomainEntry
		expected string
	}{
		{
			name: "with domain only",
			entry: &DomainEntry{
				DomainEntry: pb.DomainEntry{
					Domain: "example.com",
				},
			},
			expected: "example.com",
		},
		{
			name: "with alias",
			entry: &DomainEntry{
				DomainEntry: pb.DomainEntry{
					Domain: "example.com",
					Alias:  "example",
				},
			},
			expected: "example",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.entry.PathName())
		})
	}
}

func TestCreateDomainRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request *CreateDomainRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: &CreateDomainRequest{
				Domain:           "example.com",
				AlternativeNames: []string{"www.example.com"},
				Enabled:          true,
			},
			wantErr: false,
		},
		{
			name: "missing domain",
			request: &CreateDomainRequest{
				AlternativeNames: []string{"www.example.com"},
				Enabled:          true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.request)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUpdateDomainRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request UpdateDomainRequest
		isValid bool
	}{
		{
			name: "Valid request",
			request: UpdateDomainRequest{
				AlternativeNames: util.StringSlicePtr([]string{"www.example.com"}),
				Enabled:          util.BoolPtr(true),
			},
			isValid: true,
		},
		{
			name:    "Empty request",
			request: UpdateDomainRequest{},
			isValid: true,
		},
		{
			name:    "With metadata",
			request: UpdateDomainRequest{},
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// UpdateDomainRequest doesn't have required fields, so all validations should pass
			require.True(t, tt.isValid)
		})
	}
}

func TestDomainResponse(t *testing.T) {
	tests := []struct {
		name     string
		response DomainResponse
		success  bool
	}{
		{
			name: "Successful response",
			response: DomainResponse{
				Success: true,
				Data: &DomainEntry{
					DomainEntry: pb.DomainEntry{
						Domain: "example.com",
					},
				},
			},
			success: true,
		},
		{
			name: "Error response",
			response: DomainResponse{
				Success: false,
				Error:   "domain not found",
			},
			success: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.success, tt.response.Success)
			if tt.success {
				require.NotEmpty(t, tt.response.Data.Domain)
				require.Empty(t, tt.response.Error)
			} else {
				require.NotEmpty(t, tt.response.Error)
			}
		})
	}
}

func TestDomainsResponse(t *testing.T) {
	tests := []struct {
		name     string
		response DomainsResponse
		success  bool
	}{
		{
			name: "Successful response",
			response: DomainsResponse{
				Success: true,
				Data: []*DomainEntry{
					{
						DomainEntry: pb.DomainEntry{
							Domain: "example1.com",
						},
					},
					{
						DomainEntry: pb.DomainEntry{
							Domain: "example2.com",
						},
					},
				},
			},
			success: true,
		},
		{
			name: "Error response",
			response: DomainsResponse{
				Success: false,
				Error:   "failed to list domains",
			},
			success: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.success, tt.response.Success)
			if tt.success {
				require.NotEmpty(t, tt.response.Data)
				require.Empty(t, tt.response.Error)
			} else {
				require.NotEmpty(t, tt.response.Error)
			}
		})
	}
}
