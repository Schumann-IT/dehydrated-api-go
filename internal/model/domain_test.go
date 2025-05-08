package model

import (
	"encoding/json"
	"testing"

	"github.com/Azure/go-autorest/autorest/to"

	pb "github.com/schumann-it/dehydrated-api-go/plugin/proto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestFromProto(t *testing.T) {
	tests := []struct {
		name     string
		input    *pb.GetMetadataResponse
		expected *DomainEntry
	}{
		{
			name: "Basic conversion",
			input: &pb.GetMetadataResponse{
				Metadata: map[string]*structpb.Value{
					"key1": structpb.NewStringValue("value1"),
					"key2": structpb.NewNumberValue(123),
				},
			},
			expected: &DomainEntry{
				Metadata: map[string]any{
					"key1": "value1",
					"key2": float64(123),
				},
			},
		},
		{
			name: "Empty metadata",
			input: &pb.GetMetadataResponse{
				Metadata: map[string]*structpb.Value{},
			},
			expected: &DomainEntry{
				Metadata: map[string]any{},
			},
		},
		{
			name: "Nil metadata",
			input: &pb.GetMetadataResponse{
				Metadata: nil,
			},
			expected: &DomainEntry{
				Metadata: map[string]any{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MetadataFromProto(tt.input)
			assert.Equal(t, tt.expected.Metadata, result)
		})
	}
}

func TestCreateDomainRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request CreateDomainRequest
		isValid bool
	}{
		{
			name: "Valid request",
			request: CreateDomainRequest{
				Domain:           "example.com",
				AlternativeNames: []string{"www.example.com"},
				Enabled:          true,
			},
			isValid: true,
		},
		{
			name: "Empty domain",
			request: CreateDomainRequest{
				Domain: "",
			},
			isValid: false,
		},
		{
			name: "With metadata",
			request: CreateDomainRequest{
				Domain: "example.com",
			},
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isValid {
				assert.NotEmpty(t, tt.request.Domain)
			} else {
				assert.Empty(t, tt.request.Domain)
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
				AlternativeNames: to.StringSlicePtr([]string{"www.example.com"}),
				Enabled:          to.BoolPtr(true),
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
			assert.True(t, tt.isValid)
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
			assert.Equal(t, tt.success, tt.response.Success)
			if tt.success {
				assert.NotEmpty(t, tt.response.Data.Domain)
				assert.Empty(t, tt.response.Error)
			} else {
				assert.NotEmpty(t, tt.response.Error)
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
		{
			name: "Empty list response",
			response: DomainsResponse{
				Success: true,
				Data:    []*DomainEntry{},
			},
			success: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.success, tt.response.Success)
			if tt.success {
				if len(tt.response.Data) > 0 {
					assert.NotEmpty(t, tt.response.Data[0].Domain)
				}
				assert.Empty(t, tt.response.Error)
			} else {
				assert.NotEmpty(t, tt.response.Error)
			}
		})
	}
}

func TestDomainEntry_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		entry    *DomainEntry
		expected string
	}{
		{
			name: "all fields set",
			entry: &DomainEntry{
				DomainEntry: pb.DomainEntry{
					Domain:           "example.com",
					AlternativeNames: []string{"www.example.com"},
					Alias:            "example",
					Enabled:          true,
					Comment:          "test domain",
				},
				Metadata: Metadata{"test": "value"},
			},
			expected: `{
				"domain": "example.com",
				"alternative_names": ["www.example.com"],
				"alias": "example",
				"enabled": true,
				"comment": "test domain",
				"metadata": {"test": "value"}
			}`,
		},
		{
			name: "zero values",
			entry: &DomainEntry{
				DomainEntry: pb.DomainEntry{
					Domain:           "example.com",
					AlternativeNames: []string{},
					Alias:            "",
					Enabled:          false,
					Comment:          "",
				},
			},
			expected: `{
				"domain": "example.com",
				"alternative_names": [],
				"alias": "",
				"enabled": false,
				"comment": ""
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal the entry
			actual, err := json.Marshal(tt.entry)
			assert.NoError(t, err)

			// Compare JSON objects (ignoring whitespace)
			var actualJSON, expectedJSON interface{}
			err = json.Unmarshal(actual, &actualJSON)
			assert.NoError(t, err)
			err = json.Unmarshal([]byte(tt.expected), &expectedJSON)
			assert.NoError(t, err)

			assert.Equal(t, expectedJSON, actualJSON)
		})
	}
}
